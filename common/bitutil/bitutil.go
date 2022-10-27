/*
Package bitutil
该文件定义了若干对01比特的操作方法：
	- 异或运算
	- 与运算
	- 或运算
	- 检查给定字节切片中是否存在值为非0的字节
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

// 🌹异或运算🌹

// XORBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// XORBytes 方法接受三个入参，分别是dst、a、b三个字节切片，该方法实现对给定的a、b两个字节切片进行异或运算，并将结果
// 存储到dst中，如果运行该方法的计算机架构属于{386、amd64、ppc64、ppc64le、s390x}这其中的某一个，则执行快速算法
// fastXORBytes 来进行异或运算，否则采用常规的算法 safeXORBytes。该方法的返回值表示对a或b中多少个字节进行了异或运算。
//
//	例如：输入a=[12 34 28] b=[3 67 98 55]，经过运算，dst=[15 97 126]
//	12 xor 3 -> 1100 ^ 0011 -> 1111 -> 15
//	34 xor 67 -> 0100010 ^ 1000011 -> 1100001 -> 97
//	28 xor 98 -> 0011100 ^ 1100010 -> 1111110 -> 126
func XORBytes(dst, a, b []byte) int {
	if supportUnaligned {
		return fastXORBytes(dst, a, b)
	}
	return safeXORBytes(dst, a, b)
}

// safeXORBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// safeXORBytes 方法接受3个入参，分别是dst、a、b，该方法就是对给定的a、b两个字节切片，进行逐字节的异或运算。不要求
// a、b两个字节切片的长度必须一样，但是要求dst参数的长度至少等于a、b中长度最短的那一个，该方法的返回值表示对a或b中多少
// 个字节进行了异或运算。
//
//	例如：输入a=[12 34 28] b=[3 67 98 55]，经过运算，dst=[15 97 126]
//	12 xor 3 -> 1100 ^ 0011 -> 1111 -> 15
//	34 xor 67 -> 0100010 ^ 1000011 -> 1100001 -> 97
//	28 xor 98 -> 0011100 ^ 1100010 -> 1111110 -> 126
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
// 第一那行代码的作用是从切片a的左侧开始，每8个字节会被以小端存储模式连成一个64位的地址值，然后存入到新的切片aw里，
// 尽管这样，aw的长度并不是等于len(a)/8，而是等于len(a)，但是只有aw[:len(a)/8]这一段才有意义，bw与aw同理。要
// 问为什么这里是每8个字节会被以小端存储模式连成一个64位的地址值，是因为在64位的Ubuntu 20.04操作系统中，地址值的
// 长度就是64位，即正好8个字节。
// 例如，如果切片a=[1 1 1 1 0 0 0 1]，则aw[0]的值如下所示：
//
//	aw[0] = 0000000100000000000000000000000000000001000000010000000100000001，用整数表示为：72057594054770945
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
//
// 该方法的返回值表示对a或b中多少个字节进行了异或运算。
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

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 🌹与运算🌹

// ANDBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// ANDBytes 方法接受3个入参，分别是dst、a、b三个字节切片，该方法实现对给定的a、b两个字节切片进行与运算，并将结果
// 存储到dst中，如果运行该方法的计算机架构属于{386、amd64、ppc64、ppc64le、s390x}这其中的某一个，则执行快速算
// 法 fastANDBytes 来进行与运算，否则采用常规的算法 safeANDBytes。该方法的返回值表示对a或b中多少个字节进行了与
// 运算。
//
//	例如：输入a=[12 34 28] b=[3 67 98 55]，经过运算，dst=[0 2 0]
//	12 xor 3 -> 1100 & 0011 -> 0000 -> 0
//	34 xor 67 -> 0100010 & 1000011 -> 0000010 -> 2
//	28 xor 98 -> 0011100 & 1100010 -> 0000000 -> 0
func ANDBytes(dst, a, b []byte) int {
	if supportUnaligned {
		return fastANDBytes(dst, a, b)
	}
	return safeANDBytes(dst, a, b)
}

// safeANDBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// safeANDBytes 方法接受3个入参，分别是dst、a、b，该方法就是对给定的a、b两个字节切片，进行逐字节的与运算。不要求
// a、b两个字节切片的长度必须一样，但是要求dst参数的长度至少等于a、b中长度最短的那一个，该方法的返回值表示对a或b中多
// 少个字节进行了与运算。
//
//	例如：输入a=[12 34 28] b=[3 67 98 55]，经过运算，dst=[0 2 0]
//	12 xor 3 -> 1100 & 0011 -> 0000 -> 0
//	34 xor 67 -> 0100010 & 1000011 -> 0000010 -> 2
//	28 xor 98 -> 0011100 & 1100010 -> 0000000 -> 0
func safeANDBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] & b[i]
	}
	return n
}

// fastANDBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// fastANDBytes 方法接受3个入参，分别是dst、a、b，该方法就是对给定的a、b两个字节切片进行与运算，并将结果存到
// dst中，但是该方法比 safeANDBytes 方法更快，根据benchmark的性能表现，对长度为100字节的两个字节切片进行与
// 运算，该方法比 safeANDBytes 方法的38.49 ns/op，快了大约1.5倍。
// 该方法的运算速度比较快的原因在于下面的代码：
//
//	aw := *(*[]uintptr)(unsafe.Pointer(&a))
//	bw := *(*[]uintptr)(unsafe.Pointer(&b))
//
// 第一行代码的作用是从切片a的左侧开始，每8个字节会被以小端存储模式连成一个64位的地址值，然后存入到新的切片aw里，
// 尽管这样，aw的长度并不是等于len(a)/8，而是等于len(a)，但是只有aw[:len(a)/8]这一段才有意义，bw与aw同理。
// 要问为什么这里是每8个字节会被以小端存储模式连成一个64位的地址值，是因为在64位的Ubuntu 20.04操作系统中，地址
// 值的长度就是64位，即正好8个字节。
// 例如，如果切片a=[1 1 1 1 0 0 0 1]，则aw[0]的值如下所示：
//
//	aw[0] = 0000000100000000000000000000000000000001000000010000000100000001，用整数表示为：72057594054770945
//
// 接着，我们再利用如下代码进行与运算：
//
//	for i := 0; i < n / wordSize; i++ {
//		dw[i] = aw[i] & bw[i]
//	}
//
// 速度会明显加快，因为上面代码里的一次循环是对8对字节做与运算。由于给定的a、b字节切片的长度不一定是8的整数倍，因此我们还需要
// 利用如下代码，对剩下的字节做与运算：
//
//	for i := n - n%wordSize; i < n; i++ {
//		dst[i] = a[i] & b[i]
//	}
//
// 该方法的返回值表示对a或b中多少个字节进行了与运算。
func fastANDBytes(dst, a, b []byte) int {
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
			dw[i] = aw[i] & bw[i]
		}
	}
	for i := n - n%wordSize; i < n; i++ {
		dst[i] = a[i] & b[i]
	}
	return n
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 🌹或运算🌹

// ORBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// ORBytes 方法接受3个入参，分别是dst、a、b三个字节切片，该方法实现对给定的a、b两个字节切片进行或运算，并将结果
// 存储到dst中，如果运行该方法的计算机架构属于{386、amd64、ppc64、ppc64le、s390x}这其中的某一个，则执行快速算
// 法 fastORBytes 来进行与运算，否则采用常规的算法 safeORBytes。该方法的返回值表示对a或b中多少个字节进行了或
// 运算。
//
//	例如：输入a=[12 34 28] b=[3 67 98 55]，经过运算，dst=[15 99 126]
//	12 xor 3 -> 1100 | 0011 -> 0000 -> 15
//	34 xor 67 -> 0100010 | 1000011 -> 0000010 -> 99
//	28 xor 98 -> 0011100 | 1100010 -> 0000000 -> 126
func ORBytes(dst, a, b []byte) int {
	if supportUnaligned {
		return fastORBytes(dst, a, b)
	}
	return safeORBytes(dst, a, b)
}

// safeORBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// safeORBytes 方法接受3个入参，分别是dst、a、b，该方法就是对给定的a、b两个字节切片，进行逐字节的或运算。不要求
// a、b两个字节切片的长度必须一样，但是要求dst参数的长度至少等于a、b中长度最短的那一个，该方法的返回值表示对a或b中
// 多少个字节进行了或运算。
//
//	例如：输入a=[12 34 28] b=[3 67 98 55]，经过运算，dst=[15 99 126]
//	12 xor 3 -> 1100 | 0011 -> 1111 -> 15
//	34 xor 67 -> 0100010 | 1000011 -> 1100011 -> 99
//	28 xor 98 -> 0011100 | 1100010 -> 1111110 -> 126
func safeORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] | b[i]
	}
	return n
}

// fastORBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// fastORBytes 方法接受3个入参，分别是dst、a、b，该方法就是对给定的a、b两个字节切片进行或运算，并将结果存到
// dst中，但是该方法比 safeORBytes 方法更快，根据benchmark的性能表现，对长度为100字节的两个字节切片进行或
// 运算，该方法比 safeORBytes 方法的40.69 ns/op，快了大约1.8倍。
// 该方法的运算速度比较快的原因在于下面的代码：
//
//	aw := *(*[]uintptr)(unsafe.Pointer(&a))
//	bw := *(*[]uintptr)(unsafe.Pointer(&b))
//
// 第一行代码的作用是从切片a的左侧开始，每8个字节会被以小端存储模式连成一个64位的地址值，然后存入到新的切片aw里，
// 尽管这样，aw的长度并不是等于len(a)/8，而是等于len(a)，但是只有aw[:len(a)/8]这一段才有意义，bw与aw同理。
// 要问为什么这里是每8个字节会被以小端存储模式连成一个64位的地址值，是因为在64位的Ubuntu 20.04操作系统中，地址
// 值的长度就是64位，即正好8个字节。
// 例如，如果切片a=[1 1 1 1 0 0 0 1]，则aw[0]的值如下所示：
//
//	aw[0] = 0000000100000000000000000000000000000001000000010000000100000001，用整数表示为：72057594054770945
//
// 接着，我们再利用如下代码进行或运算：
//
//	for i := 0; i < n / wordSize; i++ {
//		dw[i] = aw[i] | bw[i]
//	}
//
// 速度会明显加快，因为上面代码里的一次循环是对8对字节做或运算。由于给定的a、b字节切片的长度不一定是8的整数倍，因此我们还需要
// 利用如下代码，对剩下的字节做或运算：
//
//	for i := n - n%wordSize; i < n; i++ {
//		dst[i] = a[i] | b[i]
//	}
//
// 该方法的返回值表示对a或b中多少个字节进行了或运算。
func fastORBytes(dst, a, b []byte) int {
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
			dw[i] = aw[i] | bw[i]
		}
	}
	for i := n - n%wordSize; i < n; i++ {
		dst[i] = a[i] | b[i]
	}
	return n
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// TestBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// TestBytes 方法接受一个字节切片p作为输入参数，该方法实现对给定的字节切片p进行检查，判断p中
// 是否存在值为非0的字节，如果存在，直接返回true，否则返回false。如果运行该方法的计算机架构属
// 于{386、amd64、ppc64、ppc64le、s390x}这其中的某一个，则执行快速算法 fastTestBytes
// 进行计算，否则采用常规的算法 safeTestBytes。
func TestBytes(p []byte) bool {
	if supportUnaligned {
		return fastTestBytes(p)
	}
	return safeTestBytes(p)
}

// safeTestBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// safeTestBytes 接受一个字节切片p作为入参，然后逐个字节检查是否存在不等于0的字节，如果存在，则直接返回true，
// 如果遍历完也没发现不等于0的字节，即p中的所有字节都等于0，则返回false。
func safeTestBytes(p []byte) bool {
	for i := 0; i < len(p); i++ {
		if p[i] != 0 {
			return true
		}
	}
	return false
}

// fastTestBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/27|
//
// fastTestBytes 接受一个字节切片p作为入参，该方法也是判断字节切片p中是否存在不为0的字节，如果存在，则返回true，否则返
// 回false，但是该方法比 safeTestBytes 方法的执行速度要快，根据benchmark的性能表现，对长度为100字节的字节切片进行计算，
// 该方法比 safeTestBytes 方法的48.69 ns/op，快了大约3倍。（测试的字节切片中每个字节的值都等于0）
// 该方法的运算速度比较快的原因在于下面的代码：
//
//	pw := *(*[]uintptr)(unsafe.Pointer(&p))
//
// 上面代码的作用是从切片p的左侧开始，每8个字节会被以小端存储模式连成一个64位的地址值，然后存入到新的切片pw里，
// 尽管这样，pw的长度并不是等于len(a)/8，而是等于len(a)，但是只有aw[:len(a)/8]这一段才有意义。要问为什么
// 这里是每8个字节会被以小端存储模式连成一个64位的地址值，是因为在64位的Ubuntu 20.04操作系统中，地址值的长度
// 就是64位，即正好8个字节。
// 例如，如果切片p=[1 1 1 1 0 0 0 1]，则pw[0]的值如下所示：
//
//	pw[0] = 0000000100000000000000000000000000000001000000010000000100000001，用整数表示为：72057594054770945
//
// 接着，我们再利用如下代码判断字节是否等于0：
//
//	for i := 0; i < n / wordSize; i++ {
//		if pw[i] != 0 {
//			return true
//		}
//	}
//
// 速度会明显加快，因为上面代码里的一次循环是对8字节进行判断。由于给定的字节切片p的长度不一定是8的整数倍，因此我们还需要
// 利用如下代码，对剩下的字节进行判断：
//
//	for i := n - n%wordSize; i < n; i++ {
//		if p[i] != 0 {
//			return true
//		}
//	}
func fastTestBytes(p []byte) bool {
	n := len(p)
	w := n / wordSize
	if w > 0 {
		pw := *(*[]uintptr)(unsafe.Pointer(&p))
		for i := 0; i < w; i++ {
			if pw[i] != 0 {
				return true
			}
		}
	}
	for i := n - n%wordSize; i < n; i++ {
		if p[i] != 0 {
			return true
		}
	}
	return false
}