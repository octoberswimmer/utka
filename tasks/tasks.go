package tasks

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/octoberswimmer/utka/client"
)

type TaskManager struct {
	client *client.Client
}

func NewTaskManager(c *client.Client) *TaskManager {
	return &TaskManager{client: c}
}

type Task struct {
	GID             string        `json:"gid"`
	ResourceType    string        `json:"resource_type"`
	Name            string        `json:"name"`
	ResourceSubtype string        `json:"resource_subtype,omitempty"`
	Completed       bool          `json:"completed"`
	CompletedAt     string        `json:"completed_at,omitempty"`
	CompletedBy     *User         `json:"completed_by,omitempty"`
	CreatedAt       string        `json:"created_at,omitempty"`
	DueAt           string        `json:"due_at,omitempty"`
	DueOn           string        `json:"due_on,omitempty"`
	HTMLNotes       string        `json:"html_notes,omitempty"`
	ModifiedAt      string        `json:"modified_at,omitempty"`
	Notes           string        `json:"notes,omitempty"`
	NumSubtasks     int           `json:"num_subtasks,omitempty"`
	StartAt         string        `json:"start_at,omitempty"`
	StartOn         string        `json:"start_on,omitempty"`
	Assignee        *User         `json:"assignee,omitempty"`
	AssigneeSection *Section      `json:"assignee_section,omitempty"`
	CustomFields    []CustomField `json:"custom_fields,omitempty"`
	Followers       []User        `json:"followers,omitempty"`
	Parent          *Task         `json:"parent,omitempty"`
	Projects        []Project     `json:"projects,omitempty"`
	Tags            []Tag         `json:"tags,omitempty"`
	Workspace       *Workspace    `json:"workspace,omitempty"`
	Memberships     []Membership  `json:"memberships,omitempty"`
	Dependencies    []Dependency  `json:"dependencies,omitempty"`
	Dependents      []Dependency  `json:"dependents,omitempty"`
}

type User struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type Section struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type CustomField struct {
	GID          string      `json:"gid"`
	ResourceType string      `json:"resource_type"`
	Name         string      `json:"name,omitempty"`
	DisplayValue string      `json:"display_value,omitempty"`
	Type         string      `json:"type,omitempty"`
	Value        interface{} `json:"value,omitempty"`
}

type Project struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type Tag struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
	Color        string `json:"color,omitempty"`
}

type Workspace struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type Membership struct {
	Project *Project `json:"project,omitempty"`
	Section *Section `json:"section,omitempty"`
}

type Dependency struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type TasksResponse struct {
	Data     []Task    `json:"data"`
	NextPage *NextPage `json:"next_page,omitempty"`
}

type TaskResponse struct {
	Data *Task `json:"data"`
}

type NextPage struct {
	Offset string `json:"offset,omitempty"`
	Path   string `json:"path,omitempty"`
	URI    string `json:"uri,omitempty"`
}

func (tm *TaskManager) ListByProject(projectGID string, completed bool, limit int) ([]Task, error) {
	allTasks := []Task{}
	params := url.Values{}
	params.Add("project", projectGID)
	params.Add("completed_since", "now") // Include all tasks
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	params.Add("opt_fields", "name,completed,completed_at,completed_by.name,created_at,due_on,due_at,notes,assignee.name,assignee_section.name,tags.name,tags.color,num_subtasks,parent.name,memberships.section.name,resource_subtype,start_on")

	for {
		endpoint := "/tasks"
		respBody, err := tm.client.Get(endpoint, params)
		if err != nil {
			return nil, fmt.Errorf("failed to list tasks: %w", err)
		}

		var response TasksResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Filter based on completed status if needed
		for _, task := range response.Data {
			if completed || !task.Completed {
				allTasks = append(allTasks, task)
			}
		}

		if response.NextPage == nil || response.NextPage.Offset == "" {
			break
		}

		params.Set("offset", response.NextPage.Offset)
	}

	return allTasks, nil
}

func (tm *TaskManager) ListByAssignee(assigneeGID string, workspaceGID string, completed bool, limit int) ([]Task, error) {
	allTasks := []Task{}
	params := url.Values{}
	params.Add("assignee", assigneeGID)
	params.Add("workspace", workspaceGID)
	params.Add("completed_since", "now") // Include all tasks
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	params.Add("opt_fields", "name,completed,completed_at,created_at,due_on,due_at,notes,projects.name,assignee_section.name,tags.name,num_subtasks")

	for {
		endpoint := "/tasks"
		respBody, err := tm.client.Get(endpoint, params)
		if err != nil {
			return nil, fmt.Errorf("failed to list tasks: %w", err)
		}

		var response TasksResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Filter based on completed status if needed
		for _, task := range response.Data {
			if completed || !task.Completed {
				allTasks = append(allTasks, task)
			}
		}

		if response.NextPage == nil || response.NextPage.Offset == "" {
			break
		}

		params.Set("offset", response.NextPage.Offset)
	}

	return allTasks, nil
}

func (tm *TaskManager) ListBySection(sectionGID string, completed bool, limit int) ([]Task, error) {
	allTasks := []Task{}
	params := url.Values{}
	params.Add("section", sectionGID)
	params.Add("completed_since", "now") // Include all tasks
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	params.Add("opt_fields", "name,completed,completed_at,created_at,due_on,due_at,notes,assignee.name,tags.name,num_subtasks")

	for {
		endpoint := "/tasks"
		respBody, err := tm.client.Get(endpoint, params)
		if err != nil {
			return nil, fmt.Errorf("failed to list tasks: %w", err)
		}

		var response TasksResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Filter based on completed status if needed
		for _, task := range response.Data {
			if completed || !task.Completed {
				allTasks = append(allTasks, task)
			}
		}

		if response.NextPage == nil || response.NextPage.Offset == "" {
			break
		}

		params.Set("offset", response.NextPage.Offset)
	}

	return allTasks, nil
}

func (tm *TaskManager) Get(taskGID string) (*Task, error) {
	endpoint := fmt.Sprintf("/tasks/%s", taskGID)
	params := url.Values{}
	params.Add("opt_fields", "name,completed,completed_at,completed_by.name,created_at,due_on,due_at,html_notes,notes,assignee.name,assignee_section.name,custom_fields,followers.name,parent.name,projects.name,tags,workspace.name,memberships.project.name,memberships.section.name,num_subtasks,resource_subtype,start_on,start_at,dependencies.name,dependents.name")

	respBody, err := tm.client.Get(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var response TaskResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}
