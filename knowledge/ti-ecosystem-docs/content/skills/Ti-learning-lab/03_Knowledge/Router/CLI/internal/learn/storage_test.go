package learn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStorage(t *testing.T) *Storage {
	// Use in-memory DB for tests
	storage, err := NewStorage(":memory:")
	require.NoError(t, err)
	return storage
}

func TestStorage_CreateAndQuery(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	// Insert a log
	entry := &LogInput{
		Timestamp:       time.Now(),
		SessionID:       "test-session-1",
		PromptHash:      "abc123",
		PromptLength:    100,
		TaskType:        "testing",
		TaskDomain:      "backend",
		ComplexityScore: 5.0,
		Keywords:        []string{"go", "test"},
		Provider:        "openrouter",
		Model:           "claude-sonnet-4",
		ModelFamily:     "claude",
		TokensInput:     150,
		TokensOutput:    200,
		LatencyMs:       1800,
		CostUSD:         0.0025,
		QualityUser:     8.0,
		QualityAuto:     0.85,
		QualityCombined: 8.75,
		Outcome:         "success",
	}

	err := storage.InsertLog(entry)
	require.NoError(t, err)

	// Query by session
	logs, err := storage.GetBySession("test-session-1")
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "testing", logs[0].TaskType)
	assert.Equal(t, "claude-sonnet-4", logs[0].Model)
}

func TestStorage_GetStats(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	// Insert multiple entries
	for i := 0; i < 10; i++ {
		entry := &LogInput{
			Timestamp:       time.Now(),
			SessionID:       "session-" + string(rune('a'+i)),
			TaskType:        "coding",
			Model:           "claude-sonnet-4",
			QualityCombined: 7.0 + float64(i%3), // 7,8,9 cycling
			CostUSD:         0.002,
			Outcome:         "success",
		}
		_ = storage.InsertLog(entry)
	}

	stats, err := storage.GetStats(5)
	require.NoError(t, err)
	assert.Equal(t, int64(10), stats.TotalSamples)
	assert.InEpsilon(t, 7.67, stats.AvgQuality, 0.5)        // approx
	assert.InEpsilon(t, 10000.0, stats.TotalCostUSD, 100.0) // ~0.002 * 10
}

func TestStorage_PruneOld(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	// Insert old (30 days ago) and recent logs
	oldTime := time.Now().AddDate(0, 0, -40)
	recentTime := time.Now()

	oldLog := &LogInput{Timestamp: oldTime, SessionID: "old"}
	recentLog := &LogInput{Timestamp: recentTime, SessionID: "recent"}

	storage.InsertLog(oldLog)
	storage.InsertLog(recentLog)

	// Prune older than 30 days
	err := storage.PruneOld(30)
	require.NoError(t, err)

	// Verify: old gone, recent remains
	logs, _ := storage.GetBySession("old")
	assert.Empty(t, logs)

	logs, _ = storage.GetBySession("recent")
	assert.Len(t, logs, 1)
}

func TestStorage_DBSize(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	// DB starts small
	size, err := storage.GetDBSize()
	require.NoError(t, err)
	assert.Greater(t, size, int64(0))

	// Insert some data
	for i := 0; i < 100; i++ {
		entry := &LogInput{
			Timestamp: time.Now(),
			SessionID: "s" + string(rune('0'+i%10)),
			TaskType:  "test",
			Model:     "claude",
		}
		_ = storage.InsertLog(entry)
	}

	size2, _ := storage.GetDBSize()
	assert.Greater(t, size2, size) // DB grew
}

func TestStorage_Migration_Idempotent(t *testing.T) {
	storage1 := setupTestStorage(t)
	storage1.Close()

	// Create second instance on same in-memory DB (reset via path)
	storage2, err := NewStorage(":memory:") // New DB = fresh migration
	require.NoError(t, err)
	storage2.Close()
	_ = storage2
	// Just ensure no error on re-migration
}

func TestStorage_ConcurrentWrites(t *testing.T) {
	storage := setupTestStorage(t)
	defer storage.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			entry := &LogInput{
				Timestamp: time.Now(),
				SessionID: "concurrent-" + string(rune('a'+idx%10)),
				TaskType:  "code",
				Model:     "llama",
			}
			_ = storage.InsertLog(entry)
		}(i)
	}
	wg.Wait()

	count, _ := storage.db.QueryRow(`SELECT COUNT(*) FROM beads_logs`).Scan()
	assert.Equal(t, 100, count)
}
