package main

import (
	"io"
)

const ()

func parse(r io.Reader) (node, error) {
	p := &parser{
		root:   newRootNode(),
		input:  lex(r),
		backup: make([]token, 0, 8),
	}
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p.root, nil
}

type parser struct {
	root   node
	input  chan token
	backup []token
}

func (p *parser) parse() error {
	if p.root == nil {
		p.root = newRootNode()
	}
	return p.root.parse(p)
}

// returns the next token and advances the input stream
func (p *parser) next() token {
	if len(p.backup) > 0 {
		t := p.backup[len(p.backup)-1]
		p.backup = p.backup[:len(p.backup)-1]
		return t
	}
	return <-p.input
}
