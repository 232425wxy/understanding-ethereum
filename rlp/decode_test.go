package rlp

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBigEndianInt(t *testing.T) {
	buf := getEncBuffer()
	i := uint64(123456789)
	buf.writeUint64(i)
	auxiliary := [8]byte{}
	for index := range auxiliary {
		auxiliary[index] = 0
	}
	start := 8 - (len(buf.str) - 1)
	j := 1
	for ; start < len(auxiliary); start++ {
		auxiliary[start] = buf.str[j]
		j++
	}
	x := binary.BigEndian.Uint64(auxiliary[:])
	t.Log(x)
}

type Dog struct {
	Name     string
	Age      uint8
	Location string
}

func TestReadKind(t *testing.T) {
	d := Dog{
		Name:     "大黄",
		Age:      3,
		Location: "中国安徽合肥庐阳区三孝口街道杏花社区大门口",
	}
	bz, err := EncodeToBytes(d)
	assert.Nil(t, err)
	t.Log(bz)
	t.Log(len(bz))
}
