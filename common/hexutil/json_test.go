package hexutil

import (
	"github.com/stretchr/testify/assert"
	"math/big"
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

func TestBytes_UnmarshalGraphQL(t *testing.T) {
	b := new(Bytes)
	err := b.UnmarshalGraphQL("1234")
	assert.Equal(t, err, ErrMissingPrefix)
	err = b.UnmarshalGraphQL("0x5448")
	assert.Nil(t, err)
	t.Log(b)
}

func TestBytes_UnmarshalJSON(t *testing.T) {
	b := new(Bytes)
	// 1. 测试奇数个16进制数错误
	input := []byte{'"', '0', 'x', 'a', 'b', 'c', '"'}
	err := b.UnmarshalJSON(input)
	assert.Equal(t, err, wrapTypeError(ErrOddLength, bytesT))
	// 2. 测试正常输入
	input = []byte{'"', '0', 'x', 'a', 'b', '"'}
	err = b.UnmarshalJSON(input)
	assert.Nil(t, err)
	// 3. 测试只有左边有引号
	input = []byte{'"', '0', 'x', 'a', 'b'}
	err = b.UnmarshalJSON(input)
	assert.Equal(t, err, errNonString(bytesT))
	// 4. 测试只有右边有引号
	input = []byte{'0', 'x', 'a', 'b', '"'}
	err = b.UnmarshalJSON(input)
	assert.Equal(t, err, errNonString(bytesT))
	// 5. 测试没有前缀
	input = []byte{'"', 'a', 'b', '"'}
	err = b.UnmarshalJSON(input)
	assert.Equal(t, err, wrapTypeError(ErrMissingPrefix, bytesT))
	// 6. 测试错误的前缀
	input = []byte{'"', 'o', 'x', 'a', 'b', '"'}
	err = b.UnmarshalJSON(input)
	assert.Equal(t, err, wrapTypeError(ErrMissingPrefix, bytesT))
	// 7. 测试空字符串
	input = []byte{'"', '"'}
	err = b.UnmarshalJSON(input)
	assert.Equal(t, err, nil)
	// 8. 测试空值数字
	input = []byte{'"', '0', 'x', '"'}
	err = b.UnmarshalJSON(input)
	assert.Equal(t, err, nil)
}

func TestBig_UnmarshalText(t *testing.T) {
	bInt := new(big.Int).SetInt64(38478678)
	b := (*Big)(bInt)
	marshal, err := b.MarshalText()
	assert.Nil(t, err)
	t.Log(string(marshal))
	unmarshal := new(Big)
	err = unmarshal.UnmarshalText(marshal)
	assert.Nil(t, err)
	t.Log((*big.Int)(unmarshal).Int64())
	// 故意编码一个负数，然后再对其解码，看看是否出错
	bInt.SetInt64(-123)
	marshal, err = b.MarshalText()
	assert.Nil(t, err)
	t.Log(string(marshal))
	err = unmarshal.UnmarshalText(marshal)
	assert.Equal(t, err, ErrMissingPrefix)
}

func TestUint64_MarshalText(t *testing.T) {
	i := Uint64(75)
	bz, err := i.MarshalText()
	assert.Nil(t, err)
	t.Log(bz, string(bz))
}

func TestUint64_UnmarshalText(t *testing.T) {
	i := Uint64(75)
	bz, err := i.MarshalText()
	assert.Nil(t, err)
	i2 := new(Uint64)
	err = i2.UnmarshalText(bz)
	assert.Nil(t, err)
	assert.Equal(t, *i2, i)
}
