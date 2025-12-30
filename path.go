package stream

import (
	"fmt"
	"strconv"
	"strings"
)

// pathOptions 路径生成选项
type pathOptions struct {
	excludeTop      bool // 是否排除顶层 frame（用于 ObjectEnd/ArrayEnd）
	excludeTopIndex bool // 是否排除顶层 array 的 index（用于 ArrayStart）
}

func buildPathFromSegments(segments []PathSegment, opt pathOptions) string {
	if len(segments) == 0 {
		return "$"
	}

	var sb strings.Builder
	sb.WriteString("$")

	start := 0
	end := len(segments)
	if opt.excludeTop {
		if len(segments) == 1 {
			return "$"
		}
		end = len(segments) - 1
	}

	for i := start; i < end; i++ {
		seg := segments[i]
		switch seg.Kind {
		case SegField:
			sb.WriteString(".")
			sb.WriteString(seg.Value)
		case SegIndex:
			if !(opt.excludeTopIndex && i == len(segments)-1) {
				sb.WriteString("[")
				sb.WriteString(seg.Value)
				sb.WriteString("]")
			}
		}
	}

	return sb.String()
}

// SegmentKind 表示路径段的类型
type SegmentKind int

const (
	// SegField 对象字段，如 .field
	SegField SegmentKind = iota
	// SegIndex 数组索引，如 [0]
	SegIndex
	// SegWildcard 通配符，如 [*]
	SegWildcard
)

// PathSegment 表示路径的一个段
type PathSegment struct {
	Kind  SegmentKind // 段的类型
	Value string      // 段的值（字段名或索引字符串）
}

// PathPattern 表示编译后的路径模式
type PathPattern struct {
	Segments []PathSegment // 路径段数组
}

func parseFieldSegment(remaining string) (PathSegment, string, error) {
	remaining = remaining[1:]
	end := len(remaining)
	for i, r := range remaining {
		if r == '.' || r == '[' {
			end = i
			break
		}
	}
	if end == 0 {
		return PathSegment{}, "", fmt.Errorf("%w: empty field name", ErrInvalidPattern)
	}
	fieldName := remaining[:end]
	return PathSegment{
		Kind:  SegField,
		Value: fieldName,
	}, remaining[end:], nil
}

func parseArraySegment(remaining string) (PathSegment, string, error) {
	remaining = remaining[1:]
	closeIdx := strings.Index(remaining, "]")
	if closeIdx == -1 {
		return PathSegment{}, "", fmt.Errorf("%w: missing closing ]", ErrInvalidPattern)
	}
	indexStr := remaining[:closeIdx]
	remaining = remaining[closeIdx+1:]

	if indexStr == "*" {
		return PathSegment{Kind: SegWildcard}, remaining, nil
	}

	if _, err := strconv.Atoi(indexStr); err != nil {
		return PathSegment{}, "", fmt.Errorf("%w: invalid array index: %s", ErrInvalidPattern, indexStr)
	}
	return PathSegment{Kind: SegIndex, Value: indexStr}, remaining, nil
}

// CompilePattern 编译路径模式表达式为 PathPattern
func CompilePattern(expr string) (PathPattern, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return PathPattern{}, fmt.Errorf("%w: empty pattern", ErrInvalidPattern)
	}

	if !strings.HasPrefix(expr, "$") {
		return PathPattern{}, fmt.Errorf("%w: pattern must start with $", ErrInvalidPattern)
	}

	if expr == "$" {
		return PathPattern{Segments: []PathSegment{}}, nil
	}

	var segments []PathSegment
	remaining := expr[1:]

	for len(remaining) > 0 {
		remaining = strings.TrimSpace(remaining)

		if strings.HasPrefix(remaining, ".") {
			seg, rest, err := parseFieldSegment(remaining)
			if err != nil {
				return PathPattern{}, err
			}
			segments = append(segments, seg)
			remaining = rest
			continue
		}

		if strings.HasPrefix(remaining, "[") {
			seg, rest, err := parseArraySegment(remaining)
			if err != nil {
				return PathPattern{}, err
			}
			segments = append(segments, seg)
			remaining = rest
			continue
		}

		return PathPattern{}, fmt.Errorf("%w: unexpected character at: %s", ErrInvalidPattern, remaining)
	}

	return PathPattern{Segments: segments}, nil
}

func (p *Parser) updateCachedSegments() {
	segments := p.cachedSegments[:0]
	for _, f := range p.stack {
		switch f.kind {
		case frameObject:
			if f.key != "" {
				segments = append(segments, PathSegment{
					Kind:  SegField,
					Value: f.key,
				})
			}
		case frameArray:
			segments = append(segments, PathSegment{
				Kind:  SegIndex,
				Value: strconv.Itoa(f.index),
			})
		}
	}
	p.cachedSegments = segments
}
