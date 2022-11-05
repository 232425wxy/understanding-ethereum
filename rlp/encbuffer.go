package rlp

import (
	"github.com/232425wxy/understanding-ethereum/common/math"
	"io"
	"math/big"
	"reflect"
	"sync"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// encBuffer ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// encBuffer 结构体被用于在编码数据时存储编码结果。
type encBuffer struct {
	str          []byte     // str 包含了除列表头之外的所有编码信息
	lHeads       []listHead // 存储了所有列表头信息，官方源码的写法是"lheads"
	lHeadsSize   int        // 官方源码写法是"lhsize"，表示所有头加一起的长度
	auxiliaryBuf [9]byte    // 官方源码写法是"sizebuf"
}

// encBufferPool ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// encBufferPool 是一个全局池子，我们可以从里面拿到一个 encBuffer 实例，每次从这个池子里拿一个 encBuffer 之后，
// 如果不放回去，那么下次再拿的话就不是我们刚刚拿的那个 encBuffer 了，除非我们拿了用完之后在给它放回去。
var encBufferPool = sync.Pool{New: func() interface{} { return new(encBuffer) }}

// getEncBuffer ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// getEncBuffer 方法从 encBufferPool 池里拿出一个 encBuffer 实例。
func getEncBuffer() *encBuffer {
	buf := encBufferPool.Get().(*encBuffer)
	buf.reset()
	return buf
}

// reset ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 该方法会执行以下代码来重置 encBuffer：
//
//	buf.lHeadsSize = 0
//	buf.str = buf.str[:0]
//	buf.lHeads = buf.lHeads[:0]
func (buf *encBuffer) reset() {
	buf.lHeadsSize = 0
	buf.str = buf.str[:0]
	buf.lHeads = buf.lHeads[:0]
}

// size ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// size 方法返回已编码数据的长度：len(encBuffer.str)+encBuffer.lHeadsSize，该方法返回的值就是编码数据的
// 结果的完整长度，例如原始数据是data，编码后的结果是result，那么该方法返回的结果相当于len(result)，
func (buf *encBuffer) size() int {
	return len(buf.str) + buf.lHeadsSize
}

// makeBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// makeBytes 方法的作用就是将编码结果完整的返回出来。
func (buf *encBuffer) makeBytes() []byte {
	result := make([]byte, buf.size())
	buf.copyTo(result)
	return result
}

// copyTo ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// copyTo 方法接受一个字节切片参数buf，该方法的作用是将 encBuffer 内存储的编码数据拷贝到buf里，同时还需要配合
// encBuffer.lHeads 字段将列表头或者编码的数据头编码到buf里。
func (buf *encBuffer) copyTo(dst []byte) {
	strPos := 0
	pos := 0
	for _, head := range buf.lHeads {
		// 第一个head的offset必等于0，buf.str[strPos:head.offset]表示前一个列表头到当前列表头之间的字符串
		n := copy(dst[pos:], buf.str[strPos:head.offset])
		pos += n
		strPos += n
		enc := head.encodeHead(dst[pos:])
		pos += len(enc)
	}
	// 下面这句很关键，如果我们编码的数据完全是字符串，那么上面的for循环根本不会执行，那么下面这段代码就可以将编码的
	// 字符串数据拷贝出来；而如果我们编码的数据是一个列表，那么下面这行代码可以将最后一个头后面跟着的编码数据拷贝出来
	copy(dst[pos:], buf.str[strPos:])
}

// writeTo ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// writeTo 方法接受一个 io.Writer 参数，该方法将编码结果完整地写入到给定的 io.Writer 里，官方的实现代码请跳转到：
//
//	https://github.com/ethereum/go-ethereum/blob/972007a517c49ee9e2a359950d81c74467492ed2/rlp/encbuffer.go#L79
//
// 我对官方的实现进行了简化，因为我们前面的 makeBytes 方法就可以获得完整的编码结果，何故再利用一个新的逻辑去获取编码结果呢？
func (buf *encBuffer) writeTo(w io.Writer) error {
	bz := buf.makeBytes()
	if _, err := w.Write(bz); err != nil {
		return err
	}
	return nil
}

// Write ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// Write 方法实现了 io.Writer 接口，该方法直接将给定的字节切片追加到 encBuffer.str 后面。返回值有两个，第一个返回值表示
// 给定切片的长度，第二个返回值永远为nil。
func (buf *encBuffer) Write(bz []byte) (int, error) {
	buf.str = append(buf.str, bz...)
	return len(bz), nil
}

// writeBool ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// writeBool 该方法接受一个bool类型的变量，然后根据其值将其编码到 encBuffer.str 里，如果给的值等于true，那么就往str里写
// 入0x01，否则写入0x80，0x80表示的是一个空值。
func (buf *encBuffer) writeBool(b bool) {
	if b {
		buf.str = append(buf.str, 0x01)
	} else {
		buf.str = append(buf.str, 0x80)
	}
}

// writeUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// writeUint64 方法接受一个64位的无符号整数作为输入参数，然后将其编码到 encBuffer.str 里，如果输入的整数大小小于128，则可
// 以将其看成是一个单独的ASCII码，那么该整数自身就可以作为编码结果写入到str里；如果给定的整数大于或等于128，则会计算需要多少个
// 字节才能存储这个数，例如需要n个，那么编码结果就是（0x80+n||整数的字节表现形式）；如果给定的整数等于0，则将0x80写入到str里，
// 代表空值，下面给出三个示例来对该方法的逻辑进行解释：
//   - 0：append(str, 0x80)
//   - 123：append(str, byte(123))
//   - 1024：append(str, []byte{0x80+2, 000000100, 00000000})
func (buf *encBuffer) writeUint64(i uint64) {
	if i == 0 {
		buf.str = append(buf.str, 0x80)
	} else if i < 128 {
		buf.str = append(buf.str, byte(i))
	} else {
		n := putInt(buf.auxiliaryBuf[1:], i)
		buf.auxiliaryBuf[0] = 0x80 + byte(n)
		buf.str = append(buf.str, buf.auxiliaryBuf[:1+n]...)
	}
}

// writeBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// writeBytes 方法接受一个字节切片bz，该方法的目的就是将字节切片bz编码到 encBuffer.str 里。当bz满足不同情况时，编码方式也不
// 同，具体会遇到以下两种方式：
//   - 给定的字节切片bz长度等于1，并且里面唯一的字节小于或等于0x7F，即127，那么该字节切片（或者说该字节更准确）会被直接追加到str后
//   - 给定的字节切片长度大于1，那么会将bz的长度先编码到str里，这里又会遇到两种情况：1.长度小于56；2.长度大于55。面对不同的情况，
//     如何对字节长度进行编码必须遵守以下准则：
//     · 当长度size小于56，则将0x80+size的值追加到str后
//     · 当长度size大于55，例如1024，存储1024最少需要2个字节，并且这两个字节分别是00000100和00000000，那么就将0xB7+2和这
//     两个字节追加到str后
//     字节的长度被编码到str里后，剩下的工作则是直接将切片bz追加到str后。
func (buf *encBuffer) writeBytes(bz []byte) {
	if len(bz) == 1 && bz[0] < 0x80 {
		buf.str = append(buf.str, bz[0])
	} else {
		// 先把字节切片的长度编码了
		buf.encodeStringHeader(len(bz))
		// 剩下的就是直接将字节切片追加到str后面
		buf.str = append(buf.str, bz...)
	}
}

// writeString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// writeString 方法接受一个字符串参数s，该方法的作用就是将s编码到 encBuffer.str 里，实际上，该方法的逻辑是调用了如下函数，来实现
// 将s编码到str里：
//
//	buf.writeBytes([]byte(s))
func (buf *encBuffer) writeString(s string) {
	buf.writeBytes([]byte(s))
}

// writeBigInt ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// writeBigInt 方法接受一个大整数i，i的类型是*big.Int，该方法的作用就是将大整数编码到 encBuffer.str 里。事实上，大整数不一定就比
// 最大的64位无符号整数大，在这种情况下，我们可以调用 writeUint64 方法将该所谓的大整数编码到str里；但是，如果给定的大整数大于最大的64
// 位无符号整数，在这种情况下，我们需要将其看成是一个字节切片进行编码，此时该大整数需要超过8个字节来存储。这里，我们对官方的源码作为一些改
// 动，经过测试，改动后的效果和官方源码一样。
func (buf *encBuffer) writeBigInt(i *big.Int) {
	if i.BitLen() <= 64 {
		buf.writeUint64(i.Uint64())
		return
	}
	// 计算该大整数需要多少个字节来存储，这里官方的计算方式是：
	// n := ((i.BitLen() + 7) & -8) >> 3
	// 不懂为何要这样写，显得更高级？
	n := (i.BitLen() + 7) / 8
	buf.encodeStringHeader(n)
	enc, index := make([]byte, n), n
	for _, word := range i.Bits() {
		for j := 0; j < math.WordBytes && index > 0; j++ {
			index--
			enc[index] = byte(word)
			word >>= 8
		}
	}
	buf.str = append(buf.str, enc...)
}

// encodeStringHeader ♏ |作者：吴翔宇| 🍁 |日期：2022/11/4|
//
// encodeStringHeader 方法接受一个整型size作为输入，顾名思义，该方法的作用就是在编码字符串数据时，将字符串的长度编码到 encBuffer.str
// 里，该方法仅在编码长度大于1的字符串或者字节切片时才会被使用。编码时，需要遵守以下两条规则：
//
//	· 当长度size小于56，则将0x80+size的值追加到str后
//	· 当长度size大于55，例如1024，存储1024最少需要2个字节，并且这两个字节分别是00000100和00000000，那么就将0xB7+2和这
//	  两个字节追加到str后
func (buf *encBuffer) encodeStringHeader(size int) {
	if size < 56 {
		buf.str = append(buf.str, 0x80+byte(size))
	} else {
		n := putInt(buf.auxiliaryBuf[1:], uint64(size))
		buf.auxiliaryBuf[0] = 0xB7 + byte(n)
		buf.str = append(buf.str, buf.auxiliaryBuf[:1+n]...)
	}
}

// listStart ♏ |作者：吴翔宇| 🍁 |日期：2022/11/5|
//
// listStart 方法用来往 encBuffer.lHeads 里添加一个 listHead 实例，该方法在编码列表数据前被调用，用来为编码列表数据作准
// 备，从该方法的逻辑也可以看出，listHead.offset 的作用就是标记编码结果在 encBuffer.str 的什么位值处开始被追加。
func (buf *encBuffer) listStart() int {
	head := listHead{offset: len(buf.str), size: buf.lHeadsSize}
	buf.lHeads = append(buf.lHeads, head)
	return len(buf.lHeads) - 1
}

// listEnd ♏ |作者：吴翔宇| 🍁 |日期：2022/11/5|
//
// listEnd 方法接受一个整型参数index，该方法在编码列表数据结束之后被调用，其目的就是更新 encBuffer.lHeads 上给定index索引
// 位值处的 listHead.size 和 encBuffer.lHeadsSize 两个字段。
// 结合 listStart 方法，我们对一个结构体实例进行编码，来看看 listStart 和 listEnd 究竟是如何工作的，给出两个结构体：
//
//	type Store struct {
//		Location string
//	}
//
//	type Dog struct {
//		Name  string
//		Age   uint64
//		Store Store
//	}
//
// 实例化一个Dog实例：d := Dog{Name: "aaa", Age: 12, Store: Store{Location: "Hefei"}}，然后对其进行编码，这里我们可
// 以看到我们给出的结构体实例是结构体中套结构体，那么对于rlp编码来说就是，列表中套列表，因此这里或产生两个表头，我们先给出编码结果：
//
//	[204 131 97 97 97 12 198 133 72 101 102 101 105]
//
// 可以看到，编码结果中的[198 133 72 101 102 101 105]是对Store: Store{Location: "Hefei"}列表数据编码产生的结果，它的表
// 头前缀是[198]，然后[204 131 97 97 97 12 198 133 72 101 102 101 105]是对整个Dog实例（外面的列表）编码的结果，它的表头
// 前缀是[204]，我们把所有表头的前缀去掉，将得到 encBuffer.str：[131 97 97 97 12 133 72 101 102 101 105]，那么第一个表头
// 的offset应当等于0，size则等于12，之所以等于12，是因为内部列表编码的结果（包括表头前缀也被算在内）长度等于7，加上5就等于12，表示
// 的是[131 97 97 97 12 198 133 72 101 102 101 105]的长度，第二个表头的offset应当等于5，size则等于6，6表示
// [133 72 101 102 101 105]的长度。由于两个表头前缀的长度都等于1，所以 encBuffer.lHeadsSize 最终等于2。
func (buf *encBuffer) listEnd(index int) {
	head := &buf.lHeads[index]
	// 此时的buf.size()=x原+h原+x新，head.offset=x原，head.size=h原，之所以此时的buf.size()不能加上h新，是因为h新还没
	// 有被计算出来呢
	// 那么下面的公式可以替换为：x原+h原+x新-x原-h原=x新
	head.size = buf.size() - head.offset - head.size
	if head.size < 56 {
		buf.lHeadsSize++
	} else {
		buf.lHeadsSize += 1 + intSize(uint64(head.size))
	}
}

// encode ♏ |作者：吴翔宇| 🍁 |日期：2022/11/5|
//
// encode 接受一个参数interface{}，然后根据该参数的类型选择对应的编码器，利用选中的编码器编码给定的参数数据。
func (buf *encBuffer) encode(val interface{}) error {
	rVal := reflect.ValueOf(val)
	encoder, err := cachedWriter(rVal.Type())
	if err != nil {
		return err
	}
	return encoder(rVal, buf)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// encReader ♏ |作者：吴翔宇| 🍁 |日期：2022/11/5|
//
// encReader
type encReader struct {
	buf    *encBuffer // 我们从 buf 缓冲里读取数据
	lhpos  int        // 我们正在读取的列表数据的列表头的索引位值
	strpos int        // 正在读取 encBuffer.str 的位值
	piece  []byte     // 下一个要被读取的数据片段
}

// Read ♏ |作者：吴翔宇| 🍁 |日期：2022/11/5|
//
// Read 方法接受一个字节切片缓冲区用于存储读取的数据，该方法在循环中不停地调用 next 方法获取接下来要被读取的数据，这样
// 可以在给定的缓冲区过小的情况下，记下读取到了哪里，在以后继续读取的过程中，从之前停下的位值处继续读取，如果 encBuffer
// 里面的数据读取完，会返回 io.EOF，然后 encReader.encBuffer 会被重新放到 encBufferPool池子里。
func (r *encReader) Read(bz []byte) (n int, err error) {
	for {
		if r.piece = r.next(); r.piece == nil {
			if r.buf != nil {
				encBufferPool.Put(r.buf)
				r.buf = nil
			}
			return n, io.EOF
		}
		m := copy(bz[n:], r.piece)
		n += m
		if m < len(r.piece) {
			// r.piece 里面的数据没有被读取完，那么就只好等待下次被读取咯！
			r.piece = r.piece[m:]
			return n, nil
		}
		r.piece = nil
	}
}

// next ♏ |作者：吴翔宇| 🍁 |日期：2022/11/5|
//
// next 方法返回 encBuffer 里面接下来要被读取的数据，如果 Read 方法给定的bz足够大，并且假设缓冲区等待被读取的数据为：
//
//	[204 131 97 97 97 12 198 133 72 101 102 101 105]
//
// 那么该方法在for循环里每次返回的数据如下所示：
//  1. 第一次循环：返回第一个列表头的编码数据[204]
//  2. 第二次循环：返回第一个列表头到第二个列表头之间的数据[131 97 97 97 12]
//  3. 第三次循环：返回第二个列表头的编码数据[198]
//  4. 第四次循环：由于第二个列表头就是最后一个列表头，因此直接返回str缓冲区剩下的所有数据[133 72 101 102 101 105]
//  5. 第5次循环：由于在第4次循环里，所有数据都被读取完，所以这一次返回值为nil
func (r *encReader) next() []byte {
	switch {
	case r.buf == nil:
		return nil
	case r.piece != nil:
		// 当前仍然有可被读取的数据，这种情况在给定的字节切片缓冲区不够大的情况下会出现
		return r.piece
	case r.lhpos < len(r.buf.lHeads):
		// 目前还没读取到最后一个列表头，获取当前正在读取的列表数据的头
		head := r.buf.lHeads[r.lhpos]
		// 计算在当前列表头之前还有多少数据未被读取，该逻辑设计的很巧妙，因为我们知道，第一个列表头的offset一定等于0，所以，
		// 第一次返回的要被读取的数据就是列表头的编码数据，而不是str缓冲区里的数据，因此strpos还会等于0，接下来，第二个列表
		// 头的offset一定大于0，此时我们就知道在第二个列表头前，str缓冲区里有一部分数据未被读取，而该段数据正是第一个列表头
		// 后面跟着的数据
		sizeBefore := head.offset - r.strpos
		if sizeBefore > 0 {
			// 如果在当前列表头前还有数据未被读取
			p := r.buf.str[r.strpos:head.offset] // 获取这段未被读取的数据
			r.strpos += sizeBefore
			// 将这段未被读取数据返回出去作为下一个即将被读取的数据
			return p
		}
		// 当前列表头之前的所有数据都被读取完毕，将列表头的索引值加1
		r.lhpos++
		// 将列表头的前缀返回出去，作为下一个即将被读取的数据
		return head.encodeHead(r.buf.auxiliaryBuf[:])
	case r.strpos < len(r.buf.str):
		// 字符串缓冲区的数据还没读取完，但是目前已经读取到最后一个列表头了，那么索性将字符串缓冲区里剩下的数据全部返回出去
		p := r.buf.str[r.strpos:]
		r.strpos = len(r.buf.str)
		return p
	default:
		return nil
	}
}
