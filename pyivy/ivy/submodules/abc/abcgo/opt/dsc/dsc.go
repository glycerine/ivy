// Package dsc implements bottom-up disjoint-support decomposition of Boolean
// truth tables, following Callegaro et al. (ICCD 2015).
package dsc

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
)

// DscNode is a partial decomposition candidate. The truth slices are immutable
// inputs to subsequent candidates; grouping copies cofactors before applying
// additional cube restrictions.
type DscNode struct {
	NegCof, PosCof, BoolDiff []uint64
	On, Off                  []uint32
	Expr                     string
}

func truthWords(nVars int32) int {
	if nVars < 0 || nVars > DSC_MAX_VAR {
		panic("dsc: variable count outside supported range 0..16")
	}
	return int(util.Abc_TtWordNum(nVars))
}

func checkTruth(truth []uint64, nVars int32) int {
	words := truthWords(nVars)
	if len(truth) < words {
		panic(fmt.Sprintf("dsc: truth table has %d words; %d required", len(truth), words))
	}
	return words
}

// Dsc_AllocPool returns reusable scratch storage for decomposition calls.
// Passing nil to Dsc_Decompose is also supported.
func Dsc_AllocPool(nVars int32) []uint64 {
	if nVars < 0 || nVars > DSC_MAX_VAR {
		panic("dsc: pool variable count outside supported range 0..16")
	}
	if nVars == 0 {
		return nil
	}
	return make([]uint64, 3*truthWords(nVars)*int(nVars))
}

// Dsc_FreePool releases the caller's reference to a scratch pool. Go owns the
// storage, so clearing it is unnecessary; the function remains convenient for
// callers that mirror the original allocate/decompose/free flow.
func Dsc_FreePool(pool []uint64) { _ = pool }

// Dsc_Decompose returns a DSD expression and a status. Status is zero when a
// complete decomposition is found, and -1 when the bottom-up procedure cannot
// derive one. The expression uses ! for complement, (...) for AND, and [...] for
// XOR. A caller may reuse a pool from Dsc_AllocPool.
func Dsc_Decompose(truth []uint64, nVars int32, pool []uint64) (string, int32) {
	words := checkTruth(truth, nVars)
	if nVars == 0 {
		if util.Abc_TtIsConst0(truth, int32(words)) != 0 {
			return "0", 0
		}
		if util.Abc_TtIsConst1(truth, int32(words)) != 0 {
			return "1", 0
		}
		return "", -1
	}
	need := 3 * words * int(nVars)
	if pool == nil {
		pool = make([]uint64, need)
	} else if len(pool) < need {
		panic(fmt.Sprintf("dsc: scratch pool has %d words; %d required", len(pool), need))
	}
	pool = pool[:need]

	nodes := make([]DscNode, int(nVars))
	newNodes := make([]*DscNode, 0, int(nVars))
	for variable := int32(0); variable < nVars; variable++ {
		base := int(variable) * 3 * words
		neg, pos, diff := pool[base:base+words], pool[base+words:base+2*words], pool[base+2*words:base+3*words]
		util.Abc_TtCofactor0p(neg, truth, int32(words), variable)
		if util.Abc_TtEqual(neg, truth, int32(words)) != 0 {
			continue
		}
		util.Abc_TtCofactor1p(pos, truth, int32(words), variable)
		util.Abc_TtXor(diff, neg, pos, int32(words), 0)
		node := &nodes[variable]
		node.NegCof, node.PosCof, node.BoolDiff = neg, pos, diff
		node.On = []uint32{1, uint32(variable<<1) | 1}
		node.Off = []uint32{1, uint32(variable << 1)}
		node.Expr = string(rune('a' + variable))
		newNodes = append(newNodes, node)
	}
	if len(newNodes) == 0 {
		if util.Abc_TtIsConst0(truth, int32(words)) != 0 {
			return "0", 0
		}
		if util.Abc_TtIsConst1(truth, int32(words)) != 0 {
			return "1", 0
		}
		return "", -1
	}

	oldNodes := make([]*DscNode, 0, int(nVars))
	freeNodes := make([]DscNode, DSC_MAX_VAR)
	nFree := 0
	for len(newNodes) > 0 {
		nextNodes := make([]*DscNode, 0, len(newNodes))
		for _, ni := range newNodes {
			var grouped *DscNode
			for j, nj := range oldNodes {
				if leftPol, rightPol, ok := dscAndTest(ni, nj); ok {
					grouped = dscAndGroup(allocDscNode(freeNodes, &nFree), ni, leftPol, nj, rightPol, words)
				} else if dscXorTest(ni, nj) {
					grouped = dscXorGroup(allocDscNode(freeNodes, &nFree), ni, nj, words)
				}
				if grouped != nil {
					oldNodes[j] = oldNodes[len(oldNodes)-1]
					oldNodes = oldNodes[:len(oldNodes)-1]
					break
				}
			}
			if grouped != nil {
				nextNodes = append(nextNodes, grouped)
			} else {
				oldNodes = append(oldNodes, ni)
			}
		}
		newNodes = nextNodes
	}
	if len(oldNodes) != 1 {
		return "", -1
	}
	solution := oldNodes[0]
	if util.Abc_TtIsConst0(solution.NegCof, int32(words)) != 0 && util.Abc_TtIsConst1(solution.PosCof, int32(words)) != 0 {
		return solution.Expr, 0
	}
	if util.Abc_TtIsConst1(solution.NegCof, int32(words)) != 0 && util.Abc_TtIsConst0(solution.PosCof, int32(words)) != 0 {
		return "!" + solution.Expr, 0
	}
	return "", -1
}

func allocDscNode(nodes []DscNode, next *int) *DscNode {
	if *next == len(nodes) {
		panic("dsc: internal group-node capacity exceeded")
	}
	result := &nodes[*next]
	*next++
	return result
}

func dscAndTest(a, b *DscNode) (aPolarity, bPolarity int32, ok bool) {
	switch {
	case util.Abc_TtEqual(a.NegCof, b.NegCof, int32(len(a.NegCof))) != 0:
		return 1, 1, true
	case util.Abc_TtEqual(a.NegCof, b.PosCof, int32(len(a.NegCof))) != 0:
		return 1, 0, true
	case util.Abc_TtEqual(a.PosCof, b.NegCof, int32(len(a.NegCof))) != 0:
		return 0, 1, true
	case util.Abc_TtEqual(a.PosCof, b.PosCof, int32(len(a.NegCof))) != 0:
		return 0, 0, true
	default:
		return 0, 0, false
	}
}

func dscXorTest(a, b *DscNode) bool {
	return util.Abc_TtEqual(a.BoolDiff, b.BoolDiff, int32(len(a.BoolDiff))) != 0
}

func dscAndGroup(out, a *DscNode, aPolarity int32, b *DscNode, bPolarity int32, words int) *DscNode {
	out.Expr = dscConcat('(', ')', a.Expr, aPolarity, b.Expr, bPolarity)
	aOff, aOn := a.Off, a.On
	bOff, bOn := b.Off, b.On
	if aPolarity == 0 {
		aOff, aOn = a.On, a.Off
	}
	if bPolarity == 0 {
		bOff, bOn = b.On, b.Off
	}
	if len(aOff) <= len(bOff) {
		out.Off = append(out.Off[:0], aOff...)
	} else {
		out.Off = append(out.Off[:0], bOff...)
	}
	out.NegCof = selectCof(a, aPolarity)
	out.On = append(out.On[:0], aOn[1:]...)
	out.On = append(out.On, bOn[1:]...)
	out.On = append([]uint32{uint32(len(out.On))}, out.On...)
	if len(aOn) >= len(bOn) {
		out.PosCof = append([]uint64(nil), selectCof(a, 1-aPolarity)...)
		cubeCofactor(out.PosCof, bOn, words)
	} else {
		out.PosCof = append([]uint64(nil), selectCof(b, 1-bPolarity)...)
		cubeCofactor(out.PosCof, aOn, words)
	}
	out.BoolDiff = append([]uint64(nil), selectCof(b, bPolarity)...)
	xorInPlace(out.BoolDiff, out.PosCof)
	return out
}

func selectCof(node *DscNode, polarity int32) []uint64 {
	if polarity != 0 {
		return node.NegCof
	}
	return node.PosCof
}

func dscXorGroup(out, a, b *DscNode, words int) *DscNode {
	aOff, aOn, bOff, bOn := a.Off, a.On, b.Off, b.On
	aOffN, aOnN, bOffN, bOnN := len(aOff)-1, len(aOn)-1, len(bOff)-1, len(bOn)-1
	minSize, minPolarity, minNode := aOffN, int32(0), a
	if minSize > aOnN {
		minSize, minPolarity = aOnN, 1
	}
	if minSize > bOffN {
		minSize, minPolarity, minNode = bOffN, 0, b
	}
	if minSize > bOnN {
		minSize, minPolarity, minNode = bOnN, 1, b
	}
	out.Expr = dscConcat('[', ']', a.Expr, 1, b.Expr, 1)
	var offCube, onCube []uint32
	if minNode == a {
		if minPolarity != 0 {
			out.NegCof = append([]uint64(nil), b.PosCof...)
			cubeCofactor(out.NegCof, aOn, words)
			out.PosCof = append([]uint64(nil), b.NegCof...)
			cubeCofactor(out.PosCof, aOn, words)
		} else {
			out.NegCof = append([]uint64(nil), b.NegCof...)
			cubeCofactor(out.NegCof, aOff, words)
			out.PosCof = append([]uint64(nil), b.PosCof...)
			cubeCofactor(out.PosCof, aOff, words)
		}
	} else if minPolarity != 0 {
		out.NegCof = append([]uint64(nil), a.PosCof...)
		cubeCofactor(out.NegCof, bOn, words)
		out.PosCof = append([]uint64(nil), a.NegCof...)
		cubeCofactor(out.PosCof, bOn, words)
	} else {
		out.NegCof = append([]uint64(nil), a.NegCof...)
		cubeCofactor(out.NegCof, bOff, words)
		out.PosCof = append([]uint64(nil), a.PosCof...)
		cubeCofactor(out.PosCof, bOff, words)
	}
	out.BoolDiff = append([]uint64(nil), a.BoolDiff...)
	if aOffN+bOffN <= aOnN+bOnN {
		offCube = append(append([]uint32(nil), aOff[1:]...), bOff[1:]...)
	} else {
		offCube = append(append([]uint32(nil), aOn[1:]...), bOn[1:]...)
	}
	if aOffN+bOnN <= aOnN+bOffN {
		onCube = append(append([]uint32(nil), aOff[1:]...), bOn[1:]...)
	} else {
		onCube = append(append([]uint32(nil), aOn[1:]...), bOff[1:]...)
	}
	out.Off = append([]uint32{uint32(len(offCube))}, offCube...)
	out.On = append([]uint32{uint32(len(onCube))}, onCube...)
	return out
}

func dscConcat(begin, end byte, a string, aPolarity int32, b string, bPolarity int32) string {
	result := make([]byte, 0, len(a)+len(b)+4)
	result = append(result, begin)
	if aPolarity == 0 {
		result = append(result, '!')
	}
	result = append(result, a...)
	if bPolarity == 0 {
		result = append(result, '!')
	}
	result = append(result, b...)
	result = append(result, end)
	return string(result)
}

func cubeCofactor(truth []uint64, cube []uint32, words int) {
	for _, literal := range cube[1:] {
		if literal&1 != 0 {
			util.Abc_TtCofactor1(truth, int32(words), int32(literal>>1))
		} else {
			util.Abc_TtCofactor0(truth, int32(words), int32(literal>>1))
		}
	}
}

func xorInPlace(out, other []uint64) {
	for i := range out {
		out[i] ^= other[i]
	}
}

// Dsc_ComputeMatches returns the matching byte index for every opening
// delimiter. Non-openers map to zero. It rejects malformed expressions.
func Dsc_ComputeMatches(expression string) []int {
	matches := make([]int, len(expression))
	stack := make([]int, 0, DSC_MAX_VAR)
	for i := range expression {
		switch expression[i] {
		case '(', '[', '<', '{':
			stack = append(stack, i)
		case ')', ']', '>', '}':
			if len(stack) == 0 {
				panic(fmt.Sprintf("dsc: unmatched closing delimiter at byte %d", i))
			}
			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !dscDelimiterPair(expression[open], expression[i]) {
				panic(fmt.Sprintf("dsc: mismatched delimiter at byte %d", i))
			}
			matches[open] = i
		}
	}
	if len(stack) > 0 {
		panic(fmt.Sprintf("dsc: unclosed delimiter at byte %d", stack[len(stack)-1]))
	}
	return matches
}

func dscDelimiterPair(open, close byte) bool {
	switch open {
	case '(':
		return close == ')'
	case '[':
		return close == ']'
	case '<':
		return close == '>'
	case '{':
		return close == '}'
	default:
		return false
	}
}

// Dsc_CountAnds counts two-input AND-equivalent nodes in a DSD expression.
func Dsc_CountAnds(expression string) int32 {
	if len(expression) <= 1 {
		return 0
	}
	matches := Dsc_ComputeMatches(expression)
	position := 0
	return dscCountAndsNode(expression, &position, matches)
}

func dscCountAndsNode(expression string, position *int, matches []int) int32 {
	if *position >= len(expression) {
		panic("dsc: unexpected end of expression")
	}
	if expression[*position] == '!' {
		*position++
	}
	for *position < len(expression) && ((expression[*position] >= 'A' && expression[*position] <= 'F') || (expression[*position] >= '0' && expression[*position] <= '9')) {
		*position++
	}
	if *position >= len(expression) {
		panic("dsc: expression prefix has no operand")
	}
	ch := expression[*position]
	if ch >= 'a' && ch <= 'z' {
		*position++
		return 0
	}
	if ch != '(' && ch != '[' && ch != '<' && ch != '{' {
		panic(fmt.Sprintf("dsc: unexpected token %q at byte %d", ch, *position))
	}
	open := *position
	close := matches[open]
	if close <= open {
		panic(fmt.Sprintf("dsc: missing match at byte %d", open))
	}
	if ch == '<' && close+1 < len(expression) && expression[close+1] == '{' {
		*position = close + 1
		return dscCountAndsNode(expression, position, matches)
	}
	*position++
	var total int32
	children := 0
	if ch == '(' || ch == '[' {
		addon := int32(1)
		if ch == '[' {
			addon = 3
		}
		for *position < close {
			total += addon + dscCountAndsNode(expression, position, matches)
			children++
		}
		if *position != close || children < 2 {
			panic(fmt.Sprintf("dsc: malformed operator at byte %d", open))
		}
		*position++
		return total - addon
	}
	total = 3
	for *position < close {
		total += dscCountAndsNode(expression, position, matches)
		children++
	}
	if *position != close || children < 2 {
		panic(fmt.Sprintf("dsc: malformed mux at byte %d", open))
	}
	*position++
	return total
}
