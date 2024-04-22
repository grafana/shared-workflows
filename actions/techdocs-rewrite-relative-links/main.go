package main

import (
	"bytes"
	"context"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/lmittmann/tint"
	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/urfave/cli/v2"
	"github.com/willabides/actionslog"
	"github.com/willabides/actionslog/human"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

func main() {
	app := cli.App{
		Name: "techdocs-rewrite-relative-links",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "root-dir",
				Usage:    "Path to the directory where the mkdocs.yml file is located",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "repo-url",
				Usage:    "Full URL of the repository on GitHub",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "default-branch",
				Usage:    "Name of the default branch",
				Required: true,
			},
		},
		Action: func(cliCtx *cli.Context) error {
			repoURL := cliCtx.String("repo-url")
			defaultBranch := cliCtx.String("default-branch")
			rootDir := cliCtx.String("root-dir")

			ctx := cliCtx.Context
			var logger *slog.Logger
			if os.Getenv("GITHUB_ACTIONS") == "true" {
				handler := &human.Handler{}
				logger = slog.New(&actionslog.Wrapper{
					Handler: handler.WithOutput,
				})
			} else {
				if term.IsTerminal(int(os.Stderr.Fd())) {
					logger = slog.New(
						tint.NewHandler(os.Stderr, nil),
					)
				} else {
					logger = slog.New(
						slog.NewTextHandler(os.Stderr, nil),
					)
				}
			}
			ctrl := controller{
				logger:        logger,
				rootDirectory: rootDir,
				repoURL:       repoURL,
				defaultBranch: defaultBranch,
			}
			return ctrl.run(ctx)
		},
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	app.RunContext(ctx, os.Args)
}

type controller struct {
	logger        *slog.Logger
	defaultBranch string
	repoURL       string
	rootDirectory string

	// Private: This will be set during the run
	docsRootDirectory string
}

func (ctrl *controller) run(ctx context.Context) error {
	docsRootDirectory, err := ctrl.getDocsDir(ctx, ctrl.rootDirectory)
	ctrl.docsRootDirectory = docsRootDirectory
	if err != nil {
		return err
	}
	return filepath.WalkDir(docsRootDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			if err := ctrl.updateRelativeLinks(ctx, path); err != nil {
				return err
			}
		}
		return nil
	})
}

func (ctrl *controller) updateRelativeLinks(ctx context.Context, path string) error {
	logger := ctrl.logger.With(slog.String("path", path))
	renderer := markdown.NewRenderer()
	markdown := goldmark.New(goldmark.WithRenderer(renderer))
	transformer := &relativeLinkASTTransformer{
		ctx:               ctx,
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
		// Only update the original file if there were actual changes made:
		if err := os.WriteFile(path, out.Bytes(), 0600); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *controller) getDocsDir(ctx context.Context, root string) (string, error) {
	cfg := mkdocsYaml{}
	content, err := os.ReadFile(filepath.Join(root, "mkdocs.yml"))
	if err != nil {
		return "", err
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return "", err
	}
	if filepath.IsAbs(cfg.DocsDir) {
		return cfg.DocsDir, nil
	}
	return filepath.Join(root, cfg.DocsDir), nil
}

type mkdocsYaml struct {
	DocsDir string `yaml:"docs_dir"`
}

type relativeLinkASTTransformer struct {
	ctx               context.Context
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
					transformer.logger.InfoContext(transformer.ctx, "rewriting path", slog.String("old-dest", dst), slog.String("new-dest", newDest))
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
