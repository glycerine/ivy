package abc

// Derived from ABC src/proof/live/ltl_parser.c.

import (
	"fmt"
	"strings"
	"unicode"
)

// LTLOperator is the node kind used by the liveness formula parser.
type LTLOperator uint8

const (
	LTLAnd LTLOperator = iota
	LTLOr
	LTLNot
	LTLImplies
	LTLGlobally
	LTLEventually
	LTLNext
	LTLUntil
	LTLAtom
)

// LTLNode is an owned syntax tree. Atom nodes have a name, unary nodes have
// Left, and binary nodes have Left and Right.
type LTLNode struct {
	Operator    LTLOperator
	Name        string
	Left, Right *LTLNode
}

type LTLParseOptions struct {
	MaxBytes, MaxNodes, MaxDepth int
	Stop                         func() bool
}

// LTLParseError identifies the byte that made a formula invalid.
type LTLParseError struct {
	Offset  int
	Message string
}

func (e *LTLParseError) Error() string {
	return fmt.Sprintf("invalid LTL formula at byte %d: %s", e.Offset, e.Message)
}

type ltlPrefixParser struct {
	text          string
	position      int
	nodes         int
	maxNodes      int
	maxDepth      int
	stop          func() bool
	cancelCounter int
}

func ParseLTLPrefix(text string, options LTLParseOptions) (*LTLNode, error) {
	if options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxDepth < 0 {
		return nil, ErrDimension
	}
	maxBytes, maxNodes, maxDepth := options.MaxBytes, options.MaxNodes, options.MaxDepth
	if maxBytes == 0 {
		maxBytes = 1 << 20
	}
	if maxNodes == 0 {
		maxNodes = 1 << 16
	}
	if maxDepth == 0 {
		maxDepth = 2048
	}
	if len(text) > maxBytes {
		return nil, ErrLimit
	}
	if strings.IndexByte(text, 0) >= 0 {
		return nil, &LTLParseError{Offset: strings.IndexByte(text, 0), Message: "NUL byte"}
	}
	p := ltlPrefixParser{text: text, maxNodes: maxNodes, maxDepth: maxDepth, stop: options.Stop}
	node, err := p.parse(1)
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	if p.position != len(text) {
		return nil, &LTLParseError{Offset: p.position, Message: "trailing formula"}
	}
	return node, nil
}

func (p *ltlPrefixParser) skipSpace() {
	for p.position < len(p.text) && unicode.IsSpace(rune(p.text[p.position])) {
		p.position++
	}
}

func (p *ltlPrefixParser) makeNode(depth int, operator LTLOperator) (*LTLNode, error) {
	if p.maxDepth > 0 && depth > p.maxDepth {
		return nil, ErrLimit
	}
	p.nodes++
	if p.maxNodes > 0 && p.nodes > p.maxNodes {
		return nil, ErrLimit
	}
	p.cancelCounter++
	if p.stop != nil && (p.cancelCounter&31 == 1) && p.stop() {
		return nil, ErrCancelled
	}
	return &LTLNode{Operator: operator}, nil
}

func (p *ltlPrefixParser) parse(depth int) (*LTLNode, error) {
	p.skipSpace()
	if p.position >= len(p.text) {
		return nil, &LTLParseError{Offset: p.position, Message: "unexpected end"}
	}
	start := p.position
	switch p.text[p.position] {
	case '&', '+':
		op := LTLAnd
		if p.text[p.position] == '+' {
			op = LTLOr
		}
		p.position++
		node, err := p.makeNode(depth, op)
		if err != nil {
			return nil, err
		}
		if node.Left, err = p.parse(depth + 1); err != nil {
			return nil, err
		}
		if node.Right, err = p.parse(depth + 1); err != nil {
			return nil, err
		}
		return node, nil
	case '!', '~':
		p.position++
		node, err := p.makeNode(depth, LTLNot)
		if err != nil {
			return nil, err
		}
		node.Left, err = p.parse(depth + 1)
		return node, err
	case ':':
		p.position++
		p.skipSpace()
		if p.position >= len(p.text) {
			return nil, &LTLParseError{Offset: p.position, Message: "expected temporal operator"}
		}
		var operator LTLOperator
		switch p.text[p.position] {
		case 'G':
			operator = LTLGlobally
		case 'F':
			operator = LTLEventually
		case 'X':
			operator = LTLNext
		case 'U':
			operator = LTLUntil
		default:
			return nil, &LTLParseError{Offset: p.position, Message: "expected G, F, X, or U"}
		}
		p.position++
		node, err := p.makeNode(depth, operator)
		if err != nil {
			return nil, err
		}
		if node.Left, err = p.parse(depth + 1); err != nil {
			return nil, err
		}
		if operator == LTLUntil {
			if node.Right, err = p.parse(depth + 1); err != nil {
				return nil, err
			}
		}
		return node, nil
	}
	for p.position < len(p.text) {
		c := p.text[p.position]
		if c == ':' || unicode.IsSpace(rune(c)) {
			break
		}
		p.position++
	}
	if p.position == start {
		return nil, &LTLParseError{Offset: start, Message: "expected variable"}
	}
	node, err := p.makeNode(depth, LTLAtom)
	if err != nil {
		return nil, err
	}
	node.Name = p.text[start:p.position]
	return node, nil
}

func (node *LTLNode) Clone() *LTLNode {
	if node == nil {
		return nil
	}
	return &LTLNode{Operator: node.Operator, Name: node.Name, Left: node.Left.Clone(), Right: node.Right.Clone()}
}

func (node *LTLNode) Validate(options LTLParseOptions) error {
	if options.MaxNodes < 0 || options.MaxDepth < 0 {
		return ErrDimension
	}
	maxNodes, maxDepth := options.MaxNodes, options.MaxDepth
	if maxNodes == 0 {
		maxNodes = 1 << 16
	}
	if maxDepth == 0 {
		maxDepth = 2048
	}
	nodes := 0
	states := make(map[*LTLNode]uint8)
	var visit func(*LTLNode, int) error
	visit = func(current *LTLNode, depth int) error {
		if current == nil {
			return ErrDimension
		}
		if states[current] == 1 {
			return ErrInvariant
		}
		if states[current] == 2 {
			return nil
		}
		states[current] = 1
		defer func() { states[current] = 2 }()
		if depth > maxDepth {
			return ErrLimit
		}
		nodes++
		if nodes > maxNodes {
			return ErrLimit
		}
		if options.Stop != nil && nodes&31 == 1 && options.Stop() {
			return ErrCancelled
		}
		switch current.Operator {
		case LTLAtom:
			if current.Name == "" || current.Left != nil || current.Right != nil {
				return ErrDimension
			}
			return nil
		case LTLNot, LTLGlobally, LTLEventually, LTLNext:
			if current.Left == nil || current.Right != nil {
				return ErrDimension
			}
			return visit(current.Left, depth+1)
		case LTLAnd, LTLOr, LTLImplies, LTLUntil:
			if current.Left == nil || current.Right == nil {
				return ErrDimension
			}
			if err := visit(current.Left, depth+1); err != nil {
				return err
			}
			return visit(current.Right, depth+1)
		default:
			return ErrDimension
		}
	}
	return visit(node, 1)
}

func (node *LTLNode) Prefix() (string, error) {
	if err := node.Validate(LTLParseOptions{}); err != nil {
		return "", err
	}
	var output strings.Builder
	var write func(*LTLNode)
	write = func(current *LTLNode) {
		switch current.Operator {
		case LTLAtom:
			output.WriteString(current.Name)
		case LTLAnd, LTLOr, LTLImplies, LTLUntil:
			if current.Operator == LTLAnd {
				output.WriteByte('&')
			} else if current.Operator == LTLOr {
				output.WriteByte('+')
			} else if current.Operator == LTLImplies {
				output.WriteString("+!")
				write(current.Left)
				output.WriteByte(' ')
				write(current.Right)
				return
			} else {
				output.WriteString(":U")
			}
			write(current.Left)
			output.WriteByte(' ')
			write(current.Right)
		case LTLNot:
			output.WriteByte('!')
			write(current.Left)
		case LTLGlobally, LTLEventually, LTLNext:
			output.WriteByte(':')
			output.WriteByte(map[LTLOperator]byte{LTLGlobally: 'G', LTLEventually: 'F', LTLNext: 'X'}[current.Operator])
			write(current.Left)
		}
	}
	write(node)
	return output.String(), nil
}

// Infix returns the parenthesized presentation used by the source's postfix-
// named traversal routine.
func (node *LTLNode) Infix() (string, error) {
	if err := node.Validate(LTLParseOptions{}); err != nil {
		return "", err
	}
	var output strings.Builder
	var write func(*LTLNode)
	write = func(current *LTLNode) {
		switch current.Operator {
		case LTLAtom:
			output.WriteString(current.Name)
		case LTLNot:
			output.WriteByte('~')
			write(current.Left)
		case LTLGlobally, LTLEventually, LTLNext:
			if current.Operator == LTLGlobally {
				output.WriteByte('G')
			} else if current.Operator == LTLEventually {
				output.WriteByte('F')
			} else {
				output.WriteByte('X')
			}
			write(current.Left)
		case LTLAnd, LTLOr, LTLImplies, LTLUntil:
			output.WriteByte('(')
			write(current.Left)
			switch current.Operator {
			case LTLAnd:
				output.WriteByte('&')
			case LTLOr:
				output.WriteByte('+')
			case LTLImplies:
				output.WriteString("->")
			case LTLUntil:
				output.WriteByte('U')
			}
			write(current.Right)
			output.WriteByte(')')
		}
	}
	write(node)
	return output.String(), nil
}

func (node *LTLNode) IsNonTemporal() bool {
	states := make(map[*LTLNode]uint8)
	results := make(map[*LTLNode]bool)
	var check func(*LTLNode) bool
	check = func(current *LTLNode) bool {
		if current == nil || states[current] == 1 {
			return false
		}
		if states[current] == 2 {
			return results[current]
		}
		states[current] = 1
		result := false
		defer func() { states[current], results[current] = 2, result }()
		switch current.Operator {
		case LTLAtom:
			result = current.Name != "" && current.Left == nil && current.Right == nil
		case LTLNot:
			result = current.Right == nil && check(current.Left)
		case LTLAnd, LTLOr, LTLImplies:
			result = check(current.Left) && check(current.Right)
		}
		return result
	}
	return check(node)
}

// IsGFWellFormed preserves the restricted source checker: temporal nodes are
// accepted only as G(F(combinational-formula)).
func (node *LTLNode) IsGFWellFormed() bool {
	states := make(map[*LTLNode]uint8)
	results := make(map[*LTLNode]bool)
	var check func(*LTLNode) bool
	check = func(current *LTLNode) bool {
		if current == nil || states[current] == 1 {
			return false
		}
		if states[current] == 2 {
			return results[current]
		}
		states[current] = 1
		result := false
		defer func() { states[current], results[current] = 2, result }()
		switch current.Operator {
		case LTLAtom:
			result = current.Name != "" && current.Left == nil && current.Right == nil
		case LTLNot:
			result = current.Right == nil && check(current.Left)
		case LTLAnd, LTLOr, LTLImplies:
			result = check(current.Left) && check(current.Right)
		case LTLGlobally:
			result = current.Right == nil && current.Left != nil && current.Left.Operator == LTLEventually && current.Left.Right == nil && current.Left.Left.IsNonTemporal()
		}
		return result
	}
	return check(node)
}

type LTLValuation map[string]bool

// EvaluateLTL uses strong finite-trace semantics: X is false at the final
// frame and U requires its right operand to occur before the trace ends.
func EvaluateLTL(node *LTLNode, trace []LTLValuation, frame int, options LTLParseOptions) (bool, error) {
	if frame < 0 || frame >= len(trace) || len(trace) == 0 {
		return false, ErrDimension
	}
	if err := node.Validate(options); err != nil {
		return false, err
	}
	indices := make(map[*LTLNode]int)
	var assignIndices func(*LTLNode)
	assignIndices = func(current *LTLNode) {
		if _, exists := indices[current]; exists {
			return
		}
		indices[current] = len(indices)
		if current.Left != nil {
			assignIndices(current.Left)
		}
		if current.Right != nil {
			assignIndices(current.Right)
		}
	}
	assignIndices(node)
	values := make(map[[2]int]bool)
	done := make(map[[2]int]bool)
	var eval func(*LTLNode, int) (bool, error)
	eval = func(current *LTLNode, at int) (bool, error) {
		key := [2]int{indices[current], at}
		if done[key] {
			return values[key], nil
		}
		if options.Stop != nil && options.Stop() {
			return false, ErrCancelled
		}
		var value bool
		var err error
		switch current.Operator {
		case LTLAtom:
			if current.Name == "true" {
				value = true
			} else if current.Name != "false" {
				value = trace[at][current.Name]
			}
		case LTLNot:
			value, err = eval(current.Left, at)
			value = !value
		case LTLAnd, LTLOr, LTLImplies:
			var left, right bool
			left, err = eval(current.Left, at)
			if err == nil {
				right, err = eval(current.Right, at)
			}
			if current.Operator == LTLAnd {
				value = left && right
			} else if current.Operator == LTLOr {
				value = left || right
			} else {
				value = !left || right
			}
		case LTLNext:
			if at+1 < len(trace) {
				value, err = eval(current.Left, at+1)
			}
		case LTLGlobally:
			value = true
			for i := at; i < len(trace) && value && err == nil; i++ {
				value, err = eval(current.Left, i)
			}
		case LTLEventually:
			for i := at; i < len(trace) && !value && err == nil; i++ {
				value, err = eval(current.Left, i)
			}
		case LTLUntil:
			value = false
			for i := at; i < len(trace) && err == nil; i++ {
				var right bool
				right, err = eval(current.Right, i)
				if right || err != nil {
					value = right
					break
				}
				var left bool
				left, err = eval(current.Left, i)
				if !left {
					break
				}
			}
		}
		if err != nil {
			return false, err
		}
		done[key], values[key] = true, value
		return value, nil
	}
	return eval(node, frame)
}

type LTLDatabase struct{ formulas []string }

func NewLTLDatabase(formulas []string) (*LTLDatabase, error) {
	copyOf := make([]string, len(formulas))
	for i, formula := range formulas {
		if strings.IndexByte(formula, 0) >= 0 {
			return nil, ErrDimension
		}
		copyOf[i] = formula
	}
	return &LTLDatabase{formulas: copyOf}, nil
}

func (database *LTLDatabase) Formulas() []string {
	if database == nil {
		return nil
	}
	return append([]string(nil), database.formulas...)
}

func CheckLTLSignalNames(node *LTLNode, names []string) error {
	known := make(map[string]bool, len(names)+2)
	known["true"], known["false"] = true, true
	for _, name := range names {
		if name == "" {
			return ErrDimension
		}
		known[name] = true
	}
	var check func(*LTLNode) error
	check = func(current *LTLNode) error {
		if current.Operator == LTLAtom && !known[current.Name] {
			return fmt.Errorf("unknown LTL signal %q", current.Name)
		}
		if current.Left != nil {
			if err := check(current.Left); err != nil {
				return err
			}
		}
		if current.Right != nil {
			return check(current.Right)
		}
		return nil
	}
	if err := node.Validate(LTLParseOptions{}); err != nil {
		return err
	}
	return check(node)
}

type LTLSignal struct {
	Name string
	Lit  AigLit
}

// BuildLTLCombinational builds only the Boolean subset of a formula, matching
// the source's combinational-only builder without storing graph pointers in the
// syntax tree.
func BuildLTLCombinational(graph *AigManager, node *LTLNode, signals []LTLSignal) (AigLit, error) {
	if graph == nil || !node.IsNonTemporal() {
		return InvalidAigLit, ErrDimension
	}
	lookup := make(map[string]AigLit, len(signals))
	for _, signal := range signals {
		if signal.Name == "" || !graph.validSignal(signal.Lit) {
			return InvalidAigLit, ErrDimension
		}
		lookup[signal.Name] = signal.Lit
	}
	var build func(*LTLNode) (AigLit, error)
	build = func(current *LTLNode) (AigLit, error) {
		switch current.Operator {
		case LTLAtom:
			if current.Name == "true" {
				return AigConst1, nil
			}
			if current.Name == "false" {
				return AigConst0, nil
			}
			literal, ok := lookup[current.Name]
			if !ok {
				return InvalidAigLit, fmt.Errorf("unknown LTL signal %q", current.Name)
			}
			return literal, nil
		case LTLNot:
			literal, err := build(current.Left)
			return literal.Not(), err
		case LTLAnd, LTLOr, LTLImplies:
			left, err := build(current.Left)
			if err != nil {
				return InvalidAigLit, err
			}
			right, err := build(current.Right)
			if err != nil {
				return InvalidAigLit, err
			}
			if current.Operator == LTLAnd {
				return graph.And(left, right)
			}
			if current.Operator == LTLOr {
				return graph.Or(left, right)
			}
			return graph.Or(left.Not(), right)
		default:
			return InvalidAigLit, ErrDimension
		}
	}
	return build(node)
}
