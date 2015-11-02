package patcher

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Patcher use to Patch values of a struct
type Patcher interface {
	PatchIt(it interface{}, patches map[string]interface{}) []string
}

// NewPatcher use to create new patcher
func NewPatcher() Patcher {
	return &reflectPatcher{
		cache: make(map[string]reflect.Value),
	}
}

var NumberType reflect.Type = reflect.TypeOf(json.Number(""))

type reflectPatcher struct {
	cache map[string]reflect.Value
}

func (p *reflectPatcher) tokenize(path string) []string {
	var toks []string
	for _, tok := range strings.Split(path, ".") {
		toks = append(toks, upperFirst(tok))
	}
	return toks
}

func (p *reflectPatcher) PatchIt(target interface{}, patches map[string]interface{}) []string {
	var patchedSegs []string
	targetValue := reflect.ValueOf(target)
	for path, value := range patches {
		tokens := p.tokenize(path)
		patchedSegs = append(patchedSegs, tokens[0])
		v, vname := p.patchRecursive("", targetValue, tokens, "", value)
		if v == nil {
			panic(fmt.Sprintf("Can found field %v to patch", tokens))
		}
		if !v.CanSet() {
			panic("12121")
		}
		valueType := reflect.TypeOf(value)
		if valueType == NumberType {
			nv := value.(json.Number)
			switch v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				n, err := strconv.ParseInt(string(nv), 10, 64)
				if err != nil || v.OverflowInt(n) {
					panic(fmt.Sprintf("field: %s, number %v as %s patch failure err: %v", vname, value, v.Type(), err))
				}
				v.SetInt(n)
				continue

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				n, err := strconv.ParseUint(string(nv), 10, 64)
				if err != nil || v.OverflowUint(n) {
					panic(fmt.Sprintf("field: %v, number %v as %s patch failure err: %v", vname, value, v.Type(), err))
				}
				v.SetUint(n)
				continue

			case reflect.Float32, reflect.Float64:
				n, err := strconv.ParseFloat(string(nv), v.Type().Bits())
				if err != nil || v.OverflowFloat(n) {
					panic(fmt.Sprintf("field: %v, number %v as %s patch failure err: %v", value, v.Type(), err))
				}
				v.SetFloat(n)
				continue
			default:
				panic(fmt.Sprintf("field %s use value %v to patch %s type", vname, value, v.Kind()))
			}
		} else {
			vv := reflect.ValueOf(value)
			if v.Type() != vv.Type() {
				panic(fmt.Sprintf("field %s use value %v to patch %s type", vname, vv.Type(), v.Type()))
			}
			v.Set(vv)
		}
	}
	return patchedSegs
}

func (p *reflectPatcher) patchRecursive(fieldName string, targetValue reflect.Value, tokens []string, path string, value interface{}) (*reflect.Value, string) {
	switch targetValue.Kind() {
	case reflect.Ptr:
		originalValue := targetValue.Elem()
		return p.patchRecursive(fieldName, originalValue, tokens, path, value)
	case reflect.Interface:
		originalValue := targetValue.Elem()
		return p.patchRecursive(fieldName, originalValue, tokens, path, value)
	case reflect.Struct:
		typeOfT := targetValue.Type()
		for i := 0; i < targetValue.NumField(); i += 1 {
			currentTok := tokens[0]
			if typeOfT.Field(i).Name == currentTok {
				path = path + "." + currentTok
				return p.patchRecursive(typeOfT.Field(i).Name, targetValue.Field(i), tokens[1:], path, value)
			}
		}
	case reflect.Array, reflect.Slice:
		return nil, ""
	default:
		if targetValue.IsValid() {
			return &targetValue, fieldName
		}
		return nil, ""
	}
	return nil, ""
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}
