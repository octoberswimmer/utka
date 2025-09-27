package users

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/octoberswimmer/utka/client"
)

type UserManager struct {
	client *client.Client
}

func NewUserManager(c *client.Client) *UserManager {
	return &UserManager{client: c}
}

type User struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
	Email        string `json:"email"`
}

type UsersResponse struct {
	Data []User `json:"data"`
}

func (um *UserManager) ListInWorkspace(workspaceGID string) ([]User, error) {
	endpoint := fmt.Sprintf("/workspaces/%s/users", workspaceGID)
	params := url.Values{}
	params.Set("opt_fields", "gid,name,email")

	respBody, err := um.client.Get(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var response UsersResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}
