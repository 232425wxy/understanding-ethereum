package common

import (
	"encoding/hex"
	"errors"
	"github.com/232425wxy/understanding-ethereum/common/hexutil"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API

// FromHex ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// FromHex 方法接受一个16进制编码的字符串作为入参，该字符串可能含有"0x"或"0X"前缀，如果有的话，
// 则将其去除掉，然后将剩余的部分利用 hex.DecodeString 方法解码成字符串，最终返回结果的字节切
// 片形式。
func FromHex(s string) []byte {
	if has0xPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// CopyBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// CopyBytes 接受一个字节切片作为输入参数，然后复制这个字节切片，并将复制品返回出去。
func CopyBytes(bz []byte) (cpy []byte) {
	if bz == nil {
		return nil
	}
	cpy = make([]byte, len(bz))
	copy(cpy, bz)
	return cpy
}

// Hex2Bytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// Hex2Bytes 方法接受一个16禁止编码的字符串，该字符串不能含有"0x"或"0X"前缀，然后将给定的字符串
// 利用 hex.DecodeString 方法解码成字符串，最终返回结果的字节切片形式。
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// Bytes2Hex ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// Bytes2Hex 方法接受一个字节切片作为输入参数，然后调用 hex.EncodeToString 方法将该字节切片编码成
// 16进制的字符串，例如将"Hello"编码成"48656c6c6f"。
func Bytes2Hex(bz []byte) string {
	return hex.EncodeToString(bz)
}

// ParseHexOrString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// ParseHexOrString 方法接受一个字符串作为输入参数，如果输入的字符串不含有"0x"或"0X"前缀，则返回输入
// 的字符串本身，否则调用 hex.DecodeString 方法去解码前缀之后的字符串。
func ParseHexOrString(str string) ([]byte, error) {
	bz, err := hexutil.Decode(str)
	if errors.Is(err, hexutil.ErrMissingPrefix) {
		return []byte(str), nil
	}
	return bz, err
}

// RightPadBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// RightPadBytes 方法接受两个参数，第一个参数是一个字节切片，第二个参数是一个整数l，该方法的目的就是将字节
// 切片的长度扩展成l，右边新增的字节将用0来填充。
func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

// LeftPadBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// LeftPadBytes 方法接受两个参数，第一个参数是一个字节切片，第二个参数是一个整数l，该方法的目的就是将字节
// 切片的长度扩展成l，左边新增的字节将用0来填充。
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

// TrimLeftZeroes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// TrimLeftZeroes 接受一个字节切片作为输入参数，该方法的目的就是将给定的字节切片左边的所有0给去掉，然后返回
// 右边剩下的字节切片。
func TrimLeftZeroes(s []byte) []byte {
	idx := 0
	for ; idx < len(s); idx++ {
		if s[idx] != 0 {
			break
		}
	}
	return s[idx:]
}

// TrimRightZeroes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// TrimRightZeroes 接受一个字节切片作为输入参数，该方法的目的就是将给定的字节切片右边的所有0给去掉，然后返回
// 左边剩下的字节切片。
func TrimRightZeroes(s []byte) []byte {
	idx := len(s)
	for ; idx > 0; idx-- {
		if s[idx-1] != 0 {
			break
		}
	}
	return s[:idx]
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的函数

// has0xPrefix ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// has0xPrefix 方法接受一个字符串作为入参，并判断该字符串是否含有"0x"或"0X"前缀。
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// isHexCharacter ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// isHexCharacter 方法接受一个字节作为输入参数，然后判断该字节是不是合法的16进制字节，16进制字节的取值范围如下所示：
//
//	'0' ~ '9' | 'a' ~ 'f' | 'A' ~ 'F'
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

// isHex ♏ |作者：吴翔宇| 🍁 |日期：2022/11/19|
//
// isHex 方法接受一个字符串作为输入参数，该方法的功能是判断字符串里的每个字节是否都在16进制字节的取值范围之内，同时还
// 需要判断给定的字符串的长度是否为偶数，因为16进制编码的内容，两个字节代表一个字节。例如字节'00010010'(18)的16进制
// 编码结果为：'00000001 00000010'(1 2)。
func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}
