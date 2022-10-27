/*
Package hexutil
该文件里定义了以下三种自定义类型：
  - type Bytes []byte
  - type Big big.Int
  - type Uint64 uint64
  - type Uint uint

每个类型都为其实现了 MarshalText、UnmarshalText、UnmarshalJSON方法，
UnmarshalText与UnmarshalJSON之间的关系为：当我们调用json.Unmarshal去解码数据时，如果给定的指针所代表的数据类型实现了
UnmarshalJSON方法，则会调用该类型自定义的UnmarshalJSON方法进行解码；否则如果给定的指针所代表的数据类型实现了UnmarshalText
方法，并且需要解码的数据被双引号包围，则会调用该类型自定义的UnmarshalText方法进行解码（解码的时候会把引号去掉）。

其中，Bytes、Big和Uint64三个类型还实现了 ImplementsGraphQLType 和 UnmarshalGraphQL 两个方法。
*/
package hexutil

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 三个API函数

// UnmarshalFixedJSON ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// UnmarshalFixedJSON 方法接收3个入参，其中第一个入参是要解码的数据类型，第二个入参是编码数据，第三个参数是接收解码结果的一个字节切片，
// 第二个入参和第三个入参都是字节切片类型的，UnmarshalFixedJSON 对这两个参数具有如下要求：
//  1. 首先，第二个入参是编码数据，要求这个编码数据必须被双引号包围，并且必须含有"0x"或"0X"前缀
//  2. 其次，要求第二个参数除前缀外，剩下的部分的长度值必须是偶数
//  3. 最后，第三个参数作为接收解码结果的一个容器，编码数据是16进制形式的，所以要求第三个参数的切片长度必须等于第二个参数去掉前缀后，剩下
//     部分长度的一半，即：len(out) = len(input[1:len(input)-1]。
func UnmarshalFixedJSON(typ reflect.Type, input, out []byte) error {
	if !isString(input) {
		return errNonString(typ)
	}
	err := UnmarshalFixedText(typ.String(), input[1:len(input)-1], out)
	return wrapTypeError(err, typ)
}

// UnmarshalFixedText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// UnmarshalFixedText 方法接收3个入参，其中第一个入参是要解码的数据类型，第二个入参是编码数据，第三个参数是接收解码结果的一个字节切片，
// 第二个入参和第三个入参都是字节切片类型的，UnmarshalFixedText 对这两个参数具有如下要求：
//  1. 首先，第二个入参是编码数据，要求这个编码数据必须含有"0x"或"0X"前缀
//  2. 其次，要求第二个参数除前缀外，剩下的部分的长度值必须是偶数
//  3. 最后，第三个参数作为接收解码结果的一个容器，编码数据是16进制形式的，所以要求第三个参数的切片长度必须等于第二个参数去掉前缀后，剩下
//     部分长度的一半，即：len(out) = len(input[1:len(input)-1]。
func UnmarshalFixedText(typName string, input, out []byte) error {
	raw, err := checkText(input, true)
	if err != nil {
		return nil
	}
	if len(raw)/2 != len(out) {
		return fmt.Errorf("hex string has length %d, want %d for %s", len(raw), len(out)*2, typName)
	}
	for _, b := range raw {
		if decodeNibble(b) == badNibble {
			return ErrSyntax
		}
	}
	_, err = hex.Decode(out, raw)
	return err
}

// UnmarshalFixedUnPrefixedText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// UnmarshalFixedUnPrefixedText 方法接收3个入参，其中第一个入参是要解码的数据类型，第二个入参是编码数据，第三个参数是接收解码结果的
// 一个字节切片，第二个入参和第三个入参都是字节切片类型的，UnmarshalFixedUnPrefixedText 对这两个参数具有如下要求：
//  1. 首先，第二个入参是编码数据，这个编码数据不用必须含有"0x"或"0X"前缀
//  2. 其次，如果第二个参数含有前缀，要求其除前缀外，剩下的部分的长度值必须是偶数
//  3. 最后，第三个参数作为接收解码结果的一个容器，编码数据是16进制形式的，（如果第二个参数含有前缀）所以要求第三个参数的切片长度必须等于
//     第二个参数去掉前缀后，剩下部分长度的一半，即：len(out) = len(input[1:len(input)-1]。
func UnmarshalFixedUnPrefixedText(typName string, input, out []byte) error {
	raw, err := checkText(input, false)
	if err != nil {
		return err
	}
	if len(raw)/2 != len(out) {
		return fmt.Errorf("hex string has length %d, want %d for %s", len(raw), len(out)*2, typName)
	}
	for _, b := range raw {
		if decodeNibble(b) == badNibble {
			return ErrSyntax
		}
	}
	_, err = hex.Decode(out, raw)
	return err
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

var (
	bytesT  = reflect.TypeOf(Bytes(nil))
	bigT    = reflect.TypeOf((*Big)(nil))
	uint64T = reflect.TypeOf(Uint64(0))
	uintT   = reflect.TypeOf(Uint(0))
)

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
// UnmarshalText方法实现了encoding.TextUnmarshaler接口，该方法将 MarshalText 的编码结果解码成原始数据,
// 由于是 MarshalText 的编码结果，所以给定的输入参数必然要含有"0x"前缀，不然会直接报错。
func (b *Bytes) UnmarshalText(input []byte) error {
	raw, err := checkText(input, true)
	if err != nil {
		return err
	}
	result := make([]byte, len(raw)/2)
	if _, err = hex.Decode(result, raw); err != nil {
		err = mapError(err)
	} else {
		*b = result
	}
	return err
}

// UnmarshalJSON ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 json.Unmarshaler 接口，该方法将给定的字节切片解码成 Bytes，但是给定的字节切片需要满足以下条件：
//  1. 切片的两端必须是引号'"'
//  2. 切片左边的引号之后必须紧跟"0x"或"0X"前缀
//
// 该方法实际上是调用 Bytes 的 UnmarshalText 方法对去掉两端引号后的字节切片进行解码。
func (b *Bytes) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(bytesT)
	}
	err := b.UnmarshalText(input[1 : len(input)-1])
	return wrapTypeError(err, bytesT)
}

// ImplementsGraphQLType ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// ImplementsGraphQLType 方法的输入参数如果是"Bytes"，则该方法返回true。
// 第三方库github.com/graph-gophers/graphql-go对该方法的解释是：
//
//	ImplementsGraphQLType将实现的自定义Go类型映射到模式中的GraphQL标量类型。
func (b Bytes) ImplementsGraphQLType(name string) bool {
	return name == "Bytes"
}

// UnmarshalGraphQL ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// UnmarshalGraphQL 方法的输入参数是一个interface{}，如果该参数的实际类型是string，则调用 Decode 方法对该参数
// 进行解码，并将得到的结果赋值给该方法的接收器 Bytes。
//
//	🚨注意：由于它调用 Decode 方法进行解码，所以要求输入的字符串参数必须含有"0x"或"0X"前缀，否则会报错。
//
// 第三方库github.com/graph-gophers/graphql-go对该方法的解释是：
//
//	UnmarshalGraphQL是实现类型的自定义unmarshaler，每当你使用自定义GraphQL标量类型作为输入时，就会调用这个函数。
func (b *Bytes) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		data, err := Decode(input)
		if err != nil {
			return err
		}
		*b = data
	default:
		err = fmt.Errorf("unexpected type %T for Bytes", input)
	}
	return err
}

// String ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 输出Bytes的字符串表现形式，将给定的数据编码成带有"0x"前缀的16进制数据。
func (b Bytes) String() string {
	return Encode(b)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Big ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 我们自定义的Big类型，其底层就是big.Int，自定义一个Big类型方便我们对大整数进行marshal/unmarshal，
// 大整数里的0会被编码成"0x0"。不支持对负数进行unmarshal，因为负数的编码结果的表现形式是"-0x..."这样
// 的，所以在利用 checkNumberText 方法验证需要编码数据是否具有"0x"或"0X"前缀时，验证结果会显示不存在
// 前缀，因为这里的前缀等于"-0"。比特位数大于256位的大整数在解码时会报错，但是在编码时不会报错。
type Big big.Int

// MarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 encoding.TextMarshaler 接口，对大整数进行编码，得到含有"0x"前缀的16进制数字字符串，
// 然后返回该字符串的字节切片形式，返回的第二个参数永远都是nil。
func (b Big) MarshalText() ([]byte, error) {
	big := (*big.Int)(&b)
	return []byte(EncodeBig(big)), nil
}

// UnmarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 encoding.TextUnmarshaler 接口，将给定的带有"0x"或"0X"前缀的字节切片数据解码成大整数。
//
//	 🚨注意：给定的字节切片必须含有前缀！否则会报错。另外去掉前缀后的字节切片的长度不能超过64，不然也会报错，
//		因为我们无法对比特位数超过256位的大整数进行解码。
func (b *Big) UnmarshalText(input []byte) error {
	raw, err := checkNumberText(input)
	if err != nil {
		return err
	}
	if len(raw) > 64 {
		return ErrBig256Range
	}
	words := make([]big.Word, len(raw)/bigWordNibbles+1)
	end := len(raw)
	for i := range words {
		start := end - bigWordNibbles
		if start < 0 {
			start = 0
		}
		for j := start; j < end; j++ {
			nibble := decodeNibble(raw[j])
			if nibble == badNibble {
				return ErrSyntax
			}
			words[i] *= 16
			words[i] += big.Word(nibble)
		}
		end = start
	}
	var result big.Int
	result.SetBits(words)
	*b = Big(result)
	return nil
}

// UnmarshalJSON ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 json.Unmarshaler 接口，该方法将给定的字节切片解码成 Big，但是给定的字节切片需要满足以下条件：
//  1. 切片的两端必须是引号'"'
//  2. 切片左边的引号之后必须紧跟"0x"或"0X"前缀
//
// 实际上该方法是调用 Big 的 UnmarshalText 方法对去掉两端引号后的字节切片进行解码。
func (b *Big) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(bigT)
	}
	err := b.UnmarshalText(input[1 : len(input)-1])
	return wrapTypeError(err, bigT)
}

// ImplementsGraphQLType ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// ImplementsGraphQLType 方法的输入参数如果是"BigInt"，则该方法返回true。
// 第三方库github.com/graph-gophers/graphql-go对该方法的解释是：
//
//	ImplementsGraphQLType将实现的自定义Go类型映射到模式中的GraphQL标量类型。
func (b Big) ImplementsGraphQLType(name string) bool {
	return name == "BigInt"
}

// UnmarshalGraphQL ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// UnmarshalGraphQL 方法的输入参数是一个interface{}，如果该参数的实际类型是string，则调用 Big 的 UnmarshalText 方法对该参数
// 的字节切片进行解码，并将得到的结果赋值给该方法的接收器 Big；如果参数的实际类型是int32，则调用 big.Int 的 SetInt64 方法将该参数
// 赋值给该方法的接收器 Big。如果输入的参数不是以上两种类型中的其中之一，则返回错误。
// 第三方库github.com/graph-gophers/graphql-go对该方法的解释是：
//
//	UnmarshalGraphQL是实现类型的自定义unmarshaler，每当你使用自定义GraphQL标量类型作为输入时，就会调用这个函数。
func (b *Big) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return b.UnmarshalText([]byte(input))
	case int32:
		var num big.Int
		num.SetInt64(int64(input))
		*b = Big(num)
	default:
		err = fmt.Errorf("unexpected type %T for BigInt", input)
	}
	return err
}

// ToInt ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// ToInt 方法将 *Big 转换成 *big.Int。
func (b *Big) ToInt() *big.Int {
	return (*big.Int)(b)
}

// String ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// String 方法返回 Big 的字符串形式，实际上，就是对大整数进行编码，得到含有"0x"前缀的16进制数字字符串形式的结果。
func (b *Big) String() string {
	return EncodeBig(b.ToInt())
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Uint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 我们自定义了一个Uint64类型，它的底层实现其实就是Go内置的uint64类型，定义Uint64是为了方便对64位
// 无符号整型进行marshal/unmarshal，0会被编码成"0x0"。
type Uint64 uint64

// MarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 encoding.TextMarshaler 接口，该方法实际上是调用 strconv.AppendUInt 方法对 Uint64
// 进行编码，编码的规则是：先将 Uint64 转换成16进制，如75被转换成4b，然后用ASCII码字符对应的编码逐个替换
// 4b里的4和b，4在ASCII码里对应的编码是52，b在ASCII码里对应的编码是98，所以最终75被转换为[52 98]，在此
// 基础上，我们还要在转换的结果前加上"0x"前缀，0在ASCII码里对应的编码是48，x在ASCII码里对应的编码是120，
// 所以，如果该方法的接收器的值是75，调用 UnmarshalText 方法得到的结果将是[48 120 52 98]。
func (i Uint64) MarshalText() ([]byte, error) {
	// 64位无符号整型最多需要8个字节的存储空间
	result := make([]byte, 2, 10) // 还要再加上两字节的前缀
	copy(result, "0x")
	result = strconv.AppendUint(result, uint64(i), 16)
	return result, nil
}

// UnmarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 encoding.TextUnmarshaler 接口，该方法就是将 Uint64 的 MarshalText 方法的编码结果再
// 解码成 Uint64。
//
//	🚨注意：该方法要求传入的字节切片参数必须含有"0x"或"0X"前缀。
func (i *Uint64) UnmarshalText(input []byte) error {
	raw, err := checkNumberText(input)
	if err != nil {
		return err
	}
	if len(raw) > 16 {
		return ErrUint64Range
	}
	var result uint64
	for _, b := range raw {
		nibble := decodeNibble(b)
		if nibble == badNibble {
			return ErrSyntax
		}
		result *= 16
		result += nibble
	}
	*i = Uint64(result)
	return nil
}

// UnmarshalJSON ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 json.Unmarshaler 接口，该方法将给定的字节切片解码成 Uint64，但是给定的字节切片需要满足以下条件：
//  1. 切片的两端必须是引号'"'
//  2. 切片左边的引号之后必须紧跟"0x"或"0X"前缀
//
// 实际上该方法是调用 Uint64 的 UnmarshalText 方法对去掉两端引号后的字节切片进行解码。
func (i *Uint64) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(uint64T)
	}
	err := i.UnmarshalText(input[1 : len(input)-1])
	return wrapTypeError(err, uint64T)
}

// ImplementsGraphQLType ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// ImplementsGraphQLType 方法的输入参数如果是"Long"，则该方法返回true。
// 第三方库github.com/graph-gophers/graphql-go对该方法的解释是：
//
//	ImplementsGraphQLType将实现的自定义Go类型映射到模式中的GraphQL标量类型。
func (i Uint64) ImplementsGraphQLType(name string) bool {
	return name == "Long"
}

// UnmarshalGraphQL ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// UnmarshalGraphQL 方法的输入参数是一个interface{}，如果该参数的实际类型是string，则调用 Uint64 的 UnmarshalText 方法对
// 该参数的字节切片进行解码，并将得到的结果赋值给该方法的接收器 Uint64；如果参数的实际类型是int32，则将该参数强制类型转换成 Uint64。
// 如果输入的参数不是以上两种类型中的其中之一，则返回错误。
// 第三方库github.com/graph-gophers/graphql-go对该方法的解释是：
//
//	UnmarshalGraphQL是实现类型的自定义unmarshaler，每当你使用自定义GraphQL标量类型作为输入时，就会调用这个函数。
func (i *Uint64) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		return i.UnmarshalText([]byte(input))
	case int32:
		*i = Uint64(input)
	default:
		err = fmt.Errorf("unexpected type %T for Long", input)
	}
	return err
}

// String ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// String 方法返回 Uint64 的字符串形式，实际上就是将64位的无符号整型转换成带有"0x"前缀的16进制数据。
//
//	例如：Uint64 的实例是24，得到结果"0x18"；Uint64 的实例是7，得到结果"0x7"
func (i Uint64) String() string {
	return EncodeUint64(uint64(i))
}

// Uint ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// Uint 是我们自定义的一个数据类型，其底层实现其实就是Go内置的uint，定义 Uint 是为
// 了方便对uint进行marshal/unmarshal，0会被编码成"0x0"。
type Uint uint

// MarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 encoding.TextMarshaler 接口，实际上是将 Uint 强制类型转换成 Uint64，
// 然后调用 Uint64 的 MarshalText 方法对无符号整数进行编码。
func (i Uint) MarshalText() ([]byte, error) {
	return Uint64(i).MarshalText()
}

// UnmarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 encoding.TextUnmarshaler 接口，该方法是对 Uint 的 MarshalText 方法产生的编码结果进行解码，
// 由于 Uint 的 MarshalText实际上是通过调用 Uint64 的 MarshalText 方法实现的，所以 Uint 的 UnmarshalText
// 方法也是通过调用 Uint64 的 UnmarshalText 方法实现的。
func (i *Uint) UnmarshalText(input []byte) error {
	var result Uint64
	err := result.UnmarshalText(input)
	if result > Uint64(^uint(0)) || err == ErrUint64Range {
		// 实际上在64位的Ubuntu 20.04操作系统中，^uint64(0) = ^uint(0)
		return ErrUintRange
	} else if err != nil {
		return err
	}
	*i = Uint(result)
	return nil
}

// UnmarshalJSON ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法实现了 json.Unmarshaler 接口，该方法将给定的字节切片解码成 Uint，但是给定的字节切片需要满足以下条件：
//  1. 切片的两端必须是引号'"'
//  2. 切片左边的引号之后必须紧跟"0x"或"0X"前缀
//
// 实际上该方法是调用 Uint 的 UnmarshalText 方法对去掉两端引号后的字节切片进行解码。
func (i *Uint) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(uintT)
	}
	err := i.UnmarshalText(input[1 : len(input)-1])
	return wrapTypeError(err, uintT)
}

// String ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// String 方法返回 Uint 的字符串形式，实际上是先将 Uint 强制类型转换为 uint64，然后利用 EncodeUint64 方法对其进行编码，
// 将64位的无符号整型转换成带有"0x"前缀的16进制数据。
//
//	例如：输入24，得到结果"0x18"；输入7，得到结果"0x7"
func (i Uint) String() string {
	return EncodeUint64(uint64(i))
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// isString ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 该方法的逻辑就是判断给定的输入参数是否是字符串，判断的依据是检查给定的字节切片的第一个字节和最后一个字节是否是'"'，如果是，
// 则说明是字符串，否则就不是。
func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

// bytesHave0xPrefix ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 判断给定的字节切片是否含有"0x"或"0X"前缀，该方法相当于 has0xPrefix(string(input)) 方法的效果。
func bytesHave0xPrefix(input []byte) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

// checkText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/26|
//
// 对给定的字节切片进行检查，检查分两种情况进行：
//  1. 如果第二个输入参数的值是true，则当给定的字节切片符合以下情况之一，checkText 方法就会报错：
//     - 给定的字节切片不含有"0x"或者"0X"前缀，例如：['a', 'b', 'c', 'd']
//     - 给定的字节切片长度是奇数，例如：['0', 'x', 'a', 'b', 'c']
//  2. 如果第二个输入参数的值是false，则当给定的字节切片符合以下情况，checkText 方法就会报错：
//     - 给定的字节切片长度是奇数，例如：['a', 'b', 'c']
//
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

// checkNumberText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 检查16进制数字字节切片形式的格式是否正确，如果给定的字节切片符合以下情况之一，则会报错：
//  1. 给定的字节切片不含有"0x"或"0X"前缀：['a', 'b', 'c', 'd']
//  2. 给定的字节切片仅仅只含有前缀"0x"或"0X"：['0', 'x']
//  3. 给定的16进制数非零，但是紧跟在前缀后面的值等于'0'：['0', 'x', '0', '1']
//
// 如果通过格式检查，则返回去掉前缀的字节切片。
func checkNumberText(input []byte) (raw []byte, err error) {
	if len(input) == 0 {
		return nil, nil
	}
	if !bytesHave0xPrefix(input) {
		return nil, ErrMissingPrefix
	}
	input = input[2:]
	if len(input) == 0 {
		return nil, ErrEmptyNumber
	}
	if len(input) > 1 && input[0] == '0' {
		return nil, ErrLeadingZero
	}
	return input, nil
}

// wrapTypeError ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 给定的错误如果是*decError的实例，则将其包装成*json.UnmarshalTypeError并返回，否返将给定的错误原封不动的返回。
func wrapTypeError(err error, typ reflect.Type) error {
	if _, ok := err.(*decError); ok {
		return &json.UnmarshalTypeError{Value: err.Error(), Type: typ}
	}
	return err
}

// errNonString ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// 返回一个*json.UnmarshalTypeError实例，这个错误的提示信息为"non-string"，表示”非字符串“的意思。
func errNonString(typ reflect.Type) error {
	return &json.UnmarshalTypeError{Value: "non-string", Type: typ}
}
