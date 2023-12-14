package main

import (
	"bytes"
	"fmt"
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

func (a App) server() string {
	instanceToHost := map[string]string{
		"dev":  "argo-workflows-dev",
		"ops": "argo-workflows",
	}
	return fmt.Sprintf("%s.grafana.net:443", instanceToHost[a.instance])
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
	output := input.String()

	matches := nameRe.FindStringSubmatch(output)

	var uri string
	if len(matches) == 2 {
		uri = fmt.Sprintf("https://%s/workflows/%s/%s", a.server(), a.namespace, matches[1])
	}

	return uri, strings.TrimSuffix(output, "\n")
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

func (a *App) setURIAsJobOutput(uri string) {
	if a.command != "submit" {
		a.logger.With("command", a.command).Debug("not setting job output, command is not `submit`")
		return
	}

	githubOutput := os.Getenv("GITHUB_OUTPUT")

	if githubOutput == "" {
		a.logger.Warn("GITHUB_OUTPUT not set, not setting job output")
		return
	}

	a.logger.With("github_output", githubOutput).Debug("setting job output")

	f, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		a.logger.With("error", err).Error("failed to open file")
		return
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("uri=%s\n", uri))
	if err != nil {
		a.logger.With("error", err).Error("failed to write to file")
	}
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

	if uri != "" {
		a.setURIAsJobOutput(uri)
	}

	fmt.Println(out)

	return nil
}
