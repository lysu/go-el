package patcher

import (
	"fmt"
	"reflect"
	"strconv"
)

type Value struct {
	val  reflect.Value
	safe bool // used to indicate whether a Value needs explicit escaping in the template
}

func AsValue(i interface{}) *Value {
	return &Value{
		val: reflect.ValueOf(i),
	}
}

func AsSafeValue(i interface{}) *Value {
	return &Value{
		val:  reflect.ValueOf(i),
		safe: true,
	}
}

func (v *Value) getResolvedValue() reflect.Value {
	if v.val.IsValid() && v.val.Kind() == reflect.Ptr {
		return v.val.Elem()
	}
	return v.val
}

func (v *Value) IsString() bool {
	return v.getResolvedValue().Kind() == reflect.String
}

func (v *Value) IsBool() bool {
	return v.getResolvedValue().Kind() == reflect.Bool
}

func (v *Value) IsFloat() bool {
	return v.getResolvedValue().Kind() == reflect.Float32 ||
		v.getResolvedValue().Kind() == reflect.Float64
}

func (v *Value) IsInteger() bool {
	return v.getResolvedValue().Kind() == reflect.Int ||
		v.getResolvedValue().Kind() == reflect.Int8 ||
		v.getResolvedValue().Kind() == reflect.Int16 ||
		v.getResolvedValue().Kind() == reflect.Int32 ||
		v.getResolvedValue().Kind() == reflect.Int64 ||
		v.getResolvedValue().Kind() == reflect.Uint ||
		v.getResolvedValue().Kind() == reflect.Uint8 ||
		v.getResolvedValue().Kind() == reflect.Uint16 ||
		v.getResolvedValue().Kind() == reflect.Uint32 ||
		v.getResolvedValue().Kind() == reflect.Uint64
}

func (v *Value) IsNumber() bool {
	return v.IsInteger() || v.IsFloat()
}

func (v *Value) IsNil() bool {
	return !v.getResolvedValue().IsValid()
}

func (v *Value) Integer() int {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(v.getResolvedValue().Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int(v.getResolvedValue().Uint())
	case reflect.Float32, reflect.Float64:
		return int(v.getResolvedValue().Float())
	case reflect.String:
		// Try to convert from string to int (base 10)
		f, err := strconv.ParseFloat(v.getResolvedValue().String(), 64)
		if err != nil {
			return 0
		}
		return int(f)
	default:
		return 0
	}
}

func (v *Value) Float() float64 {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.getResolvedValue().Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.getResolvedValue().Uint())
	case reflect.Float32, reflect.Float64:
		return v.getResolvedValue().Float()
	case reflect.String:
		// Try to convert from string to float64 (base 10)
		f, err := strconv.ParseFloat(v.getResolvedValue().String(), 64)
		if err != nil {
			return 0.0
		}
		return f
	default:
		return 0.0
	}
}

func (v *Value) Bool() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Bool:
		return v.getResolvedValue().Bool()
	default:
		return false
	}
}

func (v *Value) IsTrue() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.getResolvedValue().Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.getResolvedValue().Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.getResolvedValue().Float() != 0
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.getResolvedValue().Len() > 0
	case reflect.Bool:
		return v.getResolvedValue().Bool()
	case reflect.Struct:
		return true // struct instance is always true
	default:
		return false
	}
}

func (v *Value) Negate() *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Integer() != 0 {
			return AsValue(0)
		}
		return AsValue(1)
	case reflect.Float32, reflect.Float64:
		if v.Float() != 0.0 {
			return AsValue(float64(0.0))
		}
		return AsValue(float64(1.1))
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return AsValue(v.getResolvedValue().Len() == 0)
	case reflect.Bool:
		return AsValue(!v.getResolvedValue().Bool())
	default:
		return AsValue(true)
	}
}

func (v *Value) Len() int {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return v.getResolvedValue().Len()
	case reflect.String:
		runes := []rune(v.getResolvedValue().String())
		return len(runes)
	default:
		return 0
	}
}

func (v *Value) Slice(i, j int) *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice:
		return AsValue(v.getResolvedValue().Slice(i, j).Interface())
	case reflect.String:
		runes := []rune(v.getResolvedValue().String())
		return AsValue(string(runes[i:j]))
	default:
		return AsValue([]int{})
	}
}

func (v *Value) Index(i int) *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice:
		if i >= v.Len() {
			return AsValue(nil)
		}
		return AsValue(v.getResolvedValue().Index(i).Interface())
	case reflect.String:
		//return AsValue(v.getResolvedValue().Slice(i, i+1).Interface())
		s := v.getResolvedValue().String()
		runes := []rune(s)
		if i < len(runes) {
			return AsValue(string(runes[i]))
		}
		return AsValue("")
	default:
		return AsValue([]int{})
	}
}

func (v *Value) String() string {
	if v.IsNil() {
		return ""
	}

	switch v.getResolvedValue().Kind() {
	case reflect.String:
		return v.getResolvedValue().String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.getResolvedValue().Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.getResolvedValue().Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.getResolvedValue().Float())
	case reflect.Bool:
		if v.Bool() {
			return "True"
		}
		return "False"
	case reflect.Struct:
		if t, ok := v.Interface().(fmt.Stringer); ok {
			return t.String()
		}
	}

	return v.getResolvedValue().String()
}

func (v *Value) Interface() interface{} {
	if v.val.IsValid() {
		return v.val.Interface()
	}
	return nil
}
