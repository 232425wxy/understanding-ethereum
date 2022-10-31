package rlp

import (
	"sync/atomic"
	"testing"
)

func TestCurAtomicValue(t *testing.T) {
	cur := new(atomic.Value)
	cur.Store(make(map[int][]string))
	m := map[int][]string{0: []string{"dog", "cat"}}
	old := cur.Swap(m)
	t.Log(old)
	m1 := map[int][]string{1: []string{"💯", "😯"}}
	old = cur.Swap(m1)
	t.Log(old)
	m1[1] = []string{"🐲"}
	old = cur.Load()
	t.Log(old)
	old1 := cur.Load()
	t.Log(old1)
}
