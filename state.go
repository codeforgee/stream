package stream

import "fmt"

// frameKind 表示结构帧的类型
type frameKind int

const (
	// frameObject 对象类型
	frameObject frameKind = iota
	// frameArray 数组类型
	frameArray
)

// frame 表示一个结构帧（用于维护解析上下文）
type frame struct {
	kind  frameKind // 帧类型（object 或 array）
	key   string    // object 当前字段名
	index int       // array 当前索引
}

// parserState 表示 parser 的状态
type parserState int

const (
	// pIdle 空闲状态
	pIdle parserState = iota
	// pObjExpectKey 对象中期望 key
	pObjExpectKey
	// pObjAfterKey key 之后，等待冒号
	pObjAfterKey
	// pObjExpectValue 对象中期望 value
	pObjExpectValue
	// pObjAfterValue value 之后，等待逗号或结束
	pObjAfterValue
	// pArrExpectValue 数组中期望 value
	pArrExpectValue
	// pArrAfterValue value 之后，等待逗号或结束
	pArrAfterValue
)

// String 返回 parser 状态的字符串表示
func (ps parserState) String() string {
	switch ps {
	case pIdle:
		return "Idle"
	case pObjExpectKey:
		return "ObjExpectKey"
	case pObjAfterKey:
		return "ObjAfterKey"
	case pObjExpectValue:
		return "ObjExpectValue"
	case pObjAfterValue:
		return "ObjAfterValue"
	case pArrExpectValue:
		return "ArrExpectValue"
	case pArrAfterValue:
		return "ArrAfterValue"
	default:
		return fmt.Sprintf("ParserState(%d)", ps)
	}
}
