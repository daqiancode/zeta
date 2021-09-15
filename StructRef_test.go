package zeta_test

import (
	"testing"

	"github.com/daqiancode/zeta"
	"github.com/stretchr/testify/assert"
)

type TestBasic struct {
	ID int
}

type TestStruct struct {
	TestBasic
	Name string
	Age  int
}

func TestStructRef(t *testing.T) {
	var obj TestStruct
	s := zeta.NewStructRef(&obj)
	s.Set("ID", 1)
	s.SetIgnoreCase("name", "jim")
	assert.Equal(t, 1, obj.ID)
	assert.Equal(t, 1, s.Get("ID"))
	assert.Equal(t, 1, s.GetIgnoreCase("ID"))
	ageRef := s.GetRef("Age")
	*ageRef.(*int) = 20
	assert.Equal(t, 20, obj.Age)
}
