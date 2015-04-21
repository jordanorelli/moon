package moon

import (
	"bytes"
	"math"
	"reflect"
	"strconv"
)

func Encode(v interface{}) ([]byte, error) {
	e := &encoder{}
	if err := e.encode(v); err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

type encoder struct {
	bytes.Buffer
	scratch [64]byte
}

func (e *encoder) encode(v interface{}) error {
	e.encodeValue(reflect.ValueOf(v))
	return nil
}

func (e *encoder) encodeValue(v reflect.Value) {
	fn := valueEncoder(v)
	fn(e, v)
}

func valueEncoder(v reflect.Value) encodeFn {
	if !v.IsValid() {
		return encodeNull
	}
	return typeEncoder(v.Type())
}

func typeEncoder(t reflect.Type) encodeFn {
	switch t.Kind() {
	case reflect.Bool:
		return encodeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return encodeInt
	case reflect.Float32:
		return encodeFloat32
	case reflect.Float64:
		return encodeFloat64
	case reflect.String:
		return encodeString
	default:
		panic("I don't know what to do here")
	}
}

type encodeFn func(e *encoder, v reflect.Value)

func encodeBool(e *encoder, v reflect.Value) {
	if v.Bool() {
		e.WriteString("true")
	} else {
		e.WriteString("false")
	}
}

func encodeInt(e *encoder, v reflect.Value) {
	b := strconv.AppendInt(e.scratch[:0], v.Int(), 10)
	e.Write(b)
}

func encodeNull(e *encoder, v reflect.Value) {
	e.WriteString("null")
}

func encodeFloat(bits int) encodeFn {
	return func(e *encoder, v reflect.Value) {
		f := v.Float()
		if math.IsInf(f, 0) || math.IsNaN(f) {
			panic("that value is bad") // TODO: not this
		}
		b := strconv.AppendFloat(e.scratch[:0], f, 'f', 1, bits)
		e.Write(b)
	}
}

var (
	encodeFloat32 = encodeFloat(32)
	encodeFloat64 = encodeFloat(64)
)

// this is a really over-simplified string emitter.  I highly doubt it can stay
// like this.
func encodeString(e *encoder, v reflect.Value) {
	s := v.String()
	e.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			e.WriteByte('\\')
			e.WriteByte('\\')
		case '"':
			e.WriteByte('\\')
			e.WriteByte('"')
		default:
			e.WriteRune(r)
		}
	}
	e.WriteByte('"')
}
