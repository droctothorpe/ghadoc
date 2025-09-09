package cmd

import (
	"fmt"

	"github.com/droctothorpe/gha-docs/internal/generate"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen"},
	Short:   "Generate markdown documentation for GitHub Actions workflows",
	Long: `Generate a markdown table summarizing GitHub Actions workflows in a specified directory.

The table includes the following columns:
- Filename: Name of the workflow file with a link to the file
- Description: Extracted from the first line starting with "##" in the workflow file
- On Push: Indicates if the workflow runs on push events
- On PR: Indicates if the workflow runs on pull request events

Output is written to workflows.md in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowDir, _ := cmd.Flags().GetString("workflows")
		output, _ := cmd.Flags().GetString("output")

		err := generate.Generate(workflowDir, output)
		if err != nil {
			fmt.Printf("Error generating workflow documentation: %v\n", err)
		}
	},
}

func init() {
	generateCmd.Flags().StringP("workflows", "w", ".", "Directory containing GitHub workflow files")
	generateCmd.Flags().StringP("output", "o", "./workflows.md", "Output file for the markdown table")
	generateCmd.MarkFlagRequired("workflows")
	rootCmd.AddCommand(generateCmd)
}
