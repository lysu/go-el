package el

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// NumberType patch api use this type to deserialize JSON request in Golang
var NumberType = reflect.TypeOf(json.Number(""))

type Value struct {
	val       reflect.Value
	keySetter *KeySetter
}

type KeySetter struct {
	prev *Value
	key  reflect.Value
}

func AsValue(i interface{}) *Value {
	return &Value{
		val: reflect.ValueOf(i),
	}
}

func AsValueWithSetter(i interface{}, keySetter *KeySetter) *Value {
	return &Value{
		val:       reflect.ValueOf(i),
		keySetter: keySetter,
	}
}

func (v *Value) getResolvedValue() reflect.Value {
	if v.val.IsValid() && v.val.Kind() == reflect.Ptr {
		return v.val.Elem()
	}
	return v.val
}

func (v *Value) IsKeySetter() bool {
	return v.keySetter != nil
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

func (v *Value) Contains(other *Value) bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Struct:
		fieldValue := v.getResolvedValue().FieldByName(other.String())
		return fieldValue.IsValid()
	case reflect.Map:
		var mapValue reflect.Value
		switch other.Interface().(type) {
		case int:
			mapValue = v.getResolvedValue().MapIndex(other.getResolvedValue())
		case string:
			mapValue = v.getResolvedValue().MapIndex(other.getResolvedValue())
		default:
			return false
		}

		return mapValue.IsValid()
	case reflect.String:
		return strings.Contains(v.getResolvedValue().String(), other.String())

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.getResolvedValue().Len(); i++ {
			item := v.getResolvedValue().Index(i)
			if other.Interface() == item.Interface() {
				return true
			}
		}
		return false

	default:
		return false
	}
}

func (v *Value) CanSlice() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return true
	}
	return false
}

func (v *Value) EqualValueTo(other *Value) bool {
	if v.IsInteger() && other.IsInteger() {
		return v.Integer() == other.Integer()
	}
	return v.Interface() == other.Interface()
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

func (v *Value) SetNumber(nv json.Number) error {
	resolvedValue := v.getResolvedValue()
	switch resolvedValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(string(nv), 10, 64)
		if err != nil || resolvedValue.OverflowInt(n) {
			return fmt.Errorf("Can not use number %v as %s patch failure err: %v", nv, resolvedValue.Type(), err)
		}
		resolvedValue.SetInt(n)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(string(nv), 10, 64)
		if err != nil || resolvedValue.OverflowUint(n) {
			return fmt.Errorf("Can not use number %v as %s patch failure err: %v", nv, resolvedValue.Type(), err)
		}
		resolvedValue.SetUint(n)
		return nil

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(string(nv), resolvedValue.Type().Bits())
		if err != nil || resolvedValue.OverflowFloat(n) {
			return fmt.Errorf("Can not use number %v as %s patch failure err: %v", resolvedValue, resolvedValue.Type(), err)
		}
		resolvedValue.SetFloat(n)
		return nil

	default:
		return fmt.Errorf("Can not use use value %v to patch %s type", resolvedValue, resolvedValue.Kind())
	}
	return nil
}

func (v *Value) ToRealNumber(nv json.Number, valueType reflect.Type) interface{} {
	switch valueType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(string(nv), 10, 64)
		if err != nil {
			return fmt.Errorf("Can not use number %v as %s patch failure err: %v", nv, valueType, err)
		}
		switch k := valueType.Kind(); k {
		default:
			panic(&reflect.ValueError{"Transform to int failure, err: %v", valueType.Kind()})
		case reflect.Int:
			return int(n)
		case reflect.Int8:
			return int8(n)
		case reflect.Int16:
			return int16(n)
		case reflect.Int32:
			return int32(n)
		case reflect.Int64:
			return n
		}
		return n

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(string(nv), 10, 64)
		if err != nil {
			return fmt.Errorf("Can not use number %v as %s patch failure err: %v", nv, valueType, err)
		}
		switch k := valueType.Kind(); k {
		default:
			panic(&reflect.ValueError{"Transform to uint failure, err: %v", valueType.Kind()})
		case reflect.Uint:
			return uint(n)
		case reflect.Uint8:
			return uint8(n)
		case reflect.Uint16:
			return uint16(n)
		case reflect.Uint32:
			return uint32(n)
		case reflect.Uint64:
			return n
		case reflect.Uintptr:
			return uintptr(n)
		}
		return n

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(string(nv), valueType.Bits())
		if err != nil {
			return fmt.Errorf("Can not use number %v as %s patch failure err: %v", valueType, valueType, err)
		}
		switch k := valueType.Kind(); k {
		default:
			panic(&reflect.ValueError{"Transform to float failure, err: %v", valueType.Kind()})
		case reflect.Float32:
			return int32(n)
		case reflect.Float64:
			return n
		}
		return n

	default:
		return fmt.Errorf("Can not use use value %v to patch %s type", valueType, valueType.Kind())
	}
	return nil
}

func (v *Value) SetValue(rightValue interface{}) error {

	rvType := reflect.TypeOf(rightValue)

	resolvedValue := v.getResolvedValue()

	if rvType == NumberType && !v.IsKeySetter() {
		nv := rightValue.(json.Number)
		return v.SetNumber(nv)
	}

	if v.IsKeySetter() {
		setter := v.keySetter
		target := setter.prev.getResolvedValue()
		switch target.Kind() {
		case reflect.Map:
			if rvType == NumberType {
				nv := rightValue.(json.Number)
				rightValue = v.ToRealNumber(nv, target.Type().Elem())
			}
			target.SetMapIndex(setter.key, reflect.ValueOf(rightValue))
			return nil
		}
	}

	if rvType != resolvedValue.Type() {
		return fmt.Errorf("Can not use use value %v to patch %s type", rvType, resolvedValue.Type())
	}
	if !resolvedValue.CanSet() {
		return fmt.Errorf("Var %#v is not settable", v.val)
	}
	resolvedValue.Set(reflect.ValueOf(rightValue))
	return nil
}
