package zeta

type tableUpdateListener struct {
	listeners map[string][]func(table string, oldModel interface{}, newModel interface{})
}

var TableUpdateListener tableUpdateListener = tableUpdateListener{listeners: make(map[string][]func(table string, oldModel interface{}, newModel interface{}))}

func (s *tableUpdateListener) On(table string, fn func(table string, oldModel interface{}, newModel interface{})) {
	s.listeners[table] = append(s.listeners[table], fn)
}
func (s *tableUpdateListener) Trigger(table string, oldModel interface{}, newModel interface{}) {
	if fns, ok := s.listeners[table]; ok {
		for _, fn := range fns {
			fn(table, oldModel, newModel)
		}
	}
}
