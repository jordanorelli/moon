package moon

import (
	"reflect"
	"testing"
)

func TestFillList(t *testing.T) {
	doc, err := ReadString(`
    # yay for lists
    servers: [
        {hostname: dev.example.com; label: dev}
        {hostname: prod.example.com; label: prod}
    ]
    numbers: [1 1 2 3 5 8 13]
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
}
