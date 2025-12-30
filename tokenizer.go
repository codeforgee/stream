package stream

import (
	"fmt"
	"unicode"
)

// TokenType 表示 token 类型
type TokenType int

const (
	// TokenLBrace 左花括号 {
	TokenLBrace TokenType = iota
	// TokenRBrace 右花括号 }
	TokenRBrace
	// TokenLBracket 左方括号 [
	TokenLBracket
	// TokenRBracket 右方括号 ]
	TokenRBracket
	// TokenColon 冒号 :
	TokenColon
	// TokenComma 逗号 ,
	TokenComma
	// TokenStringChunk 字符串片段（可多次）
	TokenStringChunk
	// TokenStringEnd 字符串结束
	TokenStringEnd
	// TokenNumberChunk 数字片段
	TokenNumberChunk
	// TokenNumberEnd 数字结束
	TokenNumberEnd
	// TokenBool 布尔值 true/false
	TokenBool
	// TokenNull null 值
	TokenNull
)

// Token 表示一个 token
type Token struct {
	// Type token 类型
	Type TokenType
	// Value string/number chunk 的值
	Value string
	// Bool true/false 的值
	Bool bool
}

// tokenizerState 表示 tokenizer 的状态
type tokenizerState int

const (
	// tIdle 空闲状态，等待下一个 token
	tIdle tokenizerState = iota
	// tString 在字符串内部
	tString
	// tStringEscape 字符串转义态
	tStringEscape
	// tNumber 在数字内部
	tNumber
	// tKeyword 在关键字内部（true/false/null）
	tKeyword
)

// String 返回 tokenizer 状态的字符串表示
func (ts tokenizerState) String() string {
	switch ts {
	case tIdle:
		return "Idle"
	case tString:
		return "String"
	case tStringEscape:
		return "StringEscape"
	case tNumber:
		return "Number"
	case tKeyword:
		return "Keyword"
	default:
		return fmt.Sprintf("TokenizerState(%d)", ts)
	}
}

// Tokenizer 将字符流转换为 token 流
type Tokenizer struct {
	state tokenizerState // 当前状态
	buf   []rune         // 临时缓冲区（用于 keyword）
	emit  func(Token)    // token 输出回调
}

// NewTokenizer 创建一个新的 Tokenizer
func NewTokenizer(emit func(Token)) *Tokenizer {
	return &Tokenizer{
		state: tIdle,
		emit:  emit,
	}
}

// Consume 消费一个 rune，可能产生 0 个或多个 token
func (t *Tokenizer) Consume(r rune) {
	switch t.state {
	case tIdle:
		t.consumeIdle(r)
	case tString:
		t.consumeString(r)
	case tStringEscape:
		t.consumeStringEscape(r)
	case tNumber:
		t.consumeNumber(r)
	case tKeyword:
		t.consumeKeyword(r)
	}
}

func (t *Tokenizer) consumeIdle(r rune) {
	switch r {
	case '{':
		t.emit(Token{Type: TokenLBrace})
	case '}':
		t.emit(Token{Type: TokenRBrace})
	case '[':
		t.emit(Token{Type: TokenLBracket})
	case ']':
		t.emit(Token{Type: TokenRBracket})
	case ':':
		t.emit(Token{Type: TokenColon})
	case ',':
		t.emit(Token{Type: TokenComma})
	case '"':
		t.state = tString
		t.buf = t.buf[:0]
	case ' ', '\n', '\r', '\t':
	default:
		if t.isDigit(r) || r == '-' {
			t.state = tNumber
			t.buf = append(t.buf[:0], r)
			t.emit(Token{Type: TokenNumberChunk, Value: string(r)})
			return
		}
		if t.isKeywordStart(r) {
			t.state = tKeyword
			t.buf = append(t.buf[:0], r)
			return
		}
	}
}

func (t *Tokenizer) consumeString(r rune) {
	switch r {
	case '\\':
		t.state = tStringEscape
	case '"':
		t.emit(Token{Type: TokenStringEnd})
		t.state = tIdle
	default:
		t.emit(Token{
			Type:  TokenStringChunk,
			Value: string(r),
		})
	}
}

func (t *Tokenizer) consumeStringEscape(r rune) {
	t.emit(Token{
		Type:  TokenStringChunk,
		Value: string(r),
	})
	t.state = tString
}

func (t *Tokenizer) consumeNumber(r rune) {
	if t.isNumberChar(r) {
		t.emit(Token{
			Type:  TokenNumberChunk,
			Value: string(r),
		})
		return
	}

	t.emit(Token{Type: TokenNumberEnd})
	t.state = tIdle
	t.consumeIdle(r)
}

func (t *Tokenizer) consumeKeyword(r rune) {
	if !unicode.IsLetter(r) {
		t.state = tIdle
		t.consumeIdle(r)
		return
	}

	t.buf = append(t.buf, r)
	bufLen := len(t.buf)

	switch bufLen {
	case 1, 2, 3:
		return
	case 4:
		if t.buf[0] == 't' {
			word := string(t.buf)
			if word == "true" {
				t.emit(Token{Type: TokenBool, Bool: true})
				t.state = tIdle
				t.buf = t.buf[:0]
			}
			return
		}
		if t.buf[0] == 'n' {
			word := string(t.buf)
			if word == "null" {
				t.emit(Token{Type: TokenNull})
				t.state = tIdle
				t.buf = t.buf[:0]
			}
			return
		}
		return
	case 5:
		if t.buf[0] == 'f' {
			word := string(t.buf)
			if word == "false" {
				t.emit(Token{Type: TokenBool, Bool: false})
				t.state = tIdle
				t.buf = t.buf[:0]
			}
			return
		}
		t.state = tIdle
		t.buf = t.buf[:0]
		t.consumeIdle(r)
	default:
		t.state = tIdle
		t.buf = t.buf[:0]
		t.consumeIdle(r)
	}
}

// Close 关闭 tokenizer，处理未完成的状态
func (t *Tokenizer) Close() {
	switch t.state {
	case tString:
		t.emit(Token{Type: TokenStringEnd})
	case tNumber:
		t.emit(Token{Type: TokenNumberEnd})
	}
	t.state = tIdle
}

func (t *Tokenizer) isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func (t *Tokenizer) isKeywordStart(r rune) bool {
	return r == 't' || r == 'f' || r == 'n'
}

func (t *Tokenizer) isNumberChar(r rune) bool {
	return t.isDigit(r) || r == '.' || r == 'e' || r == 'E' || r == '+' || r == '-'
}
