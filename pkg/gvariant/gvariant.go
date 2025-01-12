// Package gvariant provides functions for serializing and deserializing a Go type
// to or from a byte stream using the rules described by the gvariant specification
// found at https://developer.gnome.org/documentation/specifications/gvariant-specification-1.0.html
//
// This process maps Go types to the basic and container types described in the specification.  Based
// on this mapping, it should be possible to craft go structs that represent any arbitrary gvariant
// type notation string.  For example:
//
//	aString := ""              // matches 's'
//	someStrings := []string{}  // matches 'as'
//	type AStruct struct {      // matches '(suay)'
//	  Field1 string
//	  Field2 Uint32
//	  Field3 []uint8
//	}
//	moreStructs := []AStruct{} // matches 'a(suay)'
//
//	The complete mapping of gvariant notions to Go types:
//
//	b : bool
//	y : int8/uint8
//	n : int16
//	q : uint16
//	i : int32
//	u : uint32
//	x : int64
//	t : uint64
//	d : float64/float32
//	s : string
//	o : TODO
//	g : TODO
//	v : [Variant]
//	m : TODO maybe type (will be as a `*type`)
//	a : `[]type`
//	( types ) : `type AStruct { <fields per "types" }`
//	{ base_type type } : `map[<base_type>]<type>`
//
// **NOTE:** Stuct fields must be exported and defined in the order they appear in the notation.
//
// **NOTE:** dictionaries ('{ base_tyoe type }') currently only decodes a single key/value pair.
// typically a notation will use 'a{ base_type type}' to indicate a map as intended with multiple key/value pairs
// the go type for this needs to be `[]map[<base_type]type` and will result in a slice of map with one key/value
// each.  This is for simplicity, but also to accomodate for the fact that the gvariant serialization of
// 'a{ base_type type}' means that key duplication is possible.
package gvariant

import (
	"encoding/binary"
	"math"
	"reflect"
)

// Variant is used to receive a variant type from the encoding
type Variant struct {
	Data   []byte
	Format string
}

func frameOffsetSizeForContainerSize(size int) int {
	size_f := float64(size)
	i := 0
	for {
		if size_f <= math.Pow(2, float64(8*i)) {
			return i
		}

		i *= 2
		if i == 0 {
			i += 1
		}
	}
}

func isFixedWidth(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Bool, reflect.Float32, reflect.Float64:
		return true
	// case reflect.Slice, reflect.Array:
	// 	return isFixedWidth(reflect.New(v.Type().Elem()).Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isFixedWidth(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

func typeWidth(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Int8, reflect.Uint8, reflect.Bool:
		return 1
	case reflect.Int16, reflect.Uint16:
		return 2
	case reflect.Int32, reflect.Uint32:
		return 4
	case reflect.Int64, reflect.Uint64, reflect.Float32, reflect.Float64:
		return 8
	case reflect.Struct:
		if !isFixedWidth(v) {
			return 0
		}
		alignment := typeAlignment(v)
		width := 0
		for i := 0; i < v.NumField(); i++ {
			fieldWidth := typeWidth(v.Field(i))
			width += nextAlignment(fieldWidth, alignment)
		}
		return width
	}

	return 0
}

func typeAlignment(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Int8, reflect.Uint8, reflect.Bool, reflect.String:
		return 1
	case reflect.Int16, reflect.Uint16:
		return 2
	case reflect.Int32, reflect.Uint32:
		return 4
	case reflect.Int64, reflect.Uint64, reflect.Float32, reflect.Float64:
		return 8
	case reflect.Slice, reflect.Array:
		return typeAlignment(reflect.New(v.Type().Elem()).Elem())
	case reflect.Struct:
		if structIsVariant(v) {
			return 8
		}
		structAlignment := 0
		for i := 0; i < v.NumField(); i++ {
			a := typeAlignment(v.Field(i))
			if a > structAlignment {
				structAlignment = a
			}
		}
		return structAlignment
	case reflect.Map:
		keyType := v.Type().Key()
		valType := v.Type().Elem()
		keyValue := reflect.New(keyType)
		valValue := reflect.New(valType)
		keyAlignment := typeAlignment(keyValue.Elem())
		valAlignment := typeAlignment(valValue.Elem())

		if keyAlignment > valAlignment {
			return keyAlignment
		}

		return valAlignment
	}

	return 0
}

func bytesToInt(offset []byte, byteOrder binary.ByteOrder) uint64 {
	if len(offset) == 1 {
		return uint64(uint8(offset[0]))
	}
	if len(offset) == 2 {
		return uint64(byteOrder.Uint16(offset))
	}
	if len(offset) == 4 {
		return uint64(byteOrder.Uint32(offset))
	}
	if len(offset) == 8 {
		return uint64(byteOrder.Uint64(offset))
	}
	return 0
}

func nextAlignment(pos, align int) int {
	if (pos % align) != 0 {
		return pos + (align - (pos % align))
	}
	return pos
}

func structIsVariant(v reflect.Value) bool {
	if v.Kind() != reflect.Struct {
		return false
	}

	if v.Type().Name() != "Variant" {
		return false
	}

	if v.Type().PkgPath() != "github.com/chrisportman/go-gvariant/gvariant" && v.Type().PkgPath() != "github.com/joonas-fi/joonas-sys/pkg/gvariant" {
		return false
	}

	return true
}

func unshift[T any](s *[]T, c int) []T {
	sliceDeref := *s
	r, newSlice := sliceDeref[:c], sliceDeref[c:]
	*s = newSlice
	return r
}
