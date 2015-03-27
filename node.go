package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type nodeType int

const (
	n_error nodeType = iota
	n_root
	n_comment
	n_assignment
	n_string
)

var indent = "  "

type node interface {
	Type() nodeType
	parse(*parser) error
	pretty(io.Writer, string) error
	eval(map[string]interface{}) (interface{}, error)
}

type rootNode struct {
	children []node
}

func newRootNode() node {
	return &rootNode{children: make([]node, 0, 8)}
}

func (n *rootNode) Type() nodeType {
	return n_root
}

func (n *rootNode) parse(p *parser) error {
	for {
		t := p.next()
		switch t.t {
		case t_error:
			return fmt.Errorf("parse error: saw lex error while parsing root node: %v", t)
		case t_eof:
			return nil
		case t_comment:
			n.addChild(&commentNode{t.s})
		case t_name:
			nn := &assignmentNode{name: t.s}
			if err := nn.parse(p); err != nil {
				return err
			}
			n.addChild(nn)
		default:
			return fmt.Errorf("parse error: unexpected token type %v while parsing root node", t.t)
		}
	}
}

func (n *rootNode) addChild(child node) {
	if n.children == nil {
		n.children = make([]node, 0, 8)
	}
	n.children = append(n.children, child)
}

func (n *rootNode) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	for _, child := range n.children {
		fmt.Fprintf(&buf, "%s, ", child)
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 2)
	}
	buf.WriteString("}")
	return buf.String()
}

func (n *rootNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sroot{\n", prefix)
	for _, child := range n.children {
		if err := child.pretty(w, prefix+indent); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "%s}\n", prefix)
	return nil
}

func (n *rootNode) eval(ctx map[string]interface{}) (interface{}, error) {
	for _, child := range n.children {
		if _, err := child.eval(ctx); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type commentNode struct {
	body string
}

func (n *commentNode) Type() nodeType {
	return n_comment
}

func (n *commentNode) parse(p *parser) error {
	return nil
}

func (n *commentNode) String() string {
	return fmt.Sprintf("{comment: %s}", n.body)
}

func (n *commentNode) pretty(w io.Writer, prefix string) error {
	r := bufio.NewReader(strings.NewReader(n.body))
	fmt.Fprintf(w, "%scomment{\n", prefix)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			if line != "" {
				fmt.Fprintf(w, "%s%s%s\n", prefix, indent, line)
			}
			break
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s%s%s\n", prefix, indent, line)
	}
	fmt.Fprintf(w, "%s}\n", prefix)
	return nil
}

func (n *commentNode) eval(ctx map[string]interface{}) (interface{}, error) {
	return nil, nil
}

type assignmentNode struct {
	name  string
	value node
}

func (n *assignmentNode) Type() nodeType {
	return n_assignment
}

func (n *assignmentNode) parse(p *parser) error {
	t := p.next()
	switch t.t {
	case t_error:
		return fmt.Errorf("parse error: saw lex error while parsing assignment node: %v", t.s)
	case t_eof:
		return fmt.Errorf("parse error: unexpected eof in assignment node")
	case t_object_separator:
	default:
		return fmt.Errorf("parse error: unexpected %v token after name, expected =", t.t)
	}

	v, err := p.parseValue()
	if err != nil {
		return err
	}
	n.value = v
	return nil
}

func (n *assignmentNode) String() string {
	return fmt.Sprintf("{assign: name=%s, val=%v}", n.name, n.value)
}

func (n *assignmentNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sassign{\n", prefix)
	fmt.Fprintf(w, "%s%sname: %s\n", prefix, indent, n.name)
	if err := n.value.pretty(w, fmt.Sprintf("%s%svalue: ", prefix, indent)); err != nil {
		return err
	}
	fmt.Fprintf(w, "%s}\n", prefix)
	return nil
}

func (n *assignmentNode) eval(ctx map[string]interface{}) (interface{}, error) {
	if _, ok := ctx[n.name]; ok {
		return nil, fmt.Errorf("invalid re-declaration: %s", n.name)
	}
	ctx[n.name] = n.value
	return nil, nil
}

type stringNode string

func (s *stringNode) Type() nodeType {
	return n_string
}

func (s *stringNode) parse(p *parser) error {
	t := p.next()
	if t.t != t_string {
		return fmt.Errorf("unexpected %s while looking for string token", t.t)
	}
	*s = stringNode(t.s)
	return nil
}

func (s *stringNode) pretty(w io.Writer, prefix string) error {
	_, err := fmt.Fprintf(w, "%s%s\n", prefix, string(*s))
	return err
}

func (s *stringNode) eval(ctx map[string]interface{}) (interface{}, error) {
	return string(*s), nil
}

type list []interface{}
type object map[string]interface{}
