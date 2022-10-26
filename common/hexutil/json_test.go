package hexutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBytes_MarshalText(t *testing.T) {
	b := Bytes([]byte{'a', 'b', 'c'})
	res, err := b.MarshalText()
	assert.Nil(t, err)
	t.Log(res)
	// 'a' -> 97 -> 64+32+1 -> 01100001
	// 54 -> 32+16+4+2 ->      00110110
	// 49 -> 32+16+1 ->        00110001
}
