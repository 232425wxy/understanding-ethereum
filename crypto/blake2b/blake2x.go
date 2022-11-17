package blake2b

import "io"

// XOF ♏ |作者：吴翔宇| 🍁 |日期：2022/11/17|
//
// XOF 定义了支持任意长度输出的哈希函数的接口。
type XOF interface {
	// Writer 吸收更多的数据到哈希的状态中。如果在 "读 "之后调用它，就会出现恐慌。
	io.Writer
	// Reader 从哈希中读取更多的输出，如果达到极限，它将返回io.EOF。
	io.Reader
	// Clone 返回当前状态下的XOF的副本。
	Clone() XOF
	// Reset 将XOF重置成它的初始状态。
	Reset()
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义常量

const (
	// OutputLengthUnknown 可以作为NewXOF的大小参数，表示输出的长度事先不知道。
	OutputLengthUnknown = 0
	// magicUnknownOutputLength 是输出大小的一个魔法值，表示输出字节数的未知数，4294967295。
	magicUnknownOutputLength = (1 << 32) - 1
	// maxOutputLength 是当输出字节数未知时产生的绝对最大字节数。
	maxOutputLength = (1 << 32) * 64
)
