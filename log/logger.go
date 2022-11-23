package log

import (
	"fmt"
	"github.com/go-stack/stack"
	"os"
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

// write ♏ |作者：吴翔宇| 🍁 |日期：2022/11/23|
//
// write 方法将给定的日志消息、日志等级、日志里出现的键值对组装成一条完整的日志记录，然后将其打印出去。
func (l *logger) write(msg string, lvl Lvl, ctx []interface{}, skip int) {
	r := &Record{
		Time: time.Time{},
		Lvl:  lvl,
		Msg:  msg,
		Ctx:  newContext(l.ctx, ctx),
		Call: stack.Caller(skip),
		KeyNames: RecordKeyNames{
			Time: timeKey,
			Msg:  msgKey,
			Lvl:  lvlKey,
			Ctx:  ctxKey,
		},
	}
	_ = l.h.Log(r)
}

// New ♏ |作者：吴翔宇| 🍁 |日期：2022/11/23|
//
// New 方法接受一个参数ctx，该方法会从父辈日志记录器衍生出一个新的日志记录器，这个新的日志记录器
// 继承了父辈日志记录器的ctx和handler，handler决定了日志信息被打印到什么地方：文件？网络连接？
// 然后输入的参数ctx是这个新日志记录器自己的ctx，例如"consensus=hotstuff"。
func (l *logger) New(ctx ...interface{}) Logger {
	child := &logger{ctx: newContext(l.ctx, ctx), h: new(swapHandler)}
	child.SetHandler(l.h)
	return child
}

func (l *logger) Trace(msg string, ctx ...interface{}) {
	l.write(msg, LvlTrace, ctx, skipLevel)
}

func (l *logger) Debug(msg string, ctx ...interface{}) {
	l.write(msg, LvlDebug, ctx, skipLevel)
}

func (l *logger) Info(msg string, ctx ...interface{}) {
	l.write(msg, LvlInfo, ctx, skipLevel)
}

func (l *logger) Warn(msg string, ctx ...interface{}) {
	l.write(msg, LvlWarn, ctx, skipLevel)
}

func (l *logger) Error(msg string, ctx ...interface{}) {
	l.write(msg, LvlError, ctx, skipLevel)
}

func (l *logger) Crit(msg string, ctx ...interface{}) {
	l.write(msg, LvlCrit, ctx, skipLevel)
	os.Exit(1)
}

func (l *logger) GetHandler() Handler {
	return l.h.Get()
}

// SetHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/23|
//
// SetHandler 重新设置日志信息如何被打印的处理器。
func (l *logger) SetHandler(h Handler) {
	l.h.Swap(h)
}

type Lazy struct {
	Fn interface{}
}

// Ctx ♏ |作者：吴翔宇| 🍁 |日期：2022/11/23|
//
// Ctx 结构存储了日志信息里的键值对。
type Ctx map[string]interface{}

// toArray ♏ |作者：吴翔宇| 🍁 |日期：2022/11/23|
//
// toArray 方法将 Ctx 里的键值对“扁平化处理”，生成键值相互交替出现的数组。
func (c Ctx) toArray() []interface{} {
	arr := make([]interface{}, len(c)*2)
	i := 0
	for k, v := range c {
		arr[i] = k
		arr[i+1] = v
		i += 2
	}
	return arr
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// newContext ♏ |作者：吴翔宇| 🍁 |日期：2022/11/23|
//
// newContext 方法接受两个参数，prefix和suffix，这两个参数的类型都是[]interface{}，prefix
// 和suffix可以被认为是日志信息里的键值对信息，该方法就是将prefix和suffix里的键值对信息拼接到一
// 起，或者说，就是将suffix追加到prefix的后面。
func newContext(prefix []interface{}, suffix []interface{}) []interface{} {
	normalizedSuffix := normalize(suffix)
	newCtx := make([]interface{}, len(prefix)+len(normalizedSuffix))
	n := copy(newCtx, prefix)
	copy(newCtx[n:], normalizedSuffix)
	return newCtx
}

// normalize ♏ |作者：吴翔宇| 🍁 |日期：2022/11/23|
//
// normalize 方法接受一个interface{}切片ctx，ctx的长度如果等于1，那么极有可能它的第一个元素的
// 类型是 Ctx，Ctx 是一种map类型，那么我们就将这个map里的键值对取出，组成键值相互交替出现的数组，
// 并将这个得到的数组返回出去。如果ctx的长度等于1，但是第一个元素的类型不是 Ctx，那么就在ctx后面
// 补全一个元素，并加上一对键值对，说明日志信息里的键值对信息不是双数个数。如果ctx的长度不等于1，则
// 判断完ctx的长度为偶数后，就将其直接返回出去。
func normalize(ctx []interface{}) []interface{} {
	if len(ctx) == 1 {
		if ctxMap, ok := ctx[0].(Ctx); ok {
			ctx = ctxMap.toArray()
		}
	}
	if len(ctx)%2 != 0 {
		ctx = append(ctx, nil, errorKey, "Normalized odd number of arguments by adding nil")
	}
	return ctx
}
