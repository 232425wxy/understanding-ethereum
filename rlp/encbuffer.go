package rlp

import "sync"

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// encBuffer ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// encBuffer 结构体被用于在编码数据时存储编码结果。
type encBuffer struct {
	str          []byte     // str 包含了除列表头之外的所有编码信息
	lHeads       []listHead // 存储了所有列表头信息，官方源码的写法是"lheads"
	lHeadsSize   int        // 官方源码写法是"lhsize"
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
