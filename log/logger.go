package log

import (
	"fmt"
	"github.com/go-stack/stack"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义常量

const (
	timeKey  = "t"
	lvlKey   = "lvl"
	msgKey   = "msg"
	ctxKey   = "ctx"
	errorKey = "LOG15_ERROR"
)
const skipLevel = 2

type Lvl int

const (
	LvlCrit Lvl = iota
	LvlError
	LvlWarn
	LvlInfo
	LvlDebug
	LvlTrace
)

// AlignedString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// AlignedString 方法返回日志级别的字符串名，返回的字符串名都是由5个字符组成，分别如下所示：
//
//	LvlTrace: "TRACE"
//	LvlDebug: "DEBUG"
//	LvlInfo:  "INFO "
//	LvlWarn:  "WARN "
//	LvlError: "ERROR"
//	LvlCrit:  "CRIT "
func (l Lvl) AlignedString() string {
	switch l {
	case LvlTrace:
		return "TRACE"
	case LvlDebug:
		return "DEBUG"
	case LvlInfo:
		return "INFO "
	case LvlWarn:
		return "WARN "
	case LvlError:
		return "ERROR"
	case LvlCrit:
		return "CRIT "
	default:
		panic("bad level")
	}
}

// String ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// String 方法返回日志级别的字符串名，不同于 AlignedString 方法，它返回的字符串仅有4个字符长度，
// 且是小写形式的，分别如下所示：
//
//	LvlTrace: "trce"
//	LvlDebug: "dbug"
//	LvlInfo:  "info "
//	LvlWarn:  "warn "
//	LvlError: "eror"
//	LvlCrit:  "crit "
func (l Lvl) String() string {
	switch l {
	case LvlTrace:
		return "trce"
	case LvlDebug:
		return "dbug"
	case LvlInfo:
		return "info"
	case LvlWarn:
		return "warn"
	case LvlError:
		return "eror"
	case LvlCrit:
		return "crit"
	default:
		panic("bad level")
	}
}

// LvlFromString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// LvlFromString 方法接受一个字符串形式的日志等级作为输入，然后返回 Lvl 类型的值，如下所示：
//
//	"trace", "trce": LvlTrace
//	"debug", "dbug": LvlDebug
//	"info": LvlInfo
//	"warn": LvlWarn
//	"error", "eror": LvlError
//	"crit": LvlCrit
func LvlFromString(lvlString string) (Lvl, error) {
	switch lvlString {
	case "trace", "trce":
		return LvlTrace, nil
	case "debug", "dbug":
		return LvlDebug, nil
	case "info":
		return LvlInfo, nil
	case "warn":
		return LvlWarn, nil
	case "error", "eror":
		return LvlError, nil
	case "crit":
		return LvlCrit, nil
	default:
		return LvlDebug, fmt.Errorf("unknown level: %v", lvlString)
	}
}

// RecordKeyNames ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// RecordKeyNames
type RecordKeyNames struct {
	Time string
	Msg  string
	Lvl  string
	Ctx  string
}

// Record ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// Record
type Record struct {
	Time     time.Time
	Lvl      Lvl
	Msg      string
	Ctx      []interface{}
	Call     stack.Call
	KeyNames RecordKeyNames
}

// Logger ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// Logger
type Logger interface {
	// New 生成一个新的日志记录器，给定的ctx是若干对键值对，这若干对键值对会在以后每次输出日志记录时被输出出去
	New(ctx ...interface{}) Logger

	GetHandler() Handler

	SetHandler(h Handler)

	// Trace ctx是若干对键值对
	Trace(msg string, ctx ...interface{})
	Debug(msg string, ctx ...interface{})
	Info(msg string, ctx ...interface{})
	Warn(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
	Crit(msg string, ctx ...interface{})
}

type logger struct {
	ctx []interface{}
	h   *swapHandler
}

type Lazy struct {
	Fn interface{}
}
