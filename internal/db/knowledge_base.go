package db

import "time"

// RAGKnowledgeBase describes a RAG knowledge base instance. It is referenced
// by tests and by callers that enumerate/manage knowledge bases.
type RAGKnowledgeBase struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DocCount    int       `json:"doc_count"`
	VectorCount int       `json:"vector_count"`
}
