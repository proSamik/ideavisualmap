-- Create mind_maps table
CREATE TABLE mind_maps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_public BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active',
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create nodes table
CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mind_map_id UUID NOT NULL,
    parent_id UUID,
    content TEXT NOT NULL,
    position_x FLOAT NOT NULL,
    position_y FLOAT NOT NULL,
    node_type VARCHAR(50) DEFAULT 'default',
    style_data JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT fk_mind_map FOREIGN KEY (mind_map_id) REFERENCES mind_maps(id) ON DELETE CASCADE,
    CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES nodes(id) ON DELETE CASCADE
);

-- Create edges table
CREATE TABLE edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mind_map_id UUID NOT NULL,
    source_id UUID NOT NULL,
    target_id UUID NOT NULL,
    edge_type VARCHAR(50) DEFAULT 'default',
    style_data JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_mind_map FOREIGN KEY (mind_map_id) REFERENCES mind_maps(id) ON DELETE CASCADE,
    CONSTRAINT fk_source FOREIGN KEY (source_id) REFERENCES nodes(id) ON DELETE CASCADE,
    CONSTRAINT fk_target FOREIGN KEY (target_id) REFERENCES nodes(id) ON DELETE CASCADE,
    CONSTRAINT unique_connection UNIQUE(source_id, target_id)
);

-- Create indexes for performance
CREATE INDEX idx_mind_maps_user_id ON mind_maps(user_id);
CREATE INDEX idx_nodes_mind_map_id ON nodes(mind_map_id);
CREATE INDEX idx_nodes_parent_id ON nodes(parent_id);
CREATE INDEX idx_edges_mind_map_id ON edges(mind_map_id);
CREATE INDEX idx_edges_source_id ON edges(source_id);
CREATE INDEX idx_edges_target_id ON edges(target_id);
