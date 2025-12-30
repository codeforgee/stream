package stream

import (
	"fmt"
	"strconv"
)

// ValueKind 表示值的类型
type ValueKind int

const (
	// ValueString 字符串类型
	ValueString ValueKind = iota
	// ValueNumber 数字类型
	ValueNumber
	// ValueBool 布尔类型
	ValueBool
	// ValueNull null 类型
	ValueNull
	// ValueObject 对象类型
	ValueObject
	// ValueArray 数组类型
	ValueArray
)

// String 返回值类型的字符串表示
func (vk ValueKind) String() string {
	switch vk {
	case ValueString:
		return "String"
	case ValueNumber:
		return "Number"
	case ValueBool:
		return "Bool"
	case ValueNull:
		return "Null"
	case ValueObject:
		return "Object"
	case ValueArray:
		return "Array"
	default:
		return fmt.Sprintf("ValueKind(%d)", vk)
	}
}

// PartialValue 表示一个部分值，支持流式追加和完成标记
type PartialValue struct {
	Kind     ValueKind // 值类型
	Value    any       // 值内容
	Append   bool      // 是否为追加模式
	Complete bool      // 是否完成
	Aborted  bool      // 是否中断
}

// String 转换为字符串
func (pv *PartialValue) String() string {
	if pv == nil || pv.Value == nil {
		return ""
	}
	if s, ok := pv.Value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", pv.Value)
}

func (pv *PartialValue) convertToInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int64:
		return val, true
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case uint:
		return int64(val), true
	case uint8:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint64:
		return int64(val), true
	case float32:
		return int64(val), true
	case float64:
		return int64(val), true
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return int64(f), true
		}
	}
	return 0, false
}

func (pv *PartialValue) convertToFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0.0, false
}

// Int 转换为 int
func (pv *PartialValue) Int() int {
	if pv == nil || pv.Value == nil {
		return 0
	}
	if i64, ok := pv.convertToInt64(pv.Value); ok {
		return int(i64)
	}
	return 0
}

// Int64 转换为 int64
func (pv *PartialValue) Int64() int64 {
	if pv == nil || pv.Value == nil {
		return 0
	}
	if i64, ok := pv.convertToInt64(pv.Value); ok {
		return i64
	}
	return 0
}

// Float64 转换为 float64
func (pv *PartialValue) Float64() float64 {
	if pv == nil || pv.Value == nil {
		return 0.0
	}
	if f64, ok := pv.convertToFloat64(pv.Value); ok {
		return f64
	}
	return 0.0
}

// Bool 转换为 bool
func (pv *PartialValue) Bool() bool {
	if pv == nil || pv.Value == nil {
		return false
	}
	switch v := pv.Value.(type) {
	case bool:
		return v
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
		switch v {
		case "true", "True", "TRUE":
			return true
		case "false", "False", "FALSE":
			return false
		}
		if i, err := strconv.Atoi(v); err == nil {
			return i != 0
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f != 0.0
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return pv.Int() != 0
	case float32, float64:
		return pv.Float64() != 0.0
	}
	return false
}

// IsNull 判断是否为 null
func (pv *PartialValue) IsNull() bool {
	return pv == nil || pv.Value == nil || pv.Kind == ValueNull
}

// IsEmpty 判断是否为空
func (pv *PartialValue) IsEmpty() bool {
	if pv == nil || pv.Value == nil {
		return true
	}
	switch pv.Kind {
	case ValueNull:
		return true
	case ValueString:
		return pv.String() == ""
	case ValueNumber:
		return pv.Float64() == 0.0
	case ValueBool:
		return !pv.Bool()
	default:
		return false
	}
}
