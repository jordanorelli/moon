/*
The Moon configuration language.

Purpose

The Moon configuration language is intended to be an alternative to json as
a configuration language for Go projects.  Moon has the following explicit
design goals:

  - to be reasonable for a human to write
  - to be reasonable for a human to read
  - to be reasonable for a machine to generate
  - to be reasonable for a programmer to parse
  - to accomodate documents both large and small

That is, none of these goals is heralded as being the single most important
goal, and Moon makes no claim at being the best format for any of these
individual goals, but it does attempt to consider each of them to at least
some degree.
*/
package moon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Object is a representation of a Moon object in its native form.  It has no
// configured options and deals only with opaque types.
type Object struct {
	items map[string]interface{}
}

func (o *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.items)
}

func (o *Object) MarshalMoon() ([]byte, error) {
	var buf bytes.Buffer
	for k, v := range o.items {
		buf.WriteString(k)
		buf.WriteByte(':')
		buf.WriteByte(' ')
		b, err := Encode(v)
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(b); err != nil {
			return nil, err
		}
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

// Get reads a value from the Moon object at a given path, assigning the
// value to supplied destination pointer. The argument dest MUST be a pointer,
// otherwise Get will be unable to overwrite the value that it (should) point
// to.
//
// path may represent either a top-level key or a path to a value found within the object.
//
// Let's say you have a simple document that contains just a few values:
//
//   name: jordan
//   city: brooklyn
//
// Calling Get("name", &name) would read the value in the document at the
// "name" key and assign it to the location pointed at by the *string name.
//
// Let's take a more complex example:
//
//   @webserver: {
//       host: www.example.com
//       port: 80
//   }
//
//   @mailserver: {
//       host: mail.example.com
//       port: 25
//   }
//
//   servers: [@webserver @mailserver]
//
// Calling Get("servers/1/host", &host) would look a list at the key "servers",
// retrieve the item at index 1 in that list, and then read the field named
// "host" within that item, assigning the value to the address pointed to by
// the *string named host
func (o *Object) Get(path string, dest interface{}) error {
	if o.items == nil {
		return fmt.Errorf("no item found at path %s (object is empty)", path)
	}

	var v interface{}
	parts := strings.Split(path, "/")

	v, err := seekValue(path, parts, o)
	if err != nil {
		return err
	}

	dt := reflect.TypeOf(dest)
	if dt.Kind() != reflect.Ptr {
		return fmt.Errorf("destination is of type %v; a pointer type is required", dt)
	}

	dv := reflect.ValueOf(dest)
	dve := dv.Elem()

	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		dve.Set(reflect.ValueOf(v).Elem())
	} else {
		dve.Set(reflect.ValueOf(v))
	}
	return nil
}

// Fill takes the raw values from the Moon object and assigns them to the
// fields of the struct pointed at by dest. Dest must be a struct pointer; any
// other type for dest will result in an error. Please see the Parse
// documentation for a description of how the values will be filled.
func (o *Object) Fill(dest interface{}) error {
	// dt = destination type
	dt := reflect.TypeOf(dest)
	if dt.Kind() != reflect.Ptr {
		return fmt.Errorf("destination is of type %v (%v); a pointer type is required", dt, dt.Kind())
	}

	// ensure the pointer points to a struct type
	if dt.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("destination is a pointer to a non-struct type: %v (pointer to struct required)", dt.Elem())
	}

	// value of our struct pointer
	pv := reflect.ValueOf(dest)

	// value of the struct being pointed to
	v := pv.Elem()

	return o.fillValue(v)
}

func (o *Object) fillValue(dv reflect.Value) error {
	switch dv.Kind() {
	case reflect.Struct:
		// this is fine
	case reflect.Ptr:
		if dv.IsNil() {
			dv.Set(reflect.New(dv.Type().Elem()))
		}
		dv = dv.Elem()
	default:
		if reflect.TypeOf(o.items).AssignableTo(dv.Type()) {
			dv.Set(reflect.ValueOf(o.items))
			return nil
		}
		return fmt.Errorf("moon object can only fillValue to a struct value, saw %v (%v)", dv.Type(), dv.Kind())
	}

	// the destination defines the requirements (i.e., the method of unpacking
	// our moon data)
	reqs, err := requirements(dv.Type())
	if err != nil {
		return fmt.Errorf("unable to gather requirements: %v", err)
	}

	for fname, req := range reqs {
		// field value
		fv := dv.FieldByName(fname)
		// object value
		ov, ok := o.items[req.name]
		if !ok {
			// moon data is missing expected field
			if req.required {
				// if the field is required, that's an error
				return fmt.Errorf("required field missing: %s", fname)
			}
			if req.d_fault != nil {
				// otherwise, we look for a user-defined default value
				fv.Set(reflect.ValueOf(req.d_fault))
			}
			continue
		}

		switch t_ov := ov.(type) {
		case *Object:
			if err := t_ov.fillValue(fv); err != nil {
				return err
			}
		case List:
			if err := t_ov.fillValue(fv); err != nil {
				return err
			}
		default:
			if !fv.Type().AssignableTo(reflect.TypeOf(ov)) {
				return fmt.Errorf("unable to assign field %s: source type %v is not assignable to destination type %v", req.name, reflect.TypeOf(ov), fv.Type())
			}
			fv.Set(reflect.ValueOf(ov))
		}
	}
	return nil
}

// NoValue is the error type returned when attempting to get a value from a
// moon doc that isn't found.
type NoValue struct {
	fullpath string
	relpath  string
}

func (n NoValue) Error() string {
	return fmt.Sprintf("no value found for path %s", n.relpath)
}

func seekValue(fullpath string, parts []string, root interface{}) (interface{}, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("path is empty")
	}

	head, tail := parts[0], parts[1:]
	n, err := strconv.Atoi(head)
	if err == nil {
		l, ok := root.(List)
		if !ok {
			return nil, fmt.Errorf("can only index a List, root is %s", reflect.TypeOf(root))
		}
		if n >= len(l) {
			return nil, fmt.Errorf("path %s is out of bounds, can't get the %d index from a slice of len %d", fullpath, n, len(l))
		}
		v := l[n]
		if len(tail) == 0 {
			return v, nil
		}
		return seekValue(fullpath, tail, v)
	}

	m, ok := root.(*Object)
	if !ok {
		return nil, fmt.Errorf("can only key an Object, root is %v", reflect.TypeOf(root))
	}

	v, ok := m.items[head]
	if !ok {
		return nil, NoValue{fullpath, head}
	}

	if len(tail) == 0 {
		return v, nil
	}
	return seekValue(fullpath, tail, v)
}
