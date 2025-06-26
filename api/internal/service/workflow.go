package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"workflow-code-test/api/internal/execution"
	"workflow-code-test/api/internal/models"
	"workflow-code-test/api/internal/repository"
)

type WorkflowService struct {
	repo           *repository.WorkflowRepository
	executionEngine *execution.Engine
}

func NewWorkflowService(repo *repository.WorkflowRepository) *WorkflowService {
	// Initialize execution services
	weatherService := execution.NewDefaultWeatherService(false) // Use real API
	emailService := execution.NewInMemoryEmailService()
	validator := execution.NewDefaultInputValidator()
	
	// Create execution engine
	executionEngine := execution.NewEngine(weatherService, emailService, validator)
	
	return &WorkflowService{
		repo:           repo,
		executionEngine: executionEngine,
	}
}

// GetWorkflowWithNodesAndEdges retrieves a complete workflow with all its nodes and edges
func (s *WorkflowService) GetWorkflowWithNodesAndEdges(ctx context.Context, workflowID uuid.UUID) (*models.WorkflowResponse, error) {
	// Get the workflow
	workflow, err := s.repo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	
	// Get all nodes for the workflow
	nodes, err := s.repo.GetNodesByWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	
	// Get all edges for the workflow
	edges, err := s.repo.GetEdgesByWorkflow(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}
	
	// Convert to response format
	nodeResponses := make([]models.NodeResponse, len(nodes))
	for i, node := range nodes {
		nodeResponses[i] = node.ToResponse()
	}
	
	edgeResponses := make([]models.EdgeResponse, len(edges))
	for i, edge := range edges {
		edgeResponses[i] = edge.ToResponse()
	}
	
	response := &models.WorkflowResponse{
		ID:    workflow.ID.String(),
		Name:  workflow.Name,
		Nodes: nodeResponses,
		Edges: edgeResponses,
	}
	
	return response, nil
}

// SaveWorkflowFromRequest saves a workflow from a frontend request
func (s *WorkflowService) SaveWorkflowFromRequest(ctx context.Context, req *models.WorkflowRequest) error {
	// Parse workflow ID
	workflowID, err := uuid.Parse(req.ID)
	if err != nil {
		return fmt.Errorf("invalid workflow ID: %w", err)
	}
	
	// Validate nodes and edges
	if err := s.validateWorkflowRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Create workflow entity
	workflow := &models.Workflow{
		ID:   workflowID,
		Name: req.Name,
	}
	
	// Convert request nodes to entities
	nodes := make([]models.Node, len(req.Nodes))
	for i, nodeReq := range req.Nodes {
		node := nodeReq.ToNode()
		node.WorkflowID = workflowID
		nodes[i] = *node
	}
	
	// Convert request edges to entities
	edges := make([]models.Edge, len(req.Edges))
	for i, edgeReq := range req.Edges {
		edge := edgeReq.ToEdge()
		edge.WorkflowID = workflowID
		edges[i] = *edge
	}
	
	// Save to database
	return s.repo.SaveWorkflow(ctx, workflow, nodes, edges)
}

// validateWorkflowRequest validates the workflow request
func (s *WorkflowService) validateWorkflowRequest(req *models.WorkflowRequest) error {
	// Validate nodes
	for _, node := range req.Nodes {
		if err := node.Validate(); err != nil {
			return fmt.Errorf("invalid node %s: %w", node.ID, err)
		}
	}
	
	// Validate edges
	for _, edge := range req.Edges {
		if err := edge.Validate(); err != nil {
			return fmt.Errorf("invalid edge %s: %w", edge.ID, err)
		}
	}
	
	// Check that we have exactly one start node and one end node
	startNodes := 0
	endNodes := 0
	
	for _, node := range req.Nodes {
		switch node.Type {
		case models.NodeTypeStart:
			startNodes++
		case models.NodeTypeEnd:
			endNodes++
		}
	}
	
	if startNodes != 1 {
		return fmt.Errorf("workflow must have exactly one start node, found %d", startNodes)
	}
	
	if endNodes != 1 {
		return fmt.Errorf("workflow must have exactly one end node, found %d", endNodes)
	}
	
	return nil
}

// ExecuteWorkflow executes a workflow using the execution engine
func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, workflow *models.WorkflowResponse, req *models.ExecutionRequest) (*models.ExecutionResponse, error) {
	return s.executionEngine.ExecuteWorkflow(ctx, workflow, req)
}

