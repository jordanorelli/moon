package main

import (
	"testing"
)

var parseTests = []parseTest{
	{
		in:   ``,
		desc: "an empty string is a valid config",
		configTests: []configTest{
			{
				desc: "undefined name field should not exist",
				pass: inv(hasKey("name")),
			},
		},
	},
	{
		in:        `name `,
		desc:      "eof after name",
		errorType: e_unexpected_eof,
	},
	{
		in:        `firstname lastname`,
		desc:      "two names in a row",
		errorType: e_unexpected_token,
	},
	{
		in:        `name = `,
		desc:      "eof after equals",
		errorType: e_unexpected_eof,
	},
	{
		in:   `name = "jordan"`,
		desc: "assign a value",
		configTests: []configTest{
			{
				desc: "should have name",
				pass: hasValue("name", "jordan"),
			},
		},
	},
}

// a boolean statement about a config struct
type configPredicate func(*Config) bool

// a suite of tests for parsing potm input
type parseTest struct {
	in          string
	desc        string
	configTests []configTest
	errorType   parseErrorType
}

func (p *parseTest) run(t *testing.T) {
	c, err := parseString(p.in)
	if err != nil {
		t.Logf("test %s has error %v", p.desc, err)
		e, ok := err.(parseError)
		if !ok {
			t.Errorf("unexpected error: %s", e)
			return
		}
		if p.errorType == e.t {
			t.Logf("OK: got expected error type %v for %s", e.t, p.desc)
		} else {
			t.Errorf("unexpected parse error: %s", e)
			return
		}
	}
	t.Logf("parsed config for %s", p.desc)
	t.Log(c)
	p.runConfigTests(t, c)
}

func (p *parseTest) runConfigTests(t *testing.T, c *Config) {
	ok := true
	for _, test := range p.configTests {
		if test.pass(c) {
			t.Logf("OK: %s", test.desc)
		} else {
			t.Errorf("config predicate failed: %s", test.desc)
			ok = false
		}
	}
	if ok {
		t.Logf("OK: %s", p.desc)
	}
}

// an individual test for confirming that a parsed config struct meets some
// predicate
type configTest struct {
	desc string
	pass configPredicate
}

// inverts a given config predicate
func inv(fn configPredicate) configPredicate {
	return func(c *Config) bool {
		return !fn(c)
	}
}

func hasKey(s string) configPredicate {
	return func(c *Config) bool {
		return c.hasKey(s)
	}
}

func hasValue(key string, expected interface{}) configPredicate {
	switch t := expected.(type) {
	case string:
		return hasStringValue(key, t)
	default:
		panic("no we can't do that yet")
	}
}

func hasStringValue(key string, expected string) configPredicate {
	return func(c *Config) bool {
		return c.GetString(key) == expected
	}
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		test.run(t)
	}
}
