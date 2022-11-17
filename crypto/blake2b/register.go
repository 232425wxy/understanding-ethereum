package blake2b

import (
	"crypto"
	"hash"
)

func init() {
	newHash256 := func() hash.Hash {
		h, _ := New256(nil)
		return h
	}
	newHash384 := func() hash.Hash {
		h, _ := New384(nil)
		return h
	}

	newHash512 := func() hash.Hash {
		h, _ := New512(nil)
		return h
	}

	// 这里将ethereum实现的哈希函数注册到crypto.hashes里，假如以后想要使用blake2b.go里实现
	// 的blake2b_384哈希函数，则按照以下方法去实例化这个哈希函数就可以了：
	//	h := crypto.BLAKE2b_384.New()
	crypto.RegisterHash(crypto.BLAKE2b_256, newHash256)
	crypto.RegisterHash(crypto.BLAKE2b_384, newHash384)
	crypto.RegisterHash(crypto.BLAKE2b_512, newHash512)
}
