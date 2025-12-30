package stream

import (
	"testing"
)

func TestPartialValue_String(t *testing.T) {
	tests := []struct {
		name  string
		value *PartialValue
		want  string
	}{
		{
			name:  "nil value",
			value: nil,
			want:  "",
		},
		{
			name: "string value",
			value: &PartialValue{
				Kind:  ValueString,
				Value: "hello",
			},
			want: "hello",
		},
		{
			name: "number to string",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "42",
			},
			want: "42",
		},
		{
			name: "int to string",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: 42,
			},
			want: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartialValue_Int(t *testing.T) {
	tests := []struct {
		name  string
		value *PartialValue
		want  int
	}{
		{
			name:  "nil value",
			value: nil,
			want:  0,
		},
		{
			name: "string number",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "42",
			},
			want: 42,
		},
		{
			name: "int value",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: 42,
			},
			want: 42,
		},
		{
			name: "float to int",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: 42.7,
			},
			want: 42,
		},
		{
			name: "invalid string",
			value: &PartialValue{
				Kind:  ValueString,
				Value: "not a number",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.Int()
			if got != tt.want {
				t.Errorf("Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartialValue_Int64(t *testing.T) {
	tests := []struct {
		name  string
		value *PartialValue
		want  int64
	}{
		{
			name: "string number",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "1234567890123",
			},
			want: 1234567890123,
		},
		{
			name: "int64 value",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: int64(1234567890123),
			},
			want: 1234567890123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.Int64()
			if got != tt.want {
				t.Errorf("Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartialValue_Float64(t *testing.T) {
	tests := []struct {
		name  string
		value *PartialValue
		want  float64
	}{
		{
			name: "string float",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "3.14",
			},
			want: 3.14,
		},
		{
			name: "float64 value",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: 3.14,
			},
			want: 3.14,
		},
		{
			name: "int to float",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: 42,
			},
			want: 42.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.Float64()
			if got != tt.want {
				t.Errorf("Float64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartialValue_Bool(t *testing.T) {
	tests := []struct {
		name  string
		value *PartialValue
		want  bool
	}{
		{
			name: "bool true",
			value: &PartialValue{
				Kind:  ValueBool,
				Value: true,
			},
			want: true,
		},
		{
			name: "bool false",
			value: &PartialValue{
				Kind:  ValueBool,
				Value: false,
			},
			want: false,
		},
		{
			name: "string true",
			value: &PartialValue{
				Kind:  ValueString,
				Value: "true",
			},
			want: true,
		},
		{
			name: "string false",
			value: &PartialValue{
				Kind:  ValueString,
				Value: "false",
			},
			want: false,
		},
		{
			name: "number non-zero",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "42",
			},
			want: true,
		},
		{
			name: "number zero",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "0",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.Bool()
			if got != tt.want {
				t.Errorf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartialValue_IsNull(t *testing.T) {
	tests := []struct {
		name  string
		value *PartialValue
		want  bool
	}{
		{
			name:  "nil value",
			value: nil,
			want:  true,
		},
		{
			name: "null kind",
			value: &PartialValue{
				Kind:  ValueNull,
				Value: nil,
			},
			want: true,
		},
		{
			name: "not null",
			value: &PartialValue{
				Kind:  ValueString,
				Value: "hello",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.IsNull()
			if got != tt.want {
				t.Errorf("IsNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartialValue_IsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value *PartialValue
		want  bool
	}{
		{
			name:  "nil value",
			value: nil,
			want:  true,
		},
		{
			name: "empty string",
			value: &PartialValue{
				Kind:  ValueString,
				Value: "",
			},
			want: true,
		},
		{
			name: "non-empty string",
			value: &PartialValue{
				Kind:  ValueString,
				Value: "hello",
			},
			want: false,
		},
		{
			name: "zero number",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "0",
			},
			want: true,
		},
		{
			name: "non-zero number",
			value: &PartialValue{
				Kind:  ValueNumber,
				Value: "42",
			},
			want: false,
		},
		{
			name: "false bool",
			value: &PartialValue{
				Kind:  ValueBool,
				Value: false,
			},
			want: true,
		},
		{
			name: "true bool",
			value: &PartialValue{
				Kind:  ValueBool,
				Value: true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.IsEmpty()
			if got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
