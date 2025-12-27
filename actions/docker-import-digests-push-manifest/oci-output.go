package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

type Base struct {
	Name      string `json:"name"`
	MediaType string `json:"mediaType"`
}

type Tag struct {
	Base
	Digest    string     `json:"digest"`
	Manifests []Manifest `json:"manifests"`
}

type Manifest struct {
	Base
	Platform    string       `json:"platform"`
	Annotations []Annotation `json:"annotations"`
}

type Annotation struct {
	annotation       string `json:"key"`
	annotation_value string `json:"value"`
}

type Output struct {
	IndexDigest string     `json:"indexDigest"`
	Manifests   []Manifest `json:"manifests"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var out Output
	var currentPlatform string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "Digest:") && out.IndexDigest == "":
			out.IndexDigest = strings.TrimSpace(strings.TrimPrefix(line, "Digest:"))

		case strings.HasPrefix(line, "Platform:"):
			currentPlatform = strings.TrimSpace(strings.TrimPrefix(line, "Platform:"))

		case strings.HasPrefix(line, "Digest:") && currentPlatform != "":
			out.Manifests = append(out.Manifests, Manifest{
				Platform: currentPlatform,
				Digest:   strings.TrimSpace(strings.TrimPrefix(line, "Digest:")),
			})
			currentPlatform = ""
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
