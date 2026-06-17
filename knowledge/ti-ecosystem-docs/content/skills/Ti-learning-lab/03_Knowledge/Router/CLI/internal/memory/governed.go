package memory

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Layer string

const (
	LayerCanonical Layer = "canonical"
	LayerEpisodic  Layer = "episodic"
	LayerLearned   Layer = "learned"
)

type Item struct {
	ID         string         `json:"id"`
	Layer      Layer          `json:"layer"`
	Kind       string         `json:"kind"`
	Scope      string         `json:"scope,omitempty"`
	Source     string         `json:"source,omitempty"`
	Agent      string         `json:"agent,omitempty"`
	Confidence float64        `json:"confidence,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
	TTLSeconds int            `json:"ttl_seconds,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type Stats struct {
	Total     int           `json:"total"`
	ByLayer   map[Layer]int `json:"by_layer"`
	Canonical int           `json:"canonical"`
	Episodic  int           `json:"episodic"`
	Learned   int           `json:"learned"`
}

type GovernedStore struct{ root string }

func NewGovernedStore(root string) (*GovernedStore, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &GovernedStore{root: root}, nil
}
func (s *GovernedStore) Close() error { return nil }

func (s *GovernedStore) SaveCanonical(kind, scope, source string, payload map[string]any) error {
	return s.append(Item{Layer: LayerCanonical, Kind: kind, Scope: scope, Source: source, Confidence: 1.0, Payload: payload})
}
func (s *GovernedStore) RecordEpisodic(kind, scope, source string, payload map[string]any) error {
	return s.append(Item{Layer: LayerEpisodic, Kind: kind, Scope: scope, Source: source, Confidence: 0.7, Payload: payload})
}
func (s *GovernedStore) RecordLearned(kind, scope, source string, payload map[string]any) error {
	return s.append(Item{Layer: LayerLearned, Kind: kind, Scope: scope, Source: source, Confidence: 0.8, Payload: payload})
}

func (s *GovernedStore) append(item Item) error {
	if item.ID == "" {
		item.ID = randomID()
	}
	if item.Timestamp.IsZero() {
		item.Timestamp = time.Now()
	}
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	p := s.layerPath(item.Layer)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

func (s *GovernedStore) Stats() (*Stats, error) {
	st := &Stats{ByLayer: map[Layer]int{}}
	for _, layer := range []Layer{LayerCanonical, LayerEpisodic, LayerLearned} {
		entries, err := s.readLayer(layer, 0)
		if err != nil {
			return nil, err
		}
		n := len(entries)
		st.ByLayer[layer] = n
		st.Total += n
		switch layer {
		case LayerCanonical:
			st.Canonical = n
		case LayerEpisodic:
			st.Episodic = n
		case LayerLearned:
			st.Learned = n
		}
	}
	return st, nil
}

func (s *GovernedStore) Recent(layer Layer, limit int) ([]Item, error) {
	return s.readLayer(layer, limit)
}

func (s *GovernedStore) readLayer(layer Layer, limit int) ([]Item, error) {
	p := s.layerPath(layer)
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	out := []Item{}
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item Item
		if json.Unmarshal([]byte(line), &item) == nil {
			out = append(out, item)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

func (s *GovernedStore) layerPath(layer Layer) string {
	return filepath.Join(s.root, string(layer)+".jsonl")
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
