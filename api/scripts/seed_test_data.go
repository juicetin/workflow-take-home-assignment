package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	"workflow-code-test/api/internal/models"
	"workflow-code-test/api/internal/repository"
	"workflow-code-test/api/pkg/db"
)

// SeedTestData creates test workflow data in the database
func SeedTestData(databaseURL string) error {
	// Create database config and sql.DB connection for Jet
	dbConfig := db.DefaultConfig()
	dbConfig.URI = databaseURL

	sqlDB, err := db.GetJetDB(dbConfig)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// Create repository using sql.DB
	repo := repository.NewWorkflowRepository(sqlDB)

	// Create test workflow
	workflowID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	workflow := &models.Workflow{
		ID:   workflowID,
		Name: "Weather Alert Workflow",
	}

	// Create test nodes - using RawData only for database storage
	nodes := []models.Node{
		{
			ID:         "start",
			Type:       models.NodeTypeStart,
			PositionX:  -160.0,
			PositionY:  300.0,
			RawData:    json.RawMessage(`{"label":"Start","description":"Begin weather check workflow","metadata":{"hasHandles":{"source":true,"target":false}}}`),
			WorkflowID: workflowID,
		},
		{
			ID:         "form",
			Type:       models.NodeTypeForm,
			PositionX:  152.0,
			PositionY:  304.0,
			RawData:    json.RawMessage(`{"label":"User Input","description":"Process collected data - name, email, location","metadata":{"hasHandles":{"source":true,"target":true},"inputFields":["name","email","city"],"outputVariables":["name","email","city"]}}`),
			WorkflowID: workflowID,
		},
		{
			ID:         "weather-api",
			Type:       models.NodeTypeIntegration,
			PositionX:  460.0,
			PositionY:  304.0,
			RawData:    json.RawMessage(`{"label":"Weather API","description":"Fetch current temperature for {{city}}","metadata":{"hasHandles":{"source":true,"target":true},"inputVariables":["city"],"apiEndpoint":"https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true","options":[{"city":"Sydney","lat":-33.8688,"lon":151.2093},{"city":"Melbourne","lat":-37.8136,"lon":144.9631},{"city":"Brisbane","lat":-27.4698,"lon":153.0251},{"city":"Perth","lat":-31.9505,"lon":115.8605},{"city":"Adelaide","lat":-34.9285,"lon":138.6007}],"outputVariables":["temperature"]}}`),
			WorkflowID: workflowID,
		},
		{
			ID:         "condition",
			Type:       models.NodeTypeCondition,
			PositionX:  794.0,
			PositionY:  304.0,
			RawData:    json.RawMessage(`{"label":"Check Condition","description":"Evaluate temperature threshold","metadata":{"hasHandles":{"source":["true","false"],"target":true},"conditionExpression":"temperature {{operator}} {{threshold}}","outputVariables":["conditionMet"]}}`),
			WorkflowID: workflowID,
		},
		{
			ID:         "email",
			Type:       models.NodeTypeEmail,
			PositionX:  1096.0,
			PositionY:  88.0,
			RawData:    json.RawMessage(`{"label":"Send Alert","description":"Email weather alert notification","metadata":{"hasHandles":{"source":true,"target":true},"inputVariables":["name","city","temperature"],"emailTemplate":{"subject":"Weather Alert","body":"Weather alert for {{city}}! Temperature is {{temperature}}°C!"},"outputVariables":["emailSent"]}}`),
			WorkflowID: workflowID,
		},
		{
			ID:         "end",
			Type:       models.NodeTypeEnd,
			PositionX:  1360.0,
			PositionY:  302.0,
			RawData:    json.RawMessage(`{"label":"Complete","description":"Workflow execution finished","metadata":{"hasHandles":{"source":false,"target":true}}}`),
			WorkflowID: workflowID,
		},
	}

	// Create test edges
	edges := []models.Edge{
		{
			ID:               "e1",
			Source:           "start",
			Target:           "form",
			Type:             stringPtr(models.EdgeTypeSmoothstep),
			Animated:         true,
			StyleStroke:      stringPtr("#10b981"),
			StyleStrokeWidth: float64Ptr(3),
			Label:            stringPtr("Initialize"),
			WorkflowID:       workflowID,
		},
		{
			ID:               "e2",
			Source:           "form",
			Target:           "weather-api",
			Type:             stringPtr(models.EdgeTypeSmoothstep),
			Animated:         true,
			StyleStroke:      stringPtr("#3b82f6"),
			StyleStrokeWidth: float64Ptr(3),
			Label:            stringPtr("Submit Data"),
			WorkflowID:       workflowID,
		},
		{
			ID:               "e3",
			Source:           "weather-api",
			Target:           "condition",
			Type:             stringPtr(models.EdgeTypeSmoothstep),
			Animated:         true,
			StyleStroke:      stringPtr("#f97316"),
			StyleStrokeWidth: float64Ptr(3),
			Label:            stringPtr("Temperature Data"),
			WorkflowID:       workflowID,
		},
		{
			ID:                   "e4",
			Source:               "condition",
			Target:               "email",
			Type:                 stringPtr(models.EdgeTypeSmoothstep),
			SourceHandle:         stringPtr("true"),
			Animated:             true,
			StyleStroke:          stringPtr("#10b981"),
			StyleStrokeWidth:     float64Ptr(3),
			Label:                stringPtr("✓ Condition Met"),
			LabelStyleFill:       stringPtr("#10b981"),
			LabelStyleFontWeight: stringPtr("bold"),
			WorkflowID:           workflowID,
		},
		{
			ID:                   "e5",
			Source:               "condition",
			Target:               "end",
			Type:                 stringPtr(models.EdgeTypeSmoothstep),
			SourceHandle:         stringPtr("false"),
			Animated:             true,
			StyleStroke:          stringPtr("#6b7280"),
			StyleStrokeWidth:     float64Ptr(3),
			Label:                stringPtr("✗ No Alert Needed"),
			LabelStyleFill:       stringPtr("#6b7280"),
			LabelStyleFontWeight: stringPtr("bold"),
			WorkflowID:           workflowID,
		},
		{
			ID:                   "e6",
			Source:               "email",
			Target:               "end",
			Type:                 stringPtr(models.EdgeTypeSmoothstep),
			Animated:             true,
			StyleStroke:          stringPtr("#ef4444"),
			StyleStrokeWidth:     float64Ptr(2),
			Label:                stringPtr("Alert Sent"),
			LabelStyleFill:       stringPtr("#ef4444"),
			LabelStyleFontWeight: stringPtr("bold"),
			WorkflowID:           workflowID,
		},
	}

	// Populate strongly typed data from raw data for all nodes
	for i := range nodes {
		if err := nodes[i].LoadDataFromRaw(); err != nil {
			return fmt.Errorf("failed to load data for node %s: %w", nodes[i].ID, err)
		}
	}

	// Save the workflow
	ctx := context.Background()
	if err := repo.SaveWorkflow(ctx, workflow, nodes, edges); err != nil {
		return err
	}

	log.Printf("Successfully seeded test workflow with ID: %s", workflowID)
	return nil
}

// Helper functions for pointer values
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
