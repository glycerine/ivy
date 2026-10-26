package abc

// Bounded DSD formula parsing and evaluation from opt/dau/dauDsd.c.

import (
	"fmt"
	"math/bits"
	"sort"
	"strings"
)

type DauDSDOptions struct {
	MaxBytes, MaxNodes, MaxDepth, MaxVariables, MaxWords int
	Stop                                                 func() bool
}

type DauDSDNodeKind uint8

const (
	DauDSDNone DauDSDNodeKind = iota
	DauDSDConstant
	DauDSDVariable
	DauDSDAnd
	DauDSDXor
	DauDSDMux
	DauDSDPrime
)

type dauDSDNode struct {
	kind       DauDSDNodeKind
	complement bool
	value      int
	children   []int
	bindings   []int
	primeBits  []bool
}

// DauDSDFormula is an immutable, owned parsed formula. Its internal indexed
// arena avoids interface-valued syntax trees and makes work accounting exact.
type DauDSDFormula struct {
	nodes     []dauDSDNode
	root      int
	variables int
}

type dauDSDParser struct {
	text       string
	pos, depth int
	opts       DauDSDOptions
	nodes      []dauDSDNode
}

func normalizeDauDSDOptions(o DauDSDOptions) (DauDSDOptions, error) {
	if o.MaxBytes < 0 || o.MaxNodes < 0 || o.MaxDepth < 0 || o.MaxVariables < 0 || o.MaxWords < 0 {
		return o, ErrDimension
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = DauMaxFormulaBytes
	}
	if o.MaxNodes == 0 {
		o.MaxNodes = 2000
	}
	if o.MaxDepth == 0 {
		o.MaxDepth = 256
	}
	if o.MaxVariables == 0 {
		o.MaxVariables = DauMaxVariables
	}
	if o.MaxWords == 0 {
		o.MaxWords = 1 << 20
	}
	return o, nil
}

// DauDSDMatches returns the matching byte index for every delimiter and -1 for
// other bytes. Unlike the assert-based source helper, malformed input is an error.
func DauDSDMatches(text string, maxBytes int) ([]int, error) {
	if maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = DauMaxFormulaBytes
	}
	if len(text) > maxBytes {
		return nil, ErrLimit
	}
	result := make([]int, len(text))
	for i := range result {
		result[i] = -1
	}
	stack := make([]int, 0, 32)
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(', '[', '<', '{':
			stack = append(stack, i)
		case ')', ']', '>', '}':
			if len(stack) == 0 {
				return nil, fmt.Errorf("DSD byte %d: %w", i, ErrInvariant)
			}
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if (text[j] == '(' && text[i] != ')') || (text[j] == '[' && text[i] != ']') ||
				(text[j] == '<' && text[i] != '>') || (text[j] == '{' && text[i] != '}') {
				return nil, fmt.Errorf("DSD bytes %d and %d: %w", j, i, ErrInvariant)
			}
			result[i], result[j] = j, i
		}
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("DSD byte %d: %w", stack[len(stack)-1], ErrInvariant)
	}
	return result, nil
}

func ParseDauDSD(text string, options DauDSDOptions) (*DauDSDFormula, error) {
	o, err := normalizeDauDSDOptions(options)
	if err != nil {
		return nil, err
	}
	if len(text) == 0 {
		return nil, ErrDimension
	}
	if len(text) > o.MaxBytes {
		return nil, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	if _, err = DauDSDMatches(text, o.MaxBytes); err != nil {
		return nil, err
	}
	p := &dauDSDParser{text: text, opts: o}
	var root int
	if text == "0" || text == "1" {
		root, err = p.add(dauDSDNode{kind: DauDSDConstant, value: int(text[0] - '0')})
		p.pos = 1
	} else {
		root, err = p.parseExpr()
	}
	if err != nil {
		return nil, err
	}
	if p.pos != len(text) {
		return nil, fmt.Errorf("DSD byte %d: trailing input: %w", p.pos, ErrInvariant)
	}
	variables := 0
	for _, n := range p.nodes {
		if n.kind == DauDSDVariable && n.value+1 > variables {
			variables = n.value + 1
		}
	}
	if variables > o.MaxVariables {
		return nil, ErrLimit
	}
	f := &DauDSDFormula{nodes: p.nodes, root: root, variables: variables}
	if err = f.validateBindings(root, nil); err != nil {
		return nil, err
	}
	return f, nil
}

func (p *dauDSDParser) add(n dauDSDNode) (int, error) {
	if len(p.nodes) >= p.opts.MaxNodes {
		return 0, ErrLimit
	}
	p.nodes = append(p.nodes, n)
	return len(p.nodes) - 1, nil
}

func (p *dauDSDParser) parseExpr() (int, error) {
	if p.opts.Stop != nil && p.opts.Stop() {
		return 0, ErrCancelled
	}
	if p.pos >= len(p.text) {
		return 0, fmt.Errorf("DSD byte %d: missing expression: %w", p.pos, ErrInvariant)
	}
	if p.depth >= p.opts.MaxDepth {
		return 0, ErrLimit
	}
	p.depth++
	defer func() { p.depth-- }()
	complement := false
	if p.text[p.pos] == '!' {
		complement = true
		p.pos++
	}
	if p.pos >= len(p.text) {
		return 0, fmt.Errorf("DSD byte %d: missing complemented expression: %w", p.pos, ErrInvariant)
	}
	c := p.text[p.pos]
	if c >= 'a' && c <= 'z' {
		p.pos++
		return p.add(dauDSDNode{kind: DauDSDVariable, complement: complement, value: int(c - 'a')})
	}
	if c == '(' || c == '[' || c == '<' {
		p.pos++
		close := byte(')')
		kind := DauDSDAnd
		if c == '[' {
			close, kind = ']', DauDSDXor
		}
		if c == '<' {
			close, kind = '>', DauDSDMux
		}
		children, err := p.parseList(close)
		if err != nil {
			return 0, err
		}
		if kind == DauDSDMux && len(children) != 3 {
			return 0, fmt.Errorf("DSD mux has %d inputs: %w", len(children), ErrDimension)
		}
		if kind != DauDSDMux && len(children) < 2 {
			return 0, fmt.Errorf("DSD group has %d inputs: %w", len(children), ErrDimension)
		}
		n := dauDSDNode{kind: kind, complement: complement, children: children}
		if kind == DauDSDMux && p.pos < len(p.text) && p.text[p.pos] == '{' {
			p.pos++
			n.bindings, err = p.parseList('}')
			if err != nil {
				return 0, err
			}
			if len(n.bindings) == 0 {
				return 0, fmt.Errorf("DSD mux binding list: %w", ErrDimension)
			}
			if len(n.bindings) > p.opts.MaxVariables {
				return 0, ErrLimit
			}
		}
		return p.add(n)
	}
	if isDauPrimeDigit(c) {
		start := p.pos
		for p.pos < len(p.text) && isDauPrimeDigit(p.text[p.pos]) {
			p.pos++
		}
		digits := p.text[start:p.pos]
		nvars, primeBits, err := parseDauPrime(digits)
		if err != nil {
			return 0, err
		}
		if nvars > p.opts.MaxVariables {
			return 0, ErrLimit
		}
		if p.pos >= len(p.text) || p.text[p.pos] != '{' {
			return 0, fmt.Errorf("DSD byte %d: prime fanins missing: %w", p.pos, ErrInvariant)
		}
		p.pos++
		children, err := p.parseList('}')
		if err != nil {
			return 0, err
		}
		if len(children) != nvars {
			return 0, fmt.Errorf("DSD prime has %d fanins, want %d: %w", len(children), nvars, ErrDimension)
		}
		return p.add(dauDSDNode{kind: DauDSDPrime, complement: complement, children: children, primeBits: primeBits})
	}
	return 0, fmt.Errorf("DSD byte %d: invalid symbol %q: %w", p.pos, c, ErrInvariant)
}

func (p *dauDSDParser) parseList(close byte) ([]int, error) {
	children := make([]int, 0, 4)
	for {
		if p.pos >= len(p.text) {
			return nil, fmt.Errorf("DSD byte %d: missing %q: %w", p.pos, close, ErrInvariant)
		}
		if p.text[p.pos] == close {
			p.pos++
			return children, nil
		}
		child, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
}

func isDauPrimeDigit(c byte) bool { return c >= '0' && c <= '9' || c >= 'A' && c <= 'F' }

func parseDauPrime(s string) (int, []bool, error) {
	if len(s) == 0 {
		return 0, nil, ErrDimension
	}
	nvars := 0
	if len(s) == 1 {
		switch s[0] {
		case '0', 'F':
			nvars = 0
		case '5', 'A':
			nvars = 1
		default:
			nvars = 2
		}
	} else {
		if len(s)&(len(s)-1) != 0 {
			return 0, nil, fmt.Errorf("DSD prime has %d hex digits: %w", len(s), ErrDimension)
		}
		nvars = 2 + bits.Len(uint(len(s))) - 1
	}
	values := make([]bool, 1<<uint(nvars))
	for digitIndex := 0; digitIndex < len(s); digitIndex++ {
		c := s[len(s)-1-digitIndex]
		v := int(c - '0')
		if c >= 'A' {
			v = int(c-'A') + 10
		}
		for bit := 0; bit < 4; bit++ {
			i := 4*digitIndex + bit
			if i < len(values) {
				values[i] = v&(1<<uint(bit)) != 0
			}
		}
	}
	return nvars, values, nil
}

func (f *DauDSDFormula) validateBindings(index int, local *int) error {
	if f == nil || index < 0 || index >= len(f.nodes) {
		return ErrInvariant
	}
	n := &f.nodes[index]
	if n.kind == DauDSDVariable && local != nil && n.value >= *local {
		return fmt.Errorf("DSD local variable %c exceeds %d bindings: %w", 'a'+n.value, *local, ErrDimension)
	}
	for _, child := range n.children {
		next := local
		if n.kind == DauDSDMux && len(n.bindings) != 0 {
			count := len(n.bindings)
			next = &count
		}
		if err := f.validateBindings(child, next); err != nil {
			return err
		}
	}
	for _, child := range n.bindings {
		if err := f.validateBindings(child, local); err != nil {
			return err
		}
	}
	return nil
}

func (f *DauDSDFormula) VariableCount() int {
	if f == nil {
		return 0
	}
	return f.variables
}

func (f *DauDSDFormula) AndCount() (int, error) {
	if f == nil || f.root < 0 || f.root >= len(f.nodes) {
		return 0, ErrDimension
	}
	var count func(int) int
	count = func(i int) int {
		n := &f.nodes[i]
		total := 0
		parts := n.children
		if n.kind == DauDSDMux && len(n.bindings) != 0 {
			parts = n.bindings
		}
		for _, child := range parts {
			total += count(child)
		}
		switch n.kind {
		case DauDSDAnd:
			total += len(n.children) - 1
		case DauDSDXor:
			total += 3 * (len(n.children) - 1)
		case DauDSDMux, DauDSDPrime:
			total += 3
		}
		return total
	}
	return count(f.root), nil
}

func (f *DauDSDFormula) String() string {
	if f == nil {
		return ""
	}
	return f.render(f.root, false)
}

func (f *DauDSDFormula) render(index int, normalized bool) string {
	n := &f.nodes[index]
	var body string
	switch n.kind {
	case DauDSDConstant:
		body = string(byte('0' + n.value))
	case DauDSDVariable:
		body = string(byte('a' + n.value))
	case DauDSDAnd, DauDSDXor, DauDSDMux:
		parts := make([]string, len(n.children))
		for i, child := range n.children {
			// The source treats the expression before a mux binding list as
			// a local truth template and only normalizes the bindings.
			parts[i] = f.render(child, normalized && !(n.kind == DauDSDMux && len(n.bindings) != 0))
		}
		if normalized && (n.kind == DauDSDAnd || n.kind == DauDSDXor) {
			sort.SliceStable(parts, func(i, j int) bool {
				left, right := dauDSDVariableSequence(parts[i]), dauDSDVariableSequence(parts[j])
				if left == right {
					return parts[i] < parts[j]
				}
				return left < right
			})
		}
		open, close := "(", ")"
		if n.kind == DauDSDXor {
			open, close = "[", "]"
		}
		if n.kind == DauDSDMux {
			open, close = "<", ">"
		}
		body = open + strings.Join(parts, "") + close
		if len(n.bindings) != 0 {
			parts = make([]string, len(n.bindings))
			for i, child := range n.bindings {
				parts[i] = f.render(child, normalized)
			}
			body += "{" + strings.Join(parts, "") + "}"
		}
	case DauDSDPrime:
		// Prime spelling is retained indirectly through its truth bits.
		digits := make([]byte, max(1, len(n.primeBits)/4))
		for d := range digits {
			v := 0
			for bit := 0; bit < 4; bit++ {
				i := 4*d + bit
				// ABC stretches zero- and one-variable functions across a
				// complete hexadecimal digit.
				if n.primeBits[i%len(n.primeBits)] {
					v |= 1 << uint(bit)
				}
			}
			if v < 10 {
				digits[len(digits)-1-d] = byte('0' + v)
			} else {
				digits[len(digits)-1-d] = byte('A' + v - 10)
			}
		}
		parts := make([]string, len(n.children))
		for i, child := range n.children {
			parts[i] = f.render(child, normalized)
		}
		body = string(digits) + "{" + strings.Join(parts, "") + "}"
	}
	if n.complement {
		return "!" + body
	}
	return body
}

func (f *DauDSDFormula) Normalized() string {
	if f == nil {
		return ""
	}
	return f.render(f.root, true)
}

func dauDSDVariableSequence(text string) string {
	var result strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] >= 'a' && text[i] <= 'z' {
			result.WriteByte(text[i])
		}
	}
	return result.String()
}

// TruthTable evaluates the formula for the requested global variable domain.
// Special mux bindings are evaluated in the outer domain, then substituted for
// the local variables inside the mux, matching Dau_Dsd6ToTruth_rec.
func (f *DauDSDFormula) TruthTable(variables, maxWords int, stop func() bool) (*TruthTable, error) {
	if f == nil || variables < f.variables || variables > 36 || maxWords < 0 {
		return nil, ErrDimension
	}
	t, err := NewConstantTruthTable(variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	assignments := uint64(1) << uint(variables)
	for assignment := uint64(0); assignment < assignments; assignment++ {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		value, err := f.eval(f.root, assignment, nil)
		if err != nil {
			return nil, err
		}
		if value {
			if err = t.Set(assignment, true); err != nil {
				return nil, err
			}
		}
	}
	return t, nil
}

func (f *DauDSDFormula) eval(index int, assignment uint64, local []bool) (bool, error) {
	n := &f.nodes[index]
	var value bool
	switch n.kind {
	case DauDSDConstant:
		value = n.value != 0
	case DauDSDVariable:
		if local != nil {
			if n.value >= len(local) {
				return false, ErrInvariant
			}
			value = local[n.value]
		} else {
			value = assignment&(uint64(1)<<uint(n.value)) != 0
		}
	case DauDSDAnd:
		value = true
		for _, child := range n.children {
			v, err := f.eval(child, assignment, local)
			if err != nil {
				return false, err
			}
			value = value && v
		}
	case DauDSDXor:
		for _, child := range n.children {
			v, err := f.eval(child, assignment, local)
			if err != nil {
				return false, err
			}
			value = value != v
		}
	case DauDSDMux:
		env := local
		if len(n.bindings) != 0 {
			env = make([]bool, len(n.bindings))
			for i, child := range n.bindings {
				v, err := f.eval(child, assignment, local)
				if err != nil {
					return false, err
				}
				env[i] = v
			}
		}
		c, err := f.eval(n.children[0], assignment, env)
		if err != nil {
			return false, err
		}
		a, err := f.eval(n.children[1], assignment, env)
		if err != nil {
			return false, err
		}
		b, err := f.eval(n.children[2], assignment, env)
		if err != nil {
			return false, err
		}
		value = c && a || !c && b
	case DauDSDPrime:
		truthIndex := 0
		for i, child := range n.children {
			v, err := f.eval(child, assignment, local)
			if err != nil {
				return false, err
			}
			if v {
				truthIndex |= 1 << uint(i)
			}
		}
		value = n.primeBits[truthIndex]
	default:
		return false, ErrInvariant
	}
	if n.complement {
		value = !value
	}
	return value, nil
}
