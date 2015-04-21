package moon

import (
	"testing"
)

var valueTests = []struct {
	in  interface{}
	out string
}{
	{nil, "null"},
	{true, "true"},
	{false, "false"},
	{0, "0"},
	{1, "1"},
	{12345, "12345"},
	{.1, "0.1"},
	{1.0, "1.0"},
	{1.0e9, "1000000000.0"},
	// this is kinda gross, but it's the only way I've figured out how to
	// prevent 1.0 printing out as 1 and thus having its type changed from
	// float to int. Sometimes having obnoxious string representations is
	// better than something having things change type.
	{"a string", `"a string"`},
	{`it's got "quotes"`, `"it's got \"quotes\""`},
}

func TestWriteValues(t *testing.T) {
	for _, test := range valueTests {
		out, err := Encode(test.in)
		if err != nil {
			t.Error(err)
			continue
		}
		if string(out) != test.out {
			t.Errorf("expected '%s', saw '%s'", test.out, string(out))
		}
	}
}
