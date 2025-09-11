package webhooks

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/octoberswimmer/utka/client"
)

type WebhookManager struct {
	client *client.Client
}

func NewWebhookManager(c *client.Client) *WebhookManager {
	return &WebhookManager{client: c}
}

type Webhook struct {
	GID                      string           `json:"gid,omitempty"`
	Resource                 *WebhookResource `json:"resource,omitempty"`
	Target                   string           `json:"target,omitempty"`
	Active                   bool             `json:"active"`
	CreatedAt                string           `json:"created_at,omitempty"`
	Filters                  []WebhookFilter  `json:"filters,omitempty"`
	LastSuccessAt            string           `json:"last_success_at,omitempty"`
	LastFailureAt            string           `json:"last_failure_at,omitempty"`
	LastFailureContent       string           `json:"last_failure_content,omitempty"`
	DeliveryRetryCount       int              `json:"delivery_retry_count,omitempty"`
	NextAttemptAfter         string           `json:"next_attempt_after,omitempty"`
	FailureDeletionTimestamp string           `json:"failure_deletion_timestamp,omitempty"`
	IsWorkspaceWebhook       bool             `json:"is_workspace_webhook,omitempty"`
}

type WebhookResource struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type,omitempty"`
	Name         string `json:"name,omitempty"`
}

type WebhookFilter struct {
	ResourceType    string   `json:"resource_type"`
	ResourceSubtype string   `json:"resource_subtype,omitempty"`
	Action          string   `json:"action,omitempty"`
	Fields          []string `json:"fields,omitempty"`
}

type WebhookRequest struct {
	Data *Webhook `json:"data"`
}

type WebhookResponse struct {
	Data *Webhook `json:"data"`
}

type WebhooksListResponse struct {
	Data []Webhook `json:"data"`
}

func (wm *WebhookManager) Create(resourceGID, targetURL string, filters []WebhookFilter) (*Webhook, error) {
	webhook := &Webhook{
		Resource: &WebhookResource{
			GID: resourceGID,
		},
		Target:  targetURL,
		Active:  true,
		Filters: filters,
	}

	reqBody := WebhookRequest{Data: webhook}
	respBody, err := wm.client.Post("/webhooks", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	var response WebhookResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

func (wm *WebhookManager) List(workspace string, resourceGID string) ([]Webhook, error) {
	params := url.Values{}
	if workspace != "" {
		params.Add("workspace", workspace)
	}
	if resourceGID != "" {
		params.Add("resource", resourceGID)
	}

	respBody, err := wm.client.Get("/webhooks", params)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	var response WebhooksListResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

func (wm *WebhookManager) Get(webhookGID string) (*Webhook, error) {
	endpoint := fmt.Sprintf("/webhooks/%s", webhookGID)
	respBody, err := wm.client.Get(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	var response WebhookResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

func (wm *WebhookManager) Delete(webhookGID string) error {
	endpoint := fmt.Sprintf("/webhooks/%s", webhookGID)
	if _, err := wm.client.Delete(endpoint); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	return nil
}

func (wm *WebhookManager) Update(webhookGID string, filters []WebhookFilter) (*Webhook, error) {
	endpoint := fmt.Sprintf("/webhooks/%s", webhookGID)

	webhook := &Webhook{
		Filters: filters,
	}

	reqBody := WebhookRequest{Data: webhook}
	respBody, err := wm.client.Put(endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	var response WebhookResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

func (wm *WebhookManager) UpdateFilters(webhookGID string, filters []WebhookFilter) (*Webhook, error) {
	endpoint := fmt.Sprintf("/webhooks/%s", webhookGID)

	// Create update request with only filters field
	// Using a map to ensure we only send the filters field
	updateData := map[string]interface{}{
		"filters": filters,
	}

	reqBody := map[string]interface{}{
		"data": updateData,
	}

	respBody, err := wm.client.Put(endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook filters: %w", err)
	}

	var response WebhookResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}
