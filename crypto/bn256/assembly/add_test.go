package assembly

import "testing"

func TestAssemblyAdd(t *testing.T) {
	t.Log(add(1, 2))
}

func BenchmarkAddAssembly(b *testing.B) {
	var x, y int64 = 100, 200
	for i := 0; i < b.N; i++ {
		add(x, y)
	}
}

func BenchmarkAddNormal(b *testing.B) {
	var x, y int64 = 100, 200
	for i := 0; i < b.N; i++ {
		addNormal(x, y)
	}
}
