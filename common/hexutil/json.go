package hexutil

import "encoding/hex"

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Bytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 我们自己定义了一个切片类型Bytes，它的底层实现就是[]byte，定义该类型的目的是方便对字节切片进行编解码。
type Bytes []byte

// MarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 该方法实现了 encoding.TextMarshaler 接口，该方法的作用就是将字节切片转换为16进制数据，
// 我们知道一个字节可以代表两个16进制数，所以转换为16进制数据后，长度会扩大一倍，只是在此
// 基础上，我们还要在转换后的数据前加上`0x`前缀，所以长度还要再加2。
func (b Bytes) MarshalText() ([]byte, error) {
	result := make([]byte, len(b)*2+2)
	copy(result, `0x`)
	hex.Encode(result[2:], b) // hex包有它自己的编码规则
	return result, nil
}

// UnmarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// UnmarshalText方法实现了encoding.TextUnmarshaler接口，该方法将MarshalText的编码结果解码成原始数据。

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// bytesHave0xPrefix ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 判断给定的字节切片是否含有"0x"或"0X"前缀，该方法相当于 has0xPrefix(string(input)) 方法的效果。
func bytesHave0xPrefix(input []byte) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

// checkText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 对给定的字节切片进行检查，检查分两种情况进行：
//	1. 如果第二个输入参数的值是true，则当给定的字节切片符合以下情况之一，checkText 方法就会报错：
//		- 给定的字节切片不含有"0x"或者"0X"前缀，例如：['a', 'b', 'c', 'd']
//		- 给定的字节切片长度是奇数，例如：['a', 'b', 'c']
//	2. 如果第二个输入参数的值是false，则当给定的字节切片符合以下情况，checkText 方法就会报错：
//		- 给定的字节切片长度是奇数，例如：['a', 'b', 'c']
// 如果给定的切片合法，并且假如给定的切片含有"0x"或者"0X"前缀，则将其前缀去掉并返回，否则不做任何改变直接返回该切片。
func checkText(input []byte, wantPrefix bool) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}
	if bytesHave0xPrefix(input) {
		input = input[2:]
	} else if wantPrefix { // 没有前缀，但是又想有前缀
		return nil, ErrMissingPrefix
	}
	if len(input)%2 != 0 {
		return nil, ErrOddLength
	}
	return input, nil

}