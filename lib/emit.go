package moon

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strconv"
)

type Marshaler interface {
	MarshalMoon() ([]byte, error)
}

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

func (e *encoder) encode(v interface{}) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, ok := r.(runtime.Error); ok {
			panic(r)
		}
		if s, ok := r.(string); ok {
			panic(s)
		}
		err = r.(error)
	}()
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

var (
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
)

func typeEncoder(t reflect.Type) encodeFn {
	if t.Implements(marshalerType) {
		return marshalerEncoder
	}

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
	case reflect.Struct:
		return encodeStruct
	case reflect.Slice:
		return encodeSlice
	case reflect.Interface:
		return encodeInterface
	case reflect.Ptr:
		return encodePointer
	case reflect.Map:
		return encodeMap
	default:
		panic(fmt.Errorf("unhandled type: %v kind: %v", t, t.Kind()))
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
		b := strconv.AppendFloat(e.scratch[:0], f, 'g', -1, bits)
		e.Write(b)
	}
}

func encodeStruct(e *encoder, v reflect.Value) {
	t := v.Type()
	e.WriteByte('{')
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		e.WriteString(f.Name)
		e.WriteByte(':')
		e.WriteByte(' ')
		fv := v.FieldByName(f.Name)
		e.encodeValue(fv)
		if i != t.NumField()-1 {
			e.WriteByte(' ')
		}
	}
	e.WriteByte('}')
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

func encodeSlice(e *encoder, v reflect.Value) {
	e.WriteByte('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			e.WriteByte(' ')
		}
		e.encodeValue(v.Index(i))
	}
	e.WriteByte(']')
}

func encodeInterface(e *encoder, v reflect.Value) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	e.encodeValue(v.Elem())
}

func encodeMap(e *encoder, v reflect.Value) {
	t := v.Type()
	if t.Key().Kind() != reflect.String {
		panic(fmt.Errorf("unsupported map key type: %v", t.Key().Kind()))
	}
	keys := v.MapKeys()
	e.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			e.WriteByte(' ')
		}
		e.WriteString(key.String()) // TODO: escape this?
		e.WriteByte(':')
		e.WriteByte(' ')
		elem := v.MapIndex(key)
		e.encodeValue(elem)
	}
	e.WriteByte('}')
}

func encodePointer(e *encoder, v reflect.Value) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	e.encodeValue(v.Elem())
}

func marshalerEncoder(e *encoder, v reflect.Value) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		e.WriteString("null")
		return
	}
	m := v.Interface().(Marshaler)
	b, err := m.MarshalMoon()
	if err != nil {
		panic(err)
	}
	e.Write(b)
}
