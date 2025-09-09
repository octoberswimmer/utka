package cmd

import (
	"encoding/json"
	"testing"

	"github.com/expr-lang/expr"
)

func TestEventFiltering(t *testing.T) {
	tests := []struct {
		name      string
		events    []map[string]interface{}
		filter    string
		wantCount int
		wantError bool
	}{
		{
			name: "filter by action",
			events: []map[string]interface{}{
				{"action": "changed", "type": "task"},
				{"action": "added", "type": "task"},
				{"action": "changed", "type": "task"},
			},
			filter:    `event.action == "changed"`,
			wantCount: 2,
			wantError: false,
		},
		{
			name: "filter by nested field",
			events: []map[string]interface{}{
				{
					"action": "changed",
					"resource": map[string]interface{}{
						"resource_subtype": "default_task",
					},
				},
				{
					"action": "changed",
					"resource": map[string]interface{}{
						"resource_subtype": "milestone",
					},
				},
			},
			filter:    `event.action == "changed" && event.resource.resource_subtype == "default_task"`,
			wantCount: 1,
			wantError: false,
		},
		{
			name: "filter with nil change field",
			events: []map[string]interface{}{
				{
					"action": "changed",
					"change": map[string]interface{}{
						"field": "name",
					},
				},
				{
					"action": "added",
					// no change field
				},
			},
			filter:    `event.change.field == "name"`,
			wantCount: 1,
			wantError: false,
		},
		{
			name: "filter by custom field new_value",
			events: []map[string]interface{}{
				{
					"action": "changed",
					"change": map[string]interface{}{
						"field": "custom_fields",
						"new_value": map[string]interface{}{
							"gid":           "123456",
							"display_value": "In Progress",
						},
					},
				},
				{
					"action": "changed",
					"change": map[string]interface{}{
						"field": "custom_fields",
						"new_value": map[string]interface{}{
							"gid":           "789012",
							"display_value": "Done",
						},
					},
				},
			},
			filter:    `event.change.new_value.gid == "123456"`,
			wantCount: 1,
			wantError: false,
		},
		{
			name: "complex filter with multiple conditions",
			events: []map[string]interface{}{
				{
					"action": "changed",
					"type":   "task",
					"user": map[string]interface{}{
						"name": "John Doe",
					},
					"change": map[string]interface{}{
						"field": "name",
					},
				},
				{
					"action": "changed",
					"type":   "task",
					"user": map[string]interface{}{
						"name": "Jane Smith",
					},
					"change": map[string]interface{}{
						"field": "name",
					},
				},
				{
					"action": "added",
					"type":   "task",
					"user": map[string]interface{}{
						"name": "John Doe",
					},
				},
			},
			filter:    `event.action == "changed" && event.user.name == "John Doe" && event.change.field == "name"`,
			wantCount: 1,
			wantError: false,
		},
		{
			name: "filter handles nil gracefully",
			events: []map[string]interface{}{
				{
					"action": "changed",
					"change": nil,
				},
				{
					"action": "changed",
					"change": map[string]interface{}{
						"new_value": nil,
					},
				},
				{
					"action": "changed",
					"change": map[string]interface{}{
						"new_value": map[string]interface{}{
							"gid": "valid",
						},
					},
				},
			},
			filter:    `event.change.new_value.gid == "valid"`,
			wantCount: 1,
			wantError: false,
		},
		{
			name: "invalid expression syntax",
			events: []map[string]interface{}{
				{"action": "changed"},
			},
			filter:    `event.action ==`,
			wantCount: 0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile the expression
			program, err := expr.Compile(tt.filter, expr.AsBool())
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected compilation error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			// Filter events
			var filteredEvents []map[string]interface{}
			for _, event := range tt.events {
				env := map[string]interface{}{
					"event": event,
				}

				result, err := expr.Run(program, env)
				if err != nil {
					// Skip events that cause evaluation errors
					continue
				}

				if result == true {
					filteredEvents = append(filteredEvents, event)
				}
			}

			if len(filteredEvents) != tt.wantCount {
				t.Errorf("Expected %d filtered events, got %d", tt.wantCount, len(filteredEvents))
			}
		})
	}
}

func TestEventFilteringWithRealJSON(t *testing.T) {
	// Test with actual JSON structure from Asana
	eventJSON := `[
		{
			"action": "changed",
			"type": "task",
			"created_at": "2025-09-09T15:29:56.550Z",
			"change": {
				"field": "custom_fields",
				"action": "changed",
				"new_value": {
					"gid": "1210930954402852",
					"name": "Marketing Materials Status",
					"display_value": "Client Meeting or Introduction",
					"enum_value": {
						"gid": "1210930954402856",
						"name": "Client Meeting or Introduction"
					}
				}
			}
		},
		{
			"action": "changed",
			"type": "task",
			"created_at": "2025-09-09T15:29:26.698Z",
			"change": {
				"field": "completed",
				"action": "changed"
			}
		},
		{
			"action": "added",
			"type": "task",
			"created_at": "2025-09-09T15:14:59.602Z",
			"resource": {
				"gid": "1211305321446116",
				"resource_type": "task",
				"name": "Another subtask",
				"resource_subtype": "default_task"
			}
		}
	]`

	var events []map[string]interface{}
	if err := json.Unmarshal([]byte(eventJSON), &events); err != nil {
		t.Fatalf("Failed to unmarshal events: %v", err)
	}

	tests := []struct {
		name      string
		filter    string
		wantCount int
	}{
		{
			name:      "filter by custom field gid",
			filter:    `event.change.new_value.gid == "1210930954402852"`,
			wantCount: 1,
		},
		{
			name:      "filter by display value",
			filter:    `event.change.new_value.display_value == "Client Meeting or Introduction"`,
			wantCount: 1,
		},
		{
			name:      "filter by change field",
			filter:    `event.change.field == "completed"`,
			wantCount: 1,
		},
		{
			name:      "filter by action and resource subtype",
			filter:    `event.action == "added" && event.resource.resource_subtype == "default_task"`,
			wantCount: 1,
		},
		{
			name:      "filter all changed events",
			filter:    `event.action == "changed"`,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.filter, expr.AsBool())
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			var filteredEvents []map[string]interface{}
			for _, event := range events {
				env := map[string]interface{}{
					"event": event,
				}

				result, err := expr.Run(program, env)
				if err != nil {
					continue
				}

				if result == true {
					filteredEvents = append(filteredEvents, event)
				}
			}

			if len(filteredEvents) != tt.wantCount {
				t.Errorf("Expected %d filtered events, got %d", tt.wantCount, len(filteredEvents))
				for i, event := range filteredEvents {
					eventJSON, _ := json.MarshalIndent(event, "", "  ")
					t.Logf("Event %d: %s", i, string(eventJSON))
				}
			}
		})
	}
}
