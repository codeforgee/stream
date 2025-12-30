package stream

import (
	"testing"
)

func TestTokenizer_SingleCharTokens(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试单字符 token
	testCases := []struct {
		input rune
		want  TokenType
	}{
		{'{', TokenLBrace},
		{'}', TokenRBrace},
		{'[', TokenLBracket},
		{']', TokenRBracket},
		{':', TokenColon},
		{',', TokenComma},
	}

	for _, tc := range testCases {
		tokens = tokens[:0]
		tok.Consume(tc.input)
		if len(tokens) != 1 {
			t.Errorf("expected 1 token for %c, got %d", tc.input, len(tokens))
			continue
		}
		if tokens[0].Type != tc.want {
			t.Errorf("expected %v, got %v", tc.want, tokens[0].Type)
		}
	}
}

func TestTokenizer_String(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试完整字符串
	tokens = tokens[:0]
	for _, r := range `"hello"` {
		tok.Consume(r)
	}

	// 应该产生：StringChunk(h), StringChunk(e), StringChunk(l), StringChunk(l), StringChunk(o), StringEnd
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d: %v", len(tokens), tokens)
	}

	// 检查前5个是 StringChunk
	for i := 0; i < 5; i++ {
		if tokens[i].Type != TokenStringChunk {
			t.Errorf("token[%d] expected StringChunk, got %v", i, tokens[i].Type)
		}
	}

	// 检查最后一个是 StringEnd
	if tokens[5].Type != TokenStringEnd {
		t.Errorf("expected StringEnd, got %v", tokens[5].Type)
	}
}

func TestTokenizer_StringCrossChunk(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 模拟跨 chunk 的字符串
	// chunk 1: "hel
	// chunk 2: lo"
	tokens = tokens[:0]
	for _, r := range `"hel` {
		tok.Consume(r)
	}

	// 应该产生 3 个 StringChunk
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}

	// 继续第二个 chunk
	for _, r := range `lo"` {
		tok.Consume(r)
	}

	// 现在应该有 5 个 StringChunk + 1 个 StringEnd
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[5].Type != TokenStringEnd {
		t.Errorf("expected StringEnd, got %v", tokens[5].Type)
	}
}

func TestTokenizer_StringEscape(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试转义字符（MVP：原样透传）
	for _, r := range `"a\bc"` {
		tok.Consume(r)
	}

	// 根据实现，转义字符本身不emit，只emit转义后的字符
	// 应该产生：StringChunk(a), StringChunk(b), StringChunk(c), StringEnd
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d: %v", len(tokens), tokens)
	}

	// 检查转义后的字符（b）
	if tokens[1].Type != TokenStringChunk || tokens[1].Value != "b" {
		t.Errorf("expected StringChunk(b), got %v", tokens[1])
	}
}

func TestTokenizer_Number(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试数字
	for _, r := range `123` {
		tok.Consume(r)
	}

	// 需要添加一个非数字字符来结束 number
	tok.Consume(',')

	// 应该产生：NumberChunk(1), NumberChunk(2), NumberChunk(3), NumberEnd
	// 然后 Comma 会被重新消费
	if len(tokens) < 4 {
		t.Fatalf("expected at least 4 tokens, got %d: %v", len(tokens), tokens)
	}

	// 检查前3个是 NumberChunk
	for i := 0; i < 3; i++ {
		if tokens[i].Type != TokenNumberChunk {
			t.Errorf("token[%d] expected NumberChunk, got %v", i, tokens[i].Type)
		}
	}

	// 检查 NumberEnd
	if tokens[3].Type != TokenNumberEnd {
		t.Errorf("expected NumberEnd, got %v", tokens[3].Type)
	}

	// 检查逗号被重新消费
	if tokens[4].Type != TokenComma {
		t.Errorf("expected Comma, got %v", tokens[4].Type)
	}
}

func TestTokenizer_NumberCrossChunk(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 模拟跨 chunk 的数字
	// chunk 1: 42
	// chunk 2: 0
	tokens = tokens[:0]
	for _, r := range `42` {
		tok.Consume(r)
	}

	// 应该产生 2 个 NumberChunk
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	// 继续第二个 chunk
	for _, r := range `0,` {
		tok.Consume(r)
	}

	// 现在应该有 3 个 NumberChunk(4,2,0) + 1 个 NumberEnd + 1 个 Comma
	if len(tokens) < 5 {
		t.Fatalf("expected at least 5 tokens, got %d: %v", len(tokens), tokens)
	}

	// tokens[3] 应该是 NumberEnd（在遇到逗号时触发）
	if tokens[3].Type != TokenNumberEnd {
		t.Errorf("expected NumberEnd at index 3, got %v at index %d", tokens[3].Type, 3)
	}
}

func TestTokenizer_NumberNegative(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试负数
	for _, r := range `-42,` {
		tok.Consume(r)
	}

	// 应该产生：NumberChunk(-), NumberChunk(4), NumberChunk(2), NumberEnd, Comma
	if len(tokens) < 5 {
		t.Fatalf("expected at least 5 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[0].Type != TokenNumberChunk || tokens[0].Value != "-" {
		t.Errorf("expected NumberChunk(-), got %v", tokens[0])
	}
}

func TestTokenizer_KeywordTrue(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试 true
	for _, r := range `true` {
		tok.Consume(r)
	}

	// 需要添加一个非字母字符来结束
	tok.Consume(',')

	// 应该产生：TokenBool(true), Comma
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[0].Type != TokenBool {
		t.Errorf("expected TokenBool, got %v", tokens[0].Type)
	}
	if !tokens[0].Bool {
		t.Errorf("expected true, got false")
	}
}

func TestTokenizer_KeywordFalse(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试 false
	for _, r := range `false` {
		tok.Consume(r)
	}
	tok.Consume(',')

	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[0].Type != TokenBool {
		t.Errorf("expected TokenBool, got %v", tokens[0].Type)
	}
	if tokens[0].Bool {
		t.Errorf("expected false, got true")
	}
}

func TestTokenizer_KeywordNull(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试 null
	for _, r := range `null` {
		tok.Consume(r)
	}
	tok.Consume(',')

	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[0].Type != TokenNull {
		t.Errorf("expected TokenNull, got %v", tokens[0].Type)
	}
}

func TestTokenizer_KeywordInvalid(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试无效的关键字（trux）
	for _, r := range `trux` {
		tok.Consume(r)
	}
	tok.Consume(',')

	// 应该丢弃无效关键字，只产生 Comma
	// 或者可能产生其他 token（取决于实现）
	// 关键是不应该产生 TokenBool
	for _, tok := range tokens {
		if tok.Type == TokenBool {
			t.Errorf("should not produce TokenBool for invalid keyword")
		}
	}
}

func TestTokenizer_Close_String(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试未完成的字符串
	for _, r := range `"hello` {
		tok.Consume(r)
	}

	// 关闭 tokenizer
	tok.Close()

	// 应该产生：StringChunk(h), StringChunk(e), StringChunk(l), StringChunk(l), StringChunk(o), StringEnd
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d: %v", len(tokens), tokens)
	}

	// 最后一个应该是 StringEnd
	if tokens[len(tokens)-1].Type != TokenStringEnd {
		t.Errorf("expected StringEnd, got %v", tokens[len(tokens)-1].Type)
	}
}

func TestTokenizer_Close_Number(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试未完成的数字
	for _, r := range `123` {
		tok.Consume(r)
	}

	// 关闭 tokenizer
	tok.Close()

	// 应该产生：NumberChunk(1), NumberChunk(2), NumberChunk(3), NumberEnd
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[len(tokens)-1].Type != TokenNumberEnd {
		t.Errorf("expected NumberEnd, got %v", tokens[len(tokens)-1].Type)
	}
}

func TestTokenizer_Whitespace(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试空白字符被忽略
	for _, r := range ` { } ` {
		tok.Consume(r)
	}

	// 应该只产生 LBrace 和 RBrace，空白字符被忽略
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}

	if tokens[0].Type != TokenLBrace {
		t.Errorf("expected LBrace, got %v", tokens[0].Type)
	}
	if tokens[1].Type != TokenRBrace {
		t.Errorf("expected RBrace, got %v", tokens[1].Type)
	}
}

func TestTokenizer_ComplexJSON(t *testing.T) {
	var tokens []Token
	tok := NewTokenizer(func(t Token) {
		tokens = append(tokens, t)
	})

	// 测试复杂 JSON
	json := `{"a": 1, "b": [true, false], "c": null}`
	for _, r := range json {
		tok.Consume(r)
	}

	// 验证基本结构
	// 应该有：LBrace, StringChunk/End, Colon, NumberChunk/End, Comma, ...
	// 这里只做基本检查
	if len(tokens) < 10 {
		t.Fatalf("expected at least 10 tokens, got %d", len(tokens))
	}

	// 第一个应该是 LBrace
	if tokens[0].Type != TokenLBrace {
		t.Errorf("expected LBrace, got %v", tokens[0].Type)
	}

	// 最后一个应该是 RBrace
	if tokens[len(tokens)-1].Type != TokenRBrace {
		t.Errorf("expected RBrace, got %v", tokens[len(tokens)-1].Type)
	}
}
