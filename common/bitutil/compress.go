/*
Package bitutil
该文件定义了对字节切片进行解压缩的方法：
  - 压缩：CompressBytes
  - 解压：DecompressBytes
*/
package bitutil

import "errors"

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义一堆错误

var (
	// errMissingData 压缩数据不完整时会报该错误
	errMissingData = errors.New("missing bytes on input")
	// errUnreferencedData 解压缩时，读取压缩数据的长度不等于压缩数据的长度时会报该错误
	errUnreferencedData = errors.New("extra bytes on input")
	// errExceededTarget 解压缩时，给的压缩数据长度大于原始数据的长度时会报该错误
	errExceededTarget = errors.New("target data size exceeded")
	// errZeroContent 压缩数据中含有零值会报该错误
	errZeroContent = errors.New("zero byte in input content")
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 🌹压缩🌹

// CompressBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// CompressBytes 调用 bitsetEncodeBytes 方法实现对给定字节切片进行压缩，压缩过程分三种情况：
//  1. 如果给定的data是一个空切片，压缩结果就是nil
//  2. 如果给定的data长度等于1，并且里面唯一的字节等于0，压缩结果就是nil，否则就是data本身
//  3. 如果给定的data长度大于1，那么压缩结果的结构如下
//     |...|标记nonZeroBitset_1中非0字节位置的比特数组nonZeroBitset_2|标记data中非0字节位置的比特数组nonZeroBitset_1|data中非0字节拼接在一起|
//     例如data是一个长度为32的字节切片，它的第4、14、24下标处的值分别等于1、2、3，其余位置的值都等于0，即
//     data = [0 0 0 0 1 0 0 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 0 0 3 0 0 0 0 0 0 0 0]
//     压缩结果：[208 1 2 128 1 2 3]
func CompressBytes(data []byte) []byte {
	if out := bitsetEncodeBytes(data); len(out) < len(data) {
		return out
	}
	cpy := make([]byte, len(data))
	copy(cpy, data)
	return cpy
}

// bitsetEncodeBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// bitsetEncodeBytes 是一个递归方法，它接受一个字节切片data作为入参，它第一次递归的目的是将data里所有非零值单独取出，记录到一个新的
// 字节切片里：nonZeroBytes := make([]byte, 0, len(data))，然后用比特位数组标记出data哪些下标位置的字节值不等于0，这里会用一个
// 额外的字节切片来作记录：nonZeroBitset := make([]byte, (len(data)+7)/8)；第二次递归会将nonZeroBitset作为 bitsetEncodeBytes
// 方法的输入，也就是说第二次递归的任务是对比特位数组做压缩，然后将压缩的结果和nonZeroBytes数组拼接起来，得到最终的压缩结果。我们以一个
// 例子作为说明，来具体看看压缩的过程：
//
//	假设data是一个长度为32的字节切片，它的第4、14、24下标处的值分别等于1、2、3，其余位置的值都等于0，即
//		data = [0 0 0 0 1 0 0 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 0 0 3 0 0 0 0 0 0 0 0]
//	第一次迭代，取出data里不为0的字节并放入到nonZeroBytes_1里，然后在nonZeroBitset_1的第4、第14和第24比特位标记1，其余位置标记0，
//	得到结果：
//		nonZeroBytes_1 = [1 2 3] nonZeroBitset_1 = [00001000 00000010 00000000 10000000] = [8 2 0 128]
//	第二次迭代，取出nonZeroBitset_1里不为0的字节放入到nonZeroBytes_2里，然后在nonZeroBitset_2的第0、第1、第3比特位标记1，其余
//	位置标记0，得到结果：
//		nonZeroBytes_2 = [8 2 128] nonZeroBitset_2 = [11010000] = [208]
//	第三次迭代，由于nonZeroBitset_2的长度等于1，并且里面唯一的字节不等于0，所以直接返回[208]，迭代过程结束。
//
// 现在就是将上述迭代过程产生的中间值进行组合就可以得到最终的压缩结果：[208 8 2 128 1 2 3]
// 总结一下：先把原始字节切片中不为0的字节全部取出来按顺序排成一排，形成一个新的数组，这样我们就失去了原始字节切片中值为0的字节和不为0的
// 字节在原始切片中的位置信息，为此，我们用比特位标注出原始字节切片中哪些下标处的字节不为0，不为0的地方标1，为0的地方标0，这样我们就又得
// 到了一个新的字节切片来存储原始切片中各个字节的位置信息，所以为了尽量压缩数据的存储空间，我们利用相同的方法继续对该新字节切片进行压缩，
// 直到我们退出压缩过程。
func bitsetEncodeBytes(data []byte) []byte {
	// 空切片的压缩结果就是nil
	if len(data) == 0 {
		return nil
	}
	// 如果切片长度等于1，并且存储的字节等于0，那么压缩结果是nil，否则就是该字节本身
	if len(data) == 1 {
		if data[0] == 0 {
			return nil
		}
		return data
	}
	// 假如data的长度等于13，那么nonZeroBitset的长度等于(13+7)/8=2
	// nonZeroBytes的容量等于13
	nonZeroBitset := make([]byte, (len(data)+7)/8)
	nonZeroBytes := make([]byte, 0, len(data))
	for i, b := range data {
		if b != 0 {
			nonZeroBytes = append(nonZeroBytes, b)
			// 当我们处理data里面下标为2的字节时，nonZeroBitset[0] = nonZeroBitset[0] | (1 << byte(7-2))
			// --> x = x | (00100000)：这样可以在字节的对应位置的比特位标记1来记录该位置的字节为非零值
			// 当我们处理data里面下标为4的字节时，nonZeroBitset[0] = nonZeroBitset[0] | (1 << byte(7-4))
			// --> x = x | (00001000)
			nonZeroBitset[i/8] = nonZeroBitset[i/8] | (1 << byte(7-i%8))
		}
	}
	if len(nonZeroBytes) == 0 {
		// data切片里所有字节都为0
		return nil
	}
	return append(bitsetEncodeBytes(nonZeroBitset), nonZeroBytes...)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 🌹解压🌹

// DecompressBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// DecompressBytes 方法接受两个参数，第一个参数是经过压缩后的字节切片，第二个参数表示压缩前原始字节切片的长度，
// 该方法实际上是调用 bitsetDecodePartialBytes 方法来解压数据。
func DecompressBytes(data []byte, target int) ([]byte, error) {
	if len(data) > target {
		return nil, errExceededTarget
	}
	if len(data) == target {
		cpy := make([]byte, target)
		copy(cpy, data)
		return cpy, nil
	}
	return bitsetDecodeBytes(data, target)
}

// bitsetDecodeBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// bitsetDecodeBytes 方法接受两个参数，第一个参数是经过压缩后的字节切片，第二个参数表示压缩前原始字节切片的长度，
// 该方法实际上是调用 bitsetDecodePartialBytes 方法来解压数据。
func bitsetDecodeBytes(data []byte, target int) ([]byte, error) {
	out, size, err := bitsetDecodePartialBytes(data, target)
	if err != nil {
		return nil, err
	}
	if size != len(data) {
		return nil, errUnreferencedData
	}
	return out, nil
}

// bitsetDecodePartialBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/10/28|
//
// bitsetDecodePartialBytes 是一个递归方法，它接受两个参数，第一个参数是一个字节切片，它是对某个切片进行压缩后的结果，第二个参数是
// 一个整数，它表示压缩前原始数据的长度。该方法是从最内层逐渐反馈到最外层递归，我们以一个例子来说明该方法解压缩的过程：
//
//	假设data的值等于[208 1 2 128 1 2 3]，它是对一个长度为32的字节切片进行压缩后的结果，所以 bitsetDecodePartialBytes 方法的第
//	二个输入参数就是32，那么下面就开始介绍递归过程：
//		🚩第一次递归，nonZeroBitset_1, ptr_1, err_1 := bitsetDecodePartialBytes(data, 32)，此时data等于
//		[208 8 2 128 1 2 3]，target等于32，由于target不等于1，所以进入第二次迭代；
//		🚩第二次迭代，nonZeroBitset_2, ptr_2, err_2 := bitsetDecodePartialBytes(data, (target+7)/8)，此时，data等于
//		[208 8 2 128 1 2 3]，target等于4，由于target不等于1，所以进入第三次迭代；
//		🚩第三次迭代，nonZeroBitset_3, ptr_3, err_3 := bitsetDecodePartialBytes(data, (target+7)/8)，此时，data依然是
//		[208 8 2 128 1 2 3]，target等于(4+7)/8=1，由于target等于1，所以返回data[0], 1, nil。接着我们就要返回到第二次迭代
//	 	过程里了；
//		🚩回到第二次迭代过程，nonZeroBitset_2=[208]=[11010000]，ptr_2=1，err_2=nil，执行for循环，得到result=[8 2 0 128]，ptr=4，
//		然后返回到第一次迭代过程里；
//		🚩回到第一次迭代过程，nonZeroBitset_1=[8 2 0 128]=[00001000 00000010 00000000 10000000]，ptr_1=4, err_1=nil，执行
//		for循环，得到result=[0 0 0 0 1 0 0 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 0 0 3 0 0 0 0 0 0 0]，ptr=7。
func bitsetDecodePartialBytes(data []byte, target int) ([]byte, int, error) {
	if target == 0 {
		return nil, 0, nil
	}
	result := make([]byte, target)
	if len(data) == 0 {
		return result, 0, nil
	}
	if target == 1 {
		result[0] = data[0]
		if data[0] != 0 {
			return result, 1, nil
		}
		return result, 0, nil
	}
	nonZeroBitset, ptr, err := bitsetDecodePartialBytes(data, (target+7)/8)
	if err != nil {
		return nil, ptr, err
	}
	for i := 0; i < 8*len(nonZeroBitset); i++ { // 8*len(nonZeroBitset)代表nonZeroBitset里面有多少个比特位
		if nonZeroBitset[i/8]&(1<<byte(7-i%8)) != 0 {
			if ptr >= len(data) {
				return nil, 0, errMissingData
			}
			if i >= len(result) {
				return nil, 0, errExceededTarget
			}
			if data[ptr] == 0 {
				return nil, 0, errZeroContent
			}
			result[i] = data[ptr]
			ptr++
		}
	}
	return result, ptr, nil
}
