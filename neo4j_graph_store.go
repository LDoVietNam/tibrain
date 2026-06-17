// TiBrain Neo4j Graph Store
// Provides graph-based knowledge retrieval using Neo4j
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// GraphNode represents a node in the knowledge graph
type GraphNode struct {
	ID         string                 `json:"id"`
	Labels     []string               `json:"labels"`
	Properties map[string]interface{} `json:"properties"`
}

// GraphRelationship represents a relationship in the knowledge graph
type GraphRelationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	StartNode  string                 `json:"start_node"`
	EndNode    string                 `json:"end_node"`
	Properties map[string]interface{} `json:"properties"`
}

// GraphSearchResult represents results from graph queries
type GraphSearchResult struct {
	Nodes         []GraphNode         `json:"nodes"`
	Relationships []GraphRelationship `json:"relationships"`
	Score         float64             `json:"score"`
	TraversalPath []string            `json:"traversal_path"`
}

// Neo4jGraphStore manages graph-based knowledge in Neo4j
type Neo4jGraphStore struct {
	driver   neo4j.DriverWithContext
	database string
	maxDepth int
}

// Neo4jConfig holds Neo4j connection configuration
type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
	MaxDepth int
}

// NewNeo4jGraphStore creates a new Neo4j graph store
func NewNeo4jGraphStore(config Neo4jConfig) (*Neo4jGraphStore, error) {
	driver, err := neo4j.NewDriverWithContext(config.URI, neo4j.BasicAuth(config.Username, config.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("create neo4j driver: %w", err)
	}

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("verify neo4j connectivity: %w", err)
	}

	maxDepth := config.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}

	return &Neo4jGraphStore{
		driver:   driver,
		database: config.Database,
		maxDepth: maxDepth,
	}, nil
}

// Close closes the Neo4j driver connection
func (g *Neo4jGraphStore) Close(ctx context.Context) error {
	if g.driver != nil {
		return g.driver.Close(ctx)
	}
	return nil
}

// AddNode creates or updates a node in the graph
func (g *Neo4jGraphStore) AddNode(ctx context.Context, node GraphNode) (string, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.database})
	defer session.Close(ctx)

	labels := strings.Join(node.Labels, ":")
	if labels == "" {
		labels = "Entity"
	}

	cypher := fmt.Sprintf(`
		MERGE (n:%s {id: $id})
		SET n += $props
		RETURN n.id
	`, labels)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"id":    node.ID,
			"props": node.Properties,
		}
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	if err != nil {
		return "", fmt.Errorf("add node: %w", err)
	}

	return node.ID, nil
}

// AddRelationship creates a relationship between two nodes
func (g *Neo4jGraphStore) AddRelationship(ctx context.Context, rel GraphRelationship) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.database})
	defer session.Close(ctx)

	cypher := fmt.Sprintf(`
		MATCH (a {id: $startId})
		MATCH (b {id: $endId})
		MERGE (a)-[r:%s]->(b)
		SET r += $props
		RETURN r
	`, rel.Type)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{
			"startId": rel.StartNode,
			"endId":   rel.EndNode,
			"props":   rel.Properties,
		}
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	if err != nil {
		return fmt.Errorf("add relationship: %w", err)
	}

	return nil
}

// SearchGraph performs a graph traversal search starting from matching nodes
func (g *Neo4jGraphStore) SearchGraph(ctx context.Context, query string, maxDepth int) ([]GraphSearchResult, error) {
	if maxDepth <= 0 || maxDepth > g.maxDepth {
		maxDepth = g.maxDepth
	}

	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.database})
	defer session.Close(ctx)

	cypher := fmt.Sprintf(`
		MATCH (n)
		WHERE toLower(n.name) CONTAINS toLower($query) 
		   OR toLower(n.content) CONTAINS toLower($query)
		   OR toLower(n.id) CONTAINS toLower($query)
		CALL apoc.path.subgraphAll(n, {maxLevel: %d})
		YIELD nodes, relationships
		RETURN nodes, relationships
		LIMIT 10
	`, maxDepth)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{"query": query}
		return tx.Run(ctx, cypher, params)
	})

	if err != nil {
		return nil, fmt.Errorf("search graph: %w", err)
	}

	return g.parseGraphResults(result)
}

// FindRelatedEntities finds entities related to a given entity through relationships
func (g *Neo4jGraphStore) FindRelatedEntities(ctx context.Context, entityID string, relationshipTypes []string) ([]GraphSearchResult, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.database})
	defer session.Close(ctx)

	var relPattern string
	if len(relationshipTypes) > 0 {
		patterns := make([]string, len(relationshipTypes))
		for i, relType := range relationshipTypes {
			patterns[i] = fmt.Sprintf(":%s", relType)
		}
		relPattern = strings.Join(patterns, "|")
	} else {
		relPattern = "*"
	}

	cypher := fmt.Sprintf(`
		MATCH (n {id: $entityId})-[r%s]-(related)
		RETURN n, r, related
		LIMIT 50
	`, relPattern)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		params := map[string]interface{}{"entityId": entityID}
		return tx.Run(ctx, cypher, params)
	})

	if err != nil {
		return nil, fmt.Errorf("find related entities: %w", err)
	}

	return g.parseGraphResults(result)
}

// QueryCypher executes a raw Cypher query and returns results
func (g *Neo4jGraphStore) QueryCypher(ctx context.Context, cypher string, params map[string]interface{}) ([]GraphSearchResult, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.database})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, cypher, params)
	})

	if err != nil {
		return nil, fmt.Errorf("query cypher: %w", err)
	}

	return g.parseGraphResults(result)
}

// BuildKnowledgeGraph creates a knowledge graph from RAG documents
func (g *Neo4jGraphStore) BuildKnowledgeGraph(ctx context.Context, documents []RAGDocument) error {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: g.database})
	defer session.Close(ctx)

	for _, doc := range documents {
		// Create document node
		docNode := GraphNode{
			ID: fmt.Sprintf("doc_%s", doc.ID),
			Labels: []string{"Document"},
			Properties: map[string]interface{}{
				"title":    doc.Title,
				"content":  doc.Content,
				"category": doc.Category,
				"tags":     doc.Tags,
				"path":     doc.Path,
			},
		}
		if _, err := g.AddNode(ctx, docNode); err != nil {
			logger.Warn("Failed to add document node %s: %v", doc.ID, err)
			continue
		}

		// Create entities from tags
		if doc.Tags != "" {
			tags := strings.Split(doc.Tags, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag == "" {
					continue
				}
				entityNode := GraphNode{
					ID:       fmt.Sprintf("entity_%s", tag),
					Labels:   []string{"Entity", "Tag"},
					Properties: map[string]interface{}{
						"name": tag,
					},
				}
				g.AddNode(ctx, entityNode)

				// Create relationship between document and entity
				rel := GraphRelationship{
					Type:      "TAGGED_WITH",
					StartNode: fmt.Sprintf("doc_%s", doc.ID),
					EndNode:   fmt.Sprintf("entity_%s", tag),
				}
				g.AddRelationship(ctx, rel)
			}
		}

		// Create category relationship
		if doc.Category != "" {
			categoryNode := GraphNode{
				ID: fmt.Sprintf("category_%s", doc.Category),
				Labels: []string{"Category"},
				Properties: map[string]interface{}{
					"name": doc.Category,
				},
			}
			g.AddNode(ctx, categoryNode)

			rel := GraphRelationship{
				Type:      "BELONGS_TO",
				StartNode: fmt.Sprintf("doc_%s", doc.ID),
				EndNode:   fmt.Sprintf("category_%s", doc.Category),
			}
			g.AddRelationship(ctx, rel)
		}
	}

	return nil
}

// parseGraphResults converts Neo4j results to GraphSearchResult
func (g *Neo4jGraphStore) parseGraphResults(result interface{}) ([]GraphSearchResult, error) {
	results := []GraphSearchResult{}

	if rw, ok := result.(neo4j.ResultWithContext); ok {
		for rw.Next(context.Background()) {
			record := rw.Record()

			searchResult := GraphSearchResult{}

			// Extract nodes
			if nodes, ok := record.Get("nodes"); ok {
				if nodeList, ok := nodes.([]interface{}); ok {
					for _, n := range nodeList {
						if node, ok := n.(neo4j.Node); ok {
							props := make(map[string]interface{})
							for k, v := range node.Props {
								props[k] = v
							}
							searchResult.Nodes = append(searchResult.Nodes, GraphNode{
								ID:         node.ElementId,
								Labels:     node.Labels,
								Properties: props,
							})
						}
					}
				}
			}

			// Extract relationships
			if rels, ok := record.Get("relationships"); ok {
				if relList, ok := rels.([]interface{}); ok {
					for _, r := range relList {
						if rel, ok := r.(neo4j.Relationship); ok {
							props := make(map[string]interface{})
							for k, v := range rel.Props {
								props[k] = v
							}
							searchResult.Relationships = append(searchResult.Relationships, GraphRelationship{
								ID:         rel.ElementId,
								Type:       rel.Type,
								StartNode:  rel.StartElementId,
								EndNode:    rel.EndElementId,
								Properties: props,
							})
						}
					}
				}
			}

			results = append(results, searchResult)
		}
	}

	return results, nil
}

// RAGGraphIntegration provides graph-based RAG capabilities
type RAGGraphIntegration struct {
	graphStore *Neo4jGraphStore
	ragManager *RAGSystemManager
}

// NewRAGGraphIntegration creates a graph-enhanced RAG system
func NewRAGGraphIntegration(graphStore *Neo4jGraphStore, ragManager *RAGSystemManager) *RAGGraphIntegration {
	return &RAGGraphIntegration{
		graphStore: graphStore,
		ragManager: ragManager,
	}
}

// GraphEnhancedQuery performs RAG with graph context
func (gi *RAGGraphIntegration) GraphEnhancedQuery(ctx context.Context, query string) (*GraphSearchResult, error) {
	result := &GraphSearchResult{}

	// First, try graph search for entity relationships
	if gi.graphStore != nil {
		graphResults, err := gi.graphStore.SearchGraph(ctx, query, 2)
		if err != nil {
			logger.Warn("Graph search failed, falling back to regular RAG: %v", err)
		} else if len(graphResults) > 0 {
			result.Nodes = graphResults[0].Nodes
			result.Relationships = graphResults[0].Relationships
		}
	}

	// Also get traditional RAG results
	ragQuery := RAGQuery{
		Query:     query,
		Metadata:  make(map[string]string),
		Timestamp: time.Now(),
	}

	ragResult, err := gi.ragManager.ProcessQuery(ragQuery)
	if err != nil {
		return nil, err
	}

	result.Score = ragResult.Confidence
	if result.Score == 0 {
		result.Score = 0.5
	}

	return result, nil
}