package blake2b

import "errors"

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

var (
	useAVX2 bool
	useAVX  bool
	useSSE4 bool
)

var (
	errKeySize  = errors.New("blake2b: invalid key size")
	errHashSize = errors.New("blake2b: invalid hash size")
)
