package main

import (
	"fmt"
	"io"
	"os"
)

// GenerateOptions configures a run of GenerateDocs.
type GenerateOptions struct {
	// Root is the repository root to discover targets under.
	Root string
	// Formatter runs the finished markdown through prettier.
	Formatter *Formatter
	// CheckOnly reports what would change without writing anything, which is
	// what CI uses to detect drift.
	CheckOnly bool
	// Verbose logs every target rather than only the changed ones.
	Verbose bool
	// Log receives progress output. Defaults to os.Stdout when nil.
	Log io.Writer
}

// GenerateDocs brings every action README and reusable workflow doc in the repo
// into line with the YAML that declares it, and returns the repo-relative paths
// it changed (or would change, when CheckOnly is set).
//
// This is deliberately free of CLI types so it can be tested directly against a
// temporary repository; runGenerate is the thin adapter over it.
func GenerateDocs(opts GenerateOptions) ([]string, error) {
	log := opts.Log
	if log == nil {
		log = os.Stdout
	}

	targets, err := DiscoverTargets(opts.Root)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no actions or reusable workflows found under %s", opts.Root)
	}

	var changed []string
	for _, target := range targets {
		spec, err := ParseFile(target.YAML, target.Kind)
		if err != nil {
			return nil, err
		}

		existing, err := os.ReadFile(target.Doc)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			// An action or reusable workflow with no doc file at all still needs
			// its inputs documented, so start one from the target's name.
			existing = []byte("# " + target.Name() + "\n")
		}

		updated, err := RenderDoc(string(existing), spec, target.Doc, opts.Formatter)
		if err != nil {
			return nil, err
		}
		if updated == string(existing) {
			if opts.Verbose {
				fmt.Fprintf(log, "ok      %s\n", rel(opts.Root, target.Doc))
			}
			continue
		}

		changed = append(changed, rel(opts.Root, target.Doc))
		if opts.CheckOnly {
			fmt.Fprintf(log, "drift   %s\n", rel(opts.Root, target.Doc))
			continue
		}
		if err := os.WriteFile(target.Doc, []byte(updated), 0o644); err != nil {
			return nil, err
		}
		fmt.Fprintf(log, "wrote   %s\n", rel(opts.Root, target.Doc))
	}

	return changed, nil
}
