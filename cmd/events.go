package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Get and monitor Asana events",
	Long:  `Commands for retrieving and monitoring events from Asana resources.`,
}

var eventsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get events for a resource",
	Long: `Retrieve events for a specific resource with optional sync token.

If you get a 'Sync token invalid or too old' error, omit the --sync flag to fetch 
the full dataset and get a new sync token.`,
	Run: func(cmd *cobra.Command, args []string) {
		resource, _ := cmd.Flags().GetString("resource")
		syncToken, _ := cmd.Flags().GetString("sync")

		if resource == "" {
			log.Fatal("Resource GID is required")
		}

		events, err := eventManager.GetByResource(resource, syncToken)
		if err != nil {
			log.Fatalf("Failed to get events: %v", err)
		}

		printJSON(events)
	},
}

var eventsPollCmd = &cobra.Command{
	Use:   "poll",
	Short: "Poll events continuously",
	Long:  `Continuously poll for events from a specific resource.`,
	Run: func(cmd *cobra.Command, args []string) {
		resource, _ := cmd.Flags().GetString("resource")
		syncToken, _ := cmd.Flags().GetString("sync")
		interval, _ := cmd.Flags().GetDuration("interval")

		if resource == "" {
			log.Fatal("Resource GID is required")
		}

		fmt.Printf("Starting to poll events for resource %s (interval: %v)...\n", resource, interval)
		eventsChan, errorsChan := eventManager.Poll(resource, syncToken, interval)

		for {
			select {
			case event, ok := <-eventsChan:
				if !ok {
					fmt.Println("Event channel closed")
					return
				}
				printJSON(event)
			case err, ok := <-errorsChan:
				if !ok {
					fmt.Println("Error channel closed")
					return
				}
				log.Printf("Error polling events: %v", err)
			}
		}
	},
}

var eventsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Initialize or refresh sync token for a resource",
	Long: `Initialize or refresh the sync token for a resource. Use this when you get
'Sync token invalid or too old' errors. The command will fetch the current state
and return a fresh sync token for future polling.`,
	Run: func(cmd *cobra.Command, args []string) {
		resource, _ := cmd.Flags().GetString("resource")

		if resource == "" {
			log.Fatal("Resource GID is required")
		}

		// Use InitializeSync to properly handle 412 errors
		events, err := eventManager.InitializeSync(resource)
		if err != nil {
			log.Fatalf("Failed to initialize sync: %v", err)
		}

		fmt.Printf("Sync initialized for resource %s\n", resource)
		if events.Sync != "" {
			fmt.Printf("New sync token: %s\n", events.Sync)
			fmt.Printf("\nUse this token with: utka events get --resource %s --sync %s\n", resource, events.Sync)
		}
		fmt.Printf("Events in current state: %d\n", len(events.Data))
	},
}

func init() {
	eventsGetCmd.Flags().String("resource", "", "Resource GID")
	eventsGetCmd.Flags().String("sync", "", "Sync token")
	eventsGetCmd.MarkFlagRequired("resource")

	eventsSyncCmd.Flags().String("resource", "", "Resource GID")
	eventsSyncCmd.MarkFlagRequired("resource")

	eventsPollCmd.Flags().String("resource", "", "Resource GID (project, task, etc.)")
	eventsPollCmd.Flags().String("sync", "", "Initial sync token")
	eventsPollCmd.Flags().Duration("interval", 5*time.Second, "Poll interval")
	eventsPollCmd.MarkFlagRequired("resource")

	eventsCmd.AddCommand(eventsGetCmd)
	eventsCmd.AddCommand(eventsSyncCmd)
	eventsCmd.AddCommand(eventsPollCmd)

	rootCmd.AddCommand(eventsCmd)
}
