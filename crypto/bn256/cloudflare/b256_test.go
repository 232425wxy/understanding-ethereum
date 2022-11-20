package bn256

import (
	"bytes"
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestG1Marshal(t *testing.T) {
	_, Ga, err := RandomG1(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(G1)
	_, err = Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestG2Marshal(t *testing.T) {
	_, Ga, err := RandomG2(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(G2)
	_, err = Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestBilinearity(t *testing.T) {
	for i := 0; i < 2; i++ {
		a, p1, _ := RandomG1(rand.Reader)
		b, p2, _ := RandomG2(rand.Reader)
		e1 := Pair(p1, p2)

		e2 := Pair(&G1{curveGen}, &G2{twistGen})
		e2.ScalarMult(e2, a)
		e2.ScalarMult(e2, b)

		if *e1.p != *e2.p {
			t.Fatalf("bad pairing result: %s", e1)
		}
	}
}

func TestExamplePair(t *testing.T) {
	// This implements the tripartite Diffie-Hellman algorithm from "A One
	// Round Protocol for Tripartite Diffie-Hellman", A. Joux.
	// http://www.springerlink.com/content/cddc57yyva0hburb/fulltext.pdf

	// Each of three parties, a, b and c, generate a private value.
	a, _ := rand.Int(rand.Reader, Order)
	b, _ := rand.Int(rand.Reader, Order)
	c, _ := rand.Int(rand.Reader, Order)

	// Then each party calculates g₁ and g₂ times their private value.
	pa := new(G1).ScalarBaseMult(a)
	qa := new(G2).ScalarBaseMult(a)

	pb := new(G1).ScalarBaseMult(b)
	qb := new(G2).ScalarBaseMult(b)

	pc := new(G1).ScalarBaseMult(c)
	qc := new(G2).ScalarBaseMult(c)

	// Now each party exchanges its public values with the other two and
	// all parties can calculate the shared key.
	k1 := Pair(pb, qc)
	k1.ScalarMult(k1, a)

	k2 := Pair(pc, qa)
	k2.ScalarMult(k2, b)

	k3 := Pair(pa, qb)
	k3.ScalarMult(k3, c)

	// k1, k2 and k3 will all be equal.

	require.Equal(t, k1, k2)
	require.Equal(t, k1, k3)

	require.Equal(t, len(np), 4) //Avoid gometalinter varcheck err on np
}

func TestRandomG2Marshal(t *testing.T) {
	for i := 0; i < 10; i++ {
		n, g2, err := RandomG2(rand.Reader)
		if err != nil {
			t.Error(err)
			continue
		}
		t.Logf("%v: %x\n", n, g2.Marshal())
	}
}

func TestPairings(t *testing.T) {
	a1 := new(G1).ScalarBaseMult(bigFromBase10("1"))
	a2 := new(G1).ScalarBaseMult(bigFromBase10("2"))
	a37 := new(G1).ScalarBaseMult(bigFromBase10("37"))
	an1 := new(G1).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495616"))

	b0 := new(G2).ScalarBaseMult(bigFromBase10("0"))
	b1 := new(G2).ScalarBaseMult(bigFromBase10("1"))
	b2 := new(G2).ScalarBaseMult(bigFromBase10("2"))
	b27 := new(G2).ScalarBaseMult(bigFromBase10("27"))
	b999 := new(G2).ScalarBaseMult(bigFromBase10("999"))
	bn1 := new(G2).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495616"))

	p1 := Pair(a1, b1)
	pn1 := Pair(a1, bn1)
	np1 := Pair(an1, b1)
	if pn1.String() != np1.String() {
		t.Error("Pairing mismatch: e(a, -b) != e(-a, b)")
	}
	if !PairingCheck([]*G1{a1, an1}, []*G2{b1, b1}) {
		t.Error("MultiAte check gave false negative!")
	}
	p0 := new(GT).Add(p1, pn1)
	p0_2 := Pair(a1, b0)
	if p0.String() != p0_2.String() {
		t.Error("Pairing mismatch: e(a, b) * e(a, -b) != 1")
	}
	p0_3 := new(GT).ScalarMult(p1, bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495617"))
	if p0.String() != p0_3.String() {
		t.Error("Pairing mismatch: e(a, b) has wrong order")
	}
	p2 := Pair(a2, b1)
	p2_2 := Pair(a1, b2)
	p2_3 := new(GT).ScalarMult(p1, bigFromBase10("2"))
	if p2.String() != p2_2.String() {
		t.Error("Pairing mismatch: e(a, b * 2) != e(a * 2, b)")
	}
	if p2.String() != p2_3.String() {
		t.Error("Pairing mismatch: e(a, b * 2) != e(a, b) ** 2")
	}
	if p2.String() == p1.String() {
		t.Error("Pairing is degenerate!")
	}
	if PairingCheck([]*G1{a1, a1}, []*G2{b1, b1}) {
		t.Error("MultiAte check gave false positive!")
	}
	p999 := Pair(a37, b27)
	p999_2 := Pair(a1, b999)
	if p999.String() != p999_2.String() {
		t.Error("Pairing mismatch: e(a * 37, b * 27) != e(a, b * 999)")
	}
}
