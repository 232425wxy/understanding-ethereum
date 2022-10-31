package rlp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPutInt(t *testing.T) {
	var b = make([]byte, 8)
	var i uint64 = 1234
	size := putInt(b, i)
	assert.Equal(t, 2, size)
	assert.Equal(t, byte(4), b[0])
	assert.Equal(t, byte(210), b[1])
}

func TestIntSize(t *testing.T) {
	var i uint64 = 1234
	size := intSize(i)
	assert.Equal(t, 2, size)
}

func TestPutHead(t *testing.T) {
	buf := make([]byte, 3)
	res := putHead(buf, 0x80, 0xB7, 32)
	assert.Equal(t, 1, res)
	assert.Equal(t, byte(160), buf[0])

	buf = make([]byte, 3)
	res = putHead(buf, 0x80, 0xB7, 64)
	assert.Equal(t, 2, res)
	assert.Equal(t, byte(0xB8), buf[0])
	assert.Equal(t, byte(64), buf[1])

	buf = make([]byte, 3)
	res = putHead(buf, 0xC0, 0xF7, 36)
	assert.Equal(t, 1, res)
	assert.Equal(t, byte(228), buf[0])

	buf = make([]byte, 3)
	res = putHead(buf, 0xC0, 0xF7, 456)
	assert.Equal(t, 3, res)
	assert.Equal(t, byte(0xF9), buf[0])
	assert.Equal(t, byte(1), buf[1])
	assert.Equal(t, byte(200), buf[2])
}
