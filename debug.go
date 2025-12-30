package stream

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func init() {
	if defaultWriter == nil {
		defaultWriter = os.Stderr
	}
}

// DebugLevel 表示调试级别
type DebugLevel int

const (
	// DebugLevelNone 不输出调试信息
	DebugLevelNone DebugLevel = iota
	// DebugLevelError 只输出错误
	DebugLevelError
	// DebugLevelInfo 输出关键信息（状态变化、事件）
	DebugLevelInfo
	// DebugLevelVerbose 输出详细信息（token、状态细节）
	DebugLevelVerbose
)

// DebugContext 调试上下文，用于传递额外的调试信息
type DebugContext func() map[string]any

// DebugConfig 调试配置选项
type DebugConfig struct {
	Writer         io.Writer  // 日志输出目标，nil 时使用 os.Stderr
	Level          DebugLevel // 调试级别
	Prefix         string     // 日志前缀，默认为 "[DEBUG]"
	ShowTimestamp  bool       // 是否显示时间戳，默认为 true
	MaxValueLength int        // 值的最大显示长度，超过会被截断，0 表示不限制
}

// DefaultDebugConfig 返回默认的调试配置
func DefaultDebugConfig() *DebugConfig {
	return &DebugConfig{
		Writer:         nil, // 使用 os.Stderr
		Level:          DebugLevelInfo,
		Prefix:         "[DEBUG]",
		ShowTimestamp:  true,
		MaxValueLength: 30,
	}
}

// DebugLogger 调试日志记录器接口
type DebugLogger interface {
	LogToken(level DebugLevel, token Token, context DebugContext)
	LogState(level DebugLevel, oldState, newState string, context DebugContext)
	LogEvent(level DebugLevel, event Event, context DebugContext)
	LogError(level DebugLevel, err error, context DebugContext)
	LogStack(level DebugLevel, stack []frame, path string)
	LogMessage(level DebugLevel, message string, context DebugContext)
}

// defaultDebugLogger 默认的调试日志记录器（输出到 io.Writer）
type defaultDebugLogger struct {
	writer         io.Writer  // 输出目标
	level          DebugLevel // 调试级别
	prefix         string     // 日志前缀
	showTimestamp  bool       // 是否显示时间戳
	maxValueLength int        // 值的最大显示长度
}

var defaultWriter io.Writer = os.Stderr

// NewDebugLogger 使用配置创建调试日志记录器
func NewDebugLogger(config *DebugConfig) DebugLogger {
	writer := config.Writer
	if writer == nil {
		writer = defaultWriter
	}
	return &defaultDebugLogger{
		writer:         writer,
		level:          config.Level,
		prefix:         config.Prefix,
		showTimestamp:  config.ShowTimestamp,
		maxValueLength: config.MaxValueLength,
	}
}

func (d *defaultDebugLogger) shouldLog(level DebugLevel) bool {
	return level <= d.level && d.level != DebugLevelNone
}

func (d *defaultDebugLogger) log(level DebugLevel, format string, args ...any) {
	if !d.shouldLog(level) {
		return
	}
	levelStr := d.levelString(level)
	message := fmt.Sprintf(format, args...)

	var parts []string
	if d.showTimestamp {
		timestamp := time.Now().Format("15:04:05.000")
		parts = append(parts, timestamp)
	}
	parts = append(parts, fmt.Sprintf("[%s]", levelStr))
	if d.prefix != "" {
		parts = append(parts, d.prefix)
	}
	parts = append(parts, message)

	fmt.Fprintf(d.writer, "%s\n", strings.Join(parts, " "))
}

func (d *defaultDebugLogger) levelString(level DebugLevel) string {
	switch level {
	case DebugLevelError:
		return "ERROR"
	case DebugLevelInfo:
		return "INFO"
	case DebugLevelVerbose:
		return "VERBOSE"
	default:
		return "UNKNOWN"
	}
}

func (d *defaultDebugLogger) LogToken(level DebugLevel, token Token, context DebugContext) {
	if !d.shouldLog(level) {
		return
	}
	tokenStr := d.formatToken(token)
	ctxStr := d.formatContext(context)
	d.log(level, "TOKEN: %s%s", tokenStr, ctxStr)
}

func (d *defaultDebugLogger) LogState(level DebugLevel, oldState, newState string, context DebugContext) {
	if !d.shouldLog(level) {
		return
	}
	ctxStr := d.formatContext(context)
	if oldState != newState {
		d.log(level, "STATE: %s -> %s%s", oldState, newState, ctxStr)
	} else {
		d.log(level, "STATE: %s%s", newState, ctxStr)
	}
}

func (d *defaultDebugLogger) LogEvent(level DebugLevel, event Event, context DebugContext) {
	if !d.shouldLog(level) {
		return
	}
	eventStr := d.formatEvent(event)
	ctxStr := d.formatContext(context)
	d.log(level, "EVENT: %s%s", eventStr, ctxStr)
}

func (d *defaultDebugLogger) LogError(level DebugLevel, err error, context DebugContext) {
	if !d.shouldLog(level) {
		return
	}
	ctxStr := d.formatContext(context)
	d.log(level, "ERROR: %v%s", err, ctxStr)
}

func (d *defaultDebugLogger) LogStack(level DebugLevel, stack []frame, path string) {
	if !d.shouldLog(level) {
		return
	}
	stackStr := d.formatStack(stack)
	d.log(level, "STACK: depth=%d path=%s frames=[%s]", len(stack), path, stackStr)
}

func (d *defaultDebugLogger) LogMessage(level DebugLevel, message string, context DebugContext) {
	if !d.shouldLog(level) {
		return
	}
	ctxStr := d.formatContext(context)
	d.log(level, "MSG: %s%s", message, ctxStr)
}

func (d *defaultDebugLogger) formatToken(token Token) string {
	switch token.Type {
	case TokenStringChunk:
		value := token.Value
		if len(value) > 20 {
			value = value[:20] + "..."
		}
		return fmt.Sprintf("StringChunk(%q)", value)
	case TokenStringEnd:
		return "StringEnd"
	case TokenNumberChunk:
		return fmt.Sprintf("NumberChunk(%s)", token.Value)
	case TokenNumberEnd:
		return "NumberEnd"
	case TokenBool:
		return fmt.Sprintf("Bool(%v)", token.Bool)
	case TokenNull:
		return "Null"
	case TokenLBrace:
		return "LBrace {"
	case TokenRBrace:
		return "RBrace }"
	case TokenLBracket:
		return "LBracket ["
	case TokenRBracket:
		return "RBracket ]"
	case TokenColon:
		return "Colon :"
	case TokenComma:
		return "Comma ,"
	default:
		return fmt.Sprintf("Token(%d)", token.Type)
	}
}

func (d *defaultDebugLogger) formatEvent(event Event) string {
	var parts []string
	parts = append(parts, event.Type.String())
	if path := event.Path(); path != "" {
		parts = append(parts, fmt.Sprintf("path=%s", path))
	}
	if event.Value != nil {
		valueStr := d.formatPartialValue(event.Value)
		parts = append(parts, valueStr)
	}
	return strings.Join(parts, " | ")
}

func (d *defaultDebugLogger) formatPartialValue(pv *PartialValue) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("kind=%s", pv.Kind.String()))
	if pv.Value != nil {
		valueStr := fmt.Sprintf("%v", pv.Value)
		maxLen := d.maxValueLength
		if maxLen > 0 && len(valueStr) > maxLen {
			valueStr = valueStr[:maxLen] + "..."
		}
		parts = append(parts, fmt.Sprintf("value=%q", valueStr))
	}
	if pv.Append {
		parts = append(parts, "append=true")
	}
	if pv.Complete {
		parts = append(parts, "complete=true")
	}
	if pv.Aborted {
		parts = append(parts, "aborted=true")
	}
	return fmt.Sprintf("Value(%s)", strings.Join(parts, ", "))
}

func (d *defaultDebugLogger) formatStack(stack []frame) string {
	var parts []string
	for i, f := range stack {
		var frameStr string
		switch f.kind {
		case frameObject:
			if f.key != "" {
				frameStr = fmt.Sprintf("Object(key=%s)", f.key)
			} else {
				frameStr = "Object"
			}
		case frameArray:
			frameStr = fmt.Sprintf("Array(index=%d)", f.index)
		}
		parts = append(parts, fmt.Sprintf("%d:%s", i, frameStr))
	}
	return strings.Join(parts, " ")
}

func (d *defaultDebugLogger) formatContext(context DebugContext) string {
	if context == nil {
		return ""
	}
	ctx := context()
	if len(ctx) == 0 {
		return ""
	}
	var parts []string
	for k, v := range ctx {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return " | " + strings.Join(parts, " ")
}
