package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义包级全局变量

const (
	timeFormat  = time.RFC3339
	floatFormat = 'f'
)

type Format interface {
	Format(r *Record) []byte
}

func FormatFunc(f func(*Record) []byte) Format {
	return formatFunc(f)
}

type formatFunc func(*Record) []byte

// Format ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// Format 方法用于实现 Format 接口，实际上此处调用 Format 方法，其实就是在调用 formatFunc，
// 与 funcHandler 的 Log 方法是一样的道理。
func (f formatFunc) Format(r *Record) []byte {
	return f(r)
}

type TerminalStringer interface {
	TerminalString() string
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 函数

// JSONFormat ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// JSONFormat 返回一个格式化句柄，该句柄将 Record 格式化为 JSON 样式的输出格式，例如：
//
//	实例化一个 Record 对象：
//	r := Record{
//		Time: time.Now(),
//		Lvl:  3,
//		Msg:  "Start network",
//		Ctx:  []interface{}{"app", "ethereum/server", "consensus", "POS"},
//		Call: stack.Caller(2),
//		KeyNames: RecordKeyNames{
//			Time: timeKey,
//			Msg:  msgKey,
//			Lvl:  lvlKey,
//			Ctx:  ctxKey,
//		},
//	}
//	经过 JSONFormat 方法格式化后得到：
//	{"app":"ethereum/server","consensus":"POS","lvl":"info","msg":"Start network","t":"2022-11-22T16:08:06.96890076+08:00"}
func JSONFormat() Format {
	jsonMarshal := json.Marshal

	return FormatFunc(func(record *Record) []byte {
		props := make(map[string]interface{})
		props[record.KeyNames.Time] = record.Time
		props[record.KeyNames.Lvl] = record.Lvl.String()
		props[record.KeyNames.Msg] = record.Msg

		for i := 0; i < len(record.Ctx); i += 2 {
			k, ok := record.Ctx[i].(string)
			if !ok {
				props[errorKey] = fmt.Sprintf("%+v is not a string key", record.Ctx[i])
			}
			props[k] = formatJSONValue(record.Ctx[i+1])
		}
		bz, err := jsonMarshal(props)
		if err != nil {
			// 一般来讲是不会出错的
			bz, _ = jsonMarshal(map[string]string{
				errorKey: err.Error(),
			})
			return bz
		}
		bz = append(bz, '\n')
		return bz
	})
}

// FormatLogfmtInt64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// FormatLogfmtInt64 方法接受一个整数作为输入参数，注意该方法与 FormatLogfmtUint64 方法不同的地方在于，
// 那个方法仅限于接受正整数作为输入参数，而该方法可以接受正数，也可以接受负数，然后对给定的整数进行格式化输出，
// 例如：
//  1. 如果给定的整数为：1234567890，得到的输出是：1,234,567,890
//  2. 如果给定的整数为：-1234567890，得到的输出是：-1,234,567,890
func FormatLogfmtInt64(n int64) string {
	if n < 0 {
		return formatLogfmtUint64(uint64(-n), true)
	}
	return formatLogfmtUint64(uint64(n), false)
}

// FormatLogfmtUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// FormatLogfmtUint64 方法接受一个uint64类型的正整数作为输入，然后对该整数进行格式化输出，例如：
//   - 给定的正整数是：1234567890，得到的输出是：1,234,567,890
func FormatLogfmtUint64(n uint64) string {
	return formatLogfmtUint64(n, false)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// logfmt ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// logfmt
func logfmt(buf *bytes.Buffer, ctx []interface{}, color int, term bool) {
	for i := 0; i < len(ctx); i += 2 {
		if i != 0 {
			// 加一个空格
			buf.WriteByte(' ')
		}
	}
}

func formatLogfmtValue(value interface{}, term bool) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case time.Time:
		// Performance optimization: No need for escaping since the provided
		// timeFormat doesn't have any escape characters, and escaping is
		// expensive.
		return v.Format(timeFormat)

	case *big.Int:
		// Big ints get consumed by the Stringer clause so we need to handle
		// them earlier on.
		if v == nil {
			return "<nil>"
		}
		return formatLogfmtBigInt(v)
	}
	if term {
		if s, ok := value.(TerminalStringer); ok {
			// Custom terminal stringer provided, use that
			return escapeString(s.TerminalString())
		}
	}
	value = formatShared(value)
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), floatFormat, 3, 64)
	case float64:
		return strconv.FormatFloat(v, floatFormat, 3, 64)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case uint8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case uint16:
		return strconv.FormatInt(int64(v), 10)
	// Larger integers get thousands separators.
	case int:
		return FormatLogfmtInt64(int64(v))
	case int32:
		return FormatLogfmtInt64(int64(v))
	case int64:
		return FormatLogfmtInt64(v)
	case uint:
		return FormatLogfmtUint64(uint64(v))
	case uint32:
		return FormatLogfmtUint64(uint64(v))
	case uint64:
		return FormatLogfmtUint64(v)
	case string:
		return escapeString(v)
	default:
		return escapeString(fmt.Sprintf("%+v", value))
	}
}

// formatShared ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// formatShared 方法接受一个interface{}类型的value作为输入参数，value的底层类型属于以下三种类型，则会
// 做如下处理：
//  1. time.Time 类型：转换时间值的格式为"2006-01-02T15:04:05Z07:00"，得到输出例如为：2022-11-22T14:45:04+0800
//  2. error 类型：返回error.Error() string
//  3. 实现了 String() 方法的对象，返回其 String() 方法的返回值
//  4. 其他类型：不做处理，返回其原始值。
func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}

// formatJSONValue ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// formatJSONValue 方法接受一个interface{}类型的value作为参数，如果value的底层类型是数字类型
// 或字符串类型，就返回其原始值，如果是某个结构体类型，就会按照JSON格式将value完整地输出出来，包括
// 结构体的字段名。
func formatJSONValue(value interface{}) interface{} {
	value = formatShared(value)
	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64, string:
		return value
	default:
		return fmt.Sprintf("%+v", value)
	}
}

// formatLogfmtUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// formatLogfmtUint64 方法接受两个参数，第一个参数是一个uint64类型的整数，第二个参数是一个bool值，用来
// 指示整数是否是负数，该方法的作用就是对整数进行格式化输出，例如以下两个例子：
//  1. 如果给定的两个参数为：1234, false，得到的输出是：1234
//  2. 如果给定的两个参数为：1234567890, true，得到的输出是：-1,234,567,890
func formatLogfmtUint64(n uint64, neg bool) string {
	if n < 100000 {
		if neg {
			return strconv.Itoa(-int(n))
		} else {
			return strconv.Itoa(int(n))
		}
	}

	const maxLength = 26

	out := make([]byte, maxLength)
	i := maxLength - 1
	comma := 0

	for ; n > 0; i-- {
		if comma == 3 {
			comma = 0
			out[i] = ','
		} else {
			comma++
			out[i] = '0' + byte(n%10)
			n /= 10
		}
	}
	if neg {
		out[i] = '-'
		i--
	}
	return string(out[i+1:])
}

// formatLogfmtBigInt ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// formatLogfmtBigInt 接受一个大整数 *big.Int 作为输入参数，然后对该整数进行格式化输出。
func formatLogfmtBigInt(n *big.Int) string {
	if n.IsUint64() {
		return FormatLogfmtUint64(n.Uint64())
	}
	if n.IsInt64() {
		return FormatLogfmtInt64(n.Int64())
	}

	var (
		text  = n.String()
		buf   = make([]byte, len(text)+len(text)/3)
		comma = 0
		i     = len(buf) - 1
	)
	for j := len(text) - 1; j >= 0; j, i = j-1, i-1 {
		c := text[j]

		switch {
		case c == '-':
			buf[i] = c
		case comma == 3:
			buf[i] = ','
			i--
			comma = 0
			fallthrough
		default:
			buf[i] = c
			comma++
		}
	}
	return string(buf[i+1:])
}

// escapeString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/22|
//
// escapeString 方法接受一个字符串作为输入参数，该方法如果发现给定的字符串中存在一些特殊字符，就会在
// 给定的字符串两端加上双引号，否则就什么也不做，将原字符串返回。以下字符被定义为特殊字符：
//  1. ASCII码小于0x22的字符
//  2. ASCII码大于7e的字符
//  3. '='字符
func escapeString(s string) string {
	needsQuoting := false
	for _, r := range s {
		if r <= '"' || r > '~' || r == '=' {
			needsQuoting = true
			break
		}
	}
	if !needsQuoting {
		return s
	}
	return strconv.Quote(s)
}
