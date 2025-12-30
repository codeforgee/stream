package stream

// Handler 是事件处理函数类型
type Handler func(Event)

// Subscription 表示一个订阅
type Subscription struct {
	Pattern PathPattern // 编译后的路径模式
	Handler Handler     // 事件处理函数
}

func match(pattern []PathSegment, path []PathSegment) bool {
	if len(pattern) != len(path) {
		return false
	}

	for i := 0; i < len(pattern); i++ {
		p := pattern[i]
		s := path[i]

		switch p.Kind {
		case SegWildcard:
			continue
		case SegField:
			if s.Kind != SegField || p.Value != s.Value {
				return false
			}
		case SegIndex:
			if s.Kind != SegIndex || p.Value != s.Value {
				return false
			}
		}
	}

	return true
}
