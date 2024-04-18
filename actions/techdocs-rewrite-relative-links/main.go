package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func main() {
	var repoURL string
	var defaultBranch string
	flag.StringVar(&repoURL, "repo-url", "https://github.com/grafana/deployment_tools", "Full URL of the repository on GitHub")
	flag.StringVar(&defaultBranch, "default-branch", "main", "Name of the default branch")
	flag.Parse()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	docsRoot := filepath.Join(flag.Arg(0), "docs")
	ctrl := controller{
		logger:            logger,
		docsRootDirectory: docsRoot,
		rootDirectory:     flag.Arg(0),
		repoURL:           repoURL,
		defaultBranch:     defaultBranch,
	}
	fmt.Println(filepath.WalkDir(docsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			if err := ctrl.updateRelativeLinks(ctx, path); err != nil {
				return err
			}
		}
		return nil
	}))
}

type controller struct {
	logger            *slog.Logger
	defaultBranch     string
	repoURL           string
	rootDirectory     string
	docsRootDirectory string
}

func (ctrl *controller) updateRelativeLinks(ctx context.Context, path string) error {
	logger := ctrl.logger.With(slog.String("path", path))
	renderer := markdown.NewRenderer()
	markdown := goldmark.New(goldmark.WithRenderer(renderer))
	transformer := &relativeLinkASTTransformer{
		logger:            logger,
		rootDirectory:     ctrl.rootDirectory,
		docsRootDirectory: ctrl.docsRootDirectory,
		path:              path,
		repoURL:           ctrl.repoURL,
		defaultBranch:     ctrl.defaultBranch,
	}
	markdown.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(transformer, 999)))
	source, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := bytes.Buffer{}
	if err := markdown.Convert(source, &out); err != nil {
		return err
	}
	if transformer.changed {
		if err := os.WriteFile(path, out.Bytes(), 0600); err != nil {
			return err
		}
	}
	return nil
}

type relativeLinkASTTransformer struct {
	logger            *slog.Logger
	rootDirectory     string
	docsRootDirectory string
	path              string
	repoURL           string
	defaultBranch     string
	changed           bool
}

func (transformer *relativeLinkASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := n.(*ast.Link); ok {
			// If the destination points somewhere outside of the root
			// directory, then we care:
			dst := string(link.Destination)
			if strings.HasPrefix(dst, "..") {
				absPath := filepath.Join(filepath.Dir(transformer.path), dst)
				destStat, err := os.Stat(absPath)
				if err != nil {
					return ast.WalkStop, err
				}
				if rel, _ := filepath.Rel(transformer.docsRootDirectory, absPath); strings.HasPrefix(rel, "..") {
					typeSegment := "tree"
					if !destStat.IsDir() {
						typeSegment = "blob"
					}
					newDest := strings.Replace(absPath, strings.TrimSuffix(transformer.rootDirectory, "/"), transformer.repoURL+"/"+typeSegment+"/"+transformer.defaultBranch, 1)
					transformer.logger.Info("rewriting path", slog.String("old-dest", dst), slog.String("new-dest", newDest))
					// If this is too aggressive, we should be able to
					// determine the line of the link and then just replace the
					// string in a separate run
					link.Destination = []byte(newDest)
					transformer.changed = true
				}
			}
		}
		return ast.WalkContinue, nil
	})
}
