package live

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"strings"
	"sync"
)

type tokenType int32

const (
	AND tokenType = iota
	OR
	NOT
	IMPLY
	GLOBALLY
	EVENTUALLY
	NEXT
	UNTIL
	BOOL
)

type LTLToken = tokenType
type ltlNode struct {
	Type        tokenType
	Name        string
	Object      aig.Edge
	HasObject   bool
	Left, Right *ltlNode
}
type LTLNode = ltlNode

func generateTypedNode(kind tokenType) *ltlNode { return &ltlNode{Type: kind, Object: aig.InvalidEdge} }

// ParseFormula reads the source prefix notation: &,+,! and :G,:F,:X,:U.
// Unlike the C global cursor, every invocation owns its parsing position.
func ParseFormula(text string) (*LTLNode, error) { return parseFormula(text, true, true) }

// Source tokenization does not recognize '>' as an operator, and C returns
// after the first expression rather than rejecting trailing input.
func parseFormula(text string, implication, strict bool) (*LTLNode, error) {
	node, _, err := parseFormulaAt(text, 0, implication, strict)
	return node, err
}
func parseFormulaAt(text string, start int, implication, strict bool) (*LTLNode, int, error) {
	position := start
	var parse func() (*ltlNode, error)
	skip := func() {
		for position < len(text) && strings.ContainsRune(" \t\n\r\v\f", rune(text[position])) {
			position++
		}
	}
	parse = func() (*ltlNode, error) {
		skip()
		if position == len(text) {
			return nil, fmt.Errorf("live: missing operand at byte %d", position)
		}
		kind := BOOL
		c := text[position]
		position++
		switch c {
		case '&':
			kind = AND
		case '+':
			kind = OR
		case '!':
			kind = NOT
		case '>':
			if implication {
				kind = IMPLY
			}
		case ':':
			if position == len(text) {
				return nil, fmt.Errorf("live: missing temporal operator")
			}
			switch text[position] {
			case 'G':
				kind = GLOBALLY
			case 'F':
				kind = EVENTUALLY
			case 'X':
				kind = NEXT
			case 'U':
				kind = UNTIL
			default:
				return nil, fmt.Errorf("live: unknown temporal operator at byte %d", position)
			}
			position++
		}
		node := generateTypedNode(kind)
		if kind == BOOL {
			start := position - 1
			// getVarName in ltl_parser.c stops at exactly these four
			// delimiters; CR, vertical tab and form feed inside a name are
			// retained even though the parser skips them between tokens.
			for position < len(text) && !strings.ContainsRune(" \t\n:", rune(text[position])) {
				position++
			}
			node.Name = text[start:position]
			return node, nil
		}
		var err error
		node.Left, err = parse()
		if err != nil {
			return nil, err
		}
		if kind == AND || kind == OR || kind == IMPLY || kind == UNTIL {
			node.Right, err = parse()
			if err != nil {
				return nil, err
			}
		}
		return node, nil
	}
	root, err := parse()
	if err != nil {
		return nil, position, err
	}
	if strict {
		skip()
	}
	if strict && position != len(text) {
		return nil, position, fmt.Errorf("live: trailing formula input at byte %d", position)
	}
	return root, position, nil
}

// Source readLtlFormula uses this process-wide cursor; the exported native
// parser remains reentrant. A lock protects the Go representation from races.
var sourceParserMu sync.Mutex
var startOfSuffixString int

func resetGlobalVar() { sourceParserMu.Lock(); startOfSuffixString = 0; sourceParserMu.Unlock() }
func readLtlFormula(text string) *ltlNode {
	sourceParserMu.Lock()
	defer sourceParserMu.Unlock()
	if startOfSuffixString >= len(text) {
		return nil
	}
	node, end, err := parseFormulaAt(text, startOfSuffixString, false, false)
	startOfSuffixString = end
	if err != nil {
		return nil
	}
	return node
}
func parseFormulaCreateAST(text string) *ltlNode {
	sourceParserMu.Lock()
	defer sourceParserMu.Unlock()
	defer func() { startOfSuffixString = 0 }()
	if startOfSuffixString >= len(text) {
		return nil
	}
	node, _, err := parseFormulaAt(text, startOfSuffixString, false, false)
	if err != nil {
		return nil
	}
	return node
}
func checkBooleanConstant(name string) int32 {
	if name == "true" {
		return 1
	}
	if name == "false" {
		return 0
	}
	return -1
}
func isNonTemporalSubformula(node *ltlNode) int32 {
	if node == nil {
		return 0
	}
	switch node.Type {
	case BOOL:
		return 1
	case NOT:
		return isNonTemporalSubformula(node.Left)
	case AND, OR, IMPLY:
		if isNonTemporalSubformula(node.Left) != 0 && isNonTemporalSubformula(node.Right) != 0 {
			return 1
		}
	}
	return 0
}
func isWellFormed(node *ltlNode) int32 {
	if node == nil {
		return 0
	}
	switch node.Type {
	case BOOL:
		return 1
	case NOT:
		return isWellFormed(node.Left)
	case AND, OR, IMPLY:
		if isWellFormed(node.Left) != 0 && isWellFormed(node.Right) != 0 {
			return 1
		}
	case GLOBALLY:
		if node.Left != nil && node.Left.Type == EVENTUALLY {
			return isNonTemporalSubformula(node.Left.Left)
		}
	}
	return 0
}
func walkLTL(node *ltlNode, visit func(*ltlNode)) {
	if node == nil {
		return
	}
	visit(node)
	walkLTL(node.Left, visit)
	walkLTL(node.Right, visit)
}
func checkSignalNameExistence(network *Network, node *ltlNode) int32 {
	valid := true
	walkLTL(node, func(n *ltlNode) {
		if n.Type == BOOL && checkBooleanConstant(n.Name) < 0 {
			found := false
			for i := 0; i < trueOutputs(network.Graph); i++ {
				found = found || outputName(network, i) == n.Name
			}
			valid = valid && found
		}
	})
	if valid {
		return 1
	}
	return 0
}
func populateBoolWithAigNodePtr(network *Network, old, new *aig.Graph, node *ltlNode) {
	walkLTL(node, func(n *ltlNode) {
		if n.Type != BOOL {
			return
		}
		constant := checkBooleanConstant(n.Name)
		if constant >= 0 {
			n.Object = aig.Edge(1 - constant)
			n.HasObject = true
			return
		}
		for i := 0; i < trueOutputs(old); i++ {
			if outputName(network, i) == n.Name {
				driver := outputDriver(old, i)
				if old == new {
					n.Object = driver
				} else {
					copy := aig.Aig_ObjCopy(old, aig.Aig_Regular(driver))
					if copy == aig.InvalidEdge {
						panic("live: signal has no copied image")
					}
					n.Object = copy ^ aig.Edge(driver&1)
				}
				n.HasObject = true
				return
			}
		}
		panic("live: unknown formula signal " + n.Name)
	})
}
func checkAllBoolHaveAIGPointer(node *ltlNode) int32 {
	valid := true
	walkLTL(node, func(n *ltlNode) {
		if n.Type == BOOL && !n.HasObject {
			valid = false
		}
	})
	if valid {
		return 1
	}
	return 0
}
func buildLogicFromLTLNode(p *aig.Graph, node *ltlNode) aig.Edge {
	if node == nil {
		panic("live: nil formula node")
	}
	if node.HasObject {
		return node.Object
	}
	a := buildLogicFromLTLNode(p, node.Left)
	var result aig.Edge
	switch node.Type {
	case NOT:
		result = a.Not()
	case AND:
		result = aig.Aig_And(p, a, buildLogicFromLTLNode(p, node.Right))
	case OR:
		result = aig.Aig_Or(p, a, buildLogicFromLTLNode(p, node.Right))
	case IMPLY:
		result = aig.Aig_Or(p, a.Not(), buildLogicFromLTLNode(p, node.Right))
	default:
		panic("live: temporal or Boolean formula node has no assigned signal")
	}
	node.Object, node.HasObject = result, true
	return result
}
func buildLogicFromLTLNode_combinationalOnly(p *aig.Graph, node *ltlNode) aig.Edge {
	if isNonTemporalSubformula(node) == 0 {
		panic("live: noncombinational GF operand")
	}
	return buildLogicFromLTLNode(p, node)
}
func populateAigPointerUnitGF(p *aig.Graph, node *ltlNode, signals *[]aig.Edge, groups *[][]*ltlNode) {
	if node == nil {
		return
	}
	if node.Type == GLOBALLY {
		if node.Left == nil || node.Left.Type != EVENTUALLY {
			panic("live: expected GF property")
		}
		signal := buildLogicFromLTLNode_combinationalOnly(p, node.Left.Left)
		index := -1
		for i, e := range *signals {
			if e == signal {
				index = i
				break
			}
		}
		if index < 0 {
			index = len(*signals)
			*signals = append(*signals, signal)
			*groups = append(*groups, nil)
		}
		(*groups)[index] = append((*groups)[index], node)
		return
	}
	populateAigPointerUnitGF(p, node.Left, signals, groups)
	populateAigPointerUnitGF(p, node.Right, signals, groups)
}
func setAIGNodePtrOfGloballyNode(node *ltlNode, edge aig.Edge) {
	node.Object, node.HasObject = edge, true
}
func retriveAIGPointerFromLTLNode(node *ltlNode) aig.Edge {
	if !node.HasObject {
		panic("live: formula has no circuit")
	}
	return node.Object
}
func (node *ltlNode) String() string {
	if node == nil {
		return ""
	}
	if node.Type == BOOL {
		return node.Name
	}
	operator := map[tokenType]string{AND: "&", OR: "+", NOT: "!", IMPLY: ">", GLOBALLY: ":G", EVENTUALLY: ":F", NEXT: ":X", UNTIL: ":U"}[node.Type]
	parts := []string{operator, node.Left.String()}
	if node.Right != nil {
		parts = append(parts, node.Right.String())
	}
	return strings.Join(parts, " ")
}
func traverseAbstractSyntaxTree(node *ltlNode) { fmt.Print(node.String(), " ") }
func traverseAbstractSyntaxTree_postFix(node *ltlNode) {
	if node == nil {
		return
	}
	traverseAbstractSyntaxTree_postFix(node.Left)
	traverseAbstractSyntaxTree_postFix(node.Right)
	if node.Type == BOOL {
		fmt.Print(node.Name, " ")
	} else {
		fmt.Print(map[tokenType]string{AND: "&", OR: "+", NOT: "!", IMPLY: ">", GLOBALLY: "G", EVENTUALLY: "F", NEXT: "X", UNTIL: "U"}[node.Type], " ")
	}
}
func Abc_FrameCopyLTLDataBase(frame *Frame, network *Network) {
	frame.LTLProperties = make([]string, len(network.LTLProperties))
	copy(frame.LTLProperties, network.LTLProperties)
}
