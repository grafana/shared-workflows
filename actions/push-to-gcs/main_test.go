package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.ArtifactDir(), "workspace")
	for _, name := range []string{"dist/package.tar.gz", "dist/package with spaces.tar.gz", "dist/nested/package.sha256", "dist/.hidden/file.txt"} {
		putFile(t, filepath.Join(root, name))
	}
	return root
}

func putFile(t *testing.T, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte("artifact"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func objects(plan uploadPlan) []string {
	names := make([]string, len(plan.files))
	for i, file := range plan.files {
		names[i] = file.object
	}
	return names
}

func TestPlanUploads(t *testing.T) {
	tests := []struct {
		name   string
		source string
		glob   string
		parent bool
		prefix string
		want   []string
	}{
		{name: "nested and hidden files", source: "dist", want: []string{"releases/.hidden/file.txt", "releases/nested/package.sha256", "releases/package with spaces.tar.gz", "releases/package.tar.gz"}},
		{name: "directory parent", source: "dist/", parent: true, glob: "nested/*", want: []string{"releases/dist/nested/package.sha256"}},
		{name: "file without parent", source: "dist/package.tar.gz", want: []string{"releases/package.tar.gz"}},
		{name: "file ignores parent", source: "dist/package.tar.gz", parent: true, want: []string{"releases/package.tar.gz"}},
		{name: "relative glob", source: "dist", glob: "nested/*.sha256", want: []string{"releases/nested/package.sha256"}},
		{name: "recursive glob", source: "dist", glob: "**/*.txt", want: []string{"releases/.hidden/file.txt"}},
		{name: "literal extglob", source: "dist", glob: "**/@(package.sha256|file.txt)", want: []string{"releases/.hidden/file.txt", "releases/nested/package.sha256"}},
		{name: "negative extglob", source: "dist", glob: "!(*.tar.gz)", want: []string{}},
		{name: "negative literal extglob", source: "dist", glob: "!(package.tar.gz)", want: []string{"releases/package with spaces.tar.gz"}},
		{name: "glob with spaces", source: "dist", glob: "package with *.tar.gz", want: []string{"releases/package with spaces.tar.gz"}},
		{name: "workspace root", source: ".", parent: true, glob: "dist/*.tar.gz", want: []string{"releases/dist/package with spaces.tar.gz", "releases/dist/package.tar.gz"}},
		{name: "absolute source", source: "ABSOLUTE", glob: "*.tar.gz", want: []string{"releases/package with spaces.tar.gz", "releases/package.tar.gz"}},
		{name: "absolute glob", source: "dist", glob: "ABSOLUTE", want: []string{"releases/package with spaces.tar.gz", "releases/package.tar.gz"}},
		{name: "default bucket path", source: "dist/package.tar.gz", prefix: ".", want: []string{"package.tar.gz"}},
		{name: "no matches", source: "dist", glob: "*.missing", want: []string{}},
		{name: "outside source", source: "dist", glob: "../outside.txt", want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := fixture(t)
			putFile(t, filepath.Join(root, "outside.txt"))
			c := config{workspace: root, source: tt.source, glob: tt.glob, parent: tt.parent, destination: "bucket/releases/"}
			if tt.source == "ABSOLUTE" {
				c.source = filepath.Join(root, "dist")
			}
			if tt.glob == "ABSOLUTE" {
				putFile(t, filepath.Join(root, "dist/nested/extra.tar.gz"))
				c.glob = filepath.Join(root, "dist/*.tar.gz")
			}
			if tt.prefix != "" {
				c.destination = "bucket/" + tt.prefix
			}
			plan, err := planUploads(t.Context(), c)
			if err != nil {
				t.Fatal(err)
			}
			if plan.bucket != "bucket" || !slices.Equal(objects(plan), tt.want) {
				t.Fatalf("plan = %q/%q, want bucket/%q", plan.bucket, objects(plan), tt.want)
			}
			for _, file := range plan.files {
				if !filepath.IsAbs(file.source) {
					t.Errorf("source is not absolute: %q", file.source)
				}
				if _, err := os.Stat(file.source); err != nil {
					t.Errorf("source does not exist: %v", err)
				}
			}
		})
	}
}

func TestDirectorySymlinksAndCycles(t *testing.T) {
	root := fixture(t)
	putFile(t, filepath.Join(root, "shared/deep/asset.js"))
	if err := os.Symlink("../shared", filepath.Join(root, "dist/vendor")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../shared", filepath.Join(root, "dist/alias")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing", filepath.Join(root, "dist/dangling")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("..", filepath.Join(root, "dist/nested/back")); err != nil {
		t.Fatal(err)
	}
	plan, err := planUploads(t.Context(), config{workspace: root, source: "dist", destination: "bucket/releases"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.files) != 6 || !slices.Contains(objects(plan), "releases/vendor/deep/asset.js") || !slices.Contains(objects(plan), "releases/alias/deep/asset.js") {
		t.Fatalf("incomplete or repeated traversal: %q", objects(plan))
	}
}

func TestLiteralFilenamesAndCommands(t *testing.T) {
	root := fixture(t)
	for _, name := range []string{"app/[slug]/page.js", "app/s/page.js", "app/bundle*.js", "app/bundle?.js", "app/closing].js", "app/package $(command).js"} {
		putFile(t, filepath.Join(root, name))
	}
	c := config{workspace: root, source: "app", destination: "bucket/releases"}
	plan, err := planUploads(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.files) != 6 {
		t.Fatalf("selected %d files, want 6", len(plan.files))
	}
	for _, file := range plan.files {
		args := uploadArgs(plan.bucket, c, file)
		if !strings.HasSuffix(args[len(args)-1], "/") {
			t.Errorf("destination is not a directory prefix: %q", args)
		}
		if strings.Contains(file.source, "[slug]") && !strings.Contains(args[len(args)-2], "[[]slug]") {
			t.Errorf("brackets not escaped for gcloud: %q", args)
		}
		if strings.Contains(file.source, "bundle*") && !strings.Contains(args[len(args)-2], "bundle[*]") {
			t.Errorf("asterisk not escaped for gcloud: %q", args)
		}
		if strings.Contains(file.source, "bundle?") && !strings.Contains(args[len(args)-2], "bundle[?]") {
			t.Errorf("question mark not escaped for gcloud: %q", args)
		}
	}
}

func TestUploadArguments(t *testing.T) {
	for _, tt := range []struct {
		name   string
		c      config
		object string
		want   []string
	}{
		{name: "root prefix", object: "package.tar.gz", want: []string{"storage", "cp", "/tmp/package.tar.gz", "gs://bucket/"}},
		{name: "nested prefix and ACL", c: config{acl: "projectPrivate"}, object: "releases/nested/package.tar.gz", want: []string{"storage", "cp", "--predefined-acl=projectPrivate", "/tmp/package.tar.gz", "gs://bucket/releases/nested/"}},
		{name: "gzip and WIF", c: config{gzip: true, acl: " "}, object: "releases/package.tar.gz", want: []string{"storage", "cp", "--gzip-local-all", "/tmp/package.tar.gz", "gs://bucket/releases/"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := uploadArgs("bucket", tt.c, transfer{source: "/tmp/package.tar.gz", object: tt.object})
			if !slices.Equal(got, tt.want) {
				t.Fatalf("args = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlanningErrors(t *testing.T) {
	root := fixture(t)
	for _, tt := range []struct{ name, source, glob string }{
		{"missing source", "missing", ""},
		{"file and glob", "dist/package.tar.gz", "*"},
		{"brace expansion", "dist", "**/*.{gz,txt}"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := planUploads(t.Context(), config{workspace: root, source: tt.source, glob: tt.glob, destination: "bucket"})
			if err == nil {
				t.Fatal("expected a planning error")
			}
		})
	}
}

func TestUploadConcurrencyLimit(t *testing.T) {
	for _, tt := range []struct {
		name  string
		args  []string
		limit int32
	}{
		{name: "default", limit: 4},
		{name: "serial", args: []string{"-concurrency=1"}, limit: 1},
		{name: "two", args: []string{"-concurrency", "2"}, limit: 2},
		{name: "six", args: []string{"-concurrency=6"}, limit: 6},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := fixture(t)
			for i := range 4 {
				putFile(t, filepath.Join(root, fmt.Sprintf("dist/extra-%d", i)))
			}
			env := map[string]string{"GITHUB_WORKSPACE": root, "SOURCE_PATH": "dist", "DESTINATION": "bucket", "GITHUB_OUTPUT": filepath.Join(t.ArtifactDir(), "output")}
			synctest.Test(t, func(t *testing.T) {
				release := make(chan struct{})
				done := make(chan error, 1)
				var active, started atomic.Int32
				go func() {
					done <- run(context.Background(), tt.args, func(name string) string { return env[name] }, func(context.Context, []string) error {
						started.Add(1)
						active.Add(1)
						defer active.Add(-1)
						<-release
						return nil
					}, io.Discard)
				}()
				synctest.Wait()
				if active.Load() != tt.limit || started.Load() != tt.limit {
					t.Errorf("active/started = %d/%d, want %d/%d", active.Load(), started.Load(), tt.limit, tt.limit)
				}
				close(release)
				if err := <-done; err != nil {
					t.Fatal(err)
				}
				if active.Load() != 0 || started.Load() != 8 {
					t.Fatalf("unfinished uploads: active/started = %d/%d", active.Load(), started.Load())
				}
			})
		})
	}
}

func TestInvalidConcurrencyFlag(t *testing.T) {
	for _, args := range [][]string{{"-concurrency=0"}, {"-concurrency=-1"}, {"-concurrency=invalid"}, {"-concurrency"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			err := run(t.Context(), args, func(string) string {
				t.Error("invalid concurrency read upload configuration")
				return ""
			}, func(context.Context, []string) error {
				t.Error("invalid concurrency started an upload")
				return nil
			}, io.Discard)
			if err == nil || !strings.Contains(err.Error(), "concurrency") {
				t.Fatalf("error = %v, want concurrency validation", err)
			}
		})
	}
}

func TestUploadFailureCancelsPeers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		failure := errors.New("gcloud failed")
		plan := uploadPlan{bucket: "bucket", files: make([]transfer, 8)}
		fail := make(chan struct{})
		done := make(chan error, 1)
		var started, canceled atomic.Int32
		go func() {
			done <- uploadAll(context.Background(), config{concurrency: 2}, plan, func(ctx context.Context, _ []string) error {
				if started.Add(1) == 1 {
					<-fail
					return failure
				}
				<-ctx.Done()
				canceled.Add(1)
				return ctx.Err()
			})
		}()
		synctest.Wait()
		close(fail)
		if err := <-done; !errors.Is(err, failure) {
			t.Fatalf("error = %v, want original upload failure", err)
		}
		if started.Load() != 2 || canceled.Load() != 1 {
			t.Fatalf("started/canceled = %d/%d, want 2/1", started.Load(), canceled.Load())
		}
	})
}

func TestAlreadyCanceledUpload(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := uploadAll(ctx, config{concurrency: 4}, uploadPlan{files: []transfer{{source: "unused"}}}, func(context.Context, []string) error {
		t.Fatal("started a canceled upload")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want cancellation", err)
	}
}

func TestRunOutput(t *testing.T) {
	for _, mode := range []string{"success", "no matches", "failure"} {
		t.Run(mode, func(t *testing.T) {
			root := fixture(t)
			output := filepath.Join(t.ArtifactDir(), "github-output")
			env := map[string]string{"GITHUB_WORKSPACE": root, "SOURCE_PATH": "dist", "DESTINATION": "bucket/releases", "GITHUB_OUTPUT": output, "INCLUDE_PARENT": "false", "GZIP": "false"}
			if mode == "no matches" {
				env["SOURCE_GLOB"] = "*.missing"
			}
			var log bytes.Buffer
			var mu sync.Mutex
			var calls int
			failure := errors.New("upload failed")
			err := run(t.Context(), nil, func(name string) string { return env[name] }, func(context.Context, []string) error {
				mu.Lock()
				calls++
				mu.Unlock()
				if mode == "failure" {
					return failure
				}
				return nil
			}, &log)
			if mode == "failure" {
				if !errors.Is(err, failure) {
					t.Fatalf("error = %v, want upload failure", err)
				}
				if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("failed upload published an output")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
			if len(lines) != 3 || lines[0] != "uploaded<<"+lines[2] {
				t.Fatalf("invalid GitHub output: %q", data)
			}
			if mode == "no matches" {
				if calls != 0 || lines[1] != "" || !strings.Contains(log.String(), "::warning::") {
					t.Fatalf("unexpected empty upload: calls=%d, output=%q, log=%q", calls, data, log.String())
				}
			} else if calls != 4 || lines[1] != "releases/.hidden/file.txt,releases/nested/package.sha256,releases/package with spaces.tar.gz,releases/package.tar.gz" {
				t.Fatalf("unexpected successful output: calls=%d, output=%q", calls, data)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	base := map[string]string{"GITHUB_WORKSPACE": "/workspace", "SOURCE_PATH": "dist", "DESTINATION": "bucket", "GITHUB_OUTPUT": "/output"}
	for _, key := range []string{"GITHUB_WORKSPACE", "SOURCE_PATH", "DESTINATION", "GITHUB_OUTPUT", "INCLUDE_PARENT", "GZIP"} {
		t.Run(key, func(t *testing.T) {
			_, err := readConfig(func(name string) string {
				if name == key {
					if key == "INCLUDE_PARENT" || key == "GZIP" {
						return "invalid"
					}
					return ""
				}
				return base[name]
			})
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("error = %v, want %s validation", err, key)
			}
		})
	}
	c, err := readConfig(func(name string) string { return base[name] })
	if err != nil || !c.parent || !c.gzip {
		t.Fatalf("defaults = %+v, %v; want parent and gzip enabled", c, err)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestOutputWriteError(t *testing.T) {
	want := errors.New("output unavailable")
	if err := writeOutput(failingWriter{want}, []string{"object"}); !errors.Is(err, want) {
		t.Fatalf("error = %v, want write failure", err)
	}
	if err := writeOutput(io.Discard, []string{"object\nwith newline"}); err != nil {
		t.Fatal(err)
	}
}
