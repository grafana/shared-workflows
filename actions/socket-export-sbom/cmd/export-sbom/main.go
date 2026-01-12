package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/grafana/shared-workflows/actions/socket-export-sbom"
)

func main() {
	Usage := `
    Usage: main.go <repo name> <output file>
    `
	key := os.Getenv("SOCKET_API_TOKEN")
	if key == "" {
		fmt.Fprintln(os.Stderr, "SOCKET_API_TOKEN not provided")
		os.Exit(1)
	}
	org := os.Getenv("SOCKET_ORG")
	if org == "" {
		fmt.Fprintln(os.Stderr, "please specify socket org name, e.g. 'grafana'")
		os.Exit(1)
	}
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stdout, Usage)
		os.Exit(1)
	}
	repo, output := os.Args[1], os.Args[2]
	baseURL := os.Getenv("SOCKET_BASE_URL")
	client := socket.NewClient(key, org, socket.WithBaseURL(baseURL))
	sbom, err := client.GetSBOM(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	data, err := json.Marshal(sbom)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = os.WriteFile(output, data, 0o600)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, "written sbom to output file: ", output)
}
