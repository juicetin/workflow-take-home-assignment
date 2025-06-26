package workflow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"workflow-code-test/api/internal/models"
)

func (s *Service) HandleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	slog.Debug("Getting workflow definition for id", "id", id)
	
	// Parse workflow ID
	workflowID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Invalid workflow ID", "id", id, "error", err)
		http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
		return
	}
	
	// Get workflow with nodes and edges
	workflow, err := s.workflowService.GetWorkflowWithNodesAndEdges(r.Context(), workflowID)
	if err != nil {
		slog.Error("Failed to get workflow", "id", id, "error", err)
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(workflow); err != nil {
		slog.Error("Failed to encode workflow response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Service) HandleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	slog.Debug("Handling workflow execution for id", "id", id)

	// Parse workflow ID to validate format
	if _, err := uuid.Parse(id); err != nil {
		slog.Error("Invalid workflow ID", "id", id, "error", err)
		http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var executeRequest struct {
		FormData  map[string]interface{} `json:"formData"`
		Condition map[string]interface{} `json:"condition"`
		Nodes     []models.NodeRequest   `json:"nodes"`
		Edges     []models.EdgeRequest   `json:"edges"`
	}

	if err := json.NewDecoder(r.Body).Decode(&executeRequest); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create workflow request for persistence
	workflowRequest := &models.WorkflowRequest{
		ID:    id,
		Name:  "Weather Alert Workflow", // TODO: Make this configurable
		Nodes: executeRequest.Nodes,
		Edges: executeRequest.Edges,
	}

	// Save/update workflow definition in database
	if err := s.workflowService.SaveWorkflowFromRequest(r.Context(), workflowRequest); err != nil {
		slog.Error("Failed to save workflow", "id", id, "error", err)
		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
		return
	}

	slog.Info("Workflow saved/updated successfully", "id", id)

	// Generate current timestamp for execution response
	currentTime := time.Now().Format(time.RFC3339)

	// this is the result of an execution
	executionJSON := fmt.Sprintf(`{
		"executedAt": "%s",
		"status": "completed",
		"steps": [
			{
				"nodeId": "start",
				"type": "start",
				"label": "Start",
				"description": "Begin weather check workflow",
				"status": "completed"
			},
			{
				"nodeId": "form",
				"type": "form",
				"label": "User Input",
				"description": "Process collected data - name, email, location",
				"status": "completed",
				"output": {
					"name": "Alice",
					"email": "alice@example.com",
					"city": "Sydney"
				}
			},
			{
				"nodeId": "weather-api",
				"type": "integration",
				"label": "Weather API",
				"description": "Fetch current temperature for Sydney",
				"status": "completed",
				"output": {
					"temperature": 28.5,
					"location": "Sydney"
				}
			},
			{
				"nodeId": "condition",
				"type": "condition",
				"label": "Check Condition",
				"description": "Evaluate temperature threshold",
				"status": "completed",
				"output": {
					"conditionMet": true,
					"threshold": 25,
					"operator": "greater_than",
					"actualValue": 28.5,
					"message": "Temperature 28.5°C is greater than 25°C - condition met"
				}
			},
			{
				"nodeId": "email",
				"type": "email",
				"label": "Send Alert",
				"description": "Email weather alert notification",
				"status": "completed",
				"output": {
					"emailDraft": {
						"to": "alice@example.com",
						"from": "weather-alerts@example.com",
						"subject": "Weather Alert",
						"body": "Weather alert for Sydney! Temperature is 28.5°C!",
						"timestamp": "2024-01-15T14:30:24.856Z"
					},
					"deliveryStatus": "sent",
					"messageId": "msg_abc123def456",
					"emailSent": true
				}
			},
			{
				"nodeId": "end",
				"type": "end",
				"label": "Complete",
				"description": "Workflow execution finished",
				"status": "completed"
			}
		]
	}`, currentTime)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(executionJSON))
}