package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/octoberswimmer/utka/client"
	"github.com/octoberswimmer/utka/events"
	"github.com/octoberswimmer/utka/webhooks"
	"github.com/spf13/cobra"
)

var (
	asanaClient    *client.Client
	webhookManager *webhooks.WebhookManager
	eventManager   *events.EventManager
)

var rootCmd = &cobra.Command{
	Use:   "utka",
	Short: "Asana API client for managing webhooks and events",
	Long: `utka is a command-line tool for interacting with the Asana API.
It provides commands for managing webhooks and retrieving events from Asana.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}

		token := os.Getenv("ASANA_PERSONAL_ACCESS_TOKEN")
		if token == "" {
			log.Fatal("ASANA_PERSONAL_ACCESS_TOKEN not set in environment or .env file")
		}

		asanaClient = client.NewClient(token)
		webhookManager = webhooks.NewWebhookManager(asanaClient)
		eventManager = events.NewEventManager(asanaClient)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
