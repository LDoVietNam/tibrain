package knowledge

import (
	"context"
	"fmt"
	"time"
)

// IngestionService handles knowledge ingestion from external sources
type IngestionService struct {
	store *KnowledgeStore
}

// NewIngestionService creates a new ingestion service
func NewIngestionService(store *KnowledgeStore) *IngestionService {
	return &IngestionService{store: store}
}

// IngestFromNotionSync ingests pages from Notion sync
func (s *IngestionService) IngestFromNotionSync(ctx context.Context, pages []NotionPage) error {
	for _, page := range pages {
		doc := Document{
			ID:       page.ID,
			Title:    page.Title,
			Content:  page.Content,
			Metadata: map[string]interface{}{
				"source":  "notion",
				"synced":  time.Now().Unix(),
				"parent":  page.ParentID,
				"url":     page.URL,
			},
		}
		if err := s.store.Store(ctx, doc); err != nil {
			return fmt.Errorf("failed to store document %s: %w", page.ID, err)
		}
	}
	return nil
}

// NotionPage represents a page from Notion API
type NotionPage struct {
	ID       string
	Title    string
	Content  string
	ParentID string
	URL      string
}

// SyncFromNotion syncs knowledge from Notion database
func (s *IngestionService) SyncFromNotion(ctx context.Context, databaseID string) error {
	// This would integrate with notionprovider
	_, err := s.store.List(ctx) // placeholder
	return err
}