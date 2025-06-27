package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"workflow-code-test/api/internal/db/gen/workflow_engine/public/model"
	. "workflow-code-test/api/internal/db/gen/workflow_engine/public/table"
	"workflow-code-test/api/internal/models"
)

type WorkflowRepository struct {
	db *sql.DB
}

func NewWorkflowRepository(db *sql.DB) *WorkflowRepository {
	return &WorkflowRepository{
		db: db,
	}
}

// GetWorkflow retrieves a workflow by ID
func (r *WorkflowRepository) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.Workflow, error) {
	stmt := postgres.SELECT(
		Workflows.ID,
		Workflows.Name,
		Workflows.CreatedAt,
		Workflows.UpdatedAt,
	).FROM(
		Workflows,
	).WHERE(
		Workflows.ID.EQ(postgres.UUID(workflowID)),
	)

	var dest model.Workflows
	err := stmt.QueryContext(ctx, r.db, &dest)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow not found: %s", workflowID)
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Convert db model to domain model
	workflow := &models.Workflow{
		ID:   dest.ID,
		Name: dest.Name,
	}
	if dest.CreatedAt != nil {
		workflow.CreatedAt = *dest.CreatedAt
	}
	if dest.UpdatedAt != nil {
		workflow.UpdatedAt = *dest.UpdatedAt
	}

	return workflow, nil
}

// GetNodesByWorkflow retrieves all nodes for a given workflow
func (r *WorkflowRepository) GetNodesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Node, error) {
	stmt := postgres.SELECT(
		Nodes.ID,
		Nodes.Type,
		Nodes.PositionX,
		Nodes.PositionY,
		Nodes.Data,
		Nodes.WorkflowID,
		Nodes.CreatedAt,
		Nodes.UpdatedAt,
	).FROM(
		Nodes,
	).WHERE(
		Nodes.WorkflowID.EQ(postgres.UUID(workflowID)),
	).ORDER_BY(
		Nodes.CreatedAt.ASC(),
	)

	var dbNodes []model.Nodes
	err := stmt.QueryContext(ctx, r.db, &dbNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}

	// Convert db models to domain models
	nodes := make([]models.Node, len(dbNodes))
	for i, dbNode := range dbNodes {
		node := models.Node{
			ID:         dbNode.ID,
			Type:       dbNode.Type,
			PositionX:  dbNode.PositionX,
			PositionY:  dbNode.PositionY,
			RawData:    []byte(dbNode.Data), // Jet sees JSONB as string, convert to []byte
			WorkflowID: dbNode.WorkflowID,
		}
		if dbNode.CreatedAt != nil {
			node.CreatedAt = *dbNode.CreatedAt
		}
		if dbNode.UpdatedAt != nil {
			node.UpdatedAt = *dbNode.UpdatedAt
		}

		// Load strongly typed data from raw data
		if len(node.RawData) > 0 {
			if err := node.LoadDataFromRaw(); err != nil {
				return nil, fmt.Errorf("failed to load typed data for node %s: %w", node.ID, err)
			}
		}

		nodes[i] = node
	}

	return nodes, nil
}

// GetEdgesByWorkflow retrieves all edges for a given workflow
func (r *WorkflowRepository) GetEdgesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Edge, error) {
	stmt := postgres.SELECT(
		Edges.ID,
		Edges.Source,
		Edges.Target,
		Edges.Type,
		Edges.Animated,
		Edges.StyleStroke,
		Edges.StyleStrokewidth,
		Edges.Label,
		Edges.LabelstyleFill,
		Edges.LabelstyleFontweight,
		Edges.SourceHandle,
		Edges.TargetHandle,
		Edges.WorkflowID,
		Edges.CreatedAt,
		Edges.UpdatedAt,
	).FROM(
		Edges,
	).WHERE(
		Edges.WorkflowID.EQ(postgres.UUID(workflowID)),
	).ORDER_BY(
		Edges.CreatedAt.ASC(),
	)

	var dbEdges []model.Edges
	err := stmt.QueryContext(ctx, r.db, &dbEdges)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}

	// Convert db models to domain models
	edges := make([]models.Edge, len(dbEdges))
	for i, dbEdge := range dbEdges {
		edge := models.Edge{
			ID:         dbEdge.ID,
			Source:     dbEdge.Source,
			Target:     dbEdge.Target,
			WorkflowID: dbEdge.WorkflowID,
		}

		// Handle optional fields (nullable in database)
		if dbEdge.Type != nil {
			edge.Type = dbEdge.Type
		}
		if dbEdge.Animated != nil {
			edge.Animated = *dbEdge.Animated
		}
		if dbEdge.StyleStroke != nil {
			edge.StyleStroke = dbEdge.StyleStroke
		}
		if dbEdge.StyleStrokewidth != nil {
			edge.StyleStrokeWidth = dbEdge.StyleStrokewidth
		}
		if dbEdge.Label != nil {
			edge.Label = dbEdge.Label
		}
		if dbEdge.LabelstyleFill != nil {
			edge.LabelStyleFill = dbEdge.LabelstyleFill
		}
		if dbEdge.LabelstyleFontweight != nil {
			edge.LabelStyleFontWeight = dbEdge.LabelstyleFontweight
		}
		if dbEdge.SourceHandle != nil {
			edge.SourceHandle = dbEdge.SourceHandle
		}
		if dbEdge.TargetHandle != nil {
			edge.TargetHandle = dbEdge.TargetHandle
		}
		if dbEdge.CreatedAt != nil {
			edge.CreatedAt = *dbEdge.CreatedAt
		}
		if dbEdge.UpdatedAt != nil {
			edge.UpdatedAt = *dbEdge.UpdatedAt
		}

		edges[i] = edge
	}

	return edges, nil
}

// SaveWorkflow creates or updates a workflow and its associated nodes and edges
func (r *WorkflowRepository) SaveWorkflow(ctx context.Context, workflow *models.Workflow, nodes []models.Node, edges []models.Edge) error {
	// Start a database transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or update workflow using UPSERT
	workflowStmt := Workflows.INSERT(
		Workflows.ID,
		Workflows.Name,
		Workflows.CreatedAt,
		Workflows.UpdatedAt,
	).VALUES(
		workflow.ID,
		workflow.Name,
		postgres.NOW(),
		postgres.NOW(),
	).ON_CONFLICT(Workflows.ID).DO_UPDATE(
		postgres.SET(
			Workflows.Name.SET(postgres.String(workflow.Name)),
			Workflows.UpdatedAt.SET(postgres.NOW()),
		),
	)

	_, err = workflowStmt.ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	deleteEdgesStmt := Edges.DELETE().WHERE(
		Edges.WorkflowID.EQ(postgres.UUID(workflow.ID)),
	)
	_, err = deleteEdgesStmt.ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to delete existing edges: %w", err)
	}

	deleteNodesStmt := Nodes.DELETE().WHERE(
		Nodes.WorkflowID.EQ(postgres.UUID(workflow.ID)),
	)
	_, err = deleteNodesStmt.ExecContext(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to delete existing nodes: %w", err)
	}

	if len(nodes) > 0 {
		// Start with the base INSERT statement
		insertNodesStmt := Nodes.INSERT(
			Nodes.ID,
			Nodes.Type,
			Nodes.PositionX,
			Nodes.PositionY,
			Nodes.Data,
			Nodes.WorkflowID,
			Nodes.CreatedAt,
			Nodes.UpdatedAt,
		)

		// Add VALUES for each node
		for _, node := range nodes {
			// Ensure RawData is up to date
			if err := node.UpdateRawDataFromData(); err != nil {
				return fmt.Errorf("failed to update raw data for node %s: %w", node.ID, err)
			}

			insertNodesStmt = insertNodesStmt.VALUES(
				node.ID,
				node.Type,
				node.PositionX,
				node.PositionY,
				string(node.RawData), // Convert []byte to string for JSONB
				workflow.ID,
				postgres.NOW(),
				postgres.NOW(),
			)
		}

		_, err = insertNodesStmt.ExecContext(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to insert nodes: %w", err)
		}
	}

	// Insert edges using batch insert
	if len(edges) > 0 {
		// Start with the base INSERT statement
		insertEdgesStmt := Edges.INSERT(
			Edges.ID,
			Edges.Source,
			Edges.Target,
			Edges.Type,
			Edges.Animated,
			Edges.StyleStroke,
			Edges.StyleStrokewidth,
			Edges.Label,
			Edges.LabelstyleFill,
			Edges.LabelstyleFontweight,
			Edges.SourceHandle,
			Edges.TargetHandle,
			Edges.WorkflowID,
			Edges.CreatedAt,
			Edges.UpdatedAt,
		)

		// Add VALUES for each edge
		for _, edge := range edges {
			insertEdgesStmt = insertEdgesStmt.VALUES(
				edge.ID,
				edge.Source,
				edge.Target,
				edge.Type,
				edge.Animated,
				edge.StyleStroke,
				edge.StyleStrokeWidth,
				edge.Label,
				edge.LabelStyleFill,
				edge.LabelStyleFontWeight,
				edge.SourceHandle,
				edge.TargetHandle,
				workflow.ID,
				postgres.NOW(),
				postgres.NOW(),
			)
		}

		_, err = insertEdgesStmt.ExecContext(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to insert edges: %w", err)
		}
	}

	return tx.Commit()
}
