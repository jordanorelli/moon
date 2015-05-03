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

// retrieve a value from the moon document and assign it to provided
// destination.  dest must be a pointer.  path may be a path to the desired
// field, e.g., people/1/name would get the name field of the second element of
// the people list.
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
	return setValue(dve, reflect.ValueOf(v))
}

func setValue(dest, src reflect.Value) error {
	if dest.Type() == src.Type() {
		dest.Set(src)
		return nil
	}
	switch dest.Kind() {
	case reflect.Bool:
		setBoolValue(dest, src)
		return nil
	default:
		return fmt.Errorf("src and destination are not of same type. src type: %s, dest type: %s", src.Type().Name(), dest.Type().Name())
	}
}

func setBoolValue(dest, src reflect.Value) error {
	switch src.Kind() {
	case reflect.String:
		s := src.String()
		if s == "false" { // lol
			dest.Set(reflect.ValueOf(false))
			return nil
		}
		dest.Set(reflect.ValueOf(true))
		return nil
	default:
		return fmt.Errorf("dunno how to set a bool with that damned value")
	}
}

// Fill takes the raw values from the moon document and assigns them to the
// struct pointed at by dest.  Dest must be a struct pointer; any other type
// for dest will result in an error.
func (d *Doc) Fill(dest interface{}) error {
	dt := reflect.TypeOf(dest)
	if dt.Kind() != reflect.Ptr {
		return fmt.Errorf("destination is of type %v; a pointer type is required", dt)
	}

	dv := reflect.ValueOf(dest).Elem()

	n := dv.NumField()
	for i := 0; i < n; i++ {
		field, fv := dv.Type().Field(i), dv.Field(i)

		tag := field.Tag
		fopts, err := ReadString(string(tag))
		if err != nil {
			return fmt.Errorf("unable to parse options for type %s, field %s: %s",
				dv.Type().Name(), field.Name, err)
		}

		var fname string
		if err := fopts.Get("name", &fname); err != nil {
			if _, ok := err.(NoValue); !ok {
				return err
			}
		}

		var frequired bool
		if err := fopts.Get("required", &frequired); err != nil {
			if _, ok := err.(NoValue); !ok {
				return err
			}
		}

		fdefault := fopts.items["default"]

		v, ok := d.items[fname]
		if !ok {
			if frequired {
				return fmt.Errorf("required field missing: %s", fname)
			} else {
				fv.Set(reflect.ValueOf(fdefault))
				continue
			}
		}
		fv.Set(reflect.ValueOf(v))
	}
	return nil
}

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
