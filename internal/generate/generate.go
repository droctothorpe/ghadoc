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
	OnPush      bool
	OnPR        bool
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
	markdownTable := generateMarkdownTable(workflows, workflowsDir)

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
	isFirstLine := true
	foundDescription := false

	for scanner.Scan() { // Continue until we find a non-## line after finding at least one ## line
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check if this is the first line
		if isFirstLine {
			isFirstLine = false
			// If the first line doesn't start with ##, don't include any description
			if !strings.HasPrefix(trimmedLine, "##") {
				break
			}
			// First line starts with ##, extract the description
			descriptionLine := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "##"))
			descriptionLines = append(descriptionLines, descriptionLine)
			foundDescription = true
		} else if strings.HasPrefix(trimmedLine, "##") {
			// Extract the description by removing the ## prefix
			descriptionLine := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "##"))
			descriptionLines = append(descriptionLines, descriptionLine)
		} else if foundDescription {
			// Stop parsing when we find a non-## line after finding at least one ## line
			break
		}
	}
	
	// Join description lines with line breaks for markdown
	if len(descriptionLines) > 0 {
		workflow.Description = strings.Join(descriptionLines, "<br>")
	}

	// Parse YAML to check for "on" field with "push" and "pull_request"
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(content, &yamlData)
	if err != nil {
		return workflow, err
	}

	// Check if "on" field exists
	if onField, ok := yamlData["on"]; ok {
		// Check for push trigger
		switch v := onField.(type) {
		case map[string]interface{}:
			_, workflow.OnPush = v["push"]
			_, workflow.OnPR = v["pull_request"]
		case []interface{}:
			for _, item := range v {
				if item == "push" {
					workflow.OnPush = true
				}
				if item == "pull_request" {
					workflow.OnPR = true
				}
			}
		case string:
			if v == "push" {
				workflow.OnPush = true
			}
			if v == "pull_request" {
				workflow.OnPR = true
			}
		}
	}

	return workflow, nil
}

// generateMarkdownTable creates a markdown table from workflow information
func generateMarkdownTable(workflows []WorkflowInfo, basePath string) string {
	var sb strings.Builder

	// Write table header
	sb.WriteString("# GitHub Workflows Summary\n\n")
	sb.WriteString("| Filename | Description | On Push | On PR |\n")
	sb.WriteString("| --- | --- | :---: | :---: |\n")

	// Write table rows
	for _, workflow := range workflows {
		// Create relative link to the file
		fileLink := fmt.Sprintf("[%s](%s)", workflow.Filename, filepath.Join(basePath, workflow.Filename))

		// Create checkmarks for triggers
		pushCheck := ""
		if workflow.OnPush {
			pushCheck = "✓"
		}

		prCheck := ""
		if workflow.OnPR {
			prCheck = "✓"
		}

		// Write row
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			fileLink,
			workflow.Description,
			pushCheck,
			prCheck))
	}

	return sb.String()
}
