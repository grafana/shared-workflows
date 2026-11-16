package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sync/errgroup"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/pattern"
	"mvdan.cc/sh/v3/syntax"
)

type config struct {
	concurrency                                       int
	workspace, source, glob, destination, output, acl string
	parent, gzip                                      bool
}

type transfer struct{ source, object string }
type uploadPlan struct {
	bucket string
	files  []transfer
}

type commandRunner func(context.Context, []string) error

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Getenv, runGcloud, os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readConfig(env func(string) string) (config, error) {
	c := config{
		workspace: env("GITHUB_WORKSPACE"), source: env("SOURCE_PATH"),
		glob: env("SOURCE_GLOB"), destination: env("DESTINATION"),
		output: env("GITHUB_OUTPUT"), acl: strings.TrimSpace(env("PREDEFINED_ACL")),
	}
	for _, key := range []string{"GITHUB_WORKSPACE", "SOURCE_PATH", "DESTINATION", "GITHUB_OUTPUT"} {
		if env(key) == "" {
			return config{}, fmt.Errorf("%s is required", key)
		}
	}
	for key, value := range map[string]*bool{"INCLUDE_PARENT": &c.parent, "GZIP": &c.gzip} {
		raw := env(key)
		if raw == "" {
			raw = "true"
		}
		var err error
		*value, err = strconv.ParseBool(raw)
		if err != nil {
			return config{}, fmt.Errorf("%s: %w", key, err)
		}
	}
	return c, nil
}

func planUploads(ctx context.Context, c config) (uploadPlan, error) {
	if err := ctx.Err(); err != nil {
		return uploadPlan{}, err
	}
	if strings.ContainsAny(c.glob, "{}") {
		return uploadPlan{}, errors.New("brace expansion is unsupported in CLI glob patterns")
	}
	bucket, prefix, _ := strings.Cut(strings.TrimPrefix(c.destination, "gs://"), "/")
	if bucket == "" {
		return uploadPlan{}, errors.New("destination bucket is required")
	}
	source := c.source
	if !filepath.IsAbs(source) {
		source = filepath.Join(c.workspace, source)
	}
	source, err := filepath.Abs(source)
	if err != nil {
		return uploadPlan{}, fmt.Errorf("resolve source: %w", err)
	}
	info, err := os.Stat(source)
	if err != nil {
		return uploadPlan{}, fmt.Errorf("stat source: %w", err)
	}
	root, parent := source, ""
	var names []string
	if info.IsDir() {
		if c.parent && filepath.Clean(c.source) != "." && filepath.Base(source) != string(filepath.Separator) {
			parent = filepath.Base(source)
		}
		glob := c.glob
		if glob == "" {
			glob = "**/*"
		}
		if filepath.IsAbs(glob) {
			glob, err = filepath.Rel(root, glob)
			if err != nil {
				return uploadPlan{}, fmt.Errorf("resolve glob: %w", err)
			}
		}
		names, err = globFiles(ctx, root, glob)
		if err != nil {
			return uploadPlan{}, fmt.Errorf("select files: %w", err)
		}
	} else {
		if c.glob != "" {
			return uploadPlan{}, errors.New("the glob input cannot be used with a file path")
		}
		root, names = filepath.Dir(source), []string{filepath.Base(source)}
	}
	plan := uploadPlan{bucket: bucket}
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return uploadPlan{}, err
		}
		file := filepath.Join(root, name)
		info, err := os.Stat(file)
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ELOOP) {
			continue
		}
		if err != nil {
			return uploadPlan{}, fmt.Errorf("stat selected file: %w", err)
		}
		if info.Mode().IsRegular() {
			plan.files = append(plan.files, transfer{source: file, object: path.Join(prefix, parent, filepath.ToSlash(name))})
		}
	}
	slices.SortFunc(plan.files, func(a, b transfer) int { return strings.Compare(a.source, b.source) })
	return plan, nil
}

func globFiles(ctx context.Context, root, glob string) ([]string, error) {
	var scanErr error
	readDir := func(name string) ([]fs.DirEntry, error) {
		if err := ctx.Err(); err != nil {
			scanErr = err
			return nil, err
		}
		info, err := os.Stat(name)
		if err == nil && filepath.Clean(name) != root {
			// Compare ancestors, rather than all visited directories, to retain symlink aliases.
			for ancestor := filepath.Dir(name); ; ancestor = filepath.Dir(ancestor) {
				previous, statErr := os.Stat(ancestor)
				if statErr != nil {
					err = statErr
					break
				}
				if os.SameFile(info, previous) {
					return nil, fs.ErrNotExist
				}
				if ancestor == root || ancestor == filepath.Dir(ancestor) {
					break
				}
			}
		}
		var entries []fs.DirEntry
		if err == nil {
			entries, err = os.ReadDir(name)
		}
		if errors.Is(err, syscall.ELOOP) {
			return nil, fs.ErrNotExist
		}
		if err != nil && !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, syscall.ENOTDIR) {
			scanErr = err
		}
		return entries, err
	}
	parts := strings.Split(filepath.ToSlash(glob), "/")
	for i, part := range parts {
		if part == ".." {
			return nil, nil
		}
		// Ensure literal-only extglobs take the library's pattern-matching path.
		if (strings.Contains(part, "@(") || strings.Contains(part, "!(") || strings.Contains(part, "+(")) && !pattern.HasMeta(part, 0) {
			for j, r := range part {
				if !strings.ContainsRune("!@+?*()|[]\\^-", r) {
					parts[i] = part[:j] + "[" + string(r) + "]" + part[j+len(string(r)):]
					break
				}
			}
		}
	}
	cfg := &expand.Config{
		Env: expand.ListEnviron("PWD=" + root), ReadDir2: readDir,
		GlobStar: true, DotGlob: true, ExtGlob: true, NullGlob: true,
	}
	// A literal syntax node expands globs without evaluating shell expressions from inputs.
	names, err := expand.Fields(cfg, &syntax.Word{Parts: []syntax.WordPart{&syntax.Lit{Value: "./" + strings.Join(parts, "/")}}})
	return names, errors.Join(err, scanErr)
}

func uploadArgs(bucket string, c config, file transfer) []string {
	args := []string{"storage", "cp"}
	if c.gzip {
		args = append(args, "--gzip-local-all")
	}
	if acl := strings.TrimSpace(c.acl); acl != "" {
		args = append(args, "--predefined-acl="+acl)
	}
	// gcloud also expands wildcard characters in quoted source arguments.
	literal := strings.NewReplacer("[", "[[]", "*", "[*]", "?", "[?]").Replace(file.source)
	directory := path.Dir(file.object)
	if directory == "." {
		directory = ""
	}
	target := "gs://" + bucket + "/"
	if directory != "" {
		target += directory + "/"
	}
	return append(args, literal, target)
}

func uploadAll(ctx context.Context, c config, plan uploadPlan, command commandRunner) error {
	group, uploadCtx := errgroup.WithContext(ctx)
	group.SetLimit(c.concurrency)
	for _, file := range plan.files {
		if uploadCtx.Err() != nil {
			break
		}
		group.Go(func() error {
			if err := uploadCtx.Err(); err != nil {
				return err
			}
			if err := command(uploadCtx, uploadArgs(plan.bucket, c, file)); err != nil {
				return fmt.Errorf("upload %q: %w", file.object, err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}
	return ctx.Err()
}

func runGcloud(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func run(ctx context.Context, args []string, env func(string) string, command commandRunner, log io.Writer) error {
	flags := flag.NewFlagSet("upload", flag.ContinueOnError)
	flags.SetOutput(log)
	concurrency := flags.Int("concurrency", 1, "maximum simultaneous gcloud uploads (must be positive)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %q", flags.Args())
	}
	if *concurrency < 1 {
		return errors.New("-concurrency must be at least 1")
	}
	c, err := readConfig(env)
	if err != nil {
		return err
	}
	c.concurrency = *concurrency
	plan, err := planUploads(ctx, c)
	if err != nil {
		return err
	}
	if err := uploadAll(ctx, c, plan, command); err != nil {
		return err
	}
	if len(plan.files) == 0 {
		if _, err := fmt.Fprintln(log, "::warning::No files matched the upload path and glob"); err != nil {
			return err
		}
	}
	output, err := os.OpenFile(c.output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open GitHub output: %w", err)
	}
	names := make([]string, len(plan.files))
	for i, file := range plan.files {
		names[i] = file.object
	}
	return errors.Join(writeOutput(output, names), output.Close())
}

func writeOutput(w io.Writer, names []string) error {
	delimiter := "uploaded_" + rand.Text()
	_, err := fmt.Fprintf(w, "uploaded<<%s\n%s\n%s\n", delimiter, strings.Join(names, ","), delimiter)
	return err
}
