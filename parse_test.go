package main

import (
	"reflect"
	"strings"
	"testing"
)

var parseTests = []parseTest{
	{
		source: ``,
		root: &rootNode{
			children: []node{},
		},
	},
	{
		source: `# just a comment`,
		root: &rootNode{
			children: []node{
				&commentNode{" just a comment"},
			},
		},
	},
	{
		source: `name = "jordan"`,
		root: &rootNode{
			children: []node{
				&assignmentNode{"name", "jordan"},
			},
		},
	},
	{
		source: `
        hostname = "jordanorelli.com"
	    port = 9000
        freq = 1e9
        duty = 0.2
        neg = -2
        neg2 = -2.3
        imag = 1+2i
	    `,
		root: &rootNode{
			children: []node{
				&assignmentNode{"hostname", "jordanorelli.com"},
				&assignmentNode{"port", 9000},
				&assignmentNode{"freq", 1e9},
				&assignmentNode{"duty", 0.2},
				&assignmentNode{"neg", -2},
				&assignmentNode{"neg2", -2.3},
				&assignmentNode{"imag", 1 + 2i},
			},
		},
	},
	{
		source: `
	       first_name = "jordan" # yep, that's my name
	       last_name = "orelli"  # comments should be able to follow other shit
	       `,
		root: &rootNode{
			children: []node{
				&assignmentNode{"first_name", "jordan"},
				&commentNode{" yep, that's my name"},
				&assignmentNode{"last_name", "orelli"},
				&commentNode{" comments should be able to follow other shit"},
			},
		},
	},
	{
		source: `
	       heroes = ["lina", "cm"]
	       `,
		root: &rootNode{
			children: []node{
				&assignmentNode{"heroes", list{"lina", "cm"}},
			},
		},
	},
	{
		source: `
	       nested = [["one", "two"], ["three", "four"]]
	       `,
		root: &rootNode{
			children: []node{
				&assignmentNode{"nested", list{list{"one", "two"}, list{"three", "four"}}},
			},
		},
	},
	{
		source: `
	       nested = [
	           ["one", "two"],
	           ["three", "four"],
	       ]
	       `,
		root: &rootNode{
			children: []node{
				&assignmentNode{"nested", list{list{"one", "two"}, list{"three", "four"}}},
			},
		},
	},
	{
		source: `
	    admin = {first_name: "jordan", last_name: "orelli"}
	    `,
		root: &rootNode{
			children: []node{
				&assignmentNode{"admin", object{
					"first_name": "jordan",
					"last_name":  "orelli",
				}},
			},
		},
	},
	{
		source: `
	       http = {
	           port: 9000,
	           routes: "/path/to/some/file",
	       }
	       `,
		root: &rootNode{
			children: []node{
				&assignmentNode{"http", object{
					"port":   9000,
					"routes": "/path/to/some/file",
				}},
			},
		},
	},
}

type parseTest struct {
	source string
	root   node
}

func (p *parseTest) run(t *testing.T) {
	r := strings.NewReader(p.source)
	n, err := parse(r)
	if err != nil {
		t.Errorf("parse error: %v", err)
		return
	}
	if !reflect.DeepEqual(p.root, n) {
		t.Errorf("trees are not equal.  expected:\n%v\nsaw:\n%v", p.root, n)
	} else {
		t.Logf("OK trees are equal: %v = %v", p.root, n)
	}
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		test.run(t)
	}
}
