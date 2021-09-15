package zeta

import (
	"reflect"
	"strings"
)

type StructRef struct {
	// A struct pointer
	Ref         interface{}
	refValue    reflect.Value
	structValue reflect.Value
	structType  reflect.Type
}

// NewStructRef structRef should be a struct point
func NewStructRef(structRef interface{}) *StructRef {
	r := &StructRef{Ref: structRef}
	r.refValue = reflect.ValueOf(structRef)
	r.structValue = reflect.Indirect(r.refValue)
	r.structType = r.structValue.Type()
	return r
}

func (s *StructRef) Create() *StructRef {
	refValue := reflect.New(s.structType)
	ref := refValue.Interface()
	r := &StructRef{
		Ref:         ref,
		refValue:    refValue,
		structValue: reflect.Indirect(refValue),
	}
	r.structType = r.structValue.Type()

	return r

}

func (s *StructRef) Set(name string, value interface{}) *StructRef {
	s.structValue.FieldByName(name).Set(reflect.ValueOf(value))
	return s
}

func (s *StructRef) SetIgnoreCase(name string, value interface{}) *StructRef {
	s.getFieldValueIgnoreCase(name).Set(reflect.ValueOf(value))
	return s
}
func (s *StructRef) getFieldValueIgnoreCase(name string) reflect.Value {
	return s.structValue.FieldByNameFunc(func(s string) bool { return strings.EqualFold(s, name) })
}

func (s *StructRef) Get(name string) interface{} {
	return s.structValue.FieldByName(name).Interface()
}

func (s *StructRef) GetIgnoreCase(name string) interface{} {
	return s.getFieldValueIgnoreCase(name).Interface()
}

func (s *StructRef) GetInto(name string, valueRef interface{}) {
	reflect.Indirect(reflect.ValueOf(valueRef)).Set(s.structValue.FieldByName(name))
}
func (s *StructRef) GetIntoIgnoreCase(name string, valueRef interface{}) {
	reflect.Indirect(reflect.ValueOf(valueRef)).Set(s.getFieldValueIgnoreCase(name))
}
func (s *StructRef) GetRef(name string) interface{} {
	return s.structValue.FieldByName(name).Addr().Interface()
}
func (s *StructRef) GetRefIgnoreCase(name string) interface{} {
	return s.getFieldValueIgnoreCase(name).Addr().Interface()
}

func StructFields(root reflect.Type) []reflect.StructField {
	for root.Kind() == reflect.Ptr {
		root = root.Elem()
	}
	n := root.NumField()
	var r []reflect.StructField
	for i := 0; i < n; i++ {
		f := root.Field(i)
		if !f.Anonymous {
			r = append(r, f)
		} else {
			r = append(r, StructFields(f.Type)...)
		}
	}
	return r
}
func (s *StructRef) Names() []string {
	fields := StructFields(s.structType)
	r := make([]string, len(fields))
	for i, f := range fields {
		r[i] = f.Name
	}
	return r
}

func (s *StructRef) NewSlice(len, cap int) interface{} {
	return reflect.MakeSlice(reflect.SliceOf(s.structType), len, cap).Interface()
}
func (s *StructRef) Map() map[string]interface{} {
	names := s.Names()
	r := make(map[string]interface{}, len(names))
	for _, v := range names {
		r[v] = s.Get(v)
	}
	return r
}
