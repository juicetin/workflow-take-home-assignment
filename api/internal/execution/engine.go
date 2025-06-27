package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"workflow-code-test/api/internal/models"
)

// Engine handles workflow execution logic
type Engine struct {
	integrationService *IntegrationService
	emailService       EmailService
	validator          InputValidator
}

// EmailService defines the interface for email operations
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// InputValidator defines the interface for input validation
type InputValidator interface {
	ValidateFormData(formData map[string]interface{}, nodeData *models.FormNodeData) error
}

// APIClient interface for making HTTP calls
type APIClient interface {
	CallAPI(ctx context.Context, url string) (map[string]interface{}, error)
}

// NewEngine creates a new workflow execution engine
func NewEngine(emailService EmailService, validator InputValidator) *Engine {
	return &Engine{
		integrationService: NewIntegrationService(NewHTTPAPIClient()),
		emailService:       emailService,
		validator:          validator,
	}
}

// NewEngineWithAPIClient creates a new workflow execution engine with custom API client
func NewEngineWithAPIClient(emailService EmailService, validator InputValidator, apiClient APIClient) *Engine {
	return &Engine{
		integrationService: NewIntegrationService(apiClient),
		emailService:       emailService,
		validator:          validator,
	}
}

// ExecuteWorkflow executes a workflow in memory
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflow *models.WorkflowResponse, req *models.ExecutionRequest) (*models.ExecutionResponse, error) {
	execCtx := models.NewExecutionContext(workflow.ID, req.FormData)

	// Store condition data in context for later use
	if req.Condition != nil {
		for key, value := range req.Condition {
			execCtx.SetVariable("condition_"+key, value)
		}
	}

	// Build node and edge maps for efficient lookup
	nodeMap := make(map[string]*models.NodeResponse)
	for i := range workflow.Nodes {
		nodeMap[workflow.Nodes[i].ID] = &workflow.Nodes[i]
	}

	edgeMap := make(map[string][]models.EdgeResponse) // source -> []edges
	for _, edge := range workflow.Edges {
		edgeMap[edge.Source] = append(edgeMap[edge.Source], edge)
	}

	// Find start node
	var startNode *models.NodeResponse
	for _, node := range workflow.Nodes {
		if node.Type == models.NodeTypeStart {
			startNode = &node
			break
		}
	}

	if startNode == nil {
		return &models.ExecutionResponse{
			ExecutedAt: time.Now(),
			Status:     "failed",
			Steps:      execCtx.Steps,
			Error:      stringPtr("no start node found"),
		}, nil
	}

	// Execute workflow starting from start node
	if err := e.executeNode(ctx, startNode, nodeMap, edgeMap, execCtx); err != nil {
		return &models.ExecutionResponse{
			ExecutedAt: time.Now(),
			Status:     "failed",
			Steps:      execCtx.Steps,
			Error:      stringPtr(err.Error()),
		}, nil
	}

	return &models.ExecutionResponse{
		ExecutedAt: time.Now(),
		Status:     "completed",
		Steps:      execCtx.Steps,
	}, nil
}

// executeNode executes a single node and continues to next nodes
func (e *Engine) executeNode(ctx context.Context, node *models.NodeResponse, nodeMap map[string]*models.NodeResponse, edgeMap map[string][]models.EdgeResponse, execCtx *models.ExecutionContext) error {
	stepStart := time.Now()

	step := models.ExecutionStep{
		NodeID:      node.ID,
		Type:        node.Type,
		Label:       e.getNodeLabel(node),
		Description: e.getNodeDescription(node),
		Status:      "running",
	}

	var err error
	var output interface{}

	// Execute node based on type
	switch node.Type {
	case models.NodeTypeStart:
		output, err = e.executeStartNode(ctx, node, execCtx)

	case models.NodeTypeForm:
		output, err = e.executeFormNode(ctx, node, execCtx)

	case models.NodeTypeIntegration:
		output, err = e.executeIntegrationNode(ctx, node, execCtx)

	case models.NodeTypeCondition:
		output, err = e.executeConditionNode(ctx, node, execCtx)

	case models.NodeTypeEmail:
		output, err = e.executeEmailNode(ctx, node, execCtx)

	case models.NodeTypeEnd:
		output, err = e.executeEndNode(ctx, node, execCtx)

	default:
		err = fmt.Errorf("unsupported node type: %s", node.Type)
	}

	// Update step with results
	duration := time.Since(stepStart).Milliseconds()
	step.Duration = &duration

	if err != nil {
		step.Status = "failed"
		step.Error = stringPtr(err.Error())
	} else {
		step.Status = "completed"
		if output != nil {
			outputBytes, _ := json.Marshal(output)
			step.Output = outputBytes
		}
	}

	execCtx.AddStep(step)

	if err != nil {
		return err
	}

	// Continue to next nodes based on node type and condition results
	if err := e.continueToNextNodes(ctx, node, nodeMap, edgeMap, execCtx); err != nil {
		return err
	}

	return nil
}

// continueToNextNodes determines which nodes to execute next
func (e *Engine) continueToNextNodes(ctx context.Context, currentNode *models.NodeResponse, nodeMap map[string]*models.NodeResponse, edgeMap map[string][]models.EdgeResponse, execCtx *models.ExecutionContext) error {
	edges := edgeMap[currentNode.ID]

	// For non-condition nodes, follow all edges
	if currentNode.Type != models.NodeTypeCondition {
		return e.executeNextNodes(ctx, edges, nodeMap, edgeMap, execCtx)
	}

	// Special handling for condition nodes
	conditionMet, ok := execCtx.GetVariable("conditionMet")
	if !ok {
		return fmt.Errorf("condition result not found in context")
	}

	conditionResult, ok := conditionMet.(bool)
	if !ok {
		return fmt.Errorf("condition result must be boolean")
	}

	// Find the appropriate edges based on condition result
	for _, edge := range edges {
		if e.shouldFollowEdge(edge, conditionResult) {
			nextNode := nodeMap[edge.Target]
			if nextNode == nil {
				return fmt.Errorf("next node not found: %s", edge.Target)
			}

			if err := e.executeNode(ctx, nextNode, nodeMap, edgeMap, execCtx); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeNextNodes executes all connected nodes for non-condition nodes
func (e *Engine) executeNextNodes(ctx context.Context, edges []models.EdgeResponse, nodeMap map[string]*models.NodeResponse, edgeMap map[string][]models.EdgeResponse, execCtx *models.ExecutionContext) error {
	for _, edge := range edges {
		nextNode := nodeMap[edge.Target]
		if nextNode == nil {
			return fmt.Errorf("next node not found: %s", edge.Target)
		}

		if err := e.executeNode(ctx, nextNode, nodeMap, edgeMap, execCtx); err != nil {
			return err
		}
	}
	return nil
}

// shouldFollowEdge determines if an edge should be followed based on condition result
func (e *Engine) shouldFollowEdge(edge models.EdgeResponse, conditionResult bool) bool {
	if edge.SourceHandle != nil {
		// Handle explicit source handles (true/false)
		return (*edge.SourceHandle == "true" && conditionResult) ||
			(*edge.SourceHandle == "false" && !conditionResult)
	}
	// If no source handle specified, follow if condition is true
	return conditionResult
}

// executeStartNode executes a start node
func (e *Engine) executeStartNode(ctx context.Context, node *models.NodeResponse, execCtx *models.ExecutionContext) (interface{}, error) {
	slog.Debug("Executing start node", "nodeId", node.ID)
	return map[string]interface{}{
		"message": "Begin weather check workflow",
		"nodeId":  node.ID,
	}, nil
}

// executeFormNode executes a form node with input validation
func (e *Engine) executeFormNode(ctx context.Context, node *models.NodeResponse, execCtx *models.ExecutionContext) (interface{}, error) {
	slog.Debug("Executing form node", "nodeId", node.ID)

	// Parse node data if available for validation
	var formData models.FormNodeData
	if len(node.Data) > 0 {
		if err := json.Unmarshal(node.Data, &formData); err != nil {
			slog.Warn("Failed to parse form node data, proceeding without validation", "error", err)
		} else {
			// Validate form input
			if err := e.validator.ValidateFormData(execCtx.FormData, &formData); err != nil {
				return nil, fmt.Errorf("form validation failed: %w", err)
			}
		}
	}

	// Store validated form data in context
	for key, value := range execCtx.FormData {
		execCtx.SetVariable(key, value)
	}

	return execCtx.FormData, nil
}

// executeIntegrationNode executes an integration node (API call)
func (e *Engine) executeIntegrationNode(ctx context.Context, node *models.NodeResponse, execCtx *models.ExecutionContext) (interface{}, error) {
	slog.Debug("Executing integration node", "nodeId", node.ID)

	// Prepare input variables for the integration
	inputVariables := make(map[string]interface{})
	for key, value := range execCtx.FormData {
		inputVariables[key] = value
	}

	// Execute the integration
	result, err := e.integrationService.ExecuteIntegration(ctx, node.Data, inputVariables)
	if err != nil {
		return nil, fmt.Errorf("integration execution failed: %w", err)
	}

	// Store results in execution context for downstream nodes
	if temperature, ok := result["temperature"]; ok {
		execCtx.SetVariable("temperature", temperature)
	}
	if location, ok := result["location"]; ok {
		execCtx.SetVariable("location", location)
	}

	return result, nil
}

// executeConditionNode executes a condition node with threshold comparison
func (e *Engine) executeConditionNode(ctx context.Context, node *models.NodeResponse, execCtx *models.ExecutionContext) (interface{}, error) {
	slog.Debug("Executing condition node", "nodeId", node.ID)

	// Get condition parameters from execution context (passed from frontend)
	operatorValue, ok := execCtx.GetVariable("condition_operator")
	if !ok {
		return nil, fmt.Errorf("condition operator not found")
	}

	thresholdValue, ok := execCtx.GetVariable("condition_threshold")
	if !ok {
		return nil, fmt.Errorf("condition threshold not found")
	}

	operator, ok := operatorValue.(string)
	if !ok {
		return nil, fmt.Errorf("condition operator must be a string")
	}

	threshold, ok := thresholdValue.(float64)
	if !ok {
		return nil, fmt.Errorf("condition threshold must be a number")
	}

	// Get temperature from context
	temperatureValue, ok := execCtx.GetVariable("temperature")
	if !ok {
		return nil, fmt.Errorf("temperature not found in execution context")
	}

	temperature, ok := temperatureValue.(float64)
	if !ok {
		return nil, fmt.Errorf("temperature must be a number")
	}

	// Evaluate condition based on frontend operator strings
	conditionMet, err := e.evaluateCondition(temperature, operator, threshold)
	if err != nil {
		return nil, fmt.Errorf("condition evaluation failed: %w", err)
	}

	// Store condition result in context for edge routing
	execCtx.SetVariable("conditionMet", conditionMet)

	output := map[string]interface{}{
		"conditionMet": conditionMet,
		"operator":     operator,
		"threshold":    threshold,
		"actualValue":  temperature,
		"message":      fmt.Sprintf("Temperature %.1f°C %s %.1f°C - condition %s", temperature, e.getOperatorSymbol(operator), threshold, map[bool]string{true: "met", false: "not met"}[conditionMet]),
	}

	return output, nil
}

// executeEmailNode executes an email node
func (e *Engine) executeEmailNode(ctx context.Context, node *models.NodeResponse, execCtx *models.ExecutionContext) (interface{}, error) {
	slog.Debug("Executing email node", "nodeId", node.ID)

	// Get recipient email from form data
	emailValue, ok := execCtx.GetVariable("email")
	if !ok {
		return nil, fmt.Errorf("email field not found in form data")
	}

	toEmail, ok := emailValue.(string)
	if !ok {
		return nil, fmt.Errorf("email field must be a string")
	}

	// Get name and location for email content
	name, _ := execCtx.GetVariable("name")
	location, _ := execCtx.GetVariable("location")
	temperature, _ := execCtx.GetVariable("temperature")

	// Build email content
	subject := "Weather Alert"
	body := fmt.Sprintf("Weather alert for %s! Temperature is %.1f°C!", location, temperature)
	if nameStr, ok := name.(string); ok {
		body = fmt.Sprintf("Hi %s, %s", nameStr, strings.ToLower(body))
	}

	fromEmail := "weather-alerts@example.com"

	// Send email (mock implementation)
	if err := e.emailService.SendEmail(ctx, toEmail, subject, body); err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	output := map[string]interface{}{
		"emailDraft": map[string]interface{}{
			"to":        toEmail,
			"from":      fromEmail,
			"subject":   subject,
			"body":      body,
			"timestamp": time.Now().Format(time.RFC3339),
		},
		"deliveryStatus": "sent",
		"messageId":      fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		"emailSent":      true,
	}

	return output, nil
}

// executeEndNode executes an end node
func (e *Engine) executeEndNode(ctx context.Context, node *models.NodeResponse, execCtx *models.ExecutionContext) (interface{}, error) {
	slog.Debug("Executing end node", "nodeId", node.ID)
	return map[string]interface{}{
		"message": "Workflow execution finished",
		"nodeId":  node.ID,
	}, nil
}

// Helper functions

func (e *Engine) getNodeLabel(node *models.NodeResponse) string {
	switch node.Type {
	case models.NodeTypeStart:
		return "Start"
	case models.NodeTypeForm:
		return "User Input"
	case models.NodeTypeIntegration:
		return "Weather API"
	case models.NodeTypeCondition:
		return "Check Condition"
	case models.NodeTypeEmail:
		return "Send Alert"
	case models.NodeTypeEnd:
		return "Complete"
	default:
		return "Unknown"
	}
}

func (e *Engine) getNodeDescription(node *models.NodeResponse) string {
	switch node.Type {
	case models.NodeTypeStart:
		return "Begin weather check workflow"
	case models.NodeTypeForm:
		return "Process collected data - name, email, location"
	case models.NodeTypeIntegration:
		return "Fetch current temperature"
	case models.NodeTypeCondition:
		return "Evaluate temperature threshold"
	case models.NodeTypeEmail:
		return "Email weather alert notification"
	case models.NodeTypeEnd:
		return "Workflow execution finished"
	default:
		return "Unknown node type"
	}
}

// evaluateCondition evaluates condition using frontend operator strings
func (e *Engine) evaluateCondition(temperature float64, operator string, threshold float64) (bool, error) {
	switch operator {
	case "greater_than":
		return temperature > threshold, nil
	case "less_than":
		return temperature < threshold, nil
	case "equals":
		return temperature == threshold, nil
	case "greater_than_or_equal":
		return temperature >= threshold, nil
	case "less_than_or_equal":
		return temperature <= threshold, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// getOperatorSymbol returns the mathematical symbol for the operator
func (e *Engine) getOperatorSymbol(operator string) string {
	switch operator {
	case "greater_than":
		return ">"
	case "less_than":
		return "<"
	case "equals":
		return "="
	case "greater_than_or_equal":
		return "≥"
	case "less_than_or_equal":
		return "≤"
	default:
		return "?"
	}
}

func stringPtr(s string) *string {
	return &s
}
