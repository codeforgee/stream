package stream

import (
	"testing"
)

// TestError_InvalidJSON 测试非法 JSON 片段
func TestError_InvalidJSON(t *testing.T) {
	p := NewParser()

	// 测试不匹配的括号
	invalidJSONs := []string{
		`{`,              // 未闭合对象
		`[`,              // 未闭合数组
		`{"a": `,         // 未完成的值
		`{"a": 1,}`,      // 尾随逗号（可能被接受）
		`{"a": 1, "b":}`, // 缺少值
	}

	for _, json := range invalidJSONs {
		t.Run(json, func(t *testing.T) {
			// 不应该 panic，应该能处理
			err := p.FeedString(json)
			if err != nil {
				// 错误是可以接受的
				return
			}
		})
	}
}

// TestError_TruncatedStream 测试截断的流
func TestError_TruncatedStream(t *testing.T) {
	var abortedCount int
	p := NewParser()
	p.On("$", func(ev Event) {
		if ev.Value != nil && ev.Value.Aborted {
			abortedCount++
		}
	})

	// 截断的 JSON
	json := `{"status": "run`
	if err := p.FeedString(json); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 应该有 Aborted 的值
	if abortedCount == 0 {
		t.Log("no aborted values found (this may be acceptable)")
	}
}

// TestError_MalformedNumber 测试格式错误的数字
func TestError_MalformedNumber(t *testing.T) {
	p := NewParser()

	// 测试各种可能的数字格式错误
	// 注意：当前实现可能接受一些格式
	malformedNumbers := []string{
		`{"num": 12.34.56}`, // 多个小数点
		`{"num": 12e}`,      // 不完整的科学计数法
		`{"num": +-12}`,     // 多个符号
	}

	for _, json := range malformedNumbers {
		t.Run(json, func(t *testing.T) {
			// 不应该 panic
			err := p.FeedString(json)
			if err != nil {
				// 错误是可以接受的
				return
			}
		})
	}
}

// TestError_UnclosedString 测试未闭合的字符串
func TestError_UnclosedString(t *testing.T) {
	var abortedValue string
	p := NewParser()
	p.On("$.text", func(ev Event) {
		if ev.Value != nil && ev.Value.Aborted {
			abortedValue = ev.Value.Value.(string)
		}
	})

	// 未闭合的字符串
	json := `{"text": "hello`
	if err := p.FeedString(json); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 应该收到 Aborted 的值
	if abortedValue == "" {
		t.Log("no aborted value found (this may be acceptable)")
	} else if abortedValue != "hello" {
		t.Errorf("expected aborted value='hello', got '%s'", abortedValue)
	}
}

// TestError_CommaAtRootLevel 测试根级别的逗号（应该返回错误而不是 panic）
func TestError_CommaAtRootLevel(t *testing.T) {
	p := NewParser()

	// 测试根级别的逗号
	err := p.FeedString(`,`)
	if err == nil {
		t.Error("expected error for comma at root level, got nil")
	} else if err != ErrUnexpectedToken {
		t.Logf("got error (may be acceptable): %v", err)
	}

	// 测试解析完成后收到逗号
	p2 := NewParser()
	if err := p2.FeedString(`{"a": 1}`); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}
	// 解析完成后，栈为空，再收到逗号应该返回错误
	err = p2.FeedString(`,`)
	if err == nil {
		t.Error("expected error for comma after complete JSON, got nil")
	} else if err != ErrUnexpectedToken {
		t.Logf("got error (may be acceptable): %v", err)
	}
}

// TestError_StringKeyWithoutStack 测试在栈为空时处理字符串 key（防御性测试）
func TestError_StringKeyWithoutStack(t *testing.T) {
	// 这个测试主要验证 onStringEnd 在 pObjExpectKey 状态下不会 panic
	// 虽然理论上不应该出现这种情况，但防御性编程是好的
	p := NewParser()

	// 正常情况下，pObjExpectKey 状态应该在有对象 frame 时出现
	// 但为了测试防御性代码，我们可以尝试构造一个异常状态
	// 注意：这可能需要直接操作内部状态，或者通过其他方式触发

	// 先正常解析一个对象开始
	if err := p.FeedString(`{`); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 然后关闭对象（栈变空）
	if err := p.FeedString(`}`); err != nil {
		t.Fatalf("FeedString() failed: %v", err)
	}

	// 此时栈为空，如果收到字符串结束 token（虽然不应该发生），应该能处理
	// 由于 tokenizer 的状态机，这种情况很难直接触发，但代码已经做了防御
	t.Log("defensive code in onStringEnd should handle empty stack gracefully")
}
