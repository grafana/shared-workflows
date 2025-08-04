package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
)

type App struct {
	levelVar *slog.LevelVar
	logger   *slog.Logger

	argoToken string

	addCILabels bool

	command          string
	workflowTemplate string
	extraArgs        []string

	namespace string
	instance  string

	retries uint64
}

type FullConfig struct {
	ArgoToken             string
	LogLevel              *slog.LevelVar
	AddCILabels           bool
	Command               string
	ExtraArgs             []string
	Instance              string
	Namespace             string
	Retries               uint64
	WorkflowTemplate      string
	GitHubActionsMetadata GitHubActionsMetadata
}

func (a App) PrintConfig(w io.Writer, md GitHubActionsMetadata) error {
	cfg := FullConfig{
		ArgoToken:             a.argoToken,
		LogLevel:              a.levelVar,
		Command:               a.command,
		WorkflowTemplate:      a.workflowTemplate,
		Instance:              a.instance,
		Namespace:             a.namespace,
		Retries:               a.retries,
		ExtraArgs:             a.extraArgs,
		GitHubActionsMetadata: md,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

var instanceToHost = map[string]string{
	"dev": "argo-workflows-dev.grafana.net:443",
	"ops": "argo-workflows.grafana.net:443",
}

func (a App) server() string {
	return instanceToHost[a.instance]
}

func (a App) env() []string {
	// Argo CLI requires either `Token ` or `Bearer ` before the token, add it if it's missing.
	if !strings.HasPrefix(a.argoToken, "Bearer") && !strings.HasPrefix(a.argoToken, "Token") {
		a.argoToken = "Bearer " + a.argoToken
	}

	return []string{
		"ARGO_HTTP1=true",
		"KUBECONFIG=/dev/null",
		fmt.Sprintf("ARGO_NAMESPACE=%s", a.namespace),
		fmt.Sprintf("ARGO_SERVER=%s", a.server()),
		fmt.Sprintf("ARGO_TOKEN=%s", a.argoToken),
	}
}

var nameRe = regexp.MustCompile(`^Name:\s+(.+)`)

func (a App) outputWithURI(reader io.Reader) (string, string, error) {
	scanner := bufio.NewScanner(reader)

	var uri string
	var outputBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		outputBuilder.WriteString(line + "\n")

		if uri == "" {
			matches := nameRe.FindStringSubmatch(line)
			if len(matches) == 2 {
				uri = fmt.Sprintf("https://%s/workflows/%s/%s", a.server(), a.namespace, matches[1])
				a.logger.With("uri", uri).Info("workflow URI")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return uri, outputBuilder.String(), fmt.Errorf("error reading command output: %w", err)
	}

	return uri, outputBuilder.String(), nil
}

func (a App) runCmd(md GitHubActionsMetadata) (string, string, error) {
	args := a.args(md)

	cmd := exec.Command("argo", args...)
	cmd.Env = a.env()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	a.logger.With("executable", "argo", "command", cmd.Args, "retries", a.retries).Debug("running command")

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start command: %w", err)
	}

	uri, out, scanErr := a.outputWithURI(stdoutPipe)
	if scanErr != nil {
		_ = stdoutPipe.Close()
		_ = cmd.Wait()
		return uri, out, scanErr
	}

	if err := cmd.Wait(); err != nil {
		return uri, out, fmt.Errorf("command failed: %w", err)
	}

	return uri, out, nil
}

func (a *App) setURIAsJobOutput(uri string, writer io.Writer) {
	if a.command != "submit" {
		a.logger.With("command", a.command).Debug("command is not `submit`, won't set job output")
		return
	}

	_, err := fmt.Fprintf(writer, "uri=%s\n", uri)
	if err != nil {
		a.logger.With("error", err).Error("failed to write to file, won't set job output")
	}
}

func (a *App) openGitHubOutput() io.WriteCloser {
	githubOutput := os.Getenv("GITHUB_OUTPUT")

	if githubOutput == "" {
		a.logger.Warn("GITHUB_OUTPUT not set, won't set job output")
		return nil
	}

	a.logger.With("github_output", githubOutput).Debug("setting job output")

	f, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		a.logger.With("error", err).Error("failed to open file, won't set job output")
		return nil
	}

	return f
}

var fatalErrors = []string{
	"AlreadyExists",
}

func isFatalError(err error) bool {
	if err == nil {
		return false
	}

	for _, fatalError := range fatalErrors {
		if strings.Contains(err.Error(), fatalError) {
			return true
		}
	}
	return false
}

func (a *App) Run(ctx context.Context, md GitHubActionsMetadata) error {
	bo := backoff.NewExponentialBackOff()

	var out string

	run := func() (string, error) {
		var uri string
		var err error
		uri, out, err = a.runCmd(md)

		if isFatalError(err) {
			return uri, backoff.Permanent(err)
		}

		return uri, err
	}

	uri, err := backoff.Retry(ctx, run, backoff.WithBackOff(bo), backoff.WithMaxTries(uint(a.retries)), backoff.WithNotify(func(err error, t time.Duration) {
		a.logger.With("error", err, "retry_in", t).Error("failed to run command, retrying")
	}))
	if err != nil {
		return err
	}

	writer := a.openGitHubOutput()
	defer func() {
		_ = writer.Close()
	}()

	if writer != nil && uri != "" {
		a.setURIAsJobOutput(uri, writer)
	}

	fmt.Println(out)

	return nil
}
