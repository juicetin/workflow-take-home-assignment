-- Add workflow_id foreign key to nodes table
ALTER TABLE nodes 
ADD COLUMN workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE;

-- Add workflow_id foreign key to edges table  
ALTER TABLE edges
ADD COLUMN workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE;

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_nodes_workflow_id ON nodes(workflow_id);
CREATE INDEX IF NOT EXISTS idx_edges_workflow_id ON edges(workflow_id);