package stream

import "fmt"

// EventType 表示事件类型
type EventType int

const (
	// EventObjectStart 对象开始
	EventObjectStart EventType = iota
	// EventObjectEnd 对象结束
	EventObjectEnd
	// EventArrayStart 数组开始
	EventArrayStart
	// EventArrayEnd 数组结束
	EventArrayEnd
	// EventFieldValue 字段值
	EventFieldValue
	// EventArrayItem 数组项
	EventArrayItem
	// EventStreamEnd 流正常结束
	EventStreamEnd
)

// String 返回事件类型的字符串表示
func (et EventType) String() string {
	switch et {
	case EventObjectStart:
		return "ObjectStart"
	case EventObjectEnd:
		return "ObjectEnd"
	case EventArrayStart:
		return "ArrayStart"
	case EventArrayEnd:
		return "ArrayEnd"
	case EventFieldValue:
		return "FieldValue"
	case EventArrayItem:
		return "ArrayItem"
	case EventStreamEnd:
		return "StreamEnd"
	default:
		return fmt.Sprintf("EventType(%d)", et)
	}
}

// Event 表示一个解析事件
type Event struct {
	Type         EventType     // 事件类型
	Value        *PartialValue // 部分值（可能为 nil）
	pathSegments []PathSegment // 路径段数组（用于延迟计算Path）
	pathOpts     pathOptions   // 路径计算选项
	pathCache    string        // 缓存的Path字符串（延迟计算）
}

// Path 获取路径字符串（延迟计算）
func (ev *Event) Path() string {
	if ev.pathCache == "" && len(ev.pathSegments) > 0 {
		ev.pathCache = buildPathFromSegments(ev.pathSegments, ev.pathOpts)
	}
	return ev.pathCache
}
