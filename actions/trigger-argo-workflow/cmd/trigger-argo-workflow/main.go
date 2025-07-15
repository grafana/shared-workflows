package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/lmittmann/tint"
	cli "github.com/urfave/cli/v3"
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
	flagPrintConfig      = "print-config"
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
	runMain(os.Args, os.Stdout, os.Stderr)
}

func runMain(args []string, writer io.Writer, errWriter io.Writer) {
	// If we're on a terminal, we use tint, otherwise if we're on GitHub actions
	// we use `willabites/actionslog` to log proper Actions messages, otherwise
	// we use logfmt.
	var lv slog.LevelVar

	logger := slog.New(
		slog.NewTextHandler(errWriter, &slog.HandlerOptions{
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

	app := cli.Command{}
	app.Writer = writer
	app.ErrWriter = errWriter
	app.Name = "Runs the Argo CLI"

	app.Action = func(ctx context.Context, c *cli.Command) error {
		return fmt.Errorf("please specify a command")
	}

	app.Commands = []*cli.Command{
		{
			Name:            "submit",
			SkipFlagParsing: true,
			Writer:          writer,
			ErrWriter:       errWriter,
			Action: func(ctx context.Context, c *cli.Command) error {
				return run(ctx, c, &lv, logger, "submit")
			},
		},
	}

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:     flagPrintConfig,
			Required: false,
			Usage:    "If set thie command will only print the gathered configuration and exist",
		},
		&cli.BoolFlag{
			Name: flagAddCILabels,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("ADD_CI_LABELS"),
			),
			Value: false,
			Usage: "If true, the `--labels` argument will be added with values from the environment. This is forced for the `submit` command",
		},
		&cli.StringFlag{
			Name: flagNamespace,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("ARGO_NAMESPACE"),
			),
			Required: true,
		},
		&cli.StringFlag{
			Name: flagArgoToken,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("ARGO_TOKEN"),
			),
			Usage:    "The Argo token to use for authentication",
			Required: true,
		},
		&cli.StringFlag{
			Name: flagLogLevel,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("LOG_LEVEL"),
			),
			Usage: "Which log level to use",
			Value: "info",
			Action: func(ctx context.Context, c *cli.Command, level string) error {
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
			Name: flagInstance,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("INSTANCE"),
			),
			Value: "ops",
			Action: func(ctx context.Context, c *cli.Command, instance string) error {
				// Validate it is a known instance
				instances := slices.Collect(maps.Keys(instanceToHost))
				if !slices.Contains(instances, instance) {
					return fmt.Errorf("invalid instance: `%s`. choose from: %s", instance, strings.Join(instances, ", "))
				}

				return nil
			},
		},
		&cli.StringSliceFlag{
			Name:  flagParameter,
			Usage: "Parameters to pass to the workflow template. Given as `key=value`. Specify multiple times for multiple parameters",
		},
		&cli.UintFlag{
			Name: flagRetries,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("RETRIES"),
			),
			Value: 3,
			Usage: "Number of retries to make if the command fails",
		},
		&cli.StringFlag{
			Name: flagWorkflowTemplate,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("WORKFLOW_TEMPLATE"),
			),
			Usage: "The workflow template to use",
			Action: func(ctx context.Context, c *cli.Command, tpl string) error {
				// Required if command is `submit`
				if c.Args().First() == "submit" && tpl == "" {
					return fmt.Errorf("required flag: %s", flagWorkflowTemplate)
				}
				return nil
			},
		},
	}

	if err := app.Run(context.Background(), args); err != nil {
		logger.With("error", err).Error("failed to run")
		os.Exit(1)
	}
}

func run(ctx context.Context, c *cli.Command, level *slog.LevelVar, logger *slog.Logger, command string) error {
	addCILabels := c.Bool(flagAddCILabels)
	argoToken := c.String(flagArgoToken)
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

	extraArgs = append(extraArgs, c.Args().Slice()...)

	md, err := NewGitHubActionsMetadata()
	if err != nil {
		return fmt.Errorf("failed to get GitHub Actions metadata: %w", err)
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

	if c.Bool(flagPrintConfig) {
		if err := argo.PrintConfig(c.Writer, md); err != nil {
			return err
		}
		return nil
	}

	logger.With("extraArgs", extraArgs).Info("running command")
	return argo.Run(ctx, md)
}
