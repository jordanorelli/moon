package main

import (
	"testing"
)

var primitivesTests = []struct {
	in  string
	out []token
}{
	{`"x"`, []token{{t_string, "x"}}},
	{`"yes"`, []token{{t_string, "yes"}}},
	{`"this one has spaces"`, []token{{t_string, "this one has spaces"}}},
	{`"this one has \"quotes\" in it"`, []token{{t_string, `this one has "quotes" in it`}}},
	{"`this one is delimited by backticks`", []token{{t_string, "this one is delimited by backticks"}}},
	{`  "this one has white space on either end"  `, []token{{t_string, "this one has white space on either end"}}},
	{`name`, []token{{t_name, "name"}}},
	{`name_with_underscore`, []token{{t_name, "name_with_underscore"}}},
	{`  name_surrounded_by_whitespace  `, []token{{t_name, "name_surrounded_by_whitespace"}}},
	{`name1`, []token{{t_name, "name1"}}},
	{`camelName`, []token{{t_name, "camelName"}}},
	{`Type`, []token{{t_type, "Type"}}},
	{`CamelType`, []token{{t_type, "CamelType"}}},
	{`Type_1_2`, []token{{t_type, "Type_1_2"}}},
	{`=`, []token{{t_equals, "="}}},
	{` = `, []token{{t_equals, "="}}},
	{`"x" "y"`, []token{{t_string, "x"}, {t_string, "y"}}},
	{`x = "sam"`, []token{
		{t_name, "x"},
		{t_equals, "="},
		{t_string, "sam"},
	}},
	{`# this is a comment`, []token{{t_comment, " this is a comment"}}},
	{`
    # comment line one
    # comment line two
    `, []token{{t_comment, " comment line one"}, {t_comment, " comment line two"}}},
	{`[]`, []token{{t_list_start, "["}, {t_list_end, "]"}}},
	{`["item"]`, []token{{t_list_start, "["}, {t_string, "item"}, {t_list_end, "]"}}},
}

func TestLexPrimities(t *testing.T) {
	for _, test := range primitivesTests {
		tokens, err := fullTokens(lexString(test.in))
		if err != nil {
			t.Error(err)
			continue
		}
		tokens = tokens[:len(tokens)-1]
		if len(tokens) != len(test.out) {
			t.Errorf("expected %d token, saw %d: %v", len(test.out), len(tokens), tokens)
			continue
		}
		for i := range tokens {
			if tokens[i] != test.out[i] {
				t.Errorf("token %d is %v, expected %v", i, tokens[i], test.out[i])
			}
		}
		t.Logf("OK: %s", test.in)
	}
}
