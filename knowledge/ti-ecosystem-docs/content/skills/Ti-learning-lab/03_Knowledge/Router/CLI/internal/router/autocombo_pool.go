// Package router - default auto-combo pool.
// Provides a default set of provider candidates for the chat proxy.
package router

import "github.com/ti/cli/internal/autocombo"

// AutoComboPool returns the default pool of provider candidates.
// This is used by the management chat proxy to route requests to the
// best available provider/model based on cost, latency, quota, and tier.
func AutoComboPool() []autocombo.ProviderCandidate {
	return []autocombo.ProviderCandidate{
		{
			Provider: "openrouter", Model: "anthropic/claude-sonnet-4-5",
			QuotaRemaining: 80, CircuitState: autocombo.CircuitClosed,
			CostPer1MTokens: 3.0, P95LatencyMs: 1800, LatencyStdDev: 300,
			AccountTier: autocombo.TierPro,
		},
		{
			Provider: "openrouter", Model: "openai/gpt-4o",
			QuotaRemaining: 90, CircuitState: autocombo.CircuitClosed,
			CostPer1MTokens: 5.0, P95LatencyMs: 1500, LatencyStdDev: 250,
			AccountTier: autocombo.TierStandard,
		},
		{
			Provider: "groq", Model: "llama-3.3-70b-versatile",
			QuotaRemaining: 85, CircuitState: autocombo.CircuitClosed,
			CostPer1MTokens: 0.59, P95LatencyMs: 400, LatencyStdDev: 80,
			AccountTier: autocombo.TierFree, QuotaResetSecs: 86400,
		},
		{
			Provider: "google", Model: "gemini-2.5-flash",
			QuotaRemaining: 90, CircuitState: autocombo.CircuitClosed,
			CostPer1MTokens: 0.075, P95LatencyMs: 900, LatencyStdDev: 150,
			AccountTier: autocombo.TierFree, QuotaResetSecs: 86400,
		},
	}
}
