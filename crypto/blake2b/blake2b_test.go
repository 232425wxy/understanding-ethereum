package blake2b

import (
	"testing"
)

func TestOutputIV(t *testing.T) {
	for _, v := range iv {
		t.Log(v)
	}
}

func TestArrayPtr(t *testing.T) {
	f := func(arr *[4]uint64) {
		a0 := arr[0]
		a0++
		t.Log(*arr, a0)
	}
	var arr *[4]uint64 = &[4]uint64{1, 2, 3, 4}
	f(arr)
	t.Log(*arr)

	arr_2 := [4]byte{9, 8, 7, 6}
	x := &(arr_2[2])
	t.Log(*x)
	*x += 10
	t.Log(arr_2)
}
