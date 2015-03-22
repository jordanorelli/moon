package main

import (
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
				commentNode(" just a comment"),
			},
		},
	},
	{
		source: `name = "jordan"`,
		root: &rootNode{
			children: []node{
				&assignmentNode{
					name:  "name",
					value: "jordan",
				},
			},
		},
	},
	{
		source: `
        first_name = "jordan"
        last_name = "orelli"
        `,
		root: &rootNode{},
	},
	{
		source: `
        # personal info
        first_name = "jordan"
        last_name = "orelli"
        `,
		root: &rootNode{},
	},
	{
		source: `
        first_name = "jordan" # yep, that's my name
        last_name = "orelli"  # comments should be able to follow other shit
        `,
		root: &rootNode{},
	},
	{
		source: `
        heroes = ["lina", "cm"]
        `,
		root: &rootNode{},
	},
	{
		source: `
        nested = [["one", "two"], ["three", "four"]]
        `,
		root: &rootNode{},
	},
	{
		source: `
        nested = [
            ["one", "two"],
            ["three", "four"],
        ]
        `,
		root: &rootNode{},
	},
	{
		source: `
	    admin = {first_name: "jordan", last_name: "orelli"}
	    `,
		root: &rootNode{},
	},
	{
		source: `
        http = {
            port: "9000",
            routes: "/path/to/some/file",
        }
        `,
		root: &rootNode{},
	},
}

type parseTest struct {
	source string
	root   *rootNode
}

func (p *parseTest) run(t *testing.T) {
	r := strings.NewReader(p.source)
	n, err := parse(r)
	if err != nil {
		t.Errorf("parse error: %v", err)
		return
	}
	if n.Type() != n_root {
		t.Errorf("we expected a root node object, but instead we got: %s", n.Type())
	}
	t.Logf("output: %v", n)
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		test.run(t)
	}
}
