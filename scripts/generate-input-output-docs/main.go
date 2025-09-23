package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type ActionIO struct {
	Name        string    `yaml:"-"`
	Description yaml.Node `yaml:"description"`
	Default     yaml.Node `yaml:"default,omitempty"`
	Required    bool      `yaml:"required,omitempty"`
	Type        yaml.Node `yaml:"type,omitempty"`
	Value       yaml.Node `yaml:"value,omitempty"`
}

func nodeString(n yaml.Node) string {
	return n.Value
}

type OrderedIO struct {
	IO ActionIO `yaml:",inline"`
}

type CompositeActionRaw struct {
	Inputs  map[string]ActionIO `yaml:"inputs"`
	Outputs map[string]ActionIO `yaml:"outputs"`
}

type CompositeAction struct {
	Inputs  []OrderedIO `yaml:"inputs"`
	Outputs []OrderedIO `yaml:"outputs"`
}

func (c *CompositeAction) alphabetizeIO() {
	sort.Slice(c.Inputs, func(i, j int) bool {
		return c.Inputs[i].IO.Name < c.Inputs[j].IO.Name
	})
	sort.Slice(c.Outputs, func(i, j int) bool {
		return c.Outputs[i].IO.Name < c.Outputs[j].IO.Name
	})
}

func (c *CompositeAction) parseYAML(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Step 1: unmarshal into map form
	var raw CompositeActionRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Step 2: convert + alphabetize
	for k, v := range raw.Inputs {
		io := v
		io.Name = k
		c.Inputs = append(c.Inputs, OrderedIO{IO: io})
	}
	for k, v := range raw.Outputs {
		io := v
		io.Name = k
		c.Outputs = append(c.Outputs, OrderedIO{IO: io})
	}

	c.alphabetizeIO()
	return nil
}

func (c *CompositeAction) printMarkdown(context *cli.Context, b strings.Builder) error {
	if len(c.Inputs) > 0 {
		b.WriteString("## Inputs\n")
		b.WriteString(renderTable(c.Inputs))
	}
	if len(c.Outputs) > 0 {
		b.WriteString("\n## Outputs\n")
		b.WriteString(renderTable(c.Outputs))
	}
	result := b.String()
	if context.Bool("pretty") {
		result = formatWithPrettier(result)
	}
	fmt.Print(result)
	return nil
}

func (c *CompositeAction) printYaml() error {
	var buff bytes.Buffer
	cr := CompositeActionRaw{
		Inputs:  make(map[string]ActionIO),
		Outputs: make(map[string]ActionIO),
	}

	for _, io := range c.Inputs {
		cr.Inputs[io.IO.Name] = io.IO
	}
	for _, io := range c.Outputs {
		cr.Outputs[io.IO.Name] = io.IO
	}

	yamlEncoder := yaml.NewEncoder(&buff)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(cr); err != nil {
		return err
	}
	fmt.Println(buff.String())
	return nil
}

func (c *CompositeAction) convertToReusableWorkflow() (wf *ReusableWorkflow, err error) {
	wf = &ReusableWorkflow{}
	wf.On.WorkflowCall.Inputs = c.Inputs
	wf.On.WorkflowCall.Outputs = c.Outputs

	for i := range wf.On.WorkflowCall.Inputs {
		if nodeString(wf.On.WorkflowCall.Inputs[i].IO.Type) == "" {
			wf.On.WorkflowCall.Inputs[i].IO.Type = yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "string",
				Style: 0, // plain style
			}
		}
	}

	return wf, nil
}

type ReusableWorkflowRaw struct {
	On struct {
		WorkflowCall struct {
			Inputs  map[string]ActionIO `yaml:"inputs"`
			Outputs map[string]ActionIO `yaml:"outputs"`
		} `yaml:"workflow_call"`
	} `yaml:"on"`
}

type ReusableWorkflow struct {
	On struct {
		WorkflowCall struct {
			Inputs  []OrderedIO `yaml:"inputs"`
			Outputs []OrderedIO `yaml:"outputs"`
		} `yaml:"workflow_call"`
	} `yaml:"on"`
}

func (w *ReusableWorkflow) alphabetizeIO() {
	sort.Slice(w.On.WorkflowCall.Inputs, func(i, j int) bool {
		return w.On.WorkflowCall.Inputs[i].IO.Name < w.On.WorkflowCall.Inputs[j].IO.Name
	})
	sort.Slice(w.On.WorkflowCall.Outputs, func(i, j int) bool {
		return w.On.WorkflowCall.Outputs[i].IO.Name < w.On.WorkflowCall.Outputs[j].IO.Name
	})
}

func (w *ReusableWorkflow) parseYAML(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Step 1: unmarshal into map form
	var raw ReusableWorkflowRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Step 2: convert + alphabetize
	for k, v := range raw.On.WorkflowCall.Inputs {
		io := v
		io.Name = k
		w.On.WorkflowCall.Inputs = append(w.On.WorkflowCall.Inputs, OrderedIO{IO: io})
	}
	for k, v := range raw.On.WorkflowCall.Outputs {
		io := v
		io.Name = k
		w.On.WorkflowCall.Outputs = append(w.On.WorkflowCall.Outputs, OrderedIO{IO: io})
	}

	w.alphabetizeIO()
	return nil
}

func (w *ReusableWorkflow) printMarkdown(context *cli.Context, b strings.Builder) error {
	if len(w.On.WorkflowCall.Inputs) > 0 {
		b.WriteString("## Inputs\n")
		b.WriteString(renderTable(w.On.WorkflowCall.Inputs))
	}
	if len(w.On.WorkflowCall.Outputs) > 0 {
		b.WriteString("\n## Outputs\n")
		b.WriteString(renderTable(w.On.WorkflowCall.Outputs))
	}
	result := b.String()
	if context.Bool("pretty") {
		result = formatWithPrettier(result)
	}
	fmt.Print(result)
	return nil
}

func (w *ReusableWorkflow) printYaml() error {
	var buff bytes.Buffer
	cr := ReusableWorkflowRaw{}
	cr.On.WorkflowCall.Inputs = make(map[string]ActionIO)
	cr.On.WorkflowCall.Outputs = make(map[string]ActionIO)

	for _, io := range w.On.WorkflowCall.Inputs {
		cr.On.WorkflowCall.Inputs[io.IO.Name] = io.IO
	}
	for _, io := range w.On.WorkflowCall.Outputs {
		cr.On.WorkflowCall.Outputs[io.IO.Name] = io.IO
	}

	yamlEncoder := yaml.NewEncoder(&buff)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(cr); err != nil {
		return err
	}
	fmt.Println(buff.String())
	return nil
}

func (w *ReusableWorkflow) convertToCompositeAction() (ca *CompositeAction, err error) {
	ca = &CompositeAction{}
	ca.Inputs = w.On.WorkflowCall.Inputs
	ca.Outputs = w.On.WorkflowCall.Outputs
	return ca, nil
}

// helper: safely extract the string value from a yaml.Node
func nodeString(n yaml.Node) string {
	// If the node is a zero-value, n.Value is empty string — that's fine.
	return n.Value
}

func guessType(io ActionIO) string {
	lower := strings.ToLower(nodeString(io.Default))
	if lower == "true" || lower == "false" {
		return "Boolean"
	}
	return "String"
}

func renderTable(entries []OrderedIO) string {
	var b strings.Builder
	b.WriteString("| Name | Type | Description |\n")
	b.WriteString("| ---- | ---- | ----------- |\n")
	for _, e := range entries {
		// prefer an explicit type if present, otherwise guess from default
		typ := nodeString(e.IO.Type)
		if typ == "" {
			typ = guessType(e.IO)
		}
		desc := strings.ReplaceAll(nodeString(e.IO.Description), "\n", " ")
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", e.IO.Name, typ, desc)
	}
	return b.String()
}

// formatWithPrettier runs prettier --parser markdown on given text
func formatWithPrettier(input string) string {
	cmd := exec.Command("prettier", "--parser", "markdown", "-w")
	cmd.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// fallback: return unformatted text if prettier fails
		return input
	}
	return out.String()
}

func main() {
	app := &cli.App{
		Name:  "action-inspector",
		Usage: "Parse GitHub Action composite YAML files",
		Commands: []*cli.Command{
			{
				Name:  "workflow",
				Usage: "Generate inputs/outputs from a shared workflow.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Path to shared workflow YAML",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "composite,workflow,markdown",
						Value:   "markdown",
					},
					&cli.BoolFlag{
						Name:  "pretty",
						Usage: "Run prettier on the output (requires prettier installed)",
						Value: true,
					},
				},
				Action: func(c *cli.Context) error {
					action := &ReusableWorkflow{}
					err := action.parseYAML(c.String("file"))
					if err != nil {
						return err
					}

					var b strings.Builder
					// Markdown
					if c.String("output") == "markdown" || c.String("output") == "md" {
						err = action.printMarkdown(c, b)
						if err != nil {
							return err
						}
					}
					// Composite
					if c.String("output") == "composite" || c.String("output") == "c" {
						var ca *CompositeAction
						ca, err = action.convertToCompositeAction()
						if err != nil {
							return err
						}
						err = ca.printYaml()
						if err != nil {
							return err
						}
					}
					// Workflow
					if c.String("output") == "workflow" || c.String("output") == "w" {
						err = action.printYaml()
						if err != nil {
							return err
						}
					}
					return nil
				},
			},

			{
				Name:  "composite",
				Usage: "Generate inputs/outputs from a composite action.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Path to composite action YAML",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "composite,workflow,markdown",
						Value:   "markdown",
					},
					&cli.BoolFlag{
						Name:  "pretty",
						Usage: "Run prettier on the output (requires prettier installed)",
						Value: true,
					},
				},
				Action: func(c *cli.Context) error {
					action := &CompositeAction{}
					err := action.parseYAML(c.String("file"))
					if err != nil {
						return err
					}

					var b strings.Builder
					// Markdown
					if c.String("output") == "markdown" || c.String("output") == "md" {
						err = action.printMarkdown(c, b)
						if err != nil {
							return err
						}
					}
					// Composite
					if c.String("output") == "composite" || c.String("output") == "c" {
						err = action.printYaml()
						if err != nil {
							return err
						}
					}
					// Workflow
					if c.String("output") == "workflow" || c.String("output") == "w" {
						var wf *ReusableWorkflow
						wf, err = action.convertToReusableWorkflow()
						if err != nil {
							return err
						}
						err = wf.printYaml()
						if err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
