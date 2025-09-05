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

var taskGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get task details",
	Long:  `Retrieve detailed information about a specific task.`,
	Run: func(cmd *cobra.Command, args []string) {
		gid, _ := cmd.Flags().GetString("gid")
		if gid == "" {
			log.Fatal("Task GID is required")
		}

		task, err := taskManager.Get(gid)
		if err != nil {
			log.Fatalf("Failed to get task: %v", err)
		}

		printJSON(task)
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
	taskGetCmd.MarkFlagRequired("gid")

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskGetCmd)

	rootCmd.AddCommand(taskCmd)
}
