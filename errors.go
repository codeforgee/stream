package stream

import "errors"

var (
	// ErrUnexpectedToken 意外的 token
	ErrUnexpectedToken = errors.New("unexpected token")
	// ErrInvalidState 无效的状态
	ErrInvalidState = errors.New("invalid state")
	// ErrInvalidPattern 无效的路径模式
	ErrInvalidPattern = errors.New("invalid path pattern")
	// ErrUnclosedString 未闭合的字符串
	ErrUnclosedString = errors.New("unclosed string")
	// ErrUnclosedNumber 未闭合的数字
	ErrUnclosedNumber = errors.New("unclosed number")
	// ErrMismatchedBrace 不匹配的大括号
	ErrMismatchedBrace = errors.New("mismatched brace")
	// ErrMismatchedBracket 不匹配的方括号
	ErrMismatchedBracket = errors.New("mismatched bracket")
	// ErrUnexpectedCharacter 意外的字符
	ErrUnexpectedCharacter = errors.New("unexpected character")
)
