/*
Package math

该文件类定义了一个类型：HexOrDecimal64，该类型的地层实现是uint64，通过它实现了对uint64进行marshal/unmarshal，marshal会将
64位无符号整型编码成含有"0x"前缀的16进制数，unmarshal支持将含有"0x"或"0X"前缀的16进制数或者10进制数转换成64位无符号整型。

此外，该文件还定义了如下全局函数：
  - func SafeSub(x, y uint64) (uint64, bool)
  - func SafeAdd(x, y uint64) (uint64, bool)
  - func SafeMul(x, y uint64) (uint64, bool)

以上三个全局函数的第二个返回值反映了对两个64位无符号整型进行加、减乘操作后是否会出现溢出。如果溢出，则第二个返回值为true。
*/
package math

import (
	"fmt"
	"math/bits"
	"strconv"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

const (
	// MaxInt8 <=> 01111111
	MaxInt8 = 1<<7 - 1
	// MinInt8 <=> 10000000
	MinInt8 = -1 << 7
	// MaxInt16 <=> 01111111,11111111
	MaxInt16 = 1<<15 - 1
	// MinInt16 <=> 10000000,00000000
	MinInt16 = -1 << 15
	// MaxInt32 <=> 01111111,11111111,11111111,11111111
	MaxInt32 = 1<<31 - 1
	// MinInt32 <=> 10000000,00000000,00000000,00000000
	MinInt32 = -1 << 31
	// MaxInt64 <=> 01111111,11111111,11111111,11111111,11111111,11111111,11111111,11111111
	MaxInt64 = 1<<63 - 1
	// MinInt64 <=> 10000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000
	MinInt64 = -1 << 63
	// MaxUint8 <=> 11111111
	MaxUint8 = 1<<8 - 1
	// MaxUint16 <=> 11111111,11111111
	MaxUint16 = 1<<16 - 1
	// MaxUint32 <=> 11111111,11111111,11111111,11111111
	MaxUint32 = 1<<32 - 1
	// MaxUint64 <=> 11111111,11111111,11111111,11111111,11111111,11111111,11111111,11111111
	MaxUint64 = 1<<64 - 1
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// HexOrDecimal64 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// HexOrDecimal64 类型的底层是uint64类型，
type HexOrDecimal64 uint64

// MarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// 该方法实现了 encoding.TextMarshaler 接口，实现将10进制数转换成16进制，并在前面加上"0x"前缀。
func (h HexOrDecimal64) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%#x", uint64(h))), nil
}

// UnmarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// 该方法实现了 encoding.TextUnmarshaler 接口，该方法实际上是调用 ParseUint64 方法来把给定的字节切片
// 数据解析成64位的无符号整数，支持解析含有"0x"或"0X"前缀的16进制数，也支持解析10进制数。
//
//	例如：给定的string(input)="0x123"，得到整数：291；如果给定的string(input)="123"，得到整数：123。
func (h *HexOrDecimal64) UnmarshalText(input []byte) error {
	i, ok := ParseUint64(string(input))
	if !ok {
		return fmt.Errorf("invalid hex or decimal integer %q", input)
	}
	*h = HexOrDecimal64(i)
	return nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// ParseUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// ParseUint64 方法接受一个字符串参数s，s应该是一个含有"0x"或"0X"前缀的16进制数，或者是一个10进制数，
// 然后 ParseUint64 方法将s解析成一个uint64类型的10进制整数。如果s的格式存在错误，该方法返回的第二个
// 参数将会是false。
//
//	例如：给定的s="0x123"，得到整数：291；如果给定的s="123"，得到整数：123。
func ParseUint64(s string) (uint64, bool) {
	if s == "" {
		return 0, true
	}
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		result, err := strconv.ParseUint(s[2:], 16, 64)
		return result, err == nil
	}
	result, err := strconv.ParseUint(s, 10, 64)
	return result, err == nil
}

// MustParseUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// MustParseUint64 方法接受一个字符串参数s，s应该是一个含有"0x"或"0X"前缀的16进制数，或者是一个10进制数，
// 然后 MustParseUint64 方法将s解析成一个uint64类型的10进制整数。该方法实际上是对 ParseUint64 方法的包
// 装，如果 ParseUint64 的第二个参数返回false，则 MustParseUint64 直接panic。
func MustParseUint64(s string) uint64 {
	result, ok := ParseUint64(s)
	if !ok {
		panic("invalid unsigned 64 bit integer: " + s)
	}
	return result
}

// SafeSub ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// 该方法接受两个64位无符号整数x和y，然后计算result=x-y，返回的result也是64位无符号整数，如果计算结果溢出，
// 则该方法返回的第二个参数等于true。
//
//	例如，给定x=3，y=5，得到输出是(18446744073709551614, true)；再给定x=3，y=2，得到输出是(1, false)。
func SafeSub(x, y uint64) (uint64, bool) {
	diff, overflow := bits.Sub64(x, y, 0)
	return diff, overflow != 0
}

// SafeAdd ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// 该方法接受两个64位无符号整数x和y，然后计算result=x+y，返回的result也是64位无符号整数，如果计算结果溢出，
// 则该方法返回的第二个参数等于true。
//
//	例如，给定x = MaxUint64 - 1，y = 2，得到输出是(1, true)；再给定x = MaxUint64 - 1，y = 1，得到
//	结果是(18446744073709551615, false)。
func SafeAdd(x, y uint64) (uint64, bool) {
	result, overflow := bits.Add64(x, y, 0)
	return result, overflow != 0
}

// SafeMul ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// 该方法接受两个64位无符号整数x和y，然后计算result=x*y，返回的result也是64位无符号整数，如果计算结果溢出，
// 则该方法返回的第二个参数等于true。
//
//	例如，给定x = MaxUint64 / 2，y = 3，得到输出是(9223372036854775805, true)；再给定x = MaxUint64 / 2，
//	y = 2，得到结果是(18446744073709551614, false)。
func SafeMul(x, y uint64) (uint64, bool) {
	overflow, result := bits.Mul64(x, y)
	return result, overflow != 0
}
