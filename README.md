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

# List tasks in a project
utka task list --project <project_gid>

# List webhooks in a workspace
utka webhook list --workspace <workspace_gid>

# Initialize sync token for a resource (project, task, etc.)
utka events sync --gid <resource_gid>

# Get events using the sync token
utka events get --gid <resource_gid> --sync <sync_token>
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

### Task Commands

List, manage, and edit tasks within projects:

```bash
# List all incomplete tasks in a project
utka task list --project <project_gid>

# Include tasks completed in the last 7 days
utka task list --project <project_gid> --completed 7

# List tasks in a specific section
utka task list --section <section_gid>

# List tasks assigned to a user in a workspace
utka task list --assignee <user_gid> --workspace <workspace_gid>

# Get detailed information about a specific task
utka task get --gid <task_gid>

# Edit task properties
utka task edit --gid <task_gid> --name "New task name"
utka task edit --gid <task_gid> --notes "Updated notes"
utka task edit --gid <task_gid> --assignee <user_gid>
utka task edit --gid <task_gid> --due-date 2024-12-31
utka task edit --gid <task_gid> --due-date null  # Remove due date

# Edit multiple properties at once
utka task edit --gid <task_gid> --name "New name" --assignee <user_gid> --due-date 2024-12-31

# Mark task as complete/incomplete
utka task complete --gid <task_gid>
utka task uncomplete --gid <task_gid>
```

The task list displays:
- Completion status with checkboxes
- Task name and type (milestone/subtasks)
- Assignee and due dates
- Tags with colors
- Notes (truncated)
- Section grouping when available

Pass `--completed 0` (the default) to show only incomplete tasks.

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

Monitor and retrieve events from Asana resources. The Events API requires sync tokens for proper pagination. The `events get` command automatically fetches all pages when more events are available.

```bash
# Initialize sync token for a resource (required for first-time use)
# This handles the "Sync token invalid or too old" error
utka events sync --gid <resource_gid>

# Get all events for any resource (automatically fetches all pages)
utka events get --gid <resource_gid>

# Get events with a sync token (for incremental updates)
utka events get --gid <resource_gid> --sync <sync_token>

# Filter events using expressions
utka events get --gid <resource_gid> -f 'event.action == "changed"'
utka events get --gid <resource_gid> -f 'event.change.field == "custom_fields"'
utka events get --gid <resource_gid> -f 'event.change.new_value.gid == "123456"'

# Poll events continuously (real-time monitoring)
utka events poll --gid <resource_gid>                      # Default 5s interval
utka events poll --gid <resource_gid> --interval 10s       # Custom interval
utka events poll --gid <resource_gid> --sync <sync_token>  # Start from sync point
```

#### Filtering Events

The `events get` command supports powerful filter expressions using the `-f` flag:

```bash
# Filter by action type
utka events get --gid <resource_gid> -f 'event.action == "changed"'
utka events get --gid <resource_gid> -f 'event.action == "added"'

# Filter by resource type and subtype
utka events get --gid <resource_gid> -f 'event.resource.resource_subtype == "default_task"'

# Filter by change fields
utka events get --gid <resource_gid> -f 'event.change.field == "completed"'
utka events get --gid <resource_gid> -f 'event.change.field == "name"'

# Filter by custom field changes
utka events get --gid <resource_gid> -f 'event.change.new_value.gid == "1234567890"'
utka events get --gid <resource_gid> -f 'event.change.new_value.display_value == "In Progress"'

# Complex filters with multiple conditions
utka events get --gid <resource_gid> -f 'event.action == "changed" && event.change.field == "custom_fields"'
utka events get --gid <resource_gid> -f 'event.action == "changed" && event.user.name == "John Doe"'
```

Note: Filters use lowercase field names and safely handle nil values by skipping events that cause evaluation errors.

#### Understanding Sync Tokens

The Asana Events API uses sync tokens to track your position in the event stream:

1. **First Time**: Use `utka events sync` to initialize and get your first sync token
2. **Incremental Updates**: Use the sync token from previous responses to get only new events
3. **Token Expired**: If you get a 412 error, run `utka events sync` again to refresh
4. **Pagination**: The `events get` command automatically fetches all pages when `has_more` is true

## Examples

### Working with Projects and Tasks

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

# 3. List tasks in the project
utka task list --project 2345678901
# Output:
# Found 5 task(s):
#
# [ ] Design landing page
#     GID: 3456789012
#     Assignee: Jane Smith
#     Due: 2024-11-15
#     Tags: design (red), high-priority (orange)
#
# [✓] Write copy for campaign
#     GID: 3456789013
#     Assignee: Bob Johnson
#     Completed: 2024-11-01 by Bob Johnson

# 4. Get detailed info about a specific task
utka task get --gid 3456789012
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
utka events sync --gid <project_gid>
# Output: New sync token: d1f4b7c3...

# 2. Get current events
utka events get --gid <project_gid> --sync d1f4b7c3...

# 3. Poll for new events in real-time
utka events poll --gid <project_gid> --interval 3s --sync d1f4b7c3...
```

### Working with Tasks

```bash
# Edit a task
utka task edit --gid 3456789012 --name "Update landing page design" --due-date 2024-11-20
# Output: ✓ Task updated successfully
#   Name: Update landing page design
#   Due: 2024-11-20

# Complete a task
utka task complete --gid 3456789012
# Output: ✓ Task completed: Update landing page design

# Get task events
utka events sync --gid <task_gid>
utka events get --gid <task_gid> --sync <token>

# Monitor task changes
utka events poll --gid <task_gid> --interval 2s
```

## Troubleshooting

### "Sync token invalid or too old" Error

This error occurs when your sync token has expired or is invalid. To fix:

```bash
# Reinitialize the sync token
utka events sync --gid <resource_gid>

# Use the new token for subsequent requests
utka events get --gid <resource_gid> --sync <new_token>
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
