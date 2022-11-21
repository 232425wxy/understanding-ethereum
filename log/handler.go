package log

import (
	"fmt"
	"github.com/go-stack/stack"
	"io"
	"net"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
)

// Handler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// Handler 接收日志记录器产生的日志条目，然后 Handler 定义了怎样将日志条目输出出去。
type Handler interface {
	Log(r *Record) error
}

// FuncHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// FuncHandler 根据给定的函数返回对应的 Handler。
func FuncHandler(fn func(r *Record) error) Handler {
	return funcHandler(fn)
}

type funcHandler func(r *Record) error

// Log ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// Log 我他妈真是服了，那些家伙为什么要把代码写的有这么多弯弯绕？？？
// 他这里调用 Log 方法实际上就是调用 funcHandler 函数。
func (fh funcHandler) Log(r *Record) error {
	return fh(r)
}

// FileHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// FileHandler 方法接受两个参数，第一个参数是日志文件的路径，第二个参数是记载日志的格式，如果给定的日志文件
// 存在，就在该文件后面追加日志条目，否则就创建一个日志文件。
func FileHandler(path string, fmtr Format) (Handler, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return closingHandler{f, StreamHandler(f, fmtr)}, nil
}

// NetHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// NetHandler 接受三个参数，前两个参数分别是网络类型和网络地址，例如："tcp"和"67.28.31.12"，第二个参数是
// 日志记载格式，该方法会主动拨通给定的网络地址，然后建立一个网络连接conn，并通过conn将日志内容发送给另一端的
// 网络设备，在那台网络设备上显式日志记录。
func NetHandler(network, addr string, fmtr Format) (Handler, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return closingHandler{conn, StreamHandler(conn, fmtr)}, nil
}

// closingHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// closingHandler 是一种支持关闭连接资源的 Handler，例如关闭文件、关闭网络连接等，官方说目前基本上还未用到
// closingHandler，在这里创建 closingHandler 是为了以后 Handler 能够实现Close功能做准备。
type closingHandler struct {
	io.WriteCloser
	Handler
}

func (h *closingHandler) Close() error {
	return h.WriteCloser.Close()
}

// StreamHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// StreamHandler 接受两个参数：io.Writer 和 Format，其中第一个参数用来接受日志信息，第二个参数决定将以
// 什么样的格式把日志记录写入到 io.Writer 里。
func StreamHandler(wr io.Writer, fmtr Format) Handler {
	h := FuncHandler(func(r *Record) error {
		_, err := wr.Write(fmtr.Format(r))
		return err
	})
	// 这里h是真正记录日志的句柄，SyncHandler将h包装成一个多线程安全的Handler
	return LazyHandler(SyncHandler(h))
}

// SyncHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// SyncHandler 接收一个 Handler 作为输入参数，将给定的 Handler 包装成一个多线程安全的 Handler。
func SyncHandler(h Handler) Handler {
	var mu sync.Mutex
	return FuncHandler(func(r *Record) error {
		mu.Lock()
		defer mu.Unlock()

		return h.Log(r)
	})
}

// LazyHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// LazyHandler 方法接受一个 Handler 作为输入参数，LazyHandler 其实就是再将给定的 Handler 进行包装，
// LazyHandler 内部调用 FuncHandler 函数，并将其返回值返回，FuncHandler 接受的参数是一个函数的定义，
// 这个函数的定义是由 LazyHandler 函数设计的，将来我们调用 LazyHandler 函数返回的 Handler 的Log方法
// 时，实际上就是调用 LazyHandler -> FuncHandler -> 入参函数定义，在这个函数的定义内，会将日志记录 Record
// 里的 Ctx 过滤一遍，目的就是找到 Ctx 的value里面是否存在 Lazy 的实例，如果有的话，就执行这个 Lazy 实
// 例里的Fn函数，并将Fn函数的返回值替代Ctx中对应位置处的value。
func LazyHandler(h Handler) Handler {
	return FuncHandler(func(r *Record) error {
		hadErr := false
		for i := 1; i < len(r.Ctx); i += 2 {
			lz, ok := r.Ctx[i].(Lazy)
			if ok {
				v, err := evaluateLazy(lz)
				if err != nil {
					hadErr = true
					r.Ctx[i] = err
				} else {
					if cs, ok := v.(stack.CallStack); ok {
						// r.Call 是调用栈中的一个条目，调用栈的栈顶表示最开始调用的地方，越往下代表调用的越深，
						// TrimBelow方法就是将cs这个调用栈中处在r.Call条目之下的所有调用条目去除掉，例如cs是
						// [logger_test.go:31 testing.go:1446 asm_amd64.s:1594]，r.Call是 testing.go:1446，
						// 那么调用TrimBelow之后，cs就会变成 [logger_test.go:31 testing.go:1446]，TrimRuntime
						// 方法则是将cs调用栈中调用GOROOT源码的调用条目去掉，例如这里的testing.go:1446就是GOROOT
						// 里的代码，那么在执行完TrimRuntime之后，cs就只剩下[logger_test.go:31]了。
						// 实际上，在以太坊中，这段代码似乎永远都不会调用到。
						v = cs.TrimBelow(r.Call).TrimRuntime()
					}
					r.Ctx[i] = v
				}
			}
		}

		if hadErr {
			r.Ctx = append(r.Ctx, errorKey, "bad lazy")
		}

		return h.Log(r)
	})
}

// LvlFilterHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// LvlFilterHandler 方法接受两个参数，分别是日志等级和 Handler，第一个参数设置了日志等级阈值，只有日志级别
// 小于第一个参数的日志才能被输出，众所周知，critical日志级别最高，trace日志级别最低。
func LvlFilterHandler(maxLvl Lvl, h Handler) Handler {
	return FilterHandler(func(r *Record) (pass bool) {
		return r.Lvl <= maxLvl
	}, h)
}

// FilterHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// FilterHandler 接受两个参数作为入参，分别是函数fn func(r *Record) bool和 Handler，如果fn的返回值等于
// true，则调用 Handler 的Log方法将日志内容输出出去，否则什么也不干，忽略这条日志信息。例如，我们只输出日志中
// 存在"err"键，并且其对应的值不等于"nil"的日志：
//
//	logger.SetHandler(FilterHandler(func(r *Record) bool {
//	    for i := 0; i < len(r.Ctx); i += 2 {
//	        if r.Ctx[i] == "err" {
//	            return r.Ctx[i+1] != nil
//	        }
//	    }
//	    return false
//	}, h))
func FilterHandler(fn func(r *Record) bool, h Handler) Handler {
	return FuncHandler(func(r *Record) error {
		if fn(r) {
			return h.Log(r)
		}
		return nil
	})
}

// DiscardHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// DiscardHandler 方法用于禁用日志功能。
func DiscardHandler() Handler {
	return FuncHandler(func(r *Record) error {
		return nil
	})
}

// evaluateLazy ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// evaluateLazy 方法接收一个 Lazy 实例作为参数，Lazy 是一个结构体，内部只有一个 Fn 作为其唯一的
// 字段，Fn的类型是interface{}，所以理论上可以是任意数据类型，但是 evaluateLazy 方法要求Fn必须
// 是一个函数，而且该函数不能含有入参，而且必须具有返回值。在满足以上条件之后，evaluateLazy 方法会
// 运行Fn函数，并将其返回值返回出来。
func evaluateLazy(lz Lazy) (interface{}, error) {
	t := reflect.TypeOf(lz.Fn)

	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("INVALID_LAZY, not func: %+v", lz.Fn)
	}

	if t.NumIn() > 0 {
		return nil, fmt.Errorf("INVALID_LAZY, func takes args: %+v", lz.Fn)
	}

	if t.NumOut() == 0 {
		return nil, fmt.Errorf("INVALID_LAZY, no func return val: %+v", lz.Fn)
	}

	value := reflect.ValueOf(lz.Fn)
	// 因为lz.Fn是一个不接受任何输入参数的函数，因此调用时，传入的参数就为[]reflect.Value{}。
	results := value.Call([]reflect.Value{})
	if len(results) == 1 {
		return results[0].Interface(), nil
	}
	values := make([]interface{}, len(results))
	for i, v := range results {
		values[i] = v.Interface()
	}
	return values, nil
}

// swapHandler ♏ |作者：吴翔宇| 🍁 |日期：2022/11/21|
//
// swapHandler 可以在多线程情况下安全的切换 Handler。
type swapHandler struct {
	handler atomic.Value
}

func (h *swapHandler) Log(r *Record) error {
	return (*h.handler.Load().(*Handler)).Log(r)
}

func (h *swapHandler) Swap(newHandler Handler) {
	h.handler.Store(&newHandler)
}

func (h *swapHandler) Get() Handler {
	return *h.handler.Load().(*Handler)
}
