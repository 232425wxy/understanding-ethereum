package rlp

import (
	"github.com/232425wxy/understanding-ethereum/rlp/internal/rlpstruct"
	"io"
	"reflect"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义 Encoder 接口

// Encoder ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 那些实现 Encoder 接口的类型，可以自定义编码规则。
type Encoder interface {
	EncodeRLP(io.Writer) error
}

var encoderInterface = reflect.TypeOf(new(Encoder)).Elem()

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// listHead ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// listHead 存储了一个列表头的信息，官方源码的写法是"listhead"，可是这在goland编辑器里，会显示波浪线，看着很遭心，所以我改成了"listHead"。
type listHead struct {
	offset int
	size   int
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// makeWriter ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// makeWriter 方法接受两个参数，分别是reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，然后为typ生成专属的
// 编码器，其中tag参数只在为元素为非byte类型的切片、数组和指针类型生成编码器时有用。
func makeWriter(typ reflect.Type, tag rlpstruct.Tag) (writer, error) {
	return nil, nil
}
