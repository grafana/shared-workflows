package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/go-codeowners"
	"sigs.k8s.io/yaml"
)

type ActionDef struct {
	Slug        string `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
	YAMLPath    string `json:"-"`
}

type CatalogInfo struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   CatalogInfoMetadata `json:"metadata"`
	Spec       CatalogInfoSpec     `json:"spec"`
}

type CatalogInfoMetadata struct {
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Annotations map[string]string `json:"annotations"`
	Links       []Link            `json:"links,omitempty"`
}

type CatalogInfoSpec struct {
	Type           string `json:"type"`
	Owner          string `json:"owner"`
	Lifecycle      string `json:"lifecycle"`
	SubcomponentOf string `json:"subcomponentOf,omitempty"`
}

type Link struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func main() {
	var rootDir string
	var debug bool
	var outputPath string
	flag.StringVar(&rootDir, "root-dir", ".", "Root directory of the repository")
	flag.BoolVar(&debug, "debug", false, "Debug level logging")
	flag.StringVar(&outputPath, "output-path", "-", "Output to path")
	flag.Parse()

	var level slog.Leveler
	if debug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))

	localFS := os.DirFS(rootDir)
	entries, err := fs.ReadDir(localFS, "actions")
	if err != nil {
		logger.Error("loading actions failed", "err", err.Error())
		os.Exit(1)
	}
	actions := make([]*ActionDef, 0, 10)
	for _, actionDir := range entries {
		actionDef := ActionDef{}
		if !actionDir.IsDir() {
			continue
		}
		fileNameCandidates := []string{"action.yml", "action.yaml"}
		for _, candidate := range fileNameCandidates {
			fullPath := filepath.Join("actions", actionDir.Name(), candidate)
			if _, err := fs.Stat(localFS, fullPath); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				logger.Error("loading action failed", "err", err.Error())
				os.Exit(1)
			}
			data, err := fs.ReadFile(localFS, fullPath)
			if err != nil {
				logger.Error("loading action failed", "err", err.Error())
				os.Exit(1)
			}
			if err := yaml.Unmarshal(data, &actionDef); err != nil {
				logger.Error("loading action failed", "err", err.Error())
				os.Exit(1)
			}
			actionDef.Slug = actionDir.Name()
			actionDef.YAMLPath = fullPath
			logger.Debug("action found", "action", actionDef.Slug)
			actions = append(actions, &actionDef)
			break
		}
	}

	owners, err := codeowners.FromFileWithFS(localFS, ".")
	if err != nil {
		logger.Error("parsing codeowners failed", "err", err.Error())
		os.Exit(1)
	}

	output := bytes.Buffer{}

	if err := writeYAML(&output, CatalogInfo{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       "Component",
		Metadata: CatalogInfoMetadata{
			Name:  "shared-workflows",
			Title: "shared-workflows",
			Annotations: map[string]string{
				"github.com/project-slug": "grafana/shared-workflows",
			},
		},
		Spec: CatalogInfoSpec{
			Lifecycle: "production",
			Type:      "library",
			Owner:     fmt.Sprintf("group:%s", "platform-productivity"),
		},
	}); err != nil {
		logger.Error("generating YAML failed", "err", err.Error())
		os.Exit(1)
	}

	for _, action := range actions {
		actionOwners := owners.Owners(action.YAMLPath)
		primaryOwner := ""
		for _, owner := range actionOwners {
			if !strings.HasPrefix(owner, "@grafana/") {
				continue
			}
			primaryOwner = strings.TrimPrefix(owner, "@grafana/")
			break
		}
		info := CatalogInfo{
			APIVersion: "backstage.io/v1alpha1",
			Kind:       "Component",
			Metadata: CatalogInfoMetadata{
				Name:        action.Slug,
				Title:       action.Slug,
				Description: action.Description,
				Annotations: map[string]string{
					"github.com/project-slug": "grafana/shared-workflows",
				},
				Links: []Link{
					{
						URL:   fmt.Sprintf("https://github.com/grafana/shared-workflows/blob/main/actions/%s/README.md", action.Slug),
						Title: "README",
					},
				},
			},
			Spec: CatalogInfoSpec{
				Lifecycle:      "production",
				Type:           "github-action",
				Owner:          fmt.Sprintf("group:%s", primaryOwner),
				SubcomponentOf: "component:shared-workflows",
			},
		}
		err := writeYAML(&output, info)
		if err != nil {
			logger.Error("generating YAML failed", "err", err.Error())
			os.Exit(1)
		}
	}

	if outputPath == "-" {
		fmt.Println(output.String())
		return
	}
	if err := os.WriteFile(outputPath, output.Bytes(), 0644); err != nil {
		logger.Error("writing to output failed", "err", err.Error())
		os.Exit(1)
	}
}

func writeYAML(output *bytes.Buffer, info CatalogInfo) error {
	infoYaml, err := yaml.Marshal(info)
	if err != nil {
		return err
	}
	output.WriteString("---\n")
	output.Write(infoYaml)
	output.WriteRune('\n')
	return nil
}
