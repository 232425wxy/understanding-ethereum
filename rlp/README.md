## RLP简介

<img src="https://gitee.com/Sagaya815/assets/raw/master/rlp.png" style="zoom:16%;" />

`递归长度前缀（Recursive Length Prefix，RLP）`编码是以太坊项目特别设计的一种编码方式，它的编码结果相比于`JSON`编码占用更少的存储空间，因为在`JSON`编码方式中，需要额外的字段名信息来组织编码内容。而在`RLP`编码里，则使用一种被称为**前缀**的字段来组织编码内容，这可以有效减少编码过程中产生的其他额外信息。

`RLP`编码结果由**编码前缀（Encoding Prefix，EP）**和**编码内容（Encoding content，EC）**两部分组成：

编码结果 := `EP` || `EC`

## 1. 类型标记位

在`RLP`编码中，`EP`由两部分组成，其中第一部分是占据`1`个字节存储空间的**类型标记位（Type Marker Bit，TMB）**，它可被分为五种，对应`RLP`编码中处理的五种数据类型，如下所示：

- 0\~127 | 0x0\~0x7F：单个的`ASCII`码
- 128\~183 | 0x80\~0xB7：长度在`56`以内的`string`，即`EC`的长度小于`56`
- 184\~191 | 0xB8\~0xBF：长度大于`55`的`string`，即`EC`的长度大于`55`
- 192\~247 | 0xC0\~0xF7：编码结果的长度小于`56`的`list`，即`EC`的长度小于`56`
- 248\~255 | 0xF8\~0xFF：编码结果的长度大于`55`的`list`，即`EC`的长度大于`55`

在`go`语言中，上面提到的`string`我们可以将其简单理解为简单数据类型的变量，例如`uint`、`string`、`byte`等，而`list`我们可以将其理解为复合数据类型，例如`struct`等。值得注意的是，在`go-ethereum`里，`[]byte`和`[num]byte`被归类为`string`，而`[]x`和`[num]x`（其中`x`为非`byte`类型）被归类为`list`，另外，空接口，即`interface{}`也被归类为`list`。

当我们给定一个数据对象`x`，然后利用`RLP`编码技术对`x`进行编码，得到编码结果`r`，一般情况下，`r`的第一个字节存储的就是`TMB`，因此，我们可以根据`TMB`推断出`x`的数据类型。

## 2. 可选长度编码

前面我们介绍了`EP`由两部分组成，其中第一部分是**类型标记位TMB**，第二部分就是**可选长度编码（Optional Length Coding，OLC）**。当`TMB`的取值在`184~191`或`248~255`这两个区间内时，`OLC`才会出现在`前缀`里。根据观察`TMB`的取值范围和对应编码的数据类型时，我们发现，当`TMB`取值在`128~183`和`192~247`之间时，我们只需要`TMB`就可以求出`EP`后面紧跟着的`EC`的长度，例如我们利用`RLP`去编码一段长度为`32`的`string`，`TMB`将等于`160`，它的取值落在`128~183`之间，那么利用`160`减去`128`就可以得到`EC`长度等于`32`这个结论。其实，之所以`EC`的长度也等于`32`，是因为，对于`string`类型的数据，`RLP`对其编码得到的`EC`其实就是数据它本身。

到这里，我们可能会感受到，`rlp`是一种和数据类型以及数据长度息息相关的编码技术。

如果我们编码一个长度为`1025`的`string`类型数据，那么仅仅根据`TMB`就无法计算出`EP`后面跟着的`EC`有多长了，在这种情况下，我们需要`OLC`来辅助存储编码数据的长度。首先，`RLP`会先对`1025`进行长度编码，长度编码遵从**大端编码规则（高位字节存储在低地址位）**，`1025`的二进制表现形式为`[00000100,00000001]`，其实这个`1025`的二进制表现形式就是它的长度编码结果，为了表示方便，我们用`[4,1]`来表示长度编码结果，即，如果我们编码一个长度为`1025`的`string`类型数据，那么`EP`中的`OLC`就等于`[4,1]`。现在我们先尝试组装一下编码结果，得到结果如下所示：

EP || EC $\rightarrow$ TMB || OLC || EC

当我们把这串数据发送给接收者后，接收者如何进行解码呢？我们知道，`TMP`只占据一个字节存储空间，因此接收者可以直接获得`TMB`，但是`OLC`具体占据多少字节存储空间则是未知的，接收者无法区分`OLC`和`EC`。基于这一点考虑，`RLP`将`OLC`的长度信息存储到了`TMB`中。

还以上面的例子为例，如果我们编码一个长度为`1025`的`string`类型数据，那么`TMB`的取值应当落在`184~191`之间，前面我们知道，`1025`的长度编码结果为`[00000100,00000001]`（简写为`[4,1]`），因此至少需要两个字节空间来存储`1025`的长度编码，换句话说，这个例子里的`OLC`长度为`2`，所以我们需要将`2`这个长度信息存储到`TMB`中，具体的存储方式如下：

`TMB` := 184 + (2 - 1) = 185

这样的话，接收者在拿到编码结果以后，根据`TMB`可以知道`OLC`的长度，获取到`OLC`的长度以后，就可将`OLC`与`EC`分隔开。根据`TMB`获取`OLC`的长度计算规则如下：

- if `TMB` $\in$ 184\~191, then, length $\leftarrow$ `TMB` - 183
- if `TMB` $\in$ 248\~255, then, length $\leftarrow$ `TMB` - 247

## 3. 递归编码

要理**解递归编**码这个概念，我们需要先理解`list`到底是什么，前面我们已经介绍过对`string`类型的数据进行编码，得到的`EC`就是数据本身，而对`list`类型的数据进行`RLP`编码，得到的`EC`还是数据本身吗？这里先说答案：**不是**。

我们给一个例子，例如我们对以下数据进行`RLP`编码：

> x := []string{"abc", "def"}

`x`是一个字符串切片，根据前面对`string`和`list`数据类型的介绍，我们知道：`x`属于`list`数据类型，但是`x`里面存储的`"abc"`和`"def"`却是`string`数据类型，对`string`数据类型的编码结果等于`EP`连接上`EC`，`EC`就是数据本身，因此我们只需要计算`EP`是多少就行。

以`"def"`为例，它的长度等于`3`（`EC`的长度也等于`3`），因此`EP`将只含有`TMB`，不含有`OLC`，所以`EP`=`TMB`=`128+3`=`131`，将`EP`和`EC`进行组合，得到：[131 100 101 102]，其中`100`、`101`和`102`分别是`'a'`、`'b'`和`'c'`的`ASCII`码值。

同理，对`"abc"`进行`RLP`编码，得到结果：[131 97 98 99]。

`x`内部元素`RLP`编码结果已知，现在需要返回到`list`这一层，对`x`进行编码，对`x`进行`RLP`编码，得到的编码结果中的`EC`等于什么呢？实际上，`EC`就等于`"abc"`和`"def"`的编码结果“之和”：[131 97 98 99 131 100 101 102]，那么`EC`的长度我们就可以求得等于`8`，根据`TMB`和不同数据类型的对应关系，我们可以推断`TMB`的取值应当落在`192~247`之间，且根据`OLC`的出现条件，我们可以判断此处`EC`和`TMB`可以划等号，所以`EC`=`TMB`=`192+8`=`200`，所以，`x`最终的编码结果为：[200, 131 97 98 99 131 100 101 102]。

上面的例子可能还不能很好地体会到**递归**的精髓，下面给一个新的例子：

> x := []interface{}{[]interface{}{}, \[][]interface{}{{}}}

它的`RLP`编码结果为：[195 192 193 192]。

对上面编码结果如果存在疑问，可以查看源码，其中空接口的编码方式如下：

```go
if val.IsNil() {
	buf.str = append(buf.str, 0xC0)
	return nil
}
```

## 4. 结构体中的编码规则

`rlpstruct\rlpstruct.go`文件里定义了使用`RLP`编码如何对用户自定义的数据结构进行编解码的方式，通过为结构体字段的`tag`设置不同的`rlp标签`，可以实现若干种编解码方式。`go-ethereum`定义了一个`Tags`结构体来维护结构体字段的`rlp标签`，如下所示：

```go
type Tags struct {
    NilKind NilKind
    NilManual bool
    Optional bool
    Tail bool
    Ignored bool
}
```

- **NilKind**字段定义了结构体字段的空值编码规则：`NilKindString`或`NilKindList`。

- **NilManual**如果设置为`true`，则表明结构体字段的空值编码规则被手动设置为：`rlp:"nil"`、` rlp:"nilString"`或` rlp:"nilList"`。

- **Optional**用来表示该字段的`rlp标签`是否被设置成`rlp:"optional"`，如果某个结构体定义了`4`个可导出的字段，并且在第二个字段的`rlp标签`里设置了`rlp:"optional"`，那么第三和第四个字段的`tag`里也必须要设置`rlp:optional`，除非第四个字段是一个切片，并且它的`rlp标签`已经被设置为`rlp:"tail"`，那么这第四个字段的`tag`就不能设置为`rlp:"optional"`。给结构体的字段的`rlp标签`设置成`optional`具有以下作用呢：当我们对结构体进行编码时，如果从某个字段开始往后所有字段的`tag`都被设置成`optional`，且这些字段中存在违未被始化的情况，，它会遵循以下规则进行编码：排在最后一个`rlp标签`为`optional`且值为非零值的字段前面的字段（包括该字段），不管它们的`rlp标签`有没有设置为`optional`，也不管它们的值是否等于零值，这些字段都将参与编码，而排在后面的值为零值的字段，这些字段由于它们的`rlp标签`一定被设置成`optional`，且它们的值在运行时阶段是零值，所以它们将不参与编码，下面给一个例子做为说明：

  ```go
  type People struct {
      Name string
      Age uint8 `rlp:"optional"`
      Son *People `rlp:"optional"`
      Daughter *People `rlp:"optional"`
  }
  var p1 People = People{Name: "Tom", Age: 35, Daughter: &People{Name: "Lina", Age: 8}}
  // 由于People的第二、三、四3个字段的tag都被设置成optional，所以当对p1进行编码时，因为最后一个非零值字段是Daughter，所以排在它前面的字段（包括Son字段，尽管它的值等于零值）包括它自己（Daughter字段）都会被编码，所以编码结果如下：
  // [205 131 84 111 109 35 192 198 132 76 105 110 97 8]
  var p2 People = People{Name: "Tom", Son: &People{Name: "David", Age: 10}}
  // 我们对p2进行编码，发现p2最后一个非零字段是Son，尽管它前面的Age字段是零值，但是它排在Son前面，所以依然会被编码，而Daughter字段为零值，且排在Son
  // 之后，所以不会被编码，那么编码结果就如下所示：
  // [205 131 84 111 109 128 199 133 68 97 118 105 100 10]
  ```

- **Tail**字段用来表示该字段的`rlp标签`是否被设置成`rlp:"tail"`，`RLP`编码规则规定：在任何自定义结构体中，只有最后一个可导出字段，且该字段还必须是切片类型，才能给该字段的`rlp标签`设置成`rlp:"tail"`，这也映证了`Tail`的中文含义。那么它的作用是什么呢？根据`RLP`的编码规则，我们知道那些元素类型为非`byte`类型的切片或者数组会被当成**列表**进行编码，并且在`go`代码中，切片或者数组里的所有元素类型必须统一，那么如果我们给结构体的最后一个字段的`rlp标签`设置成`tail`，在编码时，会将该字段“拆开看”，所谓拆开看就是不会将该字段看成一个整体：**list**，而是会逐一对该字段所表示的切片里的元素进行编码，例如下面给出了两个示例：

  ```go
  // 示例1
  type class struct {
  	ClassID  uint8
  	Students []string `rlp:"tail"`
  }
  var c class = class{ClassID: 3, Students: []string{"abc", "def"}}
  // 由于此时我们给class结构体的Students字段的tag设置成tail，所以在编码时，会将其拆开看，不会将其看成整体一个，也就是说在编码时是会把class结构体看
  // 成如下的结构体：
  // type class struct {
  //     ClassID  uint8
  //     Student1 string
  //     Student2 string
  //     Student3 string
  //     ...
  // }
  // 所以对上面的c进行编码的结果是：[201 3 131 97 98 99 131 100 101 102]
  // 而如果我们把Students字段的tag里的tail去掉，编码结果则变为：[202 3 200 131 97 98 99 131 100 101 102]，将其作为一个列表（整体）进行编码
  ```

  ```go
  // 示例3
  type class struct {
  	ClassID  uint8
  	Students []string `rlp:"tail"`
  }
  var c class = class{ClassID: 3}
  // 对c进行编码，因为c里面并没有初始化Students字段，所以它的值等于零值，又因为Students的tag被设置为tail，所以不会将其看成一个列表在对其进行编码，
  // 仅仅是将其看成若干个连续的string类型的字段，所以对c的编码结果为：[193 3]
  // 而如果我们把Students字段的tag里的tail去掉，编码结果则变为：[194 3 192]，因为此时会将Students字段看成是一个整体（列表），空列表的编码结果为
  // 0XC0
  ```

- **Ignored**字段用来表示该字段的`rlp标签`是否被设置成`rlp:"-"`，如果被设置成`rlp:"-"`，那么该字段在编码时会被直接忽略，不参与编码，例如下面给出了一个代码示例：

  ```go
  type student struct {
  	Name  string
      Age   uint8 `rlp:"-"`
  	Birth string
  }
  var s student = student{Name: "abc", Age: 18, Birth: "def"}
  // 比如上面给出了一个结构体student，该结构体内定义了一个学生的姓名、年龄和出生日期，一般来说，我们对数据进行编码是为了进行网络传输或者文件存储，为了
  // 减小网络开销，我们提倡只编码有用的数据，例如在这个例子里，当我们知道一个学生的出生日期，那么就可以推出该学生的年龄，所以我们忽略对student结构体的
  // Age字段进行编码，那么对s进行编码的结果是：[200 131 97 98 99 131 100 101 102]
  // 如果我们将Age字段的tag里的“-”给去掉，编码结果则变为：[201 131 97 98 99 18 131 100 101 102]
  ```

> 总结下来，利用rlp编码规则对自定义结构体进行编码，我们可以在结构体字段的tag里设置以下种编码标记：
>
> - rlp:"nil"
> - rlp:"nilString"
> - rlp:"nilList"
> - rlp:"optional"
> - rlp:"tail"
> - rlp:"-"

## 5. 案例

### 5.1 编码bool类型数据

| 原值  | 编码结果 |
| ----- | -------- |
| true  | [1]      |
| false | [128]    |

### 5.2 编码无符号整数

| 原值     | 编码结果          |
| -------- | ----------------- |
| 0        | [128]             |
| 127      | [127]             |
| 128      | [129 128]         |
| 256      | [130 1 0]         |
| 1024     | [130 4 0]         |
| 0xffffff | [131 255 255 255] |

### 5.3 编码大整数

| 原值                             | 编码结果                                                     |
| -------------------------------- | ------------------------------------------------------------ |
| 0                                | [80]                                                         |
| 1                                | [1]                                                          |
| 127                              | [127]                                                        |
| 128                              | [129 128]                                                    |
| 256                              | [130 1 0]                                                    |
| 0x123456789abcdef123456789abcdef | [143 18 52 86 120 154 188 222 241 35 69 103 137 171 205 239] |

### 5.4 编码字节数组

| 原值            | 编码结果                                                     |
| --------------- | ------------------------------------------------------------ |
| [0]byte{}       | [128]                                                        |
| [1]byte{0}      | [0]                                                          |
| [1]byte{1}      | [1]                                                          |
| [1]byte{127}    | [127]                                                        |
| [1]byte{128}    | [129 128]                                                    |
| [3]byte{1,2,3}  | [131 1 2 3]                                                  |
| [60]byte{1,2,3} | [184 60 1 2 3 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] |

### 5.5 编码字节切片

| 原值          | 编码结果    |
| ------------- | ----------- |
| []byte{}      | [128]       |
| []byte{0}     | [0]         |
| []byte{1}     | [1]         |
| []byte{127}   | [127]       |
| []byte{128}   | [129 128]   |
| []byte{1,2,3} | [131 1 2 3] |

### 5.6 编码字符串

| 原值                                                         | 编码结果                                                     |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| ""                                                           | [80]                                                         |
| "aaa"                                                        | [131 97 97 97]                                               |
| "My major is cyberspace security"                            | [159 77 121 32 109 97 106 111 114 32 105 115 32 99 121 98 101 114 115 112 97 99 101 32 115 101 99 117 114 105 116 121] |
| "RLP encoding is a new encoding method specifically implemented in the Ethereum" | [184 78 82 76 80 32 101 110 99 111 100 105 110 103 32 105 115 32 97 32 110 101 119 32 101 110 99 111 100 105 110 103 32 109 101 116 104 111 100 32 115 112 101 99 105 102 105 99 97 108 108 121 32 105 109 112 108 101 109 101 110 116 101 100 32 105 110 32 116 104 101 32 69 116 104 101 114 101 117 109] |

### 5.7 编码非字节切片

| 原值                                                         | 编码结果                                                     |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| []uint{}                                                     | [192]                                                        |
| []uint{1}                                                    | [193 1]                                                      |
| []uint{1 9 17}                                               | [195 1 9 17]                                                 |
| []interface{}{[]interface{}{}}                               | [193 192]                                                    |
| []interface{}{[]interface{}{}, uint(3)}                      | [194 192 3]                                                  |
| []interface{}{[]interface{}{}, []interface{}{[]interface{}{}}} | [195 192 193 192]                                            |
| []interface{}{[]interface{}{}, \[][]interface{}{{}}}         | [195 192 193 192]                                            |
| []string{"aaa", "bbb", "ccc"}                                | [204 131 97 97 97 131 98 98 98 131 99 99 99]                 |
| []interface{}{uint(1), uint(0xffffff), []interface{}{[]uint{4, 5, 6}}, "abc"} | [206 1 131 255 255 255 196 195 4 5 6 131 97 98 99]           |
| \[][]string{{"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}, {"aaa", "bbb", "ccc"}} | [248 65 204 131 97 97 97 131 98 98 98 131 99 99 99 204 131 97 97 97 131 98 98 98 131 99 99 99 204 131 97 97 97 131 98 98 98 131 99 99 99 204 131 97 97 97 131 98 98 98 131 99 99 99 204 131 97 97 97 131 98 98 98 131 99 99 99] |

### 5.8 编码结构体

```go
type simplestruct struct {
	A uint
	B string
}

type recstruct struct {
	I     uint
	Child *recstruct `rlp:"nil"`
}

type intField struct {
	X int
}

type ignoredFiled struct {
	A uint
	B uint `rlp:"-"`
	C uint
}

type tailStruct struct {
	A    uint
	Tail []RawValue `rlp:"tail"`
}

type optionalFields struct {
	A uint
	B uint `rlp:"optional"`
	C uint `rlp:"optional"`
}

type optionalAndTailField struct {
	A    uint
	B    uint   `rlp:"optional"`
	Tail []uint `rlp:"tail"`
}

type optionalBigIntField struct {
	A uint
	B *big.Int `rlp:"optional"`
}

type optionalPtrFiled struct {
	A uint
	B *[3]byte `rlp:"optional"`
}

type optionalPtrFieldNil struct {
	A uint
	B *[3]byte `rlp:"optional,nil"`
}
```

| 原值                                                         | 编码结果                                                     |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| simplestruct{}                                               | [194 128 128]                                                |
| simplestruct{A: 3, B: "abc"}                                 | [197 3 131 97 98 99]                                         |
| simplestruct{A: 326, B: "abc"}                               | [199 130 1 70 131 97 98 99]                                  |
| &recstruct{I: 5, Child: nil}                                 | [194 5 192]                                                  |
| &recstruct{I: 5, Child: &recstruct{I: 5, Child: &recstruct{I: 5, Child: nil}}} | [198 5 196 5 194 5 192]                                      |
| intField{X: 3}                                               | 错误："rlp: type int is not RLP-serializable (struct field rlp.intField.X)" |
| ignoredFiled{A: 1, B: 2, C: 3}                               | [194 1 3]                                                    |
| tailStruct{A: 1, Tail: nil}                                  | [193 1]                                                      |
| tailStruct{A: 1, Tail: []RawValue{{1, 2, 3}}}                | [196 1 1 2 3]                                                |
| optionalFields{A: 1, B: 2, C: 3}                             | [195 1 2 3]                                                  |
| optionalFields{A: 1, B: 0, C: 3}                             | [195 1 128 3]                                                |
| optionalFields{A: 1, B: 2}                                   | [194 1 2]                                                    |
| optionalFields{A: 1, C: 3}                                   | [195 1 128 3]                                                |
| optionalFields{A: 1, B: 2, C: 0}                             | [194 1 2]                                                    |
| &optionalAndTailField{A: 1, B: 2}                            | [194 1 2]                                                    |
| &optionalAndTailField{A: 1}                                  | [193 1]                                                      |
| &optionalAndTailField{A: 1, B: 2, Tail: []uint{3, 4}}        | [196 1 2 3 4]                                                |
| &optionalAndTailField{A: 1, Tail: []uint{3, 4}}              | [196 1 128 3 4]                                              |
| &optionalAndTailField{A: 1}                                  | [193 1]                                                      |
| &optionalPtrFiled{A: 1}                                      | [193 1]                                                      |
| optionalPtrFiled{A: 1, B: &[3]byte{1, 2, 3}}                 | [197 1 131 1 2 3]                                            |
| &optionalPtrFieldNil{A: 1}                                   | [193 1]                                                      |

