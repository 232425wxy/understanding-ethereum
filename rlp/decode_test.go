package rlp

import (
	"encoding/hex"
	"fmt"
	"io"
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
