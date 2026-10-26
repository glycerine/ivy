package abc

// Boolean-function and gate-network enumeration from src/misc/extra/extraUtilEnum.c.

import (
	"fmt"
	"math"
	"strings"
)

type ExtraEnumerationProblem struct {
	Variables      int
	Minterms       []uint32
	VariableTruths []uint32
	Functions      []uint32
}

func newExtraEnumerationProblem(minterms, functions []string) (ExtraEnumerationProblem, error) {
	if len(minterms) == 0 || len(minterms) > 32 || len(functions) == 0 {
		return ExtraEnumerationProblem{}, ErrDimension
	}
	variables := len(minterms[0])
	if variables == 0 || variables > 32 {
		return ExtraEnumerationProblem{}, ErrDimension
	}
	result := ExtraEnumerationProblem{
		Variables:      variables,
		Minterms:       make([]uint32, len(minterms)),
		VariableTruths: make([]uint32, variables),
		Functions:      make([]uint32, len(functions)),
	}
	for minterm, text := range minterms {
		if len(text) != variables {
			return ExtraEnumerationProblem{}, ErrFormat
		}
		for variable, character := range text {
			switch character {
			case '1':
				result.Minterms[minterm] |= 1 << uint(variable)
				result.VariableTruths[variable] |= 1 << uint(minterm)
			case '-':
			default:
				return ExtraEnumerationProblem{}, ErrFormat
			}
		}
	}
	for function, text := range functions {
		if len(text) != len(minterms) {
			return ExtraEnumerationProblem{}, ErrFormat
		}
		for minterm, character := range text {
			switch character {
			case '1':
				result.Functions[function] |= 1 << uint(minterm)
			case '0':
			default:
				return ExtraEnumerationProblem{}, ErrFormat
			}
		}
	}
	return result, nil
}

func ExtraEnumerationProblemFirst() ExtraEnumerationProblem {
	result, _ := newExtraEnumerationProblem([]string{
		"1-1-1-1-", "1-1--11-", "1-1-1--1", "1-1--1-1",
		"-11-1-1-", "-11--11-", "-11-1--1", "-11--1-1",
		"1--11-1-", "1--1-11-", "1--11--1", "1--1-1-1",
		"-1-11-1-", "-1-1-11-", "-1-11--1", "-1-1-1-1",
	}, []string{
		"1111101011111010", "0000010100000101", "1111110010101001", "0000001101010110",
		"1111111111001101", "0000000000110010", "1111111111111110", "0000000000000001",
	})
	return result
}

func ExtraEnumerationProblemSecond() ExtraEnumerationProblem {
	result, _ := newExtraEnumerationProblem([]string{
		"1-1---1---", "1-1----1--", "1-1-----1-", "1-1------1",
		"1--1--1---", "1--1---1--", "1--1----1-", "1--1-----1",
		"1---1-1---", "1---1--1--", "1---1---1-", "1---1----1",
		"1----11---", "1----1-1--", "1----1--1-", "1----1---1",
		"-11---1---", "-11----1--", "-11-----1-", "-11------1",
		"-1-1--1---", "-1-1---1--", "-1-1----1-", "-1-1-----1",
		"-1--1-1---", "-1--1--1--", "-1--1---1-", "-1--1----1",
		"-1---11---", "-1---1-1--", "-1---1--1-", "-1---1---1",
	}, []string{
		"11111110110010001110110010000000", "00000001001101110001001101111111",
		"10000001001001000001001001001000", "01001000000100101000000100100100",
		"00100100100000010100100000010010", "00010010010010000010010010000001",
		"11111111111111111111000000000000",
	})
	return result
}

func ExtraEnumerationProblemThird() ExtraEnumerationProblem {
	result, _ := newExtraEnumerationProblem([]string{
		"1---1---", "1----1--", "1-----1-", "1------1",
		"-1--1---", "-1---1--", "-1----1-", "-1-----1",
		"--1-1---", "--1--1--", "--1---1-", "--1----1",
		"---11---", "---1-1--", "---1--1-", "---1---1",
	}, []string{
		"1111111011001000", "0000000100110111", "1000000100100100", "0100100000010010",
		"0010010010000001", "0001001001001000", "1111111111111111",
	})
	return result
}

type ExtraDecompositionNode struct {
	Truth         uint32
	First, Second int
	Or            bool
	Gates         int
}

type ExtraFunctionDiscovery struct {
	Function, Node, Gates int
	Expression            string
}

type ExtraFunctionEnumerationOptions struct {
	MaxNodes, MaxAttempts int
	Stop                  func() bool
}

type ExtraFunctionEnumeration struct {
	Nodes       []ExtraDecompositionNode
	LayerTotals []int
	Discoveries []ExtraFunctionDiscovery
	Attempts    int
}

func formatExtraDecompositionNode(nodes []ExtraDecompositionNode, node, variables int, builder *strings.Builder) error {
	if node < 0 || node >= len(nodes) {
		return ErrDimension
	}
	if node < variables {
		builder.WriteByte(byte('a' + node))
		return nil
	}
	entry := nodes[node]
	writeFanin := func(fanin int) error {
		if fanin < variables {
			return formatExtraDecompositionNode(nodes, fanin, variables, builder)
		}
		builder.WriteByte('(')
		if e := formatExtraDecompositionNode(nodes, fanin, variables, builder); e != nil {
			return e
		}
		builder.WriteByte(')')
		return nil
	}
	if e := writeFanin(entry.First); e != nil {
		return e
	}
	if entry.Or {
		builder.WriteByte('+')
	}
	if e := writeFanin(entry.Second); e != nil {
		return e
	}
	return nil
}

func FormatExtraDecomposition(nodes []ExtraDecompositionNode, node, variables int) (string, error) {
	var builder strings.Builder
	if e := formatExtraDecompositionNode(nodes, node, variables, &builder); e != nil {
		return "", e
	}
	return builder.String(), nil
}

func EnumerateExtraFunctions(problem ExtraEnumerationProblem, maxGates int, options ExtraFunctionEnumerationOptions) (ExtraFunctionEnumeration, error) {
	fail := func(result ExtraFunctionEnumeration, e error) (ExtraFunctionEnumeration, error) { return result, e }
	if problem.Variables <= 0 || len(problem.VariableTruths) != problem.Variables || len(problem.Minterms) == 0 || len(problem.Minterms) > 32 || maxGates < 0 || options.MaxNodes < 0 || options.MaxAttempts < 0 {
		return fail(ExtraFunctionEnumeration{}, ErrDimension)
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxAttempts == 0 {
		options.MaxAttempts = 1 << 26
	}
	result := ExtraFunctionEnumeration{LayerTotals: []int{0, problem.Variables}}
	seen := make(map[uint32]struct{}, options.MaxNodes)
	for variable, truth := range problem.VariableTruths {
		if _, duplicate := seen[truth]; duplicate {
			return fail(ExtraFunctionEnumeration{}, ErrInvariant)
		}
		seen[truth] = struct{}{}
		result.Nodes = append(result.Nodes, ExtraDecompositionNode{Truth: truth, First: -1, Second: -1})
		_ = variable
	}
	found := make([]bool, len(problem.Functions))
	for gates := 1; gates <= maxGates; gates++ {
		n := gates - 1
		for firstCost := 0; firstCost <= n; firstCost++ {
			for secondCost := firstCost; secondCost <= n; secondCost++ {
				if firstCost+secondCost != n {
					continue
				}
				for first := result.LayerTotals[firstCost]; first < result.LayerTotals[firstCost+1]; first++ {
					for second := result.LayerTotals[secondCost]; second < result.LayerTotals[secondCost+1]; second++ {
						if first >= second {
							continue
						}
						for operation := 0; operation < 2; operation++ {
							if options.Stop != nil && options.Stop() {
								return fail(result, ErrCancelled)
							}
							result.Attempts++
							if result.Attempts > options.MaxAttempts {
								return fail(result, ErrLimit)
							}
							truth := result.Nodes[first].Truth & result.Nodes[second].Truth
							isOr := operation == 1
							if isOr {
								truth = result.Nodes[first].Truth | result.Nodes[second].Truth
							}
							if _, duplicate := seen[truth]; duplicate {
								continue
							}
							if len(result.Nodes) >= options.MaxNodes {
								return fail(result, ErrLimit)
							}
							seen[truth] = struct{}{}
							node := ExtraDecompositionNode{Truth: truth, First: first, Second: second, Or: isOr, Gates: gates}
							result.Nodes = append(result.Nodes, node)
							index := len(result.Nodes) - 1
							for function, target := range problem.Functions {
								if found[function] || truth != target {
									continue
								}
								found[function] = true
								expression, e := FormatExtraDecomposition(result.Nodes, index, problem.Variables)
								if e != nil {
									return fail(result, e)
								}
								result.Discoveries = append(result.Discoveries, ExtraFunctionDiscovery{function, index, gates, expression})
							}
						}
					}
				}
			}
		}
		result.LayerTotals = append(result.LayerTotals, len(result.Nodes))
	}
	return result, nil
}

type ExtraEnumeratedGate struct {
	First, Second                     int
	ComplementFirst, ComplementSecond bool
	XOR                               bool
	Truth                             uint64
	Level                             int
}

type ExtraEnumeratedNetwork struct {
	Variables int
	Gates     []ExtraEnumeratedGate
	Truth     uint64
}

type ExtraNetworkEnumerationOptions struct {
	UseXOR, ReturnNetworks bool
	MaxTries, MaxBuilds    uint64
	MaxFinished            uint64
	Stop                   func() bool
}

type ExtraNetworkEnumeration struct {
	Variables, GateCount int
	Feasible             bool
	Tries, Builds        uint64
	Finished             uint64
	Networks             []ExtraEnumeratedNetwork
}

type extraNetworkEnumerator struct {
	options ExtraNetworkEnumerationOptions
	result  ExtraNetworkEnumeration
	maxNode int
	nodes   int
	tops    int
	fan0    [16]int
	fan1    [16]int
	compl0  [16]bool
	compl1  [16]bool
	polar   [16]int
	refs    [16]int
	level   [16]int
	truth   [16]uint64
	err     error
}

func (e *extraNetworkEnumerator) ref(node int) {
	if e.refs[node] == 0 {
		e.tops--
	}
	e.refs[node]++
}

func (e *extraNetworkEnumerator) deref(node int) {
	e.refs[node]--
	if e.refs[node] == 0 {
		e.tops++
	}
}

func (e *extraNetworkEnumerator) refNode(node int) {
	e.ref(e.fan0[node])
	e.ref(e.fan1[node])
	e.tops++
	e.nodes++
}

func (e *extraNetworkEnumerator) derefNode(node int) {
	e.deref(e.fan0[node])
	e.deref(e.fan1[node])
	e.tops--
	e.nodes--
}

func extraEnumEquivalent(first, second uint64) bool { return first == second || first == ^second }

func (e *extraNetworkEnumerator) filter(node int) bool {
	first, second := e.fan0[node], e.fan1[node]
	truth := e.truth[node]
	if extraEnumEquivalent(truth, 0) {
		return true
	}
	for prior := 0; prior < node; prior++ {
		if extraEnumEquivalent(truth, e.truth[prior]) {
			return true
		}
	}
	for a := 0; a <= first; a++ {
		for b := a + 1; b <= second; b++ {
			if a == first && b == second {
				continue
			}
			x, y := e.truth[a], e.truth[b]
			if extraEnumEquivalent(truth, x&y) || extraEnumEquivalent(truth, x&^y) ||
				extraEnumEquivalent(truth, ^x&y) || extraEnumEquivalent(truth, ^x&^y) ||
				e.options.UseXOR && extraEnumEquivalent(truth, x^y) {
				return true
			}
		}
	}
	return false
}

func (e *extraNetworkEnumerator) capture() {
	if !e.options.ReturnNetworks {
		return
	}
	gates := make([]ExtraEnumeratedGate, 0, e.result.GateCount)
	for node := e.result.Variables; node < e.nodes; node++ {
		gates = append(gates, ExtraEnumeratedGate{
			First: e.fan0[node], Second: e.fan1[node], ComplementFirst: e.compl0[node], ComplementSecond: e.compl1[node],
			XOR: e.polar[node] == 4, Truth: e.truth[node], Level: e.level[node],
		})
	}
	e.result.Networks = append(e.result.Networks, ExtraEnumeratedNetwork{e.result.Variables, gates, e.truth[e.nodes-1]})
}

func (e *extraNetworkEnumerator) enumerate(newLevel bool, firstLastLevel int) {
	if e.err != nil {
		return
	}
	if e.options.Stop != nil && e.options.Stop() {
		e.err = ErrCancelled
		return
	}
	if e.nodes == e.maxNode {
		if e.tops != 1 {
			e.err = ErrInvariant
			return
		}
		if e.result.Finished >= e.options.MaxFinished {
			e.err = ErrLimit
			return
		}
		e.result.Finished++
		e.capture()
		return
	}
	node := e.nodes
	refedFans := e.maxNode - node + 1 - e.tops
	high0, high1, low0, firstPolar := e.fan1[node-1], firstLastLevel, e.fan0[node-1], e.polar[node-1]
	if newLevel {
		high0, high1, low0, firstPolar = firstLastLevel, node, 0, 0
	}
	level := e.level[high0]
	polarLimit := 4
	if e.options.UseXOR {
		polarLimit++
	}
	for second := high0; second < high1 && e.err == nil; second++ {
		if refedFans == 0 && e.refs[second] > 0 {
			continue
		}
		if e.refs[second] > 0 {
			refedFans--
		}
		for first := 0; first < second && e.err == nil; first++ {
			if second == high0 && first < low0 {
				continue
			}
			if refedFans == 0 && e.refs[first] > 0 {
				continue
			}
			if level == 0 && e.refs[first] == 0 && e.refs[second] == 0 && (first+1 != second || first > 0 && e.refs[first-1] == 0) {
				continue
			}
			if e.level[second] == 0 && e.refs[second] == 0 && e.refs[first] != 0 && second > 0 && e.refs[second-1] == 0 {
				continue
			}
			polarStart := 0
			if second == high0 && first == low0 && !newLevel {
				polarStart = firstPolar + 1
			}
			for polar := polarStart; polar < polarLimit && e.err == nil; polar++ {
				if e.level[first] == 0 && e.refs[first] == 0 && polar&1 != 0 || e.level[second] == 0 && e.refs[second] == 0 && polar&2 != 0 {
					continue
				}
				if e.result.Tries >= e.options.MaxTries {
					e.err = ErrLimit
					break
				}
				e.result.Tries++
				e.fan0[node], e.fan1[node] = first, second
				e.compl0[node], e.compl1[node] = polar&1 != 0, polar&2 != 0
				e.polar[node] = polar
				if polar == 4 {
					e.truth[node] = e.truth[first] ^ e.truth[second]
				} else {
					x, y := e.truth[first], e.truth[second]
					if polar&1 != 0 {
						x = ^x
					}
					if polar&2 != 0 {
						y = ^y
					}
					e.truth[node] = x & y
				}
				if e.filter(node) {
					continue
				}
				if e.result.Builds >= e.options.MaxBuilds {
					e.err = ErrLimit
					break
				}
				e.result.Builds++
				e.level[node] = max(e.level[first], e.level[second]) + 1
				e.refNode(node)
				nextFirst := firstLastLevel
				if newLevel {
					nextFirst = node
				}
				e.enumerate(false, nextFirst)
				e.derefNode(node)
			}
		}
		if e.refs[second] > 0 {
			refedFans++
		}
	}
	if !newLevel && e.err == nil {
		e.enumerate(true, firstLastLevel)
	}
}

func EnumerateExtraNetworks(variables, gates int, options ExtraNetworkEnumerationOptions) (ExtraNetworkEnumeration, error) {
	result := ExtraNetworkEnumeration{Variables: variables, GateCount: gates, Feasible: variables <= gates+1}
	if variables < 2 || variables > 6 || gates <= 0 || variables+gates >= 16 {
		return result, ErrDimension
	}
	if !result.Feasible {
		return result, nil
	}
	if options.MaxTries == 0 {
		options.MaxTries = math.MaxUint64
	}
	if options.MaxBuilds == 0 {
		options.MaxBuilds = math.MaxUint64
	}
	if options.MaxFinished == 0 {
		options.MaxFinished = math.MaxUint64
	}
	enumerator := &extraNetworkEnumerator{options: options, result: result, maxNode: variables + gates, nodes: variables, tops: variables}
	patterns := [...]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	copy(enumerator.truth[:], patterns[:variables])
	enumerator.enumerate(true, 0)
	return enumerator.result, enumerator.err
}

func FormatExtraEnumeratedNetwork(network ExtraEnumeratedNetwork) (string, error) {
	if network.Variables < 0 || len(network.Gates)+network.Variables >= 26 {
		return "", ErrDimension
	}
	var builder strings.Builder
	for offset, gate := range network.Gates {
		node := network.Variables + offset
		if gate.First < 0 || gate.Second <= gate.First || gate.Second >= node {
			return "", ErrInvariant
		}
		if offset > 0 {
			builder.WriteString("  ")
		}
		fmt.Fprintf(&builder, "%c=", 'a'+node)
		if gate.XOR {
			fmt.Fprintf(&builder, "%c+%c", 'a'+gate.First, 'a'+gate.Second)
		} else {
			if gate.ComplementFirst {
				builder.WriteByte('!')
			}
			builder.WriteByte(byte('a' + gate.First))
			if gate.ComplementSecond {
				builder.WriteByte('!')
			}
			builder.WriteByte(byte('a' + gate.Second))
		}
	}
	return builder.String(), nil
}
