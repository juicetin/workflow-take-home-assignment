//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"github.com/google/uuid"
	"time"
)

type Edges struct {
	ID                   string `sql:"primary_key"`
	Source               string
	Target               string
	Type                 *string
	Animated             *bool
	StyleStroke          *string
	StyleStrokewidth     *float64
	Label                *string
	LabelstyleFill       *string
	LabelstyleFontweight *string
	SourceHandle         *string
	TargetHandle         *string
	WorkflowID           uuid.UUID
	CreatedAt            *time.Time
	UpdatedAt            *time.Time
}
