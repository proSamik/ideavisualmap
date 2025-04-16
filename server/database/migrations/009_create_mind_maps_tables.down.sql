-- Drop indexes
DROP INDEX IF EXISTS idx_edges_target_id;
DROP INDEX IF EXISTS idx_edges_source_id;
DROP INDEX IF EXISTS idx_edges_mind_map_id;
DROP INDEX IF EXISTS idx_nodes_parent_id;
DROP INDEX IF EXISTS idx_nodes_mind_map_id;
DROP INDEX IF EXISTS idx_mind_maps_user_id;

-- Drop tables (in reverse order of creation to handle dependencies)
DROP TABLE IF EXISTS edges;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS mind_maps;
