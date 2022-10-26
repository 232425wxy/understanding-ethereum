package hexutil

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/big"
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

func TestWhatIsBigWord(t *testing.T) {
	b, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFF", 16)
	BB := new(big.Int).SetUint64(^uint64(0))
	t.Log(len(b.Bits()))
	t.Log(b.String())
	t.Log(b.Bits())
	t.Log(BB.Bits())
	str := hex.EncodeToString(b.Bytes())
	t.Log(str)
}

func TestDecode(t *testing.T) {
	h := "0x43444546"
	res, err := Decode(h)
	assert.Nil(t, err)
	t.Log(string(res))
}

func TestDecodeUint64(t *testing.T) {
	h := "0x1f"
	res, err := DecodeUint64(h)
	assert.Nil(t, err)
	t.Log(res)
	// output: 31
}

func TestEncode(t *testing.T) {
	bz := []byte{97, 98, 99, 100}
	result := Encode(bz)
	t.Log(result)
}

func TestEncodeUint64(t *testing.T) {
	num := 24
	res := EncodeUint64(uint64(num))
	t.Log(res)
}

func TestBigWordLittleEndian(t *testing.T) {
	bz := []byte{32, 16}
	b := new(big.Int).SetBytes(bz)
	t.Log(b.Bits())
	// output: [8208]
	// 32: 00100000
	// 16: 00010000
	// 00100000,00010000
}

func TestDecodeBig(t *testing.T) {
	h := "0x123"
	result, err := DecodeBig(h)
	assert.Nil(t, err)
	t.Log(result.Uint64())
}

func TestEncodeBig(t *testing.T) {
	b := new(big.Int).SetInt64(-12)
	result := EncodeBig(b)
	t.Log(result)
}
