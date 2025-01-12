package gvariant

import (
	"encoding/binary"
	"math"
	"reflect"
)

// InvalidUnmarshalError indicates that the reciever is inappropriate
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "gvariant: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "gvariant: Unmarshal(non-pointer " + e.Type.String() + ")"
	}

	return "gvariant: Unmarshal(nil " + e.Type.String() + ")"
}

// Unmarshal decodes the byte stream `bytes` into the referenced value `v`.
func Unmarshal(data []byte, v any, byteOrder binary.ByteOrder) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidUnmarshalError{rv.Type()}
	}

	state := newDecodeState(data, rv, byteOrder)

	return state.decode()
}

type decodeState struct {
	data            []byte
	receiver        reflect.Value
	frameOffsetSize int
	byteOrder       binary.ByteOrder
}

func newDecodeState(data []byte, rv reflect.Value, byteOrder binary.ByteOrder) *decodeState {
	ret := decodeState{
		data:            data,
		receiver:        rv,
		frameOffsetSize: frameOffsetSizeForContainerSize(len(data)),
		byteOrder:       byteOrder,
	}

	return &ret
}

func (d *decodeState) decode() error {
	if len(d.data) == 0 {
		return nil
	}

	switch d.receiver.Elem().Kind() {
	case reflect.Int8, reflect.Uint8:
		return d.decodeByte()
	case reflect.Int16, reflect.Uint16:
		return d.decodeInt16()
	case reflect.Int32, reflect.Uint32:
		return d.decodeInt32()
	case reflect.Int64, reflect.Uint64:
		return d.decodeInt64()
	case reflect.Float32, reflect.Float64:
		return d.decodeFloat()
	case reflect.Bool:
		return d.decodeBool()
	case reflect.String:
		return d.decodeString()
	case reflect.Array, reflect.Slice:
		return d.decodeArray()
	case reflect.Map:
		return d.decodeMap()
	case reflect.Struct:
		return d.decodeStruct()
	}

	return nil
}

func (d *decodeState) decodeByte() error {
	val := int8(d.data[0])

	switch d.receiver.Elem().Kind() {
	case reflect.Int8:
		d.receiver.Elem().SetInt(int64(val))
	case reflect.Uint8:
		d.receiver.Elem().SetUint(uint64(val))
	default:
		return &InvalidUnmarshalError{d.receiver.Elem().Type()}
	}
	return nil
}

func (d *decodeState) decodeInt16() error {
	val := d.byteOrder.Uint16(d.data)

	switch d.receiver.Elem().Kind() {
	case reflect.Int16:
		d.receiver.Elem().SetInt(int64(val))
	case reflect.Uint16:
		d.receiver.Elem().SetUint(uint64(val))
	default:
		return &InvalidUnmarshalError{d.receiver.Elem().Type()}
	}
	return nil
}

func (d *decodeState) decodeInt32() error {
	val := d.byteOrder.Uint32(d.data)

	switch d.receiver.Elem().Kind() {
	case reflect.Int32:
		d.receiver.Elem().SetInt(int64(val))
	case reflect.Uint32:
		d.receiver.Elem().SetUint(uint64(val))
	default:
		return &InvalidUnmarshalError{d.receiver.Elem().Type()}
	}
	return nil
}

func (d *decodeState) decodeInt64() error {
	val := d.byteOrder.Uint64(d.data)

	switch d.receiver.Elem().Kind() {
	case reflect.Int64:
		d.receiver.Elem().SetInt(int64(val))
	case reflect.Uint64:
		d.receiver.Elem().SetUint(uint64(val))
	default:
		return &InvalidUnmarshalError{d.receiver.Elem().Type()}
	}
	return nil
}

func (d *decodeState) decodeFloat() error {
	val := d.byteOrder.Uint64(d.data)

	d.receiver.Elem().SetFloat(math.Float64frombits(val))
	return nil
}

func (d *decodeState) decodeBool() error {
	val := int8(d.data[0])
	d.receiver.Elem().SetBool(val == 1)

	return nil
}

func (d *decodeState) decodeString() error {
	val := string(d.data[:len(d.data)-1]) // drop the trailing \0
	d.receiver.Elem().SetString(val)
	return nil
}

func (d *decodeState) decodeArray() error {
	val := d.receiver.Elem()
	innerType := val.Type().Elem()
	innerVal := reflect.New(innerType).Elem()

	if isFixedWidth(innerVal) {
		// we can just itterate the array in len(innerkink) bytes
		width := typeWidth(innerVal)
		currentStart := 0
		for {
			if currentStart >= len(d.data) {
				break
			}

			nextStart := currentStart + width
			rv := reflect.New(innerVal.Type())

			err := newDecodeState(d.data[currentStart:nextStart], rv, d.byteOrder).decode()
			if err != nil {
				return err
			}

			val.Set(reflect.Append(val, rv.Elem()))
			currentStart = nextStart
		}
		return nil
	}

	alignment := typeAlignment(reflect.New(innerType).Elem())
	frameOffsetStart := bytesToInt(d.data[len(d.data)-d.frameOffsetSize:], d.byteOrder)
	frameOffsets := d.data[frameOffsetStart:]
	frameCount := len(frameOffsets) / d.frameOffsetSize

	var offset uint64 = 0
	for frame := 0; frame < frameCount; frame++ {
		frameBytes := unshift(&frameOffsets, d.frameOffsetSize)
		endPosition := bytesToInt(frameBytes, d.byteOrder)
		frameData := d.data[offset:endPosition]

		rv := reflect.New(val.Type().Elem())
		err := newDecodeState(frameData, rv, d.byteOrder).decode()
		if err != nil {
			return err
		}
		val.Set(reflect.Append(val, rv.Elem()))
		offset = uint64(nextAlignment(int(endPosition), alignment))
	}

	return nil
}

func (d *decodeState) decodeStruct() error {
	val := d.receiver.Elem()

	if structIsVariant(val) {
		return d.decodeVariant()
	}

	fieldCount := val.NumField()

	// iter the fields, if a field type is fixed size, process it with that many bytes.
	// if its variable width, grap the end from the frame offsets.  If its the last field,
	// just pass the remaining bytes.
	offset := 0
	frameBoundsConsumed := 0

	for i := 0; i < fieldCount; i++ {
		field := val.Field(i)

		// TODO: pointer values represent maybe types in the gvariant spec.
		// NOTHING values will be represented by a nil pointer assigned, otherwise
		// a pointer to a real value will be created and assigned
		//
		// if field.Kind() == reflect.Pointer {
		// 	// this is a maybe field...
		// 	if field.IsNil() {
		// 		// assign a value
		// 	}
		// 	field = field.Elem()
		// }

		rv := reflect.New(field.Type())
		offset = nextAlignment(offset, typeAlignment(rv.Elem()))
		endPosition := offset

		if isFixedWidth(rv.Elem()) {
			// consume the bytes
			endPosition = offset + typeWidth(rv.Elem())
			err := newDecodeState(d.data[offset:endPosition], rv, d.byteOrder).decode()
			if err != nil {
				return err
			}
		}

		if !isFixedWidth(rv.Elem()) {
			if i == fieldCount-1 {
				//last field
				endPosition = len(d.data) - (d.frameOffsetSize * (frameBoundsConsumed))
				err := newDecodeState(d.data[offset:endPosition], rv, d.byteOrder).decode()
				if err != nil {
					return err
				}
			} else {
				boundBytesStart := len(d.data) - (d.frameOffsetSize * (frameBoundsConsumed + 1))
				boundBytes := d.data[boundBytesStart:]
				boundBytes = boundBytes[0:d.frameOffsetSize]

				endPosition = int(bytesToInt(boundBytes, d.byteOrder))
				err := newDecodeState(d.data[offset:endPosition], rv, d.byteOrder).decode()
				if err != nil {
					return err
				}
				frameBoundsConsumed++
			}
		}

		offset = endPosition
		field.Set(rv.Elem())
	}

	return nil
}

func (d *decodeState) decodeMap() error {
	mapVal := d.receiver.Elem()

	if mapVal.IsNil() {
		mapVal.Set(reflect.MakeMap(mapVal.Type()))
	}

	keyType := mapVal.Type().Key()
	valType := mapVal.Type().Elem()
	keyValue := reflect.New(keyType)
	valValue := reflect.New(valType)

	fixedWidthKey := isFixedWidth(keyValue.Elem())
	alignment := typeAlignment(mapVal)

	keyBytes := []byte{}
	valBytes := []byte{}

	if fixedWidthKey {
		// No offsets
		keyBytes = d.data[:typeWidth(keyValue)]
		valBytes = d.data[nextAlignment(typeWidth(keyValue), alignment):]
	}

	if !fixedWidthKey {
		keyBound := bytesToInt(d.data[len(d.data)-d.frameOffsetSize:], d.byteOrder)
		keyBytes = d.data[:keyBound]
		valBytes = d.data[nextAlignment(int(keyBound), alignment) : len(d.data)-d.frameOffsetSize]
	}

	err := newDecodeState(keyBytes, keyValue, d.byteOrder).decode()
	if err != nil {
		return err
	}
	err = newDecodeState(valBytes, valValue, d.byteOrder).decode()
	if err != nil {
		return err
	}

	mapVal.SetMapIndex(keyValue.Elem(), valValue.Elem())

	return nil
}

func (d *decodeState) decodeVariant() error {
	// still a bit hazy on how this works, but it seems to just be:
	// find the last \0, everything before is encoded bytes, everything
	// after is type String

	val := d.receiver.Elem()

	separatorPossition := 0
	for i := len(d.data) - 1; i >= 0; i-- {
		if d.data[i] == 0x00 {
			separatorPossition = i
			break
		}
	}

	val.FieldByName("Data").SetBytes(d.data[:separatorPossition])
	val.FieldByName("Format").SetString(string(d.data[separatorPossition+1:]))

	return nil
}
