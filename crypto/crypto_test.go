package crypto

import (
	"crypto"
	_ "github.com/232425wxy/understanding-ethereum/crypto/blake2b"
	"testing"
)

func TestBlake2b(t *testing.T) {
	h256 := crypto.BLAKE2b_256.New()
	bz256 := h256.Sum([]byte("ethereum"))
	t.Log(h256.Size(), len(bz256), bz256)

	h384 := crypto.BLAKE2b_384.New()
	bz384 := h384.Sum([]byte("ethereum"))
	t.Log(h384.Size(), len(bz384), bz384)

	h512 := crypto.BLAKE2b_512.New()
	bz512 := h512.Sum([]byte("ethereum"))
	t.Log(h512.Size(), len(bz512), bz512)
}
