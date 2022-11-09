package rlp

import (
	"github.com/232425wxy/understanding-ethereum/rlp/internal/rlpstruct"
	"github.com/stretchr/testify/assert"
	"reflect"
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

func TestEncodeBuffer_WriteString(t *testing.T) {
	buf := getEncBuffer()
	s := "123456789"
	err := writeString(reflect.ValueOf(s), buf)
	assert.Equal(t, err, nil)
	assert.Equal(t, buf.str, []byte{0x89, '1', '2', '3', '4', '5', '6', '7', '8', '9'})
}

func TestMakePtrWriter(t *testing.T) {
	var i *uint64 = new(uint64)
	i = nil
	ptrptr := &i
	typ := reflect.TypeOf(ptrptr)
	buf := getEncBuffer()
	w, err := makeWriter(typ, rlpstruct.Tag{})
	assert.Equal(t, nil, err)
	assert.NotNil(t, w)
	w(reflect.ValueOf(ptrptr), buf)
	t.Log(buf.str)
}

type Cat struct {
	Name string
}

func TestCanAddr(t *testing.T) {
	c := &Cat{Name: "Tom"}
	val := reflect.ValueOf(c)
	t.Log(val.Elem().CanAddr())
	assert.Equal(t, val.Elem().Addr(), val)
	//assert.Equal(t, reflect.PtrTo(val.Elem().Type()), val)
	val1 := reflect.ValueOf(*c)
	t.Log(val1.CanAddr())
}

func makeByteArrayWriter_(typ reflect.Type) writer {
	switch typ.Len() {
	case 0:
		return writeLengthZeroByteArray
	case 1:
		return writeLengthOneByteArray
	default:
		return func(value reflect.Value, buffer *encBuffer) error {
			buffer.encodeStringHeader(value.Len())
			for i := 0; i < value.Len(); i++ {
				b := byte(value.Index(i).Uint())
				buffer.str = append(buffer.str, b)
			}
			return nil
		}
	}
}

func TestEncodeByteArray(t *testing.T) {
	arr := [10]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
	typ := reflect.TypeOf(arr)
	val := reflect.ValueOf(arr)
	w1, _ := makeByteArrayWriter(typ)
	w2 := makeByteArrayWriter_(typ)

	buf1 := getEncBuffer()
	buf2 := getEncBuffer()

	err1 := w1(val, buf1)
	assert.Nil(t, err1)
	err2 := w2(val, buf2)
	assert.Nil(t, err2)

	assert.Equal(t, buf1.str, buf2.str)
	t.Log(buf1.str, "\n", buf2.str)
}

func TestSliceElem(t *testing.T) {
	s := []string{"sss"}
	typ := reflect.TypeOf(s)
	t.Log(typ.Elem().Kind())
}
