package rlpstruct

import (
	"reflect"
	"testing"
)

func TestOptionalTail(t *testing.T) {
	allFields := []Field{
		{Name: "Header", Index: 0, Exported: true, Type: Type{Kind: reflect.String}, Tag: `rlp:"-"`},
		{Name: "Height", Index: 1, Exported: true, Type: Type{Kind: reflect.Uint64}, Tag: `rlp:"optional"`},
		{Name: "Txs", Index: 2, Exported: true, Type: Type{Kind: reflect.Ptr}, Tag: `rlp:"optional,nilList"`},
		// tag被设置为optional的字段后面的字段的tag可以不必设置为optional，而设置为tail
		{Name: "Validators", Index: 3, Exported: true, Type: Type{Kind: reflect.Slice}, Tag: `rlp:"tail"`},
	}
	fields, tags, err := ProcessFields(allFields)
	t.Log(fields)
	t.Log(tags)
	t.Log(err)
}
