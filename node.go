package main

import (
	"bytes"
	"fmt"
)

type nodeType int

const (
	n_error nodeType = iota
	n_root
	n_comment
	n_assignment
)

type node interface {
	Type() nodeType
	parse(*parser) error
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
			return fmt.Errorf("parse error: saw lex error while parsing root node: %v", t.s)
		case t_eof:
			return nil
		case t_comment:
			shit := commentNode(t.s)
			n.addChild(&shit)
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

type commentNode string

func (n commentNode) Type() nodeType {
	return n_comment
}

func (n commentNode) parse(p *parser) error {
	return nil
}

func (n commentNode) String() string {
	return fmt.Sprintf("{comment: %s}", string(n))
}

type assignmentNode struct {
	name  string
	value interface{}
}

func (n assignmentNode) Type() nodeType {
	return n_assignment
}

func (n *assignmentNode) parse(p *parser) error {
	t := p.next()
	switch t.t {
	case t_error:
		return fmt.Errorf("parse error: saw lex error while parsing assignment node: %v", t.s)
	case t_eof:
		return fmt.Errorf("parse error: unexpected eof in assignment node")
	case t_equals:
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
	return fmt.Sprintf("{assign: name=%s, val=%s}", n.name, n.value)
}

type list struct {
	head *listElem
	tail *listElem
}

func (l list) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for e := l.head; e != nil; e = e.next {
		fmt.Fprintf(&buf, "%v, ", e.value)
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 2)
	}
	buf.WriteString("]")
	return buf.String()
}

func (l *list) append(v interface{}) {
	e := listElem{value: v}
	if l.head == nil {
		l.head = &e
	}

	if l.tail != nil {
		l.tail.next = &e
		e.prev = l.tail
	}
	l.tail = &e
}

type listElem struct {
	value interface{}
	prev  *listElem
	next  *listElem
}

type object map[string]interface{}
