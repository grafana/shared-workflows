package main

import (
	"os/exec"
	"strings"
	"testing"
)

// newTestFormatter resolves prettier or skips the test. Prettier is a hard
// requirement of the generator, but it comes from the JS toolchain, so a bare
// `go test` on a machine without `bun install` should not fail.
func newTestFormatter(t *testing.T) *Formatter {
	t.Helper()
	formatter, err := NewFormatterForRoot("", "../..")
	if err != nil {
		t.Skipf("pinned prettier not available: %v", err)
	}
	return formatter
}

func TestFormatIsAPrettierFixedPoint(t *testing.T) {
	formatter := newTestFormatter(t)

	spec, err := ParseFile("testdata/action.yaml", KindCompositeAction)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	doc, err := RenderDoc("# Fixture\n", spec, "README.md", formatter)
	if err != nil {
		t.Fatalf("RenderDoc: %v", err)
	}

	// The generated document has to survive prettier untouched. If it does not,
	// the pre-commit hook would rewrite what the generator just wrote and the CI
	// drift check could never pass.
	again, err := formatter.Format(doc, "README.md")
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if doc != again {
		t.Errorf("generated doc is not prettier-stable:\ngenerated:\n%s\nafter prettier:\n%s", doc, again)
	}
}

func TestRenderDocIsIdempotent(t *testing.T) {
	formatter := newTestFormatter(t)

	spec, err := ParseFile("testdata/action.yaml", KindCompositeAction)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	once, err := RenderDoc("# Fixture\n", spec, "README.md", formatter)
	if err != nil {
		t.Fatalf("RenderDoc: %v", err)
	}
	twice, err := RenderDoc(once, spec, "README.md", formatter)
	if err != nil {
		t.Fatalf("RenderDoc: %v", err)
	}

	if once != twice {
		t.Errorf("regenerating changed the document:\nfirst:\n%s\nsecond:\n%s", once, twice)
	}
}

func TestFormatErrorsRatherThanReturningUnformattedText(t *testing.T) {
	// The original version of this tool ran `prettier -w` against stdin, which
	// always failed, and then silently returned the input unchanged -- so it
	// looked like it worked while formatting nothing. Failing loudly is the fix.
	formatter := &Formatter{argv: []string{"definitely-not-a-real-binary-xyz"}}

	if _, err := formatter.Format("# x\n", "README.md"); err == nil {
		t.Fatal("expected an error when the formatter cannot run")
	}
}

func TestNewFormatterHonoursExplicitCommand(t *testing.T) {
	formatter, err := NewFormatter("bunx prettier")
	if err != nil {
		t.Fatalf("NewFormatter: %v", err)
	}
	if got := strings.Join(formatter.argv, " "); got != "bunx prettier" {
		t.Errorf("argv = %q, want %q", got, "bunx prettier")
	}
}

func TestResolvedFormatterMatchesThePinnedVersion(t *testing.T) {
	// The whole point of pinned resolution: whatever gets resolved must report
	// exactly the version CI will use, or generated docs differ by a byte and the
	// drift check fails on a PR whose docs were just regenerated.
	formatter := newTestFormatter(t)

	want, err := PinnedPrettierVersion("../..")
	if err != nil {
		t.Fatalf("PinnedPrettierVersion: %v", err)
	}
	got, err := formatter.version()
	if err != nil {
		t.Fatalf("version(): %v", err)
	}
	if got != want {
		t.Errorf("resolved prettier %s, want the pinned %s", got, want)
	}
}

func TestPinnedPrettierVersionIsReadFromPackageJSON(t *testing.T) {
	version, err := PinnedPrettierVersion("../..")
	if err != nil {
		t.Fatalf("PinnedPrettierVersion: %v", err)
	}
	if version == "" {
		t.Fatal("no version read")
	}
	// A range would defeat the purpose; the pin has to be exact.
	if strings.ContainsAny(version, "^~*x><= ") {
		t.Errorf("prettier is pinned as %q; an exact version is required for reproducible output", version)
	}
}

func TestPinnedPrettierVersionErrorsWithoutPackageJSON(t *testing.T) {
	if _, err := PinnedPrettierVersion(t.TempDir()); err == nil {
		t.Error("expected an error when package.json is absent")
	}
}

func TestVersionRejectsLauncherThatCannotReachPrettier(t *testing.T) {
	// The bug this guards: resolution used to accept any launcher that existed on
	// PATH. A machine with Node but no node_modules has `npx`, so `npx` resolved,
	// nothing skipped, and the failure only surfaced later as a hard error from
	// Format. Running the candidate has to reject one that exits non-zero.
	if _, err := exec.LookPath("false"); err != nil {
		t.Skip("no `false` binary to stand in for a broken launcher")
	}

	formatter := &Formatter{argv: []string{"false"}}
	if _, err := formatter.version(); err == nil {
		t.Error("version() = nil error for a launcher that exits non-zero")
	}
}

func TestVersionRejectsMissingBinary(t *testing.T) {
	formatter := &Formatter{argv: []string{"definitely-not-a-real-binary-xyz"}}
	if _, err := formatter.version(); err == nil {
		t.Error("version() = nil error for a missing binary")
	}
}
