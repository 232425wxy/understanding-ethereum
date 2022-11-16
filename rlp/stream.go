package rlp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"sync"
)

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

// 定义 ByteReader 接口

// ByteReader ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// ByteReader 接口被例如 bufio.Reader 和 bytes.Reader 实现。这里定义接口的方式与官方源码略有不同，官方源码地址：
//
//	https://github.com/ethereum/go-ethereum/blob/972007a517c49ee9e2a359950d81c74467492ed2/rlp/decode.go#L544
type ByteReader interface {
	Read(p []byte) (n int, err error) // 从源中读取至多len(p)个字节到p中
	ReadByte() (byte, error)          // 每次读取一个字节
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Stream ♏ |作者：吴翔宇| 🍁 |日期：2022/10/30|
//
// Stream
type Stream struct {
	r         ByteReader
	remaining uint64
	// size 只在Kind()方法中被显式赋予非0的值，size 表示某编码头后面跟着多少个字节是由该
	// 编码头主导的，比如某个编码头的值为0x88，那么size应当取值为8
	size         uint64   // size 表示EC的长度，EP||EC表示RLP编码结果，其中EP表示编码前缀，EC表示编码内容
	kindErr      error    // 最近一次调用 readKind 方法时产生的错误
	stack        []uint64 // stack 里面存储的是list的EC长度
	auxiliaryBuf [32]byte // 用于整数解码的辅助缓冲区
	kind         Kind
	byteVal      byte // 类型标签中的值，例如0xC0或者0x87等等
	limited      bool
}

var streamPool = sync.Pool{New: func() interface{} { return new(Stream) }}

// NewStream ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// NewStream 方法接受两个入参：io.Reader 和一个64位无符号整数 inputLimit，这两个参数用来实例化 *Stream，
// *Stream 的读取源 *Stream.r 会被设置为 io.Reader，然后如果 inputLimit 大于0，则 *Stream.limited
// 会被置为 true，而 *Stream.remaining 会被置为 inputLimit，否则 *Stream.remaining 会被设置为 io.Reader
// 的长度
func NewStream(r io.Reader, inputLimit uint64) *Stream {
	s := new(Stream)
	s.Reset(r, inputLimit)
	return s
}

// NewListStream ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// NewListStream 与 NewStream 方法相比，该方法有两处不同，一是 *Stream.kind 被设置为 List，二是 *Stream.size
// 被设置为该方法的第二个入参：inputLimit。值得一提的是，该方法只在测试文件中被调用。
func NewListStream(r io.Reader, inputLimit uint64) *Stream {
	s := new(Stream)
	s.Reset(r, inputLimit)
	s.kind = List
	s.size = inputLimit
	return s
}

// Decode ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// Decode 这个方法非常类似于 json.Unmarshal 方法，接受某个类型的指针，然后将底层stream存储的rlp编码内容解码到
// 给定类型指针指向的空间里。实际上，给定某个类型的指针，我们首先要从 typeCache 缓冲区里寻找针对该类型的解码器，找
// 到的话就直接用，找不到的话就生成一个。
func (s *Stream) Decode(val interface{}) error {
	if val == nil {
		return errDecodeIntoNil
	}
	rVal := reflect.ValueOf(val)
	rTyp := reflect.TypeOf(val)
	if rTyp.Kind() != reflect.Pointer {
		return errNoPointer
	}
	if rVal.IsNil() {
		return errDecodeIntoNil
	}
	// rTyp代表的是一个指针类型
	d, err := cachedDecoder(rTyp.Elem())
	if err != nil {
		return err
	}
	err = d(s, rVal.Elem())
	if decErr, ok := err.(*decodeError); ok && len(decErr.ctx) > 0 {
		decErr.ctx = append(decErr.ctx, fmt.Sprintf("(%v)", rTyp.Elem()))
	}
	return err
}

// Reset ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// Reset 方法接受两个入参：io.Reader 和一个64位无符号整数 inputLimit，这两个参数用来重置 *Stream，
// *Stream 的读取源 *Stream.r 会被 io.Reader 替换，然后如果 inputLimit 大于0，则 *Stream.limited
// 会被置为 true，而 *Stream.remaining 会被置为 inputLimit，否则 *Stream.remaining 会被设置为 io.Reader
// 的长度
func (s *Stream) Reset(r io.Reader, inputLimit uint64) {
	if inputLimit > 0 {
		s.remaining = inputLimit
		s.limited = true
	} else {
		switch br := r.(type) {
		case *bytes.Reader:
			s.remaining = uint64(br.Len())
			s.limited = true
		case *bytes.Buffer:
			s.remaining = uint64(br.Len())
			s.limited = true
		case *strings.Reader:
			s.remaining = uint64(br.Len())
			s.limited = true
		default:
			s.limited = false
		}
	}
	//
	byteReader, ok := r.(ByteReader)
	if !ok {
		// bufio.Reader 实现了 Read 和 ReadByte 两个方法
		byteReader = bufio.NewReader(r)
	}
	s.r = byteReader
	s.stack = s.stack[:0]
	s.size = 0
	s.kind = -1
	s.kindErr = nil
	s.byteVal = 0
	s.auxiliaryBuf = [32]byte{}
}

// ListStart ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// ListStart 官方源码的写法是："List"，我将其改成了："ListStart"，该方法返回的第一个参数表示list
// 编码数据EC部分的长度。
//
// 接下来要解码的数据是一个list的RLP编码结果，在解码前，需要做一些准备工作。
func (s *Stream) ListStart() (size uint64, err error) {
	kind, size, err := s.Kind()
	if err != nil {
		return 0, err
	}
	if kind != List {
		return 0, ErrExpectedList
	}
	if inList, listLimit := s.listLimit(); inList {
		s.stack[len(s.stack)-1] = listLimit - size
	}
	s.stack = append(s.stack, size)
	s.kind = -1
	s.size = 0
	return size, nil
}

// ListEnd ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// ListEnd
func (s *Stream) ListEnd() error {
	if inList, listLimit := s.listLimit(); !inList {
		return errNotInList
	} else if listLimit > 0 {
		return errNotAtEOL
	}
	s.stack = s.stack[:len(s.stack)-1]
	s.kind = -1
	s.size = 0
	return nil
}

// Kind ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// Kind 方法返回下一个编码数据的类型和其EC部分的大小，类型就三类：Byte、String、List。
// 如果每次在 ListStart 方法被调用之后再调用此方法，会从底层stream中读取一个字节的TMB（类型标记位），因此，
// Stream.remaining 和 Stream.stack 里的最后一个元素会被减一。
func (s *Stream) Kind() (kind Kind, size uint64, err error) {
	if s.kind >= 0 {
		return s.kind, s.size, s.kindErr
	}
	// 当我们刚开始初始化Stream的时候，比如给它底层的输入数据是"c80102030405060708"，尽管我们给的是一个list
	// 编码数据，但是此时第一次调用listLimit()方法获得的第一个返回值依然是false
	inList, listLimit := s.listLimit()
	if inList && listLimit == 0 {
		return 0, 0, EOL
	}
	// 在这里会从"c80102030405060708"中读取一个字节的内容
	s.kind, s.size, s.kindErr = s.readKind()
	if s.kindErr == nil {
		if inList && s.size > listLimit {
			s.kindErr = ErrElemTooLarge
		} else if s.limited && s.size > s.remaining {
			s.kindErr = ErrValueTooLarge
		}
	}
	return s.kind, s.size, s.kindErr
}

// readKind ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// readKind 方法从底层stream中读取一个字节，这个字节指示了编码的类型标签，根据这个标签返回编码对象是什么类型的，
// 例如 Byte、String 或 List，第二个参数表示类型标签后面有多少个字节是编码结果，第三个参数是一个error。下面给出一个例
// 子：
//
//	例如底层的stream存储的内容是[248 73 134 229 164 167 233 187 132 3 184 63 228 184 173 229 155 189 229 174 137
//	229 190 189 229 144 136 232 130 165 229 186 144 233 152 179 229 140 186 228 184 137 229 173 157 229 143
//	163 232 161 151 233 129 147 230 157 143 232 138 177 231 164 190 229 140 186 229 164 167 233 151 168 229
//	143 163]
//
// 则运行该方法返回的值将会是：List, 73, nil
// 注意，我们这里读取的数据来获取kind和size，是实实在在的读取出来的，也就是说，读完之后，存储kind和size信息的数据就不再存在于底层的
// stream里了。
func (s *Stream) readKind() (kind Kind, size uint64, err error) {
	b, err := s.readByte()
	if err != nil {
		if len(s.stack) == 0 {
			switch err {
			case io.ErrUnexpectedEOF, ErrValueTooLarge:
				err = io.EOF
			}
		}
		return 0, 0, err
	}
	s.byteVal = 0
	switch {
	case b < 0x80:
		s.byteVal = b
		return Byte, 0, nil
	case b < 0xB8: // 0-55个字节组成的字符串
		return String, uint64(b - 0x80), nil
	case b < 0xC0:
		size, err = s.readUint(b - 0xB7)
		if err == nil && size < 56 {
			err = ErrCanonSize
		}
		return String, size, err
	case b < 0xF8:
		return List, uint64(b - 0xC0), nil
	default:
		size, err = s.readUint(b - 0xF7)
		if err == nil && size < 56 {
			err = ErrCanonSize
		}
		return List, size, err
	}
}

// readUint ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// readUint 接受一个整数size，如果size等于0，该方法将直接返回0和nil，如果等于1，则该方法从 Stream 的底层数据池里
// 读取一个字节的内容，并将该字节转换为uint64类型然后返回，否则就从 Stream 的底层数据池读取 size 个字节，然后将这size
// 个字节解码成uint64类型的整数并返回。例如，我们假设size等于3，然后读取的三个字节分别是：00110000，00100000，00010000，
// 那么我们将这三个字节拼接起来：001100000010000000010000，得到一个24比特位的整数，用十进制表示就是：3153936。
//
//	🚨注意：size的大小不能超过8。
func (s *Stream) readUint(size byte) (uint64, error) {
	switch size {
	case 0:
		s.kind = -1
		return 0, nil
	case 1:
		b, err := s.readByte()
		return uint64(b), err
	default:
		// 无符号整数最多只需要8个字节去存储
		buffer := s.auxiliaryBuf[:8]
		for i := range buffer {
			buffer[i] = 0
		}
		start := int(8 - size)
		if err := s.readFull(buffer[start:]); err != nil {
			return 0, err
		}
		if buffer[start] == 0 {
			return 0, ErrCanonSize
		}
		// binary.BigEndian.Uint64方法要求传入的字节切片长度至少为8
		return binary.BigEndian.Uint64(buffer[:]), nil
	}
}

// readFull ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// readFull 方法接受一个字节切片buf作为入参，然后从底层的stream里读取len(buf)个字节到buf里。
func (s *Stream) readFull(buf []byte) error {
	if err := s.willRead(uint64(len(buf))); err != nil {
		return err
	}
	var n, m int
	var err error
	for n < len(buf) && err == nil {
		// 在不出错的情况下，不用担心读不够n个字节，因为已经通过了willRead方法的验证了
		m, err = s.r.Read(buf[n:])
		n += m
	}
	// 读完了，但是可能也遇到错误了
	if err == io.EOF {
		if n < len(buf) {
			// 读取的字节数不够
			err = io.ErrUnexpectedEOF
		} else {
			// 底层的stream被读完的同时，刚好buf也被填满了，皆大欢喜
			err = nil
		}
	}
	return err
}

// readByte ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// readByte 从底层的stream里面读取一个字节。
func (s *Stream) readByte() (byte, error) {
	if err := s.willRead(1); err != nil {
		return 0, err
	}
	b, err := s.r.ReadByte()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return b, err
}

// willRead ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// willRead 接受一个参数n，它是一个64位无符号整数，该方法会在其他所有read方法调用前调用，目的是为了判断如果
// 读取n个字节会不会出错，比如要读取的内容会不会过大。
func (s *Stream) willRead(n uint64) error {
	s.kind = -1
	if inList, limit := s.listLimit(); inList {
		if n > limit {
			return ErrElemTooLarge
		}
		// 我们将会读取n个字节，这n最多等于limit，也就是s.stack的最后一个元素，那么读完后，我们需要更新一下s.stack的
		// 最后一个元素，他这个最后一个元素代表最内层列表的大小
		s.stack[len(s.stack)-1] = limit - n
	}
	if s.limited {
		if n > s.remaining {
			return ErrValueTooLarge
		}
		s.remaining -= n
	}
	return nil
}

// listLimit ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// listLimit 方法返回两个参数，第一个参数是一个bool类型，如果 *Stream.stack 切片为空，则返回false，否则
// 返回true，第二个参数是一个64位无符号整数类型，返回 *Stream.stack 切片中最后一个元素（整数）。
func (s *Stream) listLimit() (inList bool, limit uint64) {
	if len(s.stack) == 0 {
		return false, 0
	}
	return true, s.stack[len(s.stack)-1]
}

// decodeBigInt ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// decodeBigInt 方法接受一个大整数的指针 *big.Int，底层stream接下来存储的数据是某个大整数rlp编码的内容，
// 该方法的作用就是将stream接下来存储的数据编码成一个大整数对象。
func (s *Stream) decodeBigInt(x *big.Int) error {
	var buffer []byte
	kind, size, err := s.Kind()
	switch {
	case err != nil:
		return err
	case kind == List:
		return ErrExpectedString
	case kind == Byte:
		// 单个ASCII码
		buffer = s.auxiliaryBuf[:1]
		buffer[0] = s.byteVal
		s.kind = -1
	case size == 0:
		s.kind = -1
	case size <= uint64(len(s.auxiliaryBuf)):
		// 256位以内的大整数，避免给buffer分配空间
		buffer = s.auxiliaryBuf[:size]
		if err = s.readFull(buffer); err != nil {
			return err
		}
		if size == 1 && buffer[0] < 0x80 {
			return ErrCanonSize
		}
	default:
		buffer = make([]byte, size)
		if err = s.readFull(buffer); err != nil {
			return err
		}
	}
	if len(buffer) > 0 && buffer[0] == 0 {
		return ErrCanonInt
	}
	x.SetBytes(buffer)
	return nil
}

// Bytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// Bytes 方法返回底层stream中存储的接下来的字符串解码结果，不能是列表数据。
func (s *Stream) Bytes() ([]byte, error) {
	kind, size, err := s.Kind()
	if err != nil {
		return nil, err
	}
	switch kind {
	case Byte:
		s.kind = -1
		return []byte{s.byteVal}, nil
	case String:
		bz := make([]byte, size)
		if err = s.readFull(bz); err != nil {
			return nil, err
		}
		if size == 1 && bz[0] < 0x80 {
			return nil, ErrCanonSize
		}
		return bz, nil
	default:
		return nil, ErrExpectedString
	}
}

// ReadBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// ReadBytes 方法接受一个字节切片bz，从底层stream解码出相应长度的字符串，非列表数据。
func (s *Stream) ReadBytes(bz []byte) error {
	kind, size, err := s.Kind()
	if err != nil {
		return err
	}
	switch kind {
	case Byte:
		if len(bz) != 1 {
			return fmt.Errorf("input value has wrong size 1, want %d", len(bz))
		}
		bz[0] = s.byteVal
		s.kind = -1
		return nil
	case String:
		if uint64(len(bz)) != size {
			return fmt.Errorf("input value has wrong size %d, want %d", size, len(bz))
		}
		if err = s.readFull(bz); err != nil {
			return err
		}
		if size == 1 && bz[0] < 0x80 {
			return ErrCanonSize
		}
		return nil
	default:
		return ErrExpectedString
	}
}

// Raw ♏ |作者：吴翔宇| 🍁 |日期：2022/11/11|
//
// Raw 方法返回stream里存储的 RawValue 数据。
func (s *Stream) Raw() ([]byte, error) {
	// 获取下一段数据的类型，size反映出stream里接下来存储的RawValue的大小
	kind, size, err := s.Kind()
	if err != nil {
		return nil, err
	}
	if kind == Byte {
		// 将kind设置为-1的目的是为了避免将来调用Kind()方法返回的kind还是之前编码数据片段的kind
		s.kind = -1
		return []byte{s.byteVal}, nil
	}
	// 计算编码前缀的的大小
	prefixSize := headSize(size)
	buf := make([]byte, uint64(prefixSize)+size)
	if err = s.readFull(buf[prefixSize:]); err != nil {
		return nil, err
	}
	if kind == String {
		putHead(buf, 0x80, 0xB7, size)
	} else {
		putHead(buf, 0xC0, 0xF7, size)
	}
	return buf, nil
}

// Uint64 ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// Uint64 方法从底层stream解码出一个64位无符号整数。
func (s *Stream) Uint64() (uint64, error) {
	return s.uint(64)
}

// bool ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// bool 方法解码底层stream接下来的数据成bool类型。
func (s *Stream) bool() (bool, error) {
	num, err := s.uint(8)
	if err != nil {
		return false, err
	}
	switch num {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("rlp: invalid boolean value: %d", num)
	}
}

// uint ♏ |作者：吴翔宇| 🍁 |日期：2022/11/10|
//
// uint 方法接受一个整数maxBits，该方法从底层stream里读取一个整数，该整数占用的比特数必须不大于maxBits，否则报错。
// 首先 uint 方法会调用 Kind 方法从底层数据池里读取
func (s *Stream) uint(maxBits int) (uint64, error) {
	kind, size, err := s.Kind()
	if err != nil {
		return 0, err
	}
	switch kind {
	case Byte:
		if s.byteVal == 0 {
			return 0, ErrCanonInt
		}
		s.kind = -1
		return uint64(s.byteVal), nil
	case String:
		// 是一个大于127的整数，或者是0
		if size > uint64(maxBits/8) {
			return 0, errUintOverflow
		}
		v, err := s.readUint(byte(size))
		switch {
		case err == ErrCanonSize:
			return 0, ErrCanonInt
		case err != nil:
			return 0, err
		case size > 0 && v < 128:
			return 0, ErrCanonSize
		default:
			return v, nil
		}
	default:
		return 0, ErrExpectedString
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义 Kind 类型，Kind 类型标示出了编码数据时所面临的不同规则。

type Kind int8

const (
	Byte Kind = iota
	String
	List
)

func (k Kind) String() string {
	switch k {
	case Byte:
		return "Byte"
	case String:
		return "String"
	case List:
		return "List"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}
