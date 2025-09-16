package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/expr-lang/expr"
	eventsLib "github.com/octoberswimmer/utka/events"
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
the full dataset and get a new sync token.

You can filter events using an expression with the -f flag. The expression has access
to the 'event' variable with all its fields. Examples:
  -f 'event.action == "changed"'
  -f 'event.action == "changed" && event.resource.resource_subtype == "default_task"'
  -f 'event.change.field == "name"'
  -f 'event.user.name == "John Doe"'
  -f 'event.action == "changed" && event.change.new_value.gid == "1210930954402852"'
  -f 'event.change.new_value.display_value == "Client Meeting or Introduction"'

Note: The filter will skip events where the expression cannot be evaluated.`,
	Run: func(cmd *cobra.Command, args []string) {
		resource, _ := cmd.Flags().GetString("gid")
		syncToken, _ := cmd.Flags().GetString("sync")
		filterExpr, _ := cmd.Flags().GetString("filter")

		if resource == "" {
			log.Fatal("Resource GID is required")
		}

		events, err := eventManager.GetByResource(resource, syncToken)
		if err != nil {
			log.Fatalf("Failed to get events: %v", err)
		}

		// Apply filter if provided
		if filterExpr != "" {
			// Compile the expression without type checking to work with dynamic maps
			program, err := expr.Compile(filterExpr, expr.AsBool())
			if err != nil {
				log.Fatalf("Failed to compile filter expression: %v", err)
			}

			var filteredEvents []eventsLib.Event
			for _, event := range events.Data {
				// Convert event to a map for safer field access
				eventJSON, _ := json.Marshal(event)
				var eventMap map[string]interface{}
				json.Unmarshal(eventJSON, &eventMap)

				env := map[string]interface{}{
					"event": eventMap,
				}

				result, err := expr.Run(program, env)
				if err != nil {
					// Skip events that cause evaluation errors
					continue
				}

				if result == true {
					filteredEvents = append(filteredEvents, event)
				}
			}

			// Update response with filtered events
			events.Data = filteredEvents
		}

		printJSON(events)
	},
}

var eventsPollCmd = &cobra.Command{
	Use:   "poll",
	Short: "Poll events continuously",
	Long: `Continuously poll for events from a specific resource.

If no sync token is provided, the command will automatically fetch one to start polling.`,
	Run: func(cmd *cobra.Command, args []string) {
		resource, _ := cmd.Flags().GetString("gid")
		syncToken, _ := cmd.Flags().GetString("sync")
		interval, _ := cmd.Flags().GetDuration("interval")

		if resource == "" {
			log.Fatal("Resource GID is required")
		}

		// If no sync token provided, get one automatically
		if syncToken == "" {
			fmt.Printf("No sync token provided. Fetching initial sync token for resource %s...\n", resource)
			events, err := eventManager.InitializeSync(resource)
			if err != nil {
				log.Fatalf("Failed to initialize sync: %v", err)
			}
			syncToken = events.Sync
			fmt.Printf("Got sync token: %s\n", syncToken)
			fmt.Printf("Found %d events in current state\n\n", len(events.Data))
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
		resource, _ := cmd.Flags().GetString("gid")

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
			fmt.Printf("\nUse this token with: utka events get --gid %s --sync %s\n", resource, events.Sync)
		}
		fmt.Printf("Events in current state: %d\n", len(events.Data))
	},
}

func init() {
	eventsGetCmd.Flags().String("gid", "", "Resource GID (project, task, portfolio, etc.)")
	eventsGetCmd.Flags().String("sync", "", "Sync token")
	eventsGetCmd.Flags().StringP("filter", "f", "", "Filter expression to apply to events")
	eventsGetCmd.MarkFlagRequired("gid")

	eventsSyncCmd.Flags().String("gid", "", "Resource GID (project, task, portfolio, etc.)")
	eventsSyncCmd.MarkFlagRequired("gid")

	eventsPollCmd.Flags().String("gid", "", "Resource GID (project, task, portfolio, etc.)")
	eventsPollCmd.Flags().String("sync", "", "Initial sync token (optional, will be fetched automatically if not provided)")
	eventsPollCmd.Flags().Duration("interval", 5*time.Second, "Poll interval")
	eventsPollCmd.MarkFlagRequired("gid")

	eventsCmd.AddCommand(eventsGetCmd)
	eventsCmd.AddCommand(eventsSyncCmd)
	eventsCmd.AddCommand(eventsPollCmd)

	rootCmd.AddCommand(eventsCmd)
}
