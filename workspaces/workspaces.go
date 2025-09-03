package workspaces

import (
	"encoding/json"
	"fmt"

	"github.com/octoberswimmer/utka/client"
)

type WorkspaceManager struct {
	client *client.Client
}

func NewWorkspaceManager(c *client.Client) *WorkspaceManager {
	return &WorkspaceManager{client: c}
}

type Workspace struct {
	GID            string   `json:"gid"`
	ResourceType   string   `json:"resource_type"`
	Name           string   `json:"name"`
	EmailDomains   []string `json:"email_domains,omitempty"`
	IsOrganization bool     `json:"is_organization"`
}

type WorkspacesResponse struct {
	Data []Workspace `json:"data"`
}

type WorkspaceResponse struct {
	Data *Workspace `json:"data"`
}

func (wm *WorkspaceManager) List() ([]Workspace, error) {
	respBody, err := wm.client.Get("/workspaces", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	var response WorkspacesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

func (wm *WorkspaceManager) Get(workspaceGID string) (*Workspace, error) {
	endpoint := fmt.Sprintf("/workspaces/%s", workspaceGID)
	respBody, err := wm.client.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	var response WorkspaceResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}
