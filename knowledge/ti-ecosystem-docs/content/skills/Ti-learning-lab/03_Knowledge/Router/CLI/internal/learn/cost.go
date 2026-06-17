package learn

import (
	"sync"
	"time"
)

// BudgetLimits defines spending caps.
type BudgetLimits struct {
	Daily   float64 `json:"daily"`
	Monthly float64 `json:"monthly"`
}

// CostTracker monitors and optimizes for budget constraints.
type CostTracker struct {
	mu           sync.RWMutex
	daily        float64
	monthly      float64
	limits       BudgetLimits
	dailyReset   time.Time
	monthlyReset time.Time
}

// NewCostTracker creates a new tracker with given limits.
func NewCostTracker(limits BudgetLimits) *CostTracker {
	now := time.Now()
	ct := &CostTracker{
		limits:       limits,
		daily:        0.0,
		monthly:      0.0,
		dailyReset:   time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1),
		monthlyReset: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0),
	}

	// Load persisted usage from DB
	ct.loadFromDB()
	return ct
}

// AddSpent records spending and checks budget.
func (ct *CostTracker) AddSpent(amount float64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	now := time.Now()

	// Reset daily if needed
	if now.After(ct.dailyReset) {
		ct.daily = 0.0
		ct.dailyReset = now.AddDate(0, 0, 1)
	}

	// Reset monthly if needed
	if now.After(ct.monthlyReset) {
		ct.monthly = 0.0
		ct.monthlyReset = now.AddDate(0, 1, 0)
	}

	ct.daily += amount
	ct.monthly += amount

	// Persist to DB asynchronously
	go ct.persist()
}

// CanAfford checks if a cost is within budget.
func (ct *CostTracker) CanAfford(cost float64, priority int) bool {
	// Priority: 1=critical, 5=nice-to-have
	if priority <= 2 {
		return true // High priority always allowed
	}

	ct.mu.RLock()
	defer ct.mu.RUnlock()

	remainingDaily := ct.limits.Daily - ct.daily
	remainingMonthly := ct.limits.Monthly - ct.monthly

	// If either budget exhausted, cannot afford
	if cost > remainingDaily || cost > remainingMonthly {
		return false
	}

	return true
}

// GetDailyUsed returns today's spend.
func (ct *CostTracker) GetDailyUsed() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.daily
}

// GetMonthlyUsed returns this month's spend.
func (ct *CostTracker) GetMonthlyUsed() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.monthly
}

// GetDailyRemaining returns daily budget remaining.
func (ct *CostTracker) GetDailyRemaining() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.limits.Daily - ct.daily
}

// GetMonthlyRemaining returns monthly budget remaining.
func (ct *CostTracker) GetMonthlyRemaining() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.limits.Monthly - ct.monthly
}

// SetLimits updates budget limits.
func (ct *CostTracker) SetLimits(limits BudgetLimits) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.limits = limits
}

// loadFromDB reads daily/monthly usage from DB.
func (ct *CostTracker) loadFromDB() {
	// TODO: Load from budget_tracker table for today
	// For now, use in-memory only
}

// persist saves current usage to DB.
func (ct *CostTracker) persist() {
	// TODO: Upsert into budget_tracker(date, daily_used, monthly_used)
}
