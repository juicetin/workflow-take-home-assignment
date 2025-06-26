-- Create workflows table
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    starting_node_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key constraint to nodes table
    CONSTRAINT fk_workflows_starting_node FOREIGN KEY (starting_node_id) REFERENCES nodes(id) ON DELETE RESTRICT
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_workflows_name ON workflows(name);
CREATE INDEX IF NOT EXISTS idx_workflows_starting_node ON workflows(starting_node_id);
CREATE INDEX IF NOT EXISTS idx_workflows_created_at ON workflows(created_at);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_workflows_updated_at
    BEFORE UPDATE ON workflows
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add foreign key constraints from nodes and edges to workflows
ALTER TABLE nodes 
ADD CONSTRAINT fk_nodes_workflow FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE;

ALTER TABLE edges
ADD CONSTRAINT fk_edges_workflow FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE;