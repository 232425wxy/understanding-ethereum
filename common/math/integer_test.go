package math

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseUint64(t *testing.T) {
	u1, ok1 := ParseUint64("0x123")
	assert.True(t, ok1)
	t.Log(u1)

	u2, ok2 := ParseUint64("123")
	assert.True(t, ok2)
	t.Log(u2)
}

func TestSafeSub(t *testing.T) {
	x := uint64(3)
	y := uint64(5)
	res, over := SafeSub(x, y)
	t.Log(res)
	t.Log(over)
}

func TestSafeAdd(t *testing.T) {
	//x := 1<<64 - 1 - 1
	y := uint64(1)
	res, over := SafeAdd(uint64(1<<64-1-1), y)
	t.Log(res == MaxUint64)
	t.Log(over)
}

func TestSafeMul(t *testing.T) {
	y := uint64(2)
	res, over := SafeMul(uint64((1<<64-1)/2), y)
	t.Log(res)
	t.Log(over)
}
