package bls12381

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestG2EncodePoint(t *testing.T) {
	f := make([]byte, 96)

	g := NewG2()
	r, err := g.MapToCurve(f)
	assert.Nil(t, err)
	t.Log(g.EncodePoint(r))
}
