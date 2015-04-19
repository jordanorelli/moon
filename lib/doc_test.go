package moon

import (
	"testing"
)

func TestDoc(t *testing.T) {
	doc, err := ReadString(`
    name: "jordan"
    `)
	if err != nil {
		t.Error(err)
		return
	}
	var name string
	if err := doc.Get("name", &name); err != nil {
		t.Error(err)
		return
	}
	if name != "jordan" {
		t.Errorf("unexpected name value, expected 'jordan', saw '%s'", name)
	}
}
