/*
Package hexutil这个文件定义了如何在uint64和16进制之间进行编解码；定义了如何在大整数和16进制之间进行编解码；
定义了如何在普通字符串或字节切片与16进制之间进行编解码。

乙太坊中定义的16进制数据必定要以"0x"作为前缀，乙太坊规定用16进制最大只能定义256比特位的大整数。
*/
package hexutil

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// uintBits ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// uintBits定义了一个uint类型占用了多少个比特位，一般来讲uint类型的数据要么占据32位，要么占据64位，
// 所以，我们先获取uint类型的最大值：^uint(0)，然后将其右移63个比特位，如果本系统支持的是64位操作系
// 统，那么^uint(0)右移63位应当等于1，否则等于0。
// 下面的代码与[const uintBits = 32 << (^uint(0) >> 63)]等效。
const uintBits = 32 << (uint64(^uint(0)) >> 63)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// decError ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 在解码16进制字符串数据时可能遇到各种各样的错误，decError负责对这些错误进行统一管理。
type decError struct {
	msg string
}

func (err decError) Error() string {
	return err.msg
}

// Errors ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 定义了9个项目全局范围内的decError实例，用来反映在解码16进制字符串时可能遇到的错误。
var (
	// ErrEmptyString 如果给定的16进制数是空的，就是连前缀都不含有，则报告该错误。
	ErrEmptyString = &decError{msg: "empty hex string"}
	// ErrSyntax 16进制数的取值范围是[0, F]，一一例举的话就是：{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, a, b, c, d, e, f}，
	// 不在这个范围内的都会报这个错误。
	ErrSyntax        = &decError{msg: "invalid hex string"}
	ErrMissingPrefix = &decError{msg: "hex string without 0x prefix"}
	// ErrOddLength 16进制数由四个比特表示，表示范围是[0000, 1111]，而一个字节由8个比特组成，因此一个字节可以代表两个16进制
	// 数，也就是说必须两个16进制数才能组成一个字节，所以一个数据被编码成16进制数，那么结果的长度必然是偶数。
	ErrOddLength = &decError{msg: "hex string of odd length"}
	// ErrEmptyNumber 如果给定的16进制数等于0，换句话说就是只有"0x"前缀，则报告该错误，它与ErrEmptyString有一点点不一样。
	ErrEmptyNumber = &decError{msg: "hex string \"0x\""}
	// ErrLeadingZero 如果给定的16进制数不等于0，但是紧跟在前缀"0x"后面的数是"0"，则报告此错误。
	ErrLeadingZero = &decError{msg: "hex number with leading zero digits"}
	// ErrUint64Range 一个64位的无符号整型由8个字节组成
	ErrUint64Range = &decError{msg: "hex number > 64 bits"}
	ErrUintRange   = &decError{msg: fmt.Sprintf("hex number > %d bits", uintBits)}
	ErrBig256Range = &decError{msg: "hex number > 256 bits"}
)

// bigWordNibbles ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// bigWordNibbles定义了一个big.Word可以存储多少个16进制数字，不同的操作系统具有不同结果，
// 我们这里在运行时计算获取一个big.Word能存储多少个16进制数，用一个nibble代表一个16进制数。
var bigWordNibbles int

// init ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 用来计算bigWordNibbles的值等于多少，我们知道^uint64(0)用16进制表示等于"ffffffffffffffff"，由16个"f"组成，换句话说就是，
// ^uint64(0)由16个16进制数组成，在64位操作系统中，如果一个big.Word就可以存储^uint64(0)，则表明一个big.Word可以存储16个16进制数。
//
//	🚨注意：这个函数没有按照官方的方式去实现。
func init() {
	b := new(big.Int).SetUint64(^uint64(0))
	switch len(b.Bits()) {
	case 1:
		bigWordNibbles = 16
	case 2:
		bigWordNibbles = 8
	default:
		panic(fmt.Sprintf("invalid big.Word size %d", len(b.Bits())))
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Decode ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 解码16进制字符串数据，得到字节切片.
//
//	例如：输入"0x43444546"， 解码得到：[67 68 69 70]
func Decode(number string) ([]byte, error) {
	if len(number) == 0 {
		return nil, ErrEmptyString
	}
	if !has0xPrefix(number) {
		return nil, ErrMissingPrefix
	}
	b, err := hex.DecodeString(number[2:])
	if err != nil {
		err = mapError(err)
	}
	return b, err
}

// MustDecode ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 对Decode方法进行了包装，实际上该方法还是调用了Decode方法，然后如果Decode方法返回了错误，
// 则MustDecode方法会直接panic。
func MustDecode(number string) []byte {
	result, err := Decode(number)
	if err != nil {
		panic(err)
	}
	return result
}

// DecodeUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// DecodeUint64将给定的16进制数字（字符串形式，必须带有"0x"或"0X"前缀）解析成10进制数字。
//
//	例如：输入"0x1f"，得到结果：31
func DecodeUint64(number string) (uint64, error) {
	raw, err := checkNumber(number)
	if err != nil {
		return 0, err
	}
	result, err := strconv.ParseUint(raw, 16, 64)
	if err != nil {
		err = mapError(err)
	}
	return result, err
}

// MustDecodeUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 对DecodeUint64方法进行包装，实际上MustDecodeUint64调用DecodeUint64方法解码给定的16进制数字，如果返回错误，则直接panic。
func MustDecodeUint64(number string) uint64 {
	result, err := DecodeUint64(number)
	if err != nil {
		panic(err)
	}
	return result
}

// DecodeBig ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 解码大整数，给定的大整数的16进制字符串如果符合以下5个情况之一，则无法解码：
//  1. 给定的字符串为空：""
//  2. 给定的字符串没有"0x"或"0X"前缀："1234"
//  3. 给定的字符串只含有前缀："0x" 或 "0X"
//  4. 给定的非零16进制数的字符串开头等于0："0x01"
//  5. 去掉前缀后字符串的长度大于64
//
// 根据上述第5个情况所描述的规则，给定的16进制数必须小于^uint256(0)。
//
//	例如：输入"0x123"，得到291
func DecodeBig(number string) (*big.Int, error) {
	raw, err := checkNumber(number)
	if err != nil {
		return nil, err
	}
	if len(raw) > 64 {
		return nil, ErrBig256Range
	}
	words := make([]big.Word, len(raw)/bigWordNibbles+1) // 计算需要多少个big.Word来存储该大整数
	end := len(raw)
	for i := range words {
		start := end - bigWordNibbles
		if start < 0 {
			start = 0
		}
		for j := start; j < end; j++ {
			nibble := decodeNibble(raw[j])
			if nibble == badNibble {
				return nil, ErrSyntax
			}
			words[i] *= 16
			words[i] += big.Word(nibble)
		}
		end = start
	}
	result := new(big.Int).SetBits(words)
	return result, nil
}

// MustDecodeBig ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// MustDecodeBig实际上执行DecodeBig来解码16进制的大数字，DecodeBig如果返回错误，则直接panic。
func MustDecodeBig(number string) *big.Int {
	result, err := DecodeBig(number)
	if err != nil {
		panic(err)
	}
	return result
}

// Encode ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// Encode将给定的数据编码成带有"0x"前缀的16进制数据。
//
//	例如：输入[97 98 99 100]， 输出："0x61626364"
func Encode(bz []byte) string {
	result := make([]byte, len(bz)*2+2)
	copy(result, "0x")
	hex.Encode(result[2:], bz)
	return string(result)
}

// EncodeUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 将64位的无符号整型转换成带有"0x"前缀的16进制数据。
//
//	例如：输入24，得到结果"0x18"；输入7，得到结果"0x7"
func EncodeUint64(number uint64) string {
	result := make([]byte, 2, 10)
	copy(result, "0x")
	return string(strconv.AppendUint(result, number, 16))
}

// EncodeBig ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 对大整数进行编码，得到含有"0x"前缀的16进制数字字符串形式的结果。
//
//	例如输入的大整数为-12，得到结果"-0xc"
func EncodeBig(number *big.Int) string {
	switch sign := number.Sign(); sign {
	case 0:
		return "0x0"
	case 1:
		return "0x" + number.Text(16)
	case -1:
		return "-0x" + number.Text(16)[1:]
	default:
		return ""
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// has0xPrefix ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 判断给定的字符串是否含有前缀"0x"或者"0X"。
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// checkNumber ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 判断给定的16进制数（字符串形式）格式是否合法，以下几种情况皆为不合法：
//  1. 给定的字符串为空：""
//  2. 给定的字符串没有"0x"或"0X"前缀："1234"
//  3. 给定的字符串只含有前缀："0x" 或 "0X"
//  4. 给定的非零16进制数的字符串开头等于0："0x01"
//
// 如果给定的16进制数是合法的，则去掉该数的"0x"前缀，并返回剩下的部分，例如"0x1234"得到"1234"
func checkNumber(number string) (raw string, err error) {
	if len(number) == 0 {
		return "", ErrEmptyString
	}
	if !has0xPrefix(number) {
		return "", ErrMissingPrefix
	}
	withoutPrefix := number[2:]
	if len(withoutPrefix) == 0 {
		return "", ErrEmptyNumber
	}
	if len(withoutPrefix) > 1 && withoutPrefix[0] == '0' {
		return "", ErrLeadingZero
	}
	return withoutPrefix, nil
}

// badNibble ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// badNibble定义了64位无符号整型的最大值，如果在解码16进制数字时，遇到超出[0, f]范围的数字，则用badNibble来替换它原本的值，
// 如果用16进制来表示badNibble的值，它应该等于"ffffffffffffffff"。
const badNibble = ^uint64(0)

// decodeNibble ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 解码单独的一个16进制数字，解码规则如下（区间表示被解码的数字属于哪个范围）：
//
//	['0', '9'] => uint64(x - '0)
//	['a', 'f'] => uint64(x - 'a' + 10)
//	['A', 'F'] => uint64(x - 'A' + 10)
//	其他 => badNibble
func decodeNibble(x byte) uint64 {
	switch {
	case x >= '0' && x <= '9':
		return uint64(x - '0')
	case x >= 'a' && x <= 'f':
		return uint64(x - 'a' + 10)
	case x >= 'A' && x <= 'F':
		return uint64(x - 'A' + 10)
	default:
		return badNibble
	}
}

// mapError ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// mapError将Go定义的错误类型转换成我们自己定义的错误类型，转换规则如下：
//  1. strconv.ErrRange => ErrUint64Range
//  2. strconv.ErrSyntax => ErrSyntax
//  3. hex.InvalidByteError = > ErrSyntax
//  4. hex.ErrLength => ErrOddLength
//  5. 其他 => 原样返回
func mapError(err error) error {
	if err, ok := err.(*strconv.NumError); ok {
		switch err.Err {
		case strconv.ErrRange:
			return ErrUint64Range
		case strconv.ErrSyntax:
			return ErrSyntax
		}
	}
	if _, ok := err.(hex.InvalidByteError); ok {
		return ErrSyntax
	}
	if err == hex.ErrLength {
		return ErrOddLength
	}
	return err
}
