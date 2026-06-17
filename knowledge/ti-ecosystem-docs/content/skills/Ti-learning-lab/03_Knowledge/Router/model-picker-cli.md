# Model Picker CLI Tool

> **Created**: 2026-04-29  
> **Purpose**: Tài liệu về Model Picker CLI Tool cho Ti Router  
> **Language**: Vietnamese

---

## Overview

Model Picker CLI Tool là công cụ dòng lệnh để chọn model nhanh chóng từ danh sách tất cả models có sẵn. Tool này tương tự như `claude-pick` trong free-claude-code-main nhưng được implement bằng Go thay vì fzf.

## Features

1. **List Models** - Liệt kê tất cả models từ tất cả providers
2. **Interactive Selection** - Chọn model interactively với keyboard
3. **Filter Options** - Filter theo tier, provider, cost
4. **Config Generation** - Generate config snippet cho model đã chọn
5. **Model Metadata** - Hiển thị metadata (cost, context size, capabilities)

## Implementation Plan

### 1. Create CLI Structure

```
cmd/modelpick/
├── main.go          # CLI entry point
├── picker.go        # Model picker logic
├── filters.go       # Filter logic
└── config.go        # Config generation
```

### 2. Interactive Selection

Sử dụng Go library cho interactive selection:

```go
package main

import (
    "fmt"
    "os"
    "github.com/manifoldco/promptui"
)

type ModelPicker struct {
    models []ModelMetadata
    filter ModelFilter
}

func (p *ModelPicker) Select() (ModelMetadata, error) {
    // Convert models to promptui items
    items := make([]string, len(p.models))
    for i, model := range p.models {
        items[i] = fmt.Sprintf("%s (%s) - %s", model.ID, model.Provider, model.Tier)
    }
    
    prompt := promptui.Select{
        Label: "Select Model",
        Items: items,
    }
    
    _, result, err := prompt.Run()
    if err != nil {
        return ModelMetadata{}, err
    }
    
    // Find selected model
    for _, model := range p.models {
        if fmt.Sprintf("%s (%s) - %s", model.ID, model.Provider, model.Tier) == result {
            return model, nil
        }
    }
    
    return ModelMetadata{}, fmt.Errorf("model not found")
}
```

### 3. Filter Options

```go
type ModelFilter struct {
    Tier     string
    Provider string
    MaxCost  float64
    MinCtx   int
}

func (f *ModelFilter) Apply(models []ModelMetadata) []ModelMetadata {
    var filtered []ModelMetadata
    
    for _, model := range models {
        if f.Tier != "" && model.Tier != f.Tier {
            continue
        }
        if f.Provider != "" && model.Provider != f.Provider {
            continue
        }
        if f.MaxCost > 0 && model.CostPer1K > f.MaxCost {
            continue
        }
        if f.MinCtx > 0 && model.ContextSize < f.MinCtx {
            continue
        }
        filtered = append(filtered, model)
    }
    
    return filtered
}
```

### 4. Config Generation

```go
func GenerateConfigSnippet(model ModelMetadata) string {
    return fmt.Sprintf(`# Model: %s
# Provider: %s
# Tier: %s
# Cost: $%.4f/1K tokens
# Context: %d tokens

model_tiers:
  %s: "%s"
`, model.ID, model.Provider, model.Tier, model.CostPer1K, model.ContextSize, model.Tier, model.ID)
}
```

### 5. CLI Interface

```go
func main() {
    // Load models from router config
    models := loadModels()
    
    // Apply filters
    filter := parseFilters()
    filtered := filter.Apply(models)
    
    // Interactive selection
    picker := NewModelPicker(filtered)
    selected, err := picker.Select()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    
    // Generate config
    config := GenerateConfigSnippet(selected)
    fmt.Println(config)
}
```

## Usage

```bash
# Build
cd Z:\Ti\router
go build -o bin/modelpick.exe ./cmd/modelpick

# Run
./bin/modelpick.exe

# With filters
./bin/modelpick.exe --tier sonnet --provider groq
./bin/modelpick.exe --max-cost 0.01
./bin/modelpick.exe --min-ctx 32000
```

## Testing

```go
func TestModelPicker(t *testing.T) {
    models := []ModelMetadata{
        {ID: "groq/llama-3.3-70b", Tier: "opus", Provider: "groq"},
        {ID: "openrouter/deepseek", Tier: "sonnet", Provider: "openrouter"},
    }
    
    picker := NewModelPicker(models)
    
    // Test filter
    filter := ModelFilter{Tier: "opus"}
    filtered := filter.Apply(models)
    
    if len(filtered) != 1 {
        t.Errorf("Expected 1 model, got %d", len(filtered))
    }
    
    // Test config generation
    config := GenerateConfigSnippet(filtered[0])
    if !strings.Contains(config, "groq/llama-3.3-70b") {
        t.Error("Config should contain model ID")
    }
}
```

## Integration with Ti Router

### Current State

- Router có model registry với ModelMetadata
- Router có config file với model definitions
- Router có provider registry

### Changes Needed

1. Create cmd/modelpick/ directory
2. Implement CLI tool with interactive selection
3. Add model loading from router config
4. Add filter logic
5. Add config generation
6. Build and add to core/tools/

## Success Criteria

- [x] CLI tool created
- [x] Interactive model selection works
- [x] Filter options implemented
- [x] Config generation works
- [x] Tests pass
- [x] Binary added to core/tools/
- [x] AGENTS.md updated

---

**Last Updated**: 2026-04-29
