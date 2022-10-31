package rlp

import (
	"fmt"
	"github.com/232425wxy/understanding-ethereum/rlp/internal/rlpstruct"
	"reflect"
	"sync"
	"sync/atomic"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义编码器和解码器

// writer ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// writer 是一个函数类型，编码时会遇到各种各样的数据类型，为此需要针对不同的数据类型设计不同的编码规则。
type writer func(reflect.Value, *encBuffer) error

// decoder ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// decoder 是一个函数类型，解码时会遇到各种各样的数据类型，为此需要针对不同的数据类型设计不同的解码规则。
type decoder func(*Stream, reflect.Value) error

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// typeInfo ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// typeInfo 结构体是专为不同数据类型定义的，一个 typeInfo 实例专门维护了针对某个特定数据类型的编码器和解码器。
// 官方源码的写法是"typeinfo"，可是这样在goland里会显示波浪线，看着很遭心，所以我改成了"typeInfo"。
type typeInfo struct {
	decoder    decoder
	decoderErr error // 在为某个特定的数据类型生成解码器时遇到的错误
	writer     writer
	writerErr  error // 在为某个特定的数据类型生成编码器时遇到的错误
}

// makeDecoderAndWriter ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// makeDecoderAndWriter 该方法接受两个参数，分别是reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，
// 然后调用 makeDecoder 和 makeWriter 来为typ所代表的数据类型生成专有的编解码器。官方源码的写法是"generate"，我将
// 它改成了"makeDecoderAndWriter"。
func (ti *typeInfo) makeDecoderAndWriter(typ reflect.Type, tag rlpstruct.Tag) {
	ti.decoder, ti.decoderErr = makeDecoder(typ, tag)
	ti.writer, ti.writerErr = makeWriter(typ, tag)
}

// typeKey ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// 我们在观察 typeInfo 结构体时，不禁会发出一个疑问，那就是我们都知道了 typeInfo 是为了某些特定的数据类型而设计
// 的用来维护编解码器的结构体，但是在 typeInfo 结构体里我们并没有发现存储数据类型的信息，为此，typeKey 结构体被
// 设计了出来，typeKey 用来存储数据类型的详细信息，实际上，typeKey 与 typeInfo 是成对出现的，它们被分别作为key
// 和value存储在一个map里：typeKey->typeInfo，所以当我们想要为某个特定的数据类型进行编码时，拿到 typeKey 就可
// 以找到对应的 typeInfo，然后利用 typeInfo.writer 对数据进行编码，解码的过程也是一样的。
// 官方的写法是"typekey"，我将其改成了"typeKey"，强迫症啊！
type typeKey struct {
	reflect.Type
	rlpstruct.Tag
}

// typeCache ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// 我们前面介绍了 typeInfo 和 typeKey，然后知道了 typeKey 和 typeInfo 分别被作为key和value存储在map里，这不，
// 那个所谓的map就被定义在 typeCache 结构体里，从它的名字“类型缓存”也能看出来它存储了乙太坊运行过程中所遇到的所有需
// 要经历rlp编码的类型信息，和对应的编解码器。
type typeCache struct {
	cur  atomic.Value
	mu   sync.Mutex
	next map[typeKey]*typeInfo
}

// cachedWriter ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// cachedWriter 方法接受一个参数，那就是 reflect.Type 类型的typ，然后该方法从缓冲区获取针对该typ的 typeInfo 实
// 例，缓冲区里面没有也可以，它会现场生成，然后返回针对该typ的编码器和生成编码器时可能产生的错误。
func cachedWriter(typ reflect.Type) (writer, error) {
	info := theTC.info(typ)
	return info.writer, info.writerErr
}

// cachedDecoder ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// cachedDecoder 方法接受一个参数，那就是 reflect.Type 类型的typ，然后该方法从缓冲区获取针对该typ的 typeInfo 实
// 例，缓冲区里面没有也可以，它会现场生成，然后返回针对该typ的解码器和生成解码器时可能产生的错误。
func cachedDecoder(typ reflect.Type) (decoder, error) {
	info := theTC.info(typ)
	return info.decoder, info.decoderErr
}

// info ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// info 方法接受一个参数，那就是 reflect.Type 类型的typ，info 方法利用typ构建 typeKey 实例：
//
//	key := typeKey{Type: typ}
//
// 然后利用这个key到 typeCache.cur 缓存区中寻找对应的 typeInfo 实例，如果找不到，就调用 typeCache.generate 方法，
// 即时生成一个针对typ的 typeInfo 实例，注意，typeCache.generate 方法接受两个参数，分别是 reflect.Type 类型的typ，
// 另一个是 rlpstruct.Tag 类型的 tag，第一个参数就沿用 info 方法的typ，至于第二个参数，就用一个空的 rlpstruct.Tag{}。
func (tc *typeCache) info(typ reflect.Type) *typeInfo {
	key := typeKey{Type: typ}
	cur := tc.cur.Load().(map[typeKey]*typeInfo)
	if info := cur[key]; info != nil {
		return info
	}
	// 缓存区里没有，需要现在立马为给定的typ生成对应的 typeInfo
	return tc.generate(typ, rlpstruct.Tag{})
}

// generate ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// generate 方法接受两个参数，分别是 reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，该方法的最终目的
// 就是获得针对typ的 typeInfo 实例，它先从缓存区的 typeCache.cur 里面寻找存不存在针对typ的 typeInfo，如果没有的
// 话，就先把 typeCache.cur 里面的内容搬到 typeCache.next 里面，然后调用 typeCache.infoWhileGenerating 方法
// 现场生成针对typ的 typeInfo，在 typeCache.infoWhileGenerating 方法里，新生成的 typeInfo 实例会被存到 typeCache.next
// 里，然后再把 typeCache.next 赋值给 typeCache.cur，为什么先把 cur 里的内容搬到 next 里，再把 next 赋值给 cur
// 呢？官方的设计很耐人寻味，何不直接设计一个支持多线程安全的map来存储 typeKey->typeInfo 呢？最后新生成的 typeInfo
// 实例被作为方法的唯一返回参数返回出去。
func (tc *typeCache) generate(typ reflect.Type, tag rlpstruct.Tag) *typeInfo {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	key := typeKey{typ, tag}
	cur := tc.cur.Load().(map[typeKey]*typeInfo)
	if info := cur[key]; info != nil {
		// 先从cur里面找一找
		return info
	}
	tc.next = make(map[typeKey]*typeInfo, len(cur)+1)
	for k, v := range cur {
		tc.next[k] = v
	}
	info := tc.infoWhileGenerating(typ, tag)
	tc.cur.Store(tc.next)
	tc.next = nil // 将 tc.next 设置为nil，不会影响到 tc.cur
	return info
}

// infoWhileGenerating ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// infoWhileGenerating 方法接受两个参数，分别是reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，该
// 方法先将typ和tag组成一个typeKey，然后利用这个key去 typeCache 里寻找对应的 typeInfo，耐人寻味的是，它从 typeCache.next
// 里去寻找，而不是 typeCache.cur，如果找不到的话，就调用 typeInfo 的 makeDecoderAndWriter 方法即时为 typ 生成
// 专属的编解码器，生成的新 typeInfo 会先被存到 typeCache.next 里，然后再作为函数的返回参数返回出去。
func (tc *typeCache) infoWhileGenerating(typ reflect.Type, tag rlpstruct.Tag) *typeInfo {
	key := typeKey{typ, tag}
	if info := tc.next[key]; info != nil {
		// 如果缓存区有针对给定的typ的 typeInfo，则直接返回用这个 typeInfo
		return info
	}
	// 目前缓存区没有针对给定typ的 typeInfo，只能现场生成了
	info := new(typeInfo)
	tc.next[key] = info
	info.makeDecoderAndWriter(typ, tag)
	return info
}

// theTC ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// theTC 是一个包级别的全局变量，可以实现在rlp包里任何地方被使用，它其实是 typeCache 的一个实例，准确来说，该变量存储
// 了乙太坊在运行过程中所遇到的所有需要被rlp编码的数据类型和对应的编解码器。
var theTC = newTypeCache()

// newTypeCache ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// newTypeCache
func newTypeCache() *typeCache {
	c := new(typeCache)
	c.cur.Store(make(map[typeKey]*typeInfo))
	return c
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 处理结构体所有的可导出字段

// field ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 定义 field 结构体是为了方便管理自定义结构体里所有可导出字段的编解码问题。
type field struct {
	index    int
	info     *typeInfo // 存储了针对该字段的编解码器
	optional bool
}

// processStructFields ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// processStructFields 方法接受某个结构体的 reflect.Type，然后基于此来处理给定的结构体里所有可导出字段，包括每个字段
// 的tag，最终目的是为了获取所有参与
func processStructFields(typ reflect.Type) (fields []field, err error) {
	// 将结构体的字段转换为 rlpstruct.Field
	var allStructFields []rlpstruct.Field
	for i := 0; i < typ.NumField(); i++ {
		rf := typ.Field(i)
		allStructFields = append(allStructFields, rlpstruct.Field{
			Name:     rf.Name,
			Index:    i,
			Exported: rf.IsExported(),
			Type:     *reflectTypeToRLPType(rf.Type, nil),
			Tag:      string(rf.Tag),
		})
	}
	// 过滤和验证结构体的所有字段
	structFields, structTags, err := rlpstruct.ProcessFields(allStructFields)
	if err != nil {
		if tagErr, ok := err.(rlpstruct.TagError); ok {
			tagErr.StructType = typ.String()
			return nil, tagErr
		}
		return nil, err
	}
	// 为结构体里每个字段生成对应的编解码器
	for i, sf := range structFields {
		t := typ.Field(sf.Index).Type
		tag := structTags[i]
		info := theTC.infoWhileGenerating(t, tag)
		fields = append(fields, field{sf.Index, info, tag.Optional})
	}
	return fields, nil
}

// reflectTypeToRLPType ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// reflectTypeToRLPType 方法将 reflect.Type 转换为 rlpstruct.Type，官方源码的写法是"rtypeToStructType"。
func reflectTypeToRLPType(typ reflect.Type, rec map[reflect.Type]*rlpstruct.Type) *rlpstruct.Type {
	if typ.Kind() == reflect.Invalid {
		panic("invalid kind")
	}
	if prev := rec[typ]; prev != nil {
		// 从已经注册过的map里面尝试获取针对typ的 rlpstruct.Type
		return prev
	}
	if rec == nil {
		rec = make(map[reflect.Type]*rlpstruct.Type)
	}
	t := &rlpstruct.Type{
		Name:      typ.Name(),
		Kind:      typ.Kind(),
		IsEncoder: typ.Implements(encoderInterface),
		IsDecoder: typ.Implements(decoderInterface),
	}
	rec[typ] = t
	if typ.Kind() == reflect.Array || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Ptr {
		t.Elem = reflectTypeToRLPType(typ.Elem(), rec)
	}
	return t
}

// firstOptionalField ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 该方法返回某个结构体中第一个tag被设置为"optional"的字段的索引值，如果没有字段的tag被设置为"optional"，
// 那么就直接返回给定切片的长度。
func firstOptionalField(fields []field) int {
	for i, f := range fields {
		if f.optional {
			return i
		}
	}
	return len(fields)
}

// typeNilKind ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// typeNilKind 方法接受两个参数，分别是 reflect.Type 类型的typ，另一个是 rlpstruct.Tag 类型的 tag，依据这两
// 个参数，判断typ在编码时的零值类型，要么是 List，要么是 String，该方法仅在为指针类型数据生成编解码器时被调用。
func typeNilKind(typ reflect.Type, tag rlpstruct.Tag) Kind {
	rlpTyp := reflectTypeToRLPType(typ, nil)
	var nilKind rlpstruct.NilKind
	if tag.NilManual {
		// 如果我们自己设定了零值类型
		nilKind = tag.NilKind
	} else {
		nilKind = rlpTyp.DefaultNilValue()
	}
	switch nilKind {
	case rlpstruct.NilKindString:
		return String
	case rlpstruct.NilKindList:
		return List
	default:
		panic("invalid nil kind value")
	}
}

// isUint ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 该反法仅接受一个参数：k reflect.Kind，该方法的目的就是判断给定的 reflect.Kind 是否是无符号整数类型。
func isUint(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uintptr
}

// isByte ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 该方法接受一个参数：typ reflect.Type，该方法判断给定的typ是否是 reflect.Uint8 类型，且必须没有实现 Encoder 接口，
// 因为要是实现了 Encoder 接口，即便给定的数据类型是byte类型，那么我们也没法按照rlp编码规则对该数据进行编解码，只能按照开
// 发者自定义的规则对数据进行编解码。
func isByte(typ reflect.Type) bool {
	return typ.Kind() == reflect.Uint8 && !typ.Implements(encoderInterface)
}

// structFieldError ♏ |作者：吴翔宇| 🍁 |日期：2022/10/31|
//
// 定义该结构体，是为了在生成针对结构体的编解码器时方便对遇到的错误进行统一管理。
type structFieldError struct {
	typ        reflect.Type
	fieldIndex int // 官方源码的写法是："field"，感觉不是很一目了然
	err        error
}

func (e structFieldError) Error() string {
	return fmt.Sprintf("%v (struct field %v.%s)", e.err, e.typ, e.typ.Field(e.fieldIndex).Name)
}
