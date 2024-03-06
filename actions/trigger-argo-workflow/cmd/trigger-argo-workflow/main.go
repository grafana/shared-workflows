package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	cli "github.com/urfave/cli/v2"
	"github.com/willabides/actionslog"
	"github.com/willabides/actionslog/human"
	"golang.org/x/term"
)

const (
	flagAddCILabels      = "add-ci-labels"
	flagArgoToken        = "argo-token"
	flagInstance         = "instance"
	flagLogLevel         = "log-level"
	flagNamespace        = "namespace"
	flagParameter        = "parameter"
	flagRetries          = "retries"
	flagWorkflowTemplate = "workflow-template"
)

func parseLogLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: `%s`. choose from: debug, info, warn, error", level)
	}
}

func main() {
	// If we're on a terminal, we use tint, otherwise if we're on GitHub actions
	// we use `willabites/actionslog` to log proper Actions messages, otherwise
	// we use logfmt.
	var lv slog.LevelVar

	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: &lv,
		}),
	)

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		handler := &human.Handler{Level: &lv}

		logger = slog.New(&actionslog.Wrapper{
			Handler: handler.WithOutput,
		})
	}

	if term.IsTerminal(int(os.Stderr.Fd())) {
		logger = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level: &lv,
			}),
		)
	}

	app := cli.NewApp()
	app.Name = "Runs the Argo CLI"

	app.Action = func(c *cli.Context) error { return run(c, &lv, logger) }

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    flagAddCILabels,
			EnvVars: []string{"ADD_CI_LABELS"},
			Value:   false,
			Usage:   "If true, the `--labels` argument will be added with values from the environment. This is forced for the `submit` command",
		},
		&cli.StringFlag{
			Name:     flagNamespace,
			EnvVars:  []string{"ARGO_NAMESPACE"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     flagArgoToken,
			EnvVars:  []string{"ARGO_TOKEN"},
			Usage:    "The Argo token to use for authentication",
			Required: true,
		},
		&cli.StringFlag{
			Name:    flagLogLevel,
			EnvVars: []string{"LOG_LEVEL"},
			Usage:   "Which log level to use",
			Value:   "info",
			Action: func(c *cli.Context, level string) error {
				level = strings.ToLower(level)

				lvl, err := parseLogLevel(level)
				if err != nil {
					return err
				}

				lv.Set(lvl)

				return nil
			},
		},
		&cli.StringFlag{
			Name:    flagInstance,
			EnvVars: []string{"INSTANCE"},
			Value:   "ops",
			Action: func(c *cli.Context, instance string) error {
				// Validate it is "dev" or "ops"
				if instance != "dev" && instance != "ops" {
					return fmt.Errorf("invalid instance: `%s`. choose from: dev, ops", instance)
				}

				return nil
			},
		},
		&cli.StringSliceFlag{
			Name:  flagParameter,
			Usage: "Parameters to pass to the workflow template. Given as `key=value`. Specify multiple times for multiple parameters",
		},
		&cli.UintFlag{
			Name:    flagRetries,
			EnvVars: []string{"RETRIES"},
			Value:   3,
			Usage:   "Number of retries to make if the command fails",
		},
		&cli.StringFlag{
			Name:    flagWorkflowTemplate,
			EnvVars: []string{"WORKFLOW_TEMPLATE"},
			Usage:   "The workflow template to use",
			Action: func(c *cli.Context, tpl string) error {
				// Required if command is `submit`
				if c.Args().First() == "submit" && tpl == "" {
					return fmt.Errorf("required flag: %s", flagWorkflowTemplate)
				}
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.With("error", err).Error("failed to run")
		os.Exit(1)
	}
}

func run(c *cli.Context, level *slog.LevelVar, logger *slog.Logger) error {
	addCILabels := c.Bool(flagAddCILabels)
	argoToken := c.String(flagArgoToken)
	command := c.Args().First()
	instance := c.String(flagInstance)
	namespace := c.String(flagNamespace)
	parameters := c.StringSlice(flagParameter)
	retries := c.Uint64(flagRetries)
	workflowTemplate := c.String(flagWorkflowTemplate)

	var extraArgs []string
	for _, param := range parameters {
		if !strings.Contains(param, "=") {
			return fmt.Errorf("invalid parameter: `%s`. must be in the format `key=value`", param)
		}

		extraArgs = append(extraArgs, "--parameter", param)
	}

	extraArgs = append(extraArgs, c.Args().Tail()...)

	md, err := NewGitHubActionsMetadata()
	if err != nil {
		return fmt.Errorf("failed to get GitHub Actions metadata: %w", err)
	}
	prinfo, err := NewPullRequestInfo(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get GitHub pull request metadata: %w", err)
	}

	logger = logger.With(
		"argo_wf_instance", instance,
		"org", md.Repo.Owner,
		"repo", md.Repo.Name,
		"commit", md.Commit,
		"commit_author", md.CommitAuthor,
		"build_number", md.BuildNumber,
		"build_event", md.BuildEvent,
		"command", command,
		"namespace", namespace,
	)

	logger.With("extraArgs", extraArgs).Info("running command")

	argo := App{
		levelVar: level,
		logger:   logger,

		argoToken: argoToken,

		addCILabels: addCILabels,

		command:          command,
		workflowTemplate: workflowTemplate,
		extraArgs:        extraArgs,

		namespace: namespace,
		instance:  instance,

		retries: retries,
	}

	return argo.Run(md, prinfo)
}
