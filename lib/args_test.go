package moon

import (
	"testing"
)

func TestArgs(t *testing.T) {
	var one struct {
		Host     string `name: host; short: h; default: localhost`
		Port     int    `name: port; short: p; required: true`
		UseSSL   bool   `name: ssl_enabled; long: ssl-enabled`
		CertPath string `name: ssl_cert; long: ssl-cert`
	}
	args := []string{"program", "--host=example.com", "--port", "9000"}
	vals, err := parseArgs(args, &one)
	if err != nil {
		t.Error(err)
		return
	}

	if vals["host"] != "example.com" {
		t.Errorf("expected host 'example.com', saw host '%s'", vals["host"])
	}

	if vals["port"] != 9000 {
		t.Errorf("expected port 9000, saw port %d", vals["port"])
	}
}
