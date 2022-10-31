package rlp

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
