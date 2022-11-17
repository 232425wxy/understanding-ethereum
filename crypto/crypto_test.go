package crypto

import (
	"crypto"
	_ "github.com/232425wxy/understanding-ethereum/crypto/blake2b"
	"testing"
)

func TestBlake2b(t *testing.T) {
	h := crypto.BLAKE2b_256.New()
	size := h.Size()
	t.Log(size)

}
