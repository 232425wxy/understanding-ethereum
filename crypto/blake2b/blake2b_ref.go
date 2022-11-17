//go:build !amd64 || appengine || gccgo

package blake2b

// f 实际上没有调用该函数，调用的是blake2bAVX2_amd64.go里的f函数
func f(h *[8]uint64, m *[16]uint64, c0, c1 uint64, flag uint64, rounds uint64) {
	fGeneric(h, m, c0, c1, flag, rounds)
}
