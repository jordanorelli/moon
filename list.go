package moon

import (
	"fmt"
	"reflect"
)

type List []interface{}

func (l List) fillValue(v reflect.Value) error {
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("moon List can only fillValue to a slice, saw %v (%v)", v.Type(), v.Kind())
	}
	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), len(l), cap(l)))
	}
	for idx, item := range l {
		dv := v.Index(idx)

		switch t_sv := item.(type) {
		case *Object:
			if err := t_sv.fillValue(dv); err != nil {
				return err
			}
		case List:
			if err := t_sv.fillValue(dv); err != nil {
				return err
			}
		default:
			sv := reflect.ValueOf(item)
			if !sv.Type().AssignableTo(dv.Type()) {
				return fmt.Errorf("unable to assign element %d: source type %v is not assignable to destination type %v", idx, sv.Type(), dv.Type())
			}
			dv.Set(sv)
		}
	}
	return nil
}
