package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

func TestUpdateRelativeLinks(t *testing.T) {
	logger := slogt.New(t)
	tests := map[string]struct {
		testDirectory       map[string]string
		fileUnderTest       string
		expectedFileContent string
	}{
		"rewrites-outside-link": {
			testDirectory: map[string]string{
				"README.md":      "outside file",
				"docs/README.md": "Hello world [inside](./README.md) [outside](../README.md)",
			},
			fileUnderTest:       "docs/README.md",
			expectedFileContent: "Hello world [inside](./README.md) [outside](https://github.com/grafana/dummy/blob/main/README.md)\n",
		},
		"ignores-external-links": {
			testDirectory: map[string]string{
				"docs/README.md": "Hello world [external](https://example.org)",
			},
			fileUnderTest:       "docs/README.md",
			expectedFileContent: "Hello world [external](https://example.org)",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			rootDir := t.TempDir()
			createTestDirectory(t, rootDir, test.testDirectory)
			ctrl := controller{
				rootDirectory:     rootDir,
				logger:            logger,
				defaultBranch:     "main",
				repoURL:           "https://github.com/grafana/dummy",
				docsRootDirectory: filepath.Join(rootDir, "docs"),
			}
			testFilePath := filepath.Join(rootDir, test.fileUnderTest)
			require.NoError(t, ctrl.updateRelativeLinks(context.Background(), filepath.Join(rootDir, test.fileUnderTest)))
			updatedContent, err := os.ReadFile(testFilePath)
			require.NoError(t, err)
			require.Equal(t, test.expectedFileContent, string(updatedContent))
		})
	}
}

func TestGetDocsRoot(t *testing.T) {
	logger := slogt.New(t)
	tests := map[string]struct {
		expectError             bool
		expectedRelativeDocsDir string
		expectedAbsoluteDocsDir string
		rootDirectory           map[string]string
	}{
		"no-mkdocs-yml": {
			expectError:   true,
			rootDirectory: map[string]string{},
		},
		"mkdocs-yml-relative-docs-dir": {
			expectError: false,
			rootDirectory: map[string]string{
				"mkdocs.yml": `docs_dir: "docs"`,
			},
			expectedRelativeDocsDir: "docs",
		},
		"mkdocs-yml-absolute-docs-dir": {
			expectError: false,
			rootDirectory: map[string]string{
				"mkdocs.yml": `docs_dir: "/tmp/docs"`,
			},
			expectedAbsoluteDocsDir: "/tmp/docs",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			rootDir := t.TempDir()
			createTestDirectory(t, rootDir, test.rootDirectory)
			ctrl := controller{
				rootDirectory: rootDir,
				logger:        logger,
				defaultBranch: "main",
				repoURL:       "https://github.com/grafana/dummy",
			}
			docsDir, err := ctrl.getDocsDir(context.Background(), rootDir)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if test.expectedAbsoluteDocsDir != "" {
				require.Equal(t, test.expectedAbsoluteDocsDir, docsDir)
			}
			if test.expectedRelativeDocsDir != "" {
				require.Equal(t, filepath.Join(rootDir, test.expectedRelativeDocsDir), docsDir)
			}
		})
	}
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
