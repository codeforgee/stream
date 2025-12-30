package stream

import (
	"strings"
)

// valueKind 表示当前正在构建的值类型
type valueKind int

const (
	// valNone 无值
	valNone valueKind = iota
	// valString 字符串值
	valString
	// valNumber 数字值
	valNumber
)

// Parser 将 token 流转换为结构事件
type Parser struct {
	state          parserState     // 当前状态
	stack          []frame         // 结构帧栈
	curValueKind   valueKind       // 当前值的类型
	curString      strings.Builder // string 临时拼装
	curNumber      strings.Builder // number 临时拼装
	closing        bool            // 是否正在关闭
	chunkBuffer    strings.Builder // 字符串 chunk 缓冲区
	subs           []*Subscription // 订阅列表
	tokenizer      *Tokenizer      // tokenizer 实例
	err            error           // 解析过程中的错误
	logger         DebugLogger     // 调试日志记录器
	debugLevel     DebugLevel      // 调试级别
	cachedSegments []PathSegment   // 缓存的路径段数组
	segmentsDirty  bool            // 标记 segments 是否需要重新计算
	lastValueKind  ValueKind       // 当前值的类型
}

// NewParser 创建一个新的 Parser
func NewParser() *Parser {
	p := &Parser{
		state: pIdle,
		subs:  make([]*Subscription, 0, 8),
		stack: make([]frame, 0, 32),
	}

	p.tokenizer = NewTokenizer(func(tok Token) {
		p.OnToken(tok)
	})

	return p
}

// On 订阅指定路径的事件
func (p *Parser) On(expr string, h Handler) *Parser {
	pat, err := CompilePattern(expr)
	if err != nil {
		panic(err)
	}
	p.subs = append(p.subs, &Subscription{
		Pattern: pat,
		Handler: h,
	})
	return p
}

func (p *Parser) emit(ev Event) {
	if p.segmentsDirty || len(p.cachedSegments) != len(p.stack) {
		p.updateCachedSegments()
		p.segmentsDirty = false
	}

	segments := p.cachedSegments

	if len(segments) == 0 {
		ev.pathSegments = nil
	} else {
		ev.pathSegments = append([]PathSegment(nil), segments...)
	}

	p.debugLogEvent(ev, func() map[string]any {
		return map[string]any{
			"subs_count":  len(p.subs),
			"stack_depth": len(p.stack),
		}
	})

	for _, sub := range p.subs {
		if match(sub.Pattern.Segments, segments) {
			sub.Handler(ev)
		}
	}
}

func (p *Parser) topFrame() *frame {
	if len(p.stack) == 0 {
		return nil
	}
	return &p.stack[len(p.stack)-1]
}

func (p *Parser) parentFrame() *frame {
	if len(p.stack) < 2 {
		return nil
	}
	return &p.stack[len(p.stack)-2]
}

// OnToken 处理一个 token
func (p *Parser) OnToken(tok Token) {
	p.debugLogToken(tok)
	switch tok.Type {
	case TokenLBrace:
		p.onObjectStart()
	case TokenRBrace:
		p.onObjectEnd()
	case TokenLBracket:
		p.onArrayStart()
	case TokenRBracket:
		p.onArrayEnd()
	case TokenStringChunk:
		p.onStringChunk(tok.Value)
	case TokenStringEnd:
		p.onStringEnd()
	case TokenNumberChunk:
		p.onNumberChunk(tok.Value)
	case TokenNumberEnd:
		p.onNumberEnd()
	case TokenBool:
		p.onPrimitive(tok.Bool)
	case TokenNull:
		p.onPrimitive(nil)
	case TokenColon:
		p.onColon()
	case TokenComma:
		p.onComma()
	}
}

func (p *Parser) onObjectStart() {
	oldState := p.state
	p.stack = append(p.stack, frame{kind: frameObject})
	p.segmentsDirty = true
	p.state = pObjExpectKey
	p.debugLogStateChange(oldState, p.state, func() map[string]any {
		return map[string]any{
			"action":      "object_start",
			"stack_depth": len(p.stack),
		}
	})
	p.debugLogStack()
	p.emit(Event{
		Type:     EventObjectStart,
		pathOpts: pathOptions{},
	})
}

func (p *Parser) onObjectEnd() {
	top := p.topFrame()
	if top == nil {
		p.err = ErrMismatchedBrace
		p.debugLogError(p.err, func() map[string]any {
			return map[string]any{
				"action":      "object_end",
				"stack_empty": true,
			}
		})
		return
	}
	if top.kind != frameObject {
		p.err = ErrMismatchedBrace
		p.debugLogError(p.err, func() map[string]any {
			return map[string]any{
				"action":   "object_end",
				"expected": "frameObject",
				"got":      top.kind,
			}
		})
		return
	}

	p.emit(Event{
		Type:     EventObjectEnd,
		pathOpts: pathOptions{excludeTop: true},
	})

	parentFrame := p.parentFrame()

	oldState := p.state
	p.stack = p.stack[:len(p.stack)-1]
	p.segmentsDirty = true
	topAfterPop := p.topFrame()
	if topAfterPop == nil {
		p.state = pIdle
		p.debugLogStateChange(oldState, p.state, func() map[string]any {
			return map[string]any{
				"action": "object_end",
				"result": "idle",
			}
		})
		return
	}

	if parentFrame != nil && parentFrame.kind == frameArray {
		p.emit(Event{
			Type:     EventArrayItem,
			pathOpts: pathOptions{},
			Value: &PartialValue{
				Kind:     ValueObject,
				Complete: true,
			},
		})
		topAfterPop.index++
		p.segmentsDirty = true
		p.state = pArrAfterValue
		p.debugLogStateChange(oldState, p.state, func() map[string]any {
			return map[string]any{
				"action":      "object_end",
				"result":      "array_item",
				"array_index": topAfterPop.index,
			}
		})
		return
	}

	switch topAfterPop.kind {
	case frameArray:
		p.state = pArrAfterValue
	case frameObject:
		p.state = pObjAfterValue
	}
	p.debugLogStateChange(oldState, p.state, func() map[string]any {
		return map[string]any{
			"action":      "object_end",
			"parent_kind": topAfterPop.kind,
		}
	})
}

func (p *Parser) onArrayStart() {
	oldState := p.state
	p.stack = append(p.stack, frame{kind: frameArray})
	p.segmentsDirty = true
	p.state = pArrExpectValue
	p.debugLogStateChange(oldState, p.state, func() map[string]any {
		return map[string]any{
			"action":      "array_start",
			"stack_depth": len(p.stack),
		}
	})
	p.debugLogStack()
	p.emit(Event{
		Type:     EventArrayStart,
		pathOpts: pathOptions{excludeTopIndex: true},
	})
}

func (p *Parser) onArrayEnd() {
	top := p.topFrame()
	if top == nil {
		p.err = ErrMismatchedBracket
		p.debugLogError(p.err, func() map[string]any {
			return map[string]any{
				"action":      "array_end",
				"stack_empty": true,
			}
		})
		return
	}
	if top.kind != frameArray {
		p.err = ErrMismatchedBracket
		p.debugLogError(p.err, func() map[string]any {
			return map[string]any{
				"action":   "array_end",
				"expected": "frameArray",
				"got":      top.kind,
			}
		})
		return
	}

	p.emit(Event{
		Type:     EventArrayEnd,
		pathOpts: pathOptions{excludeTop: true},
	})

	parentFrame := p.parentFrame()

	oldState := p.state
	p.stack = p.stack[:len(p.stack)-1]
	p.segmentsDirty = true
	topAfterPop := p.topFrame()
	if topAfterPop == nil {
		p.state = pIdle
		p.debugLogStateChange(oldState, p.state, func() map[string]any {
			return map[string]any{
				"action": "array_end",
				"result": "idle",
			}
		})
		return
	}

	if parentFrame != nil && parentFrame.kind == frameArray {
		p.emit(Event{
			Type:     EventArrayItem,
			pathOpts: pathOptions{},
			Value: &PartialValue{
				Kind:     ValueArray,
				Complete: true,
			},
		})
		topAfterPop.index++
		p.segmentsDirty = true
		p.state = pArrAfterValue
		p.debugLogStateChange(oldState, p.state, func() map[string]any {
			return map[string]any{
				"action":      "array_end",
				"result":      "array_item",
				"array_index": topAfterPop.index,
			}
		})
		return
	}

	switch topAfterPop.kind {
	case frameArray:
		p.state = pArrAfterValue
	case frameObject:
		p.state = pObjAfterValue
	}
	p.debugLogStateChange(oldState, p.state, func() map[string]any {
		return map[string]any{
			"action":      "array_end",
			"parent_kind": topAfterPop.kind,
		}
	})
}

func (p *Parser) onStringChunk(s string) {
	p.curString.WriteString(s)

	if p.state != pObjExpectKey {
		p.curValueKind = valString
		p.chunkBuffer.WriteString(s)
	}
}

func (p *Parser) flushStringChunk() {
	if p.curValueKind != valString {
		return
	}

	chunk := p.chunkBuffer.String()
	if chunk == "" {
		return
	}

	p.emit(Event{
		Type:     EventFieldValue,
		pathOpts: pathOptions{},
		Value: &PartialValue{
			Kind:   ValueString,
			Value:  chunk,
			Append: true,
		},
	})

	p.chunkBuffer.Reset()
}

func (p *Parser) onStringEnd() {
	switch p.state {
	case pObjExpectKey:
		top := p.topFrame()
		if top == nil {
			p.err = ErrUnexpectedToken
			return
		}
		top.key = p.curString.String()
		p.segmentsDirty = true
		p.curString.Reset()
		p.state = pObjAfterKey
	case pObjExpectValue, pArrExpectValue:
		if p.chunkBuffer.Len() > 0 {
			bufferLen := p.chunkBuffer.Len()
			curStringLen := p.curString.Len()
			if curStringLen > bufferLen {
				p.flushStringChunk()
			}
		}

		if p.closing && p.curString.Len() == 0 {
			return
		}
		complete := !p.closing
		aborted := p.closing
		p.emit(Event{
			Type:     EventFieldValue,
			pathOpts: pathOptions{},
			Value: &PartialValue{
				Kind:     ValueString,
				Value:    p.curString.String(),
				Complete: complete,
				Aborted:  aborted,
			},
		})
		p.curString.Reset()
		p.chunkBuffer.Reset()
		p.curValueKind = valNone
		p.lastValueKind = ValueString
		if !p.closing {
			p.advanceAfterValue()
		}
	}
}

func (p *Parser) onNumberChunk(s string) {
	p.curValueKind = valNumber
	p.curNumber.WriteString(s)
}

func (p *Parser) onNumberEnd() {
	if p.closing && p.curNumber.Len() == 0 {
		return
	}
	val := p.curNumber.String()
	p.curNumber.Reset()
	p.curValueKind = valNone
	p.lastValueKind = ValueNumber
	complete := !p.closing
	aborted := p.closing
	p.emit(Event{
		Type:     EventFieldValue,
		pathOpts: pathOptions{},
		Value: &PartialValue{
			Kind:     ValueNumber,
			Value:    val,
			Complete: complete,
			Aborted:  aborted,
		},
	})
	if !p.closing {
		p.advanceAfterValue()
	}
}

func (p *Parser) onPrimitive(v any) {
	kind := ValueBool
	if v == nil {
		kind = ValueNull
	}
	p.lastValueKind = kind
	p.emit(Event{
		Type:     EventFieldValue,
		pathOpts: pathOptions{},
		Value: &PartialValue{
			Kind:     kind,
			Value:    v,
			Complete: true,
		},
	})
	p.advanceAfterValue()
}

func (p *Parser) onColon() {
	p.state = pObjExpectValue
}

func (p *Parser) onComma() {
	top := p.topFrame()
	if top == nil {
		p.err = ErrUnexpectedToken
		p.debugLogError(p.err, func() map[string]any {
			return map[string]any{
				"action":      "comma",
				"stack_empty": true,
			}
		})
		return
	}
	oldState := p.state
	switch top.kind {
	case frameObject:
		p.state = pObjExpectKey
	case frameArray:
		p.state = pArrExpectValue
	}
	p.debugLogStateChange(oldState, p.state, func() map[string]any {
		return map[string]any{
			"action":     "comma",
			"frame_kind": top.kind,
		}
	})
}

func (p *Parser) advanceAfterValue() {
	top := p.topFrame()
	if top == nil {
		p.state = pIdle
		return
	}
	switch top.kind {
	case frameObject:
		p.state = pObjAfterValue
	case frameArray:
		p.emit(Event{
			Type:     EventArrayItem,
			pathOpts: pathOptions{},
			Value: &PartialValue{
				Kind:     p.lastValueKind,
				Complete: true,
			},
		})
		top.index++
		p.segmentsDirty = true
		p.state = pArrAfterValue
	}
}

func (p *Parser) flushIncompleteValue() {
	if p.curValueKind == valString {
		p.flushStringChunk()
	}

	if p.curValueKind == valNone {
		return
	}

	var value *PartialValue
	switch p.curValueKind {
	case valString:
		if p.curString.Len() > 0 {
			value = &PartialValue{
				Kind:    ValueString,
				Value:   p.curString.String(),
				Aborted: true,
			}
			p.curString.Reset()
			p.chunkBuffer.Reset()
		}
	case valNumber:
		if p.curNumber.Len() > 0 {
			value = &PartialValue{
				Kind:    ValueNumber,
				Value:   p.curNumber.String(),
				Aborted: true,
			}
			p.curNumber.Reset()
		}
	}

	if value != nil {
		p.emit(Event{
			Type:     EventFieldValue,
			pathOpts: pathOptions{},
			Value:    value,
		})
	}

	p.curValueKind = valNone
}

func (p *Parser) closeUnfinishedFrames() {
	for top := p.topFrame(); top != nil; top = p.topFrame() {
		switch top.kind {
		case frameObject:
			p.emit(Event{
				Type:     EventObjectEnd,
				pathOpts: pathOptions{excludeTop: true},
			})
		case frameArray:
			p.emit(Event{
				Type:     EventArrayEnd,
				pathOpts: pathOptions{excludeTop: true},
			})
		}
		p.stack = p.stack[:len(p.stack)-1]
	}
}

// Close 关闭 parser，处理未完成的状态
func (p *Parser) Close(normal bool) error {
	if p.tokenizer != nil {
		p.tokenizer.Close()
	}

	p.closing = true

	p.flushIncompleteValue()

	p.closeUnfinishedFrames()

	if normal {
		p.emit(Event{
			Type:     EventStreamEnd,
			pathOpts: pathOptions{},
		})
	} else {
		p.emit(Event{
			Type:     EventStreamAbort,
			pathOpts: pathOptions{},
		})
	}

	return nil
}

func (p *Parser) checkState() error {
	if p.tokenizer == nil {
		return ErrInvalidState
	}
	return p.err
}

// Feed 输入字节数据
func (p *Parser) Feed(data []byte) error {
	return p.FeedString(string(data))
}

// FeedString 输入字符串数据
func (p *Parser) FeedString(s string) error {
	if err := p.checkState(); err != nil {
		return err
	}
	for _, r := range s {
		p.tokenizer.Consume(r)
		if p.err != nil {
			return p.err
		}
	}
	p.flushStringChunk()
	return nil
}

// Err 返回解析过程中的错误
func (p *Parser) Err() error {
	return p.err
}
