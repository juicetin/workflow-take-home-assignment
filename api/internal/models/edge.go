package models

import (
	"time"

	"github.com/google/uuid"
)

// Edge represents a workflow edge connecting two nodes
type Edge struct {
	ID                   string    `json:"id" db:"id"`
	Source               string    `json:"source" db:"source"`
	Target               string    `json:"target" db:"target"`
	Type                 *string   `json:"type,omitempty" db:"type"`
	Animated             bool      `json:"animated" db:"animated"`
	StyleStroke          *string   `json:"-" db:"style_stroke"`
	StyleStrokeWidth     *float64  `json:"-" db:"style_strokewidth"`
	Label                *string   `json:"label,omitempty" db:"label"`
	LabelStyleFill       *string   `json:"-" db:"labelstyle_fill"`
	LabelStyleFontWeight *string   `json:"-" db:"labelstyle_fontweight"`
	SourceHandle         *string   `json:"sourceHandle,omitempty" db:"source_handle"`
	TargetHandle         *string   `json:"targetHandle,omitempty" db:"target_handle"`
	WorkflowID           uuid.UUID `json:"-" db:"workflow_id"`
	CreatedAt            time.Time `json:"-" db:"created_at"`
	UpdatedAt            time.Time `json:"-" db:"updated_at"`
}

// Style represents the style properties of an edge
type Style struct {
	Stroke      string  `json:"stroke"`
	StrokeWidth float64 `json:"strokeWidth"`
}

// LabelStyle represents the label style properties of an edge
type LabelStyle struct {
	Fill       string `json:"fill"`
	FontWeight string `json:"fontWeight"`
}

// EdgeResponse represents an edge as returned to the frontend
type EdgeResponse struct {
	ID           string      `json:"id"`
	Source       string      `json:"source"`
	Target       string      `json:"target"`
	Type         *string     `json:"type,omitempty"`
	Animated     bool        `json:"animated"`
	Style        *Style      `json:"style,omitempty"`
	Label        *string     `json:"label,omitempty"`
	LabelStyle   *LabelStyle `json:"labelStyle,omitempty"`
	SourceHandle *string     `json:"sourceHandle,omitempty"`
	TargetHandle *string     `json:"targetHandle,omitempty"`
}

// ToResponse converts an Edge to EdgeResponse format for API responses
func (e *Edge) ToResponse() EdgeResponse {
	response := EdgeResponse{
		ID:           e.ID,
		Source:       e.Source,
		Target:       e.Target,
		Type:         e.Type,
		Animated:     e.Animated,
		Label:        e.Label,
		SourceHandle: e.SourceHandle,
		TargetHandle: e.TargetHandle,
	}

	// Convert style fields
	if e.StyleStroke != nil && e.StyleStrokeWidth != nil {
		response.Style = &Style{
			Stroke:      *e.StyleStroke,
			StrokeWidth: *e.StyleStrokeWidth,
		}
	}

	// Convert label style fields
	if e.LabelStyleFill != nil && e.LabelStyleFontWeight != nil {
		response.LabelStyle = &LabelStyle{
			Fill:       *e.LabelStyleFill,
			FontWeight: *e.LabelStyleFontWeight,
		}
	}

	return response
}

// EdgeRequest represents an edge as sent from the frontend
type EdgeRequest struct {
	ID           string      `json:"id"`
	Source       string      `json:"source"`
	Target       string      `json:"target"`
	Type         *string     `json:"type,omitempty"`
	Animated     bool        `json:"animated"`
	Style        *Style      `json:"style,omitempty"`
	Label        *string     `json:"label,omitempty"`
	LabelStyle   *LabelStyle `json:"labelStyle,omitempty"`
	SourceHandle *string     `json:"sourceHandle,omitempty"`
	TargetHandle *string     `json:"targetHandle,omitempty"`
}

// ToEdge converts an EdgeRequest to an Edge for database storage
func (er *EdgeRequest) ToEdge() *Edge {
	edge := &Edge{
		ID:           er.ID,
		Source:       er.Source,
		Target:       er.Target,
		Type:         er.Type,
		Animated:     er.Animated,
		Label:        er.Label,
		SourceHandle: er.SourceHandle,
		TargetHandle: er.TargetHandle,
	}

	// Convert style fields
	if er.Style != nil {
		edge.StyleStroke = &er.Style.Stroke
		edge.StyleStrokeWidth = &er.Style.StrokeWidth
	}

	// Convert label style fields
	if er.LabelStyle != nil {
		edge.LabelStyleFill = &er.LabelStyle.Fill
		edge.LabelStyleFontWeight = &er.LabelStyle.FontWeight
	}

	return edge
}