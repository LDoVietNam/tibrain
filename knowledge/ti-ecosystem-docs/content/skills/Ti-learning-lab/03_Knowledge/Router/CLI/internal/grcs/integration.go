package grcs

import (
	"context"
	"fmt"
	"github.com/ti/cli/internal/agent"
	"github.com/ti/cli/internal/memory"
	"github.com/ti/cli/internal/routeagent"
	"github.com/ti/cli/internal/session"
	"net/http"
)

// GRCSIntegration integrates GRCS with Ti core systems
type GRCSIntegration struct {
	grcs     *GRCS
	memory   *memory.Palace
	sessions *session.Store
	mux      *http.ServeMux
}

func NewGRCSIntegration(grcs *GRCS, mem *memory.Palace, sessions *session.Store, registry *agent.Registry) *GRCSIntegration {
	return &GRCSIntegration{
		grcs:     grcs,
		memory:   mem,
		sessions: sessions,
		mux:      routeagent.AgentMux(registry),
	}
}

// RunHybridWorkflow executes Agent + GRCS Judge pipeline
func (gi *GRCSIntegration) RunHybridWorkflow(ctx context.Context, task string) (string, error) {
	// Step 1: Generate K candidate outputs using Agent
	candidates, err := gi.generateCandidates(ctx, task)
	if err != nil {
		return "", err
	}

	// Step 2: GRCS Rank & Select best candidate
	best, err := gi.grcs.SelectBest(candidates)
	if err != nil {
		return "", err
	}

	// Step 3: Store ranking history in Memory Palace
	gi.storeRanking(candidates, best)

	return best.Content, nil
}

func (gi *GRCSIntegration) generateCandidates(ctx context.Context, task string) ([]*Candidate, error) {
	var candidates []*Candidate

	// Generate K independent candidate outputs via session store
	for i := 0; i < gi.grcs.K; i++ {
		_ = ctx
		// Placeholder: in production, dispatch to routeagent mux via HTTP
		candidates = append(candidates, &Candidate{
			Content: fmt.Sprintf("candidate-%d for: %s", i, task),
		})
	}

	return candidates, nil
}

func (gi *GRCSIntegration) storeRanking(candidates []*Candidate, best *Candidate) {
	// Store scoring history in Memory Palace using AddDrawer
	for i, candidate := range candidates {
		isSelected := candidate == best
		importance := candidate.Score
		if isSelected {
			importance += 1.0 // boost selected candidate
		}
		_ = gi.memory.AddDrawer("grcs", "rankings", memory.Drawer{
			ID:         fmt.Sprintf("grcs-rank-%d", i+1),
			Text:       candidate.Content,
			Importance: importance,
			Category:   "grcs_ranking",
		})
	}
}
