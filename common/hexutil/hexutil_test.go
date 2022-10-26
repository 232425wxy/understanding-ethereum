package hexutil

import (
	"encoding/hex"
	"strconv"
	"testing"
)

func TestTheLargestOfUint(t *testing.T) {
	res1 := ^uint(0)
	res2 := uint64(^uint(0))
	res3 := ^uint64(0)
	t.Log(res1>>63, res1)
	t.Log(res2>>63, res2)
	t.Log(res3>>63, res3)
}

func TestLargestUint64ToHex(t *testing.T) {
	largestUInt64 := ^uint64(0)
	bz := make([]byte, 0)
	bz = strconv.AppendUint(bz, largestUInt64, 16)
	t.Log(bz)
	// output: [102 102 102 102 102 102 102 102 102 102 102 102 102 102 102 102]
	// 102 = 64 + 32 + 4 + 2 ==> 01100110
	h := hex.EncodeToString(bz)
	t.Log(h)
	// output: 66666666666666666666666666666666
}
