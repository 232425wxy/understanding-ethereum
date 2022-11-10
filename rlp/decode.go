package rlp

import (
	"errors"
	"fmt"
	"github.com/232425wxy/understanding-ethereum/rlp/internal/rlpstruct"
	"reflect"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义全局错误

var (
	ErrCanonSize      = errors.New("rlp: non-canonical size information")
	ErrValueTooLarge  = errors.New("rlp: value size exceeds available input length")
	ErrExpectedString = errors.New("rlp: expected String or Byte")
	ErrExpectedList   = errors.New("rlp: expected List")
	ErrCanonInt       = errors.New("rlp: non-canonical integer format")
)

// 定义内部错误

var (
	errUintOverflow = errors.New("rlp: uint overflow")
	errNotAtEOL     = errors.New("rlp: call of ListEnd not positioned at EOL")
)

// 自定义错误类型

// decodeError ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// decodeError 定义解码时可能遇到的错误
type decodeError struct {
	msg string
	typ reflect.Type
	ctx []string
}

func (err *decodeError) Error() string {
	ctx := ""
	if len(err.ctx) > 0 {
		ctx = ", decoding into "
		for i := len(err.ctx) - 1; i >= 0; i-- {
			ctx += err.ctx[i]
		}
	}
	return fmt.Sprintf("rlp: %s for %v%s", err.msg, err.typ, ctx)
}

// addErrorContext ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// addErrorContext 该方法接受两个参数：error 和一个字符串ctx，如果给定的error的类型是 *decodeError，
// 则将参数ctx添加到 *decodeError.ctx 中。
func addErrorContext(err error, ctx string) error {
	if decErr, ok := err.(*decodeError); ok {
		decErr.ctx = append(decErr.ctx, ctx)
	}
	return err
}

// wrapStreamError ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// wrapStreamError 方法接受两个入参：error 和 reflect.Type，如果给定的 error 属于以下自定义的错误：
//
//	ErrCanonInt、ErrCanonSize、ErrExpectedList、ErrExpectedString、errUintOverflow、errNotAtEOL
//
// 则将给定的错误包装成 *decodeError。
func wrapStreamError(err error, typ reflect.Type) error {
	switch err {
	case ErrCanonInt:
		return &decodeError{msg: "non-canonical integer (leading zero bytes)", typ: typ}
	case ErrCanonSize:
		return &decodeError{msg: "non-canonical size information", typ: typ}
	case ErrExpectedList:
		return &decodeError{msg: "expected input list", typ: typ}
	case ErrExpectedString:
		return &decodeError{msg: "expected input string or byte", typ: typ}
	case errUintOverflow:
		return &decodeError{msg: "input string too long", typ: typ}
	case errNotAtEOL:
		return &decodeError{msg: "input list has too many elements", typ: typ}
	}
	return err
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义 Decoder 接口

// Decoder ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 那些实现 Decoder 接口的类型，可以自定义解码规则。
type Decoder interface {
	DecodeRLP(*Stream) error
}

var decoderInterface = reflect.TypeOf(new(Decoder)).Elem()

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Stream ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// Stream
type Stream struct {
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义 Kind 类型，Kind 类型标示出了编码数据时所面临的不同规则。

type Kind int8

const (
	Byte Kind = iota
	String
	List
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// makeDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// makeDecoder 方法接受两个参数，分别是reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，然后为typ生成专属的
// 解码器，其中tag参数只在为切片、数组和指针类型生成解码器时有用。
func makeDecoder(typ reflect.Type, tag rlpstruct.Tag) (decoder, error) {
	return nil, nil
}
