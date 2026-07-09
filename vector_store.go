// TiBrain Vector Store
// SQLite-based vector storage using the existing rag_vector_index table
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// VectorChunk represents an embedded document chunk stored in SQLite
type VectorChunk struct {
	ID         string            `json:"id"`
	DocumentID string            `json:"document_id"`
	Vector     []float32         `json:"vector"`
	Content    string            `json:"content"`
	Metadata   map[string]string `json:"metadata"`
	ChunkOrder int               `json:"chunk_order"`
	CreatedAt  int64             `json:"created_at"`
}

// VectorStore manages vector data in the hub database
type VectorStore struct {
	db        *sql.DB
	dimension int
}

// NewVectorStore creates a vector store using the hub database
func NewVectorStore(db *sql.DB, dimension int) *VectorStore {
	return &VectorStore{
		db:        db,
		dimension: dimension,
	}
}

// AddEmbedding adds or updates a vector embedding in the store
func (vs *VectorStore) AddEmbedding(chunk VectorChunk) error {
	vectorJSON, err := json.Marshal(chunk.Vector)
	if err != nil {
		return fmt.Errorf("marshal vector: %w", err)
	}

	_, err = vs.db.Exec(`
		INSERT OR REPLACE INTO rag_vector_index
		(id, document_id, vector_data, chunk_text, chunk_order, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, chunk.ID, chunk.DocumentID, string(vectorJSON), chunk.Content, chunk.ChunkOrder, chunk.CreatedAt)

	if err != nil {
		return fmt.Errorf("insert embedding: %w", err)
	}
	return nil
}

// GetEmbeddingByDocument retrieves all embeddings for a document
func (vs *VectorStore) GetEmbeddingByDocument(docID string) ([]VectorChunk, error) {
	rows, err := vs.db.Query(`
		SELECT id, document_id, vector_data, chunk_text, chunk_order, created_at
		FROM rag_vector_index
		WHERE document_id = ?
		ORDER BY chunk_order ASC
	`, docID)
	if err != nil {
		return nil, fmt.Errorf("query embeddings: %w", err)
	}
	defer rows.Close()

	var chunks []VectorChunk
	for rows.Next() {
		var chunk VectorChunk
		var vectorJSON string
		err := rows.Scan(&chunk.ID, &chunk.DocumentID, &vectorJSON, &chunk.Content, &chunk.ChunkOrder, &chunk.CreatedAt)
		if err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(vectorJSON), &chunk.Vector); err != nil {
			continue
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// GetAllEmbeddings retrieves all embeddings from the store
func (vs *VectorStore) GetAllEmbeddings() ([]VectorChunk, error) {
	rows, err := vs.db.Query(`
		SELECT id, document_id, vector_data, chunk_text, chunk_order, created_at
		FROM rag_vector_index
	`)
	if err != nil {
		return nil, fmt.Errorf("query all embeddings: %w", err)
	}
	defer rows.Close()

	var chunks []VectorChunk
	for rows.Next() {
		var chunk VectorChunk
		var vectorJSON string
		err := rows.Scan(&chunk.ID, &chunk.DocumentID, &vectorJSON, &chunk.Content, &chunk.ChunkOrder, &chunk.CreatedAt)
		if err != nil {
			logger.Warn("Failed to scan embedding: %v", err)
			continue
		}
		if err := json.Unmarshal([]byte(vectorJSON), &chunk.Vector); err != nil {
			logger.Warn("Failed to unmarshal vector: %v", err)
			continue
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// DeleteEmbeddingsByDocument removes all embeddings for a document
func (vs *VectorStore) DeleteEmbeddingsByDocument(docID string) error {
	_, err := vs.db.Exec(`
		DELETE FROM rag_vector_index WHERE document_id = ?
	`, docID)
	if err != nil {
		return fmt.Errorf("delete embeddings: %w", err)
	}
	return nil
}

// SearchSimilar performs cosine similarity search against embeddings, optionally filtered by a SQL clause
func (vs *VectorStore) SearchSimilar(queryVector []float32, topK int, threshold float64, sqlFilter string) ([]VectorSearchResult, error) {
	var chunks []VectorChunk
	var err error

	if sqlFilter != "" {
		// Pre-filter vector index using a SQL JOIN on documents matching the query scope filter
		query := `
			SELECT v.id, v.document_id, v.vector_data, v.chunk_text, v.chunk_order, v.created_at
			FROM rag_vector_index v
			JOIN rag_documents d ON v.document_id = d.id
			WHERE d.status = 'active' ` + sqlFilter

		rows, errQuery := vs.db.Query(query)
		if errQuery != nil {
			return nil, fmt.Errorf("query filtered embeddings: %w", errQuery)
		}
		defer rows.Close()

		for rows.Next() {
			var chunk VectorChunk
			var vectorJSON string
			errScan := rows.Scan(&chunk.ID, &chunk.DocumentID, &vectorJSON, &chunk.Content, &chunk.ChunkOrder, &chunk.CreatedAt)
			if errScan != nil {
				continue
			}
			if errScan = json.Unmarshal([]byte(vectorJSON), &chunk.Vector); errScan != nil {
				continue
			}
			chunks = append(chunks, chunk)
		}
	} else {
		chunks, err = vs.GetAllEmbeddings()
		if err != nil {
			return nil, err
		}
	}

	if len(chunks) == 0 {
		return nil, nil
	}

	results := make([]VectorSearchResult, 0)

	for _, chunk := range chunks {
		if len(chunk.Vector) != len(queryVector) {
			continue
		}

		score := cosineSimilarity(queryVector, chunk.Vector)
		if score >= threshold {
			results = append(results, VectorSearchResult{
				ChunkID:    chunk.ID,
				DocumentID: chunk.DocumentID,
				Content:    chunk.Content,
				Score:      score,
				ChunkOrder: chunk.ChunkOrder,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top-K
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// CountEmbeddings returns the total number of embeddings
func (vs *VectorStore) CountEmbeddings() (int, error) {
	var count int
	err := vs.db.QueryRow(`SELECT COUNT(*) FROM rag_vector_index`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count embeddings: %w", err)
	}
	return count, nil
}

// VectorSearchResult represents a similarity search result
type VectorSearchResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	ChunkOrder int     `json:"chunk_order"`
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
