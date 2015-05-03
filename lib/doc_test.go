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

func TestFill(t *testing.T) {
	doc, err := ReadString(`
    # this is just a comment
    firstname: jordan
    port: 9000
    `)
	if err != nil {
		t.Error(err)
		return
	}

	var dest struct {
		FirstName string `name: firstname; required: true`
		HostName  string `name: hostname; default: localhost`
		Port      int    `name: port; default: 9001`
	}

	if err := doc.Fill(&dest); err != nil {
		t.Error(err)
	}

	if dest.FirstName != "jordan" {
		t.Errorf("unexpected firstname value, expected 'jordan', saw '%s'", dest.FirstName)
	}
	if dest.HostName != "localhost" {
		t.Errorf("unexpected hostname value, expected 'localhost', saw '%s'", dest.HostName)
	}
	if dest.Port != 9000 {
		t.Errorf("unexpected port value, expected 9000, saw %d", dest.Port)
	}

}
