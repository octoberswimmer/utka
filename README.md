# utka

A command-line Asana API client in Go for managing webhooks and events, built with Cobra.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/octoberswimmer/utka.git
cd utka

# Build the binary
go build -o utka .

# Or install globally
go install
```

## Configuration

Create a `.env` file in your working directory with your Asana Personal Access Token:

```env
ASANA_PERSONAL_ACCESS_TOKEN=your_token_here
```

### Getting an Asana Personal Access Token

1. Go to https://app.asana.com/0/my-apps
2. Click "Create New Token"
3. Give it a descriptive name
4. Copy the token immediately (you won't be able to see it again)
5. Save it in your `.env` file

## Quick Start

```bash
# List your workspaces to get workspace GIDs
utka workspace list

# List projects in a workspace
utka project list --workspace <workspace_gid>

# List webhooks in a workspace
utka webhook list --workspace <workspace_gid>

# Initialize sync token for a resource (project, task, etc.)
utka events sync --resource <resource_gid>

# Get events using the sync token
utka events get --resource <resource_gid> --sync <sync_token>
```

## Commands

### Workspace Commands

Get information about your Asana workspaces:

```bash
# List all workspaces you have access to
utka workspace list

# Get detailed information about a specific workspace
utka workspace get --gid <workspace_gid>
```

### Project Commands

List and manage projects within workspaces or teams:

```bash
# List all active projects in a workspace
utka project list --workspace <workspace_gid>

# List projects in a specific team
utka project list --team <team_gid>

# Include archived projects
utka project list --workspace <workspace_gid> --archived

# Limit the number of results
utka project list --workspace <workspace_gid> --limit 10

# Output as JSON for processing
utka project list --workspace <workspace_gid> --json

# Get detailed information about a specific project
utka project get --gid <project_gid>
```

### Webhook Commands

Manage Asana webhooks for real-time notifications:

```bash
# List all webhooks (requires either workspace or resource)
utka webhook list --workspace <workspace_gid>
utka webhook list --resource <resource_gid>

# Get details of a specific webhook
utka webhook get --gid <webhook_gid>

# Create a new webhook for a resource
utka webhook create --resource <resource_gid> --target <callback_url>

# Delete a webhook
utka webhook delete --gid <webhook_gid>

# Update webhook filters (currently updates with empty filters)
utka webhook update --gid <webhook_gid>
```

### Event Commands

Monitor and retrieve events from Asana resources. The Events API requires sync tokens for proper pagination.

```bash
# Initialize sync token for a resource (required for first-time use)
# This handles the "Sync token invalid or too old" error
utka events sync --resource <resource_gid>

# Get events for any resource (project, task, portfolio, etc.)
utka events get --resource <resource_gid>

# Get events with a sync token (for incremental updates)
utka events get --resource <resource_gid> --sync <sync_token>

# Poll events continuously (real-time monitoring)
utka events poll --resource <resource_gid>                      # Default 5s interval
utka events poll --resource <resource_gid> --interval 10s       # Custom interval
utka events poll --resource <resource_gid> --sync <sync_token>  # Start from sync point
```

#### Understanding Sync Tokens

The Asana Events API uses sync tokens to track your position in the event stream:

1. **First Time**: Use `utka events sync` to initialize and get your first sync token
2. **Incremental Updates**: Use the sync token from previous responses to get only new events
3. **Token Expired**: If you get a 412 error, run `utka events sync` again to refresh

## Examples

### Working with Projects

```bash
# 1. Get your workspace GID
utka workspace list
# Output: • My Workspace (Organization) - GID: 1234567890

# 2. List all projects in the workspace
utka project list --workspace 1234567890
# Output: 
# Found 3 project(s):
# 
# • Marketing Campaign 2024
#   GID: 2345678901
#   Color: light-green
#   Owner: John Doe
#   Status: On Track (green)
#   Due: 2024-12-31

# 3. Get detailed info about a specific project
utka project get --gid 2345678901
```

### Setting up a Webhook

```bash
# 1. Get your workspace and project GIDs
utka workspace list
utka project list --workspace <workspace_gid>

# 2. Create a webhook for a project
utka webhook create --resource <project_gid> --target https://your-server.com/webhook

# 3. Verify the webhook was created
utka webhook list --workspace <workspace_gid>
```

### Monitoring Project Events

```bash
# 1. Initialize sync for a project
utka events sync --resource <project_gid>
# Output: New sync token: d1f4b7c3...

# 2. Get current events
utka events get --resource <project_gid> --sync d1f4b7c3...

# 3. Poll for new events in real-time
utka events poll --resource <project_gid> --interval 3s --sync d1f4b7c3...
```

### Working with Tasks

```bash
# Get task events
utka events sync --resource <task_gid>
utka events get --resource <task_gid> --sync <token>

# Monitor task changes
utka events poll --resource <task_gid> --interval 2s
```

## Troubleshooting

### "Sync token invalid or too old" Error

This error occurs when your sync token has expired or is invalid. To fix:

```bash
# Reinitialize the sync token
utka events sync --resource <resource_gid>

# Use the new token for subsequent requests
utka events get --resource <resource_gid> --sync <new_token>
```

### "You should specify one of workspace" Error

When listing webhooks, you must provide either a workspace or resource filter:

```bash
# First, get your workspace GID
utka workspace list

# Then use it to list webhooks
utka webhook list --workspace <workspace_gid>
```

## Help

Get detailed help for any command:

```bash
utka --help                    # General help
utka webhook --help            # Webhook commands help
utka events --help             # Events commands help  
utka workspace --help          # Workspace commands help
utka webhook create --help     # Specific command help
```

## API Reference

This tool implements the following Asana API endpoints:
- [Workspaces API](https://developers.asana.com/reference/workspaces)
- [Webhooks API](https://developers.asana.com/reference/webhooks)
- [Events API](https://developers.asana.com/reference/events)

## License

MIT