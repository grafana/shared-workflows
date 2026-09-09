package main

import (
	"strings"
)

// RenderInputs formats inputs as a GitHub-flavored markdown table.
//
// Required and Default are generated from the YAML rather than left to prose so
// they cannot fall out of step with the action -- the same reason the tool exists
// at all.
//
// Cells are emitted unpadded; prettier owns the final layout. Do not be tempted
// to align the columns here: prettier does not only pad cells, it also rewrites
// their content (collapsing a ```x``` code fence to `x`, for example), so any
// alignment computed before that rewrite is wrong. Format() runs the finished
// document through prettier instead.
//
// Returns an empty string when there is nothing to render.
func RenderInputs(entries []IOEntry) string {
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []string{
			codeSpan(e.Name),
			e.DisplayType(),
			yesNo(e.Required),
			renderDefault(e),
			cleanCell(e.Description),
		})
	}
	return renderTable([]string{"Name", "Type", "Required", "Default", "Description"}, rows)
}

// RenderOutputs formats outputs as a markdown table.
//
// Outputs are only a name, a description and an expression: neither composite
// actions nor reusable workflows give them a type, a default or a required flag,
// so there are only two columns worth rendering.
func RenderOutputs(entries []IOEntry) string {
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []string{
			codeSpan(e.Name),
			cleanCell(e.Description),
		})
	}
	return renderTable([]string{"Name", "Description"}, rows)
}

func renderTable(header []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("| " + strings.Join(header, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat(" --- |", len(header)) + "\n")
	for _, row := range rows {
		b.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}
	return b.String()
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// renderDefault shows the default as a code span, or leaves the cell empty when
// the input has none. An explicitly empty default renders empty too: to a reader
// "no default" and "defaults to nothing" are the same statement.
func renderDefault(e IOEntry) string {
	value := cleanCell(e.Default)
	if value == "" {
		return ""
	}
	return codeSpan(value)
}

// codeSpan wraps s in backticks, widening the fence if s itself contains a run
// of backticks, per the CommonMark rule for code spans.
func codeSpan(s string) string {
	if s == "" {
		return ""
	}
	longest := 0
	current := 0
	for _, r := range s {
		if r == '`' {
			current++
			if current > longest {
				longest = current
			}
			continue
		}
		current = 0
	}
	if longest == 0 {
		return "`" + s + "`"
	}
	fence := strings.Repeat("`", longest+1)
	// A code span whose content starts or ends with a backtick needs padding
	// spaces so the fence is unambiguous.
	return fence + " " + s + " " + fence
}

// cleanCell flattens a YAML scalar into a single markdown table cell.
//
// Descriptions and defaults are routinely written as folded or literal block
// scalars, so they arrive with newlines and a trailing newline attached.
// Collapsing every run of whitespace to one space keeps the cell on one line,
// escaping pipes stops the value from breaking out of the table, and escaping
// angle brackets stops a placeholder from being eaten as HTML.
func cleanCell(s string) string {
	escaped := strings.ReplaceAll(s, "|", `\|`)
	return strings.Join(strings.Fields(escapeAngleBrackets(escaped)), " ")
}

// escapeAngleBrackets replaces `<` with `&lt;` outside code spans.
//
// Descriptions describe formats using bare placeholders -- `<image>:<tag>@<digest>`,
// `<base-ref>..<commit-sha>`, `gh project view <number> --owner <org>`. Markdown
// treats those as raw HTML tags, so the renderer swallows them and the reader
// sees `:@` or `..`. Escaping makes them display literally.
//
// Text inside a code span is skipped: backticks already render their contents
// literally, and an `&lt;` written there would show up as the entity itself
// rather than as `<`.
func escapeAngleBrackets(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if s[i] == '`' {
			// Copy the whole code span verbatim: an opening run of N backticks is
			// closed by the next run of exactly N.
			fence := runLength(s, i, '`')
			end := findClosingFence(s, i+fence, fence)
			b.WriteString(s[i:end])
			i = end
			continue
		}
		if s[i] == '<' {
			b.WriteString("&lt;")
			i++
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// runLength counts consecutive occurrences of c starting at i.
func runLength(s string, i int, c byte) int {
	n := 0
	for i+n < len(s) && s[i+n] == c {
		n++
	}
	return n
}

// findClosingFence returns the index just past a backtick run of exactly want,
// searching from i. If the span is never closed, the rest of the string is
// treated as part of it, matching how a renderer would give up on the fence.
func findClosingFence(s string, i, want int) int {
	for i < len(s) {
		if s[i] != '`' {
			i++
			continue
		}
		n := runLength(s, i, '`')
		if n == want {
			return i + n
		}
		i += n
	}
	return len(s)
}
