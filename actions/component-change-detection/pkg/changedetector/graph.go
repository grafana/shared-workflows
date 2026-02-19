package changedetector

import (
	"fmt"
	"os"
	"strings"
)

// Graph represents a dependency graph
type Graph struct {
	nodes map[string][]string // node -> list of nodes that depend on it
}

// NewGraph creates a new empty dependency graph
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string][]string),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(name string) {
	if _, exists := g.nodes[name]; !exists {
		g.nodes[name] = make([]string, 0)
	}
}

// AddEdge adds a dependency edge (from depends on to)
func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	// Store reverse edges: to -> from means "from depends on to"
	// so if "to" changes, "from" needs to be marked as changed
	g.nodes[to] = append(g.nodes[to], from)
}

// BuildDependencyGraph creates a dependency graph from the configuration
func BuildDependencyGraph(config *Config) (*Graph, error) {
	graph := NewGraph()

	// Add all components as nodes
	for name := range config.Components {
		graph.AddNode(name)
	}

	// Add dependency edges
	for name, comp := range config.Components {
		for _, dep := range comp.Dependencies {
			graph.AddEdge(name, dep)
		}
	}

	// Check for cycles
	if err := detectCycle(graph, config); err != nil {
		return nil, err
	}

	return graph, nil
}

// detectCycle checks for circular dependencies in the graph
func detectCycle(graph *Graph, config *Config) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for node := range graph.nodes {
		// Only start DFS from unvisited nodes to avoid redundant work
		if !visited[node] {
			if cycle := dfs(node, config, visited, recStack); cycle != nil {
				return fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " -> "))
			}
		}
	}

	return nil
}

// dfs performs depth-first search to detect cycles
func dfs(node string, config *Config, visited, recStack map[string]bool) []string {
	visited[node] = true
	recStack[node] = true

	// Check all dependencies of this node
	if comp, exists := config.Components[node]; exists {
		for _, dep := range comp.Dependencies {
			if !visited[dep] {
				if cycle := dfs(dep, config, visited, recStack); cycle != nil {
					return append([]string{node}, cycle...)
				}
			} else if recStack[dep] {
				// Found a cycle
				return []string{node, dep}
			}
		}
	}

	recStack[node] = false
	return nil
}

// PropagateChanges marks all downstream dependents of changed components
func PropagateChanges(graph *Graph, directChanges map[string]bool) map[string]bool {
	result := make(map[string]bool)

	// Copy direct changes
	for comp, changed := range directChanges {
		result[comp] = changed
	}

	// For each changed component, mark all dependents as changed
	for comp, changed := range directChanges {
		if changed {
			markDependentsWithLogging(graph, comp, comp, result)
		}
	}

	return result
}

// markDependents recursively marks all components that depend on the given component
func markDependents(graph *Graph, component string, changed map[string]bool) {
	if dependents, exists := graph.nodes[component]; exists {
		for _, dependent := range dependents {
			if !changed[dependent] {
				changed[dependent] = true
				// Recursively mark downstream dependents
				markDependents(graph, dependent, changed)
			}
		}
	}
}

// markDependentsWithLogging is like markDependents but logs why components are marked as changed
func markDependentsWithLogging(graph *Graph, rootCause string, component string, changed map[string]bool) {
	if dependents, exists := graph.nodes[component]; exists {
		for _, dependent := range dependents {
			if !changed[dependent] {
				changed[dependent] = true
				fmt.Fprintf(os.Stderr, "%s: (dependency on %s)\n", dependent, rootCause)
				// Recursively mark downstream dependents
				markDependentsWithLogging(graph, rootCause, dependent, changed)
			}
		}
	}
}
