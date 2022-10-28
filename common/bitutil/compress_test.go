package bitutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBitsetEncodeBytes(t *testing.T) {
	// data = [0 0 0 0 1 0 0 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 0 0 3 0 0 0 0 0 0 0 0]
	data := make([]byte, 32)
	data[4] = 1
	data[14] = 2
	data[24] = 3
	result := bitsetEncodeBytes(data)
	t.Log(result)
}

func TestBitsetDecodedPartialBytes(t *testing.T) {
	data := []byte{208, 8, 2, 128, 1, 2, 3}
	result, ptr, err := bitsetDecodePartialBytes(data, 32)
	assert.Nil(t, err)
	t.Log("解压缩结果:", result)
	t.Log("ptr:", ptr)
}
