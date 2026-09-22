package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestRepo builds a throwaway repository containing one composite action, one
// reusable workflow and one ordinary workflow, and returns its root.
//
// package.json is copied from the real repo so the pinned prettier version
// resolves the same way it does in anger.
func newTestRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	mustMkdirAll(t, filepath.Join(root, "actions", "thing"))
	mustMkdirAll(t, filepath.Join(root, ".github", "workflows"))

	pkg, err := os.ReadFile(filepath.Join("..", "..", "package.json"))
	if err != nil {
		t.Fatalf("reading package.json: %v", err)
	}
	mustWrite(t, filepath.Join(root, "package.json"), string(pkg))

	mustWrite(t, filepath.Join(root, "actions", "thing", "action.yml"), `name: Thing
description: A composite action.

inputs:
  alpha:
    description: "The alpha input."
    required: true
  beta:
    description: "The beta input."
    default: "two"

outputs:
  result:
    description: "The result."
    value: ${{ steps.run.outputs.result }}

runs:
  using: composite
  steps:
    - id: run
      shell: bash
      run: echo "result=ok" >> "$GITHUB_OUTPUT"
`)

	mustWrite(t, filepath.Join(root, ".github", "workflows", "reusable.yml"), `name: Reusable

on:
  workflow_call:
    inputs:
      flag:
        description: "A flag."
        type: boolean
        default: false

jobs:
  run:
    runs-on: ubuntu-latest
    steps:
      - run: "true"
`)

	// An ordinary workflow must be ignored entirely -- no doc created for it.
	mustWrite(t, filepath.Join(root, ".github", "workflows", "ordinary.yml"), `name: Ordinary

on: push

jobs:
  run:
    runs-on: ubuntu-latest
    steps:
      - run: "true"
`)

	return root
}

func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func generateInto(t *testing.T, root string, checkOnly bool) []string {
	t.Helper()
	changed, err := GenerateDocs(GenerateOptions{
		Root:      root,
		Formatter: newTestFormatter(t),
		CheckOnly: checkOnly,
		Log:       io.Discard,
	})
	if err != nil {
		t.Fatalf("GenerateDocs: %v", err)
	}
	return changed
}

func TestGenerateDocsCreatesMissingDocs(t *testing.T) {
	// The path CI depends on most: a brand-new action with no README at all.
	// Nothing else exercises it, and because a created file is untracked, this is
	// exactly the case a plain `git diff` drift check would miss.
	root := newTestRepo(t)

	changed := generateInto(t, root, false)
	if len(changed) != 2 {
		t.Fatalf("changed = %v, want 2 files (action README + workflow doc)", changed)
	}

	readme := mustRead(t, filepath.Join(root, "actions", "thing", "README.md"))
	if !strings.HasPrefix(readme, "# thing\n") {
		t.Errorf("created README should be titled from the action directory:\n%s", readme)
	}
	for _, want := range []string{"## Inputs", "`alpha`", "`beta`", "`two`", "## Outputs", "`result`"} {
		if !strings.Contains(readme, want) {
			t.Errorf("created README missing %q:\n%s", want, readme)
		}
	}

	doc := mustRead(t, filepath.Join(root, ".github", "workflows", "reusable.md"))
	if !strings.Contains(doc, "`flag`") || !strings.Contains(doc, "boolean") {
		t.Errorf("created workflow doc missing its input:\n%s", doc)
	}
}

func TestGenerateDocsIgnoresOrdinaryWorkflows(t *testing.T) {
	root := newTestRepo(t)
	generateInto(t, root, false)

	// `on: push` carries no inputs, so it must not gain a doc.
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "ordinary.md")); !os.IsNotExist(err) {
		t.Errorf("ordinary.md should not have been created (err = %v)", err)
	}
}

func TestGenerateDocsIsANoOpWhenUpToDate(t *testing.T) {
	root := newTestRepo(t)
	generateInto(t, root, false)

	if changed := generateInto(t, root, false); len(changed) != 0 {
		t.Errorf("second run changed %v, want nothing", changed)
	}
}

func TestGenerateDocsCheckOnlyReportsWithoutWriting(t *testing.T) {
	root := newTestRepo(t)
	generateInto(t, root, false)

	// Knock the table out of sync the way a YAML edit would.
	readme := filepath.Join(root, "actions", "thing", "README.md")
	before := mustRead(t, readme)
	mustWrite(t, readme, strings.Replace(before, "The alpha input.", "Stale description.", 1))
	stale := mustRead(t, readme)

	changed := generateInto(t, root, true)
	if len(changed) != 1 || !strings.Contains(changed[0], "README.md") {
		t.Errorf("changed = %v, want the one stale README", changed)
	}
	if got := mustRead(t, readme); got != stale {
		t.Error("CheckOnly must not write to disk")
	}

	// And without CheckOnly the same drift is repaired.
	if changed := generateInto(t, root, false); len(changed) != 1 {
		t.Errorf("changed = %v, want the README repaired", changed)
	}
	if got := mustRead(t, readme); got != before {
		t.Errorf("README not restored to generated form:\n%s", got)
	}
}

func TestGenerateDocsPreservesHandWrittenProse(t *testing.T) {
	root := newTestRepo(t)
	generateInto(t, root, false)

	readme := filepath.Join(root, "actions", "thing", "README.md")
	withProse := mustRead(t, readme) + "\n## Notes\n\nHand-written, must survive.\n"
	mustWrite(t, readme, withProse)

	// Regenerating must leave everything outside the markers alone.
	generateInto(t, root, false)

	if got := mustRead(t, readme); !strings.Contains(got, "Hand-written, must survive.") {
		t.Errorf("prose outside the markers was lost:\n%s", got)
	}
}

func TestGenerateDocsErrorsOnEmptyRepo(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "actions"))
	mustMkdirAll(t, filepath.Join(root, ".github", "workflows"))

	_, err := GenerateDocs(GenerateOptions{
		Root:      root,
		Formatter: newTestFormatter(t),
		Log:       io.Discard,
	})
	if err == nil {
		t.Error("expected an error when no actions or workflows are found")
	}
}
