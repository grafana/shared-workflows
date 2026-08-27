// Command generate-input-output-docs keeps the input/output tables in this
// repo's action READMEs and reusable workflow docs in sync with the YAML that
// declares them.
//
// See README.md in this directory for usage.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

const usage = `generate-input-output-docs keeps action and reusable workflow docs in sync with their YAML.

Usage:
  generate-input-output-docs <command> [flags]

Commands:
  generate   Rewrite the generated tables in every doc file in the repo
  check      Report which doc files are out of date, without writing (exit 1 on drift)
  parity     Verify a reusable workflow forwards the inputs/outputs of its action
  print      Print the tables for a single YAML file to stdout

Run "generate-input-output-docs <command> -h" for the flags of a command.
`

func run(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return errors.New("no command given")
	}

	switch args[0] {
	case "generate":
		return runGenerate(args[1:], false)
	case "check":
		return runGenerate(args[1:], true)
	case "parity":
		return runParity(args[1:])
	case "print":
		return runPrint(args[1:])
	case "-h", "--help", "help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

// runGenerate rewrites every doc file in the repo. When checkOnly is set it
// reports what would change and fails instead of writing, which is what CI runs.
func runGenerate(args []string, checkOnly bool) error {
	name := "generate"
	if checkOnly {
		name = "check"
	}
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	root := fs.String("root-dir", ".", "path to the repository root")
	verbose := fs.Bool("verbose", false, "log every target, not just the changed ones")
	prettier := fs.String("prettier", "", "prettier command to format with (default: the version pinned in package.json)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	formatter, err := NewFormatterForRoot(*prettier, *root)
	if err != nil {
		return err
	}

	targets, err := DiscoverTargets(*root)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no actions or reusable workflows found under %s", *root)
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
			if *verbose {
				fmt.Printf("ok      %s\n", rel(*root, target.Doc))
			}
			continue
		}

		changed = append(changed, rel(*root, target.Doc))
		if checkOnly {
			fmt.Printf("drift   %s\n", rel(*root, target.Doc))
			continue
		}
		if err := os.WriteFile(target.Doc, []byte(updated), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote   %s\n", rel(*root, target.Doc))
	}

	if checkOnly && len(changed) > 0 {
		return fmt.Errorf("%d doc file(s) are out of date; run `go run . generate` to update them", len(changed))
	}
	return nil
}

func runParity(args []string) error {
	fs := flag.NewFlagSet("parity", flag.ExitOnError)
	root := fs.String("root-dir", ".", "path to the repository root")
	config := fs.String("config", "", "path to the parity rules file (default <root-dir>/scripts/generate-input-output-docs/parity.yaml)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	path := *config
	if path == "" {
		path = filepath.Join(*root, "scripts", "generate-input-output-docs", "parity.yaml")
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
		problems, err := rule.Check(*root)
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
func runPrint(args []string) error {
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	file := fs.String("file", "", "path to an action.yml or reusable workflow (required)")
	kindFlag := fs.String("kind", "auto", "one of: auto, composite, workflow")
	root := fs.String("root-dir", "../..", "path to the repository root, used to find the pinned prettier")
	prettier := fs.String("prettier", "", "prettier command to format with (default: the version pinned in package.json)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return errors.New("-file is required")
	}

	formatter, err := NewFormatterForRoot(*prettier, *root)
	if err != nil {
		return err
	}

	kind, err := resolveKind(*file, *kindFlag)
	if err != nil {
		return err
	}
	spec, err := ParseFile(*file, kind)
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
		return 0, fmt.Errorf("unknown -kind %q, want auto, composite or workflow", kindFlag)
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
