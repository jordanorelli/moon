package moon

import (
	"testing"
)

func TestFillList(t *testing.T) {
	doc, err := ReadString(`
    # yay for lists
    servers: [
        {hostname: dev.example.com; label: dev}
        {hostname: prod.example.com; label: prod}
    ]
    `)
	if err != nil {
		t.Error(err)
		return
	}
	var config struct {
		Servers []struct {
			Hostname string `name: hostname; required: true`
			Label    string `name: label; required: true`
		} `name: servers`
	}
	if err := doc.Fill(&config); err != nil {
		t.Error(err)
		return
	}
	if config.Servers == nil {
		t.Error("servers is nil for some reason")
		return
	}
	if len(config.Servers) != 2 {
		t.Errorf("expected 2 servers, saw %d", len(config.Servers))
		return
	}
	if config.Servers[0].Hostname != "dev.example.com" {
		t.Errorf("wut lol %v", config.Servers[0])
	}
	if config.Servers[0].Label != "dev" {
		t.Errorf("wut lol %v", config.Servers[0])
	}
	if config.Servers[1].Hostname != "prod.example.com" {
		t.Errorf("wut 1 lol %v", config.Servers[1])
	}
	if config.Servers[1].Label != "prod" {
		t.Errorf("wut 2 lol %v", config.Servers[1])
	}
}
