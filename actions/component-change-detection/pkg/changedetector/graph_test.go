package changedetector

import (
	"strings"
	"testing"
)

func TestBuildDependencyGraph(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid graph",
			config: &Config{
				Components: map[string]ComponentConfig{
					"migrator": {
						Paths:        []string{"migrations/**"},
						Dependencies: []string{},
					},
					"apiserver": {
						Paths:        []string{"pkg/**"},
						Dependencies: []string{"migrator"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "circular dependency",
			config: &Config{
				Components: map[string]ComponentConfig{
					"a": {
						Paths:        []string{"a/**"},
						Dependencies: []string{"b"},
					},
					"b": {
						Paths:        []string{"b/**"},
						Dependencies: []string{"a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "self dependency",
			config: &Config{
				Components: map[string]ComponentConfig{
					"a": {
						Paths:        []string{"a/**"},
						Dependencies: []string{"a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "three node cycle",
			config: &Config{
				Components: map[string]ComponentConfig{
					"a": {
						Paths:        []string{"a/**"},
						Dependencies: []string{"b"},
					},
					"b": {
						Paths:        []string{"b/**"},
						Dependencies: []string{"c"},
					},
					"c": {
						Paths:        []string{"c/**"},
						Dependencies: []string{"a"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "complex graph with cycle",
			config: &Config{
				Components: map[string]ComponentConfig{
					"root": {
						Paths: []string{"root/**"},
					},
					"a": {
						Paths:        []string{"a/**"},
						Dependencies: []string{"root"},
					},
					"b": {
						Paths:        []string{"b/**"},
						Dependencies: []string{"a", "c"},
					},
					"c": {
						Paths:        []string{"c/**"},
						Dependencies: []string{"b"}, // b->c->b cycle
					},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildDependencyGraph(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildDependencyGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("BuildDependencyGraph() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestPropagateChanges(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		directChanges map[string]bool
		want          map[string]bool
	}{
		{
			name: "no dependencies - no propagation",
			config: &Config{
				Components: map[string]ComponentConfig{
					"a": {Paths: []string{"a/**"}},
					"b": {Paths: []string{"b/**"}},
				},
			},
			directChanges: map[string]bool{"a": true, "b": false},
			want:          map[string]bool{"a": true, "b": false},
		},
		{
			name: "linear dependency chain",
			config: &Config{
				Components: map[string]ComponentConfig{
					"migrator": {
						Paths: []string{"migrations/**"},
					},
					"apiserver": {
						Paths:        []string{"api/**"},
						Dependencies: []string{"migrator"},
					},
				},
			},
			directChanges: map[string]bool{"migrator": true, "apiserver": false},
			want:          map[string]bool{"migrator": true, "apiserver": true},
		},
		{
			name: "multiple dependents",
			config: &Config{
				Components: map[string]ComponentConfig{
					"migrator": {
						Paths: []string{"migrations/**"},
					},
					"apiserver": {
						Paths:        []string{"api/**"},
						Dependencies: []string{"migrator"},
					},
					"controller": {
						Paths:        []string{"controller/**"},
						Dependencies: []string{"migrator"},
					},
				},
			},
			directChanges: map[string]bool{"migrator": true, "apiserver": false, "controller": false},
			want:          map[string]bool{"migrator": true, "apiserver": true, "controller": true},
		},
		{
			name: "transitive dependencies",
			config: &Config{
				Components: map[string]ComponentConfig{
					"a": {Paths: []string{"a/**"}},
					"b": {Paths: []string{"b/**"}, Dependencies: []string{"a"}},
					"c": {Paths: []string{"c/**"}, Dependencies: []string{"b"}},
				},
			},
			directChanges: map[string]bool{"a": true, "b": false, "c": false},
			want:          map[string]bool{"a": true, "b": true, "c": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := BuildDependencyGraph(tt.config)
			if err != nil {
				t.Fatalf("BuildDependencyGraph() failed: %v", err)
			}

			got := PropagateChanges(graph, tt.directChanges)

			for comp, wantChanged := range tt.want {
				if got[comp] != wantChanged {
					t.Errorf("PropagateChanges() component %q = %v, want %v", comp, got[comp], wantChanged)
				}
			}
		})
	}
}
