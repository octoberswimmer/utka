package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/octoberswimmer/utka/webhooks"
	"github.com/spf13/cobra"
)

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage Asana webhooks",
	Long:  `Commands for creating, listing, updating, and deleting Asana webhooks.`,
}

var webhookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	Long:  `List all webhooks, optionally filtered by workspace or resource.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspace, _ := cmd.Flags().GetString("workspace")
		resource, _ := cmd.Flags().GetString("resource")

		webhooks, err := webhookManager.List(workspace, resource)
		if err != nil {
			log.Fatalf("Failed to list webhooks: %v", err)
		}

		printJSON(webhooks)
	},
}

var webhookGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific webhook",
	Long:  `Retrieve details of a specific webhook by its GID.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		webhook, err := webhookManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get webhook: %v", err)
		}

		printJSON(webhook)
	},
}

var webhookCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new webhook",
	Long:  `Create a new webhook for a specific resource with a target URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		resource, _ := cmd.Flags().GetString("resource")
		target, _ := cmd.Flags().GetString("target")

		if resource == "" || target == "" {
			log.Fatal("Both resource GID and target URL are required")
		}

		filters := []webhooks.WebhookFilter{}

		webhook, err := webhookManager.Create(resource, target, filters)
		if err != nil {
			log.Fatalf("Failed to create webhook: %v", err)
		}

		printJSON(webhook)
	},
}

var webhookDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a webhook",
	Long:  `Delete a specific webhook by its GID.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		err := webhookManager.Delete(gid)
		if err != nil {
			log.Fatalf("Failed to delete webhook: %v", err)
		}

		fmt.Println("Webhook deleted successfully")
	},
}

var webhookUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a webhook",
	Long:  `Update a webhook's filters by its GID.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		filters := []webhooks.WebhookFilter{}

		webhook, err := webhookManager.Update(gid, filters)
		if err != nil {
			log.Fatalf("Failed to update webhook: %v", err)
		}

		printJSON(webhook)
	},
}

func init() {
	webhookListCmd.Flags().String("workspace", "", "Workspace GID")
	webhookListCmd.Flags().String("resource", "", "Resource GID")

	webhookGetCmd.Flags().String("gid", "", "Webhook GID")
	webhookGetCmd.MarkFlagRequired("gid")

	webhookCreateCmd.Flags().String("resource", "", "Resource GID")
	webhookCreateCmd.Flags().String("target", "", "Target URL")
	webhookCreateCmd.MarkFlagRequired("resource")
	webhookCreateCmd.MarkFlagRequired("target")

	webhookDeleteCmd.Flags().String("gid", "", "Webhook GID")
	webhookDeleteCmd.MarkFlagRequired("gid")

	webhookUpdateCmd.Flags().String("gid", "", "Webhook GID")
	webhookUpdateCmd.MarkFlagRequired("gid")

	webhookCmd.AddCommand(webhookListCmd)
	webhookCmd.AddCommand(webhookGetCmd)
	webhookCmd.AddCommand(webhookCreateCmd)
	webhookCmd.AddCommand(webhookDeleteCmd)
	webhookCmd.AddCommand(webhookUpdateCmd)

	rootCmd.AddCommand(webhookCmd)
}

func printJSON(v interface{}) {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println(string(output))
}
