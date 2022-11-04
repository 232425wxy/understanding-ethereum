/*
Package math

该文件内定义了两个结构体：HexOrDecimal256和Decimal256，它们的底层实现都是big.Int，定义这两个结构体是为了实现对big.Int进行
marshal/unmarshal，它们俩都各自实现了MarshalText和UnmarshalText方法，HexOrDecimal256和Decimal256不同的地方在于：
  - HexOrDecimal256 的MarshalText方法会将大整数编码成16进制数，并在前面加上"0x"前缀，而 Decimal256 只会在原10进制数
    两边加上双引号得到字符串，然后再将字符串转换成字节切片，更本不会将原大整数转换成16进制，更不会在前面加上"0x"前缀。
  - HexOrDecimal256 的UnmarshalText和 Decimal256 功能一样，都能将含有前缀的16进制或者不含前缀的10进制数据解析成大整数。

需要注意的地方是，无论 HexOrDecimal256 还是 Decimal256，它们所能支持的大整数必须在256比特以内。

随后，该文件还定义了以下方法：

  - func BigPow(a, b int64) *big.Int

  - func BigMax(x, y *big.Int) *big.Int

  - func BigMin(x, y *big.Int) *big.Int

  - func FirstBitSet(i *big.Int) int

  - func ReadBits(bigInt *big.Int, buf []byte)

  - func PaddedBigBytes(bigInt *big.Int, n int) []byte

  - func Byte(bigInt *big.Int, padLen, n int) byte

    U256Bytes 方法将一个给定的大整数（比特位小于等于256）填充为一个含有32个字节的大整数，从而转换成乙太坊虚拟机里的支持的数字

  - func U256Bytes(n *big.Int) []byte

    S256 方法接受一个大整数x作为入参，如果x小于2^255，则直接返回x，否则计算x-2^256，并返回计算结果

  - func S256(x *big.Int) *big.Int

    Exp 方法接受两个大整数base和exponent作为入参，然后计算result=base^exponent，并将result返回

  - func Exp(base, exponent *big.Int) *big.Int
*/
package math

import (
	"fmt"
	"math/big"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

var (
	// tt255 = 2^255
	tt255 = BigPow(2, 255)
	// tt256 = 2^256
	tt256 = BigPow(2, 256)
	// tt256m1 = 2^256-1
	tt256m1 = new(big.Int).Sub(tt256, big.NewInt(1))
	// tt63 = 2^63
	tt63 = BigPow(2, 63)
	// MaxBig256 是256比特位所能表示的最大整数：115792089237316195423570985008687907853269984665640564039457584007913129639935
	MaxBig256 = new(big.Int).Set(tt256m1)
	// MaxBig63 = 2^63-1，该变量在官方的项目源码里并没有被使用
	MaxBig63 = new(big.Int).Sub(tt63, big.NewInt(1))
)

const (
	// WordBits 用来表示一个 big.Word 可以存放多少个比特位，在64位的Ubuntu 20.04操作系统中，wordBits等于64。
	WordBits = 32 << (uint64(^(big.Word(0))) >> 63)
	// WordBytes 用来表示一个 big.Word 最多可以存放多少个字节，在64位的Ubuntu 20.04操作系统中，wordBytes等于8。
	WordBytes = WordBits / 8
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// HexOrDecimal256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// HexOrDecimal256 的底层类型是 big.Int，定义该类型实现了对256位比特以内的大整数进行marshal/unmarshal。
// 实现了 MarshalText 和 UnmarshalText 两个方法，MarshalText 方法将 HexOrDecimal256 编码成16进制的
// 数，并含有 "0x"前缀；UnmarshalText 方法对给定的字节切片内容进行解码，如果给定的字节切片含有前缀，则把给定
// 的字节切片看成是16进制的数，然后解析成10进制的大整数，如果不含前缀，则看成是10进制的数。
//
//	 🚨注意：HexOrDecimal256 作为10进制数，其最大取值是：
//		115792089237316195423570985008687907853269984665640564039457584007913129639935
type HexOrDecimal256 big.Int

// NewHexOrDecimal256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// NewHexOrDecimal256 方法接受一个int64类型的参数x，然后将x转换成一个大整数，最后再将这个大整数通过强制类型转换为 HexOrDecimal256。
func NewHexOrDecimal256(x int64) *HexOrDecimal256 {
	bigInt := big.NewInt(x)
	h := HexOrDecimal256(*bigInt)
	return &h
}

// MarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法实现了 encoding.TextMarshaler 接口，将数字转换为16进制，然后在前面加上"0x"前缀。
//
//	例如：HexOrDecimal256的值等于255，编码后的结果为：[48 120 102 102]，字符串形式："0xff"
//
// 该方法返回的第二个参数永远都是nil。
func (h *HexOrDecimal256) MarshalText() ([]byte, error) {
	if h == nil {
		return []byte("0x0"), nil
	}
	return []byte(fmt.Sprintf("%#x", (*big.Int)(h))), nil
}

// UnmarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法实现了 encoding.TextUnmarshaler 接口，该方法将给定的字符串解析成一个大整数，给定的字符串要么含有"0x"或"0X"前
// 缀，要么不含有，这会决定不同的解析方式：
//  1. 如果给定的字符串含有前缀，则给定的字符串必须满足以下两个条件才能解析成功：
//     - 字符串里除了前缀外，其余字符的取值范围只能在{0 1 2 3 4 5 6 7 8 9 a b c d e f A B C D E F}里
//     - 实际上，尽管给的字符串是16进制的，但是解析后得到大整数是10进制形式的，如果解析后的结果需要超过256位比特去存储，则该
//     方法认为解析失败，默认解析后所能获得的最大整数是：
//     115792089237316195423570985008687907853269984665640564039457584007913129639935
//  2. 如果给定的字符串不含有前缀，则给定的字符串必须满足以下两个条件才能解析成功：
//     - 字符串里每个字符的取值范围只能在{0 1 2 3 4 5 6 7 8 9}里
//     - 解码后得到的大整数必须小于115792089237316195423570985008687907853269984665640564039457584007913129639935
//     给出两个例子：
//     - 输入"0x12a"，得到：298
//     - 输入"123"，得到：123
func (h *HexOrDecimal256) UnmarshalText(input []byte) error {
	bigInt, ok := ParseBig256(string(input))
	if !ok {
		return fmt.Errorf("invalid hex or decimal integer %q", input)
	}
	*h = HexOrDecimal256(*bigInt)
	return nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// ParseBig256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// PareBig256 方法将给定的字符串解析成一个大整数，给定的字符串要么含有"0x"或"0X"前缀，要么不含有，这会决定不同的解析方式：
//  1. 如果给定的字符串含有前缀，则给定的字符串必须满足以下两个条件才能解析成功：
//     - 字符串里除了前缀外，其余字符的取值范围只能在{0 1 2 3 4 5 6 7 8 9 a b c d e f A B C D E F}里
//     - 实际上，尽管给的字符串是16进制的，但是解析后得到大整数是10进制形式的，如果解析后的结果需要超过256位比特去存储，则该
//     方法认为解析失败，默认解析后所能获得的最大整数是：
//     115792089237316195423570985008687907853269984665640564039457584007913129639935
//  2. 如果给定的字符串不含有前缀，则给定的字符串必须满足以下两个条件才能解析成功：
//     - 字符串里每个字符的取值范围只能在{0 1 2 3 4 5 6 7 8 9}里
//     - 解析后得到的大整数必须小于115792089237316195423570985008687907853269984665640564039457584007913129639935
//     给出两个例子：
//     - 输入"0x12a"，得到：298
//     - 输入"123"，得到：123
func ParseBig256(s string) (*big.Int, bool) {
	if s == "" {
		return new(big.Int), true
	}
	var bigInt *big.Int
	var ok bool
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		bigInt, ok = new(big.Int).SetString(s[2:], 16)
	} else {
		bigInt, ok = new(big.Int).SetString(s, 10)
	}
	if ok && bigInt.BitLen() > 256 {
		return nil, false
	}
	return bigInt, ok
}

// MustParseBig256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法实际上就是调用 ParseBig256 方法，如果 ParseBig256 无法解析字符串得到大整数，则直接panic。
func MustParseBig256(s string) *big.Int {
	result, ok := ParseBig256(s)
	if !ok {
		panic("invalid 256 bit integer: " + s)
	}
	return result
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Decimal256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// Decimal256 类型的底层实现是 big.Int，定义该类型是为了实现对大整数的marshal/unmarshal，Decimal256 实现了
// MarshalText 和 UnmarshalText 两个方法，Decimal256 与 HexOrDecimal256 不同的地方在于，Decimal256 的
// MarshalText 方法不会把数字编码成16进制形式，即不会含有"0x"或"0X"前缀；但是 Decimal256 支持对含有前缀的字节
// 切片进行解码，并且解码得到的是10进制的大整数。
type Decimal256 big.Int

// NewDecimal256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// NewDecimal256 方法接受一个int64类型的数字作为入参，然后将其转换为 Decimal256 类型。
func NewDecimal256(x int64) *Decimal256 {
	bigInt := big.NewInt(x)
	d := Decimal256(*bigInt)
	return &d
}

// MarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// MarshalText 方法实现了 encoding.TextMarshaler 接口，直接给数字的两端加上双引号得到字符串，
// 然后将字符串转换为字节切片并返回。
func (d *Decimal256) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法实现了 encoding.TextUnmarshaler 接口，将给定的字节切片转换为十进制的大整数，给定的字节切片需要
// 满足以下两个条件：
//   - 字符串里每个字符的取值范围只能在{0 1 2 3 4 5 6 7 8 9}里
//   - 解码后得到的大整数必须小于115792089237316195423570985008687907853269984665640564039457584007913129639935
func (d *Decimal256) UnmarshalText(input []byte) error {
	bigInt, ok := ParseBig256(string(input))
	if !ok {
		return fmt.Errorf("invalid hex or decimal integer %q", input)
	}
	*d = Decimal256(*bigInt)
	return nil
}

// String ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法返回 Decimal256 的字符串形式。
func (d *Decimal256) String() string {
	if d == nil {
		return "0"
	}
	return fmt.Sprintf("%#d", (*big.Int)(d))
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// BigPow ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法接受两个int64类型的的参数a和b，然后求r=a**b，并将r转换成大整数并返回。
func BigPow(a, b int64) *big.Int {
	r := big.NewInt(a)
	return r.Exp(r, big.NewInt(b), nil)
}

// BigMax ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法接受两个大整数x和y，然后返回这两个数中较大的那一个。
func BigMax(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}
	return x
}

// BigMin ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法接受两个大整数x和y，然后返回这两个数中较小的那一个。
func BigMin(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return x
	}
	return y
}

// FirstBitSet ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// 该方法接受一个大整数作为入参，该方法从给定的大整数i的最低位开始遍历，直到遇到第一个比特位等于1的位置结束，
// 并返回该位置处的比特索引位置（最低有效位），如果找不到，则直接返回该大整数的比特长度。
//
//	例如：给定一个大整数：134348928，它的二进制表现形式是：0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0
//	它的最低有效位则是7。
func FirstBitSet(i *big.Int) int {
	for j := 0; j < i.BitLen(); j++ {
		if i.Bit(j) > 0 {
			return j
		}
	}
	return i.BitLen()
}

// ReadBits ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// ReadBits 接受两个入参，第一个入参是一个大整数bigInt，第二个入参是一个字节切片buf，该方的功能就是将给
// 定的大整数的所有比特位拷贝到给定的字节切片里（从buf的右边开始拷贝），如果给定的字节切片不够长，就拷贝
// len(buf)*8个比特；如果太长，字节切片左边剩下的字节就置为0.
func ReadBits(bigInt *big.Int, buf []byte) {
	i := len(buf)
	for _, word := range bigInt.Bits() {
		for j := 0; j < WordBytes && i > 0; j++ { // 逐个遍历 big.Word 里的每个字节
			i--
			buf[i] = byte(word) // 取 big.Word 最右边的8个比特构成一个字节
			word >>= 8          // 当前 big.Word 右移8个比特位
		}
	}
}

// PaddedBigBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// PaddedBigBytes 该方法接受两个参数，第一个入参是一个大整数bigInt，第二个参数是一个整数n，该方法的目的
// 就是在不改变bigInt大小的情况下，给bigInt填充n-bigInt.BitLen()/8个空字节。
//
//	例如，给定的大整数bigInt=575648，它的字节切片表现形式是：[8 200 160]=[0001000 11001000 10100000]；
//	然后给定的n为5，那么经过填充后，得到的结果是：[0 0 8 200 160]。
func PaddedBigBytes(bigInt *big.Int, n int) []byte {
	if bigInt.BitLen() >= n*8 {
		return bigInt.Bytes()
	}
	result := make([]byte, n)
	ReadBits(bigInt, result)
	return result
}

// bigEndianByteAt ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// bigEndianByteAt 方法接受两个入参，第一个入参是一个大整数bigInt，第二个入参是一个整数n，该方法的目的是获
// 取bigInt中从最低位字节开始第n个字节的值。在这里需要提一下，bigInt里字节的编码模式是大端模式，即高位字节存
// 放在低地址处，第n个字节是从低位字节开始数，第一个字节索引值是0，它是大整数里的最低有效字节，所以当n等于0时，
// 该方法会返回bigInt中最低位的字节，即存放在最高地址位上的字节。
//
//	例如，给定一个大整数bigInt=20441799243135961136544514575630，它的字节切片表现形式为：[1 2 3 4 5 6 7 8 9 10 11 12 13 14]，
//	在这里，存放14这个字节的地址位最高，而14这个字节代表的是bigInt最低有效字节；然后给定的n等于11，执行该方法，
//	返回的值等于byte(3)。
func bigEndianByteAt(bigInt *big.Int, n int) byte {
	words := bigInt.Bits()
	i := n / WordBytes
	if i >= len(words) {
		return byte(0)
	}
	word := words[i]
	shift := 8 * uint(n%WordBytes)
	return byte(word >> shift)
}

// Byte ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// Byte 方法接受三个入参，第一个入参是一个大整数bigInt，第二和第三两个入参分别是padLen和n，需要注意的是padLen这个参数，
// 它表示对原始大整数进行填充后大整数所含有的字节数，这里指的填充可以认为是调用 PaddedBigBytes 方法对大整数进行填充。该
// 方法的目的就是获取bigInt中从最高位字节开始第n个字节的值，与 bigEndianByteAt 这个方法的目的正好相反，实际上，Byte 方
// 法就是通过调用 bigEndianByteAt 方法实现的：return bigEndianByteAt(bigInt, padLen-1-n)
//
//	例如，先定义一个大整数bigInt=4328719365，它的字节切片表现形式是：[1 2 3 4 5]，调用 PaddedBigBytes(bigInt, 17)
//	方法对其进行填充，从bigInt的最低地址位开始填充12个0：[0 0 0 0 0 0 0 0 0 0 0 0 1 2 3 4 5]，然后给定padLen=17和
//	n=13，调用bigEndianByteAt(bigInt, 3)，根据 bigEndianByteAt 的规则从最低有效字节处开始，返回第3个字节，即byte(2)。
func Byte(bigInt *big.Int, padLen, n int) byte {
	if n >= padLen {
		return byte(0)
	}
	return bigEndianByteAt(bigInt, padLen-1-n)
}

// U256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// U256 该方法接受一个入参，该入参是一个大整数x，然后将x与 tt256m1 作与运算，并将结果返回。
//
//		例如：x=5，执行该方法返回：5。
//	 🚨注意：*big.Int 有两个 Add(*big.Int) 方法！！！并且这两个方法的功能不一样！！！
func U256(x *big.Int) *big.Int {
	return x.And(x, tt256m1)
}

// U256Bytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// U256Bytes 方法将一个给定的大整数（比特位小于等于256）填充为一个含有32个字节的大整数，从而转换成乙太坊虚拟机里的支持的数字。
//
//	例如：给定的大整数n=123，执行该方法后得到结果：[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 123]
func U256Bytes(n *big.Int) []byte {
	return PaddedBigBytes(U256(n), 32)
}

// S256 ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// S256 方法接受一个大整数x作为入参，如果x小于2^255，则直接返回x，否则计算x-2^256，并返回计算结果。
//
//	例如：给定的x=2^255+1，执行该方法后得到结果：-57896044618658097711785492504343953926634992332820282019728792003956564819967
func S256(x *big.Int) *big.Int {
	if x.Cmp(tt255) < 0 {
		return x
	}
	return new(big.Int).Sub(x, tt256)
}

// Exp ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// Exp 方法接受两个大整数base和exponent作为入参，然后计算result=base^exponent，并将result返回。
// 如果计算得到的result所表示的数需要超过256个比特位存储，那么仅保留低位的256个比特。
func Exp(base, exponent *big.Int) *big.Int {
	result := big.NewInt(1)
	for _, word := range exponent.Bits() {
		for i := 0; i < WordBytes; i++ {
			if word&1 == 1 {
				U256(result.Mul(result, base))
			}
			U256(base.Mul(base, base))
			word >>= 1
		}
	}
	return result
}
