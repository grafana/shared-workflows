package parser

import (
	"bufio"
	"bytes"
	"strings"
)

type Manifest struct {
	Name        string            `json:"name,omitempty"`
	MediaType   string            `json:"mediaType,omitempty"`
	Digest      string            `json:"digest"`
	Platform    string            `json:"platform,omitempty"`
	Kind        string            `json:"kind,omitempty"`     // image | attestation
	RefersTo    string            `json:"refersTo,omitempty"` // digest of image
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Report struct {
	Tag         string     `json:"tag,omitempty"`
	IndexDigest string     `json:"indexDigest,omitempty"`
	Manifests   []Manifest `json:"manifests"`
}

func extractDigestFromName(name string) string {
	if i := strings.LastIndex(name, "@sha256:"); i != -1 {
		return name[i+1:]
	}
	return ""
}

func ParseInspectOutput(input []byte) Report {
	scanner := bufio.NewScanner(bytes.NewReader(input))

	var out Report
	var current Manifest
	var inManifests bool
	var inAnnotations bool

	flush := func() {
		if current.Digest == "" {
			current = Manifest{}
			return
		}

		if current.Platform == "unknown/unknown" {
			current.Platform = ""
			current.Kind = "attestation"
		} else {
			current.Kind = "image"
		}

		if current.Kind == "attestation" && current.Annotations != nil {
			if d, ok := current.Annotations["vnd.docker.reference.digest"]; ok {
				current.RefersTo = d
			}
		}

		out.Manifests = append(out.Manifests, current)
		current = Manifest{}
		inAnnotations = false
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "Digest:") && !inManifests && out.IndexDigest == "":
			out.IndexDigest = strings.TrimSpace(strings.TrimPrefix(line, "Digest:"))

		case strings.HasPrefix(line, "Manifests"):
			inManifests = true

		case strings.HasPrefix(line, "Name:") && inManifests:
			flush()
			current = Manifest{}
			current.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
			current.Digest = extractDigestFromName(current.Name)

		case strings.HasPrefix(line, "MediaType:") && inManifests:
			current.MediaType = strings.TrimSpace(strings.TrimPrefix(line, "MediaType:"))

		case strings.HasPrefix(line, "Platform:") && inManifests:
			current.Platform = strings.TrimSpace(strings.TrimPrefix(line, "Platform:"))

		case strings.HasPrefix(line, "Annotations:") && inManifests:
			inAnnotations = true
			current.Annotations = map[string]string{}

		case inAnnotations && strings.Contains(line, ":"):
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			current.Annotations[key] = val
		}
	}

	flush()

	if out.Manifests == nil {
		out.Manifests = []Manifest{}
	}
	return out
}
