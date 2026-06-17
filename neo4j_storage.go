package main

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jStorage provides Neo4j graph database integration for RAG
type Neo4jStorage struct {
	driver neo4j.DriverWithContext
	ctx    context.Context
}

// GraphEntity represents a node in the knowledge graph
type GraphEntity struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties"`
}

// GraphRelation represents an edge between entities
type GraphRelation struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Properties map[string]interface{} `json:"properties"`
}

// NewNeo4jStorage creates a new Neo4j storage connection
func NewNeo4jStorage(uri, username, password string) (*Neo4jStorage, error) {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	return &Neo4jStorage{
		driver: driver,
		ctx:    ctx,
	}, nil
}

// Close closes the Neo4j driver connection
func (s *Neo4jStorage) Close() error {
	return s.driver.Close(s.ctx)
}

// CreateEntity creates a new entity node in the graph
func (s *Neo4jStorage) CreateEntity(entity GraphEntity) error {
	session := s.driver.NewSession(s.ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
	})
	defer session.Close(s.ctx)

	query := `
		MERGE (e:Entity {id: $id})
		SET e.type = $type, e.name = $name, e.description = $description
		WITH e
		UNWIND keys($properties) AS key
		SET e[key] = $properties[key]
		RETURN e
	`

	props := map[string]interface{}{
		"id":          entity.ID,
		"type":        entity.Type,
		"name":        entity.Name,
		"description": entity.Description,
		"properties":  entity.Properties,
	}

	_, err := session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(s.ctx, query, props)
		if err != nil {
			return nil, err
		}
		_, err = result.Single(s.ctx)
		return nil, err
	})

	return err
}

// CreateRelation creates a relationship between two entities
func (s *Neo4jStorage) CreateRelation(relation GraphRelation) error {
	session := s.driver.NewSession(s.ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
	})
	defer session.Close(s.ctx)

	query := `
		MATCH (from:Entity {id: $fromId})
		MATCH (to:Entity {id: $toId})
		MERGE (from)-[r:RELATES {type: $type}]->(to)
		SET r.id = $id
		WITH r
		UNWIND keys($properties) AS key
		SET r[key] = $properties[key]
		RETURN r
	`

	props := map[string]interface{}{
		"fromId":     relation.From,
		"toId":       relation.To,
		"type":       relation.Type,
		"id":         relation.ID,
		"properties": relation.Properties,
	}

	_, err := session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(s.ctx, query, props)
		if err != nil {
			return nil, err
		}
		_, err = result.Single(s.ctx)
		return nil, err
	})

	return err
}

// QueryRelatedEntities finds entities related to a given entity
func (s *Neo4jStorage) QueryRelatedEntities(entityID string, depth int) ([]GraphEntity, error) {
	session := s.driver.NewSession(s.ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
	})
	defer session.Close(s.ctx)

	query := `
		MATCH (e:Entity {id: $entityId})
		MATCH path = (e)-[:RELATES*1..$depth]-(related)
		RETURN DISTINCT related
	`

	result, err := session.ExecuteRead(s.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(s.ctx, query, map[string]interface{}{
			"entityId": entityID,
			"depth":    depth,
		})
		if err != nil {
			return nil, err
		}
		return result.Collect(s.ctx)
	})

	if err != nil {
		return nil, err
	}

	records, ok := result.([]neo4j.Record)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	var entities []GraphEntity
	for _, record := range records {
		relatedVal, _ := record.Get("related")
		if node, ok := relatedVal.(neo4j.Node); ok {
			entities = append(entities, GraphEntity{
				ID:          node.Props["id"].(string),
				Type:        node.Props["type"].(string),
				Name:        node.Props["name"].(string),
				Description: node.Props["description"].(string),
				Properties:  node.Props,
			})
		}
	}

	return entities, nil
}

// BuildKnowledgeGraph creates entities and relationships from RAG data
func (s *Neo4jStorage) BuildKnowledgeGraph(entities []GraphEntity, relations []GraphRelation) error {
	log.Println("🏗️  Building knowledge graph in Neo4j...")

	// Create all entities first
	for _, entity := range entities {
		if err := s.CreateEntity(entity); err != nil {
			log.Printf("Warning: Failed to create entity %s: %v", entity.ID, err)
		}
	}

	// Create all relations
	for _, relation := range relations {
		if err := s.CreateRelation(relation); err != nil {
			log.Printf("Warning: Failed to create relation %s: %v", relation.ID, err)
		}
	}

	log.Println("✅ Knowledge graph built successfully")
	return nil
}

// DiscoverRelationships identifies relationships between existing entities
func (s *Neo4jStorage) DiscoverRelationships(documents []string) ([]GraphRelation, error) {
	log.Println("🔍 Discovering relationships in knowledge base...")

	// This would analyze document content and create relationship candidates
	// For now, return empty slice as placeholder
	var relations []GraphRelation

	// In production, this would:
	// 1. Parse document content for entity mentions
	// 2. Use NLP to identify relationships
	// 3. Score relationships by confidence
	// 4. Return top candidates for insertion

	return relations, nil
}