package abc

// Array representation of DSD formulas from opt/dau/dauArray.c.

import (
	"fmt"
	"strings"
)

type DauDSDArrayNode struct {
	Variable, Fanins, Depth, Offset, Data int
	Kind                                  DauDSDNodeKind
	Complement                            bool
}

// DauDSDArray owns a pre-order, sentinel-free DSD node sequence. Accessors
// return copies so callers cannot corrupt structural fanin/depth invariants.
type DauDSDArray struct {
	nodes []DauDSDArrayNode
}

func DauDSDArrayFromFormula(text string, options DauDSDOptions) (*DauDSDArray, error) {
	formula, err := ParseDauDSD(text, options)
	if err != nil {
		return nil, err
	}
	result := &DauDSDArray{nodes: make([]DauDSDArrayNode, 0, len(formula.nodes))}
	var appendNode func(int, int) (int, error)
	appendNode = func(index, depth int) (int, error) {
		node := formula.nodes[index]
		if depth > 31 || len(node.children) > 31 {
			return 0, ErrLimit
		}
		if len(node.bindings) != 0 {
			return 0, fmt.Errorf("DSD array cannot encode mux-local bindings: %w", ErrDimension)
		}
		position := len(result.nodes)
		row := DauDSDArrayNode{Kind: node.kind, Complement: node.complement, Depth: depth, Variable: node.value, Data: 31}
		if node.kind == DauDSDConstant {
			row.Data = 0
			row.Complement = node.value != 0 != node.complement
		}
		result.nodes = append(result.nodes, row)
		minimum := 31
		if node.kind == DauDSDVariable {
			minimum = node.value
			result.nodes[position].Data = minimum
			return minimum, nil
		}
		for _, child := range node.children {
			childMinimum, err := appendNode(child, depth+1)
			if err != nil {
				return 0, err
			}
			if childMinimum < minimum {
				minimum = childMinimum
			}
		}
		result.nodes[position].Fanins = len(node.children)
		if minimum != 31 {
			result.nodes[position].Data = minimum
		}
		return minimum, nil
	}
	if _, err = appendNode(formula.root, 0); err != nil {
		return nil, err
	}
	if err = result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}

func (a *DauDSDArray) Nodes() []DauDSDArrayNode {
	if a == nil {
		return nil
	}
	result := make([]DauDSDArrayNode, len(a.nodes))
	copy(result, a.nodes)
	return result
}

func (a *DauDSDArray) Check() error {
	if a == nil || len(a.nodes) == 0 {
		return ErrDimension
	}
	position := 0
	var visit func(int) error
	visit = func(depth int) error {
		if position >= len(a.nodes) {
			return ErrInvariant
		}
		node := a.nodes[position]
		position++
		if node.Depth != depth || node.Depth > 31 || node.Fanins < 0 || node.Fanins > 31 || node.Variable < 0 || node.Variable > 25 || node.Data < 0 || node.Data > 31 {
			return ErrInvariant
		}
		switch node.Kind {
		case DauDSDConstant, DauDSDVariable:
			if node.Fanins != 0 {
				return ErrInvariant
			}
		case DauDSDAnd, DauDSDXor:
			if node.Fanins < 2 {
				return ErrInvariant
			}
		case DauDSDMux:
			if node.Fanins != 3 {
				return ErrInvariant
			}
		case DauDSDPrime:
		default:
			return ErrInvariant
		}
		for i := 0; i < node.Fanins; i++ {
			if err := visit(depth + 1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(0); err != nil {
		return err
	}
	if position != len(a.nodes) {
		return ErrInvariant
	}
	return nil
}

// AndCount preserves Dau_DsdCountAnd's per-node cost model.
func (a *DauDSDArray) AndCount() (int, error) {
	if err := a.Check(); err != nil {
		return 0, err
	}
	cost := 0
	for _, node := range a.nodes {
		switch node.Kind {
		case DauDSDAnd:
			cost++
		case DauDSDXor, DauDSDMux, DauDSDPrime:
			cost += 3
		}
	}
	return cost, nil
}

// WithDepths returns an owned copy with depths recomputed from fanin structure.
func (a *DauDSDArray) WithDepths() (*DauDSDArray, error) {
	if err := a.Check(); err != nil {
		return nil, err
	}
	result := &DauDSDArray{nodes: a.Nodes()}
	position := 0
	var visit func(int)
	visit = func(depth int) {
		index := position
		position++
		result.nodes[index].Depth = depth
		for i := 0; i < result.nodes[index].Fanins; i++ {
			visit(depth + 1)
		}
	}
	visit(0)
	return result, nil
}

type dauDSDArrayTree struct {
	node     DauDSDArrayNode
	children []*dauDSDArrayTree
}

func (a *DauDSDArray) tree() (*dauDSDArrayTree, error) {
	if err := a.Check(); err != nil {
		return nil, err
	}
	position := 0
	var read func() *dauDSDArrayTree
	read = func() *dauDSDArrayTree {
		node := a.nodes[position]
		position++
		result := &dauDSDArrayTree{node: node, children: make([]*dauDSDArrayTree, node.Fanins)}
		for i := range result.children {
			result.children[i] = read()
		}
		return result
	}
	return read(), nil
}

// WithoutUselessAssociativeNodes preserves Dau_DsdRemoveUseless: nested XORs
// are flattened, as are non-complemented nested ANDs. The receiver is unchanged.
func (a *DauDSDArray) WithoutUselessAssociativeNodes() (*DauDSDArray, error) {
	root, err := a.tree()
	if err != nil {
		return nil, err
	}
	var flatten func(*dauDSDArrayTree)
	flatten = func(node *dauDSDArrayTree) {
		for _, child := range node.children {
			flatten(child)
		}
		children := make([]*dauDSDArrayTree, 0, len(node.children))
		for _, child := range node.children {
			if child.node.Kind == node.node.Kind && (node.node.Kind == DauDSDXor || node.node.Kind == DauDSDAnd && !child.node.Complement) {
				children = append(children, child.children...)
			} else {
				children = append(children, child)
			}
		}
		node.children = children
		node.node.Fanins = len(children)
	}
	flatten(root)
	result := &DauDSDArray{}
	var write func(*dauDSDArrayTree, int) int
	write = func(node *dauDSDArrayTree, depth int) int {
		position := len(result.nodes)
		row := node.node
		row.Depth = depth
		row.Fanins = len(node.children)
		result.nodes = append(result.nodes, row)
		minimum := 31
		if row.Kind == DauDSDVariable {
			minimum = row.Variable
		}
		for _, child := range node.children {
			if value := write(child, depth+1); value < minimum {
				minimum = value
			}
		}
		if minimum != 31 {
			result.nodes[position].Data = minimum
		}
		return minimum
	}
	write(root, 0)
	return result, result.Check()
}

// Format returns Dau_DsdPrint's deterministic diagnostic without stdout.
func (a *DauDSDArray) Format() (string, error) {
	root, err := a.tree()
	if err != nil {
		return "", err
	}
	var output strings.Builder
	var write func(*dauDSDArrayTree)
	write = func(node *dauDSDArrayTree) {
		if node.node.Kind == DauDSDConstant {
			if node.node.Complement {
				output.WriteByte('1')
			} else {
				output.WriteByte('0')
			}
			return
		}
		if node.node.Complement {
			output.WriteByte('!')
		}
		switch node.node.Kind {
		case DauDSDVariable:
			output.WriteByte(byte('a' + node.node.Variable))
		default:
			open, close := byte('('), byte(')')
			if node.node.Kind == DauDSDXor {
				open, close = '[', ']'
			}
			if node.node.Kind == DauDSDMux {
				open, close = '<', '>'
			}
			if node.node.Kind == DauDSDPrime {
				open, close = '{', '}'
			}
			output.WriteByte(open)
			output.WriteByte(byte('a' + node.node.Data))
			fmt.Fprint(&output, node.node.Depth)
			for _, child := range node.children {
				write(child)
			}
			output.WriteByte(close)
		}
	}
	write(root)
	output.WriteByte('\n')
	return output.String(), nil
}
