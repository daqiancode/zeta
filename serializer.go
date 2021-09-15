package zeta

import "encoding/json"

type ValueSerializer interface {
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte, value interface{}) error
}

type JSONValueSerializer struct {
}

func (s *JSONValueSerializer) Marshal(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (s *JSONValueSerializer) Unmarshal(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}
