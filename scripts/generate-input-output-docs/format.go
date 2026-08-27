package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Formatter runs markdown through prettier.
//
// Prettier is the repo's formatter of record -- a pre-commit hook checks every
// markdown file against it -- so generated docs have to come out the far side of
// prettier byte-for-byte, or the hook and this generator would take turns
// rewriting each other and the CI drift check would never pass.
type Formatter struct {
	argv []string
}

// PinnedPrettierVersion reads the prettier version pinned in package.json.
//
// The exact version matters, not just "some prettier": 3.8.3 escapes an asterisk
// in a table cell as `\*` and 3.9.6 leaves it bare. Generating with a different
// version than CI uses produces a one-byte difference, which shows up as a drift
// failure on a PR whose docs were just regenerated.
func PinnedPrettierVersion(root string) (string, error) {
	path := filepath.Join(root, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var pkg struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}

	version := pkg.DevDependencies["prettier"]
	if version == "" {
		return "", fmt.Errorf("%s does not pin a prettier version in devDependencies", path)
	}
	return version, nil
}

// candidateFormatters returns the commands to try, best first.
//
// The repo-local install is preferred because it is the pinned version by
// definition; bunx and npx resolve it too once dependencies are installed. A
// global prettier comes last and is only accepted if its version happens to
// match, since a contributor's global install is the most likely source of skew.
func candidateFormatters(root string) [][]string {
	return [][]string{
		{filepath.Join(root, "node_modules", ".bin", "prettier")},
		{"bunx", "prettier"},
		{"npx", "--no-install", "prettier"},
		{"prettier"},
	}
}

// NewFormatter resolves a prettier matching the version pinned in package.json.
//
// A non-empty command overrides detection entirely and is split on spaces, so
// `-prettier "bunx prettier"` works. The override is trusted without a version
// check -- it exists so callers can point at something deliberate, including a
// specific version for testing.
//
// Otherwise each candidate is run, not merely looked up on PATH: the launcher
// being present says nothing about prettier being reachable through it. A machine
// with Node but no node_modules has `npx`, and `npx --no-install prettier` fails
// there. Candidates whose version does not match the pin are rejected.
func NewFormatter(command string) (*Formatter, error) {
	return newFormatter(command, ".")
}

// NewFormatterForRoot resolves prettier relative to a repository root, so the
// repo-local install and package.json are found when running from elsewhere.
func NewFormatterForRoot(command, root string) (*Formatter, error) {
	return newFormatter(command, root)
}

func newFormatter(command, root string) (*Formatter, error) {
	if strings.TrimSpace(command) != "" {
		return &Formatter{argv: strings.Fields(command)}, nil
	}

	want, err := PinnedPrettierVersion(root)
	if err != nil {
		return nil, err
	}

	var tried []string
	for _, candidate := range candidateFormatters(root) {
		if _, err := exec.LookPath(candidate[0]); err != nil {
			continue
		}
		formatter := &Formatter{argv: candidate}
		got, err := formatter.version()
		if err != nil {
			continue
		}
		if got == want {
			return formatter, nil
		}
		tried = append(tried, fmt.Sprintf("%s (%s)", candidate[0], got))
	}

	if len(tried) > 0 {
		return nil, fmt.Errorf(
			"prettier %s is required (pinned in package.json) but only found %s; run `bun install`, or pass -prettier to override",
			want, strings.Join(tried, ", "),
		)
	}
	return nil, fmt.Errorf(
		"could not find prettier %s (pinned in package.json); run `bun install`, or pass -prettier to override",
		want,
	)
}

// version runs the candidate with --version and returns what it reports, which
// doubles as a check that prettier really is reachable through this launcher.
func (f *Formatter) version() (string, error) {
	argv := append(append([]string{}, f.argv...), "--version")
	cmd := exec.Command(argv[0], argv[1:]...) //nolint:gosec // argv comes from candidateFormatters or an explicit flag

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running %s: %w: %s", strings.Join(argv, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Format returns doc as prettier would write it. filename is passed through as
// --stdin-filepath so prettier resolves .prettierrc and picks the right parser
// from the extension.
func (f *Formatter) Format(doc, filename string) (string, error) {
	argv := append(append([]string{}, f.argv...), "--stdin-filepath", filename)
	cmd := exec.Command(argv[0], argv[1:]...) //nolint:gosec // argv is a fixed command plus a repo path
	cmd.Stdin = strings.NewReader(doc)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Never fall back to the unformatted text here. Silently returning
		// unformatted markdown is what made the original version of this tool
		// look like it worked while doing nothing.
		return "", fmt.Errorf("running %s: %w: %s", strings.Join(argv, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
