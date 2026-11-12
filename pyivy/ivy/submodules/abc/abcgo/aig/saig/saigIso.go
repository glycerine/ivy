package saig

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
)

// isoSignatureColors performs deterministic color refinement on the terminal
// and node graph. Primary inputs and latch outputs start in distinct classes;
// each round refines a terminal by its sorted fanout contexts. Ties are kept
// stable by their source terminal IDs.
func isoSignatureColors(p *aig.Graph) []string {
	inputs := p.Inputs()
	outputs := p.Outputs()
	nRegs := aig.Aig_ManRegNum(p)
	nPis := int32(len(inputs)) - nRegs
	nPos := int32(len(outputs)) - nRegs
	colors := make([]string, len(inputs))
	for i := range inputs {
		if int32(i) < nPis {
			colors[i] = "PI"
		} else {
			colors[i] = "LO"
		}
	}
	for round := 0; round <= len(inputs); round++ {
		contexts := make([][]string, len(inputs))
		inputIndex := make(map[aig.Edge]int, len(inputs))
		for i, edge := range inputs {
			inputIndex[edge] = i
			contexts[i] = make([]string, 0)
		}
		edgeColor := isoEdgeColorer(p, colors)
		// Build source-ID-indexed fanout contexts. The node slice position is
		// its stable AIG object ID, including holes left by graph rewriting.
		for _, node := range p.Nodes() {
			if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
				continue
			}
			if node.Kind == aig.Buffer {
				addInputContext(contexts, inputIndex, node.Fanins[0], fmt.Sprintf("B:%s", edgeColor(node.Fanins[0])))
				continue
			}
			left, right := node.Fanins[0], node.Fanins[1]
			leftColor := edgeColor(left)
			rightColor := edgeColor(right)
			kind := "A"
			if node.Kind == aig.ExclusiveOr {
				kind = "X"
			}
			addInputContext(contexts, inputIndex, left, fmt.Sprintf("%s:%s:%s", kind, rightColor, leftColor))
			addInputContext(contexts, inputIndex, right, fmt.Sprintf("%s:%s:%s", kind, leftColor, rightColor))
		}
		for i, output := range outputs {
			fanin := aig.Aig_ObjChild0(p, output)
			kind := "PO"
			if int32(i) >= nPos {
				kind = "LI"
			}
			addInputContext(contexts, inputIndex, fanin, kind+":"+edgeColor(fanin))
		}
		next := make([]string, len(inputs))
		for i := range inputs {
			sort.Strings(contexts[i])
			next[i] = colors[i] + "|" + fmt.Sprint(contexts[i])
		}
		next = compressColorStrings(next)
		if sameStrings(colors, next) {
			break
		}
		colors = next
	}
	return colors
}

func isoEdgeColorer(p *aig.Graph, colors []string) func(aig.Edge) string {
	cache := make(map[aig.Edge]string)
	var signature func(aig.Edge) string
	signature = func(edge aig.Edge) string {
		regular := aig.Aig_Regular(edge)
		base, ok := cache[regular]
		if !ok {
			switch {
			case aig.Aig_ObjIsConst1(p, regular):
				base = "C"
			case aig.Aig_ObjIsCi(p, regular):
				id := int(aig.Aig_ObjCioId(p, regular))
				if id >= 0 && id < len(colors) {
					base = colors[id]
				} else {
					base = "I?"
				}
			default:
				left := signature(aig.Aig_ObjChild0(p, regular))
				if aig.Aig_ObjIsBuf(p, regular) {
					base = "B" + digestISO(left)
				} else {
					right := signature(aig.Aig_ObjChild1(p, regular))
					if right < left {
						left, right = right, left
					}
					kind := "A"
					if aig.Aig_ObjIsExor(p, regular) {
						kind = "X"
					}
					base = kind + digestISO(left+":"+right)
				}
			}
			cache[regular] = base
		}
		if aig.Aig_IsComplement(edge) != 0 {
			return "!" + base
		}
		return base
	}
	return signature
}

func digestISO(value string) string {
	digest := sha256.Sum256([]byte(value))
	return string(digest[:])
}

func addInputContext(contexts [][]string, indices map[aig.Edge]int, edge aig.Edge, value string) {
	regular := aig.Aig_Regular(edge)
	if id, ok := indices[regular]; ok {
		contexts[id] = append(contexts[id], value)
	}
}

func compressColorStrings(colors []string) []string {
	unique := append([]string(nil), colors...)
	sort.Strings(unique)
	unique = compactStrings(unique)
	ranks := make(map[string]string, len(unique))
	for i, value := range unique {
		ranks[value] = strconv.Itoa(i)
	}
	result := make([]string, len(colors))
	for i, value := range colors {
		result[i] = ranks[value]
	}
	return result
}

func compactStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	write := 1
	for read := 1; read < len(values); read++ {
		if values[read] != values[write-1] {
			values[write] = values[read]
			write++
		}
	}
	return values[:write]
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func AigEdgeID(e aig.Edge) int { return e.NodeID() }

// Saig_ManFindIsoPerm derives a stable primary-input/latch-output order from
// repeated local structural refinement. The permutation contains original CI
// indices, with primary inputs first and latch outputs second.
func isoFindInputPermutation(p *aig.Graph, verbose int32) []int32 {
	inputs := p.Inputs()
	nRegs := aig.Aig_ManRegNum(p)
	nPis := len(inputs) - int(nRegs)
	colors := isoSignatureColors(p)
	indices := make([]int, len(inputs))
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(i, j int) bool {
		a, b := indices[i], indices[j]
		if (a < nPis) != (b < nPis) {
			return a < nPis
		}
		if colors[a] != colors[b] {
			return colors[a] < colors[b]
		}
		return a < b
	})
	result := make([]int32, len(indices))
	for i, id := range indices {
		result[i] = int32(id)
	}
	if verbose != 0 {
		fmt.Printf("SAIG input partition: %d primary inputs, %d latch outputs\n", nPis, nRegs)
	}
	return result
}

// Saig_ManFindIsoPermCos orders primary outputs by their driver signatures,
// then carries the latch-output permutation through to latch-input outputs.
func Saig_ManFindIsoPermCos(p *aig.Graph, permCIs []int32) []int32 {
	if len(permCIs) != p.InputCount() {
		panic("Saig_ManFindIsoPermCos: CI permutation has wrong length")
	}
	outputs := p.Outputs()
	nRegs := aig.Aig_ManRegNum(p)
	nPos := len(outputs) - int(nRegs)
	nPis := p.InputCount() - int(nRegs)
	signatures := isoNodeSignatures(p, permCIs)
	poOrder := make([]int, nPos)
	for i := range poOrder {
		poOrder[i] = i
	}
	sort.SliceStable(poOrder, func(i, j int) bool {
		a := aig.Aig_ObjChild0(p, outputs[poOrder[i]])
		b := aig.Aig_ObjChild0(p, outputs[poOrder[j]])
		if signatures[a] != signatures[b] {
			return signatures[a] < signatures[b]
		}
		return poOrder[i] < poOrder[j]
	})
	result := make([]int32, 0, len(outputs))
	for _, id := range poOrder {
		result = append(result, int32(id))
	}
	diff := nPos - nPis
	for _, ci := range permCIs[nPis:] {
		result = append(result, ci+int32(diff))
	}
	return result
}

func isoNodeSignatures(p *aig.Graph, permCIs []int32) map[aig.Edge]string {
	inputs := p.Inputs()
	inputRank := make(map[aig.Edge]int, len(inputs))
	for rank, original := range permCIs {
		if original < 0 || int(original) >= len(inputs) {
			panic("SAIG isomorphism: CI permutation contains invalid index")
		}
		inputRank[inputs[original]] = rank
	}
	cache := make(map[aig.Edge]string)
	var signature func(aig.Edge) string
	signature = func(edge aig.Edge) string {
		regular := aig.Aig_Regular(edge)
		var base string
		if cached, ok := cache[regular]; ok {
			base = cached
		} else if aig.Aig_ObjIsConst1(p, regular) {
			base = "C"
		} else if aig.Aig_ObjIsCi(p, regular) {
			base = fmt.Sprintf("I%08d", inputRank[regular])
		} else if aig.Aig_ObjIsCo(p, regular) {
			base = "O" + digestISO(signature(aig.Aig_ObjChild0(p, regular)))
		} else {
			left := aig.Aig_ObjChild0(p, regular)
			if aig.Aig_ObjIsBuf(p, regular) {
				base = "B" + digestISO(signature(left))
			} else {
				right := aig.Aig_ObjChild1(p, regular)
				leftSig, rightSig := signature(left), signature(right)
				if rightSig < leftSig {
					leftSig, rightSig = rightSig, leftSig
				}
				kind := "A"
				if aig.Aig_ObjIsExor(p, regular) {
					kind = "X"
				}
				base = kind + digestISO(leftSig+":"+rightSig)
			}
			cache[regular] = base
		}
		if aig.Aig_IsComplement(edge) != 0 {
			return "!" + base
		}
		return base
	}
	result := make(map[aig.Edge]string, len(cache)*2+len(inputs)*2)
	for i := range inputs {
		edge := inputs[i]
		result[edge] = signature(edge)
		result[edge.Not()] = signature(edge.Not())
	}
	for id := range p.Nodes() {
		edge := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(p, edge) {
			continue
		}
		result[edge] = signature(edge)
		result[edge.Not()] = signature(edge.Not())
	}
	for _, output := range p.Outputs() {
		edge := aig.Aig_ObjChild0(p, output)
		result[edge] = signature(edge)
	}
	return result
}

// Saig_ManDupIsoCanonical returns a duplicate with refined input and output
// order. The returned graph owns all of its nodes and has no metadata aliases.
func Saig_ManDupIsoCanonical(p *aig.Graph, verbose int32) *aig.Graph {
	permCIs := Saig_ManFindIsoPerm(p, verbose)
	permCos := Saig_ManFindIsoPermCos(p, permCIs)
	return duplicateWithOrder(p, permCIs, permCos)
}

func duplicateWithOrder(p *aig.Graph, permCIs, permCos []int32) *aig.Graph {
	inputs, outputs := p.Inputs(), p.Outputs()
	if len(permCIs) != len(inputs) || len(permCos) > len(outputs) || len(permCos) < int(aig.Aig_ManRegNum(p)) {
		panic("SAIG isomorphism: invalid terminal permutation")
	}
	copyGraph := aig.NewGraph()
	edgeMap := make(map[aig.Edge]aig.Edge, len(inputs)+p.NodeCount())
	for _, oldIndex := range permCIs {
		if oldIndex < 0 || int(oldIndex) >= len(inputs) {
			panic("SAIG isomorphism: invalid CI permutation")
		}
		old := inputs[oldIndex]
		edgeMap[old] = copyGraph.NewInput()
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		regular := aig.Aig_Regular(edge)
		var result aig.Edge
		if aig.Aig_ObjIsConst1(p, regular) {
			result = copyGraph.Const1()
		} else if existing, ok := edgeMap[regular]; ok {
			result = existing
		} else {
			left := copyEdge(aig.Aig_ObjChild0(p, regular))
			if aig.Aig_ObjIsBuf(p, regular) {
				result = left
			} else {
				right := copyEdge(aig.Aig_ObjChild1(p, regular))
				if aig.Aig_ObjIsExor(p, regular) {
					result = copyGraph.XOR(left, right)
				} else {
					result = copyGraph.AND(left, right)
				}
			}
			edgeMap[regular] = result
		}
		if aig.Aig_IsComplement(edge) != 0 {
			return result.Not()
		}
		return result
	}
	for _, oldIndex := range permCos {
		if oldIndex < 0 || int(oldIndex) >= len(outputs) {
			panic("SAIG isomorphism: invalid CO permutation")
		}
		co := outputs[oldIndex]
		copyGraph.AddOutput(copyEdge(aig.Aig_ObjChild0(p, co)))
	}
	aig.Aig_ManSetRegNum(copyGraph, aig.Aig_ManRegNum(p))
	return copyGraph
}

// Iso_ManCheckMapping checks whether the supplied CI map makes every internal
// node and output of pAig2 structurally match pAig1.
func Iso_ManCheckMapping(pAig1, pAig2 *aig.Graph, map2to1 []int32, verbose int32) int32 {
	if len(map2to1) != pAig2.InputCount() || pAig1.InputCount() != pAig2.InputCount() || pAig1.OutputCount() != pAig2.OutputCount() || aig.Aig_ManRegNum(pAig1) != aig.Aig_ManRegNum(pAig2) || aig.Aig_ManNodeNum(pAig1) != aig.Aig_ManNodeNum(pAig2) {
		return 0
	}
	inputs1, inputs2 := pAig1.Inputs(), pAig2.Inputs()
	edgeMap := map[aig.Edge]aig.Edge{pAig2.Const1(): pAig1.Const1()}
	used := make(map[int32]bool, len(map2to1))
	for i, target := range map2to1 {
		if target < 0 || int(target) >= len(inputs1) || used[target] {
			return 0
		}
		used[target] = true
		edgeMap[inputs2[i]] = inputs1[target]
	}
	targetNodes := make(map[isoNodeKey]aig.Edge)
	for id, node := range pAig1.Nodes() {
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer {
			continue
		}
		edge := aig.Edge(id) << 1
		key := makeIsoNodeKey(node.Kind, node.Fanins)
		targetNodes[key] = edge
	}
	var mapEdge func(aig.Edge) (aig.Edge, bool)
	visiting := make(map[aig.Edge]bool)
	inputNodes := pAig2.Nodes()
	mapEdge = func(edge aig.Edge) (aig.Edge, bool) {
		regular := aig.Aig_Regular(edge)
		mapped, ok := edgeMap[regular]
		if !ok {
			if visiting[regular] {
				return 0, false
			}
			visiting[regular] = true
			id := AigEdgeID(regular)
			if id < 0 || id >= len(inputNodes) {
				delete(visiting, regular)
				return 0, false
			}
			node := inputNodes[id]
			left, leftOK := mapEdge(node.Fanins[0])
			if !leftOK {
				delete(visiting, regular)
				return 0, false
			}
			var key isoNodeKey
			if node.Kind == aig.Buffer {
				key = isoNodeKey{kind: node.Kind, left: left}
			} else {
				right, rightOK := mapEdge(node.Fanins[1])
				if !rightOK {
					delete(visiting, regular)
					return 0, false
				}
				key = makeIsoNodeKey(node.Kind, [2]aig.Edge{left, right})
			}
			mapped, ok = targetNodes[key]
			delete(visiting, regular)
			if !ok {
				return 0, false
			}
			edgeMap[regular] = mapped
		}
		if aig.Aig_IsComplement(edge) != 0 {
			mapped = mapped.Not()
		}
		return mapped, true
	}
	outputs1, outputs2 := pAig1.Outputs(), pAig2.Outputs()
	for i := range outputs1 {
		mapped, ok := mapEdge(aig.Aig_ObjChild0(pAig2, outputs2[i]))
		if !ok || mapped != aig.Aig_ObjChild0(pAig1, outputs1[i]) {
			if verbose != 0 {
				fmt.Printf("Structural equivalence failed at output %d.\n", i)
			}
			return 0
		}
	}
	return 1
}

type isoNodeKey struct {
	kind        aig.NodeKind
	left, right aig.Edge
}

func makeIsoNodeKey(kind aig.NodeKind, fanins [2]aig.Edge) isoNodeKey {
	if kind == aig.Buffer {
		return isoNodeKey{kind: kind, left: fanins[0]}
	}
	left, right := fanins[0], fanins[1]
	if right < left {
		left, right = right, left
	}
	return isoNodeKey{kind: kind, left: left, right: right}
}

// Iso_ManNegEdgeNum counts complemented input edges without mutating manager
// caches. Output edges are included, while input terminals have no fanins.
func Iso_ManNegEdgeNum(p *aig.Graph) int32 {
	var count int32
	for id, node := range p.Nodes() {
		if pNode := aig.Edge(id) << 1; aig.Aig_ObjIsNone(p, pNode) {
			continue
		}
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			count += int32(node.Fanins[0] & 1)
			count += int32(node.Fanins[1] & 1)
		case aig.Buffer, aig.Output:
			count += int32(node.Fanins[0] & 1)
		}
	}
	return count
}

// Iso_ManFindMapping returns a CI map from pAig2 to pAig1, or nil if the
// refined candidate does not pass exact structural verification.
func Iso_ManFindMapping(pAig1, pAig2 *aig.Graph, perm1, perm2 []int32, verbose int32) []int32 {
	if pAig1.InputCount() != pAig2.InputCount() || pAig1.OutputCount() != pAig2.OutputCount() || aig.Aig_ManRegNum(pAig1) != aig.Aig_ManRegNum(pAig2) || aig.Aig_ManNodeNum(pAig1) != aig.Aig_ManNodeNum(pAig2) {
		return nil
	}
	if perm1 == nil {
		perm1 = Saig_ManFindIsoPerm(pAig1, verbose)
	}
	if perm2 == nil {
		perm2 = Saig_ManFindIsoPerm(pAig2, verbose)
	}
	if !validPermutation(perm1, pAig1.InputCount()) || !validPermutation(perm2, pAig2.InputCount()) {
		return nil
	}
	inverse2 := make([]int32, len(perm2))
	for canonical, original := range perm2 {
		inverse2[original] = int32(canonical)
	}
	result := make([]int32, len(inverse2))
	for input2, canonical := range inverse2 {
		result[input2] = perm1[canonical]
	}
	if Iso_ManCheckMapping(pAig1, pAig2, result, verbose) == 0 {
		return nil
	}
	return result
}

func validPermutation(permutation []int32, size int) bool {
	if len(permutation) != size {
		return false
	}
	seen := make([]bool, size)
	for _, value := range permutation {
		if value < 0 || int(value) >= size || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

// Iso_ManFilterPos groups primary outputs with identical structural cone
// signatures and returns a duplicate containing the first output of each
// group, followed by every latch input. Groups are candidate classes; clients
// needing semantic equivalence must verify them with SAT.
func Iso_ManFilterPos(p *aig.Graph, verbose int32) (*aig.Graph, [][]int32) {
	return filterPrimaryOutputs(p, verbose)
}

func Iso_ManFilterPos_old(p *aig.Graph, verbose int32) *aig.Graph {
	reduced, _ := filterPrimaryOutputs(p, verbose)
	return reduced
}

func filterPrimaryOutputs(p *aig.Graph, verbose int32) (*aig.Graph, [][]int32) {
	outputs := p.Outputs()
	nPos := len(outputs) - int(aig.Aig_ManRegNum(p))
	inputOrder := Saig_ManFindIsoPerm(p, 0)
	permutation := Saig_ManFindIsoPermCos(p, inputOrder)
	allSigs := isoNodeSignatures(p, inputOrder)
	indices := make([]int, nPos)
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(i, j int) bool {
		a, b := indices[i], indices[j]
		sa := allSigs[aig.Aig_ObjChild0(p, outputs[a])]
		sb := allSigs[aig.Aig_ObjChild0(p, outputs[b])]
		if sa != sb {
			return sa < sb
		}
		return a < b
	})
	classes := make([][]int32, 0)
	for _, output := range indices {
		sig := allSigs[aig.Aig_ObjChild0(p, outputs[output])]
		if len(classes) == 0 || allSigs[aig.Aig_ObjChild0(p, outputs[classes[len(classes)-1][0]])] != sig {
			classes = append(classes, []int32{int32(output)})
		} else {
			classes[len(classes)-1] = append(classes[len(classes)-1], int32(output))
		}
	}
	selected := make([]int32, 0, len(classes)+int(aig.Aig_ManRegNum(p)))
	for _, group := range classes {
		selected = append(selected, group[0])
	}
	for _, co := range permutation[nPos:] {
		selected = append(selected, co)
	}
	if verbose != 0 {
		fmt.Printf("Grouped %d primary outputs into %d structural classes.\n", nPos, len(classes))
	}
	return duplicateWithOrder(p, inputOrder, selected), classes
}

// Saig_ManIsoReduce keeps one primary output per structural class and preserves
// all register-input outputs. The returned classes identify original PO IDs.
func Saig_ManIsoReduce(p *aig.Graph, verbose int32) (*aig.Graph, [][]int32) {
	reduced, classes := Iso_ManFilterPos(p, verbose)
	return reduced, classes
}

func Iso_StoCompareVecStr(a, b []byte) int {
	return strings.Compare(string(a), string(b))
}

func Iso_ManTest(p *aig.Graph, verbose int32) *aig.Graph {
	_ = Saig_ManFindIsoPerm(p, verbose)
	return nil
}

// Iso_ManTest888 was an ABC-network conversion diagnostic; the graph package
// intentionally has no dependency on ABC network mutation APIs. It retains
// the original nil result and leaves the conversion-based diagnostic to the
// higher ABC layer.
func Iso_ManTest888(p *aig.Graph, verbose int32) *aig.Graph {
	_ = p
	_ = verbose
	return nil
}
