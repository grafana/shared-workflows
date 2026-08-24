package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/grafana/shared-workflows/actions/component-change-detection/pkg/changedetector"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	f := parseFlags()

	// Setup logger
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	if !*f.verbose {
		logger = level.NewFilter(logger, level.AllowNone())
	}

	if *f.tagsPath == "" {
		return errors.New("--tags flag is required")
	}

	config, tags, err := loadConfigAndTags(logger, *f.configPath, *f.tagsPath, *f.target)
	if err != nil {
		return err
	}

	changes, err := detectChanges(config, tags, *f.target)
	if err != nil {
		return err
	}

	// Always write boolean changes to output file
	data, err := json.MarshalIndent(changes, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling results: %w", err)
	}

	if err := writeOutput(*f.output, data, logger); err != nil {
		return err
	}

	logSummary(logger, changes)

	return nil
}

type flags struct {
	configPath *string
	tagsPath   *string
	target     *string
	output     *string
	verbose    *bool
}

func parseFlags() flags {
	f := flags{
		configPath: flag.String("config", ".component-deps.yaml", "Path to dependency configuration file"),
		tagsPath:   flag.String("tags", "", "Path to JSON file with current tags"),
		target:     flag.String("target", "HEAD", "Git ref to compare against (commit, branch, or tag)"),
		output:     flag.String("output", "", "Output file path (default: stdout)"),
		verbose:    flag.Bool("verbose", false, "Enable verbose logging"),
	}
	flag.Parse()
	return f
}

func loadConfigAndTags(
	logger log.Logger,
	configPath, tagsPath, target string,
) (*changedetector.Config, changedetector.Tags, error) {
	level.Info(logger).Log("msg", "Loading configuration", "path", configPath)
	config, err := changedetector.LoadConfig(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading config: %w", err)
	}
	level.Info(logger).Log("msg", "configuration loaded", "components", len(config.Components))

	level.Info(logger).Log("msg", "loading tags", "path", tagsPath)
	tags, err := changedetector.LoadTags(tagsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading tags: %w", err)
	}
	level.Info(logger).Log("msg", "tags loaded", "components", len(tags), "target", target)

	return config, tags, nil
}

func detectChanges(
	config *changedetector.Config,
	tags changedetector.Tags,
	target string,
) (map[string]bool, error) {
	detector := changedetector.NewDetector(config, tags, target)
	changes, err := detector.DetectChanges()
	if err != nil {
		return nil, fmt.Errorf("error detecting changes: %w", err)
	}
	return changes, nil
}

func writeOutput(output string, data []byte, logger log.Logger) error {
	if output != "" {
		if err := os.WriteFile(output, data, 0600); err != nil {
			return fmt.Errorf("error writing output file: %w", err)
		}
		level.Info(logger).Log("msg", "results written", "path", output)
	} else {
		_, _ = fmt.Println(string(data))
	}
	return nil
}

func logSummary(logger log.Logger, changes map[string]bool) {
	changedCount := 0
	for _, changed := range changes {
		if changed {
			changedCount++
		}
	}
	level.Info(logger).Log("msg", "detection complete", "changed", changedCount, "total", len(changes))
}
