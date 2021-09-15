package zeta

import (
	"reflect"
)

type SliceRef struct {
	// A slice pointer.
	Ref        interface{}
	refValue   reflect.Value
	sliceValue reflect.Value
	sliceType  reflect.Type
	eleType    reflect.Type
}

func NewSliceRef(ref interface{}) *SliceRef {
	r := &SliceRef{Ref: ref, refValue: reflect.ValueOf(ref)}
	r.sliceValue = reflect.Indirect(r.refValue)
	r.sliceType = r.sliceValue.Type()
	r.eleType = r.sliceType.Elem()
	return r
}

func (s *SliceRef) New(len, cap int) *SliceRef {
	entity := reflect.MakeSlice(s.sliceType, len, cap)
	s.sliceValue.Set(entity)
	return s
}

func (s *SliceRef) CreateInstance(len, cap int) interface{} {
	refValue := reflect.MakeSlice(s.sliceType, len, cap)
	return refValue.Interface()

	// r := &SliceRef{
	// 	Ref:        ref,
	// 	refValue:   refValue,
	// 	sliceValue: reflect.Indirect(refValue),
	// }
	// r.sliceType = r.sliceValue.Type()
	// r.eleType = r.sliceType.Elem()

	// return r
}

func (s *SliceRef) GetRef(i int) interface{} {
	return s.sliceValue.Index(i).Addr().Interface()
}
func (s *SliceRef) Set(i int, value interface{}) {
	s.sliceValue.Index(i).Set(reflect.ValueOf(value))
}

func (s *SliceRef) Get(i int) interface{} {
	return s.sliceValue.Index(i).Interface()
}
func (s *SliceRef) GetInto(i int, valueRef interface{}) {
	reflect.Indirect(reflect.ValueOf(valueRef)).Set(s.sliceValue.Index(i))
}
func (s *SliceRef) Len() int {
	if s.Ref == nil {
		return 0
	}
	return s.sliceValue.Len()
}

func (s *SliceRef) Append(value interface{}) *SliceRef {
	s.sliceValue = reflect.Append(s.sliceValue, reflect.ValueOf(value))
	reflect.Indirect(s.refValue).Set(s.sliceValue)
	return s
}
