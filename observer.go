package stream

// ParserObserver 是 Parser 的观察者接口，用于监听解析过程中的内部事件
type ParserObserver interface {
	// OnToken 当 Parser 处理 token 时调用
	OnToken(token Token, state parserState, tokenizerState tokenizerState)

	// OnStateChange 当 Parser 状态变化时调用
	// context 是延迟求值的函数，只有在需要时才会执行（避免非 debug 模式下的开销）
	OnStateChange(oldState, newState parserState, context DebugContext)

	// OnEvent 当 Parser 发出事件时调用
	// context 是延迟求值的函数，只有在需要时才会执行
	OnEvent(event Event, context DebugContext)

	// OnError 当 Parser 遇到错误时调用
	// context 是延迟求值的函数，只有在需要时才会执行
	OnError(err error, context DebugContext)

	// OnStackChange 当 Parser 堆栈变化时调用（可选，用于详细调试）
	OnStackChange(stack []frame, path func() string)
}

// noopObserver 空观察者，用于默认情况（零开销）
type noopObserver struct{}

func (n *noopObserver) OnToken(Token, parserState, tokenizerState)           {}
func (n *noopObserver) OnStateChange(parserState, parserState, DebugContext) {}
func (n *noopObserver) OnEvent(Event, DebugContext)                          {}
func (n *noopObserver) OnError(error, DebugContext)                          {}
func (n *noopObserver) OnStackChange([]frame, func() string)                 {}

var defaultObserver ParserObserver = &noopObserver{}

// debugObserver 将 ParserObserver 事件转换为 DebugLogger 日志
type debugObserver struct {
	logger     DebugLogger
	debugLevel DebugLevel
}

// NewDebugObserver 创建一个 Debug 观察者
func NewDebugObserver(config *DebugConfig) ParserObserver {
	logger := NewDebugLogger(config)
	return &debugObserver{
		logger:     logger,
		debugLevel: config.Level,
	}
}

func (d *debugObserver) OnToken(token Token, state parserState, tokenizerState tokenizerState) {
	if d.logger == nil || d.debugLevel == DebugLevelNone {
		return
	}
	level := DebugLevelVerbose
	switch token.Type {
	case TokenLBrace, TokenRBrace, TokenLBracket, TokenRBracket:
		level = DebugLevelInfo
	}
	context := func() map[string]any {
		return map[string]any{
			"state":           state.String(),
			"tokenizer_state": tokenizerState.String(),
		}
	}
	d.logger.LogToken(level, token, context)
}

func (d *debugObserver) OnStateChange(oldState, newState parserState, context DebugContext) {
	if d.logger == nil || d.debugLevel == DebugLevelNone {
		return
	}
	level := DebugLevelInfo
	if oldState == newState {
		level = DebugLevelVerbose
	}
	d.logger.LogState(level, oldState.String(), newState.String(), context)
}

func (d *debugObserver) OnEvent(event Event, context DebugContext) {
	if d.logger == nil || d.debugLevel == DebugLevelNone {
		return
	}
	level := DebugLevelInfo
	if event.Type == EventStreamAbort {
		level = DebugLevelError
	}
	d.logger.LogEvent(level, event, context)
}

func (d *debugObserver) OnError(err error, context DebugContext) {
	if d.logger == nil || d.debugLevel == DebugLevelNone {
		return
	}
	d.logger.LogError(DebugLevelError, err, context)
}

func (d *debugObserver) OnStackChange(stack []frame, path func() string) {
	if d.logger == nil || d.debugLevel < DebugLevelVerbose {
		return
	}
	d.logger.LogStack(DebugLevelVerbose, stack, path())
}
