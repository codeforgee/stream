package stream

// EnableDebug 使用配置为 Parser 启用调试功能
func (p *Parser) EnableDebug(config *DebugConfig) DebugLogger {
	logger := NewDebugLogger(config)
	p.logger = logger
	p.debugLevel = config.Level
	return logger
}

// DisableDebug 禁用调试功能
func (p *Parser) DisableDebug() {
	p.logger = nil
	p.debugLevel = DebugLevelNone
}

// IsDebugEnabled 检查是否启用了调试功能
func (p *Parser) IsDebugEnabled() bool {
	return p.logger != nil && p.debugLevel != DebugLevelNone
}

func (p *Parser) debugLogToken(token Token) {
	if !p.IsDebugEnabled() {
		return
	}
	level := DebugLevelVerbose
	switch token.Type {
	case TokenLBrace, TokenRBrace, TokenLBracket, TokenRBracket:
		level = DebugLevelInfo
	}
	context := func() map[string]any {
		return map[string]any{
			"state":           p.state.String(),
			"tokenizer_state": p.tokenizer.state.String(),
		}
	}
	p.logger.LogToken(level, token, context)
}

func (p *Parser) debugLogStateChange(oldState, newState parserState, context DebugContext) {
	if !p.IsDebugEnabled() {
		return
	}
	level := DebugLevelInfo
	if oldState == newState {
		level = DebugLevelVerbose
	}
	p.logger.LogState(level, oldState.String(), newState.String(), context)
}

func (p *Parser) debugLogEvent(event Event, context DebugContext) {
	if !p.IsDebugEnabled() {
		return
	}
	level := DebugLevelInfo
	if event.Type == EventStreamAbort {
		level = DebugLevelError
	}
	p.logger.LogEvent(level, event, context)
}

func (p *Parser) debugLogError(err error, context DebugContext) {
	if !p.IsDebugEnabled() {
		return
	}
	p.logger.LogError(DebugLevelError, err, context)
}

func (p *Parser) debugLogStack() {
	if !p.IsDebugEnabled() || p.debugLevel < DebugLevelVerbose {
		return
	}
	p.updateCachedSegments()
	path := buildPathFromSegments(p.cachedSegments, pathOptions{})
	p.logger.LogStack(DebugLevelVerbose, p.stack, path)
}
