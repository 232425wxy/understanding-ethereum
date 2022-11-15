package rlp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"math/big"
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
	if (err != nil && test.error != "" && err.Error() != test.error) || (err == nil && test.error != "") {
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
	var decodeTests = []decodeTest{
		{input: "05", ptr: new(uint32), value: uint32(5)},
		{input: "80", ptr: new(uint32), value: uint32(0)},
		{input: "820505", ptr: new(uint32), value: uint32(0x0505)},
		{input: "83050505", ptr: new(uint32), value: uint32(0x050505)},
		{input: "8405050505", ptr: new(uint32), value: uint32(0x05050505)},
		{input: "C0", ptr: new(uint32), error: "rlp: expected input string or byte for uint32"},
		{input: "B8020004", ptr: new(uint32), error: "rlp: non-canonical size information for uint32"},
		{input: "820004", ptr: new(uint32), error: "rlp: non-canonical integer (leading zero bytes) for uint32"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeSlices(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "c0", ptr: new([]uint), value: []uint{}},
		{input: "c80102030405060708", ptr: new([]uint), value: []uint{1, 2, 3, 4, 5, 6, 7, 8}},
		{input: "F8020004", ptr: new([]uint), error: "rlp: non-canonical size information for []uint"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeArrays(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "c50102030405", ptr: new([5]uint), value: [5]uint{1, 2, 3, 4, 5}},
		{input: "C0", ptr: new([5]uint), error: "rlp: input list has too few elements for [5]uint"},
		{input: "C102", ptr: new([5]uint), error: "rlp: input list has too few elements for [5]uint"},
		{input: "C6010203040506", ptr: new([5]uint), error: "rlp: input list has too many elements for [5]uint"},
		{input: "F8020004", ptr: new([5]uint), error: "rlp: non-canonical size information for [5]uint"},
		{input: "C0", ptr: new([0]uint), value: [0]uint{}},
		{input: "C101", ptr: new([0]uint), error: "rlp: input list has too many elements for [0]uint"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeByteSlice(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "01", ptr: new([]byte), value: []byte{1}},
		{input: "80", ptr: new([]byte), value: []byte{}},
		{input: "8D6162636465666768696A6B6C6D", ptr: new([]byte), value: []byte("abcdefghijklm")},
		{input: "C0", ptr: new([]byte), error: "rlp: expected input string or byte for []uint8"},
		{input: "8105", ptr: new([]byte), error: "rlp: non-canonical size information for []uint8"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeByteArrays(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "02", ptr: new([1]byte), value: [1]byte{2}},
		{input: "8180", ptr: new([1]byte), value: [1]byte{128}},
		{input: "850102030405", ptr: new([5]byte), value: [5]byte{1, 2, 3, 4, 5}},
		{input: "8400000000", ptr: new([4]byte), value: [4]byte{0x0, 0x0, 0x0, 0x0}},
		{input: "02", ptr: new([5]byte), error: "rlp: input string too short for [5]uint8"},
		{input: "80", ptr: new([5]byte), error: "rlp: input string too short for [5]uint8"},
		{input: "8400000000", ptr: new([5]byte), error: "rlp: input string too short for [5]uint8"},
		{input: "C0", ptr: new([5]byte), error: "rlp: expected input string or byte for [5]uint8"},
		{input: "C3010203", ptr: new([5]byte), error: "rlp: expected input string or byte for [5]uint8"},
		{input: "86010203040506", ptr: new([5]byte), error: "rlp: input string too long for [5]uint8"},
		{input: "8105", ptr: new([1]byte), error: "rlp: non-canonical size information for [1]uint8"},
		{input: "817F", ptr: new([1]byte), error: "rlp: non-canonical size information for [1]uint8"},
		{input: "80", ptr: new([0]byte), value: [0]byte{}},
		{input: "01", ptr: new([0]byte), error: "rlp: input string too long for [0]uint8"},
		{input: "8101", ptr: new([0]byte), error: "rlp: input string too long for [0]uint8"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeStrings(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "00", ptr: new(string), value: "\000"},
		{input: "8D6162636465666768696A6B6C6D", ptr: new(string), value: "abcdefghijklm"},
		{input: "C0", ptr: new(string), error: "rlp: expected input string or byte for string"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

var (
	veryBigInt = new(big.Int).Add(
		new(big.Int).Lsh(big.NewInt(0xFFFFFFFFFFFFFF), 16),
		big.NewInt(0xFFFF),
	)
	veryVeryBigInt = new(big.Int).Exp(veryBigInt, big.NewInt(8), nil)
)

func TestDecodeBigInt(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "80", ptr: new(*big.Int), value: big.NewInt(0)},
		{input: "80", ptr: new(*big.Int), value: new(big.Int)},
		{input: "01", ptr: new(*big.Int), value: big.NewInt(1)},
		{input: "89FFFFFFFFFFFFFFFFFF", ptr: new(*big.Int), value: veryBigInt},
		{input: "B848FFFFFFFFFFFFFFFFF800000000000000001BFFFFFFFFFFFFFFFFC8000000000000000045FFFFFFFFFFFFFFFFC800000000000000001BFFFFFFFFFFFFFFFFF8000000000000000001", ptr: new(*big.Int), value: veryVeryBigInt},
		{input: "10", ptr: new(big.Int), value: *big.NewInt(16)}, // non-pointer also works
		{input: "C0", ptr: new(*big.Int), error: "rlp: expected input string or byte for *big.Int"},
		{input: "00", ptr: new(*big.Int), error: "rlp: non-canonical integer (leading zero bytes) for *big.Int"},
		{input: "820001", ptr: new(*big.Int), error: "rlp: non-canonical integer (leading zero bytes) for *big.Int"},
		{input: "8105", ptr: new(*big.Int), error: "rlp: non-canonical size information for *big.Int"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

type bigIntStruct struct {
	I *big.Int
	B string
}

type tailUint struct {
	A    uint
	Tail []uint `rlp:"tail"`
}

type invalidNilTag struct {
	X []byte `rlp:"nil"`
}

type tailRaw struct {
	A    uint
	Tail []RawValue `rlp:"tail"`
}

type tailPrivateFields struct {
	A    uint
	Tail []uint `rlp:"tail"`
	x, y bool   //lint:ignore U1000 unused fields required for testing purposes.
}

type invalidTail2 struct {
	A uint
	B string `rlp:"tail"`
}

type invalidTail1 struct {
	A uint `rlp:"tail"`
	B string
}

type ignoredField struct {
	A uint
	B uint `rlp:"-"`
	C uint
}

type nilListUint struct {
	X *uint `rlp:"nilList"`
}

type nilStringSlice struct {
	X *[]uint `rlp:"nilString"`
}

type optionalPtrField struct {
	A uint
	B *[3]byte `rlp:"optional"`
}

func TestDecodeStructs(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "C50583343434", ptr: new(simplestruct), value: simplestruct{5, "444"}},
		{input: "C601C402C203C0", ptr: new(recstruct), value: recstruct{1, &recstruct{2, &recstruct{3, nil}}}},
		{input: "C58083343434", ptr: new(bigIntStruct), value: bigIntStruct{new(big.Int), "444"}},
		{input: "C28080", ptr: new(simplestruct), value: simplestruct{}},
		{input: "C0", ptr: new(simplestruct), error: "rlp: too few elements for rlp.simplestruct"},
		{input: "C105", ptr: new(simplestruct), error: "rlp: too few elements for rlp.simplestruct"},
		{input: "C7C50583343434C0", ptr: new([]*simplestruct), error: "rlp: too few elements for rlp.simplestruct, decoding into ([]*rlp.simplestruct)[1]"},
		{input: "83222222", ptr: new(simplestruct), error: "rlp: expected input list for rlp.simplestruct"},
		{input: "C3010101", ptr: new(simplestruct), error: "rlp: input list has too many elements for rlp.simplestruct"},
		{input: "C501C3C00000", ptr: new(recstruct), error: "rlp: expected input string or byte for uint, decoding into (rlp.recstruct).Child.I"},
		{input: "C103", ptr: new(intField), error: "rlp: type int is not RLP-serializable (struct field rlp.intField.X)"},
		{input: "C50102C20102", ptr: new(tailUint), error: "rlp: expected input string or byte for uint, decoding into (rlp.tailUint).Tail[1]"},
		{input: "C0", ptr: new(invalidNilTag), error: `rlp: invalid struct tag "nil" for rlp.invalidNilTag.X (field is not a pointer)`},
		{input: "C3010203", ptr: new(tailRaw), value: tailRaw{A: 1, Tail: []RawValue{unhex("02"), unhex("03")}}},
		{input: "C20102", ptr: new(tailRaw), value: tailRaw{A: 1, Tail: []RawValue{unhex("02")}}},
		{input: "C101", ptr: new(tailRaw), value: tailRaw{A: 1, Tail: []RawValue{}}},
		{input: "C3010203", ptr: new(tailPrivateFields), value: tailPrivateFields{A: 1, Tail: []uint{2, 3}}},
		{input: "C0", ptr: new(invalidTail1), error: `rlp: invalid struct tag "tail" for rlp.invalidTail1.A (tag "tail" is only allowed to be set on the last exportable field)`},
		{input: "C0", ptr: new(invalidTail2), error: `rlp: invalid struct tag "tail" for rlp.invalidTail2.B (tag "tail" is only allowed to be set on the slice type field)`},
		{input: "C20102", ptr: new(ignoredField), value: ignoredField{A: 1, C: 2}},
		{input: "C180", ptr: new(nilListUint), error: "rlp: wrong kind of empty value (got String, want List) for *uint, decoding into (rlp.nilListUint).X"},
		{input: "C1C0", ptr: new(nilListUint), value: nilListUint{}},
		{
			input: "C103",
			ptr:   new(nilListUint),
			value: func() interface{} {
				v := uint(3)
				return nilListUint{X: &v}
			}(),
		},
		{input: "C1C0", ptr: new(nilStringSlice), error: "rlp: wrong kind of empty value (got List, want String) for *[]uint, decoding into (rlp.nilStringSlice).X"},
		{input: "C180", ptr: new(nilStringSlice), value: nilStringSlice{}},
		{input: "C2C103", ptr: new(nilStringSlice), value: nilStringSlice{X: &[]uint{3}}},
		{input: "C101", ptr: new(optionalFields), value: optionalFields{1, 0, 0}},
		{input: "C20102", ptr: new(optionalFields), value: optionalFields{1, 2, 0}},
		{input: "C3010203", ptr: new(optionalFields), value: optionalFields{1, 2, 3}},
		{input: "C401020304", ptr: new(optionalFields), error: "rlp: input list has too many elements for rlp.optionalFields"},
		{input: "C101", ptr: new(optionalAndTailField), value: optionalAndTailField{A: 1}},
		{input: "C20102", ptr: new(optionalAndTailField), value: optionalAndTailField{A: 1, B: 2, Tail: []uint{}}},
		{input: "C401020304", ptr: new(optionalAndTailField), value: optionalAndTailField{A: 1, B: 2, Tail: []uint{3, 4}}},
		{input: "C101", ptr: new(optionalBigIntField), value: optionalBigIntField{A: 1, B: nil}},
		{input: "C20102", ptr: new(optionalBigIntField), value: optionalBigIntField{A: 1, B: big.NewInt(2)}},
		{input: "C101", ptr: new(optionalPtrField), value: optionalPtrField{A: 1}},
		{input: "C20180", ptr: new(optionalPtrField), error: "rlp: input string too short for [3]uint8, decoding into (rlp.optionalPtrField).B"},
		{input: "C20102", ptr: new(optionalPtrField), error: "rlp: input string too short for [3]uint8, decoding into (rlp.optionalPtrField).B"},
		{input: "C50183010203", ptr: new(optionalPtrField), value: optionalPtrField{A: 1, B: &[3]byte{1, 2, 3}}},
		{input: "C101", ptr: new(optionalPtrFieldNil), value: optionalPtrFieldNil{A: 1}},
		{input: "C20180", ptr: new(optionalPtrFieldNil), value: optionalPtrFieldNil{A: 1}},
		{input: "C20102", ptr: new(optionalPtrFieldNil), error: "rlp: input string too short for [3]uint8, decoding into (rlp.optionalPtrFieldNil).B"},
		{input: "C101", ptr: &optionalFields{A: 9, B: 8, C: 7}, value: optionalFields{A: 1, B: 0, C: 0}},
		{input: "C20102", ptr: &optionalFields{A: 9, B: 8, C: 7}, value: optionalFields{A: 1, B: 2, C: 0}},
		{input: "C20102", ptr: &optionalAndTailField{A: 9, B: 8, Tail: []uint{7, 6, 5}}, value: optionalAndTailField{A: 1, B: 2, Tail: []uint{}}},
		{input: "C101", ptr: &optionalPtrField{A: 9, B: &[3]byte{8, 7, 6}}, value: optionalPtrField{A: 1}},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeRawValue(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "01", ptr: new(RawValue), value: RawValue(unhex("01"))},
		{input: "82FFFF", ptr: new(RawValue), value: RawValue(unhex("82FFFF"))},
		{input: "C20102", ptr: new([]RawValue), value: []RawValue{unhex("01"), unhex("02")}},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func uintp(i uint) *uint { return &i }

func TestPointers(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "00", ptr: new(*[]byte), value: &[]byte{0}},
		{input: "80", ptr: new(*uint), value: uintp(0)},
		{input: "C0", ptr: new(*uint), error: "rlp: expected input string or byte for uint"},
		{input: "07", ptr: new(*uint), value: uintp(7)},
		{input: "817F", ptr: new(*uint), error: "rlp: non-canonical size information for uint"},
		{input: "8180", ptr: new(*uint), value: uintp(0x80)},
		{input: "C109", ptr: new(*[]uint), value: &[]uint{9}},
		{input: "C58403030303", ptr: new(*[][]byte), value: &[][]byte{{3, 3, 3, 3}}},
		{input: "C3808005", ptr: new([]*uint), value: []*uint{uintp(0), uintp(0), uintp(5)}},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

func TestDecodeInterface(t *testing.T) {
	var decodeTests = []decodeTest{
		{input: "00", ptr: new(interface{}), value: []byte{0}},
		{input: "01", ptr: new(interface{}), value: []byte{1}},
		{input: "80", ptr: new(interface{}), value: []byte{}},
		{input: "850505050505", ptr: new(interface{}), value: []byte{5, 5, 5, 5, 5}},
		{input: "C0", ptr: new(interface{}), value: []interface{}{}},
		{input: "C50183040404", ptr: new(interface{}), value: []interface{}{[]byte{1}, []byte{4, 4, 4}}},
		{input: "C3010203", ptr: new([]io.Reader), error: "rlp: type io.Reader is not RLP-serializable"},
		{input: "c330f9c030f93030ce3030303030303030bd303030303030", ptr: new(interface{}), error: "rlp: element is larger than containing list"},
	}
	for i, test := range decodeTests {
		runD(t, fd, test, i)
	}
}

type Dog struct {
	Name  string
	Child *Dog `rlp:"optional"`
}

func TestDecodeSelfStruct(t *testing.T) {
	d := &Dog{}
	input := "c7826161c3826262"
	comp := &Dog{Name: "aa", Child: &Dog{Name: "bb", Child: nil}}
	err := DecodeBytes(unhex(input), d)
	assert.Nil(t, err)
	assert.Equal(t, comp.Name, d.Name)
	assert.Equal(t, comp.Child.Name, d.Child.Name)
	assert.Equal(t, comp.Child.Child, d.Child.Child)
}
