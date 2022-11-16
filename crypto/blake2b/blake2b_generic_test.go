package blake2b

import (
	"github.com/232425wxy/understanding-ethereum/common/math"
	"testing"
)

func TestGeneric(t *testing.T) {
	var b byte = 64 // 10000000
	var c byte = 32
	c ^= b
	t.Log(b)
	t.Log(c)
}

func TestRotateLeft64(t *testing.T) {
	// 计算机里面存储是-1的补码：11111111,11111111,11111111,11111111,11111111,11111111,11111111,11111111
	var k int = -24
	const n = 64
	s := uint(k) & (n - 1)
	t.Log(uint64(1)<<s | uint64(1)>>(n-s))
	t.Log(s)
	t.Log(uint(k))
	t.Log(math.BigPow(2, 64).Uint64() - 1)

	i := uint64(4)
	t.Log(i<<2 | i>>2)
}
