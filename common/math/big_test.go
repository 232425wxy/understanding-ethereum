package math

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestHexOrDecimal256_MarshalText(t *testing.T) {
	h := NewHexOrDecimal256(255)
	bz, err := h.MarshalText()
	assert.Nil(t, err)
	t.Log(string(bz))
}

func TestParseBig256(t *testing.T) {
	s1 := "0x12a"
	b1, ok1 := ParseBig256(s1)
	assert.True(t, ok1)
	t.Log(b1)

	s2 := "123"
	b2, ok2 := ParseBig256(s2)
	assert.True(t, ok2)
	t.Log(b2)

	s3 := "123a"
	b3, err3 := ParseBig256(s3)
	assert.False(t, err3)
	t.Log(b3)
}

func TestParseBig256Upper(t *testing.T) {
	num := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	b, ok := ParseBig256(num.String())
	assert.False(t, ok)
	num.Sub(num, big.NewInt(1))
	b, ok = ParseBig256(num.String())
	assert.True(t, ok)
	t.Log(b)
}

func TestDecimal256_String(t *testing.T) {
	d1 := NewDecimal256(0)
	t.Log(d1.String())

	d2 := NewDecimal256(123)
	t.Log(d2.String())

	d3 := NewDecimal256(-123)
	t.Log(d3.String())
}

func TestDecimal256_UnmarshalText(t *testing.T) {
	d := NewDecimal256(0)
	h := NewHexOrDecimal256(1234)
	bz, err := h.MarshalText()
	assert.Nil(t, err)
	t.Log(string(bz))
	err = d.UnmarshalText(bz)
	assert.Nil(t, err)
	t.Log(d.String())
}

func TestFirstBitSet(t *testing.T) {
	bz := []byte{8, 2, 0, 128}
	i := new(big.Int).SetBytes(bz)
	t.Log(i)
	t.Log("LSB:", FirstBitSet(i))
}

func TestPaddedBigBytes(t *testing.T) {
	// bigInt = 00001000,11001000,101
	bz := []byte{8, 200, 160}
	bigInt := new(big.Int).SetBytes(bz)
	t.Log(bigInt)
	res := PaddedBigBytes(bigInt, 5)
	t.Log(res)
	newBigInt := new(big.Int).SetBytes(res)
	assert.Equal(t, newBigInt, bigInt)
}

func TestBigEndianByteAt(t *testing.T) {
	// data = [1 2 3 4 5 6 7 8 9 10 11 12 13 14] = [00000001 00000010 00000011 00000100 00000101 00000110 00000111 00001000 00001001 00001010 00001011 00001100 00001101 00001110]
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	bigInt := new(big.Int).SetBytes(data)
	t.Log(bigInt.Bytes())
	b := bigEndianByteAt(bigInt, 11)
	assert.Equal(t, b, byte(3))
}

func TestByte(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	bigInt := new(big.Int).SetBytes(data)
	t.Log(bigInt)
	bigInt.SetBytes(PaddedBigBytes(bigInt, 17))
	t.Log(PaddedBigBytes(bigInt, 17))
	t.Log(Byte(bigInt, 17, 13))
}

func TestGlobalVariables(t *testing.T) {
	t.Log(tt256m1) // 115792089237316195423570985008687907853269984665640564039457584007913129639935
}

func TestU256(t *testing.T) {
	t.Log(U256(big.NewInt(5)))
}

func TestU256Bytes(t *testing.T) {
	n := big.NewInt(123)
	t.Log(U256Bytes(n))
}

func TestS256(t *testing.T) {
	x := BigPow(2, 255)
	x.Add(x, big.NewInt(1))
	t.Log(S256(x))
}

func TestExp(t *testing.T) {
	base := big.NewInt(3)
	exponent := big.NewInt(4)
	t.Log(Exp(base, exponent))
	t.Log(BigPow(3, 5))
	t.Log(base)
	t.Log(exponent)
}
