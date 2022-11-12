package rlp

import (
	"bytes"
	"io"
	"math/big"
	"testing"
)

type testEncoder struct {
	err error
}

func (e *testEncoder) EncodeRLP(w io.Writer) error {
	if e == nil {
		panic("EncodeRLP called on nil value")
	}
	if e.err != nil {
		return e.err
	}
	w.Write([]byte{0, 1, 0, 1, 0, 1, 0, 1, 0, 1})
	return nil
}

type testEncoderValueMethod struct{}

func (e testEncoderValueMethod) EncodeRLP(w io.Writer) error {
	w.Write([]byte{0xFA, 0xFE, 0xF0})
	return nil
}

type byteEncoder byte

func (e byteEncoder) EncodeRLP(w io.Writer) error {
	w.Write(EmptyList)
	return nil
}

type undecodableEncoder func()

func (f undecodableEncoder) EncodeRLP(w io.Writer) error {
	w.Write([]byte{0xF5, 0xF5, 0xF5})
	return nil
}

type encodeableReader struct {
	A, B int
}

func (e *encodeableReader) Read(bz []byte) (int, error) {
	panic("called")
}

type namedByteType byte

var reader io.Reader = &encodeableReader{1, 2}

type encTest struct {
	val    interface{}
	output string
	error  string
}

func run(t *testing.T, f func(val interface{}) ([]byte, error), test encTest, serial int) {
	output, err := f(test.val)
	if err != nil && test.error == "" {
		t.Errorf("%d: unexpected error: %v\nvalue: 	%#v\ntype: 	%T", serial, err, test.val, test.val)
	}
	if test.error != "" && err.Error() != test.error {
		t.Errorf("%d: error mismatch\ngot: 	%v\nwant: 	%v\nvalue: 	%#v\ntype: 	%T", serial, err, test.error, test.val, test.val)
	}
	if err == nil && !bytes.Equal(output, unhex(test.output)) {
		t.Errorf("%d: encode result mismatch\ngot: 	%X\nwant: 	%s\nvalue: 	%#v\ntype:	%T", serial, output, test.output, test.val, test.val)
	}
}

func f(val interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := Encode(buffer, val)
	return buffer.Bytes(), err
}

func TestEncodeBool(t *testing.T) {
	var encTests = []encTest{
		{val: true, output: "01"},
		{val: false, output: "80"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestEncodeInteger(t *testing.T) {
	var encTests = []encTest{
		{val: uint32(0), output: "80"},
		{val: uint32(127), output: "7F"},
		{val: uint32(128), output: "8180"},
		{val: uint32(256), output: "820100"},
		{val: uint32(1024), output: "820400"},
		{val: uint32(0xffffff), output: "83ffffff"},
		{val: uint32(0xffffffff), output: "84ffffffff"},
		{val: uint64(0xffffffffff), output: "85ffffffffff"},
		{val: uint64(0xffffffffffff), output: "86ffffffffffff"},
		{val: uint64(0xffffffffffffff), output: "87ffffffffffffff"},
		{val: uint64(0xffffffffffffffff), output: "88ffffffffffffffff"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestEncodeBigInteger(t *testing.T) {
	var encTests = []encTest{
		{val: big.NewInt(0), output: "80"},
		{val: big.NewInt(1), output: "01"},
		{val: big.NewInt(2), output: "02"},
		{val: big.NewInt(127), output: "7f"},
		{val: big.NewInt(128), output: "8180"},
		{val: big.NewInt(256), output: "820100"},
		{val: new(big.Int).SetBytes(unhex("123456789abcdef123456789abcdef")), output: "8f123456789abcdef123456789abcdef"},
		{val: new(big.Int).SetBytes(unhex("123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef")), output: "b83c123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef"},
		{val: *new(big.Int).SetBytes(unhex("123456789abcdef123456789abcdef")), output: "8f123456789abcdef123456789abcdef"},
		{val: *new(big.Int).SetBytes(unhex("123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef")), output: "b83c123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef123456789abcdef"},
		{val: new(big.Int).SetInt64(-2), error: "rlp: cannot encode negative big.Int"},
		{val: *new(big.Int).SetInt64(-2), error: "rlp: cannot encode negative big.Int"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}

}

func TestEncodeByteArray(t *testing.T) {
	var encTests = []encTest{
		{val: [0]byte{}, output: "80"},
		{val: [1]byte{0}, output: "00"},
		{val: [1]byte{1}, output: "01"},
		{val: [1]byte{127}, output: "7f"},
		{val: [1]byte{128}, output: "8180"},
		{val: [1]byte{0xff}, output: "81ff"},
		{val: [3]byte{1, 2, 3}, output: "83010203"},
		{val: [60]byte{1, 2, 3}, output: "b83c010203000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestEncodeNamedByteArray(t *testing.T) {
	var encTests = []encTest{
		{val: [0]namedByteType{}, output: "80"},
		{val: [1]namedByteType{0}, output: "00"},
		{val: [1]namedByteType{127}, output: "7f"},
		{val: [1]namedByteType{128}, output: "8180"},
		{val: [2]namedByteType{1, 2}, output: "820102"},
		{val: [3]namedByteType{1, 2, 3}, output: "83010203"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestEncodeByteSlice(t *testing.T) {
	var encTests = []encTest{
		{val: []byte{}, output: "80"},
		{val: []byte{0}, output: "00"},
		{val: []byte{1}, output: "01"},
		{val: []byte{127}, output: "7f"},
		{val: []byte{128}, output: "8180"},
		{val: []byte{1, 2, 3}, output: "83010203"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestEncodeNamedByteSlice(t *testing.T) {
	var encTests = []encTest{
		{val: []namedByteType{}, output: "80"},
		{val: []namedByteType{1}, output: "01"},
		{val: []namedByteType{2}, output: "02"},
		{val: []namedByteType{127}, output: "7f"},
		{val: []namedByteType{128}, output: "8180"},
		{val: []namedByteType{1, 2}, output: "820102"},
		{val: []namedByteType{1, 2, 4}, output: "83010204"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestEncodeString(t *testing.T) {
	var encTests = []encTest{
		{val: "", output: "80"},
		{val: "aaa", output: "83616161"},
		{val: "abc", output: "83616263"},
		{val: "My major is cyberspace security", output: "9f4d79206d616a6f722069732063796265727370616365207365637572697479"},
		{val: "RLP encoding is a new encoding method specifically implemented in the Ethereum", output: "b84e524c5020656e636f64696e672069732061206e657720656e636f64696e67206d6574686f64207370656369666963616c6c7920696d706c656d656e74656420696e2074686520457468657265756d"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestNoByteSlice(t *testing.T) {
	var encTests = []encTest{
		{val: []uint{}, output: "c0"},
		{val: []uint{1}, output: "c101"},
		{val: []uint{1, 9, 17}, output: "c3010911"},
		{val: []interface{}{[]interface{}{}}, output: "c1c0"},
		{val: []interface{}{[]interface{}{}, uint(3)}, output: "c2c003"},
		{val: []interface{}{[]interface{}{}, []interface{}{[]interface{}{}}}, output: "c3c0c1c0"},
		{val: []interface{}{[]interface{}{}, [][]interface{}{{}}, []interface{}{[]interface{}{}, [][]interface{}{{}}}}, output: "c7c0c1c0c3c0c1c0"},
		{val: []string{"aaa", "bbb", "ccc"}, output: "cc836161618362626283636363"},
		{val: []interface{}{uint(1), uint(0xffffff), []interface{}{[]uint{4, 5, 6}}, "abc"}, output: "ce0183ffffffc4c304050683616263"},
		{val: [][]string{{"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}}, output: "f841cc836161618362626283636363cc836161618362626283636363cc836161618362626283636363cc836161618362626283636363cc836161618362626283636363"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}

func TestRawValue(t *testing.T) {
	var encTests = []encTest{
		{val: RawValue{}, output: ""},
		{val: RawValue{0}, output: "00"},
		{val: RawValue{1}, output: "01"},
		{val: RawValue{127}, output: "7f"},
		{val: RawValue{128}, output: "80"},
		{val: RawValue{1, 2, 3}, output: "010203"},
		{val: []RawValue{{1, 2}, {3, 4}}, output: "c401020304"},
	}
	for i, test := range encTests {
		run(t, f, test, i)
	}
}
