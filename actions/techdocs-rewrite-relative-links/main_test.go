package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestUpdateRelativeLinks(t *testing.T) {
	logger := slogt.New(t)
	tests := map[string]struct {
		testDirectory       map[string]string
		testDirectorySetup  func(afero.Fs, string)
		fileUnderTest       string
		expectedFileContent string
		dryRun              bool
	}{
		"rewrites-outside-link": {
			testDirectorySetup: func(fs afero.Fs, rootDir string) {
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "README.md"), []byte("outside file"), 0600)
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "docs/README.md"), []byte("Hello world [inside](./README.md) [outside](../README.md)"), 0600)
			},
			fileUnderTest:       "docs/README.md",
			expectedFileContent: "Hello world [inside](./README.md) [outside](https://github.com/grafana/dummy/blob/main/README.md)\n",
		},
		"rewrites-outside-link-with-hashname": {
			testDirectorySetup: func(fs afero.Fs, rootDir string) {
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "README.md"), []byte("outside file"), 0600)
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "docs/README.md"), []byte("Hello world [inside](./README.md) [outside](../README.md#somewhere)"), 0600)
			},
			fileUnderTest:       "docs/README.md",
			expectedFileContent: "Hello world [inside](./README.md) [outside](https://github.com/grafana/dummy/blob/main/README.md#somewhere)\n",
		},
		"ignores-external-links": {
			testDirectorySetup: func(fs afero.Fs, rootDir string) {
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "docs/README.md"), []byte("Hello world [external](https://example.org)"), 0600)
			},
			fileUnderTest:       "docs/README.md",
			expectedFileContent: "Hello world [external](https://example.org)",
		},
		"dry-run-does-not-modify": {
			testDirectorySetup: func(fs afero.Fs, rootDir string) {
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "README.md"), []byte("outside file"), 0600)
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "docs/README.md"), []byte("Hello world [inside](./README.md) [outside](../README.md)"), 0600)
			},
			fileUnderTest:       "docs/README.md",
			expectedFileContent: "Hello world [inside](./README.md) [outside](../README.md)",
			dryRun:              true,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			filesys := afero.NewMemMapFs()
			rootDir := "/root"
			if test.testDirectorySetup != nil {
				test.testDirectorySetup(filesys, rootDir)
			}
			ctrl := controller{
				filesys:           filesys,
				dryRun:            test.dryRun,
				rootDirectory:     rootDir,
				logger:            logger,
				defaultBranch:     "main",
				repoURL:           "https://github.com/grafana/dummy",
				docsRootDirectory: filepath.Join(rootDir, "docs"),
			}
			testFilePath := filepath.Join(rootDir, test.fileUnderTest)
			require.NoError(t, ctrl.updateRelativeLinks(context.Background(), filepath.Join(rootDir, test.fileUnderTest)))
			updatedContent, err := afero.ReadFile(filesys, testFilePath)
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
		rootDirectorySetup      func(afero.Fs, string)
	}{
		"no-mkdocs-yml": {
			expectError: true,
		},
		"mkdocs-yml-relative-docs-dir": {
			expectError: false,
			rootDirectorySetup: func(fs afero.Fs, rootDir string) {
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "mkdocs.yml"), []byte(`docs_dir: "docs"`), 0600)
			},
			expectedRelativeDocsDir: "docs",
		},
		"mkdocs-yml-absolute-docs-dir": {
			expectError: false,
			rootDirectorySetup: func(fs afero.Fs, rootDir string) {
				_ = afero.WriteFile(fs, filepath.Join(rootDir, "mkdocs.yml"), []byte(`docs_dir: "/tmp/docs"`), 0600)
			},
			expectedAbsoluteDocsDir: "/tmp/docs",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			rootDir := "/root"
			filesys := afero.NewMemMapFs()
			if test.rootDirectorySetup != nil {
				test.rootDirectorySetup(filesys, rootDir)
			}
			ctrl := controller{
				filesys:       filesys,
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
