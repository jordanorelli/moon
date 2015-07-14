package moon

import (
	"fmt"
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
		return
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

func ExampleDoc_Fill() {
	input := `
    name: jordan
    age: 29
    `

	var config struct {
		Name string `name: "name"`
		Age  int    `name: "age"`
		City string `name: "city" default: "Brooklyn"`
	}

	doc, err := ReadString(input)
	if err != nil {
		fmt.Printf("error reading input: %s", err)
		return
	}

	if err := doc.Fill(&config); err != nil {
		fmt.Printf("error filling config value: %s", err)
		return
	}

	fmt.Println(config)
	// Output: {jordan 29 Brooklyn}
}

func ExampleDoc_Get_one() {
	input := `
    name: jordan
    age: 29
    `

	doc, err := ReadString(input)
	if err != nil {
		fmt.Printf("error reading input: %s", err)
		return
	}

	var name string
	if err := doc.Get("name", &name); err != nil {
		fmt.Printf("error filling config value: %s", err)
		return
	}

	fmt.Println(name)
	// Output: jordan
}

func ExampleDoc_Get_two() {
	input := `
    @todd: {
        name: todd
        age: 38
    }

    @sean: {
        name: sean
        age: 34
    }

    @jordan: {
        name: jordan
        age: 29
    }
    brothers: [@todd @sean @jordan]
    `

	doc, err := ReadString(input)
	if err != nil {
		fmt.Printf("error reading input: %s", err)
		return
	}

	var name string
	if err := doc.Get("brothers/1/name", &name); err != nil {
		fmt.Printf("error filling config value: %s", err)
		return
	}

	fmt.Println(name)
	// Output: sean
}

func TestFillEmbeds(t *testing.T) {
	in := `top: {val: some_data}`

	var dest struct {
		Top *struct {
			Val string `name: val`
		} `name: top`
	}

	doc, err := ReadString(in)
	if err != nil {
		t.Error(err)
		return
	}

	if err := doc.Fill(&dest); err != nil {
		t.Error(err)
		return
	}

	if dest.Top == nil {
		t.Error("didn't actually set a value")
		return
	}

	if dest.Top.Val != "some_data" {
		t.Errorf("expected some_data, got %v", dest.Top.Val)
	}
}
