package rlpstruct

import (
	"fmt"
	"reflect"
	"strings"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Type ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// Type 是我们自己定义的一个结构体，用它来表示变量的类型信息，类似于简化版的 reflect.Type。
type Type struct {
	// Name 用字符串来描述该Type所指向的具体类型，例如"string"、"*[3]uint8"，甚至是我们自己定义的数据类型："math.HexOrDecimal256"
	Name string
	// Kind 用 reflect.Kind 来描述该Type所指向的具体类型，例如 reflect.String、reflect.Slice 等
	Kind reflect.Kind
	// IsEncoder 用来指示该Type描述的数据类型是否实现了 Encoder 接口，即是否实现了 EncodeRLP 方法
	IsEncoder bool
	// IsDecoder 用来指示该Type描述的数据类型是否实现了 Decoder 接口，即是否实现了 DecodeRLP 方法
	IsDecoder bool
	// Elem 如果该Type描述的是一个指针、数组或者切片，那么指针所指向的数据、数组和切片所存储的数据，
	// 这些数据的类型会被递归的获取到，并存储在Elem字段里
	Elem *Type
}

// DefaultNilValue ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// DefaultValue 方法返回 Type 描述的类型默认的零值类型：NilKindString 或者 NilKindList。
func (t Type) DefaultNilValue() NilKind {
	kind := t.Kind
	isString := kind == reflect.String
	isBool := kind == reflect.Bool
	// 判断给定的数据类型是不是无符号整数
	isUint := kind >= reflect.Uint && kind <= reflect.Uintptr
	// 判断给定的数据类型是不是字节数组或切片
	isByteArray := (kind == reflect.Slice || kind == reflect.Array) && ((t.Elem).Kind == reflect.Uint8 && !(t.Elem.IsEncoder))
	if isString || isBool || isUint || isByteArray {
		return NilKindString
	}
	return NilKindList
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Field ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// Field 结构体是我们自定义的，它用来描述目标结构体中某个字段的具体信息，比如我们定一个结构体如下代码所示：
//
//	type Dog struct {
//		Nick string `rlp:"nilString"`
//		Age  uint8
//	}
//
// 那么用来描述Dog结构体里Nick字段的 Field 实例应该是这样的：
//
//	Field{Name: "Nick", Index: 0, Exported: true, Type: {Name: "string", Kind: reflect.String, IsEncoder: false, IsDecoder: false, Elem: nil}, Tag: `rlp:"nilString"`}
type Field struct {
	Name     string
	Index    int
	Exported bool
	Type     Type
	Tag      string
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// NilKind ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// NilKind 用来表示不同类型的数据在它们取值为零值时，该被看作是什么样的零值，在rlp编码中，零值类型只有两种，
// 一种是被标记为 NilKindString 的空字符串类型，另一种是被标记为 NilKindList 的空列表类型。
type NilKind uint8

const (
	NilKindString NilKind = 0x80
	NilKindList   NilKind = 0xC0
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Tag ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// Tag 用来表示我们自定义的结构体中字段的tag值，例如`json:"name"` 或者 `rlp:"-"`等。
type Tag struct {
	// NilKind 我们在定义结构体字段时，可能会在tag处手动设置该字段的NilKind，例如：
	// 	Name string `rlp:"nilString"`
	NilKind NilKind
	// NilManual 如果我们在定义结构体字段时手动的在tag处为其设置了NIlKind，则NilManual会被设置为true
	NilManual bool
	// Optional 如果结构体字段的tag被设置为`rlp:"optional"`，那么Optional被设置为true。要求如果该结构体的编码规则被设置
	// 为`rlp:"optional"`，则其后的所有字段的rlp编码规则都必须被设置为`rlp:"optional"`，但是，有一种情况除外，那就是如果
	// 结构体的最后一个字段是一个切片，并且想为该字段设置"tail"，则不必为该字段的tag再设置"optional"。编码规则被设置为
	// `rlp:"optional"`的字段，在编码时，如果该字段的值等于零值，则不被编码，并且其后的所有字段都不会被编码（即使存在值为非空的字段）。
	Optional bool
	// Tail 如果结构体字段的tag被设置为`rlp:"tail"`，那么Tail被设置为true。只有结构体最后一个可导出且类型必须是切片类型的字段的编码
	// 规则才能被设置为`rlp:"tail"`，在对切片类型的数据编码时，数据会被看成是一个列表，如果该数据的编码规则被设置为`rlp:"tail"`，
	// 那么就不会再将其看成是列表，而是把列表里面的数据拎出来逐个进行编码。
	Tail bool
	// Ignored 结构体字段的编码规则如果被设置成`rlp:"-"`，那么Ignored被设置为true。编码规则被设置为`rlp:"-"`的字段不参与编码。
	Ignored bool
}

// TagError ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// TagError 定义了结构体字段在设置tag时可能出现的错误，这里我们只针对`rlp:"xxx"`这样的tag，像json这样的tag我们不关注。
type TagError struct {
	// StructType 指出那个结构体的tag设置出了错误
	StructType string
	// Field 进一步指出结构体哪个字段的tag设置出了错误
	Field string
	// Tag 用来显示被设置错误的tag长什么样
	Tag string
	Err string
}

// Error ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// 该方法用来实现 error 接口。
func (e TagError) Error() string {
	field := "field " + e.Field
	if e.StructType != "" {
		field = e.StructType + "." + e.Field
	}
	return fmt.Sprintf("rlp: invalid struct tag %q for %s (%s)", e.Tag, field, e.Err)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// ProcessFields ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// ProcessFields 方法用来处理自定义结构体里的字段信息，该方法接受一个入参：[]Field，该入参表示某个结构体的所有字段信息，
// 然后 ProcessFields 方法就是筛选出所有会参与rlp编码的字段，并解析这些字段的tag，得到经过过滤的[]Field和新的[]Tag。
func ProcessFields(allFields []Field) ([]Field, []Tag, error) {
	lastPublic := lastPublicField(allFields)
	var fields []Field
	var tags []Tag
	for _, field := range allFields {
		if !field.Exported {
			continue
		}
		tag, err := parseTag(field, lastPublic)
		if err != nil {
			return nil, nil, err
		}
		if tag.Ignored {
			continue
		}
		fields = append(fields, field)
		tags = append(tags, tag)
	}
	var anyOptional bool
	var firstOptionalName string
	for i, tag := range tags {
		name := fields[i].Name
		if tag.Optional || tag.Tail {
			if !anyOptional {
				firstOptionalName = name
			}
			anyOptional = true
		} else {
			if anyOptional {
				return nil, nil, TagError{Field: name, Err: fmt.Sprintf("must be optional because preceding field %q is optional", firstOptionalName)}
			}
		}
	}
	return fields, tags, nil
}

// parseTag ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// 该方法用来解析结构体字段的tag值，该方法接受两个参数，第一个参数是 Field 实例，用来描述结构体字段信息，第二个参数是
// lastPublic，用来表示结构体中最后一个可导出字段的索引值，它被用来验证tag被设置为tail的字段是否是结构体中最后一个可
// 导出字段。
func parseTag(field Field, lastPublic int) (Tag, error) {
	tag := reflect.StructTag(field.Tag)
	var result Tag
	for _, t := range strings.Split(tag.Get("rlp"), ",") {
		switch t = strings.TrimSpace(t); t {
		case "":
		// 没有为该字段设置tag
		case "-":
			result.Ignored = true
		case "nil", "nilString", "nilList":
			result.NilManual = true
			if field.Type.Kind != reflect.Ptr {
				// 只有指针类型的结构体字段才有资格在tag里设置空值类型
				return result, TagError{Field: field.Name, Tag: t, Err: "field is not a pointer"}
			}
			switch t {
			case "nil":
				result.NilKind = field.Type.Elem.DefaultNilValue()
			case "nilString":
				result.NilKind = NilKindString
			case "nilList":
				result.NilKind = NilKindList
			}
		case "optional":
			result.Optional = true
			if result.Tail {
				return result, TagError{Field: field.Name, Tag: t, Err: `also has "tail" tag`}
			}
		case "tail":
			result.Tail = true
			if field.Index != lastPublic {
				return result, TagError{Field: field.Name, Tag: t, Err: `tag "tail" is only allowed to be set on the last exportable field`}
			}
			if result.Optional {
				return result, TagError{Field: field.Name, Tag: t, Err: `also has "optional" tag`}
			}
			if field.Type.Kind != reflect.Slice {
				return result, TagError{Field: field.Name, Tag: t, Err: `tag "tail" is only allowed to be set on the slice type field`}
			}
		default:
			return result, TagError{Field: field.Name, Tag: t, Err: "unknown tag"}
		}
	}
	return result, nil
}

// lastPublicField ♏ |作者：吴翔宇| 🍁 |日期：2022/10/29|
//
// lastPublicField 方法从给定结构体的一众字段中返回最后一个可导出字段在结构体所有字段中的索引值（定义顺序）。
func lastPublicField(fields []Field) int {
	last := 0
	for _, f := range fields {
		if f.Exported {
			last = f.Index
		}
	}
	return last
}
