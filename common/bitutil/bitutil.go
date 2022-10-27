/*
Package bitutil
该文件定义了若干对01比特的操作方法
*/
package bitutil

import (
	"runtime"
	"unsafe"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// wordSize ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// wordSize表示存储一个指针数据需要多少个字节，在64位的Ubuntu 20.04操作系统中，wordSize 的值等于8.
const wordSize = int(unsafe.Sizeof(uintptr(0)))

// supportUnaligned ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// supportUnaligned用来表示当前的计算机架构是否支持内存不对齐，在64位的Ubuntu 20.04机器上，supportAligned的值恒为true。
const supportUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "amd64" || runtime.GOARCH == "ppc64" || runtime.GOARCH == "ppc64le" || runtime.GOARCH == "s390x"

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 异或运算

// safeXORBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// safeXORBytes 方法接受3个入参，分别是dst、a、b，该方法就是对给定的a、b两个字节切片，进行逐字节的异或运算。不要求
// a、b两个字节切片的长度必须一样，但是要求dst参数的长度至少等于a、b中长度最短的那一个，该方法的返回值表示对a或b中多少
// 个字节进行了异或运算。
//
//	例如：输入a=[12 34 28] b=[3 67 98 55]，经过运算，dst=[15 97 126]
//	12 xor 3 -> 1100 xor 0011 -> 1111 -> 15
//	34 xor 67 -> 0100010 xor 1000011 -> 1100001 -> 97
//	28 xor 98 -> 0011100 xor 1100010 -> 1111110 -> 126
func safeXORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}

// fastXORBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// fastXORBytes 方法接受3个入参，分别是dst、a、b，该方法就是对给定的a、b两个字节切片进行异或运算，并将结果存到
// dst中，但是该方法比 safeXORBytes 方法更快，根据benchmark的性能表现，对长度为100字节的两个字节切片进行异或
// 运算，该方法比 safeXORBytes 方法的39.45 ns/op，快了大约1.5倍。
// 该方法的运算速度比较快的原因在于下面的代码：
//
//	aw := *(*[]uintptr)(unsafe.Pointer(&a))
//	bw := *(*[]uintptr)(unsafe.Pointer(&b))
//
// 上面那行代码的作用是从切片a的左侧开始，每8个字节会被以小端存储模式连成一个64位的地址值，然后存入到新的切片aw里，
// 尽管这样，aw的长度并不是等于len(a)/8，而是等于len(a)，但是只有aw[:len(a)/8]这一段才有意义，bw与aw同理。要
// 问为什么这里是每8个字节会被以小端存储模式连成一个64位的地址值，是因为在64位的Ubuntu 20.04操作系统中，地址值的
// 长度就是64位，即正好8个字节。
// 例如，如果切片a=[1 1 1 1 0 0 0 1]，则aw的值如下所示：
//
//	aw = 0000000100000000000000000000000000000001000000010000000100000001，用整数表示为：72057594054770945
//
// 接着，我们再利用如下代码进行异或运算：
//
//	for i := 0; i < n / wordSize; i++ {
//		dw[i] = aw[i] ^ bw[i]
//	}
//
// 速度会明显加快，因为上面代码里的一次循环是对8对字节做异或运算。由于给定的a、b字节切片的长度不一定是8的整数倍，因此我们还需要
// 利用如下代码，对剩下的字节做异或运算：
//
//	for i := n - n%wordSize; i < n; i++ {
//		dst[i] = a[i] ^ b[i]
//	}
func fastXORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))
		for i := 0; i < w; i++ {
			dw[i] = aw[i] ^ bw[i]
		}
	}
	for i := n - n%wordSize; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}
