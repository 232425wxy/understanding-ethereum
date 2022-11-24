package log

import (
	"os"
)

var (
	root          = &logger{ctx: []interface{}{}, h: new(swapHandler)}
	StdoutHandler = StreamHandler(os.Stdout, LogfmtFormat())
)

func init() {
	// 默认情况下，日志记录器不输出日志
	root.SetHandler(DiscardHandler())
}

// New ♏ |作者：吴翔宇| 🍁 |日期：2022/11/24|
//
// New 方法接受若干对键值对作为新日志记录器的背景信息，每调用一次该方法，都会重新分配ctx，不会继承
// 之前的ctx，该方法返回一个全新的日志记录器。
func New(ctx ...interface{}) Logger {
	return root.New(ctx...)
}

func Root() Logger {
	return root
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 日志输出接口

func Trace(msg string, ctx ...interface{}) {
	root.write(msg, LvlTrace, ctx, skipLevel)
}

func Debug(msg string, ctx ...interface{}) {
	root.write(msg, LvlDebug, ctx, skipLevel)
}

func Info(msg string, ctx ...interface{}) {
	root.write(msg, LvlInfo, ctx, skipLevel)
}

func Warn(msg string, ctx ...interface{}) {
	root.write(msg, LvlWarn, ctx, skipLevel)
}

func Error(msg string, ctx ...interface{}) {
	root.write(msg, LvlError, ctx, skipLevel)
}

func Crit(msg string, ctx ...interface{}) {
	root.write(msg, LvlCrit, ctx, skipLevel)
}
