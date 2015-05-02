package moon

import (
	"testing"
)

type person struct {
	Name string
	Age  int
}

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
	{1.0, "1"}, // god dammit
	{1.0e9, "1e+09"},
	{"a string", `"a string"`},
	{`it's got "quotes"`, `"it's got \"quotes\""`},
	{person{"jordan", 28}, `{Name: "jordan" Age: 28}`},
	{[]int{1, 2, 3}, `[1 2 3]`},
	{[]float32{1.0, 2.2, 3.3}, `[1 2.2 3.3]`},
	{[]float64{1.0, 2.2, 3.3}, `[1 2.2 3.3]`},
	{[]string{"one", "two", "three"}, `["one" "two" "three"]`},
	{
		map[string]int{"one": 1, "two": 2, "three": 3},
		`{one: 1 two: 2 three: 3}`,
	},
	{
		map[string]interface{}{
			"one": 1,
			"two": 2.0,
			"pi":  3.14,
		},
		`{one: 1 two: 2 pi: 3.14}`,
	},
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
