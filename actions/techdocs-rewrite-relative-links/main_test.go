package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateRelativeLinks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	t.Run("rewrites-outside-link", func(t *testing.T) {
		rootDir := t.TempDir()
		createTestDirectory(t, rootDir, map[string]string{
			"README.md":      "outside file",
			"docs/README.md": "Hello world [inside](./README.md) [outside](../README.md)",
		})
		ctrl := controller{
			rootDirectory:     rootDir,
			logger:            logger,
			defaultBranch:     "main",
			repoURL:           "https://github.com/grafana/dummy",
			docsRootDirectory: filepath.Join(rootDir, "docs"),
		}
		testFilePath := filepath.Join(rootDir, "docs", "README.md")
		require.NoError(t, ctrl.updateRelativeLinks(context.Background(), filepath.Join(rootDir, "docs/README.md")))
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err)
		require.Equal(t, "Hello world [inside](./README.md) [outside](https://github.com/grafana/dummy/blob/main/README.md)\n", string(updatedContent))
	})

	t.Run("ignores-external-links", func(t *testing.T) {
		rootDir := t.TempDir()
		createTestDirectory(t, rootDir, map[string]string{
			"docs/README.md": "Hello world [external](https://example.org)",
		})
		ctrl := controller{
			rootDirectory:     rootDir,
			logger:            logger,
			defaultBranch:     "main",
			repoURL:           "https://github.com/grafana/dummy",
			docsRootDirectory: filepath.Join(rootDir, "docs"),
		}
		testFilePath := filepath.Join(rootDir, "docs", "README.md")
		require.NoError(t, ctrl.updateRelativeLinks(context.Background(), filepath.Join(rootDir, "docs/README.md")))
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err)
		require.Equal(t, "Hello world [external](https://example.org)", string(updatedContent))
	})
}

func TestGetDocsRoot(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	t.Run("no-mkdocs-yml", func(t *testing.T) {
		rootDir := t.TempDir()
		createTestDirectory(t, rootDir, map[string]string{})
		ctrl := controller{
			rootDirectory: rootDir,
			logger:        logger,
			defaultBranch: "main",
			repoURL:       "https://github.com/grafana/dummy",
		}
		_, err := ctrl.getDocsDir(context.Background(), rootDir)
		require.Error(t, err)
	})
	t.Run("mkdocs-yml-relative-docs-dir", func(t *testing.T) {
		rootDir := t.TempDir()
		createTestDirectory(t, rootDir, map[string]string{
			"mkdocs.yml": `docs_dir: "docs"`,
		})
		ctrl := controller{
			rootDirectory: rootDir,
			logger:        logger,
			defaultBranch: "main",
			repoURL:       "https://github.com/grafana/dummy",
		}
		docsDir, err := ctrl.getDocsDir(context.Background(), rootDir)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(rootDir, "docs"), docsDir)
	})
	t.Run("mkdocs-yml-absolute-docs-dir", func(t *testing.T) {
		rootDir := t.TempDir()
		createTestDirectory(t, rootDir, map[string]string{
			"mkdocs.yml": `docs_dir: "/tmp/docs"`,
		})
		ctrl := controller{
			rootDirectory: rootDir,
			logger:        logger,
			defaultBranch: "main",
			repoURL:       "https://github.com/grafana/dummy",
		}
		docsDir, err := ctrl.getDocsDir(context.Background(), rootDir)
		require.NoError(t, err)
		require.Equal(t, filepath.Join("/tmp", "docs"), docsDir)
	})
}

func createTestDirectory(t *testing.T, rootDir string, content map[string]string) {
	t.Helper()
	for filename, filecontent := range content {
		dir := filepath.Dir(filename)
		fullPath := filepath.Join(rootDir, dir)
		require.NoError(t, os.MkdirAll(fullPath, 0700))
		fullFilePath := filepath.Join(rootDir, filename)
		require.NoError(t, os.WriteFile(fullFilePath, []byte(filecontent), 0600))
	}
}
