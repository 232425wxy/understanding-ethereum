package blake2b

import "testing"

func TestMagicUnknownOutputLength(t *testing.T) {
	t.Log(magicUnknownOutputLength)
}

func TestMaxOutputLength(t *testing.T) {
	t.Log(maxOutputLength)
}
