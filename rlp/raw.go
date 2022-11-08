package rlp

import (
	"io"
	"reflect"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// RawValue ♏ |作者：吴翔宇| 🍁 |日期：2022/11/6|
//
// RawValue 官方对其解释是：
//
//	RawValue代表一个已编码的RLP值，可用于延迟RLP解码或预先计算一个编码。请注意，解码器并不验证RawValues的内容是否是有效的RLP。
type RawValue []byte

// rawValueType ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// rawValueType = reflect.TypeOf(RawValue{})
var rawValueType = reflect.TypeOf(RawValue{})

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// ListSize ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// ListSize 方法接受一个64位无符号整型参数contentSize，contentSize表示的是对一个列表进行编码后，除去
// 头剩下的编码内容的长度。这个方法在 Block 和 Transaction 两个结构体实现 DecodeRLP 方法时被调用。
func ListSize(contentSize uint64) uint64 {
	return uint64(headSize(contentSize)) + contentSize
}

// IntSize ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// IntSize 方法接受一个64位无符号整型参数x，该方法的作用是计算对x进行rlp编码后得到的结果的字节长度。对于值
// 小于128的整数，可以将其看成是一个单独的ASCII码，根据rlp编码规则，所以仅需1个字节就可以存储编码结果；如果
// 值大于128，例如1025，1025需要两个字节来存储表示：00000100,00000001，那么编码1025这个数字就需要3个字
// 节：[128+2, 4, 1]->[10000010, 00000100, 00000001]。
func IntSize(x uint64) int {
	if x < 0x80 {
		return 1
	}
	return 1 + intSize(x)
}

// Split ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// Split 方法接受一个字节切片bz作为入参，bz是一个rlp编码结果，Split 方法解析bz，返回被编码的对象是何种类型，
// Byte 、 String 或 List，然后第二个参数返回的是编码结果除编码前缀外剩下的编码内容，第三个参数返回的是其他
// 编码数据，第四个参数返回的是 Split 方法在执行过程中可能遇到的错误，这里只会遇到两种错误，一个是 ErrCanonSize，
// 另一个是 ErrValueTooLarge。
//
//	给一个例子：bz = [204 131 97 97 97 8 198 133 72 101 102 101 105]
//	对给定的bz进行解析，返回的结果将是：List, [131 97 97 97 8 198 133 72 101 102 101 105], [], nil
func Split(bz []byte) (k Kind, content, rest []byte, err error) {
	k, ps, cs, err := readKind(bz)
	if err != nil {
		return 0, nil, nil, err
	}
	return k, bz[ps : ps+cs], bz[ps+cs:], nil
}

// SplitString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// SplitString 与 Split 方法类似，不同的地方在于，SplitString 方法知道自己要解析的对象bz是一个对象字符串进
// 行rlp编码的结果，所以，如果解析得到编码类型不是 String，则会报错。然后该方法的返回值与 Split 的后三个返回值
// 具有相同的含义。
func SplitString(bz []byte) (content, rest []byte, err error) {
	k, content, rest, err := Split(bz)
	if err != nil {
		return nil, bz, err
	}
	if k == List {
		return nil, bz, ErrExpectedString
	}
	return content, rest, nil
}

// SplitUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// SplitUint64 方法接受一个字节切片bz作为入参，该方法的作用就是解析bz里面被rlp编码的整数，需要主义的是：
//  1. 整数0的编码结果是0x80，它表示的是编码头后面有0个字节的内容是为整数0编码的，那么我们在解码的时候，会
//     得到0个字节，也就是len(content)=0，整数1的编码结果就是1，直到整数127的编码结果也是127，也就是说，编
//     码任何非0数字，都至少需要1个字节，那么当编码结果的长度等于1时，编码结果是不可能等于0的。
//  2. 64位无符号整数最多由8个字节组成，超过8个字节的都是不合法的。
//
// 该方法第2个返回值很有意思，它表示的是整数编码内容后面紧跟着的数据，很像协议号的意思，例如我们给定一段数据
// bz=[129 130 97 98 99]，解析结果为：0x82, [97 98 99], nil
func SplitUint64(bz []byte) (x uint64, rest []byte, err error) {
	content, rest, err := SplitString(bz)
	if err != nil {
		return 0, bz, err
	}
	switch {
	case len(content) == 0:
		return 0, rest, nil
	case len(content) == 1:
		if content[0] == 0 {
			// bool类型的false编码结果才是0，而数字0的编码结果是0x80
			return 0, bz, ErrCanonInt
		}
		return uint64(content[0]), rest, nil
	case len(content) > 8:
		// uint64类型的数据所占据的位数不可能大于64位
		return 0, bz, errUintOverflow
	default:
		x, err = readSize(content, byte(len(content)))
		if err != nil {
			return 0, bz, ErrCanonInt
		}
		return x, rest, nil
	}
}

// SplitList ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// SplitList 与 SplitString 方法作用类似，该方法明确知道bz里面含有rlp编码列表部分的内容，如果解析出来
// 的类型显式不是 List，则返回 ErrExpectedList 错误。
func SplitList(bz []byte) (content, rest []byte, err error) {
	kind, content, rest, err := Split(bz)
	if err != nil {
		return nil, bz, err
	}
	if kind != List {
		return nil, bz, ErrExpectedList
	}
	return content, rest, nil
}

// CountValues ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// CountValues 接受一个rlp编码结果bz，该方法的功能就是计算有多少个值被编码进bz里面，例如：
//
//	给定bz=[129 130 12 132 97 97 97 97]，经过计算我们发现有三个值被编码进去了，分别是数字130、数字12
//	以及字符串"aaaa"。
func CountValues(bz []byte) (int, error) {
	i := 0
	for ; len(bz) > 0; i++ {
		_, prefixSize, contentSize, err := readKind(bz)
		if err != nil {
			return 0, err
		}
		bz = bz[prefixSize+contentSize:]
	}
	return i, nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// readKind ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// readKind 方法接受一个参数bz []byte，bz是一个rlp编码数据，bz的第一个字节是编码头，编码头的取值分5个段：
//  1. [0, 127]，表示被编码的对象是一个单独的ASCII码
//  2. [128, 183]，表示被编码的对象是一个长度小于56的字符串
//  3. [184, 191]，表示被编码的对象是一个长度大于55的字符串
//  4. [192, 247]，表示被编码的对象是一个列表，且编码结果的长度小于56
//  5. [248, 255]，表示被编码的对象是一个列表，且编码结果的长度大于55
//
// 根据编码头的取值情况，判断被编码的对象是何种类型，如果被编码的对象是一个ASCII码，则返回 Byte 类型，如果
// 被编码的对象是一个字符串，则返回 String 类型，否则返回 List 类型。该方法返回的第二个参数表示编码前缀所
// 占的字节个数，所谓编码前缀，由编码头和长度编码组成，长度编码只在长度大于55时才会被涉及。该方法返回的第三个
// 参数表示编码结果去除编码前缀后剩下的内容编码的长度，最后一个返回值表示 readKind 方法在执行过程中可能遇到
// 的错误，在这里，只可能出现两种错误，一个是 ErrCanonSize，另一个是 ErrValueTooLarge。
func readKind(bz []byte) (k Kind, prefixSize, contentSize uint64, err error) {
	if len(bz) == 0 {
		return 0, 0, 0, io.ErrUnexpectedEOF
	}
	b := bz[0]
	switch {
	case b < 0x80:
		// 编码内容是单独的ASCII码：[本身]
		k = Byte
		prefixSize = 0
		contentSize = 1
	case b < 0xB8:
		// 长度大于1小于56的字符串：[编码头|内容编码]
		k = String
		prefixSize = 1
		contentSize = uint64(b - 0x80)
		if contentSize == 1 && len(bz) > 1 && bz[1] < 128 {
			// 正常情况下是不会出现这个错误的

			return 0, 0, 0, ErrCanonSize
		}
	case b < 0xC0:
		// 长度大于55的字符串：[编码头|长度编码|内容编码]
		k = String
		prefixSize = 1 + uint64(b-0xB7) // headSize是[编码头|长度编码]所占的字节数
		contentSize, err = readSize(bz[1:], byte(prefixSize-1))
	case b < 0xF8:
		// 编码结果长度大于1小于56的列表：[编码头|内容编码]
		k = List
		prefixSize = 1
		contentSize = uint64(b - 0xC0)
	default:
		// 编码结果长度大于55的列表：[编码头|长度编码|内容编码]
		k = List
		prefixSize = uint64(b-0xF7) + 1
		contentSize, err = readSize(bz[1:], byte(prefixSize-1))
	}
	if err != nil {
		return 0, 0, 0, err
	}
	if contentSize > uint64(len(bz))-prefixSize {
		// 计算出来的内容编码长度过长！
		return 0, 0, 0, ErrValueTooLarge
	}

	return k, prefixSize, contentSize, nil
}

// readSize ♏ |作者：吴翔宇| 🍁 |日期：2022/11/7|
//
// readSize 接受两个参数：bz []byte和length byte，当我们对一个长度大于55的字符串进行rlp编码，或者编码列表得到长度大于
// 55的编码结果，那么完整的编码结果的结构如下：[编码头|长度编码|内容编码]，该方法接受的第一个参数bz存储着[长度编码|内容编码]，
// length参数标记了[长度编码]占用了多少个字节空间，正常情况下，len([长度编码])=length。例如我们编码一个长度为356的字符串数
// 据，356用二进制表示为：00000001,01100100，这个二进制需要两个字节去存储，这两个字节分别是byte(1)和byte(100)，因此，
// [长度编码]其实就是[1, 100]，那么 readSize 的功能就是执行以下步骤：
//
//	byte(1)<<8 | byte(100)
//
// 得到结果356，并返回出去。
func readSize(bz []byte, length byte) (uint64, error) {
	if int(length) > len(bz) {
		return 0, io.ErrUnexpectedEOF
	}
	var s uint64
	switch length {
	case 1:
		s = uint64(bz[0])
	case 2:
		s = uint64(bz[0])<<8 | uint64(bz[1])
	case 3:
		s = uint64(bz[0])<<16 | uint64(bz[1])<<8 | uint64(bz[2])
	case 4:
		s = uint64(bz[0])<<24 | uint64(bz[1])<<16 | uint64(bz[2])<<8 | uint64(bz[3])
	case 5:
		s = uint64(bz[0])<<32 | uint64(bz[1])<<24 | uint64(bz[2])<<16 | uint64(bz[3])<<8 | uint64(bz[4])
	case 6:
		s = uint64(bz[0])<<40 | uint64(bz[1])<<32 | uint64(bz[2])<<24 | uint64(bz[3])<<16 | uint64(bz[4])<<8 | uint64(bz[5])
	case 7:
		s = uint64(bz[0])<<48 | uint64(bz[1])<<40 | uint64(bz[2])<<32 | uint64(bz[3])<<24 | uint64(bz[4])<<16 | uint64(bz[5])<<8 | uint64(bz[6])
	case 8:
		s = uint64(bz[0])<<56 | uint64(bz[1])<<48 | uint64(bz[2])<<40 | uint64(bz[3])<<32 | uint64(bz[4])<<24 | uint64(bz[5])<<16 | uint64(bz[6])<<8 | uint64(bz[7])
	}
	if s < 56 || bz[0] == 0 {
		return 0, ErrCanonSize
	}
	return s, nil
}

// AppendUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// AppendUint64 接受两个参数，第一个参数是一个字节切片bz，第二个参数是一个64位无符号整数i，该方法的目的就是将整数i的rlp编码
// 追加到切片bz之后。例如给定切片bz=[129 137]，给定整数i=45678，执行该方法得到的切片result=[129 137 130 178 110]。
func AppendUint64(b []byte, i uint64) []byte {
	if i == 0 {
		return append(b, 0x80)
	} else if i < 128 {
		return append(b, byte(i))
	}
	switch {
	case i < (1 << 8):
		return append(b, 0x81, byte(i))
	case i < (1 << 16):
		return append(b, 0x82,
			byte(i>>8),
			byte(i),
		)
	case i < (1 << 24):
		return append(b, 0x83,
			byte(i>>16),
			byte(i>>8),
			byte(i),
		)
	case i < (1 << 32):
		return append(b, 0x84,
			byte(i>>24),
			byte(i>>16),
			byte(i>>8),
			byte(i),
		)
	case i < (1 << 40):
		return append(b, 0x85,
			byte(i>>32),
			byte(i>>24),
			byte(i>>16),
			byte(i>>8),
			byte(i),
		)

	case i < (1 << 48):
		return append(b, 0x86,
			byte(i>>40),
			byte(i>>32),
			byte(i>>24),
			byte(i>>16),
			byte(i>>8),
			byte(i),
		)
	case i < (1 << 56):
		return append(b, 0x87,
			byte(i>>48),
			byte(i>>40),
			byte(i>>32),
			byte(i>>24),
			byte(i>>16),
			byte(i>>8),
			byte(i),
		)

	default:
		return append(b, 0x88,
			byte(i>>56),
			byte(i>>48),
			byte(i>>40),
			byte(i>>32),
			byte(i>>24),
			byte(i>>16),
			byte(i>>8),
			byte(i),
		)
	}
}
