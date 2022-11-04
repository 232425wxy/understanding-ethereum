package rlp

import "sync"

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
// size 方法返回已编码数据的长度：len(encBuffer.str)+encBuffer.lHeadsSize。
func (buf *encBuffer) size() int {
	return len(buf.str) + buf.lHeadsSize
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
