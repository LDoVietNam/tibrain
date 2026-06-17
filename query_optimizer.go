package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// QueryOptimizer optimizes RAG queries for better performance
type QueryOptimizer struct {
	tibrain           *Hub
	cache             *QueryCache
	metrics           *MetricsCollector
	config            *OptimizerConfig
	queryPatterns     []QueryPattern
	optimizationRules []OptimizationRule
	queryPredictions  map[string]QueryPrediction
	performanceCache  map[string]CachedPerformance
}

// OptimizerConfig holds configuration for query optimization
type OptimizerConfig struct {
	CacheEnabled    bool          `json:"cache_enabled"`
	CacheSize       int           `json:"cache_size"`
	CacheTTL        time.Duration `json:"cache_ttl"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	Timeout         time.Duration `json:"timeout"`
	ParallelQueries int           `json:"parallel_queries"`
	BatchSize       int           `json:"batch_size"`
}

// QueryCache implements caching for RAG queries
type QueryCache struct {
	cache map[string]*CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
	size  int
}

type CacheEntry struct {
	Response  *RAGResponse
	Timestamp time.Time
	HitCount  int
	LastUsed  time.Time
}

// QueryRequest represents an optimized query request
type QueryRequest struct {
	Query     string        `json:"query"`
	Tier      string        `json:"tier"`
	TopK      int           `json:"top_k"`
	Threshold float64       `json:"threshold"`
	Category  string        `json:"category"`
	Tags      []string      `json:"tags"`
	Timeout   time.Duration `json:"timeout"`
	UserID    string        `json:"user_id"`
	SessionID string        `json:"session_id"`
}

// RAGResponse represents the response from RAG system
type RAGResponse struct {
	Results      []*RAGResult   `json:"results"`
	Query        string         `json:"query"`
	ResponseTime time.Duration  `json:"response_time"`
	Confidence   float64        `json:"confidence"`
	Tier         string         `json:"tier"`
	Cached       bool           `json:"cached"`
	Metadata     *QueryMetadata `json:"metadata"`
}

// RAGResult represents a single search result
type RAGResult struct {
	DocumentID string                 `json:"document_id"`
	FilePath   string                 `json:"file_path"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Score      float64                `json:"score"`
	Relevance  float64                `json:"relevance"`
	Category   string                 `json:"category"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// QueryMetadata contains metadata about the query
type QueryMetadata struct {
	TotalResults int           `json:"total_results"`
	SearchTime   time.Duration `json:"search_time"`
	IndexTime    time.Duration `json:"index_time"`
	FilterTime   time.Duration `json:"filter_time"`
	RankTime     time.Duration `json:"rank_time"`
	Processing   string        `json:"processing"`
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(hub *Hub, config *OptimizerConfig) *QueryOptimizer {
	cache := &QueryCache{
		cache: make(map[string]*CacheEntry),
		ttl:   config.CacheTTL,
		size:  config.CacheSize,
	}

	return &QueryOptimizer{
		tibrain: hub,
		cache:   cache,
		metrics: NewMetricsCollector(),
		config:  config,
	}
}

// OptimizeQuery optimizes a query for better performance
func (qo *QueryOptimizer) OptimizeQuery(ctx context.Context, query string, options ...QueryOption) (*RAGResponse, error) {
	// Create optimized query request
	req := &QueryRequest{
		Query:     query,
		Tier:      "hot",
		TopK:      5,
		Threshold: 0.7,
		Timeout:   qo.config.Timeout,
	}

	// Apply options
	for _, opt := range options {
		opt(req)
	}

	// Check cache first
	if qo.config.CacheEnabled {
		if cached := qo.cache.Get(req); cached != nil {
			qo.metrics.RecordCacheHit()
			cached.Cached = true
			return cached, nil
		}
	}

	// Optimize query string
	optimizedQuery := qo.optimizeQueryString(query)
	req.Query = optimizedQuery

	// Execute optimized query with retries
	response, err := qo.executeQueryWithRetries(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Cache the response
	if qo.config.CacheEnabled && response != nil {
		qo.cache.Set(req, response)
	}

	// Record metrics
	qo.metrics.RecordQuery(response)

	return response, nil
}

// QueryOption represents a query option
type QueryOption func(*QueryRequest)

// WithTier sets the query tier
func WithTier(tier string) QueryOption {
	return func(req *QueryRequest) {
		req.Tier = tier
	}
}

// WithTopK sets the top-k results
func WithTopK(topK int) QueryOption {
	return func(req *QueryRequest) {
		req.TopK = topK
	}
}

// WithThreshold sets the confidence threshold
func WithThreshold(threshold float64) QueryOption {
	return func(req *QueryRequest) {
		req.Threshold = threshold
	}
}

// WithCategory sets the query category
func WithCategory(category string) QueryOption {
	return func(req *QueryRequest) {
		req.Category = category
	}
}

// WithTags sets the query tags
func WithTags(tags ...string) QueryOption {
	return func(req *QueryRequest) {
		req.Tags = tags
	}
}

// optimizeQueryString improves the query string for better matching
func (qo *QueryOptimizer) optimizeQueryString(query string) string {
	// Convert to lowercase for case-insensitive search
	query = strings.ToLower(query)

	// Remove common stop words
	stopWords := []string{"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for"}
	for _, stopWord := range stopWords {
		query = strings.ReplaceAll(query, " "+stopWord+" ", " ")
	}

	// Remove extra spaces
	query = strings.Join(strings.Fields(query), " ")

	// Add semantic enhancements
	query = qo.addSemanticEnhancements(query)

	return query
}

// addSemanticEnhancements adds semantic improvements to the query
func (qo *QueryOptimizer) addSemanticEnhancements(query string) string {
	// Add related terms
	relatedTerms := map[string][]string{
		"implement":    {"create", "build", "develop", "code"},
		"configure":    {"setup", "install", "deploy", "settings"},
		"troubleshoot": {"fix", "debug", "solve", "resolve"},
		"optimize":     {"improve", "enhance", "speed up", "performance"},
		"test":         {"verify", "validate", "check", "confirm"},
	}

	words := strings.Fields(query)
	var enhancedWords []string

	for _, word := range words {
		enhancedWords = append(enhancedWords, word)

		// Add related terms
		if terms, exists := relatedTerms[word]; exists {
			enhancedWords = append(enhancedWords, terms...)
		}
	}

	return strings.Join(enhancedWords, " ")
}

// executeQueryWithRetries executes a query with retry logic
func (qo *QueryOptimizer) executeQueryWithRetries(ctx context.Context, req *QueryRequest) (*RAGResponse, error) {
	var lastErr error

	for attempt := 0; attempt < qo.config.MaxRetries; attempt++ {
		// Set timeout for this attempt
		ctx, cancel := context.WithTimeout(ctx, req.Timeout)

		// Execute query
		response, err := qo.tibrain.Query(ctx, req.Query, req.Tier, req.TopK, req.Threshold)

		cancel()

		if err == nil {
			return response, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			break
		}

		// Wait before retry
		time.Sleep(qo.config.RetryDelay * time.Duration(attempt+1))
	}

	return nil, lastErr
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())

	// List of retryable error patterns
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"temporary",
		"service unavailable",
		"rate limit",
		"network error",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// BatchQuery executes multiple queries in parallel
func (qo *QueryOptimizer) BatchQuery(ctx context.Context, queries []string, options ...QueryOption) ([]*RAGResponse, error) {
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries provided")
	}

	// Create query requests
	requests := make([]*QueryRequest, len(queries))
	for i, query := range queries {
		req := &QueryRequest{
			Query:   query,
			Tier:    "hot",
			TopK:    5,
			Timeout: qo.config.Timeout,
		}

		// Apply options
		for _, opt := range options {
			opt(req)
		}

		requests[i] = req
	}

	// Execute queries in parallel
	responses := make([]*RAGResponse, len(queries))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Limit concurrent queries
	maxConcurrent := qo.config.ParallelQueries
	if maxConcurrent > len(requests) {
		maxConcurrent = len(requests)
	}

	// Process queries in batches
	for i := 0; i < len(requests); i += maxConcurrent {
		end := i + maxConcurrent
		if end > len(requests) {
			end = len(requests)
		}

		batch := requests[i:end]

		for j, req := range batch {
			wg.Add(1)
			go func(idx int, r *QueryRequest) {
				defer wg.Done()

				response, err := qo.OptimizeQuery(ctx, r.Query)
				mu.Lock()
				if err != nil {
					errors = append(errors, fmt.Errorf("query %d failed: %w", idx, err))
				} else {
					responses[idx] = response
				}
				mu.Unlock()
			}(j, req)
		}

		wg.Wait()
	}

	if len(errors) > 0 {
		return responses, fmt.Errorf("batch query completed with %d errors: %v", len(errors), errors)
	}

	return responses, nil
}

// GetCacheStats returns cache statistics
func (qo *QueryOptimizer) GetCacheStats() *CacheStats {
	qo.cache.mutex.RLock()
	defer qo.cache.mutex.RUnlock()

	return &CacheStats{
		Size:      len(qo.cache.cache),
		HitCount:  qo.getHitCount(),
		MissCount: qo.getMissCount(),
		HitRate:   qo.getHitRate(),
	}
}

// CacheStats contains cache statistics
type CacheStats struct {
	Size      int     `json:"size"`
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	HitRate   float64 `json:"hit_rate"`
}

// getHitCount returns total cache hits
func (qo *QueryOptimizer) getHitCount() int64 {
	var count int64
	for _, entry := range qo.cache.cache {
		count += int64(entry.HitCount)
	}
	return count
}

// getMissCount returns total cache misses
func (qo *QueryOptimizer) getMissCount() int64 {
	return qo.metrics.GetMissCount()
}

// getHitRate returns cache hit rate
func (qo *QueryOptimizer) getHitRate() float64 {
	hits := qo.getHitCount()
	misses := qo.getMissCount()

	if hits+misses == 0 {
		return 0
	}

	return float64(hits) / float64(hits+misses) * 100
}

// Get returns a cached response
func (qc *QueryCache) Get(req *QueryRequest) *RAGResponse {
	qc.mutex.RLock()
	defer qc.mutex.RUnlock()

	key := qc.generateKey(req)

	if entry, exists := qc.cache[key]; exists {
		// Check if entry is expired
		if time.Since(entry.Timestamp) > qc.ttl {
			delete(qc.cache, key)
			return nil
		}

		// Update hit count and last used time
		entry.HitCount++
		entry.LastUsed = time.Now()

		return entry.Response
	}

	return nil
}

// Set stores a response in cache
func (qc *QueryCache) Set(req *QueryRequest, response *RAGResponse) {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()

	key := qc.generateKey(req)

	// Check if cache is full
	if len(qc.cache) >= qc.size {
		qc.evictLRU()
	}

	qc.cache[key] = &CacheEntry{
		Response:  response,
		Timestamp: time.Now(),
		HitCount:  0,
		LastUsed:  time.Now(),
	}
}

// evictLRU removes least recently used entries
func (qc *QueryCache) evictLRU() {
	var oldestTime time.Time
	var oldestKey string

	for key, entry := range qc.cache {
		if oldestTime.IsZero() || entry.LastUsed.Before(oldestTime) {
			oldestTime = entry.LastUsed
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(qc.cache, oldestKey)
	}
}

// generateKey generates a cache key for the request
func (qc *QueryCache) generateKey(req *QueryRequest) string {
	return fmt.Sprintf("%s:%s:%d:%.2f", req.Query, req.Tier, req.TopK, req.Threshold)
}

// Clear clears the cache
func (qc *QueryCache) Clear() {
	qc.mutex.Lock()
	defer qc.mutex.Unlock()

	qc.cache = make(map[string]*CacheEntry)
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	mutex             sync.RWMutex
	queryCount        int64
	cacheHits         int64
	cacheMisses       int64
	totalResponseTime time.Duration
	minResponseTime   time.Duration
	maxResponseTime   time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// RecordQuery records query metrics
func (mc *MetricsCollector) RecordQuery(response *RAGResponse) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.queryCount++
	mc.totalResponseTime += response.ResponseTime

	if mc.minResponseTime == 0 || response.ResponseTime < mc.minResponseTime {
		mc.minResponseTime = response.ResponseTime
	}

	if response.ResponseTime > mc.maxResponseTime {
		mc.maxResponseTime = response.ResponseTime
	}
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.cacheHits++
}

// GetMissCount returns cache miss count
func (mc *MetricsCollector) GetMissCount() int64 {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return mc.cacheMisses
}

// GetAverageResponseTime returns average response time
func (mc *MetricsCollector) GetAverageResponseTime() time.Duration {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if mc.queryCount == 0 {
		return 0
	}

	return mc.totalResponseTime / time.Duration(mc.queryCount)
}

// GetStats returns current statistics
func (mc *MetricsCollector) GetStats() *QueryStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return &QueryStats{
		QueryCount:          mc.queryCount,
		CacheHits:           mc.cacheHits,
		CacheMisses:         mc.cacheMisses,
		AverageResponseTime: mc.GetAverageResponseTime(),
		MinResponseTime:     mc.minResponseTime,
		MaxResponseTime:     mc.maxResponseTime,
	}
}

// QueryStats contains query statistics
type QueryStats struct {
	QueryCount          int64         `json:"query_count"`
	CacheHits           int64         `json:"cache_hits"`
	CacheMisses         int64         `json:"cache_misses"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
}
