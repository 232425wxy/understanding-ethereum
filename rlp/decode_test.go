package rlp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"strings"
	"testing"
)

type plainReader []byte

func newPlainReader(bz []byte) io.Reader {
	return (*plainReader)(&bz)
}

func (r *plainReader) Read(buf []byte) (n int, err error) {
	if len(*r) == 0 {
		return 0, io.EOF
	}
	n = copy(buf, *r)
	*r = (*r)[n:]
	return n, nil
}

// unhex 将16进制数转换为字节切片，字节切片里的元素是10进制形式的
func unhex(str string) []byte {
	b, err := hex.DecodeString(strings.ReplaceAll(str, " ", ""))
	if err != nil {
		panic(fmt.Sprintf("invalid hex string: %q", str))
	}
	return b
}

func TestStreamKind(t *testing.T) {
	tests := []struct {
		input    string
		wantKind Kind
		wantLen  uint64
	}{
		{"00", Byte, 0},
		{"01", Byte, 0},
		{"7F", Byte, 0},
		{"80", String, 0},
		{"B7", String, 55},
		{"B90400", String, 1024},
		{"BFFFFFFFFFFFFFFFFF", String, ^uint64(0)},
		{"C0", List, 0},
		{"C8", List, 8},
		{"F7", List, 55},
		{"F90400", List, 1024},
		{"FFFFFFFFFFFFFFFFFF", List, ^uint64(0)},
	}

	for i, test := range tests {
		// using plainReader to inhibit input limit errors.
		s := NewStream(newPlainReader(unhex(test.input)), 0)
		kind, len, err := s.Kind()
		if err != nil {
			t.Errorf("test %d: Kind returned error: %v", i, err)
			continue
		}
		if kind != test.wantKind {
			t.Errorf("test %d: kind mismatch: got %d, want %d", i, kind, test.wantKind)
		}
		if len != test.wantLen {
			t.Errorf("test %d: len mismatch: got %d, want %d", i, len, test.wantLen)
		}
	}

	t.Log([]byte(tests[5].input))
	t.Log(unhex(tests[5].input))
}

func TestNewListStream(t *testing.T) {
	ls := NewListStream(bytes.NewReader(unhex("8361616101020383636363")), 11)
	if k, size, err := ls.Kind(); k != List || err != nil {
		t.Errorf("Kind() returned (%v, %d, %v), expected (List, 11, nil)", k, size, err)
	}
	if size, err := ls.ListStart(); size != 11 || err != nil {
		t.Errorf("ListStart() returned (%d, %v), expected (1, nil)", size, err)
	}
	for i := 0; i < 5; i++ {
		if val, err := ls.Bytes(); err != nil {
			t.Errorf("Uint64() returned (%v), expected (nil)", err)
		} else {
			t.Log(val)
		}
	}
}

func TestStreamList(t *testing.T) {
	s := NewStream(bytes.NewReader(unhex("c80102030405060708")), 0)

	length, err := s.ListStart()
	assert.Nil(t, err)
	assert.Equal(t, uint64(8), length)

	for i := uint64(1); i <= length; i++ {
		val, err := s.Uint64()
		assert.Nil(t, err)
		assert.Equal(t, i, val)
	}

	err = s.ListEnd()
	assert.Nil(t, err)
}

func TestStreamRaw(t *testing.T) {
	tests := []struct {
		intput string
		output string
	}{
		{
			intput: "c58401010101",
			output: "8401010101",
		},
		{
			intput: "F842B84001010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
			output: "B84001010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
		},
	}
	for _, test := range tests {
		s := NewStream(bytes.NewReader(unhex(test.intput)), 0)
		s.ListStart()
		want := unhex(test.output)
		raw, err := s.Raw()
		assert.Nil(t, err)
		assert.Equal(t, want, raw)
	}
}

func TestStreamReadBytes(t *testing.T) {
	tests := []struct {
		input string
		size  int
		err   string
	}{
		{input: "c0", size: 1, err: "rlp: expected String or Byte"},
		{input: "04", size: 0, err: "input value has wrong size 1, want 0"},
		{input: "04", size: 1},
		{input: "04", size: 2, err: "input value has wrong size 1, want 2"},
		{input: "820102", size: 0, err: "input value has wrong size 2, want 0"},
		{input: "820102", size: 1, err: "input value has wrong size 2, want 1"},
		{input: "820102", size: 2},
		{input: "820102", size: 3, err: "input value has wrong size 2, want 3"},
	}

	for _, test := range tests {
		name := fmt.Sprintf("input_%s/size_%d", test.input, test.size)
		t.Run(name, func(t *testing.T) {
			s := NewStream(bytes.NewReader(unhex(test.input)), 0)
			b := make([]byte, test.size)
			err := s.ReadBytes(b)
			if test.err == "" {
				if err != nil {
					t.Errorf("unexpected error (%v)", err)
				}
			} else {
				if err == nil {
					t.Error("expected error, got nil")
				} else if err.Error() != test.err {
					t.Errorf("wrong error, got (%v), want (%v)", err, test.err)
				}
			}
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	r := bytes.NewReader(nil)

	if err := Decode(r, nil); err != errDecodeIntoNil {
		t.Errorf("Decode(r, nil) error mismatch, got %q, want %q", err, errDecodeIntoNil)
	}

	var nilPtr *struct{}
	if err := Decode(r, nilPtr); err != errDecodeIntoNil {
		t.Errorf("Decode(r, nil) error mismatch, got %q, want %q", err, errDecodeIntoNil)
	}

	if err := Decode(r, struct{}{}); err != errNoPointer {
		t.Errorf("Decode(r, struct{}{}) error mismatch, got %q, want %q", err, errNoPointer)
	}

	expectErr := "rlp: type chan bool is not RLP-serializable"
	if err := Decode(r, new(chan bool)); err == nil || err.Error() != expectErr {
		t.Errorf("Decode(r, new(chan bool)) error mismatch, got %q, want %q", err, expectErr)
	}

	if err := Decode(r, new(uint)); err != io.EOF {
		t.Errorf("Decode(r, new(int)) error mismatch, got %q, want %q", err, io.EOF)
	}
}

type decodeTest struct {
	input string
	ptr   interface{}
	value interface{}
	error string
}

func runD(t *testing.T, f func(test decodeTest) error, test decodeTest, serial int) {
	err := f(test)
	if err != nil && test.error == "" {
		t.Errorf("%d: unexpected Decode error: %v\ndecoding into %T\ninput: %q", serial, err, test.ptr, test.input)
	}
	if test.error != "" && err.Error() != test.error {
		t.Errorf("%d: Decode error mismatch\ngot:	%v\nwant:	%v\ndecoding into:	%T\ninput:	%q", serial, err, test.error, &test, test.input)
	}
	deref := reflect.ValueOf(test.ptr).Elem().Interface()
	if err == nil && !reflect.DeepEqual(deref, test.value) {
		t.Errorf("%d: Decode value mismatch\ngot:	%#v\nwant:	%#v\ndecoding into:	%T\ninput:	%q", serial, deref, test.value, test.ptr, test.input)
	}
}

func fd(test decodeTest) error {
	input := unhex(test.input)
	return Decode(bytes.NewReader(input), test.ptr)
}

func TestDecodeBool(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "01", ptr: new(bool), value: true},
		{input: "80", ptr: new(bool), value: false},
		{input: "02", ptr: new(bool), error: "rlp: invalid boolean value: 2"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeIntegers(t *testing.T) {

}
