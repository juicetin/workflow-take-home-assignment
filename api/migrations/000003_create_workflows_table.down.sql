-- Drop foreign key constraints first
ALTER TABLE edges DROP CONSTRAINT IF EXISTS fk_edges_workflow;
ALTER TABLE nodes DROP CONSTRAINT IF EXISTS fk_nodes_workflow;

-- Drop the trigger first
DROP TRIGGER IF EXISTS update_workflows_updated_at ON workflows;

-- Drop indexes
DROP INDEX IF EXISTS idx_workflows_created_at;
DROP INDEX IF EXISTS idx_workflows_starting_node;
DROP INDEX IF EXISTS idx_workflows_name;

-- Drop the workflows table
DROP TABLE IF EXISTS workflows;

-- Drop the shared trigger function if no other tables use it
DROP FUNCTION IF EXISTS update_updated_at_column();