package projects

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/octoberswimmer/utka/client"
)

type ProjectManager struct {
	client *client.Client
}

func NewProjectManager(c *client.Client) *ProjectManager {
	return &ProjectManager{client: c}
}

type Project struct {
	GID           string     `json:"gid"`
	ResourceType  string     `json:"resource_type"`
	Name          string     `json:"name"`
	Archived      bool       `json:"archived"`
	Color         string     `json:"color,omitempty"`
	CreatedAt     string     `json:"created_at,omitempty"`
	CurrentStatus *Status    `json:"current_status,omitempty"`
	DueDate       string     `json:"due_date,omitempty"`
	DueOn         string     `json:"due_on,omitempty"`
	HTMLNotes     string     `json:"html_notes,omitempty"`
	Members       []Member   `json:"members,omitempty"`
	ModifiedAt    string     `json:"modified_at,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	Public        bool       `json:"public"`
	StartOn       string     `json:"start_on,omitempty"`
	DefaultView   string     `json:"default_view,omitempty"`
	Followers     []Member   `json:"followers,omitempty"`
	Icon          string     `json:"icon,omitempty"`
	Owner         *Member    `json:"owner,omitempty"`
	PermalinkURL  string     `json:"permalink_url,omitempty"`
	ProjectBrief  *Brief     `json:"project_brief,omitempty"`
	Team          *Team      `json:"team,omitempty"`
	Workspace     *Workspace `json:"workspace,omitempty"`
}

type Status struct {
	GID          string  `json:"gid"`
	ResourceType string  `json:"resource_type"`
	Title        string  `json:"title,omitempty"`
	Text         string  `json:"text,omitempty"`
	HTMLText     string  `json:"html_text,omitempty"`
	Color        string  `json:"color,omitempty"`
	Author       *Member `json:"author,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	CreatedBy    *Member `json:"created_by,omitempty"`
	ModifiedAt   string  `json:"modified_at,omitempty"`
}

type Member struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type Brief struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
}

type Team struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type Workspace struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type ProjectsResponse struct {
	Data     []Project `json:"data"`
	NextPage *NextPage `json:"next_page,omitempty"`
}

type ProjectResponse struct {
	Data *Project `json:"data"`
}

type NextPage struct {
	Offset string `json:"offset,omitempty"`
	Path   string `json:"path,omitempty"`
	URI    string `json:"uri,omitempty"`
}

func (pm *ProjectManager) ListByWorkspace(workspaceGID string, archived bool, limit int) ([]Project, error) {
	allProjects := []Project{}
	params := url.Values{}
	params.Add("workspace", workspaceGID)
	params.Add("archived", fmt.Sprintf("%t", archived))
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	params.Add("opt_fields", "name,archived,created_at,modified_at,due_date,start_on,notes,public,color,owner.name,current_status.title,current_status.color")

	for {
		endpoint := "/projects"
		respBody, err := pm.client.Get(endpoint, params)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}

		var response ProjectsResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		allProjects = append(allProjects, response.Data...)

		if response.NextPage == nil || response.NextPage.Offset == "" {
			break
		}

		params.Set("offset", response.NextPage.Offset)
	}

	return allProjects, nil
}

func (pm *ProjectManager) ListByTeam(teamGID string, archived bool, limit int) ([]Project, error) {
	allProjects := []Project{}
	params := url.Values{}
	params.Add("team", teamGID)
	params.Add("archived", fmt.Sprintf("%t", archived))
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	params.Add("opt_fields", "name,archived,created_at,modified_at,due_date,start_on,notes,public,color,owner.name,current_status.title,current_status.color")

	for {
		endpoint := "/projects"
		respBody, err := pm.client.Get(endpoint, params)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}

		var response ProjectsResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		allProjects = append(allProjects, response.Data...)

		if response.NextPage == nil || response.NextPage.Offset == "" {
			break
		}

		params.Set("offset", response.NextPage.Offset)
	}

	return allProjects, nil
}

func (pm *ProjectManager) Get(projectGID string) (*Project, error) {
	endpoint := fmt.Sprintf("/projects/%s", projectGID)
	params := url.Values{}
	params.Add("opt_fields", "name,archived,created_at,modified_at,due_date,start_on,notes,html_notes,public,color,owner.name,current_status,team.name,workspace.name,followers.name,members.name,permalink_url,default_view,icon")

	respBody, err := pm.client.Get(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	var response ProjectResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}
