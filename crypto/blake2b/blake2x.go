package blake2b

import (
	"encoding/binary"
	"errors"
	"io"
)

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
	// maxOutputLength 是当输出字节数未知时产生的绝对最大字节数，274877906944。
	maxOutputLength = (1 << 32) * 64
)

// NewXOF ♏ |作者：吴翔宇| 🍁 |日期：2022/11/17|
//
// NewXOF 方法创建一个新的可变输出长度的哈希函数，这个哈希函数要么产生一个已知的字节数（1 <= size < 2**32-1）长度的值，
// 要么产生一个未知的字节数（size == OutputLengthUnknown）长度的值，在后一种情况下，适用256GiB的绝对限制。
// NewXOF 方法接受两个参数，第一个参数为整数型的size，size就是前面提到的size，例如：（1 <= size < 2**32-1），另一个
// 参数是一个字节切片key，如果key非空的话，则会将哈希函数转换成MAC，密钥的长度必须在0到32字节之间。
func NewXOF(size uint32, key []byte) (XOF, error) {
	if len(key) > Size {
		return nil, errKeySize
	}
	if size == magicUnknownOutputLength {
		// 2^32-1 indicates an unknown number of bytes and thus isn't a
		// valid length.
		return nil, errors.New("blake2b: XOF length too large")
	}
	if size == OutputLengthUnknown {
		size = magicUnknownOutputLength
	}
	x := &xof{
		d: digest{
			size:   Size,
			keyLen: len(key),
		},
		length: size,
	}
	copy(x.d.key[:], key)
	x.Reset()
	return x, nil
}

type xof struct {
	d                digest
	length           uint32
	remaining        uint64
	cfg, root, block [Size]byte
	offset           int
	nodeOffset       uint32
	readMode         bool
}

func (x *xof) Write(p []byte) (n int, err error) {
	if x.readMode {
		panic("blake2b: write to XOF after read")
	}
	return x.d.Write(p)
}

func (x *xof) Clone() XOF {
	clone := *x
	return &clone
}

func (x *xof) Reset() {
	x.cfg[0] = byte(Size)
	binary.LittleEndian.PutUint32(x.cfg[4:], uint32(Size)) // leaf length
	binary.LittleEndian.PutUint32(x.cfg[12:], x.length)    // XOF length
	x.cfg[17] = byte(Size)                                 // inner hash size

	x.d.Reset()
	x.d.h[1] ^= uint64(x.length) << 32

	x.remaining = uint64(x.length)
	if x.remaining == magicUnknownOutputLength {
		x.remaining = maxOutputLength
	}
	x.offset, x.nodeOffset = 0, 0
	x.readMode = false
}

func (x *xof) Read(p []byte) (n int, err error) {
	if !x.readMode {
		x.d.finalize(&x.root)
		x.readMode = true
	}

	if x.remaining == 0 {
		return 0, io.EOF
	}

	n = len(p)
	if uint64(n) > x.remaining {
		n = int(x.remaining)
		p = p[:n]
	}

	if x.offset > 0 {
		blockRemaining := Size - x.offset
		if n < blockRemaining {
			x.offset += copy(p, x.block[x.offset:])
			x.remaining -= uint64(n)
			return
		}
		copy(p, x.block[x.offset:])
		p = p[blockRemaining:]
		x.offset = 0
		x.remaining -= uint64(blockRemaining)
	}

	for len(p) >= Size {
		binary.LittleEndian.PutUint32(x.cfg[8:], x.nodeOffset)
		x.nodeOffset++

		x.d.initConfig(&x.cfg)
		x.d.Write(x.root[:])
		x.d.finalize(&x.block)

		copy(p, x.block[:])
		p = p[Size:]
		x.remaining -= uint64(Size)
	}

	if todo := len(p); todo > 0 {
		if x.remaining < uint64(Size) {
			x.cfg[0] = byte(x.remaining)
		}
		binary.LittleEndian.PutUint32(x.cfg[8:], x.nodeOffset)
		x.nodeOffset++

		x.d.initConfig(&x.cfg)
		x.d.Write(x.root[:])
		x.d.finalize(&x.block)

		x.offset = copy(p, x.block[:todo])
		x.remaining -= uint64(todo)
	}
	return
}

func (d *digest) initConfig(cfg *[Size]byte) {
	d.offset, d.c[0], d.c[1] = 0, 0, 0
	for i := range d.h {
		d.h[i] = iv[i] ^ binary.LittleEndian.Uint64(cfg[i*8:])
	}
}
