# 简介

BLS12-381是Sean Bowe在2017年设计的椭圆曲线，用于对Zcash协议进行更新，该曲线pairing-friendly并且可用于高效构建zkSnarks，许多协议使用它来实现数字签名和零知识证明。

**BLS12-381**里的两个数字解释如下：

**12**：曲线的嵌入度（后面再介绍概念）

**381**：曲线上的点坐标表示所需的bit位数，即有限域的modulus q 的位数。因为点的坐标来自质数阶有限域，我们可以用384位（48Bytes）来表示每个域元素，留3 bit来做标志位或者算术优化。这个位数是由安全需求与实现高效所共同决定的。

# 使用案例

在**以太坊**的官方源码里，**bls12381**主要在`core/vm/contracts.go`文件里被使用

**func (c \*bls12381MapG2) Run(input []byte) ([]byte, error)方法里**

```go
g := bls12381.NewG2()
// fe是一个长度为96的字节切片
r, err := g.MapToCurve(fe)
return g.EncodePoint(r)
```

**func (c \*bls12381MapG1) Run(input []byte) ([]byte, error)**

```go
g := bls12381.NewG1()
// fe是一个长度为48的字节切片
r, err := g.MapToCurve(fe)
return g.EncodePoint(r)
```

**func (c \*bls12381Pairing) Run(input []byte) ([]byte, error)**

```go
func (c *bls12381Pairing) Run(input []byte) ([]byte, error) {
// Implements EIP-2537 Pairing precompile logic.
// > Pairing call expects `384*k` bytes as an inputs that is interpreted as byte concatenation of `k` slices. Each slice has the following structure:
// > - `128` bytes of G1 point encoding
// > - `256` bytes of G2 point encoding
// > Output is a `32` bytes where last single byte is `0x01` if pairing result is equal to multiplicative identity in a pairing target field and `0x00` otherwise
// > (which is equivalent of Big Endian encoding of Solidity values `uint256(1)` and `uin256(0)` respectively).
k := len(input) / 384
if len(input) == 0 || len(input)%384 != 0 {
return nil, errBLS12381InvalidInputLength
}

// Initialize BLS12-381 pairing engine
e := bls12381.NewPairingEngine()
g1, g2 := e.G1, e.G2

// Decode pairs
for i := 0; i < k; i++ {
off := 384 * i
t0, t1, t2 := off, off+128, off+384

// Decode G1 point
p1, err := g1.DecodePoint(input[t0:t1])
if err != nil {
return nil, err
}
// Decode G2 point
p2, err := g2.DecodePoint(input[t1:t2])
if err != nil {
return nil, err
}

// 'point is on curve' check already done,
// Here we need to apply subgroup checks.
if !g1.InCorrectSubgroup(p1) {
return nil, errBLS12381G1PointSubgroup
}
if !g2.InCorrectSubgroup(p2) {
return nil, errBLS12381G2PointSubgroup
}

// Update pairing engine with G1 and G2 ponits
e.AddPair(p1, p2)
}
// Prepare 32 byte output
out := make([]byte, 32)

// Compute pairing and set the result
if e.Check() {
out[31] = 1
}
return out, nil
}
```

# 说明

因为编译器自动忽略的原因，以下源文件没有被`understanding-ethereum`包含进来：

- arithmetic_decl.go
- arithmetic_x86.s
- arithmetic_x86_adx.go
- arithmetic_x86_noadx.go