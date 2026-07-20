package trace

import (
	"reflect"
	"testing"
)

func TestInMemoryTracer_Create(t *testing.T) {
	tracer := NewInMemoryTracer()
	if tracer == nil {
		t.Errorf("Expected non-nil tracer")
		return
	}
	
	if tracer.events == nil {
		t.Errorf("Expected events map to be initialized")
		return
	}
	
	if len(tracer.events) != 0 {
		t.Errorf("Expected empty events map, got %d entries", len(tracer.events))
	}
}

func TestInMemoryTracer_TraceEvent(t *testing.T) {
	tracer := NewInMemoryTracer()
	
	// Test tracing an event
	tracer.Trace("exec123", "test_event", map[string]interface{}{"key": "value"})
	
	// Get the trace and verify
	events := tracer.GetTrace("exec123")
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
		return
	}
	
	event := events[0]
	if event.Event != "test_event" {
		t.Errorf("Expected event 'test_event', got '%s'", event.Event)
	}
	
	// Check the data
	expectedData := map[string]interface{}{"key": "value"}
	if !reflect.DeepEqual(event.Data, expectedData) {
		t.Errorf("Expected data %v, got %v", expectedData, event.Data)
	}
	
	// Check that timestamp is not zero
	if event.Timestamp.IsZero() {
		t.Errorf("Expected timestamp to be set")
	}
}

func TestInMemoryTracer_MultipleEvents(t *testing.T) {
	tracer := NewInMemoryTracer()
	
	// Trace multiple events
	tracer.Trace("exec123", "first", map[string]interface{}{"data": "data1"})
	tracer.Trace("exec123", "second", map[string]interface{}{"1": "one"})
	tracer.Trace("exec123", "third", map[string]interface{}{"items": []string{"a", "b"}})
	
	// Get the trace
	events := tracer.GetTrace("exec123")
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
		return
	}
	
	// Check first event
	if events[0].Event != "first" {
		t.Errorf("Expected first event 'first', got '%s'", events[0].Event)
	}
	if !reflect.DeepEqual(events[0].Data, map[string]interface{}{"data": "data1"}) {
		t.Errorf("Expected first data map[data:data1], got '%v'", events[0].Data)
	}
	
	// Check second event
	if events[1].Event != "second" {
		t.Errorf("Expected second event 'second', got '%s'", events[1].Event)
	}
	
	// Check third event
	if events[2].Event != "third" {
		t.Errorf("Expected third event 'third', got '%s'", events[2].Event)
	}
}

func TestInMemoryTracer_GetTrace_NonExistent(t *testing.T) {
	tracer := NewInMemoryTracer()
	
	// Try to get a trace that doesn't exist
	events := tracer.GetTrace("non-existent")
	if len(events) != 0 {
		t.Errorf("Expected empty slice for non-existent trace, got %d events", len(events))
	}
}

func TestInMemoryTracer_GetTrace_SpecificID(t *testing.T) {
	tracer := NewInMemoryTracer()
	
	// Trace to different IDs
	tracer.Trace("exec111", "event1", map[string]interface{}{"data": "data1"})
	tracer.Trace("exec222", "event2", map[string]interface{}{"data": "data2"})
	
	// Get the trace for exec111
	events := tracer.GetTrace("exec111")
	if len(events) != 1 {
		t.Errorf("Expected 1 event for exec111, got %d", len(events))
		return
	}
	
	if events[0].Event != "event1" {
		t.Errorf("Expected event 'event1', got '%s'", events[0].Event)
	}
	
	if !reflect.DeepEqual(events[0].Data, map[string]interface{}{"data": "data1"}) {
		t.Errorf("Expected data map[data:data1], got '%v'", events[0].Data)
	}
	
	// Get the trace for exec222
	events = tracer.GetTrace("exec222")
	if len(events) != 1 {
		t.Errorf("Expected 1 event for exec222, got %d", len(events))
		return
	}
	
	if events[0].Event != "event2" {
		t.Errorf("Expected event 'event2', got '%s'", events[0].Event)
	}
	
	if !reflect.DeepEqual(events[0].Data, map[string]interface{}{"data": "data2"}) {
		t.Errorf("Expected data map[data:data2], got '%v'", events[0].Data)
	}
}