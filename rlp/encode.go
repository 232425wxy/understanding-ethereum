package rlp

import (
	"errors"
	"fmt"
	"github.com/232425wxy/understanding-ethereum/rlp/internal/rlpstruct"
	"io"
	"math/big"
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

// ErrNegativeBigInt ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// ErrNegativeBigInt 被编码的大整数是一个负数时，会报该错误。
var ErrNegativeBigInt = errors.New("rlp: cannot encode negative big.Int")

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
	kind := typ.Kind()
	switch {
	case isUint(kind):
		return writeUint, nil
	case kind == reflect.Ptr:
		// 指针可能是指针的指针，因此我们需要递归地去发现该指针所指向的数据类型
		return makePtrWriter(typ, tag)
	default:
		return nil, fmt.Errorf("rlp: type %v is not RLP-serializable", typ)
	}
}

// writeRawValue ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// writeRawValue 方接受两个参数：即将被编码的 RawValue 对象的 reflect.Value 和一个 *encBuffer 实例，该方法实际上就
// 是将 RawValue 对象本身追加到 *encBuffer.str 后面。
func writeRawValue(val reflect.Value, buf *encBuffer) error {
	buf.str = append(buf.str, val.Bytes()...)
	return nil
}

// writeBigIntPtr ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// writeBigIntPtr 方法接受两个参数：即将被编码的 *big.Int 对象的 reflect.Value 和一个 *encBuffer 实例，注意这里提到的
// *big.Int 不是指针类型。该方法会调用 *encBuffer.writeBigInt 方法将给定的大整数进行编码，如果我们给定的大整数是一个负数，则
// 会报错，另外如果给定的 *big.Int 是一个空指针，则会把该大整数看成是"0"进行编码。
func writeBigIntPtr(val reflect.Value, buf *encBuffer) error {
	ptr := val.Interface().(*big.Int)
	if ptr == nil {
		buf.str = append(buf.str, 0x80)
		return nil
	}
	if ptr.Sign() == -1 {
		return ErrNegativeBigInt
	}
	buf.writeBigInt(ptr)
	return nil
}

// writeBigIntNoPtr ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// writeBigIntNoPtr 方法接受两个参数：即将被编码的 big.Int 对象的 reflect.Value 和一个 *encBuffer 实例，注意这里提到的
// big.Int 不是指针类型。该方法会调用 *encBuffer.writeBigInt 方法将给定的大整数进行编码，如果我们给定的大整数是一个负数，则
// 会报错。
func writeBigIntNoPtr(val reflect.Value, buf *encBuffer) error {
	i := val.Interface().(big.Int)
	if i.Sign() == -1 {
		return ErrNegativeBigInt
	}
	buf.writeBigInt(&i)
	return nil
}

// writeUint ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// writeUint 接受两个参数：uint类型整数的 reflect.Value 和一个 *encBuffer 实例，该方法调用 *encBuffer.writeUint64 方法
// 将给定的整数编码进 *encBuffer.str 里。
func writeUint(val reflect.Value, buf *encBuffer) error {
	buf.writeUint64(val.Uint())
	return nil
}

// writeBool ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// writeBool 方法接受两个参数：bool 的 reflect.Value 和一个 *encBuffer 实例，该方法调用 *encBuffer.writeBool 方法将布尔
// 值编码到 *encBuffer.str 里。
func writeBool(val reflect.Value, buf *encBuffer) error {
	buf.writeBool(val.Bool())
	return nil
}

// writeString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/8|
//
// writeString 方法接受两个参数：string 字符串的 reflect.Value 和一个 *encBuffer 实例，该方法将给定的字符串编码到
// *encBuffer.str 里，例如给定的字符串为"123456789"，则编码结果为：[0x89 '1' '2' '3' '4' '5' '6' '7' '8' '9']。
func writeString(val reflect.Value, buf *encBuffer) error {
	s := val.String()
	if len(s) == 1 && s[0] < 0x80 {
		// 编码单个ASCII码
		buf.str = append(buf.str, s[0])
	} else {
		// 先将字符串的长度编码到 *encBuffer.str 里
		buf.encodeStringHeader(len(s))
		buf.str = append(buf.str, s...)
	}
	return nil
}

// writeBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/9|
//
// writeBytes 方法接受两个参数：字节切片的 reflect.Value 和一个 *encBuffer 实例，该方法将给定的字节切片编码到
// *encBuffer.str 里。
func writeBytes(val reflect.Value, buf *encBuffer) error {
	buf.writeBytes(val.Bytes())
	return nil
}

// writeInterface ♏ |作者：吴翔宇| 🍁 |日期：2022/11/9|
//
// writeInterface 方法接受两个参数：interface{} 的 reflect.Value 和一个 *encBuffer 实例，该方法将某个接口类
// 型数据编码到 *encBuffer.str 里，如果给定的接口数据是nil，则把它当成空列表进行编码。随后得到接口背后的底层数据类型，
// 然后根据类型对数据进行编码。
func writeInterface(val reflect.Value, buf *encBuffer) error {
	if val.IsNil() {
		buf.str = append(buf.str, 0xC0)
		return nil
	}
	// 获取接口背后底层的数据
	eval := val.Elem()
	// 这里使用 cachedWriter 去寻找针对eval的编码器，这样如此，哪怕eval依然是一个接口，也能递归地
	// 到找到其底层的数据类型。
	w, err := cachedWriter(eval.Type())
	if err != nil {
		return err
	}
	return w(eval, buf)
}

// makePtrWriter ♏ |作者：吴翔宇| 🍁 |日期：2022/11/9|
//
// makePtrWriter 方法接受两个参数：指针类型的 reflect.Type 和一个 rlpstruct.Tag 实例，该方法就是为一个指针类型的
// 数据生成一个编码器。下面给一个例子：
//
//	给一个指针的指针：ptrptr = **uint(23)，我们现在尝试获取针对ptrptr的编码器，首先我们调用ptrptr.Elem()获取它指向
//	的第一层数据类型ptr，是*uint64，它还是个指针，此时，我们会继续获取ptr所指向的第二层数据类型（此处的逻辑由
//	infoWhileGenerating 方法实现），得到的数据类型是uint64，那么最终我们确定了针对ptrptr的编码器其实就是 writeUint。
//	那么最终的编码结果就是[23]。
//
// 如果上面举的例子中ptrptr所指向的指针等于nil，则value.Elem().IsValid()会等于false。
func makePtrWriter(typ reflect.Type, tag rlpstruct.Tag) (writer, error) {
	nilEncoding := byte(0xC0)
	if typeNilKind(typ.Elem(), tag) == String {
		nilEncoding = 0x80
	}
	// 递归地调用去发现指针所指向的数据类型
	info := theTC.infoWhileGenerating(typ.Elem(), rlpstruct.Tag{})
	if info.writerErr != nil {
		return nil, info.writerErr
	}
	var w writer = func(value reflect.Value, buffer *encBuffer) error {
		if ev := value.Elem(); ev.IsValid() {
			return info.writer(ev, buffer)
		}
		buffer.str = append(buffer.str, nilEncoding)
		return nil
	}
	return w, nil
}

// makeEncoderWriter ♏ |作者：吴翔宇| 🍁 |日期：2022/11/9|
//
// makeEncoderWriter

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
