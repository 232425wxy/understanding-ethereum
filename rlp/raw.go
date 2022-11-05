package rlp

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// RawValue ♏ |作者：吴翔宇| 🍁 |日期：2022/11/6|
//
// RawValue 官方对其解释是：
//
//	RawValue代表一个已编码的RLP值，可用于延迟RLP解码或预先计算一个编码。请注意，解码器并不验证RawValues的内容是否是有效的RLP。
type RawValue []byte
