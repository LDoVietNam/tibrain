package ticore

import (
	"strings"
	"time"

	"github.com/ti/cli/internal/planner"
)

type PlanState string
type StepStatus string

const (
	PlanStateApproved       PlanState = "approved"
	PlanStateExecuting      PlanState = "executing"
	PlanStateAwaitingVerify PlanState = "awaiting_verify"
	PlanStateBlocked        PlanState = "blocked"
	PlanStateCompleted      PlanState = "completed"
)

const (
	StepPending     StepStatus = "pending"
	StepInProgress  StepStatus = "in_progress"
	StepImplemented StepStatus = "implemented"
	StepVerified    StepStatus = "verified"
	StepCompleted   StepStatus = "completed"
	StepBlocked     StepStatus = "blocked"
	StepSkipped     StepStatus = "skipped"
)

type ApprovedStep struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Objective       string     `json:"objective,omitempty"`
	Files           []string   `json:"files,omitempty"`
	Commands        []string   `json:"commands,omitempty"`
	ExitCriteria    []string   `json:"exit_criteria,omitempty"`
	Status          StepStatus `json:"status"`
	AttemptCount    int        `json:"attempt_count,omitempty"`
	LastStartedAt   *time.Time `json:"last_started_at,omitempty"`
	LastUpdatedAt   *time.Time `json:"last_updated_at,omitempty"`
	LastCompletedAt *time.Time `json:"last_completed_at,omitempty"`
	LastNote        string     `json:"last_note,omitempty"`
}

type Progress struct {
	Total       int `json:"total"`
	Pending     int `json:"pending"`
	InProgress  int `json:"in_progress"`
	Implemented int `json:"implemented"`
	Verified    int `json:"verified"`
	Completed   int `json:"completed"`
	Blocked     int `json:"blocked"`
	Skipped     int `json:"skipped"`
}

func stepsFromPlanner(steps []planner.Step) []ApprovedStep {
	out := make([]ApprovedStep, 0, len(steps))
	for _, s := range steps {
		out = append(out, ApprovedStep{
			ID:           s.ID,
			Title:        s.Title,
			Objective:    s.Objective,
			Files:        append([]string(nil), s.Files...),
			Commands:     append([]string(nil), s.Commands...),
			ExitCriteria: append([]string(nil), s.ExitCriteria...),
			Status:       StepPending,
		})
	}
	return out
}

func (p *ApprovedPlan) CurrentStep() *ApprovedStep {
	if p == nil || p.CurrentStepID == "" {
		return nil
	}
	for i := range p.Steps {
		if p.Steps[i].ID == p.CurrentStepID {
			return &p.Steps[i]
		}
	}
	return nil
}

func (p *ApprovedPlan) FindStep(preferred string) *ApprovedStep {
	if p == nil {
		return nil
	}
	preferred = strings.TrimSpace(preferred)
	if preferred == "" {
		return nil
	}
	for i := range p.Steps {
		s := &p.Steps[i]
		if strings.EqualFold(s.ID, preferred) || strings.EqualFold(s.Title, preferred) {
			return s
		}
	}
	return nil
}

func (p *ApprovedPlan) PendingVerificationStep() *ApprovedStep {
	if p == nil {
		return nil
	}
	if cur := p.CurrentStep(); cur != nil && (cur.Status == StepImplemented || cur.Status == StepVerified) {
		return cur
	}
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.Status == StepImplemented || s.Status == StepVerified {
			return s
		}
	}
	return nil
}

func (p *ApprovedPlan) SelectNextStep(preferred string) *ApprovedStep {
	if p == nil {
		return nil
	}
	if p.PendingVerificationStep() != nil {
		return nil
	}
	if s := p.FindStep(preferred); s != nil {
		switch s.Status {
		case StepPending, StepBlocked, StepInProgress:
			return s
		default:
			return nil
		}
	}
	if cur := p.CurrentStep(); cur != nil {
		switch cur.Status {
		case StepPending, StepBlocked, StepInProgress:
			return cur
		}
	}
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.Status == StepPending || s.Status == StepBlocked {
			return s
		}
	}
	return nil
}

func (p *ApprovedPlan) SelectStepForVerification(preferred string) *ApprovedStep {
	if p == nil {
		return nil
	}
	if s := p.FindStep(preferred); s != nil {
		switch s.Status {
		case StepImplemented, StepVerified, StepInProgress, StepBlocked:
			return s
		}
	}
	if cur := p.CurrentStep(); cur != nil {
		switch cur.Status {
		case StepImplemented, StepVerified, StepInProgress, StepBlocked:
			return cur
		}
	}
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.Status == StepImplemented || s.Status == StepVerified || s.Status == StepInProgress || s.Status == StepBlocked {
			return s
		}
	}
	return nil
}

func (p *ApprovedPlan) NextActionableStep() *ApprovedStep {
	if p == nil || p.PendingVerificationStep() != nil {
		return nil
	}
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.Status == StepPending || s.Status == StepBlocked {
			return s
		}
	}
	return nil
}

func (p *ApprovedPlan) MarkStepStarted(stepID, note string) bool {
	if p == nil {
		return false
	}
	now := time.Now()
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.ID != stepID {
			continue
		}
		s.Status = StepInProgress
		s.AttemptCount++
		s.LastStartedAt = &now
		s.LastUpdatedAt = &now
		s.LastNote = strings.TrimSpace(note)
		p.CurrentStepID = s.ID
		p.State = PlanStateExecuting
		return true
	}
	return false
}

func (p *ApprovedPlan) MarkStepImplemented(stepID, note string) bool {
	return p.updateStep(stepID, StepImplemented, strings.TrimSpace(note), true)
}

func (p *ApprovedPlan) MarkStepVerified(stepID, note string) bool {
	return p.updateStep(stepID, StepVerified, strings.TrimSpace(note), true)
}

func (p *ApprovedPlan) MarkStepCompleted(stepID, note string) bool {
	return p.updateStep(stepID, StepCompleted, strings.TrimSpace(note), true)
}

func (p *ApprovedPlan) MarkStepSkipped(stepID, note string) bool {
	return p.updateStep(stepID, StepSkipped, strings.TrimSpace(note), true)
}

func (p *ApprovedPlan) MarkStepBlocked(stepID, note string) bool {
	return p.updateStep(stepID, StepBlocked, strings.TrimSpace(note), false)
}

func (p *ApprovedPlan) updateStep(stepID string, status StepStatus, note string, complete bool) bool {
	if p == nil {
		return false
	}
	now := time.Now()
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.ID != stepID {
			continue
		}
		s.Status = status
		s.LastUpdatedAt = &now
		if complete {
			s.LastCompletedAt = &now
		}
		s.LastNote = strings.TrimSpace(note)
		if status == StepCompleted || status == StepSkipped {
			p.advanceAfterCompletion(stepID)
		} else {
			p.CurrentStepID = s.ID
		}
		p.RecomputeState()
		return true
	}
	return false
}

func (p *ApprovedPlan) advanceAfterCompletion(stepID string) {
	if p == nil {
		return
	}
	seenCurrent := false
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.ID == stepID {
			seenCurrent = true
			continue
		}
		if seenCurrent && (s.Status == StepPending || s.Status == StepBlocked) {
			p.CurrentStepID = s.ID
			return
		}
	}
	for i := range p.Steps {
		s := &p.Steps[i]
		if s.Status == StepPending || s.Status == StepBlocked {
			p.CurrentStepID = s.ID
			return
		}
	}
	p.CurrentStepID = stepID
}

func (p *ApprovedPlan) RecomputeState() {
	if p == nil {
		return
	}
	progress := p.Progress()
	switch {
	case progress.Total > 0 && progress.Completed+progress.Skipped == progress.Total:
		p.State = PlanStateCompleted
	case progress.Implemented > 0 || progress.Verified > 0:
		p.State = PlanStateAwaitingVerify
	case progress.Blocked > 0 && progress.Completed+progress.Skipped < progress.Total && progress.InProgress == 0:
		p.State = PlanStateBlocked
	case progress.InProgress > 0 || progress.Completed+progress.Skipped > 0:
		p.State = PlanStateExecuting
	default:
		p.State = PlanStateApproved
	}
}

func (p *ApprovedPlan) Progress() Progress {
	prog := Progress{Total: len(p.Steps)}
	for _, s := range p.Steps {
		switch s.Status {
		case StepPending, "":
			prog.Pending++
		case StepInProgress:
			prog.InProgress++
		case StepImplemented:
			prog.Implemented++
		case StepVerified:
			prog.Verified++
		case StepCompleted:
			prog.Completed++
		case StepBlocked:
			prog.Blocked++
		case StepSkipped:
			prog.Skipped++
		}
	}
	return prog
}
