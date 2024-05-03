package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/aymanbagabas/go-udiff"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
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
			&cli.BoolFlag{
				Name:     "dry-run",
				Usage:    "Print changes without modifying the files",
				Required: false,
				Value:    false,
			},
			&cli.BoolFlag{
				Name:     "verbose",
				Usage:    "Log at info-level",
				Required: false,
				Value:    false,
			},
			&cli.BoolFlag{
				Name:     "debug",
				Usage:    "Log at debug-level",
				Required: false,
				Value:    false,
			},
		},
		Action: func(cliCtx *cli.Context) error {
			repoURL := cliCtx.String("repo-url")
			defaultBranch := cliCtx.String("default-branch")
			rootDir := cliCtx.String("root-dir")

			level := slog.LevelWarn
			if cliCtx.Bool("debug") {
				level = slog.LevelDebug
			}
			if level == slog.LevelDebug && cliCtx.Bool("verbose") {
				level = slog.LevelInfo
			}

			ctx := cliCtx.Context
			var logger *slog.Logger
			if os.Getenv("GITHUB_ACTIONS") == "true" {
				handler := &human.Handler{}
				logger = slog.New(&actionslog.Wrapper{
					Handler: handler.WithOutput,
					Level:   level,
				})
			} else {
				if term.IsTerminal(int(os.Stderr.Fd())) {
					logger = slog.New(
						tint.NewHandler(os.Stderr, &tint.Options{
							Level: level,
						}),
					)
				} else {
					logger = slog.New(
						slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
							Level: level,
						}),
					)
				}
			}
			ctrl := controller{
				filesys:       afero.NewOsFs(),
				dryRun:        cliCtx.Bool("dry-run"),
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
	dryRun        bool
	logger        *slog.Logger
	filesys       afero.Fs
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
	return afero.Walk(ctrl.filesys, docsRootDirectory, func(path string, d fs.FileInfo, err error) error {
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
		filesys:           ctrl.filesys,
		ctx:               ctx,
		logger:            logger,
		rootDirectory:     ctrl.rootDirectory,
		docsRootDirectory: ctrl.docsRootDirectory,
		path:              path,
		repoURL:           ctrl.repoURL,
		defaultBranch:     ctrl.defaultBranch,
		dryRun:            ctrl.dryRun,
	}
	markdown.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(transformer, 999)))
	source, err := afero.ReadFile(ctrl.filesys, path)
	if err != nil {
		return err
	}
	out := bytes.Buffer{}
	if err := markdown.Convert(source, &out); err != nil {
		return err
	}
	if transformer.changed {
		if ctrl.dryRun {
			// Print a diff only and then return
			changes := udiff.Strings(string(source), out.String())
			unifiedChanges, err := udiff.ToUnifiedDiff(path, path+".updated", string(source), changes, 5)
			if err != nil {
				return err
			}
			fmt.Println(unifiedChanges)
			return nil
		}
		// Only update the original file if there were actual changes made:
		if err := afero.WriteFile(ctrl.filesys, path, out.Bytes(), 0600); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *controller) getDocsDir(ctx context.Context, root string) (string, error) {
	cfg := mkdocsYaml{}
	content, err := afero.ReadFile(ctrl.filesys, filepath.Join(root, "mkdocs.yml"))
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
	filesys           afero.Fs
	rootDirectory     string
	docsRootDirectory string
	path              string
	repoURL           string
	defaultBranch     string
	changed           bool
	dryRun            bool
}

func (transformer *relativeLinkASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		// We only care about links. Images etc. are not handled:
		link, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		// If the destination points somewhere outside of the root
		// directory, then we care:
		dst := string(link.Destination)
		if !strings.HasPrefix(dst, "..") {
			return ast.WalkContinue, nil
		}

		// If the destination has a hashname, we need to strip that and append
		// it afterwards again for the filesystem checks to work:
		hashname := ""
		elems := strings.SplitN(dst, "#", 2)
		if len(elems) > 1 {
			hashname = elems[1]
			dst = elems[0]
		}

		absPath := filepath.Join(filepath.Dir(transformer.path), dst)
		destStat, err := transformer.filesys.Stat(absPath)
		if err != nil {
			transformer.logger.WarnContext(transformer.ctx, "mapped destination not found", slog.String("old-dest", dst), slog.String("abs-path", absPath))
			return ast.WalkContinue, nil
		}
		rel, _ := filepath.Rel(transformer.docsRootDirectory, absPath)
		if !strings.HasPrefix(rel, "..") {
			return ast.WalkContinue, nil
		}

		typeSegment := "tree"
		if !destStat.IsDir() {
			typeSegment = "blob"
		}
		newDest := strings.Replace(absPath, strings.TrimSuffix(transformer.rootDirectory, "/"), transformer.repoURL+"/"+typeSegment+"/"+transformer.defaultBranch, 1)
		if hashname != "" {
			newDest = newDest + "#" + hashname
		}
		transformer.logger.InfoContext(transformer.ctx, "rewriting path", slog.String("old-dest", dst), slog.String("new-dest", newDest), slog.Bool("dry-run", transformer.dryRun))
		link.Destination = []byte(newDest)
		transformer.changed = true
		return ast.WalkContinue, nil
	})
}
