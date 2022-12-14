package rlp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplit(t *testing.T) {
	bz := []byte{204, 131, 97, 97, 97, 8, 198, 133, 72, 101, 102, 101, 105}
	k, c, r, err := Split(bz)
	assert.Equal(t, k, List)
	assert.Equal(t, []byte{131, 97, 97, 97, 8, 198, 133, 72, 101, 102, 101, 105}, c)
	assert.Equal(t, []byte{}, r)
	assert.Equal(t, nil, err)
}

func TestSplitUint64(t *testing.T) {
	bz := []byte{129, 130, 97, 98, 99}
	x, rest, err := SplitUint64(bz)
	assert.Nil(t, err)
	assert.Equal(t, uint64(x), x)
	t.Log(rest)
	assert.Equal(t, "abc", string(rest))
}

func TestCountValues(t *testing.T) {
	bz := []byte{129, 130, 12, 132, 97, 97, 97, 97}
	num, err := CountValues(bz)
	assert.Nil(t, err)
	assert.Equal(t, 3, num)
}

func TestAppendUint64(t *testing.T) {
	bz := []byte{129, 137}
	var i uint64 = 45678
	res := AppendUint64(bz, i)
	t.Log(res)
}
