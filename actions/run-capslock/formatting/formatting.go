package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Read stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	text := string(data)
	text = strings.ReplaceAll(text, "New packages", "END\nNew packages")
	text += "END"
	// Match commits + capabilities + everything after
	mainRe := regexp.MustCompile(`(?s)Added (?P<numcaps>\d+) new .*?:\n(?P<capabilities>.*?)END\nNew packages`)
	match := mainRe.FindStringSubmatch(text)
	if match == nil {
		fmt.Println("No match found")
		return
	}

	result := make(map[string]string)
	for i, name := range mainRe.SubexpNames() {
		if i > 0 && name != "" {
			result[name] = match[i]
		}
	}

	// Regex that finds each section *with its content until the next section or EOF*
	// We capture the section header (capName) and the following block
	capSectionRe := regexp.MustCompile(`(?s)New packages in call paths to capability (CAPABILITY_[A-Z0-9_]+):\n\n(.*?)(?:END)`)
	capMatches := capSectionRe.FindAllStringSubmatch(text, -1)

	// ---- Markdown output ----
	scope := os.Getenv("SCOPE")
	if scope != "" {
		scope = fmt.Sprintf(" for %s", scope)
	}
	fmt.Printf("# Capability Comparison Report%s\n\n", scope)

	fmt.Printf("## ⚙️ Added Capabilities (%s)\n", result["numcaps"])
	fmt.Println("```")
	fmt.Println(strings.TrimSpace(strings.ReplaceAll(result["capabilities"], "          ", "")))
	fmt.Println("```")

	// Loop over all capability sections
	packagesRegx := regexp.MustCompile(`(?s)Package(?:s|) (?P<packages>.*?) (?:has|have).*?:(.*?)END`)
	for _, m := range capMatches {

		capName := m[1]
		fmt.Printf("## ⚠️ New Packages for %s\n", capName)
		capContent := strings.TrimSpace(m[2])
		capContent = strings.ReplaceAll(capContent, "Package", "END\nPackage")
		capContent += "END"
		matches := packagesRegx.FindAllStringSubmatch(capContent, -1)

		for _, m := range matches {
			fmt.Printf("<details><summary>📦 %s</summary>\n\n", m[1])
			r := regexp.MustCompile(`(?m)(?s)^ *?\w(\S*)\n`)
			mat := r.FindAllString(m[2], -1)
			m[2] += "\n"
			fmt.Println("```")
			fmt.Printf("- Packages tree:\n")
			for _, i := range mat {
				fmt.Printf("		%s\n", strings.TrimSpace(i))
			}
			r2 := regexp.MustCompile(`(?m)(?s)^((?:\S|\s+\().*?)\n`)
			mat2 := r2.FindAllString(m[2], -1)
			fmt.Printf("- Path:\n")
			for _, i := range mat2 {
				fmt.Printf("		%s", i)
			}
			fmt.Println("```")
			fmt.Printf("</details>\n\n")
		}
	}
}
