package cmd

import (
	"fmt"
	"log"

	"github.com/octoberswimmer/utka/users"
	"github.com/octoberswimmer/utka/workspaces"
	"github.com/spf13/cobra"
)

var userManager *users.UserManager
var userWorkspaceManager *workspaces.WorkspaceManager

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage Asana users",
	Long:  `Commands for listing and retrieving information about Asana users.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
		userManager = users.NewUserManager(asanaClient)
		userWorkspaceManager = workspaces.NewWorkspaceManager(asanaClient)
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users in workspace(s)",
	Long:  `List all users in the specified workspace, or in all workspaces if --workspace is not specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspaceGID, _ := cmd.Flags().GetString("workspace")

		if workspaceGID != "" {
			// List users for specific workspace
			users, err := userManager.ListInWorkspace(workspaceGID)
			if err != nil {
				log.Fatalf("Failed to list users: %v", err)
			}

			if len(users) == 0 {
				fmt.Println("No users found in workspace")
				return
			}

			fmt.Println("Users in workspace:")
			for _, user := range users {
				fmt.Printf("  • %s (%s) - GID: %s\n", user.Name, user.Email, user.GID)
			}
		} else {
			// List users for all workspaces
			workspaces, err := userWorkspaceManager.List()
			if err != nil {
				log.Fatalf("Failed to list workspaces: %v", err)
			}

			if len(workspaces) == 0 {
				fmt.Println("No workspaces found")
				return
			}

			for _, workspace := range workspaces {
				fmt.Printf("\n%s (%s):\n", workspace.Name, workspace.GID)

				users, err := userManager.ListInWorkspace(workspace.GID)
				if err != nil {
					fmt.Printf("  Error listing users: %v\n", err)
					continue
				}

				if len(users) == 0 {
					fmt.Println("  No users found")
					continue
				}

				for _, user := range users {
					fmt.Printf("  • %s (%s) - GID: %s\n", user.Name, user.Email, user.GID)
				}
			}
		}
	},
}

func init() {
	userListCmd.Flags().String("workspace", "", "Workspace GID (optional - if not specified, lists users from all workspaces)")

	userCmd.AddCommand(userListCmd)

	rootCmd.AddCommand(userCmd)
}
