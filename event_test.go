package stream

import (
	"testing"
)

func TestEventType(t *testing.T) {
	// 验证 EventType 常量定义
	if EventObjectStart >= EventObjectEnd {
		t.Error("EventType constants should be in order")
	}
	if EventArrayStart >= EventArrayEnd {
		t.Error("EventType constants should be in order")
	}
}

func TestPartialValue(t *testing.T) {
	// 测试 PartialValue 基本结构
	pv := &PartialValue{
		Kind:     ValueString,
		Value:    "test",
		Complete: true,
	}

	if pv.Kind != ValueString {
		t.Errorf("expected ValueString, got %v", pv.Kind)
	}
	if pv.Value != "test" {
		t.Errorf("expected 'test', got %v", pv.Value)
	}
	if !pv.Complete {
		t.Error("expected Complete to be true")
	}
}
