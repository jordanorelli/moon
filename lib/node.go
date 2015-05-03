package moon

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

type nodeType int

const (
	n_error nodeType = iota
	n_root
	n_comment
	n_assignment
	n_string
	n_number
	n_list
	n_object
	n_variable
	n_bool
)

var indent = "  "

type context struct {
	public  map[string]interface{}
	private map[string]interface{}
}

func newContext() *context {
	return &context{make(map[string]interface{}), make(map[string]interface{})}
}

func (c *context) get(name string) (interface{}, bool) {
	if v, ok := c.public[name]; ok {
		return v, true
	}
	if v, ok := c.private[name]; ok {
		return v, true
	}
	return nil, false
}

type node interface {
	Type() nodeType
	parse(*parser) error
	pretty(io.Writer, string) error
	eval(*context) (interface{}, error)
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
		case t_variable:
			nn := &assignmentNode{name: t.s, unexported: true}
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
	fmt.Fprintf(w, "%sroot:\n", prefix)
	for _, child := range n.children {
		if err := child.pretty(w, prefix+indent); err != nil {
			return err
		}
	}
	return nil
}

func (n *rootNode) eval(ctx *context) (interface{}, error) {
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
	fmt.Fprintf(w, "%scomment:\n", prefix)
	r := bufio.NewReader(strings.NewReader(n.body))
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
	return nil
}

func (n *commentNode) eval(ctx *context) (interface{}, error) {
	return nil, nil
}

type assignmentNode struct {
	name       string
	value      node
	unexported bool
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
	fmt.Fprintf(w, "%sassign:\n", prefix)
	fmt.Fprintf(w, "%s%sname:\n", prefix, indent)
	fmt.Fprintf(w, "%s%s%s%s\n", prefix, indent, indent, n.name)
	fmt.Fprintf(w, "%s%svalue:\n", prefix, indent)
	if err := n.value.pretty(w, prefix+indent+indent); err != nil {
		return err
	}
	return nil
}

func (n *assignmentNode) eval(ctx *context) (interface{}, error) {
	if _, ok := ctx.get(n.name); ok {
		return nil, fmt.Errorf("invalid re-declaration: %s", n.name)
	}
	v, err := n.value.eval(ctx)
	if err != nil {
		return nil, err
	}
	if n.unexported {
		ctx.private[n.name] = v
	} else {
		ctx.public[n.name] = v
	}
	return nil, nil
}

func (n *assignmentNode) isHidden() bool {
	return strings.HasPrefix(n.name, ".")
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
	fmt.Fprintf(w, "%sstring:\n", prefix)
	_, err := fmt.Fprintf(w, "%s%s%s\n", prefix, indent, string(*s))
	return err
}

func (s *stringNode) eval(ctx *context) (interface{}, error) {
	return string(*s), nil
}

type numberType int

const (
	num_int numberType = iota
	num_float
	num_complex
)

type numberNode struct {
	t numberType
	c complex128
	i int
	f float64
}

func (n *numberNode) Type() nodeType {
	return n_number
}

func (n *numberNode) parse(p *parser) error {
	t := p.next()
	if t.t != t_real_number {
		return fmt.Errorf("unexpected %s token while parsing number", t.t)
	}

	if p.peek().t == t_imaginary_number {
		n.t = num_complex
		s := t.s + p.next().s
		if _, err := fmt.Sscan(s, &n.c); err != nil {
			return fmt.Errorf("ungood imaginary number format %s: %s", s, err)
		}
		return nil
	}

	i, err := strconv.ParseInt(t.s, 0, 64)
	if err == nil {
		n.t = num_int
		n.i = int(i)
		return nil
	}

	f, err := strconv.ParseFloat(t.s, 64)
	if err == nil {
		n.t = num_float
		n.f = f
		return nil
	}

	return fmt.Errorf("this token broke the number parser: %s", t)
}

func (n *numberNode) pretty(w io.Writer, prefix string) error {
	v, err := n.eval(nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%snumber:\n%s%s%v\n", prefix, prefix, indent, v)
	return nil
}

func (n *numberNode) eval(ctx *context) (interface{}, error) {
	switch n.t {
	case num_int:
		return n.i, nil
	case num_float:
		return n.f, nil
	case num_complex:
		return n.c, nil
	default:
		panic("whoerps")
	}
}

type listNode []node

func (l *listNode) Type() nodeType {
	return n_list
}

func (l *listNode) parse(p *parser) error {
	if p.peek().t == t_list_end {
		p.next()
		return nil
	}

	if n, err := p.parseValue(); err != nil {
		return err
	} else {
		*l = append(*l, n)
	}

	switch t := p.peek(); t.t {
	case t_list_end:
		p.next()
		return nil
	default:
		return l.parse(p)
	}
}

func (l *listNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%slist:\n", prefix)
	for _, n := range *l {
		if err := n.pretty(w, prefix+indent); err != nil {
			return err
		}
	}
	return nil
}

func (l *listNode) eval(ctx *context) (interface{}, error) {
	out := make([]interface{}, 0, len(*l))
	for _, n := range *l {
		v, err := n.eval(ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

type objectNode map[string]node

func (o *objectNode) Type() nodeType {
	return n_object
}

func (o *objectNode) parse(p *parser) error {
	if p.peek().t == t_object_end {
		p.next()
		return nil
	}
	if err := p.ensureNext(t_name, "looking for object field name in parseObject"); err != nil {
		return err
	}
	field_name := p.next().s
	if err := p.ensureNext(t_object_separator, "looking for object separator in parseObject"); err != nil {
		return err
	}
	p.next()

	if n, err := p.parseValue(); err != nil {
		return err
	} else {
		(*o)[field_name] = n
	}

	switch t := p.peek(); t.t {
	case t_object_end:
		p.next()
		return nil
	default:
		return o.parse(p)
	}
}

func (o *objectNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sobject:\n", prefix)
	keys := make([]string, 0, len(*o))
	for key := range *o {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(w, "%s%s:\n", prefix+indent, key)
		err := (*o)[key].pretty(w, prefix+indent+indent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *objectNode) eval(ctx *context) (interface{}, error) {
	out := make(map[string]interface{}, len(*o))
	for name, node := range *o {
		v, err := node.eval(ctx)
		if err != nil {
			return nil, err
		}
		out[name] = v
	}
	return out, nil
}

type variableNode struct {
	name string
}

func (v *variableNode) Type() nodeType {
	return n_variable
}

func (v *variableNode) parse(p *parser) error {
	t := p.next()
	if t.t != t_variable {
		return fmt.Errorf("unexpected %s token when parsing variable", t.t)
	}
	v.name = t.s
	return nil
}

func (v *variableNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%svariable:\n", prefix)
	fmt.Fprintf(w, "%s%s\n", prefix+indent, v.name)
	return nil
}

func (v *variableNode) eval(ctx *context) (interface{}, error) {
	value, ok := ctx.get(v.name)
	if !ok {
		return nil, fmt.Errorf("undefined variable: %s", *v)
	}
	return value, nil
}

type boolNode bool

func (b *boolNode) Type() nodeType {
	return n_bool
}

func (b *boolNode) parse(p *parser) error {
	t := p.next()
	if t.t != t_bool {
		return fmt.Errorf("unexpected %s token while parsing bool", t.t)
	}
	switch t.s {
	case "true":
		*b = true
	case "false":
	default:
		return fmt.Errorf("illegal lexeme for bool token: %s", t.s)
	}
	return nil
}

func (b *boolNode) pretty(w io.Writer, prefix string) error {
	fmt.Fprintf(w, "%sbool:\n", prefix)
	fmt.Fprintf(w, "%s%t\n", prefix+indent, *b)
	return nil
}

func (b *boolNode) eval(ctx *context) (interface{}, error) {
	return bool(*b), nil
}
