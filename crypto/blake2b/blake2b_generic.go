package blake2b

import "math/bits"

// precomputed ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// precomputed 定义了10个长度为16的字节数数组。
var precomputed = [10][16]byte{
	{0, 2, 4, 6, 1, 3, 5, 7, 8, 10, 12, 14, 9, 11, 13, 15},
	{14, 4, 9, 13, 10, 8, 15, 6, 1, 0, 11, 5, 12, 2, 7, 3},
	{11, 12, 5, 15, 8, 0, 2, 13, 10, 3, 7, 9, 14, 6, 1, 4},
	{7, 3, 13, 11, 9, 1, 12, 14, 2, 5, 4, 15, 6, 10, 0, 8},
	{9, 5, 2, 10, 0, 7, 4, 15, 14, 11, 6, 3, 1, 12, 8, 13},
	{2, 6, 0, 8, 12, 10, 11, 3, 4, 7, 15, 1, 13, 5, 14, 9},
	{12, 1, 14, 4, 5, 15, 13, 10, 0, 6, 9, 8, 7, 3, 2, 11},
	{13, 7, 12, 3, 11, 14, 1, 9, 5, 15, 8, 2, 0, 4, 6, 10},
	{6, 14, 11, 0, 15, 9, 3, 8, 12, 13, 1, 10, 2, 7, 4, 5},
	{10, 8, 7, 1, 2, 4, 6, 5, 15, 9, 3, 13, 11, 14, 12, 0},
}

//fGeneric ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
//fGeneric
func fGeneric(h *[8]uint64, m *[16]uint64, c0, c1 uint64, flag uint64, rounds uint64) {
	// 将数组h中的8个值分别都取出来
	v0, v1, v2, v3, v4, v5, v6, v7 := h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7]
	// 将数组iv里的值也分别取出来，数组iv里的每个值都是比较大的64位无符号整数
	v8, v9, v10, v11, v12, v13, v14, v15 := iv[0], iv[1], iv[2], iv[3], iv[4], iv[5], iv[6], iv[7]
	v12 ^= c0   // 将v12与c0做异或运算
	v13 ^= c1   // 将v13与c1做异或运算
	v14 ^= flag // 将v14与flag做异或运算

	for i := 0; i < int(rounds); i++ {
		// 简单来说，取出precomputed第(i%10)个数据的地址，往后如果对s指向的存储空间里的值进行修改，则会
		// 影响precomputed[i%10]的值
		// 此处，s指向的存储空间里存储的是一个长度为16的字节数组
		s := &(precomputed[i%10])

		// v0加上m里的第s[0]个值，之后再加上v4
		v0 += m[s[0]]
		v0 += v4
		// v12先与v0做异或运算，然后再让：v12 = v12 << 32 | v12 >> 32
		v12 ^= v0
		v12 = bits.RotateLeft64(v12, -32)
		// v8加上v12，再将v4与v8做异或运算，最后：v4 = v4 << 40 | v4 >> 24
		v8 += v12
		v4 ^= v8
		v4 = bits.RotateLeft64(v4, -24)
		// s是一个长度为16的数组，让v1加上m的第s[1]个值，然后再让v1加上v5
		v1 += m[s[1]]
		v1 += v5
		// 让v13与v1做异或运算，然后：v13 = v13 << 32 | v13 >> 32
		v13 ^= v1
		v13 = bits.RotateLeft64(v13, -32)
		// 让v9加上v13，然后让v5与v9做异或运算，接着：v5 = v5 << 40 | v5 >> 24
		v9 += v13
		v5 ^= v9
		v5 = bits.RotateLeft64(v5, -24)
		//
		v2 += m[s[2]]
		v2 += v6
		v14 ^= v2
		v14 = bits.RotateLeft64(v14, -32)
		v10 += v14
		v6 ^= v10
		v6 = bits.RotateLeft64(v6, -24)
		v3 += m[s[3]]
		v3 += v7
		v15 ^= v3
		v15 = bits.RotateLeft64(v15, -32)
		v11 += v15
		v7 ^= v11
		v7 = bits.RotateLeft64(v7, -24)

		v0 += m[s[4]]
		v0 += v4
		v12 ^= v0
		v12 = bits.RotateLeft64(v12, -16)
		v8 += v12
		v4 ^= v8
		v4 = bits.RotateLeft64(v4, -63)
		v1 += m[s[5]]
		v1 += v5
		v13 ^= v1
		v13 = bits.RotateLeft64(v13, -16)
		v9 += v13
		v5 ^= v9
		v5 = bits.RotateLeft64(v5, -63)
		v2 += m[s[6]]
		v2 += v6
		v14 ^= v2
		v14 = bits.RotateLeft64(v14, -16)
		v10 += v14
		v6 ^= v10
		v6 = bits.RotateLeft64(v6, -63)
		v3 += m[s[7]]
		v3 += v7
		v15 ^= v3
		v15 = bits.RotateLeft64(v15, -16)
		v11 += v15
		v7 ^= v11
		v7 = bits.RotateLeft64(v7, -63)

		v0 += m[s[8]]
		v0 += v5
		v15 ^= v0
		v15 = bits.RotateLeft64(v15, -32)
		v10 += v15
		v5 ^= v10
		v5 = bits.RotateLeft64(v5, -24)
		v1 += m[s[9]]
		v1 += v6
		v12 ^= v1
		v12 = bits.RotateLeft64(v12, -32)
		v11 += v12
		v6 ^= v11
		v6 = bits.RotateLeft64(v6, -24)
		v2 += m[s[10]]
		v2 += v7
		v13 ^= v2
		v13 = bits.RotateLeft64(v13, -32)
		v8 += v13
		v7 ^= v8
		v7 = bits.RotateLeft64(v7, -24)
		v3 += m[s[11]]
		v3 += v4
		v14 ^= v3
		v14 = bits.RotateLeft64(v14, -32)
		v9 += v14
		v4 ^= v9
		v4 = bits.RotateLeft64(v4, -24)

		v0 += m[s[12]]
		v0 += v5
		v15 ^= v0
		v15 = bits.RotateLeft64(v15, -16)
		v10 += v15
		v5 ^= v10
		v5 = bits.RotateLeft64(v5, -63)
		v1 += m[s[13]]
		v1 += v6
		v12 ^= v1
		v12 = bits.RotateLeft64(v12, -16)
		v11 += v12
		v6 ^= v11
		v6 = bits.RotateLeft64(v6, -63)
		v2 += m[s[14]]
		v2 += v7
		v13 ^= v2
		v13 = bits.RotateLeft64(v13, -16)
		//
		v8 += v13
		v7 ^= v8
		v7 = bits.RotateLeft64(v7, -63)
		// 让v3加上m的第s[15]个值，然后让v3加上v4，接着让v14与v3做异或运算，最后：v14 = v14 << 48 | v14 >> 16
		v3 += m[s[15]]
		v3 += v4
		v14 ^= v3
		v14 = bits.RotateLeft64(v14, -16)
		// 让v9加上v14，接着让v4与v9做异或运算，然后：v4 = v4 << 1 || v4 >> 63
		v9 += v14
		v4 ^= v9
		v4 = bits.RotateLeft64(v4, -63)
	}
	h[0] ^= v0 ^ v8
	h[1] ^= v1 ^ v9
	h[2] ^= v2 ^ v10
	h[3] ^= v3 ^ v11
	h[4] ^= v4 ^ v12
	h[5] ^= v5 ^ v13
	h[6] ^= v6 ^ v14
	h[7] ^= v7 ^ v15
}
