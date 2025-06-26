package models

// ValidationError represents a workflow validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface for ValidationError
func (ve ValidationError) Error() string {
	return ve.Message
}

// ValidateWorkflow validates that a workflow has the required start and end nodes
// and that there's a valid path from start to end following the edges
func (wr *WorkflowRequest) ValidateWorkflow() []ValidationError {
	var errors []ValidationError
	
	// Check for start and end nodes
	if !wr.hasStartNode() {
		errors = append(errors, ValidationError{
			Field:   "nodes",
			Message: "workflow must have exactly one start node",
		})
	}
	
	if !wr.hasEndNode() {
		errors = append(errors, ValidationError{
			Field:   "nodes", 
			Message: "workflow must have at least one end node",
		})
	}
	
	// Check path connectivity if we have both start and end nodes
	if wr.hasStartNode() && wr.hasEndNode() {
		if !wr.hasValidPath() {
			errors = append(errors, ValidationError{
				Field:   "edges",
				Message: "no valid path exists from start node to all end nodes",
			})
		}
	}
	
	return errors
}

// hasStartNode checks if there's exactly one start node
func (wr *WorkflowRequest) hasStartNode() bool {
	startCount := 0
	for _, node := range wr.Nodes {
		if node.Type == NodeTypeStart {
			startCount++
		}
	}
	return startCount == 1
}

// hasEndNode checks if there's at least one end node
func (wr *WorkflowRequest) hasEndNode() bool {
	for _, node := range wr.Nodes {
		if node.Type == NodeTypeEnd {
			return true
		}
	}
	return false
}

// hasValidPath performs a BFS to check if there's a path from start to all end nodes
func (wr *WorkflowRequest) hasValidPath() bool {
	// Build adjacency list from edges
	graph := make(map[string][]string)
	for _, edge := range wr.Edges {
		graph[edge.Source] = append(graph[edge.Source], edge.Target)
	}
	
	// Find start node ID
	var startNodeID string
	for _, node := range wr.Nodes {
		if node.Type == NodeTypeStart {
			startNodeID = node.ID
			break
		}
	}
	
	if startNodeID == "" {
		return false
	}
	
	// Get all end node IDs
	endNodeIDs := make(map[string]bool)
	for _, node := range wr.Nodes {
		if node.Type == NodeTypeEnd {
			endNodeIDs[node.ID] = true
		}
	}
	
	if len(endNodeIDs) == 0 {
		return false
	}
	
	// BFS to find all reachable nodes from start
	visited := make(map[string]bool)
	queue := []string{startNodeID}
	visited[startNodeID] = true
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		// Add unvisited neighbors to queue
		for _, neighbor := range graph[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	
	// Check if all end nodes are reachable
	for endNodeID := range endNodeIDs {
		if !visited[endNodeID] {
			return false
		}
	}
	
	return true
}