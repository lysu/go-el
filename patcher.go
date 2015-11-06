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

// Patcher use to patch in memory struct with path
type Patcher struct {
	Cache map[string]reflect.Value
}

type Path string

func (p Path) Tokenize() []string {
	var toks []string
	for _, tok := range strings.Split(string(p), ".") {
		toks = append(toks, upperFirst(tok))
	}
	return toks
}

func (p Path) Segment() string {
	idx := strings.Index(string(p), ".")
	if idx == -1 {
		return ""
	}
	return upperFirst(string(p)[:idx])
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// NumberType avada patch api use this type to deserialize request
var NumberType = reflect.TypeOf(json.Number(""))

// DoPatch do patch job
// and generate segment update statstic
func (p *Patcher) DoPatch(target interface{}, paths map[Path]interface{}) ([]string, error) {

	targetValue := reflect.ValueOf(target)

	for path, value := range paths {

		fieldName, fieldValue, err := p.locateField(targetValue, path)
		if err != nil {
			return nil, err
		}

		err = p.setValue(fieldName, fieldValue, value)
		if err != nil {
			return nil, err
		}

	}

	return p.updatedSegments(p.Segments()), nil
}

func (p *Patcher) locateField(modified reflect.Value, path Path) (string, *reflect.Value, error) {
	tokens := path.Tokenize()
	fieldValue, fieldName := p.patchRecursive("", modified, tokens, "")
	if fieldValue == nil {
		return "", nil, fmt.Errorf("无法匹配的Patch表达式: %s", string(path))
	}
	if !fieldValue.CanSet() {
		return fieldName, fieldValue, fmt.Errorf("Patch表达式%s对应的字段%s不可写", string(path), fieldName)
	}
	return fieldName, fieldValue, nil
}

func (p *Patcher) setValue(fieldName string, fieldValue *reflect.Value, rightValue interface{}) error {

	rvType := reflect.TypeOf(rightValue)

	if rvType == NumberType {

		nv := rightValue.(json.Number)

		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(string(nv), 10, 64)
			if err != nil || fieldValue.OverflowInt(n) {
				return fmt.Errorf("field %s number %v as %s patch failure err: %v", fieldName, rightValue, fieldValue.Type(), err)
			}
			fieldValue.SetInt(n)
			return nil

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n, err := strconv.ParseUint(string(nv), 10, 64)
			if err != nil || fieldValue.OverflowUint(n) {
				return fmt.Errorf("field %s number %v as %s patch failure err: %v", fieldName, rightValue, fieldValue.Type(), err)
			}
			fieldValue.SetUint(n)
			return nil

		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(string(nv), fieldValue.Type().Bits())
			if err != nil || fieldValue.OverflowFloat(n) {
				return fmt.Errorf("field %s number %v as %s patch failure err: %v", fieldName, fieldValue, fieldValue.Type(), err)
			}
			fieldValue.SetFloat(n)
			return nil

		default:
			return fmt.Errorf("field %s use value %v to patch %s type", fieldName, fieldValue, fieldValue.Kind())
		}
	} else {
		if rvType != fieldValue.Type() {
			return fmt.Errorf("field %s use value %v to patch %s type", fieldName, rvType, fieldValue.Type())
		}
		fieldValue.Set(reflect.ValueOf(rightValue))
	}
	return nil
}

func (p *Patcher) updatedSegments(startPaths []string) []string {
	updatedSegments := make(segmentSet)
	for _, segment := range startPaths {
		switch segment {
		case "UserBasic":
			updatedSegments.put("UserBasic")
		case "Register":
			updatedSegments.put("UserRegister", "RegAudit")
		case "IdCard":
			updatedSegments.put("UserIdCard", "IdCardAudit")
		case "DriverLic":
			updatedSegments.put("UserDriverLic", "DriverLicAudit")
		case "UserQualification":
			updatedSegments.put("UserQualification", "QualificationAudit")
		case "RegAudit":
			updatedSegments.put("RegAudit")
		case "IdCardAudit":
			updatedSegments.put("IdCardAudit")
		case "DriverLicAudit":
			updatedSegments.put("DriverLicAudit")
		case "QualificationAudit":
			updatedSegments.put("QualificationAudit")
		case "Ownership":
			updatedSegments.put("Ownership")
		case "Vehicle":
			updatedSegments.put("Ownership", "VehicleBasic", "VehicleLicense", "CompulsoryInsurance", "VehicleInsurance") // maybe some opt on this.
		case "Image":
			updatedSegments.put(
				"Image",
				"UserRegister", "RegAudit", "UserIdCard", "IdCardAudit", "UserDriverLic", "DriverLicAudit", "UserQualification", "QualificationAudit",
				"Ownership", "VehicleBasic", "VehicleLicense", "CompulsoryInsurance", "VehicleInsurance",
			)
		}
	}
	return updatedSegments.asSlice()
}

func (p *Patcher) patchRecursive(fieldName string, targetValue reflect.Value, tokens []string, path string) (*reflect.Value, string) {
	switch targetValue.Kind() {
	case reflect.Ptr:
		originalValue := targetValue.Elem()
		return p.patchRecursive(fieldName, originalValue, tokens, path)
	case reflect.Interface:
		originalValue := targetValue.Elem()
		return p.patchRecursive(fieldName, originalValue, tokens, path)
	case reflect.Struct:
		typeOfT := targetValue.Type()
		for i := 0; i < targetValue.NumField(); i++ {
			currentTok := tokens[0]
			if typeOfT.Field(i).Name == currentTok {
				path = path + "." + currentTok
				return p.patchRecursive(typeOfT.Field(i).Name, targetValue.Field(i), tokens[1:], path)
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

var emptyStruct = struct{}{}

type segmentSet map[string]struct{}

func (s *segmentSet) put(segs ...string) {
	for _, seg := range segs {
		(*s)[seg] = emptyStruct
	}
}

func (s *segmentSet) asSlice() []string {
	var segs []string
	for seg := range *s {
		segs = append(segs, seg)
	}
	return segs
}
