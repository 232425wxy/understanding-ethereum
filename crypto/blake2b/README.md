# 简介

BLAKE2 系列比常见的 MD5，SHA-1，SHA-2，SHA-3 更快，同时提供不低于 SHA-3 的安全性。

BLAKE2 系列从著名的 ChaCha 算法衍生而来，有两个主要版本 BLAKE2b（BLAKE2）和 BLAKE2s。

BLAKE2b 为 64 位 CPU（包括 ARM Neon）优化，可以生成最长64字节的摘要；BLAKE2s 为 8-32 位 CPU 设计，可以生成最长 32 字节的摘要。

# 说明

以下官方源码里的文件因为各种各样的原因，为了减少分析源码的工作量，就没有被分析：

- 官方源码里的`blake2b_amd64.go`文件没有被分析，因为这个文件里的代码要求使用非`gccgo`编译器，而我自己电脑上的编译器恰恰就是`gccgo`
- 官方源码里的`blake2b_f_fuzz.go`文件没有被分析，因为我的电脑里没有装`go-fuzz`
- 官方源码里的`blake2bAVX2_amd64.go`文件没有被分析，因为这个文件里的代码要求使用非`gccgo`编译器，而我自己电脑上的编译器恰恰就是`gccgo`

# 使用方法

`regesiter.go`文件里将`ethereum`实现的`blake2b`哈希函数注册到了`golang...\crypto.go`里：

```go
crypto.RegisterHash(crypto.BLAKE2b_256, newHash256)
crypto.RegisterHash(crypto.BLAKE2b_384, newHash384)
crypto.RegisterHash(crypto.BLAKE2b_512, newHash512)
```

以后如果想要使用`blake2b_512`，按照下面的方式来实例化哈希函数就可以了：

```go
h := crypto.BLAKE2b_512.New()
```

实际上，上面代码段里的返回的`h`其实就是下面这个函数的返回结果：
```go
func New512(key []byte) (hash.Hash, error) {
	return newDigest(Size, key)
}
```

但是请注意，这里返回的`h`其实是一个接口，它的底层是`*digest`，所以，想调用`*digest`的`MarshalBinary`方法，就还得将`h`强制类型转换成`*digest`。

同理，如果想要使用`blake2b_384`和`blake2b_256`哈希函数，就分别按照下面的方法实例化哈希函数：

```go
h := crypto.BLAKE2b_384.New()
h := crypto.BLAKE2b_256.New()
```

# 使用案例

我们分别实例化了`blake2b_256`、`blake2b_384`和`blake2b_512`三个哈希函数，然后分别对`"ethereum"`进行哈希运算，得到的哈希值长度和结果分别如下所示：
```go
h256 := crypto.BLAKE2b_256.New()
bz256 := h256.Sum([]byte("ethereum"))
t.Log(h256.Size(), len(bz256), bz256)

h384 := crypto.BLAKE2b_384.New()
bz384 := h384.Sum([]byte("ethereum"))
t.Log(h384.Size(), len(bz384), bz384)

h512 := crypto.BLAKE2b_512.New()
bz512 := h512.Sum([]byte("ethereum"))
t.Log(h512.Size(), len(bz512), bz512)
```

**输出：**

>32 40 [101 116 104 101 114 101 117 109 14 87 81 192 38 229 67 178 232 171 46 176 96 153 218 161 209 229 223 71 119 143 119 135 250 171 69 205 241 47 227 168]

>48 56 [101 116 104 101 114 101 117 109 179 40 17 66 51 119 245 45 120 98 40 110 225 167 46 229 64 82 67 128 253 161 114 74 111 37 215 151 140 111 211 36 74 108 175 4 152 129 38 115 197 224 94 245 131 130 81 0]

>64 72 [101 116 104 101 114 101 117 109 120 106 2 247 66 1 89 3 198 198 253 133 37 82 210 114 145 47 71 64 225 88 71 97 138 134 226 23 247 31 84 25 210 94 16 49 175 238 88 83 19 137 100 68 147 78 176 75 144 58 104 91 20 72 183 85 213 111 112 26 254 155 226 206]