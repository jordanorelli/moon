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

// Doc is a representation of a Moon document in its native form.  It has no
// configured options and deals only with opaque types.
type Doc struct {
	items map[string]interface{}
}

func (d *Doc) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.items)
}

func (d *Doc) MarshalMoon() ([]byte, error) {
	var buf bytes.Buffer
	for k, v := range d.items {
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

// Get reads a value from the Moon document at a given path, assigning the
// value to supplied destination pointer. The argument dest MUST be a pointer,
// otherwise Get will be unable to overwrite the value that it (should) point
// to.
//
// path may represent either a top-level key or a path to a value found within the document.
//
// Let's say you have a simple document that contains just a few values:
//   name: jordan
//   city: brooklyn
// Calling Get("name", &name) would read the value in the document at the
// "name" key and assign it to the location pointed at by the *string name.
//
// Let's take a more complex example:
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
func (d *Doc) Get(path string, dest interface{}) error {
	if d.items == nil {
		return fmt.Errorf("no item found at path %s (doc is empty)", path)
	}

	var v interface{}
	parts := strings.Split(path, "/")

	v, err := seekValue(path, parts, d.items)
	if err != nil {
		return err
	}

	dt := reflect.TypeOf(dest)
	if dt.Kind() != reflect.Ptr {
		return fmt.Errorf("destination is of type %v; a pointer type is required", dt)
	}

	dv := reflect.ValueOf(dest)
	dve := dv.Elem()
	dve.Set(reflect.ValueOf(v))
	return nil
}

// Fill takes the raw values from the Moon document and assigns them to the
// fields of the struct pointed at by dest. Dest must be a struct pointer; any
// other type for dest will result in an error. Please see the Parse
// documentation for a description of how the values will be filled.
func (d *Doc) Fill(dest interface{}) error {
	// dt = destination type
	dt := reflect.TypeOf(dest)
	if dt.Kind() != reflect.Ptr {
		return fmt.Errorf("destination is of type %v; a pointer type is required", dt)
	}

	reqs, err := requirements(dest)
	if err != nil {
		return fmt.Errorf("unable to gather requirements: %s", err)
	}

	// dv = destination value
	dv := reflect.ValueOf(dest).Elem()
	for fname, req := range reqs {
		// fv = field value
		fv := dv.FieldByName(fname)
		v, ok := d.items[req.name]
		if ok {
			if !fv.Type().AssignableTo(reflect.TypeOf(v)) {
				return fmt.Errorf("unable to assign field %s: source type %v is not assignable to destination type %v", req.name, fv.Type(), reflect.TypeOf(v))
			}
			fv.Set(reflect.ValueOf(v))
		} else {
			if req.required {
				return fmt.Errorf("required field missing: %s", fname)
			}
			if req.d_fault != nil {
				fv.Set(reflect.ValueOf(req.d_fault))
			}
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
		l, ok := root.([]interface{})
		if !ok {
			return nil, fmt.Errorf("can only index a []interface{}, root is %s", reflect.TypeOf(root))
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

	m, ok := root.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("can only key a map[string]interface{}, root is %v", reflect.TypeOf(root))
	}

	v, ok := m[head]
	if !ok {
		return nil, NoValue{fullpath, head}
	}

	if len(tail) == 0 {
		return v, nil
	}
	return seekValue(fullpath, tail, v)
}
