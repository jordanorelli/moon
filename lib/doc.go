package moon

import (
	"encoding/json"
	"fmt"
	"reflect"
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

	v, ok := d.items[path]
	if !ok {
		return fmt.Errorf("no item found at path %s", path)
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
