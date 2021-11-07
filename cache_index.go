package zeta

import (
	"reflect"
	"sync"
)

// format: map[string][][]string
var cacheIndex sync.Map

func AddCacheIndex(tableStruct interface{}, index ...string) {
	for _, v := range index {
		tableValue := reflect.ValueOf(tableStruct)
		field := tableValue.FieldByName(v)
		if !field.IsValid() {
			panic("No \"" + v + "\" field in struct " + tableValue.String())
		}
	}
	var table string
	if t, ok := tableStruct.(string); ok {
		table = t
	} else {
		table = TableName(tableStruct)
	}
	if indexes, ok := cacheIndex.Load(table); ok {
		indexes = append(indexes.([][]string), index)
		cacheIndex.Store(table, indexes)
	} else {
		cacheIndex.Store(table, [][]string{index})
	}
}
