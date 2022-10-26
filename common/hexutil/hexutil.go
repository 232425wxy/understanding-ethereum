package hexutil

import "fmt"

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
//  2. 给定的字符串没有前缀："1234"
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
	withoutPrefix := number[:2]
	if len(withoutPrefix) == 0 {
		return "", ErrEmptyNumber
	}
	if len(withoutPrefix) > 1 && withoutPrefix[0] == '0' {
		return "", ErrLeadingZero
	}
	return withoutPrefix, nil
}
