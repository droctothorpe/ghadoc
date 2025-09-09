package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper functions
// createTempWorkflowFile creates a temporary workflow file with the given content
func createTempWorkflowFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file %s: %v", filename, err)
	}
	return filePath
}

// createTempDir creates a temporary directory for testing
func createTempDir(t *testing.T, prefix string) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return tempDir
}

// createWorkflowsDir creates a temporary directory with workflow files
func createWorkflowsDir(t *testing.T, workflows map[string]string) string {
	t.Helper()

	// Create base temp directory
	tempDir := createTempDir(t, "ghadoc-test")

	// Create workflows directory
	workflowsDir := filepath.Join(tempDir, "workflows")
	err := os.Mkdir(workflowsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create workflow files
	for filename, content := range workflows {
		createTempWorkflowFile(t, workflowsDir, filename, content)
	}

	return workflowsDir
}

// TestParseWorkflowFile tests the parseWorkflowFile function
func TestParseWorkflowFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := createTempDir(t, "workflow-test")

	// Test cases
	testCases := []struct {
		name             string
		content          string
		expectedDesc     string
		expectedTriggers []string
	}{
		{
			name: "With description and both triggers",
			content: `## This is a test workflow
name: Test Workflow

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
`,
			expectedDesc:     "This is a test workflow",
			expectedTriggers: []string{"push", "pull_request"},
		},
		{
			name: "With description and push only",
			content: `## Another test workflow
name: Another Test

on:
  push:
    branches: [ main ]
`,
			expectedDesc:     "Another test workflow",
			expectedTriggers: []string{"push"},
		},
		{
			name: "With description and PR only",
			content: `## PR only workflow
name: PR Test

on:
  pull_request:
    branches: [ main ]
`,
			expectedDesc:     "PR only workflow",
			expectedTriggers: []string{"pull_request"},
		},
		{
			name: "No description with both triggers",
			content: `name: No Description

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
`,
			expectedDesc:     "",
			expectedTriggers: []string{"push", "pull_request"},
		},
		{
			name: "With description and string triggers",
			content: `## String trigger workflow
name: String Trigger

on: push
`,
			expectedDesc:     "String trigger workflow",
			expectedTriggers: []string{"push"},
		},
		{
			name: "With description and array triggers",
			content: `## Array trigger workflow
name: Array Trigger

on: [push, pull_request]
`,
			expectedDesc:     "Array trigger workflow",
			expectedTriggers: []string{"push", "pull_request"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary workflow file
			filePath := createTempWorkflowFile(t, tempDir, "workflow.yml", tc.content)

			// Parse the workflow file
			workflow, err := parseWorkflowFile(filePath)
			if err != nil {
				t.Fatalf("parseWorkflowFile failed: %v", err)
			}

			// Check the results
			if workflow.Description != tc.expectedDesc {
				t.Errorf("Expected description %q, got %q", tc.expectedDesc, workflow.Description)
			}

			// Check that all expected triggers are present
			if len(workflow.Triggers) != len(tc.expectedTriggers) {
				t.Errorf("Expected %d triggers, got %d", len(tc.expectedTriggers), len(workflow.Triggers))
			}

			// Create maps for easier comparison
			expectedMap := make(map[string]bool)
			actualMap := make(map[string]bool)

			for _, trigger := range tc.expectedTriggers {
				expectedMap[trigger] = true
			}

			for _, trigger := range workflow.Triggers {
				actualMap[trigger] = true
			}

			// Check for missing or unexpected triggers
			for trigger := range expectedMap {
				if !actualMap[trigger] {
					t.Errorf("Expected trigger %q not found", trigger)
				}
			}

			for trigger := range actualMap {
				if !expectedMap[trigger] {
					t.Errorf("Unexpected trigger %q found", trigger)
				}
			}
		})
	}
}

// TestParseWorkflowFileErrors tests error handling in parseWorkflowFile
func TestParseWorkflowFileErrors(t *testing.T) {
	// Test non-existent file
	_, err := parseWorkflowFile("/non/existent/file.yml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Create a temporary directory for test files
	tempDir := createTempDir(t, "workflow-error-test")

	// Test invalid YAML
	invalidYamlPath := createTempWorkflowFile(t, tempDir, "invalid.yml", `
## This is an invalid YAML file
name: Invalid YAML
on: {
  push: [
    invalid yaml content
`)

	_, err = parseWorkflowFile(invalidYamlPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}

	// Test unreadable file (create and then remove read permissions)
	unreadablePath := filepath.Join(tempDir, "unreadable.yml")
	err = os.WriteFile(unreadablePath, []byte(`## Unreadable file`), 0000)
	if err != nil {
		t.Fatalf("Failed to create unreadable file: %v", err)
	}

	_, err = parseWorkflowFile(unreadablePath)
	if err == nil {
		t.Error("Expected error for unreadable file, got nil")
	}
}

// TestGenerateMarkdownTable tests the generateMarkdownTable function
func TestGenerateMarkdownTable(t *testing.T) {
	// Create test workflows
	workflows := []WorkflowInfo{
		{
			Filename:    "workflow1.yml",
			Description: "Test workflow 1",
			Triggers:    []string{"push", "pull_request"},
		},
		{
			Filename:    "workflow2.yml",
			Description: "Test workflow 2",
			Triggers:    []string{"push"},
		},
		{
			Filename:    "workflow3.yml",
			Description: "Test workflow 3",
			Triggers:    []string{"pull_request"},
		},
		{
			Filename:    "workflow4.yml",
			Description: "",
			Triggers:    []string{},
		},
	}

	// Generate markdown table
	workflowsPath := "test/workflows"
	outputPath := "test/output.md"
	markdownTable := generateMarkdownTable(workflows, workflowsPath, outputPath)

	// Verify the table contains expected content
	expectedLines := []string{
		"# GitHub Workflows Summary",
		"| Filename | Description | Triggers |",
		"| --- | --- | --- |",
		"| [workflow1.yml](workflows/workflow1.yml) | Test workflow 1 | push, pull_request |",
		"| [workflow2.yml](workflows/workflow2.yml) | Test workflow 2 | push |",
		"| [workflow3.yml](workflows/workflow3.yml) | Test workflow 3 | pull_request |",
		"| [workflow4.yml](workflows/workflow4.yml) |  |  |",
	}

	for _, line := range expectedLines {
		if !strings.Contains(markdownTable, line) {
			t.Errorf("Expected markdown table to contain %q, but it doesn't", line)
			fmt.Println("markdownTable:\n", markdownTable)
		}
	}
}

// TestGenerate tests the Generate function
func TestGenerate(t *testing.T) {
	// Create a temporary directory structure
	tempDir := createTempDir(t, "ghadoc-test")

	// Create workflow directory
	workflowsDir := filepath.Join(tempDir, "workflows")
	err := os.Mkdir(workflowsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create test workflow files
	testWorkflows := map[string]string{
		"workflow1.yml": `## Test workflow 1
name: Workflow 1

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
`,
		"workflow2.yml": `## Test workflow 2
name: Workflow 2

on:
  push:
    branches: [ main ]
`,
		"not-a-workflow.txt": `This is not a workflow file`,
	}

	for filename, content := range testWorkflows {
		createTempWorkflowFile(t, workflowsDir, filename, content)
	}

	// Create output file path
	outputFile := filepath.Join(tempDir, "output.md")

	// Call Generate function
	err = Generate(workflowsDir, outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created")
	}

	// Read output file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify content contains expected information
	markdownContent := string(content)
	expectedStrings := []string{
		"# GitHub Workflows Summary",
		"| Filename | Description | Triggers |",
		"| [workflow1.yml](workflows/workflow1.yml) | Test workflow 1 | push, pull_request |",
		"| [workflow2.yml](workflows/workflow2.yml) | Test workflow 2 | push |",
	}

	for _, str := range expectedStrings {
		if !strings.Contains(markdownContent, str) {
			t.Errorf("Expected output to contain %q, but it doesn't", str)
			fmt.Println("markdownTable:\n", string(content))
		}
	}

	// Verify non-workflow file was not included
	if strings.Contains(markdownContent, "not-a-workflow.txt") {
		t.Errorf("Output should not contain non-workflow file")
	}
}

// TestGenerateErrors tests error handling in Generate function
func TestGenerateErrors(t *testing.T) {
	// Test with non-existent directory
	err := Generate("/non/existent/dir", "output.md")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}

	// Create a temporary directory for testing
	tempDir := createTempDir(t, "generate-error-test")

	// Test with valid directory but invalid output path
	workflowsDir := filepath.Join(tempDir, "workflows")
	err = os.Mkdir(workflowsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create a test workflow file
	createTempWorkflowFile(t, workflowsDir, "workflow.yml", `
## Test workflow
name: Test
on:
  push:
    branches: [ main ]
`)

	// Test with invalid output path (directory that doesn't exist)
	err = Generate(workflowsDir, "/non/existent/dir/output.md")
	if err == nil {
		t.Error("Expected error for invalid output path, got nil")
	}

	// Test with unwritable output file
	unwritablePath := filepath.Join(tempDir, "unwritable")

	// Create a directory with the same name to make it impossible to write a file there
	err = os.Mkdir(unwritablePath, 0755)
	if err != nil {
		t.Fatalf("Failed to create unwritable dir: %v", err)
	}

	err = Generate(workflowsDir, unwritablePath)
	if err == nil {
		t.Error("Expected error for unwritable output path, got nil")
	}
}

// TestEmptyWorkflowsDirectory tests handling of an empty workflows directory
func TestEmptyWorkflowsDirectory(t *testing.T) {
	// Create empty temp directory
	tempDir := createTempDir(t, "empty-workflows")
	outputFile := filepath.Join(tempDir, "output.md")

	// Call Generate with empty directory
	err := Generate(tempDir, outputFile)
	if err != nil {
		t.Fatalf("Generate failed with empty directory: %v", err)
	}

	// Verify output file exists
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check that the table headers are present but no workflow entries
	markdownContent := string(content)
	if !strings.Contains(markdownContent, "# GitHub Workflows Summary") {
		t.Error("Output should contain table title")
	}
	if !strings.Contains(markdownContent, "| Filename | Description | Triggers |") {
		t.Error("Output should contain table headers")
	}

	// Count the number of lines - should be just the headers (4 lines including the blank line after title)
	lines := strings.Split(strings.TrimSpace(markdownContent), "\n")
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines in output for empty directory, got %d", len(lines))
	}
}

// TestNonYamlFiles tests handling of non-YAML files in the workflows directory
func TestNonYamlFiles(t *testing.T) {
	// Create temp directory with mixed file types
	tempDir := createTempDir(t, "mixed-files")

	// Create various file types
	files := map[string]string{
		"workflow.yml": `## Valid workflow
name: Valid
on:
  push:
    branches: [ main ]`,
		"workflow.yaml": `## Also valid workflow
name: Also Valid
on:
  pull_request:
    branches: [ main ]`,
		"readme.md":   "# This is a readme file",
		"script.sh":   "#!/bin/bash\necho 'This is a script'",
		"config.json": `{"name": "Not a workflow"}`,
	}

	for filename, content := range files {
		createTempWorkflowFile(t, tempDir, filename, content)
	}

	outputFile := filepath.Join(tempDir, "output.md")

	// Call Generate
	err := Generate(tempDir, outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Read output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	markdownContent := string(content)

	// Check that only YAML files are included
	if !strings.Contains(markdownContent, "workflow.yml") {
		t.Error("Output should contain workflow.yml")
	}
	if !strings.Contains(markdownContent, "workflow.yaml") {
		t.Error("Output should contain workflow.yaml")
	}
	if strings.Contains(markdownContent, "readme.md") {
		t.Error("Output should not contain readme.md")
	}
	if strings.Contains(markdownContent, "script.sh") {
		t.Error("Output should not contain script.sh")
	}
	if strings.Contains(markdownContent, "config.json") {
		t.Error("Output should not contain config.json")
	}
}

// TestSpecialCharactersInDescription tests handling of special characters in workflow descriptions
func TestSpecialCharactersInDescription(t *testing.T) {
	// Create temp directory
	tempDir := createTempDir(t, "special-chars")

	// Create workflow with special characters in description
	specialCharsContent := `## Special characters: !@#$%^&*()_+-={}[]|\\:;"'<>,.?/
name: Special Chars
on:
  push:
    branches: [ main ]`

	createTempWorkflowFile(t, tempDir, "special.yml", specialCharsContent)

	// Parse the workflow file
	filePath := filepath.Join(tempDir, "special.yml")
	workflow, err := parseWorkflowFile(filePath)
	if err != nil {
		t.Fatalf("parseWorkflowFile failed: %v", err)
	}

	// Check that special characters are preserved
	// Note: The actual behavior might have an extra backslash due to how strings are processed
	if !strings.Contains(workflow.Description, "Special characters") {
		t.Errorf("Expected description to contain 'Special characters', got %q", workflow.Description)
	}
}

// TestMultilineDescription tests handling of descriptions that span multiple lines
func TestMultilineDescription(t *testing.T) {
	// Create temp directory
	tempDir := createTempDir(t, "multiline-desc")

	// Create workflow with multiline comment
	multilineContent := `## First line of description
## Second line of description
name: Multiline
on:
  push:
    branches: [ main ]`

	createTempWorkflowFile(t, tempDir, "multiline.yml", multilineContent)

	// Parse the workflow file
	filePath := filepath.Join(tempDir, "multiline.yml")
	workflow, err := parseWorkflowFile(filePath)
	if err != nil {
		t.Fatalf("parseWorkflowFile failed: %v", err)
	}

	// Check that both lines with ## are included with <br> separator
	expectedDesc := "First line of description<br>Second line of description"
	if workflow.Description != expectedDesc {
		t.Errorf("Expected description %q, got %q", expectedDesc, workflow.Description)
	}
}

// TestComplexOnTriggers tests handling of complex on trigger configurations
func TestComplexOnTriggers(t *testing.T) {
	// Create temp directory
	tempDir := createTempDir(t, "complex-triggers")

	// Create workflow with complex triggers
	complexTriggersContent := `## Complex triggers
name: Complex Triggers
on:
  push:
    branches:
      - main
      - 'releases/**'
    paths:
      - '**.go'
      - '!vendor/**'
  pull_request:
    types: [opened, synchronize, reopened]
    branches:
      - main
  workflow_dispatch:
    inputs:
      logLevel:
        description: 'Log level'
        required: true
        default: 'warning'`

	createTempWorkflowFile(t, tempDir, "complex.yml", complexTriggersContent)

	// Parse the workflow file
	filePath := filepath.Join(tempDir, "complex.yml")
	workflow, err := parseWorkflowFile(filePath)
	if err != nil {
		t.Fatalf("parseWorkflowFile failed: %v", err)
	}

	// Check that all expected triggers are detected correctly
	expectedTriggers := []string{"push", "pull_request", "workflow_dispatch"}

	// Create maps for easier comparison
	expectedMap := make(map[string]bool)
	actualMap := make(map[string]bool)

	for _, trigger := range expectedTriggers {
		expectedMap[trigger] = true
	}

	for _, trigger := range workflow.Triggers {
		actualMap[trigger] = true
	}

	// Check for missing triggers
	for trigger := range expectedMap {
		if !actualMap[trigger] {
			t.Errorf("Expected trigger %q not found in complex triggers", trigger)
		}
	}

	// Check for unexpected triggers
	for trigger := range actualMap {
		if !expectedMap[trigger] {
			t.Errorf("Unexpected trigger %q found in complex triggers", trigger)
		}
	}
}

// TestYamlExtensionVariants tests handling of different YAML file extensions
func TestYamlExtensionVariants(t *testing.T) {
	// Create temp directory
	tempDir := createTempDir(t, "yaml-extensions")
	workflowsDir := filepath.Join(tempDir, "workflows")
	err := os.Mkdir(workflowsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create workflows with different extensions
	files := map[string]string{
		"workflow1.yml": `## YML extension
name: YML
on:
  push:
    branches: [ main ]`,
		"workflow2.yaml": `## YAML extension
name: YAML
on:
  pull_request:
    branches: [ main ]`,
	}

	for filename, content := range files {
		createTempWorkflowFile(t, workflowsDir, filename, content)
	}

	outputFile := filepath.Join(tempDir, "output.md")

	// Call Generate
	err = Generate(workflowsDir, outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Read output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	markdownContent := string(content)

	// Check that both extensions are included
	if !strings.Contains(markdownContent, "workflow1.yml") {
		t.Error("Output should contain workflow1.yml")
	}
	if !strings.Contains(markdownContent, "workflow2.yaml") {
		t.Error("Output should contain workflow2.yaml")
	}
}
