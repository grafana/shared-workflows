package main

import (
	"fmt"
	"strings"
)

// Section is one of the two generated blocks in a doc file.
type Section string

const (
	SectionInputs  Section = "INPUTS"
	SectionOutputs Section = "OUTPUTS"
)

// Heading is the markdown heading this section lives under.
func (s Section) Heading() string {
	switch s {
	case SectionOutputs:
		return "## Outputs"
	default:
		return "## Inputs"
	}
}

func (s Section) beginMarker() string { return fmt.Sprintf("<!-- BEGIN_%s -->", s) }
func (s Section) endMarker() string   { return fmt.Sprintf("<!-- END_%s -->", s) }

// emptyPlaceholder stands in when a section's markers exist but the action or
// workflow declares nothing, so the block is never left as blank markdown.
const emptyPlaceholder = "_None._"

// Inject rewrites the generated block for one section of doc and returns the
// updated document.
//
// It handles three cases, in order of preference:
//
//  1. The section's markers are already present -- the content between them is
//     replaced. This is the only path taken for files already checked in, and it
//     is what lets authors keep hand-written prose inside the section (pragmas,
//     footnotes) without the generator clobbering it.
//  2. The heading is present but the markers are not -- markers are inserted
//     directly beneath the heading and any table that immediately followed it is
//     dropped. This bootstraps a doc that was previously maintained by hand.
//  3. Neither is present -- a new section is appended to the end of the file.
//     This is the path a newly added action takes before anyone has written docs.
//
// A section with no entries and no existing markers is skipped entirely rather
// than adding an empty heading.
func Inject(doc string, section Section, table string) (string, error) {
	if err := validateMarkers(doc, section); err != nil {
		return "", err
	}

	body := table
	if body == "" {
		body = emptyPlaceholder
	}

	// The blank lines around the table are required: prettier inserts them
	// between an HTML comment and an adjacent table, so omitting them here would
	// make the pre-commit hook rewrite every generated block.
	block := strings.Join([]string{
		section.beginMarker(),
		"",
		strings.TrimRight(body, "\n"),
		"",
		section.endMarker(),
	}, "\n")

	if updated, ok := replaceBetweenMarkers(doc, section, block); ok {
		return updated, nil
	}
	updated, ok, err := insertUnderHeading(doc, section, block)
	if err != nil {
		return "", err
	}
	if ok {
		return updated, nil
	}
	if table == "" {
		return doc, nil
	}
	return appendSection(doc, section, block), nil
}

// validateMarkers rejects marker arrangements that cannot be updated safely.
//
// Each of these is a mistake rather than a state worth tolerating: left alone
// they produce a doc that is stable on regeneration, so the CI drift check stays
// green while the content silently rots.
func validateMarkers(doc string, section Section) error {
	begins := strings.Count(doc, section.beginMarker())
	ends := strings.Count(doc, section.endMarker())

	switch {
	case begins == 0 && ends == 0:
		return nil
	case begins > 1 || ends > 1:
		// Only the first pair would ever be rewritten; later ones keep stale
		// content forever.
		return fmt.Errorf("found %d %s and %d %s markers, want at most one of each",
			begins, section.beginMarker(), ends, section.endMarker())
	case begins != ends:
		return fmt.Errorf("found %d %s and %d %s markers; they must be paired",
			begins, section.beginMarker(), ends, section.endMarker())
	case strings.Index(doc, section.endMarker()) < strings.Index(doc, section.beginMarker()):
		return fmt.Errorf("%s appears before %s", section.endMarker(), section.beginMarker())
	}
	return nil
}

// replaceBetweenMarkers swaps out an existing marker block, reporting whether
// the markers were found. validateMarkers has already established that there is
// exactly one correctly ordered pair.
func replaceBetweenMarkers(doc string, section Section, block string) (string, bool) {
	begin := strings.Index(doc, section.beginMarker())
	if begin == -1 {
		return doc, false
	}
	end := strings.Index(doc[begin:], section.endMarker())
	if end == -1 {
		return doc, false
	}
	end += begin + len(section.endMarker())
	return doc[:begin] + block + doc[end:], true
}

// insertUnderHeading places a fresh marker block immediately below the section's
// heading, removing a hand-maintained table if one directly followed it.
func insertUnderHeading(doc string, section Section, block string) (string, bool, error) {
	lines := strings.Split(doc, "\n")
	headingIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == section.Heading() {
			headingIdx = i
			break
		}
	}
	if headingIdx == -1 {
		return doc, false, nil
	}

	// Skip blank lines, then consume a contiguous run of table rows.
	cursor := headingIdx + 1
	for cursor < len(lines) && strings.TrimSpace(lines[cursor]) == "" {
		cursor++
	}
	tableEnd := cursor
	for tableEnd < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[tableEnd]), "|") {
		tableEnd++
	}

	// Nothing directly below the heading looks like a table, but the section may
	// still hold a hand-written one further down, past a markdownlint pragma or a
	// paragraph. Inserting above it would leave two tables describing the same
	// inputs -- and because the result is idempotent, the drift check would stay
	// green over it forever. Refuse instead, and let a human place the markers.
	if tableEnd == cursor && sectionContainsTable(lines, cursor) {
		return "", false, fmt.Errorf(
			"%s has a hand-written table that is not directly under the heading; "+
				"add %s and %s around it by hand so the generator knows what to replace",
			section.Heading(), section.beginMarker(), section.endMarker())
	}

	rebuilt := make([]string, 0, len(lines)+8)
	rebuilt = append(rebuilt, lines[:headingIdx+1]...)
	rebuilt = append(rebuilt, "", block)
	rest := lines[tableEnd:]
	if len(rest) > 0 && strings.TrimSpace(rest[0]) != "" {
		rebuilt = append(rebuilt, "")
	}
	rebuilt = append(rebuilt, rest...)
	return strings.Join(rebuilt, "\n"), true, nil
}

// sectionContainsTable reports whether a markdown table appears between start
// and the next heading.
func sectionContainsTable(lines []string, start int) bool {
	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "#") {
			return false
		}
		if strings.HasPrefix(trimmed, "|") {
			return true
		}
	}
	return false
}

// appendSection adds the heading and its marker block to the end of the file.
func appendSection(doc string, section Section, block string) string {
	trimmed := strings.TrimRight(doc, "\n")
	return trimmed + "\n\n" + section.Heading() + "\n\n" + block + "\n"
}

// RenderDoc applies both sections of spec to doc, then hands the result to
// prettier so the file matches what the pre-commit hook would produce.
//
// filename is the path the document will be written to; prettier needs it to
// resolve config and infer the parser.
func RenderDoc(doc string, spec *Spec, filename string, formatter *Formatter) (string, error) {
	updated, err := Inject(doc, SectionInputs, RenderInputs(spec.Inputs))
	if err != nil {
		return "", fmt.Errorf("%s: %w", filename, err)
	}
	updated, err = Inject(updated, SectionOutputs, RenderOutputs(spec.Outputs))
	if err != nil {
		return "", fmt.Errorf("%s: %w", filename, err)
	}
	return formatter.Format(updated, filename)
}
