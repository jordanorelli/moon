package moon

import (
	"testing"
)

func TestArgs(t *testing.T) {
	var one struct {
		Host     string `name: host; short: h; default: localhost`
		Port     int    `name: port; short: p; required: true`
		User     string `name: user; short: u`
		ZipIt    bool   `name: zip; short: z`
		Verbose  bool   `name: verbose; short: v`
		UseSSL   bool   `name: ssl_enabled; long: ssl-enabled`
		CertPath string `name: ssl_cert; long: ssl-cert`
	}
	args := []string{"program", "--host=example.com", "--port", "9000", "-u", "fart", "-zv"}
	vals, err := parseArgs(args, &one)
	if err != nil {
		t.Error(err)
		return
	}

	if vals.items["host"] != "example.com" {
		t.Errorf("expected host 'example.com', saw host '%s'", vals.items["host"])
	}

	if vals.items["port"] != 9000 {
		t.Errorf("expected port 9000, saw port %d", vals.items["port"])
	}

	if vals.items["user"] != "fart" {
		t.Errorf("expected user 'fart', saw user '%s'", vals.items["user"])
	}

	if vals.items["zip"] != true {
		t.Errorf("expected zip to be true, is false")
	}

	if vals.items["verbose"] != true {
		t.Errorf("expected verbose to be true, is false")
	}
}
