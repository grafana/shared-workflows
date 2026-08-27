package main

import (
	"strings"
	"testing"
)

const table = "| Name | Description |\n| --- | --- |\n| `a` | x |\n"

func TestInjectReplacesBetweenExistingMarkers(t *testing.T) {
	doc := `# Thing

## Inputs

<!-- BEGIN_INPUTS -->

| Name | Description |
| --- | --- |
| ` + "`stale`" + ` | old |

<!-- END_INPUTS -->

Prose that must survive.
`

	got := mustInject(t, doc, SectionInputs, table)

	if strings.Contains(got, "stale") {
		t.Error("the old table should have been replaced")
	}
	if !strings.Contains(got, "| `a` | x |") {
		t.Errorf("new table missing:\n%s", got)
	}
	// The whole point of markers is that hand-written prose around them survives.
	if !strings.Contains(got, "Prose that must survive.") {
		t.Errorf("prose after the markers was lost:\n%s", got)
	}
	if strings.Count(got, "<!-- BEGIN_INPUTS -->") != 1 {
		t.Errorf("markers duplicated:\n%s", got)
	}
}

func TestInjectPreservesProseInsideSection(t *testing.T) {
	// aws-auth wraps its table in a markdownlint pragma. Injecting must not
	// disturb anything outside the markers, even within the same section.
	doc := `## Inputs

<!-- markdownlint-disable no-space-in-code -->

<!-- BEGIN_INPUTS -->

<!-- END_INPUTS -->

<!-- markdownlint-restore -->
`

	got := mustInject(t, doc, SectionInputs, table)

	for _, want := range []string{
		"<!-- markdownlint-disable no-space-in-code -->",
		"<!-- markdownlint-restore -->",
		"| `a` | x |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestInjectBootstrapsUnderExistingHeading(t *testing.T) {
	// A doc whose table was maintained by hand: markers get inserted and the
	// hand-written table is dropped.
	doc := `# Thing

## Inputs

| Name | Description |
| --- | --- |
| ` + "`handwritten`" + ` | old |

## Next
`

	got := mustInject(t, doc, SectionInputs, table)

	if strings.Contains(got, "handwritten") {
		t.Errorf("hand-written table should have been replaced:\n%s", got)
	}
	if !strings.Contains(got, "<!-- BEGIN_INPUTS -->") {
		t.Errorf("markers not inserted:\n%s", got)
	}
	if !strings.Contains(got, "## Next") {
		t.Errorf("following section was lost:\n%s", got)
	}
}

func TestInjectLeavesNonTableContentAlone(t *testing.T) {
	// trigger-argo-workflow documented its inputs as a bullet list. Silently
	// deleting prose would be worse than leaving a duplicate for a human to
	// resolve, so only a contiguous table is consumed.
	doc := `## Inputs

- ` + "`instance`" + `: some bullet documentation
`

	got := mustInject(t, doc, SectionInputs, table)

	if !strings.Contains(got, "some bullet documentation") {
		t.Errorf("bullet prose should be left for a human to remove:\n%s", got)
	}
	if !strings.Contains(got, "<!-- BEGIN_INPUTS -->") {
		t.Errorf("markers not inserted:\n%s", got)
	}
}

func TestInjectAppendsWhenNoHeadingExists(t *testing.T) {
	doc := "# Thing\n\nSome intro.\n"

	got := mustInject(t, doc, SectionInputs, table)

	if !strings.Contains(got, "## Inputs") {
		t.Errorf("heading not appended:\n%s", got)
	}
	if !strings.Contains(got, "| `a` | x |") {
		t.Errorf("table not appended:\n%s", got)
	}
	if !strings.HasPrefix(got, "# Thing\n\nSome intro.") {
		t.Errorf("original content changed:\n%s", got)
	}
}

func TestInjectSkipsEmptySectionWithNoMarkers(t *testing.T) {
	doc := "# Thing\n\nSome intro.\n"

	// An action with no outputs should not gain an empty "## Outputs" heading.
	if got := mustInject(t, doc, SectionOutputs, ""); got != doc {
		t.Errorf("document changed for an empty section:\n%s", got)
	}
}

func TestInjectEmptySectionWithMarkersGetsPlaceholder(t *testing.T) {
	// If the markers are already there, something has to go between them.
	doc := "## Outputs\n\n<!-- BEGIN_OUTPUTS -->\n\n| Name | Description |\n| --- | --- |\n| `gone` | x |\n\n<!-- END_OUTPUTS -->\n"

	got := mustInject(t, doc, SectionOutputs, "")

	if strings.Contains(got, "gone") {
		t.Errorf("stale table not removed:\n%s", got)
	}
	if !strings.Contains(got, emptyPlaceholder) {
		t.Errorf("expected %q placeholder in:\n%s", emptyPlaceholder, got)
	}
}

func TestInjectIsIdempotent(t *testing.T) {
	doc := "## Inputs\n\n<!-- BEGIN_INPUTS -->\n\n<!-- END_INPUTS -->\n"

	once := mustInject(t, doc, SectionInputs, table)
	twice := mustInject(t, once, SectionInputs, table)

	if once != twice {
		t.Errorf("second injection changed the document:\nfirst:\n%s\nsecond:\n%s", once, twice)
	}
}

func TestInjectBothSectionsIndependently(t *testing.T) {
	doc := "# Thing\n"

	got := mustInject(t, doc, SectionInputs, table)
	got = mustInject(t, got, SectionOutputs, table)

	for _, want := range []string{"## Inputs", "## Outputs", "BEGIN_INPUTS", "BEGIN_OUTPUTS"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	// Injecting outputs must not have disturbed the inputs block.
	if strings.Count(got, "BEGIN_INPUTS") != 1 {
		t.Errorf("inputs markers duplicated:\n%s", got)
	}
}

// mustInject calls Inject and fails the test on error, keeping the happy-path
// assertions focused on content.
func mustInject(t *testing.T, doc string, section Section, table string) string {
	t.Helper()
	got, err := Inject(doc, section, table)
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}
	return got
}

func TestInjectRefusesTableNotDirectlyUnderHeading(t *testing.T) {
	// A pragma between the heading and the table used to push the generated block
	// above it, leaving the hand-written table below forever. The result was
	// idempotent, so the CI drift check would never flag the duplication.
	doc := `## Inputs

<!-- markdownlint-disable no-space-in-code -->

| Name | Description |
| --- | --- |
| ` + "`handwritten`" + ` | old |

<!-- markdownlint-restore -->
`

	if _, err := Inject(doc, SectionInputs, table); err == nil {
		t.Fatal("expected an error rather than a silently duplicated table")
	}
}

func TestInjectRejectsMalformedMarkers(t *testing.T) {
	// Each of these leaves a doc that regenerates identically, so CI would stay
	// green while the content quietly went stale.
	tests := []struct {
		name string
		doc  string
	}{
		{
			"end before begin",
			"## Inputs\n\n<!-- END_INPUTS -->\n\n<!-- BEGIN_INPUTS -->\n",
		},
		{
			"begin with no end",
			"## Inputs\n\n<!-- BEGIN_INPUTS -->\n\nsome text\n",
		},
		{
			"end with no begin",
			"## Inputs\n\nsome text\n\n<!-- END_INPUTS -->\n",
		},
		{
			"two marker pairs",
			"## Inputs\n\n<!-- BEGIN_INPUTS -->\n\n<!-- END_INPUTS -->\n\n<!-- BEGIN_INPUTS -->\n\n<!-- END_INPUTS -->\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Inject(tt.doc, SectionInputs, table); err == nil {
				t.Errorf("expected an error for %s", tt.name)
			}
		})
	}
}
