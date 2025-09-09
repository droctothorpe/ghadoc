package generate

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkflowInfo stores information about a GitHub workflow
type WorkflowInfo struct {
	Filename    string
	Description string
	Triggers    []string // List of all triggers (e.g., push, pull_request, workflow_dispatch, etc.)
}

// Generate generates the workflows.md file from the workflow files in the
// specified workflowsDir.
func Generate(workflowsDir string, output string) error {
	// Get all workflow files
	files, err := os.ReadDir(workflowsDir)
	if err != nil {
		return fmt.Errorf("error reading workflows directory: %v", err)
	}

	// Store workflow information
	var workflows []WorkflowInfo

	// Process each workflow file
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (ext == ".yml" || ext == ".yaml") {
			filePath := filepath.Join(workflowsDir, file.Name())
			workflow, err := parseWorkflowFile(filePath)
			if err != nil {
				fmt.Printf("Error parsing workflow file %s: %v\n", file.Name(), err)
				continue
			}
			workflow.Filename = file.Name()
			workflows = append(workflows, workflow)
		}
	}

	// Generate markdown table
	markdownTable := generateMarkdownTable(workflows, workflowsDir, output)

	// Write to output file
	err = os.WriteFile(output, []byte(markdownTable), 0644)
	if err != nil {
		return fmt.Errorf("error writing to output file: %v", err)
	}

	fmt.Println("Successfully generated", output)
	return nil
}

// parseWorkflowFile extracts information from a GitHub workflow file
func parseWorkflowFile(filePath string) (WorkflowInfo, error) {
	workflow := WorkflowInfo{}

	// Read file content for YAML parsing
	content, err := os.ReadFile(filePath)
	if err != nil {
		return workflow, err
	}

	// Extract description from lines starting with "##", but only if the first line starts with ##
	file, err := os.Open(filePath)
	if err != nil {
		return workflow, err
	}
	defer file.Close()

	// Read the file line by line to find the description
	scanner := bufio.NewScanner(file)
	var descriptionLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if !strings.HasPrefix(trimmedLine, "##") {
			break
		}

		// Extract the description by removing the ## prefix
		descriptionLine := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "##"))
		descriptionLines = append(descriptionLines, descriptionLine)
	}

	// Join description lines with line breaks for markdown
	if len(descriptionLines) > 0 {
		workflow.Description = strings.Join(descriptionLines, "<br>")
	} else {
		workflow.Description = ""
	}

	// Parse YAML to extract all triggers from the "on" field
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(content, &yamlData)
	if err != nil {
		return workflow, err
	}

	// Check if "on" field exists
	if onField, ok := yamlData["on"]; ok {
		// Extract triggers based on the type of the "on" field
		switch v := onField.(type) {
		case map[string]interface{}:
			// If "on" is a map, each key is a trigger type
			for key := range v {
				workflow.Triggers = append(workflow.Triggers, key)
			}
		case []interface{}:
			// If "on" is an array, each item is a trigger type
			for _, item := range v {
				if str, ok := item.(string); ok {
					workflow.Triggers = append(workflow.Triggers, str)
				}
			}
		case string:
			// If "on" is a string, it's a single trigger type
			workflow.Triggers = append(workflow.Triggers, v)
		}
	}

	return workflow, nil
}

// generateMarkdownTable creates a markdown table from workflow information
func generateMarkdownTable(workflows []WorkflowInfo, workflowsDir string, outputPath string) string {
	var sb strings.Builder

	// Write table header
	sb.WriteString("# GitHub Workflows Summary\n\n")
	sb.WriteString("| Filename | Description | Triggers |\n")
	sb.WriteString("| --- | --- | --- |\n")

	// Write table rows
	for _, workflow := range workflows {
		// Create relative link to the file
		// Create link to workflow file with relative path from the markdown file
		workflowFullPath := filepath.Join(workflowsDir, workflow.Filename)
		outputDir := filepath.Dir(outputPath)

		// Calculate relative path from output directory to workflow file
		relativePath, err := filepath.Rel(outputDir, workflowFullPath)
		if err != nil {
			// Fallback to just the filename if there's an error
			relativePath = workflow.Filename
		}

		// Use forward slashes for URLs even on Windows
		relativePath = filepath.ToSlash(relativePath)
		fileLink := fmt.Sprintf("[%s](%s)", workflow.Filename, relativePath)

		// Format triggers as a comma-separated list
		triggers := strings.Join(workflow.Triggers, ", ")

		// Write row
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			fileLink,
			workflow.Description,
			triggers))
	}

	return sb.String()
}
