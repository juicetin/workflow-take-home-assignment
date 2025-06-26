-- Create edges table
CREATE TABLE IF NOT EXISTS edges (
    id VARCHAR(255) PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    target VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    animated BOOLEAN DEFAULT FALSE,
    style_stroke VARCHAR(7), -- hex color code like #10b981
    style_strokewidth DECIMAL(5,2),
    label VARCHAR(255),
    labelstyle_fill VARCHAR(7), -- hex color code like #10b981
    labelstyle_fontweight VARCHAR(20), -- like 'bold', 'normal'
    source_handle VARCHAR(50), -- for conditional nodes like "true", "false"
    target_handle VARCHAR(50),
    workflow_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key constraints to nodes table
    CONSTRAINT fk_edges_source FOREIGN KEY (source) REFERENCES nodes(id) ON DELETE CASCADE,
    CONSTRAINT fk_edges_target FOREIGN KEY (target) REFERENCES nodes(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target);
CREATE INDEX IF NOT EXISTS idx_edges_type ON edges(type);
CREATE INDEX IF NOT EXISTS idx_edges_source_target ON edges(source, target);
CREATE INDEX IF NOT EXISTS idx_edges_workflow_id ON edges(workflow_id);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_edges_updated_at
    BEFORE UPDATE ON edges
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();