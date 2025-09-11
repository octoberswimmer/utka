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

var webhookEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a webhook (placeholder for future enhancements)",
	Long:  `Edit a webhook's properties by its GID. Currently preserves existing configuration. Use 'webhook filter edit' to modify filters.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		// Get current webhook
		webhook, err := webhookManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get webhook: %v", err)
		}

		// For now, just return the current webhook
		// Future enhancements can modify other properties here
		fmt.Println("Current webhook configuration (use 'webhook filter edit' to modify filters):")
		printJSON(webhook)
	},
}

var webhookFilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Manage webhook filters",
	Long:  `Commands for managing webhook filters.`,
}

var webhookFilterAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a filter to a webhook",
	Long:  `Add a new filter to a specific webhook by its GID.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		action, _ := cmd.Flags().GetString("action")
		resourceType, _ := cmd.Flags().GetString("resource-type")
		resourceSubtype, _ := cmd.Flags().GetString("resource-subtype")

		// Handle "all" action which means no action filter
		if action == "all" {
			action = ""
		}

		// Get current webhook to see existing filters
		webhook, err := webhookManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get webhook: %v", err)
		}

		// Add the new filter to existing filters
		newFilter := webhooks.WebhookFilter{
			Action:          action,
			ResourceType:    resourceType,
			ResourceSubtype: resourceSubtype,
		}

		filters := append(webhook.Filters, newFilter)

		// Update the webhook with the new filters list
		webhook, err = webhookManager.UpdateFilters(gid, filters)
		if err != nil {
			log.Fatalf("Failed to add filter to webhook: %v", err)
		}

		fmt.Println("Filter added successfully:")
		printJSON(webhook)
	},
}

var webhookFilterEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit webhook filters",
	Long:  `Edit filters for a specific webhook by its GID.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		// Get current webhook to see existing filters
		webhook, err := webhookManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get webhook: %v", err)
		}

		filters := webhook.Filters

		// If there are multiple filters, prompt for which one to edit
		if len(filters) > 1 {
			fmt.Println("Multiple filters found:")
			for i, filter := range filters {
				fmt.Printf("%d. Action: %s, Resource Type: %s, Resource Subtype: %s\n",
					i+1, filter.Action, filter.ResourceType, filter.ResourceSubtype)
			}

			var choice int
			fmt.Print("Enter the number of the filter to edit (or 0 to add a new filter): ")
			_, err := fmt.Scanf("%d", &choice)
			if err != nil {
				log.Fatalf("Failed to read input: %v", err)
			}

			if choice == 0 {
				// Add a new filter
				action, _ := cmd.Flags().GetString("action")
				resourceType, _ := cmd.Flags().GetString("resource-type")
				resourceSubtype, _ := cmd.Flags().GetString("resource-subtype")

				// Handle "all" action which means no action filter
				if action == "all" {
					action = ""
				}

				newFilter := webhooks.WebhookFilter{
					Action:          action,
					ResourceType:    resourceType,
					ResourceSubtype: resourceSubtype,
				}
				filters = append(filters, newFilter)
			} else if choice > 0 && choice <= len(filters) {
				// Edit existing filter
				action, _ := cmd.Flags().GetString("action")
				resourceType, _ := cmd.Flags().GetString("resource-type")
				resourceSubtype, _ := cmd.Flags().GetString("resource-subtype")

				if action != "" {
					if action == "all" {
						filters[choice-1].Action = ""
					} else {
						filters[choice-1].Action = action
					}
				}
				if resourceType != "" {
					filters[choice-1].ResourceType = resourceType
				}
				if resourceSubtype != "" {
					filters[choice-1].ResourceSubtype = resourceSubtype
				}
			} else {
				log.Fatal("Invalid choice")
			}
		} else if len(filters) == 1 {
			// Single filter, edit it directly
			action, _ := cmd.Flags().GetString("action")
			resourceType, _ := cmd.Flags().GetString("resource-type")
			resourceSubtype, _ := cmd.Flags().GetString("resource-subtype")

			if action != "" {
				if action == "all" {
					filters[0].Action = ""
				} else {
					filters[0].Action = action
				}
			}
			if resourceType != "" {
				filters[0].ResourceType = resourceType
			}
			if resourceSubtype != "" {
				filters[0].ResourceSubtype = resourceSubtype
			}
		} else {
			// No filters, add a new one
			action, _ := cmd.Flags().GetString("action")
			resourceType, _ := cmd.Flags().GetString("resource-type")
			resourceSubtype, _ := cmd.Flags().GetString("resource-subtype")

			if action == "" && resourceType == "" && resourceSubtype == "" {
				log.Fatal("No filters to edit and no new filter values provided")
			}

			// Handle "all" action which means no action filter
			if action == "all" {
				action = ""
			}

			newFilter := webhooks.WebhookFilter{
				Action:          action,
				ResourceType:    resourceType,
				ResourceSubtype: resourceSubtype,
			}
			filters = []webhooks.WebhookFilter{newFilter}
		}

		// Update the webhook with modified filters (only filters, not active status)
		webhook, err = webhookManager.UpdateFilters(gid, filters)
		if err != nil {
			log.Fatalf("Failed to update webhook filters: %v", err)
		}

		printJSON(webhook)
	},
}

var webhookFilterDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a filter from a webhook",
	Long:  `Delete a specific filter from a webhook by its GID.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		// Get current webhook to see existing filters
		webhook, err := webhookManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get webhook: %v", err)
		}

		filters := webhook.Filters

		if len(filters) == 0 {
			fmt.Println("No filters found on this webhook")
			return
		}

		if len(filters) == 1 {
			// Single filter, ask for confirmation
			filter := filters[0]
			fmt.Printf("Delete filter: Action: %s, Resource Type: %s, Resource Subtype: %s? (y/N): ",
				filter.Action, filter.ResourceType, filter.ResourceSubtype)

			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Filter deletion cancelled")
				return
			}

			filters = []webhooks.WebhookFilter{}
		} else {
			// Multiple filters, prompt for which one to delete
			fmt.Println("Multiple filters found:")
			for i, filter := range filters {
				fmt.Printf("%d. Action: %s, Resource Type: %s, Resource Subtype: %s\n",
					i+1, filter.Action, filter.ResourceType, filter.ResourceSubtype)
			}

			var choice int
			fmt.Print("Enter the number of the filter to delete: ")
			_, err := fmt.Scanf("%d", &choice)
			if err != nil {
				log.Fatalf("Failed to read input: %v", err)
			}

			if choice < 1 || choice > len(filters) {
				log.Fatal("Invalid choice")
			}

			// Remove the selected filter
			filters = append(filters[:choice-1], filters[choice:]...)
		}

		// Update the webhook with the modified filters list
		webhook, err = webhookManager.UpdateFilters(gid, filters)
		if err != nil {
			log.Fatalf("Failed to delete filter from webhook: %v", err)
		}

		fmt.Println("Filter deleted successfully:")
		printJSON(webhook)
	},
}

var webhookStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show webhook delivery status and health details",
	Long:  `Display detailed information about webhook delivery success, failures, and retry status.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Webhook GID is required")
		}

		webhook, err := webhookManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get webhook: %v", err)
		}

		// Display webhook status information
		fmt.Printf("Webhook Status for GID: %s\n", webhook.GID)
		fmt.Printf("=====================================\n\n")

		// Basic info
		fmt.Printf("Target URL: %s\n", webhook.Target)
		fmt.Printf("Active: %v\n", webhook.Active)
		if webhook.Resource != nil {
			fmt.Printf("Resource: %s (%s)", webhook.Resource.GID, webhook.Resource.ResourceType)
			if webhook.Resource.Name != "" {
				fmt.Printf(" - %s", webhook.Resource.Name)
			}
			fmt.Println()
		}
		fmt.Printf("Created: %s\n", webhook.CreatedAt)
		fmt.Printf("Is Workspace Webhook: %v\n\n", webhook.IsWorkspaceWebhook)

		// Delivery status
		fmt.Printf("Delivery Status\n")
		fmt.Printf("---------------\n")

		if webhook.LastSuccessAt != "" {
			fmt.Printf("Last Success: %s\n", webhook.LastSuccessAt)
		} else {
			fmt.Printf("Last Success: Never\n")
		}

		if webhook.LastFailureAt != "" {
			fmt.Printf("Last Failure: %s\n", webhook.LastFailureAt)
			fmt.Printf("Retry Count: %d\n", webhook.DeliveryRetryCount)

			if webhook.NextAttemptAfter != "" {
				fmt.Printf("Next Retry: %s\n", webhook.NextAttemptAfter)
			}

			if webhook.FailureDeletionTimestamp != "" {
				fmt.Printf("Will be deleted if failing until: %s\n", webhook.FailureDeletionTimestamp)
			}

			if webhook.LastFailureContent != "" {
				fmt.Printf("\nLast Failure Details:\n")
				fmt.Printf("--------------------\n")
				fmt.Println(webhook.LastFailureContent)
			}
		} else {
			fmt.Printf("Last Failure: None\n")
		}

		// Show filters if present
		if len(webhook.Filters) > 0 {
			fmt.Printf("\nActive Filters\n")
			fmt.Printf("--------------\n")
			for i, filter := range webhook.Filters {
				fmt.Printf("%d. Resource Type: %s", i+1, filter.ResourceType)
				if filter.ResourceSubtype != "" {
					fmt.Printf(", Subtype: %s", filter.ResourceSubtype)
				}
				if filter.Action != "" {
					fmt.Printf(", Action: %s", filter.Action)
				}
				if len(filter.Fields) > 0 {
					fmt.Printf(", Fields: %v", filter.Fields)
				}
				fmt.Println()
			}
		}

		// Option to show full JSON
		if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
			fmt.Printf("\nFull JSON Response:\n")
			fmt.Printf("------------------\n")
			printJSON(webhook)
		}
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

	webhookEditCmd.Flags().String("gid", "", "Webhook GID")
	webhookEditCmd.MarkFlagRequired("gid")

	webhookFilterAddCmd.Flags().String("gid", "", "Webhook GID")
	webhookFilterAddCmd.Flags().String("action", "", "Filter by action (changed, added, removed, deleted, undeleted, all) - 'all' means no action filter")
	webhookFilterAddCmd.Flags().String("resource-type", "", "Resource type for the filter")
	webhookFilterAddCmd.Flags().String("resource-subtype", "", "Resource subtype for the filter")
	webhookFilterAddCmd.MarkFlagRequired("gid")

	webhookFilterEditCmd.Flags().String("gid", "", "Webhook GID")
	webhookFilterEditCmd.Flags().String("action", "", "Filter by action (changed, added, removed, deleted, undeleted, all) - 'all' removes the action filter")
	webhookFilterEditCmd.Flags().String("resource-type", "", "Resource type for the filter")
	webhookFilterEditCmd.Flags().String("resource-subtype", "", "Resource subtype for the filter")
	webhookFilterEditCmd.MarkFlagRequired("gid")

	webhookFilterDeleteCmd.Flags().String("gid", "", "Webhook GID")
	webhookFilterDeleteCmd.MarkFlagRequired("gid")

	webhookStatusCmd.Flags().String("gid", "", "Webhook GID")
	webhookStatusCmd.Flags().Bool("json", false, "Show full JSON response")
	webhookStatusCmd.MarkFlagRequired("gid")

	webhookFilterCmd.AddCommand(webhookFilterAddCmd)
	webhookFilterCmd.AddCommand(webhookFilterEditCmd)
	webhookFilterCmd.AddCommand(webhookFilterDeleteCmd)

	webhookCmd.AddCommand(webhookListCmd)
	webhookCmd.AddCommand(webhookGetCmd)
	webhookCmd.AddCommand(webhookCreateCmd)
	webhookCmd.AddCommand(webhookDeleteCmd)
	webhookCmd.AddCommand(webhookEditCmd)
	webhookCmd.AddCommand(webhookFilterCmd)
	webhookCmd.AddCommand(webhookStatusCmd)

	rootCmd.AddCommand(webhookCmd)
}

func printJSON(v interface{}) {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println(string(output))
}
