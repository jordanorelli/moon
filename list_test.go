package moon

import (
	"reflect"
	"testing"
)

func TestFillList(t *testing.T) {
	doc, err := ReadString(`
    # yay for lists
    @cool_names: [larry; curly; moe]
    servers: [
        {hostname: dev.example.com; label: dev}
        {hostname: prod.example.com; label: prod}
    ]
    numbers: [1 1 2 3 5 8 13]
    names: @cool_names
    `)
	if err != nil {
		t.Error(err)
		return
	}
	type server struct {
		Hostname string `name: hostname; required: true`
		Label    string `name: label; required: true`
	}
	var config struct {
		Servers []server `name: servers`
		Numbers []int    `name: numbers`
		Names   []string `name: names`
	}
	if err := doc.Fill(&config); err != nil {
		t.Error(err)
		return
	}
	servers := []server{
		{"dev.example.com", "dev"},
		{"prod.example.com", "prod"},
	}
	if !reflect.DeepEqual(config.Servers, servers) {
		t.Errorf("bad servers: %v", config.Servers)
	}
	if !reflect.DeepEqual(config.Numbers, []int{1, 1, 2, 3, 5, 8, 13}) {
		t.Errorf("bad numbers: %v", config.Numbers)
	}
	if !reflect.DeepEqual(config.Names, []string{"larry", "curly", "moe"}) {
		t.Errorf("bad names: %v", config.Names)
	}
}
