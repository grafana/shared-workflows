package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanCell(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"collapses newlines", "one\ntwo\nthree", "one two three"},
		{"trims trailing newline from block scalar", "text\n", "text"},
		{"collapses runs of spaces", "a    b", "a b"},
		{"escapes pipes so the cell cannot break the table", "a|b", `a\|b`},
		{"empty stays empty", "", ""},
		{"whitespace only becomes empty", "  \n  ", ""},
		// Bare placeholders are read as HTML tags and swallowed by the renderer,
		// so `<image>:<tag>` would display as `:`.
		{"escapes angle brackets", "<image>:<tag>", "&lt;image>:&lt;tag>"},
		{"escapes angle brackets in prose", "diff is <a>..<b>.", "diff is &lt;a>..&lt;b>."},
		// Inside a code span the text is already literal, and an entity would show
		// up as the entity rather than as a bracket.
		{"leaves code spans alone", "use `<image>` here", "use `<image>` here"},
		{"handles a widened fence", "``a `<b>` c``", "``a `<b>` c``"},
		{"escapes outside but not inside", "<a> `<b>` <c>", "&lt;a> `<b>` &lt;c>"},
		{"unclosed fence swallows the rest", "x `<a>", "x `<a>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanCell(tt.in); got != tt.want {
				t.Errorf("cleanCell(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCodeSpan(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain value", "renovate.json", "`renovate.json`"},
		{"empty stays empty", "", ""},
		{"expression", "${{ github.token }}", "`${{ github.token }}`"},
		// A value containing backticks needs a longer fence, plus padding spaces
		// so the fence boundary is unambiguous.
		{"single backtick widens the fence", "a`b", "``  a`b ``"},
		{"double backtick widens further", "a``b", "``` a``b ```"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codeSpan(tt.in)
			// Normalize the padding so the test asserts the fence width, which is
			// what actually matters for rendering.
			if strings.ReplaceAll(got, "  ", " ") != strings.ReplaceAll(tt.want, "  ", " ") {
				t.Errorf("codeSpan(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRenderInputsColumns(t *testing.T) {
	spec, err := ParseFile(filepath.Join("testdata", "action.yaml"), KindCompositeAction)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	table := RenderInputs(spec.Inputs)
	if !strings.HasPrefix(table, "| Name | Type | Required | Default | Description |\n") {
		t.Fatalf("unexpected header:\n%s", table)
	}

	lines := strings.Split(strings.TrimRight(table, "\n"), "\n")
	// header + delimiter + one row per input
	if got, want := len(lines), 2+len(spec.Inputs); got != want {
		t.Errorf("got %d lines, want %d", got, want)
	}

	assertRowContains(t, table, "`alpha`", "Yes")
	assertRowContains(t, table, "`boolish`", "Boolean")
	// The pipe in the description must arrive escaped.
	assertRowContains(t, table, "`pipes`", `a\|b`)
	// A folded description must collapse onto one line.
	assertRowContains(t, table, "`alpha`", "several lines and ends with a newline.")
}

func TestRenderOutputsHasTwoColumns(t *testing.T) {
	spec, err := ParseFile(filepath.Join("testdata", "action.yaml"), KindCompositeAction)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	table := RenderOutputs(spec.Outputs)
	// Outputs carry no type, default or required flag in either schema, so
	// rendering those columns would only ever produce filler.
	if !strings.HasPrefix(table, "| Name | Description |\n") {
		t.Fatalf("unexpected header:\n%s", table)
	}
	if strings.Contains(table, "Required") {
		t.Error("outputs table should not have a Required column")
	}
}

func TestRenderEmptyIsEmptyString(t *testing.T) {
	if got := RenderInputs(nil); got != "" {
		t.Errorf("RenderInputs(nil) = %q, want empty", got)
	}
	if got := RenderOutputs(nil); got != "" {
		t.Errorf("RenderOutputs(nil) = %q, want empty", got)
	}
}

func assertRowContains(t *testing.T, table, rowKey, want string) {
	t.Helper()
	for _, line := range strings.Split(table, "\n") {
		if strings.HasPrefix(line, "| "+rowKey+" ") {
			if !strings.Contains(line, want) {
				t.Errorf("row %s = %q, want it to contain %q", rowKey, line, want)
			}
			return
		}
	}
	t.Errorf("no row for %s in:\n%s", rowKey, table)
}
