-- Drop the trigger first
DROP TRIGGER IF EXISTS update_edges_updated_at ON edges;

-- Drop indexes
DROP INDEX IF EXISTS idx_edges_workflow_id;
DROP INDEX IF EXISTS idx_edges_source_target;
DROP INDEX IF EXISTS idx_edges_type;
DROP INDEX IF EXISTS idx_edges_target;
DROP INDEX IF EXISTS idx_edges_source;

-- Drop the edges table
DROP TABLE IF EXISTS edges;