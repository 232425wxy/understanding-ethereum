//go:build !nacl && !js && cgo
// +build !nacl,!js,cgo

package rlp

import (
	"reflect"
	"unsafe"
)

// byteArrayBytes ♏ |作者：吴翔宇| 🍁 |日期：2022/11/9|
//
// byteArrayBytes 将字节数组转换为字节切片
func byteArrayBytes(val reflect.Value, length int) []byte {
	var res []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&res))
	sh.Data = val.UnsafeAddr()
	sh.Cap = length
	sh.Len = length
	return res
}
