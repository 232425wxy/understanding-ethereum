package rlp

import (
	"sync"
	"testing"
)

type Computer struct {
	Brand string
}

var testPool = sync.Pool{New: func() interface{} { return new(Computer) }}

func TestSyncPool(t *testing.T) {
	c1 := testPool.Get().(*Computer)
	c1.Brand = "戴尔"
	testPool.Put(c1)

	c2 := testPool.Get().(*Computer)
	t.Log(c2.Brand)

	testPool.Put(c2)

	c3 := testPool.Get().(*Computer)
	t.Log(c3.Brand)
}
