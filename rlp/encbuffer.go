package rlp

import (
	"io"
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
// writeTo 方法接受一个 io.Writer 参数，该方法将编码结果完整地写入到给定的 io.Writer 里，官方的实现代码如下：
//
//	strpos := 0
//	for _, head := range buf.lHeads {
//		// write string data before header
//		if head.offset-strpos > 0 {
//			n, err := w.Write(buf.str[strpos:head.offset])
//			strpos += n
//			if err != nil {
//			return err
//			}
//		}
//		// write the header
//		enc := head.encodeHead(buf.auxiliaryBuf[:])
//		if _, err = w.Write(enc); err != nil {
//			return err
//		}
//	}
//	if strpos < len(buf.str) {
//		// write string data after the last list header
//		_, err = w.Write(buf.str[strpos:])
//	}
//	return err
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