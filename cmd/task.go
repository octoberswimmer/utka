package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/octoberswimmer/utka/tasks"
	"github.com/spf13/cobra"
)

var taskManager *tasks.TaskManager

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage Asana tasks",
	Long:  `Commands for listing and retrieving information about Asana tasks.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
		taskManager = tasks.NewTaskManager(asanaClient)
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List tasks in a project, section, or assigned to a user.`,
	Run: func(cmd *cobra.Command, args []string) {
		project, _ := cmd.Flags().GetString("project")
		section, _ := cmd.Flags().GetString("section")
		assignee, _ := cmd.Flags().GetString("assignee")
		workspace, _ := cmd.Flags().GetString("workspace")
		completed, _ := cmd.Flags().GetBool("completed")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		// Count how many filters are specified
		filtersSpecified := 0
		if project != "" {
			filtersSpecified++
		}
		if section != "" {
			filtersSpecified++
		}
		if assignee != "" {
			filtersSpecified++
		}

		if filtersSpecified == 0 {
			log.Fatal("One of --project, --section, or --assignee is required")
		}

		if filtersSpecified > 1 {
			log.Fatal("Please specify only one of --project, --section, or --assignee")
		}

		// If assignee is specified, workspace is required
		if assignee != "" && workspace == "" {
			log.Fatal("--workspace is required when using --assignee")
		}

		var tasksList []tasks.Task
		var err error

		switch {
		case project != "":
			tasksList, err = taskManager.ListByProject(project, completed, limit)
		case section != "":
			tasksList, err = taskManager.ListBySection(section, completed, limit)
		case assignee != "":
			tasksList, err = taskManager.ListByAssignee(assignee, workspace, completed, limit)
		}

		if err != nil {
			log.Fatalf("Failed to list tasks: %v", err)
		}

		if jsonOutput {
			printJSON(tasksList)
			return
		}

		// Pretty print tasks
		if len(tasksList) == 0 {
			fmt.Println("No tasks found")
			return
		}

		fmt.Printf("Found %d task(s):\n\n", len(tasksList))

		// Group tasks by section if they have memberships
		var sectionMap = make(map[string][]tasks.Task)
		var noSection []tasks.Task

		for _, task := range tasksList {
			sectionName := ""
			if len(task.Memberships) > 0 && task.Memberships[0].Section != nil {
				sectionName = task.Memberships[0].Section.Name
			}

			if sectionName != "" {
				sectionMap[sectionName] = append(sectionMap[sectionName], task)
			} else {
				noSection = append(noSection, task)
			}
		}

		// Print tasks without sections first
		if len(noSection) > 0 {
			printTaskList(noSection)
		}

		// Print tasks by section
		for sectionName, sectionTasks := range sectionMap {
			if len(noSection) > 0 || len(sectionMap) > 1 {
				fmt.Printf("\n📁 %s\n", sectionName)
				fmt.Println(strings.Repeat("-", 40))
			}
			printTaskList(sectionTasks)
		}
	},
}

func printTaskList(tasksList []tasks.Task) {
	for _, task := range tasksList {
		// Task name with completion status
		status := "[ ]"
		if task.Completed {
			status = "[✓]"
		}

		taskType := ""
		if task.ResourceSubtype == "milestone" {
			taskType = " 🏁"
		} else if task.NumSubtasks > 0 {
			taskType = fmt.Sprintf(" (%d subtasks)", task.NumSubtasks)
		}

		fmt.Printf("%s %s%s\n", status, task.Name, taskType)
		fmt.Printf("    GID: %s\n", task.GID)

		// Assignee
		if task.Assignee != nil && task.Assignee.Name != "" {
			fmt.Printf("    Assignee: %s\n", task.Assignee.Name)
		}

		// Due date
		if task.DueOn != "" {
			fmt.Printf("    Due: %s\n", task.DueOn)
		} else if task.DueAt != "" {
			fmt.Printf("    Due: %s\n", task.DueAt)
		}

		// Tags
		if len(task.Tags) > 0 {
			var tagNames []string
			for _, tag := range task.Tags {
				if tag.Color != "" {
					tagNames = append(tagNames, fmt.Sprintf("%s (%s)", tag.Name, tag.Color))
				} else {
					tagNames = append(tagNames, tag.Name)
				}
			}
			fmt.Printf("    Tags: %s\n", strings.Join(tagNames, ", "))
		}

		// Notes (truncated)
		if task.Notes != "" {
			notes := strings.ReplaceAll(task.Notes, "\n", " ")
			if len(notes) > 80 {
				notes = notes[:77] + "..."
			}
			fmt.Printf("    Notes: %s\n", notes)
		}

		// Completed info
		if task.Completed && task.CompletedAt != "" {
			completedBy := ""
			if task.CompletedBy != nil && task.CompletedBy.Name != "" {
				completedBy = " by " + task.CompletedBy.Name
			}
			fmt.Printf("    Completed: %s%s\n", task.CompletedAt[:10], completedBy)
		}

		fmt.Println()
	}
}

func printTaskDetails(task *tasks.Task) {
	// Header with task name and status
	status := "[ ]"
	if task.Completed {
		status = "[✓]"
	}

	taskType := ""
	if task.ResourceSubtype == "milestone" {
		taskType = " 🏁 Milestone"
	}

	fmt.Printf("%s %s%s\n", status, task.Name, taskType)
	fmt.Println(strings.Repeat("=", 80))

	// Basic Info
	fmt.Printf("GID:         %s\n", task.GID)
	fmt.Printf("URL:         https://app.asana.com/0/0/%s\n", task.GID)

	if task.Assignee != nil && task.Assignee.Name != "" {
		fmt.Printf("Assignee:    %s (GID: %s)\n", task.Assignee.Name, task.Assignee.GID)
	}

	// Dates
	if task.StartOn != "" {
		fmt.Printf("Start Date:  %s\n", task.StartOn)
	} else if task.StartAt != "" {
		fmt.Printf("Start Date:  %s\n", task.StartAt)
	}

	if task.DueOn != "" {
		fmt.Printf("Due Date:    %s\n", task.DueOn)
	} else if task.DueAt != "" {
		fmt.Printf("Due Date:    %s\n", task.DueAt)
	}

	if task.CreatedAt != "" {
		fmt.Printf("Created:     %s\n", task.CreatedAt)
	}

	if task.ModifiedAt != "" {
		fmt.Printf("Modified:    %s\n", task.ModifiedAt)
	}

	if task.Completed && task.CompletedAt != "" {
		completedBy := ""
		if task.CompletedBy != nil && task.CompletedBy.Name != "" {
			completedBy = fmt.Sprintf(" by %s", task.CompletedBy.Name)
		}
		fmt.Printf("Completed:   %s%s\n", task.CompletedAt, completedBy)
	}

	// Projects
	if len(task.Projects) > 0 {
		fmt.Printf("\nProjects:\n")
		for _, project := range task.Projects {
			fmt.Printf("  - %s (GID: %s)\n", project.Name, project.GID)
		}
	}

	// Memberships (shows project + section)
	if len(task.Memberships) > 0 {
		fmt.Printf("\nSections:\n")
		for _, membership := range task.Memberships {
			projectName := ""
			if membership.Project != nil {
				projectName = membership.Project.Name
			}
			sectionName := ""
			if membership.Section != nil {
				sectionName = membership.Section.Name
			}
			if projectName != "" && sectionName != "" {
				fmt.Printf("  - %s / %s\n", projectName, sectionName)
			} else if sectionName != "" {
				fmt.Printf("  - %s\n", sectionName)
			}
		}
	}

	// Tags
	if len(task.Tags) > 0 {
		fmt.Printf("\nTags:\n")
		for _, tag := range task.Tags {
			if tag.Color != "" {
				fmt.Printf("  - %s (%s)\n", tag.Name, tag.Color)
			} else {
				fmt.Printf("  - %s\n", tag.Name)
			}
		}
	}

	// Parent task
	if task.Parent != nil && task.Parent.Name != "" {
		fmt.Printf("\nParent Task: %s (GID: %s)\n", task.Parent.Name, task.Parent.GID)
	}

	// Subtasks
	if task.NumSubtasks > 0 {
		fmt.Printf("\nSubtasks:    %d\n", task.NumSubtasks)
	}

	// Dependencies
	if len(task.Dependencies) > 0 {
		fmt.Printf("\nDependencies (blocking this task):\n")
		for _, dep := range task.Dependencies {
			fmt.Printf("  - %s (GID: %s)\n", dep.Name, dep.GID)
		}
	}

	// Dependents
	if len(task.Dependents) > 0 {
		fmt.Printf("\nDependents (blocked by this task):\n")
		for _, dep := range task.Dependents {
			fmt.Printf("  - %s (GID: %s)\n", dep.Name, dep.GID)
		}
	}

	// Followers
	if len(task.Followers) > 0 {
		fmt.Printf("\nFollowers:\n")
		for _, follower := range task.Followers {
			fmt.Printf("  - %s (GID: %s)\n", follower.Name, follower.GID)
		}
	}

	// Custom fields
	if len(task.CustomFields) > 0 {
		fmt.Printf("\nCustom Fields:\n")
		for _, field := range task.CustomFields {
			if field.DisplayValue != "" {
				fmt.Printf("  - %s: %s\n", field.Name, field.DisplayValue)
			} else if field.Value != nil {
				fmt.Printf("  - %s: %v\n", field.Name, field.Value)
			}
		}
	}

	// Workspace
	if task.Workspace != nil && task.Workspace.Name != "" {
		fmt.Printf("\nWorkspace:   %s (GID: %s)\n", task.Workspace.Name, task.Workspace.GID)
	}

	// Notes (full text)
	if task.Notes != "" {
		fmt.Printf("\nNotes:\n")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println(task.Notes)
		fmt.Println(strings.Repeat("-", 80))
	}
}

var taskGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get task details",
	Long:  `Retrieve detailed information about a specific task.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Task GID is required")
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")

		task, err := taskManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get task: %v", err)
		}

		if jsonOutput {
			printJSON(task)
			return
		}

		printTaskDetails(task)
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a task",
	Long:  `Update task properties like name, notes, assignee, due date, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Task GID is required")
		}

		// Build update struct with only the fields that were provided
		update := &tasks.TaskUpdate{}
		hasUpdate := false

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			update.Name = &name
			hasUpdate = true
		}

		if cmd.Flags().Changed("notes") {
			notes, _ := cmd.Flags().GetString("notes")
			update.Notes = &notes
			hasUpdate = true
		}

		if cmd.Flags().Changed("assignee") {
			assignee, _ := cmd.Flags().GetString("assignee")
			update.Assignee = &assignee
			hasUpdate = true
		}

		if cmd.Flags().Changed("due-date") {
			dueDate, _ := cmd.Flags().GetString("due-date")
			update.DueOn = &dueDate
			hasUpdate = true
		}

		if cmd.Flags().Changed("start-date") {
			startDate, _ := cmd.Flags().GetString("start-date")
			update.StartOn = &startDate
			hasUpdate = true
		}

		if cmd.Flags().Changed("completed") {
			completed, _ := cmd.Flags().GetBool("completed")
			update.Completed = &completed
			hasUpdate = true
		}

		if cmd.Flags().Changed("tags") {
			tags, _ := cmd.Flags().GetStringSlice("tags")
			update.Tags = tags
			hasUpdate = true
		}

		if !hasUpdate {
			log.Fatal("No updates specified. Use flags to specify what to update.")
		}

		task, err := taskManager.Update(gid, update)
		if err != nil {
			log.Fatalf("Failed to update task: %v", err)
		}

		fmt.Printf("✓ Task updated successfully\n")
		fmt.Printf("  Name: %s\n", task.Name)
		if task.Assignee != nil {
			fmt.Printf("  Assignee: %s\n", task.Assignee.Name)
		}
		if task.DueOn != "" {
			fmt.Printf("  Due: %s\n", task.DueOn)
		}
		if task.Completed {
			fmt.Printf("  Status: Completed\n")
		}
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Mark a task as complete",
	Long:  `Mark a task as complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Task GID is required")
		}

		task, err := taskManager.Complete(gid)
		if err != nil {
			log.Fatalf("Failed to complete task: %v", err)
		}

		fmt.Printf("✓ Task completed: %s\n", task.Name)
	},
}

var taskUncompleteCmd = &cobra.Command{
	Use:   "uncomplete",
	Short: "Mark a task as incomplete",
	Long:  `Mark a completed task as incomplete.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Task GID is required")
		}

		task, err := taskManager.Uncomplete(gid)
		if err != nil {
			log.Fatalf("Failed to uncomplete task: %v", err)
		}

		fmt.Printf("✓ Task marked as incomplete: %s\n", task.Name)
	},
}

func init() {
	taskListCmd.Flags().String("project", "", "Project GID")
	taskListCmd.Flags().String("section", "", "Section GID")
	taskListCmd.Flags().String("assignee", "", "Assignee user GID")
	taskListCmd.Flags().String("workspace", "", "Workspace GID (required with --assignee)")
	taskListCmd.Flags().Bool("completed", false, "Include completed tasks")
	taskListCmd.Flags().Int("limit", 0, "Limit number of results (0 for all)")
	taskListCmd.Flags().Bool("json", false, "Output as JSON")

	taskGetCmd.Flags().String("gid", "", "Task GID")
	taskGetCmd.Flags().Bool("json", false, "Output as JSON")
	taskGetCmd.MarkFlagRequired("gid")

	taskEditCmd.Flags().String("gid", "", "Task GID")
	taskEditCmd.Flags().String("name", "", "Task name")
	taskEditCmd.Flags().String("notes", "", "Task notes")
	taskEditCmd.Flags().String("assignee", "", "Assignee GID (use 'null' to unassign)")
	taskEditCmd.Flags().String("due-date", "", "Due date (YYYY-MM-DD format, or 'null' to remove)")
	taskEditCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD format)")
	taskEditCmd.Flags().Bool("completed", false, "Mark as completed")
	taskEditCmd.Flags().StringSlice("tags", nil, "Tag GIDs (comma-separated)")
	taskEditCmd.MarkFlagRequired("gid")

	taskCompleteCmd.Flags().String("gid", "", "Task GID")
	taskCompleteCmd.MarkFlagRequired("gid")

	taskUncompleteCmd.Flags().String("gid", "", "Task GID")
	taskUncompleteCmd.MarkFlagRequired("gid")

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskUncompleteCmd)

	rootCmd.AddCommand(taskCmd)
}
