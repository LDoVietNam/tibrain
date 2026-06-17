package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RAGCache provides query and embedding caching with Redis backend
type RAGCache struct {
	queryCache     map[string]cacheEntry
	embeddingCache map[string]cacheEntry
	mu             sync.RWMutex
	ttl            time.Duration
	redisClient    *redis.Client
}

type cacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

func NewRAGCache(ttl time.Duration) *RAGCache {
	return &RAGCache{
		queryCache:     make(map[string]cacheEntry),
		embeddingCache: make(map[string]cacheEntry),
		ttl:            ttl,
	}
}

// NewRAGCacheWithRedis creates a cache with Redis backend support
func NewRAGCacheWithRedis(ttl time.Duration, redisAddr string, password string, db int) *RAGCache {
	var client *redis.Client
	if redisAddr != "" {
		client = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: password,
			DB:       db,
		})
	}

	return &RAGCache{
		queryCache:     make(map[string]cacheEntry),
		embeddingCache: make(map[string]cacheEntry),
		ttl:            ttl,
		redisClient:    client,
	}
}

// GetQuery retrieves cached query result with Redis fallback
func (c *RAGCache) GetQuery(query string) (*RAGQuery, bool) {
	key := hashString(query)

	// Try Redis first if available
	if c.redisClient != nil {
		ctx := context.Background()
		val, err := c.redisClient.Get(ctx, "query:"+key).Result()
		if err == nil {
			// Deserialize and return
			return deserializeQuery(val)
		}
	}

	// Fall back to memory cache
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.queryCache[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	result, ok := entry.Value.(*RAGQuery)
	return result, ok
}

// SetQuery stores query result in both memory and Redis
func (c *RAGCache) SetQuery(query string, result *RAGQuery) {
	key := hashString(query)

	// Store in Redis if available
	if c.redisClient != nil {
		ctx := context.Background()
		serialized := serializeQuery(result)
		c.redisClient.Set(ctx, "query:"+key, serialized, c.ttl)
	}

	// Store in memory cache
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queryCache[key] = cacheEntry{
		Value:     result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// GetEmbedding retrieves cached embedding with Redis fallback
func (c *RAGCache) GetEmbedding(text string) ([]float32, bool) {
	key := hashString(text)

	// Try Redis first if available
	if c.redisClient != nil {
		ctx := context.Background()
		val, err := c.redisClient.Get(ctx, "embedding:"+key).Result()
		if err == nil {
			return deserializeEmbedding(val)
		}
	}

	// Fall back to memory cache
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.embeddingCache[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	result, ok := entry.Value.([]float32)
	return result, ok
}

// SetEmbedding stores embedding in both memory and Redis
func (c *RAGCache) SetEmbedding(text string, embedding []float32) {
	key := hashString(text)

	// Store in Redis if available
	if c.redisClient != nil {
		ctx := context.Background()
		serialized := serializeEmbedding(embedding)
		c.redisClient.Set(ctx, "embedding:"+key, serialized, c.ttl)
	}

	// Store in memory cache
	c.mu.Lock()
	defer c.mu.Unlock()

	c.embeddingCache[key] = cacheEntry{
		Value:     embedding,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *RAGCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]interface{}{
		"query_cache_size":     len(c.queryCache),
		"embedding_cache_size": len(c.embeddingCache),
		"ttl_seconds":          c.ttl.Seconds(),
		"redis_enabled":        c.redisClient != nil,
	}
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Helper functions for serialization (simplified implementations)
func serializeQuery(q *RAGQuery) string {
	data, err := json.Marshal(q)
	if err != nil {
		return ""
	}
	return string(data)
}

func deserializeQuery(s string) (*RAGQuery, bool) {
	if strings.TrimSpace(s) == "" {
		return nil, false
	}
	var query RAGQuery
	if err := json.Unmarshal([]byte(s), &query); err != nil {
		return nil, false
	}
	return &query, true
}

func serializeEmbedding(emb []float32) string {
	data, err := json.Marshal(emb)
	if err != nil {
		return ""
	}
	return string(data)
}

func deserializeEmbedding(s string) ([]float32, bool) {
	if strings.TrimSpace(s) == "" {
		return nil, false
	}
	var embedding []float32
	if err := json.Unmarshal([]byte(s), &embedding); err != nil {
		return nil, false
	}
	return embedding, true
}
