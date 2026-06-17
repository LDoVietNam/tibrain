package patch

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ReplaceOp struct {
	Path            string `json:"path"`
	Old             string `json:"old,omitempty"`
	New             string `json:"new,omitempty"`
	ReplaceAll      bool   `json:"replace_all,omitempty"`
	CreateIfMissing bool   `json:"create_if_missing,omitempty"`
}

type OpsFile struct {
	Reason string      `json:"reason,omitempty"`
	Ops    []ReplaceOp `json:"ops"`
}

type SnapshotFile struct {
	Path    string `json:"path"`
	Existed bool   `json:"existed"`
	Content string `json:"content,omitempty"`
}

type Snapshot struct {
	ID        string         `json:"id"`
	Reason    string         `json:"reason,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	Files     []SnapshotFile `json:"files"`
}

type ApplyReport struct {
	SnapshotID   string   `json:"snapshot_id"`
	ChangedFiles int      `json:"changed_files"`
	Paths        []string `json:"paths,omitempty"`
}

type Manager struct{ root string }

func NewManager(root string) (*Manager, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("empty patch root")
	}
	if err := os.MkdirAll(filepath.Join(root, "snapshots"), 0o755); err != nil {
		return nil, err
	}
	return &Manager{root: root}, nil
}

func (m *Manager) SnapshotFiles(paths []string, reason string) (*Snapshot, error) {
	unique := dedupePaths(paths)
	snap := &Snapshot{ID: snapshotID(), Reason: strings.TrimSpace(reason), CreatedAt: time.Now()}
	for _, p := range unique {
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				snap.Files = append(snap.Files, SnapshotFile{Path: p, Existed: false})
				continue
			}
			return nil, err
		}
		snap.Files = append(snap.Files, SnapshotFile{Path: p, Existed: true, Content: string(data)})
	}
	if err := m.saveSnapshot(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func (m *Manager) ApplyReplaceOps(ops []ReplaceOp, reason string) (*ApplyReport, error) {
	if len(ops) == 0 {
		return nil, errors.New("no patch operations provided")
	}
	paths := make([]string, 0, len(ops))
	for _, op := range ops {
		if strings.TrimSpace(op.Path) == "" {
			return nil, errors.New("patch operation path is required")
		}
		paths = append(paths, op.Path)
	}
	snap, err := m.SnapshotFiles(paths, reason)
	if err != nil {
		return nil, err
	}
	changed := []string{}
	for _, op := range ops {
		path := filepath.Clean(op.Path)
		original, err := os.ReadFile(path)
		missing := os.IsNotExist(err)
		if err != nil && !missing {
			return nil, err
		}
		before := string(original)
		after, err := applyReplace(before, op, missing)
		if err != nil {
			return nil, err
		}
		if after == before {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(after), 0o644); err != nil {
			return nil, err
		}
		changed = append(changed, path)
	}
	return &ApplyReport{SnapshotID: snap.ID, ChangedFiles: len(changed), Paths: changed}, nil
}

func (m *Manager) Rollback(snapshotID string) (*ApplyReport, error) {
	snap, err := m.LoadSnapshot(snapshotID)
	if err != nil {
		return nil, err
	}
	changed := []string{}
	for _, f := range snap.Files {
		path := filepath.Clean(f.Path)
		if !f.Existed {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return nil, err
			}
			changed = append(changed, path)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(f.Content), 0o644); err != nil {
			return nil, err
		}
		changed = append(changed, path)
	}
	return &ApplyReport{SnapshotID: snap.ID, ChangedFiles: len(changed), Paths: changed}, nil
}

func (m *Manager) LoadSnapshot(id string) (*Snapshot, error) {
	data, err := os.ReadFile(filepath.Join(m.root, "snapshots", id+".json"))
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func LoadOpsFile(path string) (*OpsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f OpsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (m *Manager) saveSnapshot(snap *Snapshot) error {
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.root, "snapshots", snap.ID+".json"), data, 0o644)
}

func applyReplace(before string, op ReplaceOp, missing bool) (string, error) {
	if missing {
		if !op.CreateIfMissing {
			return "", os.ErrNotExist
		}
		return op.New, nil
	}
	if op.Old == "" {
		return op.New, nil
	}
	if !strings.Contains(before, op.Old) {
		return "", errors.New("patch old content not found for " + op.Path)
	}
	if op.ReplaceAll {
		return strings.ReplaceAll(before, op.Old, op.New), nil
	}
	return strings.Replace(before, op.Old, op.New, 1), nil
}

func dedupePaths(paths []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = filepath.Clean(strings.TrimSpace(p))
		if p == "." || p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func snapshotID() string {
	return strings.ReplaceAll(time.Now().UTC().Format("20060102T150405.000000000"), ".", "")
}
