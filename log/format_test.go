package log

import (
	"bytes"
	"encoding/hex"
	"github.com/go-stack/stack"
	"testing"
	"time"
)

func TestEscapeString(t *testing.T) {
	s1, _ := hex.DecodeString("01020304")
	_s1 := escapeString(string(s1))
	t.Log(_s1) // "\x01\x02\x03\x04"

	s2, _ := hex.DecodeString("613d6204")
	_s2 := escapeString(string(s2))
	t.Log(_s2) // "a=b\x04"

	s3, _ := hex.DecodeString("6162636465")
	_s3 := escapeString(string(s3))
	t.Log(_s3) // abcde
}

func TestFormatLogfmtUint64(t *testing.T) {
	var n1 uint64 = 1234
	var n2 uint64 = 1234567890

	t.Log(formatLogfmtUint64(n1, true))  // -1234
	t.Log(formatLogfmtUint64(n1, false)) // 1234
	t.Log(formatLogfmtUint64(n2, true))  // -1,234,567,890
	t.Log(formatLogfmtUint64(n2, false)) // 1,234,567,890
}

func TestFormatShared(t *testing.T) {
	var val1 *int
	t.Log(formatShared(val1))

	var val2 = time.Now()
	t.Log(formatShared(val2))
}

type A struct {
	X int
	Y string
	Z time.Time
}

func TestFormatJSONValue(t *testing.T) {
	a := A{
		X: 99,
		Y: "abc",
		Z: time.Now(),
	}
	t.Log(formatJSONValue(a))
}

func TestJSONFormat(t *testing.T) {
	r := Record{
		Time: time.Now(),
		Lvl:  3,
		Msg:  "Start network",
		Ctx:  []interface{}{"app", "ethereum/server", "consensus", "POS"},
		Call: stack.Caller(2),
		KeyNames: RecordKeyNames{
			Time: timeKey,
			Msg:  msgKey,
			Lvl:  lvlKey,
			Ctx:  ctxKey,
		},
	}
	bz := JSONFormat().Format(&r)
	t.Log(string(bz))
}

func TestColor(t *testing.T) {
	t.Logf("\x1b[%dm%s\x1b[0m", 35, "以太坊")
	t.Logf("\x1b[%dm%s\x1b[0m", 31, "以太坊")
	t.Logf("\x1b[%dm%s\x1b[0m", 33, "以太坊")
	t.Logf("\x1b[%dm%s\x1b[0m", 32, "以太坊")
	t.Logf("\x1b[%dm%s\x1b[0m", 36, "以太坊")
	t.Logf("\x1b[%dm%s\x1b[0m", 34, "以太坊")
}

func TestLogFmt(t *testing.T) {
	buffer := new(bytes.Buffer)
	buffer.WriteByte('\n')
	ctx1 := []interface{}{"app", "ethereum/server", "consensus", "POS", "validators", 40}
	ctx2 := []interface{}{"app", "blockchain", "consensus", "PBFT", "validators", 4}
	term := true
	logfmt(buffer, ctx1, 33, term)
	logfmt(buffer, ctx2, 31, term)
	t.Log(buffer.String())
}

func TestLogfmtFormat(t *testing.T) {
	r := Record{
		Time: time.Now(),
		Lvl:  3,
		Msg:  "Start network",
		Ctx:  []interface{}{"app", "ethereum/server", "consensus", "POS"},
		Call: stack.Caller(2),
		KeyNames: RecordKeyNames{
			Time: timeKey,
			Msg:  msgKey,
			Lvl:  lvlKey,
			Ctx:  ctxKey,
		},
	}
	format := LogfmtFormat()
	t.Log(string(format.Format(&r)))
	t.Log(byte(' '))
}
