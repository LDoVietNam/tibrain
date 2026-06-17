package learn

import (
	"sync"
	"time"
)

// Cache holds hot data in memory.
type Cache struct {
	mu          sync.RWMutex
	topModels   map[string][]string // taskType → []model keys
	topPatterns []Pattern
	stats       *Stats
	statsTime   time.Time
	hits        int
	misses      int
}

// NewCache creates a new cache.
func NewCache() *Cache {
	return &Cache{
		topModels:   make(map[string][]string),
		topPatterns: make([]Pattern, 0),
		stats:       &Stats{},
		statsTime:   time.Time{},
	}
}

// GetTopModels returns cached top models for task type.
func (c *Cache) GetTopModels(taskType string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	models := c.topModels[taskType]
	if len(models) == 0 {
		c.misses++
	} else {
		c.hits++
	}
	return models
}

// SetTopModels updates cache for task type.
func (c *Cache) SetTopModels(taskType string, models []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.topModels[taskType] = models
}

// GetStats returns cached stats.
func (c *Cache) GetStats() *Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	stats := *c.stats
	return &stats
}

// SetStats updates stats cache.
func (c *Cache) SetStats(stats *Stats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats = stats
	c.statsTime = time.Now()
}

// GetPatterns returns cached patterns.
func (c *Cache) GetPatterns() []Pattern {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.topPatterns
}

// SetPatterns updates patterns cache.
func (c *Cache) SetPatterns(patterns []Pattern) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.topPatterns = patterns
}

// Invalidate clears cache.
func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.topModels = make(map[string][]string)
	c.statsTime = time.Time{}
}

// HitRate returns cache hit ratio.
func (c *Cache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0.0
	}
	return float64(c.hits) / float64(total)
}

// RefreshTopModels reloads top models from DB.
func (c *Cache) RefreshTopModels(storage *Storage) error {
	// Get top models per task type from DB
	query := `
		SELECT task_type, model, provider, AVG(quality_combined) as avg_q, COUNT(*) as cnt
		FROM beads_logs
		WHERE quality_combined > 0
		GROUP BY task_type, provider, model
		HAVING cnt >= ?
		ORDER BY task_type, avg_q DESC, cnt DESC
	`

	rows, err := storage.db.Query(query, 5) // min 5 samples
	if err != nil {
		return err
	}
	defer rows.Close()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset
	c.topModels = make(map[string][]string)

	for rows.Next() {
		var taskType, model, provider string
		var avgQ float64
		var cnt int
		if err := rows.Scan(&taskType, &model, &provider, &avgQ, &cnt); err != nil {
			continue
		}
		key := provider + ":" + model
		c.topModels[taskType] = append(c.topModels[taskType], key)
	}

	return nil
}
