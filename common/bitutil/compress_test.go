package bitutil

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
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

func TestHugeData(t *testing.T) {
	src := make([]byte, 1024*1024*512)
	for i := 0; i < 1024*512; i++ {
		src[rand.Intn(len(src))] = 1
	}
	result := CompressBytes(src)
	dec, err := DecompressBytes(result, len(src))
	assert.Nil(t, err)
	assert.Equal(t, dec, src)
	t.Log(fmt.Sprintf("压缩前数据大小:%.2fMB", float64(len(src))/float64(1024*1024)))
	t.Log(fmt.Sprintf("压缩后数据大小:%.2fMB", float64(len(result))/float64(1024*1024)))
	// 输出：
	// compress_test.go:37: 压缩前数据大小:512.00MB
	// compress_test.go:38: 压缩后数据大小:2.02MB
}
