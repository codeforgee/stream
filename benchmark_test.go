package stream

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkParser_SimpleObject 测试简单对象的解析性能
func BenchmarkParser_SimpleObject(b *testing.B) {
	json := `{"status": "running", "progress": 42, "message": "processing"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_SimpleObjectWithSubscription 测试带订阅的简单对象解析性能
func BenchmarkParser_SimpleObjectWithSubscription(b *testing.B) {
	json := `{"status": "running", "progress": 42, "message": "processing"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		p.On("$.status", func(ev Event) {})
		p.On("$.progress", func(ev Event) {})
		p.On("$.message", func(ev Event) {})
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_LargeArray 测试大数组的解析性能
func BenchmarkParser_LargeArray(b *testing.B) {
	// 生成包含 1000 个元素的数组
	var items []string
	for i := 0; i < 1000; i++ {
		items = append(items, fmt.Sprintf(`{"id": %d, "name": "item%d", "value": %d}`, i, i, i*2))
	}
	json := "[" + strings.Join(items, ",") + "]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_LargeArrayWithWildcard 测试大数组带通配符订阅的解析性能
func BenchmarkParser_LargeArrayWithWildcard(b *testing.B) {
	// 生成包含 1000 个元素的数组
	var items []string
	for i := 0; i < 1000; i++ {
		items = append(items, fmt.Sprintf(`{"id": %d, "name": "item%d", "value": %d}`, i, i, i*2))
	}
	json := "[" + strings.Join(items, ",") + "]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		p.On("$[*].id", func(ev Event) {})
		p.On("$[*].name", func(ev Event) {})
		p.On("$[*].value", func(ev Event) {})
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_DeeplyNested 测试深度嵌套结构的解析性能
func BenchmarkParser_DeeplyNested(b *testing.B) {
	// 生成深度为 10 的嵌套对象
	json := `{"level1": {"level2": {"level3": {"level4": {"level5": {"level6": {"level7": {"level8": {"level9": {"level10": {"value": 42}}}}}}}}}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_StreamingChunks 测试流式分块输入的解析性能
func BenchmarkParser_StreamingChunks(b *testing.B) {
	// 模拟 LLM 流式输出，将 JSON 切分成多个小块
	chunks := []string{
		`{"status": "`,
		`running`,
		`", "progress": `,
		`42`,
		`, "items": [`,
		`{"id": 1, "name": "foo"}, `,
		`{"id": 2, "name": "bar"}`,
		`]}`,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		for _, chunk := range chunks {
			if err := p.FeedString(chunk); err != nil {
				b.Fatal(err)
			}
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_LargeString 测试大字符串的解析性能
func BenchmarkParser_LargeString(b *testing.B) {
	// 生成 10KB 的字符串值
	largeString := strings.Repeat("a", 10*1024)
	json := fmt.Sprintf(`{"message": "%s"}`, largeString)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_ManySubscriptions 测试多个订阅的性能
func BenchmarkParser_ManySubscriptions(b *testing.B) {
	json := `{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8, "i": 9, "j": 10}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		// 订阅 10 个不同的字段
		p.On("$.a", func(ev Event) {})
		p.On("$.b", func(ev Event) {})
		p.On("$.c", func(ev Event) {})
		p.On("$.d", func(ev Event) {})
		p.On("$.e", func(ev Event) {})
		p.On("$.f", func(ev Event) {})
		p.On("$.g", func(ev Event) {})
		p.On("$.h", func(ev Event) {})
		p.On("$.i", func(ev Event) {})
		p.On("$.j", func(ev Event) {})
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_ComplexNestedArray 测试复杂嵌套数组的解析性能
func BenchmarkParser_ComplexNestedArray(b *testing.B) {
	// 生成包含嵌套数组和对象的复杂结构
	var items []string
	for i := 0; i < 100; i++ {
		items = append(items, fmt.Sprintf(`{"id": %d, "tags": ["tag1", "tag2", "tag3"], "metadata": {"count": %d, "active": true}}`, i, i*2))
	}
	json := fmt.Sprintf(`{"items": [%s], "total": 100}`, strings.Join(items, ","))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_NumberParsing 测试数字解析的性能
func BenchmarkParser_NumberParsing(b *testing.B) {
	// 生成包含大量数字的 JSON
	var fields []string
	for i := 0; i < 1000; i++ {
		fields = append(fields, fmt.Sprintf(`"field%d": %d.123456`, i, i))
	}
	json := "{" + strings.Join(fields, ",") + "}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_StringChunkFlushing 测试字符串分块刷新的性能
func BenchmarkParser_StringChunkFlushing(b *testing.B) {
	// 模拟流式字符串输入，每个字符一个 chunk
	baseString := strings.Repeat("a", 1000)
	chunks := make([]string, 0, len(baseString))
	for _, r := range baseString {
		chunks = append(chunks, string(r))
	}
	jsonPrefix := `{"message": "`
	jsonSuffix := `"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		p.On("$.message", func(ev Event) {})
		if err := p.FeedString(jsonPrefix); err != nil {
			b.Fatal(err)
		}
		for _, chunk := range chunks {
			if err := p.FeedString(chunk); err != nil {
				b.Fatal(err)
			}
		}
		if err := p.FeedString(jsonSuffix); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_MixedTypes 测试混合类型的解析性能
func BenchmarkParser_MixedTypes(b *testing.B) {
	json := `{
		"string": "test",
		"number": 42,
		"float": 3.14,
		"bool": true,
		"null": null,
		"array": [1, 2, 3],
		"object": {"nested": "value"}
	}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTokenizer_SimpleInput 测试 tokenizer 的简单输入性能
func BenchmarkTokenizer_SimpleInput(b *testing.B) {
	json := `{"status": "running", "progress": 42}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := NewTokenizer(func(t Token) {})
		for _, r := range json {
			tok.Consume(r)
		}
		tok.Close()
	}
}

// BenchmarkTokenizer_LargeInput 测试 tokenizer 的大输入性能
func BenchmarkTokenizer_LargeInput(b *testing.B) {
	// 生成 100KB 的 JSON 字符串
	largeString := strings.Repeat("a", 100*1024)
	json := fmt.Sprintf(`{"message": "%s"}`, largeString)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tok := NewTokenizer(func(t Token) {})
		for _, r := range json {
			tok.Consume(r)
		}
		tok.Close()
	}
}

// BenchmarkPathMatching 测试路径匹配的性能
func BenchmarkPathMatching(b *testing.B) {
	json := `{"items": [{"id": 1, "name": "foo"}, {"id": 2, "name": "bar"}]}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		// 订阅多个路径模式
		p.On("$.items[*].id", func(ev Event) {})
		p.On("$.items[*].name", func(ev Event) {})
		p.On("$.items[*]", func(ev Event) {})
		if err := p.FeedString(json); err != nil {
			b.Fatal(err)
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParser_RealWorldLLMStream 测试真实 LLM 流式输出场景的性能
func BenchmarkParser_RealWorldLLMStream(b *testing.B) {
	// 模拟真实的 LLM 流式输出，包含多个字段和数组
	chunks := [][]byte{
		[]byte(`{"status": "`),
		[]byte(`processing`),
		[]byte(`", "progress": `),
		[]byte(`75`),
		[]byte(`, "items": [`),
		[]byte(`{"id": 1, "name": "item1", "score": 95.5}, `),
		[]byte(`{"id": 2, "name": "item2", "score": 87.3}, `),
		[]byte(`{"id": 3, "name": "item3", "score": 92.1}`),
		[]byte(`], "message": "`),
		[]byte(`Processing `),
		[]byte(`complete`),
		[]byte(`"}`),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser()
		p.On("$.status", func(ev Event) {})
		p.On("$.progress", func(ev Event) {})
		p.On("$.items[*].id", func(ev Event) {})
		p.On("$.items[*].name", func(ev Event) {})
		p.On("$.items[*].score", func(ev Event) {})
		p.On("$.message", func(ev Event) {})
		for _, chunk := range chunks {
			if err := p.Feed(chunk); err != nil {
				b.Fatal(err)
			}
		}
		if err := p.Close(true); err != nil {
			b.Fatal(err)
		}
	}
}
