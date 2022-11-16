/*
RLP编码数据由两部分组成：编码前缀（Encoding Prefix，EP）和编码内容（Encoding Content，EC），
其中编码前缀EP由类型标记位（Type Marker Bit，TMB）和一个可选的长度编码（Optional Length Coding，OLC）组成，
这部分内容在README里有详细介绍。
*/

package rlp

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/232425wxy/understanding-ethereum/rlp/internal/rlpstruct"
	"io"
	"math/big"
	"reflect"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API

// Decode ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// Decode 方法接受两个参数，第一个参数是一个 io.Reader，RLP编码数据被存储在里面，第二个参数是一个指针，
// 将被编码的数据解码到该指针里面。
func Decode(r io.Reader, val interface{}) error {
	stream := streamPool.Get().(*Stream)
	defer streamPool.Put(stream)
	stream.Reset(r, 0)
	return stream.Decode(val)
}

// DecodeBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// DecodeBytes 方法接受两个参数，第一个参数是一个字节切片，里面存储了原始的RLP编码数据，第二个参数是一个指针，
// 将被编码的数据解码到该指针里面。
func DecodeBytes(bz []byte, val interface{}) error {
	r := bytes.NewReader(bz)
	stream := streamPool.Get().(*Stream)
	defer streamPool.Put(stream)
	stream.Reset(r, uint64(len(bz)))
	if err := stream.Decode(val); err != nil {
		return err
	}
	if r.Len() > 0 {
		return ErrMoreThanOneValue
	}
	return nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// EOL ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// EOL "end of list"
var EOL = errors.New("rlp: end of list")

// 定义全局错误

var (
	ErrCanonSize        = errors.New("rlp: non-canonical size information")
	ErrExpectedString   = errors.New("rlp: expected String or Byte")
	ErrExpectedList     = errors.New("rlp: expected List")
	ErrCanonInt         = errors.New("rlp: non-canonical integer format")
	ErrElemTooLarge     = errors.New("rlp: element is larger than containing list")
	ErrValueTooLarge    = errors.New("rlp: value size exceeds available input length")
	ErrMoreThanOneValue = errors.New("rlp: input contains more than one value")
)

// 定义内部错误

var (
	errUintOverflow  = errors.New("rlp: uint overflow")
	errNotAtEOL      = errors.New("rlp: call of ListEnd not positioned at EOL")
	errDecodeIntoNil = errors.New("rlp: pointer given to Decode must not be nil")
	errNoPointer     = errors.New("rlp: interface given to Decode must be a pointer")
	errNotInList     = errors.New("rlp: call of ListEnd outside of any list")
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

var EmptyString = []byte{0x80}
var EmptyList = []byte{0xC0}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// makeDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// makeDecoder 方法接受两个参数，分别是reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，然后为typ生成专属的
// 解码器，其中tag参数只在为切片、数组和指针类型生成解码器时有用。
func makeDecoder(typ reflect.Type, tag rlpstruct.Tag) (decoder, error) {
	kind := typ.Kind()
	switch {
	case typ == rawValueType:
		return decodeRawValue, nil
	case typ.AssignableTo(reflect.PtrTo(reflect.TypeOf(big.Int{}))):
		return decodeBigIntPtr, nil
	case typ.AssignableTo(reflect.TypeOf(big.Int{})):
		return decodeBigIntNoPtr, nil
	case reflect.PtrTo(typ).Implements(decoderInterface):
		return decodeDecoder, nil
	case isUint(kind):
		return decodeUint, nil
	case kind == reflect.Bool:
		return decodeBool, nil
	case kind == reflect.String:
		return decodeString, nil
	case kind == reflect.Interface:
		return decodeInterface, nil
	case kind == reflect.Struct:
		return makeStructDecoder(typ)
	case kind == reflect.Slice || kind == reflect.Array:
		return makeListDecoder(typ, tag)
	case kind == reflect.Pointer:
		return makePtrDecoder(typ, tag)
	default:
		return nil, fmt.Errorf("rlp: type %v is not RLP-serializable", typ)
	}
}

// decodeString ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeString
func decodeString(s *Stream, val reflect.Value) error {
	b, err := s.Bytes()
	if err != nil {
		return wrapStreamError(err, val.Type())
	}
	val.SetString(string(b))
	return nil
}

// decodeBigIntNoPtr ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeBigIntNoPtr
func decodeBigIntNoPtr(s *Stream, val reflect.Value) error {
	return decodeBigIntPtr(s, val.Addr())
}

// decodeBigInt ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeBigInt 方法实现了 decoder 函数句柄，该方法解码rlp编码内容为 *big.Int。
// 这个方法似乎在单独解码指针类型的 big.Int 时确实调用不到，但是，如果某个结构体中含有
// 指针类型的 big.Int 字段，则在解码该结构体的后续迭代过程中，可能会调用该方法来解码该
// 结构体中的 *big.Int 字段。
func decodeBigIntPtr(s *Stream, val reflect.Value) error {
	x := val.Interface().(*big.Int)
	if x == nil {
		x = new(big.Int)
		val.Set(reflect.ValueOf(x))
	}
	err := s.decodeBigInt(x)
	if err != nil {
		return wrapStreamError(err, val.Type())
	}
	return nil
}

// decodeRawValue ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeRawValue 方法实现 decoder 函数句柄，读取stream底层的输入，将其解码为 RawValue。
func decodeRawValue(s *Stream, val reflect.Value) error {
	r, err := s.Raw()
	if err != nil {
		return err
	}
	val.SetBytes(r)
	return nil
}

// decodeUint ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeUint 方法实现了 decoder 函数句柄，读取stream底层的输入，将其解码为无符号整数。
func decodeUint(s *Stream, val reflect.Value) error {
	typ := val.Type()
	// typ.Bits() 可以精确返回该整数类型占用多少个比特空间，例如uint32类型的整数就占用32个比特空间，
	// 只能计算整数类型、浮点数类型或者复数类型的空间大小，其他数据类型调用此方法会panic。
	num, err := s.uint(typ.Bits())
	if err != nil {
		return wrapStreamError(err, val.Type())
	}
	val.SetUint(num)
	return nil
}

// decodeBool ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeBool 方法实现了 decoder 函数句柄，读取stream底层的输入，将其解码为bool类型。
func decodeBool(s *Stream, val reflect.Value) error {
	b, err := s.bool()
	if err != nil {
		return wrapStreamError(err, val.Type())
	}
	val.SetBool(b)
	return nil
}

// makeListDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// makeListDecoder
func makeListDecoder(typ reflect.Type, tag rlpstruct.Tag) (decoder, error) {
	// 获取列表中元素类型
	eTyp := typ.Elem()
	if eTyp.Kind() == reflect.Uint8 && !reflect.PtrTo(eTyp).Implements(decoderInterface) {
		if typ.Kind() == reflect.Array {
			return decodeByteArray, nil
		}
		return decodeByteSlice, nil
	}
	// 如果是非字节数组或者字节切片，就要根据数组和切片中存储的数据类型来生成对应的解码器了
	info := theTC.infoWhileGenerating(eTyp, rlpstruct.Tag{})
	if info.decoderErr != nil {
		return nil, info.decoderErr
	}
	var d decoder
	switch {
	case typ.Kind() == reflect.Array:
		d = func(stream *Stream, value reflect.Value) error {
			return decodeListArray(stream, value, info.decoder)
		}
	case tag.Tail:
		d = func(stream *Stream, value reflect.Value) error {
			return decodeSliceElems(stream, value, info.decoder)
		}
	default:
		d = func(stream *Stream, value reflect.Value) error {
			return decodeListSlice(stream, value, info.decoder)
		}
	}
	return d, nil
}

// makeStructDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// makeStructDecoder
func makeStructDecoder(typ reflect.Type) (decoder, error) {
	fields, err := processStructFields(typ)
	if err != nil {
		return nil, err
	}
	// 排除错误
	for _, f := range fields {
		if f.info.decoderErr != nil {
			return nil, structFieldError{typ: typ, fieldIndex: f.index, err: f.info.decoderErr}
		}
	}
	var d decoder = func(stream *Stream, value reflect.Value) error {
		if _, err = stream.ListStart(); err != nil {
			return wrapStreamError(err, typ)
		}
		for i, f := range fields {
			err = f.info.decoder(stream, value.Field(f.index))
			if err == EOL {
				if f.optional {
					// optional后面的字段都设置为零值
					for _, fi := range fields[i:] {
						fv := value.Field(fi.index)
						fv.Set(reflect.Zero(fv.Type()))
					}
					break
				}
				// 列表里面的数据读完了，但是结构体里的数据还没填充完，说明rlp编码数据太少了
				return &decodeError{msg: "too few elements", typ: typ}
			} else if err != nil {
				return addErrorContext(err, "."+typ.Field(f.index).Name)
			}
		}
		return wrapStreamError(stream.ListEnd(), typ)
	}
	return d, nil
}

// makePtrDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// makePtrDecoder
func makePtrDecoder(typ reflect.Type, tag rlpstruct.Tag) (decoder, error) {
	eTyp := typ.Elem()
	info := theTC.infoWhileGenerating(eTyp, rlpstruct.Tag{})
	switch {
	case info.decoderErr != nil:
		return nil, info.decoderErr
	case !tag.NilManual:
		return makeSimplePtrDecoder(eTyp, info), nil
	default:
		return makeNilPtrDecoder(eTyp, info, tag), nil
	}
}

// makeSimplePtrDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// makeSimplePtrDecoder
func makeSimplePtrDecoder(eTyp reflect.Type, info *typeInfo) decoder {
	return func(stream *Stream, value reflect.Value) error {
		newVal := value
		if value.IsNil() {
			newVal = reflect.New(eTyp)
		}
		if err := info.decoder(stream, newVal.Elem()); err == nil {
			value.Set(newVal)
		} else {
			return err
		}
		return nil
	}
}

// makeNilPtrDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// makeNilPtrDecoder
func makeNilPtrDecoder(eTyp reflect.Type, info *typeInfo, tag rlpstruct.Tag) decoder {
	typ := reflect.PtrTo(eTyp)
	nilPtr := reflect.Zero(typ)
	nilKind := typeNilKind(eTyp, tag)

	return func(stream *Stream, value reflect.Value) error {
		kind, size, err := stream.Kind()
		if err != nil {
			value.Set(nilPtr)
			return wrapStreamError(err, typ)
		}
		if kind != Byte && size == 0 {
			if kind != nilKind {
				return &decodeError{msg: fmt.Sprintf("wrong kind of empty value (got %v, want %v)", kind, nilKind), typ: typ}
			}
			stream.kind = -1
			value.Set(nilPtr)
			return nil
		}
		newVal := value
		if value.IsNil() {
			newVal = reflect.New(eTyp)
		}
		if err = info.decoder(stream, newVal.Elem()); err == nil {
			value.Set(newVal)
		} else {
			return err
		}
		return nil
	}
}

// decodeInterface ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeInterface
func decodeInterface(s *Stream, val reflect.Value) error {
	// 只能编码方法数为0的接口
	if val.Type().NumMethod() != 0 {
		return fmt.Errorf("rlp: type %v is not RLP-serializable", val.Type())
	}
	kind, _, err := s.Kind()
	if err != nil {
		return err
	}
	if kind == List {
		slice := reflect.New(reflect.TypeOf([]interface{}{})).Elem()
		if err = decodeListSlice(s, slice, decodeInterface); err != nil {
			return err
		}
		val.Set(slice)
	} else {
		b, err := s.Bytes()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(b))
	}
	return nil
}

// decodeDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeDecoder
func decodeDecoder(s *Stream, val reflect.Value) error {
	return val.Addr().Interface().(Decoder).DecodeRLP(s)
}

// decodeByteSlice ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeByteSlice
func decodeByteSlice(s *Stream, val reflect.Value) error {
	b, err := s.Bytes()
	if err != nil {
		return wrapStreamError(err, val.Type())
	}
	val.SetBytes(b)
	return nil
}

// decodeByteArray ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeByteArray
func decodeByteArray(s *Stream, val reflect.Value) error {
	kind, size, err := s.Kind()
	if err != nil {
		return err
	}
	slice := byteArrayBytes(val, val.Len())
	switch kind {
	case Byte:
		if len(slice) == 0 {
			return &decodeError{msg: "input string too long", typ: val.Type()}
		} else if len(slice) > 1 {
			return &decodeError{msg: "input string too short", typ: val.Type()}
		}
		slice[0] = s.byteVal
		s.kind = -1
	case String:
		if uint64(len(slice)) < size {
			return &decodeError{msg: "input string too long", typ: val.Type()}
		}
		if uint64(len(slice)) > size {
			return &decodeError{msg: "input string too short", typ: val.Type()}
		}
		if err = s.readFull(slice); err != nil {
			return err
		}
		if size == 1 && slice[0] < 0x80 {
			return wrapStreamError(ErrCanonSize, val.Type())
		}
	case List:
		return wrapStreamError(ErrExpectedString, val.Type())
	}
	return nil
}

// decodeListArray ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeListArray
func decodeListArray(s *Stream, val reflect.Value, elemDec decoder) error {
	if _, err := s.ListStart(); err != nil {
		return wrapStreamError(err, val.Type())
	}
	length := val.Len()
	i := 0
	for ; i < length; i++ {
		if err := elemDec(s, val.Index(i)); err == EOL {
			break
		} else if err != nil {
			return addErrorContext(err, fmt.Sprintf("[%d]", i))
		}
	}
	if i < length {
		return &decodeError{msg: "input list has too few elements", typ: val.Type()}
	}
	// 如果此时EC部分还有数据没有被读取完毕，则ListEnd方法会报错
	return wrapStreamError(s.ListEnd(), val.Type())
}

// decodeSliceElems ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeSliceElems
func decodeSliceElems(s *Stream, val reflect.Value, elemDec decoder) error {
	i := 0
	for ; ; i++ {
		if i >= val.Cap() {
			newCap := val.Cap() + val.Cap()/2
			if newCap < 4 {
				newCap = 4
			}
			newVal := reflect.MakeSlice(val.Type(), val.Len(), newCap)
			reflect.Copy(newVal, val)
			val.Set(newVal)
		}
		if i >= val.Len() {
			val.SetLen(i + 1)
		}
		if err := elemDec(s, val.Index(i)); err == EOL {
			break
		} else if err != nil {
			return addErrorContext(err, fmt.Sprint("[", i, "]"))
		}
	}
	if i < val.Len() {
		val.SetLen(i)
	}
	return nil
}

// decodeListSlice ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// decodeListSlice
func decodeListSlice(s *Stream, val reflect.Value, elemDec decoder) error {
	size, err := s.ListStart()
	if err != nil {
		return wrapStreamError(err, val.Type())
	}
	if size == 0 {
		val.Set(reflect.MakeSlice(val.Type(), 0, 0))
		return s.ListEnd()
	}
	if err = decodeSliceElems(s, val, elemDec); err != nil {
		return err
	}
	return s.ListEnd()
}
