package moon

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Doc struct {
	items map[string]interface{}
}

func (d *Doc) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.items)
}

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
		return nil, fmt.Errorf("unable to seek value at path %s: no value found for part %s", fullpath, head)
	}

	if len(tail) == 0 {
		return v, nil
	}
	return seekValue(fullpath, tail, v)
}
