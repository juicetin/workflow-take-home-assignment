//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Edges = newEdgesTable("public", "edges", "")

type edgesTable struct {
	postgres.Table

	// Columns
	ID                   postgres.ColumnString
	Source               postgres.ColumnString
	Target               postgres.ColumnString
	Type                 postgres.ColumnString
	Animated             postgres.ColumnBool
	StyleStroke          postgres.ColumnString
	StyleStrokewidth     postgres.ColumnFloat
	Label                postgres.ColumnString
	LabelstyleFill       postgres.ColumnString
	LabelstyleFontweight postgres.ColumnString
	SourceHandle         postgres.ColumnString
	TargetHandle         postgres.ColumnString
	WorkflowID           postgres.ColumnString
	CreatedAt            postgres.ColumnTimestampz
	UpdatedAt            postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
	DefaultColumns postgres.ColumnList
}

type EdgesTable struct {
	edgesTable

	EXCLUDED edgesTable
}

// AS creates new EdgesTable with assigned alias
func (a EdgesTable) AS(alias string) *EdgesTable {
	return newEdgesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new EdgesTable with assigned schema name
func (a EdgesTable) FromSchema(schemaName string) *EdgesTable {
	return newEdgesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new EdgesTable with assigned table prefix
func (a EdgesTable) WithPrefix(prefix string) *EdgesTable {
	return newEdgesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new EdgesTable with assigned table suffix
func (a EdgesTable) WithSuffix(suffix string) *EdgesTable {
	return newEdgesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newEdgesTable(schemaName, tableName, alias string) *EdgesTable {
	return &EdgesTable{
		edgesTable: newEdgesTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newEdgesTableImpl("", "excluded", ""),
	}
}

func newEdgesTableImpl(schemaName, tableName, alias string) edgesTable {
	var (
		IDColumn                   = postgres.StringColumn("id")
		SourceColumn               = postgres.StringColumn("source")
		TargetColumn               = postgres.StringColumn("target")
		TypeColumn                 = postgres.StringColumn("type")
		AnimatedColumn             = postgres.BoolColumn("animated")
		StyleStrokeColumn          = postgres.StringColumn("style_stroke")
		StyleStrokewidthColumn     = postgres.FloatColumn("style_strokewidth")
		LabelColumn                = postgres.StringColumn("label")
		LabelstyleFillColumn       = postgres.StringColumn("labelstyle_fill")
		LabelstyleFontweightColumn = postgres.StringColumn("labelstyle_fontweight")
		SourceHandleColumn         = postgres.StringColumn("source_handle")
		TargetHandleColumn         = postgres.StringColumn("target_handle")
		WorkflowIDColumn           = postgres.StringColumn("workflow_id")
		CreatedAtColumn            = postgres.TimestampzColumn("created_at")
		UpdatedAtColumn            = postgres.TimestampzColumn("updated_at")
		allColumns                 = postgres.ColumnList{IDColumn, SourceColumn, TargetColumn, TypeColumn, AnimatedColumn, StyleStrokeColumn, StyleStrokewidthColumn, LabelColumn, LabelstyleFillColumn, LabelstyleFontweightColumn, SourceHandleColumn, TargetHandleColumn, WorkflowIDColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns             = postgres.ColumnList{SourceColumn, TargetColumn, TypeColumn, AnimatedColumn, StyleStrokeColumn, StyleStrokewidthColumn, LabelColumn, LabelstyleFillColumn, LabelstyleFontweightColumn, SourceHandleColumn, TargetHandleColumn, WorkflowIDColumn, CreatedAtColumn, UpdatedAtColumn}
		defaultColumns             = postgres.ColumnList{AnimatedColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return edgesTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:                   IDColumn,
		Source:               SourceColumn,
		Target:               TargetColumn,
		Type:                 TypeColumn,
		Animated:             AnimatedColumn,
		StyleStroke:          StyleStrokeColumn,
		StyleStrokewidth:     StyleStrokewidthColumn,
		Label:                LabelColumn,
		LabelstyleFill:       LabelstyleFillColumn,
		LabelstyleFontweight: LabelstyleFontweightColumn,
		SourceHandle:         SourceHandleColumn,
		TargetHandle:         TargetHandleColumn,
		WorkflowID:           WorkflowIDColumn,
		CreatedAt:            CreatedAtColumn,
		UpdatedAt:            UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
		DefaultColumns: defaultColumns,
	}
}
