package rlp

import (
	"github.com/232425wxy/understanding-ethereum/rlp/internal/rlpstruct"
	"io"
	"reflect"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义 Encoder 接口

// Encoder ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 那些实现 Encoder 接口的类型，可以自定义编码规则。
type Encoder interface {
	EncodeRLP(io.Writer) error
}

var encoderInterface = reflect.TypeOf(new(Encoder)).Elem()

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// listHead ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// listHead 存储了一个列表头的信息，官方源码的写法是"listhead"，可是这在goland编辑器里，会显示波浪线，看着很遭心，
// 所以我改成了"listHead"。
type listHead struct {
	// offset 表明当前编码过后的列表数据的第一个字节在 encBuffer.str 里的索引位置
	offset int
	// size 表示编码列表数据后得到的编码结果的字节长度，例如有一个结构体如下：
	//	type Store struct {
	//		Location string
	//	}
	// 实例化一个Store实例：s := Store{Location: "Hefei"}，结构体会被当作列表进行编码，加上编码头得到的编码结果为：
	// 	[198 133 72 101 102 101 105]
	// 那么此时，size应该等于6，而不是7，因为不能算上编码头"198"
	size int
}

// encodeHead ♏ |作者：吴翔宇| 🍁 |日期：2022/11/1|
//
// encodeHead 方法接受一个字节切片buf作为入参，这个字节切片的长度至少要等于9，官方写法是"encode"，我将其改成了"encodeHead"。
// 由于 listHead 实例一定是在编码列表数据时才会被使用，因此 putHead 方法的第2和第3两个参数应该分别是0xC0和0xF7，该方法的作用
// 就是将 listHead.size 编码到给定的buf切片里，并且只返回编码部分的结果：buf[:size]。
func (lh *listHead) encodeHead(buf []byte) []byte {
	size := putHead(buf, 0xC0, 0xF7, uint64(lh.size))
	return buf[:size]
}

// headSize ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// headSize 方法接受一个整型参数：size，官方的写法是"headsize"，我将其改成了"headSize"，该方法的作用是计算字符串数据或列表数
// 据的头需要占用多少字节空间，传入的参数size分以下2种情况：
//   - 字符串数据或者编码后的列表数据的长度小于56
//   - 字符串数据或者编码后的列表数据的长度大于或等于56
//
// 对于小于56的情况，直接在头的tag（0x80、0xC0）上加上size即可，所以只需要1个字节就可以存储头；对于大于或等于56的情况，我们得先
// 计算需要多少个字节存储size，例如需要n个字节存储size，那么就需要在头的tag（例如0xC0、0xF7）上加上n，这只需要1个字节就够了，然
// 后还需要n个字节存储size，所以总共需要1+n个字节。
func headSize(size uint64) int {
	if size < 56 {
		return 1
	}
	return 1 + intSize(size)
}

// putHead ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 该方法接受4个参数，如下所示：
//   - []byte buf，头部数据会被编码到buf里面
//   - byte smallTag，smallTag的取值有两种：0x80和0xC0，分别对应largeTag的两种取值
//   - byte largeTag，largeTag的取值有两种：0xB7和0xF7，分别对应smallTag的两种取值
//   - uint64 size，size的取值情况分两种，大于或等于56和小于56
//
// putHead 方法的作用是在为某个数据进行编码时，我们需要在编码结果的前面加一个头，来表示头后面跟着多长的数据是对前面数
// 据进行编码后的结果。下面给出几个例子：
//   - 编码的数据是一个长度为32的字符串，那么传入的smallTag和largeTag分别应该等于0x80和0xB7，size等于32，那么编码后的结果为：
//     buf[0] = 0x80 + 32，buf[0] = 160 = 10100000
//   - 编码的数据是一个长度为64的字符串，那么传入的smallTag和largeTag分别应该等于0x80和0xB7，size等于64，那么编码后的结果为：
//     buf[0] = 0xB7 + putInt(buf[1:], size) = 0xB8 = 184，buf[1] = 01000000
//   - 编码一个列表，编码后的数据长度等于36，那么传入的smallTag和largeTag分别应该等于0xCO和0xF7，size等于36，那么编码后的结果为：
//     buf[0] = 0xC0 + 36，buf[0] = 228 = 11100100
//   - 编码一个列表，编码后的数据长度等于456，那么传入的smallTag和largeTag分别应该等于0xCO和0xF7，size等于456，那么编码后的结果为：
//     buf[0] = 0xF7 + putInt(buf[1:], size) = 0xF7 + 2 = 0xF9 = 249，buf[1] = 00000001,11001000
//
// putHead 方法返回的参数表示编码头的大小，即所占的字节数，对于编码长度小于56的字符串，或者编码列表得到长度小于56的编码结果，编码头的
// 大小始终等与1；对于编码长度大于或等于56的字符串，或者编码列表得到长度大于或等于56的编码结果，编码头的大小等于1加上对长度进行大端编码
// 后的长度，即1 + intSize(length)
func putHead(buf []byte, smallTag, largeTag byte, size uint64) int {
	if size < 56 {
		buf[0] = smallTag + byte(size)
		return 1
	}
	sizeSize := putInt(buf[1:], size) // 将size按照大端编码的方式编码到buf中，然后返回所需占用的字节数
	buf[0] = largeTag + byte(sizeSize)
	return sizeSize + 1
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// makeWriter ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// makeWriter 方法接受两个参数，分别是reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，然后为typ生成专属的
// 编码器，其中tag参数只在为元素为非byte类型的切片、数组和指针类型生成编码器时有用。
func makeWriter(typ reflect.Type, tag rlpstruct.Tag) (writer, error) {
	return nil, nil
}

// putInt ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 该方法接受两个参数，第一个参数是一个字节切片b，第二个参数是一个64位无符号整数i，该方法的目的是将i存储到b里面。我们知道，存储
// 一个64位无符号整型数字需要64个比特，即8个字节空间，但是在实际情况里，我们用到的大多数无符号整数都很小，例如1234，如果存储1234
// 这样的数字还用下面这样的8个字节来存储：
//
//	00000000,00000000,00000000,00000000,00000000,00000000,00000100,11010010
//
// 可以发现前6个字节都是0，未免过于浪费存储空间，为此我们设法只存储后面两个可以完全表示数字大小的字节：00000100和11010010，我们
// 把这两个字节的内容按照大端编码的方式存储到b里面，即00000100存储到b[0]里面，11010010存储到b[1]里面，然后 putInt 方法返回的
// 结果表示我们在b中存储i所需的字节树目，在上面的例子里，我们只需要2个字节就可以了，因此返回2。官方源码将此方法写为"putint"，我将
// 其改成了"putInt"。
func putInt(b []byte, i uint64) (size int) {
	switch {
	case i < (1 << 8):
		b[0] = byte(i)
		return 1
	case i < (1 << 16):
		b[0] = byte(i >> 8) // 大端编码，高位字节放在低地址位
		b[1] = byte(i)
		return 2
	case i < (1 << 24):
		b[0] = byte(i >> 16)
		b[1] = byte(i >> 8)
		b[2] = byte(i)
		return 3
	case i < (1 << 32):
		b[0] = byte(i >> 24)
		b[1] = byte(i >> 16)
		b[2] = byte(i >> 8)
		b[3] = byte(i)
		return 4
	case i < (1 << 40):
		b[0] = byte(i >> 32)
		b[1] = byte(i >> 24)
		b[2] = byte(i >> 16)
		b[3] = byte(i >> 8)
		b[4] = byte(i)
		return 5
	case i < (1 << 48):
		b[0] = byte(i >> 40)
		b[1] = byte(i >> 32)
		b[2] = byte(i >> 24)
		b[3] = byte(i >> 16)
		b[4] = byte(i >> 8)
		b[5] = byte(i)
		return 6
	case i < (1 << 56):
		b[0] = byte(i >> 48)
		b[1] = byte(i >> 40)
		b[2] = byte(i >> 32)
		b[3] = byte(i >> 24)
		b[4] = byte(i >> 16)
		b[5] = byte(i >> 8)
		b[6] = byte(i)
		return 7
	default:
		b[0] = byte(i >> 56)
		b[1] = byte(i >> 48)
		b[2] = byte(i >> 40)
		b[3] = byte(i >> 32)
		b[4] = byte(i >> 24)
		b[5] = byte(i >> 16)
		b[6] = byte(i >> 8)
		b[7] = byte(i)
		return 8
	}
}

// intSize ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// intSize 方法接受一个64位的无符号整数作为入参，该方法计算整数i需要多少个字节来存储，该方法的返回值含义和 putInt 方法一样。
// 官方源码的写法是"intsize"，我将其改成了"intSize"。
func intSize(i uint64) int {
	for size := 1; ; size++ {
		if i >>= 8; i == 0 {
			return size
		}
	}
}
