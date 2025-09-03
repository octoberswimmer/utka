package cmd

import (
	"fmt"
	"log"

	"github.com/octoberswimmer/utka/workspaces"
	"github.com/spf13/cobra"
)

var workspaceManager *workspaces.WorkspaceManager

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage Asana workspaces",
	Long:  `Commands for listing and retrieving information about Asana workspaces.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
		workspaceManager = workspaces.NewWorkspaceManager(asanaClient)
	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `List all workspaces accessible with your personal access token.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspaces, err := workspaceManager.List()
		if err != nil {
			log.Fatalf("Failed to list workspaces: %v", err)
		}

		if len(workspaces) == 0 {
			fmt.Println("No workspaces found")
			return
		}

		fmt.Println("Available workspaces:")
		for _, ws := range workspaces {
			orgType := "Workspace"
			if ws.IsOrganization {
				orgType = "Organization"
			}
			fmt.Printf("  • %s (%s) - GID: %s\n", ws.Name, orgType, ws.GID)
		}
	},
}

var workspaceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get workspace details",
	Long:  `Retrieve detailed information about a specific workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Workspace GID is required")
		}

		workspace, err := workspaceManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get workspace: %v", err)
		}

		printJSON(workspace)
	},
}

func init() {
	workspaceGetCmd.Flags().String("gid", "", "Workspace GID")
	workspaceGetCmd.MarkFlagRequired("gid")

	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceGetCmd)

	rootCmd.AddCommand(workspaceCmd)
}
