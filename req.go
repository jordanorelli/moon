package moon

import (
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"
)

type req struct {
	name     string      // name as it appears in moon config file.  Defaults to the field name.
	help     string      // text given in help documentation
	required bool        // whether or not the option must be configured
	d_fault  interface{} // default value for when the option is missing
	short    string      // short flag on the command line
	long     string      // long flag on the command line
	t        reflect.Type
}

func (r req) validate() error {
	if r.name == "" {
		return fmt.Errorf("invalid requirement: requirement must have a name")
	}

	if r.t.Kind() == reflect.Bool && r.required {
		return fmt.Errorf("invalid requirement: a boolean cannot be required")
	}

	if r.required && r.d_fault != nil {
		return fmt.Errorf("invalid requirement %s: a required value cannot have a default", r.name)
	}

	if utf8.RuneCountInString(r.short) > 1 {
		return fmt.Errorf("invalid requirement %s: provided short flag (%s) is more than 1 rune",
			r.name, r.short)
	}
	return nil
}

func (r req) writeHelpLine(w io.Writer) {
	if r.short != "" {
		fmt.Fprintf(w, "-%s\t%s\n\n", r.short, r.name)
		fmt.Fprintf(w, "\t%s\n\n", r.help)
	} else if r.long != "" {
		fmt.Fprintf(w, "--%s\t%s\n\n", r.long, r.name)
		fmt.Fprintf(w, "\t%s\n\n", r.help)
	}
}

func field2req(field reflect.StructField) (*req, error) {
	doc, err := ReadString(string(field.Tag))
	if err != nil {
		return nil, fmt.Errorf("unable to parse requirements for field %s: %s", field.Name, err)
	}

	req := req{
		name: field.Name,
		t:    field.Type,
		long: field.Name,
	}

	// this is called by Fill, so we have to do Fill's work by hand, otherwise
	// they would be mutually recursive.

	errors := map[string]error{
		"name":     doc.Get("name", &req.name),
		"help":     doc.Get("help", &req.help),
		"required": doc.Get("required", &req.required),
		"default":  doc.Get("default", &req.d_fault),
		"short":    doc.Get("short", &req.short),
		"long":     doc.Get("long", &req.long),
	}

	if req.long == field.Name && req.name != field.Name {
		req.long = req.name
	}

	for fname, err := range errors {
		if err == nil {
			continue
		}
		if _, ok := err.(NoValue); !ok {
			return nil, fmt.Errorf("unable to parse requirement %s: %s", fname, err)
		}
	}

	return &req, nil
}

// requirements gathers the moon requirements for a given struct type. the
// output is a mapping of field names to requirements.
func requirements(t reflect.Type) (map[string]req, error) {
	switch t.Kind() {
	case reflect.Ptr:
		t = t.Elem()
	case reflect.Struct:
	default:
		return nil, fmt.Errorf("destination is of type %v; a pointer or struct type is required", t)
	}

	n := t.NumField()
	out := make(map[string]req, t.NumField())

	for i := 0; i < n; i++ {
		field := t.Field(i)
		req, err := field2req(field)
		if err != nil {
			return nil, fmt.Errorf("unable to gather requirements for field %s: %s", field.Name, err)
		}
		out[field.Name] = *req
	}
	return out, nil
}
