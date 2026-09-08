// Command generate-input-output-docs keeps the input/output tables in this
// repo's action READMEs and reusable workflow docs in sync with the YAML that
// declares them.
//
// See README.md in this directory for usage.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

func main() {
	if err := newCommand().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// rootDirFlag is shared by every subcommand: the tool always works relative to
// a repository root, and resolves the pinned prettier and package.json from it.
func rootDirFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "root-dir",
		Usage: "path to the repository root",
		Value: "../..",
	}
}

func prettierFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "prettier",
		Usage: "prettier command to format with (default: the version pinned in package.json)",
	}
}

func newCommand() *cli.Command {
	return &cli.Command{
		Name:  "generate-input-output-docs",
		Usage: "keep action and reusable workflow docs in sync with their YAML",
		Commands: []*cli.Command{
			{
				Name:   "generate",
				Usage:  "rewrite the generated tables in every doc file in the repo",
				Flags:  []cli.Flag{rootDirFlag(), prettierFlag(), verboseFlag()},
				Action: func(_ context.Context, cmd *cli.Command) error { return runGenerate(cmd, false) },
			},
			{
				Name:   "check",
				Usage:  "report which doc files are out of date, without writing (exits non-zero on drift)",
				Flags:  []cli.Flag{rootDirFlag(), prettierFlag(), verboseFlag()},
				Action: func(_ context.Context, cmd *cli.Command) error { return runGenerate(cmd, true) },
			},
			{
				Name:   "parity",
				Usage:  "verify a reusable workflow forwards the inputs/outputs of its action",
				Flags:  []cli.Flag{rootDirFlag(), parityConfigFlag()},
				Action: func(_ context.Context, cmd *cli.Command) error { return runParity(cmd) },
			},
			{
				Name:   "print",
				Usage:  "print the tables for a single YAML file to stdout",
				Flags:  []cli.Flag{rootDirFlag(), prettierFlag(), fileFlag(), kindFlag()},
				Action: func(_ context.Context, cmd *cli.Command) error { return runPrint(cmd) },
			},
		},
	}
}

func verboseFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:  "verbose",
		Usage: "log every target, not just the changed ones",
	}
}

func parityConfigFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "config",
		Usage: "path to the parity rules file (default <root-dir>/scripts/generate-input-output-docs/parity.yaml)",
	}
}

func fileFlag() cli.Flag {
	return &cli.StringFlag{
		Name:     "file",
		Aliases:  []string{"f"},
		Usage:    "path to an action.yml or reusable workflow",
		Required: true,
	}
}

func kindFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "kind",
		Usage: "one of: auto, composite, workflow",
		Value: "auto",
	}
}

// runGenerate rewrites every doc file in the repo. When checkOnly is set it
// reports what would change and fails instead of writing, which is what CI runs.
func runGenerate(cmd *cli.Command, checkOnly bool) error {
	root := cmd.String("root-dir")
	verbose := cmd.Bool("verbose")

	formatter, err := NewFormatterForRoot(cmd.String("prettier"), root)
	if err != nil {
		return err
	}

	targets, err := DiscoverTargets(root)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no actions or reusable workflows found under %s", root)
	}

	var changed []string
	for _, target := range targets {
		spec, err := ParseFile(target.YAML, target.Kind)
		if err != nil {
			return err
		}

		existing, err := os.ReadFile(target.Doc)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			// A reusable workflow or action with no doc file at all still needs
			// its inputs documented, so start one from the target's name.
			existing = []byte("# " + target.Name() + "\n")
		}

		updated, err := RenderDoc(string(existing), spec, target.Doc, formatter)
		if err != nil {
			return err
		}
		if updated == string(existing) {
			if verbose {
				fmt.Printf("ok      %s\n", rel(root, target.Doc))
			}
			continue
		}

		changed = append(changed, rel(root, target.Doc))
		if checkOnly {
			fmt.Printf("drift   %s\n", rel(root, target.Doc))
			continue
		}
		if err := os.WriteFile(target.Doc, []byte(updated), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote   %s\n", rel(root, target.Doc))
	}

	if checkOnly && len(changed) > 0 {
		return fmt.Errorf("%d doc file(s) are out of date; run `go run . generate` to update them", len(changed))
	}
	return nil
}

func runParity(cmd *cli.Command) error {
	root := cmd.String("root-dir")

	path := cmd.String("config")
	if path == "" {
		path = filepath.Join(root, "scripts", "generate-input-output-docs", "parity.yaml")
	}
	cfg, err := LoadParityConfig(path)
	if err != nil {
		return err
	}
	if len(cfg.Rules) == 0 {
		return fmt.Errorf("%s declares no rules", path)
	}

	var total int
	for _, rule := range cfg.Rules {
		problems, err := rule.Check(root)
		if err != nil {
			return err
		}
		if len(problems) == 0 {
			fmt.Printf("ok      %s <-> %s\n", rule.Workflow, rule.Action)
			continue
		}
		fmt.Printf("drift   %s <-> %s\n", rule.Workflow, rule.Action)
		for _, p := range problems {
			fmt.Printf("        %s\n", p)
		}
		total += len(problems)
	}

	if total > 0 {
		return fmt.Errorf("%d parity problem(s) found", total)
	}
	return nil
}

// runPrint dumps the tables for one file, for use when writing docs by hand or
// eyeballing what the generator would produce.
func runPrint(cmd *cli.Command) error {
	file := cmd.String("file")

	formatter, err := NewFormatterForRoot(cmd.String("prettier"), cmd.String("root-dir"))
	if err != nil {
		return err
	}

	kind, err := resolveKind(file, cmd.String("kind"))
	if err != nil {
		return err
	}
	spec, err := ParseFile(file, kind)
	if err != nil {
		return err
	}

	var b strings.Builder
	if table := RenderInputs(spec.Inputs); table != "" {
		b.WriteString("## Inputs\n\n")
		b.WriteString(table)
	}
	if table := RenderOutputs(spec.Outputs); table != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("## Outputs\n\n")
		b.WriteString(table)
	}

	// Format as README.md so the output is identical to what `generate` would
	// write, and so it can be pasted straight into a doc.
	formatted, err := formatter.Format(b.String(), "README.md")
	if err != nil {
		return err
	}
	fmt.Print(formatted)
	return nil
}

// resolveKind picks the parser for a file. "auto" infers it from the filename,
// since every composite action in this repo is named action.yml or action.yaml.
func resolveKind(file, kindFlag string) (Kind, error) {
	switch kindFlag {
	case "composite":
		return KindCompositeAction, nil
	case "workflow":
		return KindReusableWorkflow, nil
	case "auto":
		base := filepath.Base(file)
		if base == "action.yml" || base == "action.yaml" {
			return KindCompositeAction, nil
		}
		return KindReusableWorkflow, nil
	default:
		return 0, fmt.Errorf("unknown --kind %q, want auto, composite or workflow", kindFlag)
	}
}

// rel shortens a path for logging, falling back to the full path if it does not
// sit under root.
func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}
