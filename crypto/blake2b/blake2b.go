package blake2b

import (
	"encoding/binary"
	"errors"
	"hash"
)

// 定义一些常量
const (
	// BlockSize 128字节，等于1024位
	BlockSize = 128
	// Size 64字节，等于512位，BLAKE2b-512的哈希长度
	Size = 64
	// Size384 48字节，等于384位，BLAKE2b-384的哈希长度
	Size384 = 48
	// Size256 32字节，等于256位，BLAKE2b-256的哈希长度
	Size256 = 32
)

const (
	magic = "b2b"
	// marshaledSize = 3 + 64 + 16 + 1 + 128 + 1 = 213
	marshaledSize = len(magic) + 8*8 + 2*8 + 1 + BlockSize + 1
)

var (
	useAVX2 bool
	useAVX  bool
	useSSE4 bool
)

// 定义和哈希算法相关的错误
var (
	errKeySize  = errors.New("blake2b: invalid key size")
	errHashSize = errors.New("blake2b: invalid hash size")
)

// 定义包级全局变量，iv里面的值是不会变的
var iv = [8]uint64{
	0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
	0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179,
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// digest ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// digest
type digest struct {
	h      [8]uint64
	c      [2]uint64
	size   int
	block  [BlockSize]byte
	offset int

	key    [BlockSize]byte
	keyLen int
}

func newDigest(hashSize int, key []byte) (*digest, error) {
	if hashSize < 1 || hashSize > Size {
		return nil, errHashSize
	}
	if len(key) > Size {
		return nil, errKeySize
	}
	d := &digest{
		size:   hashSize,
		keyLen: len(key),
	}
	copy(d.key[:], key)
	d.Reset()
	return d, nil
}

func (d *digest) MarshalBinary() ([]byte, error) {
	if d.keyLen != 0 {
		return nil, errors.New("crypto/blake2b: cannot marshal MACs")
	}
	// b是一个容量为213的字节切片
	b := make([]byte, 0, marshaledSize)
	// 将"b2b"作为b的前缀
	b = append(b, magic...)
	for i := 0; i < 8; i++ {
		b = appendUint64(b, d.h[i])
	}
	b = appendUint64(b, d.c[0])
	b = appendUint64(b, d.c[1])
	// Maximum value for size is 64
	b = append(b, byte(d.size))
	b = append(b, d.block[:]...)
	b = append(b, byte(d.offset))
	return b, nil
}

func (d *digest) UnmarshalBinary(b []byte) error {
	if len(b) < len(magic) || string(b[:len(magic)]) != magic {
		return errors.New("crypto/blake2b: invalid hash state identifier")
	}
	if len(b) != marshaledSize {
		return errors.New("crypto/blake2b: invalid hash state size")
	}
	b = b[len(magic):]
	for i := 0; i < 8; i++ {
		b, d.h[i] = consumeUint64(b)
	}
	b, d.c[0] = consumeUint64(b)
	b, d.c[1] = consumeUint64(b)
	d.size = int(b[0])
	b = b[1:]
	copy(d.block[:], b[:BlockSize])
	b = b[BlockSize:]
	d.offset = int(b[0])
	return nil
}

func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Size() int { return d.size }

func (d *digest) Reset() {
	d.h = iv
	d.h[0] ^= uint64(d.size) | (uint64(d.keyLen) << 8) | (1 << 16) | (1 << 24)
	d.offset, d.c[0], d.c[1] = 0, 0, 0
	if d.keyLen > 0 {
		d.block = d.key
		d.offset = BlockSize
	}
}

func (d *digest) Write(p []byte) (n int, err error) {
	n = len(p)

	if d.offset > 0 {
		remaining := BlockSize - d.offset
		if n <= remaining {
			d.offset += copy(d.block[d.offset:], p)
			return
		}
		copy(d.block[d.offset:], p[:remaining])
		hashBlocks(&d.h, &d.c, 0, d.block[:])
		d.offset = 0
		p = p[remaining:]
	}

	if length := len(p); length > BlockSize {
		nn := length &^ (BlockSize - 1)
		if length == nn {
			nn -= BlockSize
		}
		hashBlocks(&d.h, &d.c, 0, p[:nn])
		p = p[nn:]
	}

	if len(p) > 0 {
		d.offset += copy(d.block[:], p)
	}

	return
}

func (d *digest) Sum(sum []byte) []byte {
	var hash [Size]byte
	d.finalize(&hash)
	return append(sum, hash[:d.size]...)
}

func (d *digest) finalize(hash *[Size]byte) {
	var block [BlockSize]byte
	copy(block[:], d.block[:d.offset])
	remaining := uint64(BlockSize - d.offset)

	c := d.c
	if c[0] < remaining {
		c[1]--
	}
	c[0] -= remaining

	h := d.h
	hashBlocks(&h, &c, 0xFFFFFFFFFFFFFFFF, block[:])

	for i, v := range h {
		binary.LittleEndian.PutUint64(hash[8*i:], v)
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 全局函数

// Sum512 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// Sum512 方法返回数据的 BLAKE2b-512 校验和。
func Sum512(data []byte) [Size]byte {
	var sum [Size]byte
	checkSum(&sum, Size, data)
	return sum
}

// Sum384 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// Sum384 方法返回数据的 BLAKE2b-384 校验和。
func Sum384(data []byte) [Size384]byte {
	var sum [Size]byte
	var sum384 [Size384]byte
	checkSum(&sum, Size384, data)
	copy(sum384[:], sum[:Size384])
	return sum384
}

// Sum256 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// Sum256 方法返回数据的 BLAKE2b-256 校验和。
func Sum256(data []byte) [Size256]byte {
	var sum [Size]byte
	var sum256 [Size256]byte
	checkSum(&sum, Size256, data)
	copy(sum256[:], sum[:Size256])
	return sum256
}

// New512 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// New512 方法接受一个字节切片key作为输入参数，返回一个新的 hash.Hash 来计算 BLAKE2b-512
// 校验和。当输入的key等于nil计算得到的哈希值变成一个MAC。key的长度必须在0到64字节之间。
func New512(key []byte) (hash.Hash, error) { return newDigest(Size, key) }

// New384 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// New384 方法接受一个字节切片key作为输入参数，返回一个新的 hash.Hash 来计算 BLAKE2b-384
// 校验和。当输入的key等于nil计算得到的哈希值变成一个MAC。key的长度必须在0到64字节之间。
func New384(key []byte) (hash.Hash, error) { return newDigest(Size384, key) }

// New256 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// New256 方法接受一个字节切片key作为输入参数，返回一个新的 hash.Hash 来计算 BLAKE2b-256
// 校验和。当输入的key等于nil计算得到的哈希值变成一个MAC。key的长度必须在0到64字节之间。
func New256(key []byte) (hash.Hash, error) { return newDigest(Size256, key) }

// New ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// New 方法返回一个新的 hash.Hash 来计算具有自定义长度的 BLAKE2b 校验和。
// 该方法接受两个参数，第一个参数是一个int型的size，第二个参数是一个字节切片key，我们可以将其称为密钥，
// 其中size的值可以被设置为1到64中的任何一个整数，但强烈建议使用等于或大于以下的两个值：
// 	- 32：如果将size设为32，则可以将BLAKE2b作为哈希函数使用，然后key应当是nil的。
// 	- 16：如果将size设为16，则可以将BLAKE2b用作MAC函数，在这种情况下，key的长度应当介于16到64之间。
// 当key为nil时，返回的 hash.Hash 实现了 BinaryMarshaler 和 BinaryUnmarshaler。
func New(size int, key []byte) (hash.Hash, error) { return newDigest(size, key) }

// F ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// F F是BLAKE2b的一个压缩函数。它把状态向量`h`、信息块向量`m`、偏移计数器`t`、最终块指示标志`f`和
// 轮数`rounds`作为一个参数。作为第一个参数提供的状态向量被该函数修改。
func F(h *[8]uint64, m [16]uint64, c [2]uint64, final bool, rounds uint32) {
	var flag uint64
	if final {
		flag = 0xFFFFFFFFFFFFFFFF
	}
	f(h, &m, c[0], c[1], flag, uint64(rounds))
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// checkSum ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// checkSum 接受三个参数，类型和参数名如下所示：
// 	1. sum *[64]byte
//	2. hashSize int
//	3. data []byte
func checkSum(sum *[Size]byte, hashSize int, data []byte) {
	// 复制以下iv，以防破坏iv里的值
	h := iv
	h[0] ^= uint64(hashSize) | (1 << 16) | (1 << 24)
	var c [2]uint64

	if length := len(data); length > BlockSize {
		n := length &^ (BlockSize - 1)
		if length == n {
			n -= BlockSize
		}
		hashBlocks(&h, &c, 0, data[:n])
		data = data[n:]
	}

	var block [BlockSize]byte
	offset := copy(block[:], data)
	remaining := uint64(BlockSize - offset)
	if c[0] < remaining {
		c[1]--
	}
	c[0] -= remaining

	hashBlocks(&h, &c, 0xFFFFFFFFFFFFFFFF, block[:])

	for i, v := range h[:(hashSize+7)/8] {
		binary.LittleEndian.PutUint64(sum[8*i:], v)
	}
}

// 不可导出的函数

// hashBlocks ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// hashBlocks 接受四个参数，类型和参数名如下所示：
//  1. h *[8]uint64
//  2. c *[2]uint64
//  3. flag uint64
//  4. blocks []byte
func hashBlocks(h *[8]uint64, c *[2]uint64, flag uint64, blocks []byte) {
	// 定义一个长度为16的uint64数组，16*8=128，相当于128位的变量
	var m [16]uint64

	// 分别将数组c的第一个值和第二个值赋值给c0和c1，尽管传入的c是一个指针，但是被赋值
	// 的c0和c1都只是普通的数值，对c0和c1进行更改，不会影响c中的数据，确切来说，不会
	// 影响c[0]和c[1]的值
	c0, c1 := c[0], c[1]

	for i := 0; i < len(blocks); {
		c0 += BlockSize // 每循环一次c0加128
		if c0 < BlockSize {
			c1++ // 如果加上了128的c0还小于128，则给c1自增1
		}
		for j := range m {
			// 将blocks[i:]里的高位字节放在高地址位，然后转换成uint64类型的整数赋值给m的第j个值
			// 每次只取blocks[i:]的前8个字节
			m[j] = binary.LittleEndian.Uint64(blocks[i:])
			// 这里得循环16次，每循环一次，i就会自增8，等到最后一次循环结束，i将会等于128
			i += 8
		}
		f(h, &m, c0, c1, flag, 12)
	}
	c[0], c[1] = c0, c1
}

// appendUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// appendUint64 接受两个参数，第一个参数是一个byte切片，第二个参数是一个uint64类型的整数，该方法
// 首先对第二个参数按照大端编码方式编码成字节切片形式，得到一个长度为8的切片，然后将其append到第一个
// 参数后面并返回。
func appendUint64(b []byte, x uint64) []byte {
	var a [8]byte
	binary.BigEndian.PutUint64(a[:], x)
	return append(b, a[:]...)
}

// consumeUint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/16|
//
// consumeUint64 方法接受一个byte切片b，然后将b的前8个字节按照大端编码模式计算得到一个uint64类型
// 的整数x，然后返回b[8:]和得到的整数x。
func consumeUint64(b []byte) ([]byte, uint64) {
	x := binary.BigEndian.Uint64(b)
	return b[8:], x
}