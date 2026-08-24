package changedetector

import (
	"errors"
	"fmt"
	"os"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads and validates the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing YAML config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig performs validation on the loaded configuration
func validateConfig(config *Config) error {
	if len(config.Components) == 0 {
		return errors.New("configuration must define at least one component")
	}

	// Validate global exclude patterns
	for _, pattern := range config.GlobalExcludes {
		if err := validateGlobPattern(pattern); err != nil {
			return fmt.Errorf("global_excludes has invalid pattern %q: %w", pattern, err)
		}
	}

	// Validate all components in single loop
	for name, comp := range config.Components {
		// Check for empty paths
		if len(comp.Paths) == 0 {
			return fmt.Errorf("component %q must define at least one path", name)
		}

		// Check for undefined dependencies
		for _, dep := range comp.Dependencies {
			if _, exists := config.Components[dep]; !exists {
				return fmt.Errorf("component %q depends on undefined component %q", name, dep)
			}
		}

		// Validate include patterns
		for _, pattern := range comp.Paths {
			if err := validateGlobPattern(pattern); err != nil {
				return fmt.Errorf("component %q has invalid path pattern %q: %w", name, pattern, err)
			}
		}

		// Validate exclude patterns
		for _, pattern := range comp.Excludes {
			if err := validateGlobPattern(pattern); err != nil {
				return fmt.Errorf("component %q has invalid exclude pattern %q: %w", name, pattern, err)
			}
		}
	}

	return nil
}

// validateGlobPattern checks if a glob pattern is valid
func validateGlobPattern(pattern string) error {
	// Try to validate the pattern by testing it against a dummy path
	// doublestar.Match will return an error for invalid patterns
	_, err := doublestar.Match(pattern, "test")
	return err
}
