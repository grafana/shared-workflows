package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

var discard slog.Handler = discardHandler{}

// discardHandler is a slog.Handler that discards all log messages. We require a
// logger in the App struct, but we don't care about the output in the tests.
type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (d discardHandler) WithAttrs([]slog.Attr) slog.Handler      { return d }
func (d discardHandler) WithGroup(string) slog.Handler           { return d }

func app() App {
	return App{
		logger: slog.New(discard),
	}
}

func TestEnvToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "token with Bearer prefix",
			token:    "Bearer abc",
			expected: "Bearer abc",
		},
		{
			name:     "token with Token prefix",
			token:    "Token abc",
			expected: "Token abc",
		},
		{
			name:     "token without prefix gets prefix added",
			token:    "abc",
			expected: "Bearer abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := app()
			a.argoToken = tt.token

			expected := fmt.Sprintf("ARGO_TOKEN=%s", tt.expected)

			require.Contains(t, a.env(), expected)
		})
	}
}

func TestOutputWithURI(t *testing.T) {
	// The trailing whitespace here is copied from the actual output of the
	// Argo CLI.
	successOutput := `Name:                hello-world-4kx2g
Namespace:           argo
ServiceAccount:      unset
Status:              Pending
Created:             Wed Dec 13 00:00:00 +0000 (now)
Progress:
Parameters:
  message:           world
`

	tests := []struct {
		name     string
		instance string
		input    string
		expected string
	}{
		{
			name:     "output with name (dev)",
			input:    successOutput,
			instance: "dev",
			expected: "https://argo-workflows-dev.grafana.net:443/workflows/argo/hello-world-4kx2g",
		},
		{
			name:     "output with name (ops)",
			input:    successOutput,
			instance: "ops",
			expected: "https://argo-workflows.grafana.net:443/workflows/argo/hello-world-4kx2g",
		},
		{
			name: "output without name",
			input: `Namespace:           argo
ServiceAccount:      unset
Status:              Pending
Created:             Wed Dec 13 00:00:00 +0000 (now)
Progress:
Parameters:
  message:           world
`,
			instance: "dev",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := app()
			a.instance = tt.instance
			a.namespace = "argo"

			reader := bytes.NewBufferString(tt.input)
			uri, out, err := a.outputWithURI(reader)
			require.NoError(t, err, "unexpected error reading command output")

			require.Equal(t, tt.expected, uri)
			require.Equal(t, tt.input, out)
		})
	}
}
func TestSetURIAsJobOutput(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		uri            string
		expectedOutput string
	}{
		{
			name:           "Submit Command",
			command:        "submit",
			uri:            "https://example.com",
			expectedOutput: "uri=https://example.com\n",
		},
		{
			name:           "Non-Submit Command",
			command:        "cancel",
			uri:            "https://example.com",
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := app()
			a.command = tt.command

			var buf bytes.Buffer
			a.setURIAsJobOutput(tt.uri, &buf)

			require.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}

func TestIsFatalError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "error without 'AlreadyExists'",
			err:      errors.New("rpc error: code = InvalidName desc = workflows.argoproj.io \"my-workflow-@#$#$\" is invalid name"),
			expected: false,
		},
		{
			name:     "error containing 'AlreadyExists'",
			err:      errors.New("rpc error: code = AlreadyExists desc = workflows.argoproj.io \"my-workflow-1\" already exists"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFatalError(tt.err)
			require.Equal(t, result, tt.expected)
		})
	}
}
