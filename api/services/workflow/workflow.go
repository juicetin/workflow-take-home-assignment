package workflow

import (
	"encoding/json"
	"log/slog"
	"net/http"

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
	workflowID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Invalid workflow ID", "id", id, "error", err)
		http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var executeRequest models.ExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&executeRequest); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get workflow definition from database
	workflow, err := s.workflowService.GetWorkflowWithNodesAndEdges(r.Context(), workflowID)
	if err != nil {
		slog.Error("Failed to get workflow", "id", id, "error", err)
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	// Execute workflow using the execution engine
	executionResult, err := s.workflowService.ExecuteWorkflow(r.Context(), workflow, &executeRequest)
	if err != nil {
		slog.Error("Failed to execute workflow", "id", id, "error", err)
		http.Error(w, "Workflow execution failed", http.StatusInternalServerError)
		return
	}

	// Save updated workflow positions if nodes and edges are provided in the request
	if len(executeRequest.Nodes) > 0 || len(executeRequest.Edges) > 0 {
		slog.Debug("Saving updated workflow positions", "nodeCount", len(executeRequest.Nodes), "edgeCount", len(executeRequest.Edges))
		
		workflowRequest := &models.WorkflowRequest{
			ID:    id,
			Name:  workflow.Name,
			Nodes: executeRequest.Nodes,
			Edges: executeRequest.Edges,
		}

		if err := s.workflowService.SaveWorkflowFromRequest(r.Context(), workflowRequest); err != nil {
			slog.Error("Failed to save updated workflow positions", "id", id, "error", err)
			// Don't fail the execution if saving positions fails - just log it
		} else {
			slog.Debug("Successfully saved updated workflow positions", "id", id)
		}
	}

	// Return execution result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(executionResult); err != nil {
		slog.Error("Failed to encode execution result", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
