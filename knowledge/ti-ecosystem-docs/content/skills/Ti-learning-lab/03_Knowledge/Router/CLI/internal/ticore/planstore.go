package ticore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type PlanStore struct{ path string }

func NewPlanStore(path string) (*PlanStore, error) {
	if path == "" {
		return nil, errors.New("empty plan store path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &PlanStore{path: path}, nil
}

func (s *PlanStore) Save(plan *ApprovedPlan) error {
	if plan == nil {
		return errors.New("nil approved plan")
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return err
	}
	archiveDir := filepath.Join(filepath.Dir(s.path), "plans")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(archiveDir, plan.ID+".json"), data, 0o644)
}

func (s *PlanStore) LoadCurrent() (*ApprovedPlan, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var plan ApprovedPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *PlanStore) History(limit int) ([]ApprovedPlan, error) {
	archiveDir := filepath.Join(filepath.Dir(s.path), "plans")
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	plans := make([]ApprovedPlan, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(archiveDir, entry.Name()))
		if err != nil {
			continue
		}
		var plan ApprovedPlan
		if json.Unmarshal(data, &plan) == nil {
			plans = append(plans, plan)
		}
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].ApprovedAt.After(plans[j].ApprovedAt) })
	if limit > 0 && len(plans) > limit {
		plans = plans[:limit]
	}
	return plans, nil
}
