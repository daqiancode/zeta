package zeta_test

import (
	"testing"

	"github.com/daqiancode/zeta"
	"github.com/stretchr/testify/assert"
)

func TestGetRef(t *testing.T) {
	var a []string
	sr := zeta.NewSliceRef(&a)
	sr.New(3, 3)
	assert.Equal(t, 3, sr.Len())
	sr.Set(0, "jim")
	assert.Equal(t, "jim", sr.Get(0))
	sr.Append("tom")
	assert.Equal(t, 4, sr.Len())
	assert.Equal(t, "jim", sr.Get(0))
	assert.Equal(t, "tom", sr.Get(3))
	ref := sr.GetRef(1)
	*ref.(*string) = "lucy"
	assert.Equal(t, "lucy", sr.Get(1))

}
