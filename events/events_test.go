package events

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/octoberswimmer/utka/client"
)

func TestGetByResource(t *testing.T) {
	tests := []struct {
		name           string
		resourceGID    string
		syncToken      string
		responses      []string
		wantEventCount int
		wantSync       string
		wantErr        bool
	}{
		{
			name:        "single page of events",
			resourceGID: "123456",
			syncToken:   "",
			responses: []string{
				`{"data":[{"action":"changed","type":"task"}],"sync":"token1","has_more":false}`,
			},
			wantEventCount: 1,
			wantSync:       "token1",
			wantErr:        false,
		},
		{
			name:        "multiple pages of events",
			resourceGID: "123456",
			syncToken:   "",
			responses: []string{
				`{"data":[{"action":"changed","type":"task"}],"sync":"token1","has_more":true}`,
				`{"data":[{"action":"added","type":"task"}],"sync":"token2","has_more":true}`,
				`{"data":[{"action":"removed","type":"task"}],"sync":"token3","has_more":false}`,
			},
			wantEventCount: 3,
			wantSync:       "token3",
			wantErr:        false,
		},
		{
			name:        "with sync token",
			resourceGID: "123456",
			syncToken:   "existing_token",
			responses: []string{
				`{"data":[{"action":"changed","type":"task"},{"action":"added","type":"task"}],"sync":"new_token","has_more":false}`,
			},
			wantEventCount: 2,
			wantSync:       "new_token",
			wantErr:        false,
		},
		{
			name:        "empty response",
			resourceGID: "123456",
			syncToken:   "",
			responses: []string{
				`{"data":[],"sync":"token1","has_more":false}`,
			},
			wantEventCount: 0,
			wantSync:       "token1",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if callCount >= len(tt.responses) {
					t.Fatalf("Unexpected request #%d", callCount+1)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.responses[callCount]))
				callCount++
			}))
			defer server.Close()

			c := &client.Client{}
			c.SetBaseURL(server.URL)
			c.SetAccessToken("test_token")
			c.SetHTTPClient(http.DefaultClient)

			em := NewEventManager(c)
			result, err := em.GetByResource(tt.resourceGID, tt.syncToken)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(result.Data) != tt.wantEventCount {
					t.Errorf("GetByResource() got %d events, want %d", len(result.Data), tt.wantEventCount)
				}
				if result.Sync != tt.wantSync {
					t.Errorf("GetByResource() sync = %v, want %v", result.Sync, tt.wantSync)
				}
			}

			if callCount != len(tt.responses) {
				t.Errorf("Expected %d requests, got %d", len(tt.responses), callCount)
			}
		})
	}
}

func TestInitializeSync(t *testing.T) {
	tests := []struct {
		name        string
		resourceGID string
		statusCode  int
		response    string
		wantSync    string
		wantErr     bool
	}{
		{
			name:        "successful initialization",
			resourceGID: "123456",
			statusCode:  200,
			response:    `{"data":[],"sync":"initial_token"}`,
			wantSync:    "initial_token",
			wantErr:     false,
		},
		{
			name:        "412 error with sync token",
			resourceGID: "123456",
			statusCode:  412,
			response:    `{"data":[],"sync":"new_sync_token"}`,
			wantSync:    "new_sync_token",
			wantErr:     false,
		},
		{
			name:        "API error",
			resourceGID: "123456",
			statusCode:  400,
			response:    `{"errors":[{"message":"Bad request"}]}`,
			wantSync:    "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &client.Client{}
			c.SetBaseURL(server.URL)
			c.SetAccessToken("test_token")
			c.SetHTTPClient(http.DefaultClient)

			em := NewEventManager(c)
			result, err := em.InitializeSync(tt.resourceGID)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitializeSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && result.Sync != tt.wantSync {
				t.Errorf("InitializeSync() sync = %v, want %v", result.Sync, tt.wantSync)
			}
		})
	}
}

func TestEventJSONMarshaling(t *testing.T) {
	eventJSON := `{
		"action": "changed",
		"type": "task",
		"user": {
			"gid": "12345",
			"resource_type": "user",
			"name": "Test User"
		},
		"resource": {
			"gid": "67890",
			"resource_type": "task",
			"name": "Test Task"
		},
		"change": {
			"field": "custom_fields",
			"action": "changed",
			"new_value": {
				"gid": "field123",
				"name": "Status Field",
				"display_value": "In Progress"
			}
		}
	}`

	var event Event
	err := json.Unmarshal([]byte(eventJSON), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if event.Action != "changed" {
		t.Errorf("Expected action 'changed', got '%s'", event.Action)
	}

	if event.User == nil || event.User.Name != "Test User" {
		t.Error("User data not properly unmarshaled")
	}

	if event.Change == nil || event.Change.Field != "custom_fields" {
		t.Error("Change data not properly unmarshaled")
	}

	// NewValue should be an interface{} that can hold complex objects
	if event.Change.NewValue == nil {
		t.Error("NewValue should not be nil")
	}

	// Check that we can access nested fields in NewValue
	if newValueMap, ok := event.Change.NewValue.(map[string]interface{}); ok {
		if gid, exists := newValueMap["gid"]; !exists || gid != "field123" {
			t.Error("NewValue GID not properly accessible")
		}
	} else {
		t.Error("NewValue should be a map")
	}
}
