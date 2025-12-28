package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/grafana/shared-workflows/actions/docker-import-digests-push-manifest/oci-output/internal/parser"
)

func runInspect(ctx context.Context, tag string) ([]byte, error) {
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cctx, "docker", "buildx", "imagetools", "inspect", tag)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("inspect failed for %s: %w\n%s", tag, err, string(out))
	}
	return out, nil
}

func main() {
	var (
		fromFile = flag.String("from-file", "", "read imagetools output from file")
		stdin    = flag.Bool("stdin", false, "read imagetools output from stdin")
	)
	flag.Parse()

	ctx := context.Background()

	var reports []parser.Report

	switch {
	case *fromFile != "":
		data, err := os.ReadFile(*fromFile)
		if err != nil {
			panic(err)
		}
		reports = append(reports, parser.ParseInspectOutput(data))

	case *stdin:
		data, err := os.ReadFile(os.Stdin.Name())
		if err != nil {
			panic(err)
		}
		reports = append(reports, parser.ParseInspectOutput(data))

	default:
		tags := flag.Args()
		if len(tags) == 0 {
			fmt.Fprintln(os.Stderr, "usage: manifest-report [--from-file file] [--stdin] <tag...>")
			os.Exit(2)
		}
		for _, tag := range tags {
			raw, err := runInspect(ctx, tag)
			if err != nil {
				panic(err)
			}
			r := parser.ParseInspectOutput(raw)
			r.Tag = tag
			reports = append(reports, r)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(reports)
}
