package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/octoberswimmer/utka/projects"
	"github.com/spf13/cobra"
)

var projectManager *projects.ProjectManager

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage Asana projects",
	Long:  `Commands for listing and retrieving information about Asana projects.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
		projectManager = projects.NewProjectManager(asanaClient)
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long:  `List all projects in a workspace or team.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspace, _ := cmd.Flags().GetString("workspace")
		team, _ := cmd.Flags().GetString("team")
		archived, _ := cmd.Flags().GetBool("archived")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if workspace == "" && team == "" {
			log.Fatal("Either workspace or team GID is required")
		}

		if workspace != "" && team != "" {
			log.Fatal("Please specify either workspace or team, not both")
		}

		var projectsList []projects.Project
		var err error

		if workspace != "" {
			projectsList, err = projectManager.ListByWorkspace(workspace, archived, limit)
		} else {
			projectsList, err = projectManager.ListByTeam(team, archived, limit)
		}

		if err != nil {
			log.Fatalf("Failed to list projects: %v", err)
		}

		if jsonOutput {
			printJSON(projectsList)
			return
		}

		// Pretty print projects
		if len(projectsList) == 0 {
			fmt.Println("No projects found")
			return
		}

		fmt.Printf("Found %d project(s):\n\n", len(projectsList))
		for _, project := range projectsList {
			status := ""
			if project.Archived {
				status = " [ARCHIVED]"
			}

			fmt.Printf("• %s%s\n", project.Name, status)
			fmt.Printf("  GID: %s\n", project.GID)

			if project.Color != "" {
				fmt.Printf("  Color: %s\n", project.Color)
			}

			if project.Owner != nil && project.Owner.Name != "" {
				fmt.Printf("  Owner: %s\n", project.Owner.Name)
			}

			if project.CurrentStatus != nil && project.CurrentStatus.Title != "" {
				fmt.Printf("  Status: %s (%s)\n", project.CurrentStatus.Title, project.CurrentStatus.Color)
			}

			if project.DueDate != "" {
				fmt.Printf("  Due: %s\n", project.DueDate)
			}

			if project.Notes != "" {
				// Truncate notes if too long
				notes := strings.ReplaceAll(project.Notes, "\n", " ")
				if len(notes) > 100 {
					notes = notes[:97] + "..."
				}
				fmt.Printf("  Notes: %s\n", notes)
			}

			fmt.Println()
		}
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get project details",
	Long:  `Retrieve detailed information about a specific project.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Project GID is required")
		}

		project, err := projectManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get project: %v", err)
		}

		printJSON(project)
	},
}

func init() {
	projectListCmd.Flags().String("workspace", "", "Workspace GID")
	projectListCmd.Flags().String("team", "", "Team GID")
	projectListCmd.Flags().Bool("archived", false, "Include archived projects")
	projectListCmd.Flags().Int("limit", 0, "Limit number of results (0 for all)")
	projectListCmd.Flags().Bool("json", false, "Output as JSON")

	projectGetCmd.Flags().String("gid", "", "Project GID")
	projectGetCmd.MarkFlagRequired("gid")

	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)

	rootCmd.AddCommand(projectCmd)
}
