package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMainConfiguration(t *testing.T) {
	tests := map[string]struct {
		args  []string
		env   map[string]string
		check func(t *testing.T, cfg FullConfig)
	}{
		"with-extras": {
			env: map[string]string{
				"ARGO_TOKEN": "my-token",
			},
			args: []string{
				"root",
				"--print-config",
				"--namespace", "my-namespace",
				"--log-level", "debug",
				"--instance", "ops",
				"submit",
				"--name", "something", "--param", "whatever",
			},
			check: func(t *testing.T, cfg FullConfig) {
				require.Equal(t, []string{"--name", "something", "--param", "whatever"}, cfg.ExtraArgs)
				require.Equal(t, slog.LevelDebug, cfg.LogLevel.Level())

			},
		},
		"argo-token-as-flag": {
			args: []string{
				"root",
				"--print-config",
				"--argo-token", "my-token",
				"--namespace", "my-namespace",
				"--log-level", "warn",
				"--instance", "ops",
				"submit",
			},
			check: func(t *testing.T, cfg FullConfig) {
				require.Equal(t, "my-token", cfg.ArgoToken)
			},
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			for envName, envValue := range test.env {
				t.Setenv(envName, envValue)
			}
			writer := &bytes.Buffer{}
			runMain(test.args, writer, os.Stderr)
			cfg := FullConfig{}
			require.NoError(t, json.NewDecoder(writer).Decode(&cfg))
			test.check(t, cfg)
		})
	}
}
