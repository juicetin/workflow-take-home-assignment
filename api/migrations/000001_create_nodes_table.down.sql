-- Drop the trigger first
DROP TRIGGER IF EXISTS update_nodes_updated_at ON nodes;

-- Drop indexes
DROP INDEX IF EXISTS idx_nodes_workflow_id;
DROP INDEX IF EXISTS idx_nodes_data;
DROP INDEX IF EXISTS idx_nodes_type;

-- Drop the nodes table
DROP TABLE IF EXISTS nodes;