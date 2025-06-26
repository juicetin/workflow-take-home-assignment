package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workflow-code-test/api/internal/models"
	"workflow-code-test/api/pkg/db"
)

type WorkflowRepository struct {
	conn *pgx.Conn
}

func NewWorkflowRepository(conn *pgx.Conn) *WorkflowRepository {
	return &WorkflowRepository{
		conn: conn,
	}
}

// GetWorkflow retrieves a workflow by ID
func (r *WorkflowRepository) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.Workflow, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`
	
	var workflow models.Workflow
	err := r.conn.QueryRow(ctx, query, workflowID).Scan(
		&workflow.ID,
		&workflow.Name,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("workflow not found: %s", workflowID)
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	
	return &workflow, nil
}

// GetNodesByWorkflow retrieves all nodes for a given workflow
func (r *WorkflowRepository) GetNodesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Node, error) {
	query := `
		SELECT id, type, position_x, position_y, data, workflow_id, created_at, updated_at
		FROM nodes
		WHERE workflow_id = $1
		ORDER BY created_at
	`
	
	rows, err := r.conn.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()
	
	var nodes []models.Node
	for rows.Next() {
		var node models.Node
		err := rows.Scan(
			&node.ID,
			&node.Type,
			&node.PositionX,
			&node.PositionY,
			&node.Data,
			&node.WorkflowID,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, node)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating nodes: %w", err)
	}
	
	return nodes, nil
}

// GetEdgesByWorkflow retrieves all edges for a given workflow
func (r *WorkflowRepository) GetEdgesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Edge, error) {
	query := `
		SELECT id, source, target, type, animated, style_stroke, style_strokewidth,
		       label, labelstyle_fill, labelstyle_fontweight, source_handle, target_handle,
		       workflow_id, created_at, updated_at
		FROM edges
		WHERE workflow_id = $1
		ORDER BY created_at
	`
	
	rows, err := r.conn.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}
	defer rows.Close()
	
	var edges []models.Edge
	for rows.Next() {
		var edge models.Edge
		err := rows.Scan(
			&edge.ID,
			&edge.Source,
			&edge.Target,
			&edge.Type,
			&edge.Animated,
			&edge.StyleStroke,
			&edge.StyleStrokeWidth,
			&edge.Label,
			&edge.LabelStyleFill,
			&edge.LabelStyleFontWeight,
			&edge.SourceHandle,
			&edge.TargetHandle,
			&edge.WorkflowID,
			&edge.CreatedAt,
			&edge.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}
		edges = append(edges, edge)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating edges: %w", err)
	}
	
	return edges, nil
}

// SaveWorkflow creates or updates a workflow and its associated nodes and edges
func (r *WorkflowRepository) SaveWorkflow(ctx context.Context, workflow *models.Workflow, nodes []models.Node, edges []models.Edge) error {
	return db.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Insert or update workflow
		workflowQuery := `
			INSERT INTO workflows (id, name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				updated_at = NOW()
		`
		
		_, err := tx.Exec(ctx, workflowQuery, workflow.ID, workflow.Name)
		if err != nil {
			return fmt.Errorf("failed to save workflow: %w", err)
		}
		
		// Delete existing nodes and edges for this workflow
		_, err = tx.Exec(ctx, "DELETE FROM edges WHERE workflow_id = $1", workflow.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing edges: %w", err)
		}
		
		_, err = tx.Exec(ctx, "DELETE FROM nodes WHERE workflow_id = $1", workflow.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing nodes: %w", err)
		}
		
		// Insert nodes
		for _, node := range nodes {
			nodeQuery := `
				INSERT INTO nodes (id, type, position_x, position_y, data, workflow_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			`
			_, err = tx.Exec(ctx, nodeQuery, node.ID, node.Type, node.PositionX, node.PositionY, node.Data, workflow.ID)
			if err != nil {
				return fmt.Errorf("failed to insert node %s: %w", node.ID, err)
			}
		}
		
		// Insert edges
		for _, edge := range edges {
			edgeQuery := `
				INSERT INTO edges (id, source, target, type, animated, style_stroke, style_strokewidth,
				                  label, labelstyle_fill, labelstyle_fontweight, source_handle, target_handle,
				                  workflow_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
			`
			_, err = tx.Exec(ctx, edgeQuery,
				edge.ID, edge.Source, edge.Target, edge.Type, edge.Animated,
				edge.StyleStroke, edge.StyleStrokeWidth, edge.Label,
				edge.LabelStyleFill, edge.LabelStyleFontWeight,
				edge.SourceHandle, edge.TargetHandle, workflow.ID)
			if err != nil {
				return fmt.Errorf("failed to insert edge %s: %w", edge.ID, err)
			}
		}
		
		return nil
	})
}