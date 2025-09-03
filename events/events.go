package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/octoberswimmer/utka/client"
)

type EventManager struct {
	client *client.Client
}

func NewEventManager(c *client.Client) *EventManager {
	return &EventManager{client: c}
}

type Event struct {
	User      *EventUser     `json:"user,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	Action    string         `json:"action,omitempty"`
	Resource  *EventResource `json:"resource,omitempty"`
	Parent    *EventParent   `json:"parent,omitempty"`
	Change    *EventChange   `json:"change,omitempty"`
	Type      string         `json:"type,omitempty"`
}

type EventUser struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type EventResource struct {
	GID             string `json:"gid"`
	ResourceType    string `json:"resource_type"`
	Name            string `json:"name,omitempty"`
	ResourceSubtype string `json:"resource_subtype,omitempty"`
}

type EventParent struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name,omitempty"`
}

type EventChange struct {
	Field        string      `json:"field,omitempty"`
	Action       string      `json:"action,omitempty"`
	OldValue     interface{} `json:"old_value,omitempty"`
	NewValue     interface{} `json:"new_value,omitempty"`
	AddedValue   interface{} `json:"added_value,omitempty"`
	RemovedValue interface{} `json:"removed_value,omitempty"`
}

type EventsResponse struct {
	Data     []Event   `json:"data"`
	Sync     string    `json:"sync,omitempty"`
	HasMore  bool      `json:"has_more,omitempty"`
	NextPage *NextPage `json:"next_page,omitempty"`
}

type NextPage struct {
	Offset string `json:"offset,omitempty"`
	Path   string `json:"path,omitempty"`
	URI    string `json:"uri,omitempty"`
}

func (em *EventManager) GetByResource(resourceGID string, syncToken string) (*EventsResponse, error) {
	endpoint := fmt.Sprintf("/events")
	params := url.Values{}
	params.Add("resource", resourceGID)

	if syncToken != "" {
		params.Add("sync", syncToken)
	}

	respBody, err := em.client.Get(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	var response EventsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (em *EventManager) InitializeSync(resourceGID string) (*EventsResponse, error) {
	// According to Asana docs, when you get a 412 error, the response includes a new sync token
	// We need to handle this specially
	endpoint := fmt.Sprintf("/events")
	params := url.Values{}
	params.Add("resource", resourceGID)

	// Make raw request to handle 412 specially
	fullURL := em.client.GetBaseURL() + endpoint + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+em.client.GetAccessToken())
	req.Header.Set("Accept", "application/json")

	resp, err := em.client.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response EventsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Even on 412, Asana returns the sync token in the response
	if resp.StatusCode == 412 {
		// Return the response with the sync token
		return &response, nil
	}

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(respBody, &errorResp); err == nil && len(errorResp.Errors) > 0 {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errorResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return &response, nil
}

func (em *EventManager) Poll(resourceGID string, syncToken string, pollInterval time.Duration) (<-chan Event, <-chan error) {
	eventsChan := make(chan Event)
	errorsChan := make(chan error)

	go func() {
		defer close(eventsChan)
		defer close(errorsChan)

		currentSync := syncToken

		for {
			response, err := em.GetByResource(resourceGID, currentSync)
			if err != nil {
				errorsChan <- err
				time.Sleep(pollInterval)
				continue
			}

			for _, event := range response.Data {
				eventsChan <- event
			}

			if response.Sync != "" {
				currentSync = response.Sync
			}

			time.Sleep(pollInterval)
		}
	}()

	return eventsChan, errorsChan
}
