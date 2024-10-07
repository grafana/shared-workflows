package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
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

var instanceToHost = map[string]string{
	"dev":     "argo-workflows-dev.grafana.net:443",
	"dev-aws": "argo-workflows-aws.grafana-dev.net:443",
	"ops":     "argo-workflows.grafana.net:443",
	"ops-aws": "argo-workflows-aws.grafana.net:443",
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

func (a App) outputWithURI(input *bytes.Buffer) (string, string) {
	output := strings.TrimSuffix(input.String(), "\n")

	matches := nameRe.FindStringSubmatch(output)

	if len(matches) != 2 {
		a.logger.Warn("Couldn't find workflow name in output - can't construct URI for launched workflow")
		return "", output
	}

	uri := fmt.Sprintf("https://%s/workflows/%s/%s", a.server(), a.namespace, matches[1])

	return uri, output
}

func (a App) runCmd(md GitHubActionsMetadata) (string, string, error) {
	args := a.args(md)

	cmd := exec.Command("argo", args...)
	cmdOutput := &bytes.Buffer{}

	cmd.Env = a.env()
	cmd.Stdout = cmdOutput
	cmd.Stderr = os.Stderr

	a.logger.With("executable", "argo", "command", cmd.Args, "retries", a.retries).Debug("running command")

	err := cmd.Run()
	if err != nil {
		return "", "", err
	}

	uri, output := a.outputWithURI(cmdOutput)

	return uri, output, nil
}

func (a *App) setURIAsJobOutput(uri string, writer io.Writer) {
	if a.command != "submit" {
		a.logger.With("command", a.command).Debug("command is not `submit`, won't set job output")
		return
	}

	_, err := writer.Write([]byte(fmt.Sprintf("uri=%s\n", uri)))
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

func (a *App) Run(md GitHubActionsMetadata) error {
	bo := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), a.retries)

	var uri string
	var out string

	run := func() error {
		var err error
		uri, out, err = a.runCmd(md)

		return err
	}

	err := backoff.RetryNotify(run, bo, func(err error, t time.Duration) {
		a.logger.With("error", err, "retry_in", t).Error("failed to run command, retrying")
	})
	if err != nil {
		return err
	}

	a.logger.With("uri", uri).Info("workflow URI")

	writer := a.openGitHubOutput()
	defer writer.Close()

	if writer != nil && uri != "" {
		a.setURIAsJobOutput(uri, writer)
	}

	fmt.Println(out)

	return nil
}
