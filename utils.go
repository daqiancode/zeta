package zeta

import "strings"

func hasNil(vs []interface{}) bool {
	for _, v := range vs {
		if v == nil {
			return true
		}
	}
	return false
}
func argsToMap(args ...interface{}) map[string]interface{} {
	n := len(args)/2 + len(args)%2
	r := make(map[string]interface{}, n)
	for i := 0; i < n; i++ {
		r[args[2*i].(string)] = args[2*i+1]
	}
	if len(args)%2 == 1 {
		r[args[len(args)-1].(string)] = nil
	}
	return r
}

func pick(structRef interface{}, fields ...string) map[string]interface{} {
	sr := NewStructRef(structRef)
	r := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		r[f] = sr.GetIgnoreCase(f)
	}
	return r
}
func pickFromMap(m map[string]interface{}, fields ...string) map[string]interface{} {
	r := make(map[string]interface{}, len(fields))
	for k, v := range m {
		for _, f := range fields {
			if strings.EqualFold(k, f) {
				r[f] = v
			}
		}
	}
	return r
}

// subtract a-b
func subtract(a, b []uint64) []uint64 {
	m := make(map[uint64]bool, len(b))
	for _, v := range b {
		m[v] = true
	}
	var r []uint64
	for _, v := range a {
		if _, ok := m[v]; !ok {
			r = append(r, v)
		}
	}
	return r
}
