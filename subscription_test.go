package stream

import (
	"testing"
)

// TestSubscription_SimplePath 测试简单路径订阅
func TestSubscription_SimplePath(t *testing.T) {
	var receivedEvents []Event
	p := NewParser()
	p.EnableDebug(DefaultDebugConfig())
	p.On("$.status", func(ev Event) {
		receivedEvents = append(receivedEvents, ev)
	})

	json := `{"status": "running"}`
	if err := p.FeedString(json); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 应该收到 FieldValue 事件
	if len(receivedEvents) == 0 {
		t.Fatal("expected at least one event, got 0")
	}

	found := false
	for _, ev := range receivedEvents {
		if ev.Type == EventFieldValue && ev.Path() == "$.status" {
			if ev.Value != nil && ev.Value.Value == "running" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("expected FieldValue event for $.status with value 'running'")
	}
}

// TestSubscription_Wildcard 测试通配符订阅
func TestSubscription_Wildcard(t *testing.T) {
	var receivedEvents []Event
	p := NewParser()
	p.EnableDebug(DefaultDebugConfig())
	p.On("$.items[*].id", func(ev Event) {
		receivedEvents = append(receivedEvents, ev)
	})

	json := `{"items": [{"id": 1}, {"id": 2}]}`
	if err := p.FeedString(json); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 应该收到两个 FieldValue 事件：$.items[0].id 和 $.items[1].id
	if len(receivedEvents) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(receivedEvents))
	}

	found0 := false
	found1 := false
	for _, ev := range receivedEvents {
		if ev.Type == EventFieldValue && ev.Path() == "$.items[0].id" {
			found0 = true
		}
		if ev.Type == EventFieldValue && ev.Path() == "$.items[1].id" {
			found1 = true
		}
	}

	if !found0 {
		t.Error("expected FieldValue event for $.items[0].id")
	}
	if !found1 {
		t.Error("expected FieldValue event for $.items[1].id")
	}
}

// TestSubscription_MultipleSubscriptions 测试多个订阅
func TestSubscription_MultipleSubscriptions(t *testing.T) {
	var statusEvents []Event
	var progressEvents []Event
	p := NewParser()

	p.On("$.status", func(ev Event) {
		statusEvents = append(statusEvents, ev)
	}).On("$.progress", func(ev Event) {
		progressEvents = append(progressEvents, ev)
	})

	json := `{"status": "running", "progress": 42}`
	if err := p.FeedString(json); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 应该收到 status 事件
	if len(statusEvents) == 0 {
		t.Error("expected status events, got 0")
	}

	// 应该收到 progress 事件
	if len(progressEvents) == 0 {
		t.Error("expected progress events, got 0")
	}
}

// TestSubscription_ArrayItem 测试数组项订阅
func TestSubscription_ArrayItem(t *testing.T) {
	var receivedEvents []Event
	p := NewParser()
	p.EnableDebug(DefaultDebugConfig())
	p.On("$.items[*]", func(ev Event) {
		if ev.Type == EventArrayItem {
			receivedEvents = append(receivedEvents, ev)
		}
	})

	json := `{"items": [{"id": 1}, {"id": 2}]}`
	if err := p.FeedString(json); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 应该收到两个 ArrayItem 事件
	if len(receivedEvents) != 2 {
		t.Fatalf("expected 2 ArrayItem events, got %d", len(receivedEvents))
	}

	if receivedEvents[0].Path() != "$.items[0]" {
		t.Errorf("expected Path='$.items[0]', got '%s'", receivedEvents[0].Path())
	}
	if receivedEvents[1].Path() != "$.items[1]" {
		t.Errorf("expected Path='$.items[1]', got '%s'", receivedEvents[1].Path())
	}
}

// TestSubscription_NestedPath 测试嵌套路径订阅
func TestSubscription_NestedPath(t *testing.T) {
	var receivedEvents []Event
	p := NewParser()
	p.EnableDebug(DefaultDebugConfig())
	p.On("$.data.items[*].name", func(ev Event) {
		receivedEvents = append(receivedEvents, ev)
	})

	json := `{"data": {"items": [{"name": "foo"}, {"name": "bar"}]}}`
	if err := p.FeedString(json); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 应该收到两个 FieldValue 事件
	if len(receivedEvents) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(receivedEvents))
	}

	found0 := false
	found1 := false
	for _, ev := range receivedEvents {
		if ev.Path() == "$.data.items[0].name" && ev.Value != nil && ev.Value.Value == "foo" {
			found0 = true
		}
		if ev.Path() == "$.data.items[1].name" && ev.Value != nil && ev.Value.Value == "bar" {
			found1 = true
		}
	}

	if !found0 {
		t.Error("expected FieldValue event for $.data.items[0].name")
	}
	if !found1 {
		t.Error("expected FieldValue event for $.data.items[1].name")
	}
}

// TestCompilePattern 测试路径模式编译
func TestCompilePattern(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
		check   func(PathPattern) bool
	}{
		{
			name:    "root",
			expr:    "$",
			wantErr: false,
			check: func(p PathPattern) bool {
				return len(p.Segments) == 0
			},
		},
		{
			name:    "simple field",
			expr:    "$.status",
			wantErr: false,
			check: func(p PathPattern) bool {
				return len(p.Segments) == 1 &&
					p.Segments[0].Kind == SegField &&
					p.Segments[0].Value == "status"
			},
		},
		{
			name:    "array index",
			expr:    "$.items[0]",
			wantErr: false,
			check: func(p PathPattern) bool {
				return len(p.Segments) == 2 &&
					p.Segments[0].Kind == SegField &&
					p.Segments[0].Value == "items" &&
					p.Segments[1].Kind == SegIndex &&
					p.Segments[1].Value == "0"
			},
		},
		{
			name:    "array wildcard",
			expr:    "$.items[*]",
			wantErr: false,
			check: func(p PathPattern) bool {
				return len(p.Segments) == 2 &&
					p.Segments[0].Kind == SegField &&
					p.Segments[0].Value == "items" &&
					p.Segments[1].Kind == SegWildcard
			},
		},
		{
			name:    "nested path",
			expr:    "$.items[*].id",
			wantErr: false,
			check: func(p PathPattern) bool {
				return len(p.Segments) == 3 &&
					p.Segments[0].Kind == SegField &&
					p.Segments[0].Value == "items" &&
					p.Segments[1].Kind == SegWildcard &&
					p.Segments[2].Kind == SegField &&
					p.Segments[2].Value == "id"
			},
		},
		{
			name:    "invalid: no $",
			expr:    "status",
			wantErr: true,
		},
		{
			name:    "invalid: empty field",
			expr:    "$.",
			wantErr: true,
		},
		{
			name:    "invalid: empty index",
			expr:    "$.items[]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pat, err := CompilePattern(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.check != nil && !tt.check(pat) {
				t.Errorf("pattern check failed")
			}
		})
	}
}
