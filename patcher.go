package patcher

import "reflect"

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

type reflectPatcher struct {
	cache map[string]reflect.Value
}

func (r *reflectPatcher) PatchIt(it interface{}, patches map[string]interface{}) []string {
	var effectPath []string
	return effectPath
}
