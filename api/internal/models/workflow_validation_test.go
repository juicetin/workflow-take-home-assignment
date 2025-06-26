package models

import (
	"testing"
)

func TestWorkflowRequest_ValidateWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		workflow       WorkflowRequest
		expectedErrors int
		expectedFields []string
	}{
		{
			name: "valid workflow with start and end nodes and valid path",
			workflow: WorkflowRequest{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "form-1"},
					{ID: "edge-2", Source: "form-1", Target: "end-1"},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "missing start node",
			workflow: WorkflowRequest{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Nodes: []NodeRequest{
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "form-1", Target: "end-1"},
				},
			},
			expectedErrors: 1,
			expectedFields: []string{"nodes"},
		},
		{
			name: "missing end node",
			workflow: WorkflowRequest{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "form-1"},
				},
			},
			expectedErrors: 1,
			expectedFields: []string{"nodes"},
		},
		{
			name: "multiple start nodes",
			workflow: WorkflowRequest{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "start-2", Type: NodeTypeStart},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "end-1"},
					{ID: "edge-2", Source: "start-2", Target: "end-1"},
				},
			},
			expectedErrors: 1,
			expectedFields: []string{"nodes"},
		},
		{
			name: "no path from start to end",
			workflow: WorkflowRequest{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "form-1"},
					// Missing edge from form-1 to end-1
				},
			},
			expectedErrors: 1,
			expectedFields: []string{"edges"},
		},
		{
			name: "complex valid workflow with multiple paths",
			workflow: WorkflowRequest{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "condition-1", Type: NodeTypeCondition},
					{ID: "email-1", Type: NodeTypeEmail},
					{ID: "integration-1", Type: NodeTypeIntegration},
					{ID: "end-1", Type: NodeTypeEnd},
					{ID: "end-2", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "condition-1"},
					{ID: "edge-2", Source: "condition-1", Target: "email-1"},
					{ID: "edge-3", Source: "condition-1", Target: "integration-1"},
					{ID: "edge-4", Source: "email-1", Target: "end-1"},
					{ID: "edge-5", Source: "integration-1", Target: "end-2"},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "disconnected nodes",
			workflow: WorkflowRequest{
				ID:   "test-workflow",
				Name: "Test Workflow",
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "end-1", Type: NodeTypeEnd},
					{ID: "isolated-1", Type: NodeTypeEmail}, // isolated node
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "form-1"},
					{ID: "edge-2", Source: "form-1", Target: "end-1"},
					// isolated-1 has no connections
				},
			},
			expectedErrors: 0, // This is valid - isolated nodes don't break the main path
		},
		{
			name: "all validation errors",
			workflow: WorkflowRequest{
				ID:    "test-workflow",
				Name:  "Invalid Workflow",
				Nodes: []NodeRequest{
					{ID: "form-1", Type: NodeTypeForm},
					// No start or end nodes
				},
				Edges: []EdgeRequest{
					// No edges
				},
			},
			expectedErrors: 2,
			expectedFields: []string{"nodes", "nodes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.workflow.ValidateWorkflow()
			
			if len(errors) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
			}
			
			if tt.expectedFields != nil {
				if len(errors) != len(tt.expectedFields) {
					t.Errorf("expected %d field errors, got %d", len(tt.expectedFields), len(errors))
				}
				
				for i, expectedField := range tt.expectedFields {
					if i < len(errors) && errors[i].Field != expectedField {
						t.Errorf("expected error field %s, got %s", expectedField, errors[i].Field)
					}
				}
			}
		})
	}
}

func TestWorkflowRequest_hasStartNode(t *testing.T) {
	tests := []struct {
		name     string
		workflow WorkflowRequest
		expected bool
	}{
		{
			name: "has exactly one start node",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
				},
			},
			expected: true,
		},
		{
			name: "has no start node",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "end-1", Type: NodeTypeEnd},
				},
			},
			expected: false,
		},
		{
			name: "has multiple start nodes",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "start-2", Type: NodeTypeStart},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.workflow.hasStartNode()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWorkflowRequest_hasEndNode(t *testing.T) {
	tests := []struct {
		name     string
		workflow WorkflowRequest
		expected bool
	}{
		{
			name: "has end node",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "end-1", Type: NodeTypeEnd},
				},
			},
			expected: true,
		},
		{
			name: "has multiple end nodes",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "end-1", Type: NodeTypeEnd},
					{ID: "end-2", Type: NodeTypeEnd},
				},
			},
			expected: true,
		},
		{
			name: "has no end node",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.workflow.hasEndNode()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWorkflowRequest_hasValidPath(t *testing.T) {
	tests := []struct {
		name     string
		workflow WorkflowRequest
		expected bool
	}{
		{
			name: "simple valid path",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "end-1"},
				},
			},
			expected: true,
		},
		{
			name: "complex valid path",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "condition-1", Type: NodeTypeCondition},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "form-1"},
					{ID: "edge-2", Source: "form-1", Target: "condition-1"},
					{ID: "edge-3", Source: "condition-1", Target: "end-1"},
				},
			},
			expected: true,
		},
		{
			name: "branching path to multiple end nodes",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "condition-1", Type: NodeTypeCondition},
					{ID: "end-1", Type: NodeTypeEnd},
					{ID: "end-2", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "condition-1"},
					{ID: "edge-2", Source: "condition-1", Target: "end-1"},
					{ID: "edge-3", Source: "condition-1", Target: "end-2"},
				},
			},
			expected: true,
		},
		{
			name: "no path to end",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "form-1"},
					// Missing edge from form-1 to end-1
				},
			},
			expected: false,
		},
		{
			name: "cycle but no path to end",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "form-1", Type: NodeTypeForm},
					{ID: "form-2", Type: NodeTypeForm},
					{ID: "end-1", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "form-1"},
					{ID: "edge-2", Source: "form-1", Target: "form-2"},
					{ID: "edge-3", Source: "form-2", Target: "form-1"}, // cycle
					// No path to end-1
				},
			},
			expected: false,
		},
		{
			name: "multiple end nodes but not all reachable",
			workflow: WorkflowRequest{
				Nodes: []NodeRequest{
					{ID: "start-1", Type: NodeTypeStart},
					{ID: "condition-1", Type: NodeTypeCondition},
					{ID: "end-1", Type: NodeTypeEnd},
					{ID: "end-2", Type: NodeTypeEnd},
				},
				Edges: []EdgeRequest{
					{ID: "edge-1", Source: "start-1", Target: "condition-1"},
					{ID: "edge-2", Source: "condition-1", Target: "end-1"},
					// Missing edge to end-2, so end-2 is not reachable
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.workflow.hasValidPath()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "nodes",
		Message: "workflow must have exactly one start node",
	}
	
	expected := "workflow must have exactly one start node"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}