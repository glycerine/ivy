package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"sync/atomic"
)

var dsmCalls, dsmNonDSD, dsmNonOneStep atomic.Int64

// DSDGraphOptions selects the source's balanced/folded DSD construction paths.
// Native XORs use Graph's explicit XOR node; MUXes retain its exact AIG form.
type DSDGraphOptions struct{ Balance, NativeXOR bool }
type dsdKitGraph struct{ graph *aig.Graph }

func dsdToEdge(lit int32) aig.Edge {
	if lit < 2 {
		return aig.Edge(lit ^ 1)
	}
	return aig.Edge(lit)
}
func dsdFromEdge(edge aig.Edge) int32 {
	if edge < 2 {
		return int32(edge ^ 1)
	}
	return int32(edge)
}
func (p dsdKitGraph) Const1() int32              { return 1 }
func (p dsdKitGraph) InputLiteral(i int32) int32 { return dsdFromEdge(p.graph.Inputs()[i]) }
func (p dsdKitGraph) AndLiteral(a, b int32, hash bool) int32 {
	return dsdFromEdge(p.graph.AND(dsdToEdge(a), dsdToEdge(b)))
}
func Dau_DsdAddToArray(p *aig.Graph, fans []aig.Edge, n int32, fan aig.Edge) {
	fans[n] = fan
	for i := n; i > 0; i-- {
		if p.Node(fans[i]).Level <= p.Node(fans[i-1]).Level {
			return
		}
		fans[i], fans[i-1] = fans[i-1], fans[i]
	}
}
func dsdBalance(p *aig.Graph, fans []aig.Edge, and, nativeXOR bool) aig.Edge {
	if len(fans) == 0 {
		panic("dau: empty balanced DSD operator")
	}
	for n := len(fans); n > 1; {
		a, b := fans[n-1], fans[n-2]
		n -= 2
		var fan aig.Edge
		if and {
			fan = p.AND(a, b)
		} else if nativeXOR {
			fan = p.NewXor(a, b)
		} else {
			fan = p.XOR(a, b)
		}
		Dau_DsdAddToArray(p, fans, int32(n), fan)
		n++
	}
	return fans[0]
}
func Dau_DsdBalance(p *aig.Graph, fans []aig.Edge, n, and int32) aig.Edge {
	return dsdBalance(p, fans[:n], and != 0, false)
}
func Dau_DsdToGiaCompose_rec(p *aig.Graph, function uint64, fanins []aig.Edge, n int32) aig.Edge {
	if function == 0 {
		return p.Const0()
	}
	if function == ^uint64(0) {
		return p.Const1()
	}
	if n <= 0 || n > 6 {
		panic("dau: invalid Shannon composition width")
	}
	n--
	if n == 0 {
		if function != npnVariables[0] && function != ^npnVariables[0] {
			panic("dau: invalid one-variable truth")
		}
		result := fanins[0]
		if function == ^npnVariables[0] {
			result = result.Not()
		}
		return result
	}
	if util.Abc_Tt6HasVar(function, n) == 0 {
		return Dau_DsdToGiaCompose_rec(p, function, fanins, n)
	}
	zero := Dau_DsdToGiaCompose_rec(p, util.Abc_Tt6Cofactor0(function, n), fanins, n)
	one := Dau_DsdToGiaCompose_rec(p, util.Abc_Tt6Cofactor1(function, n), fanins, n)
	return p.Mux(fanins[n], one, zero)
}
func dsdExpressionGraph(p *aig.Graph, e *dsExpr, leaves []aig.Edge, cover *vec.Vec_Int_t, options DSDGraphOptions) aig.Edge {
	var result aig.Edge
	switch e.kind {
	case 'v':
		if e.variable >= len(leaves) || e.variable >= 12 {
			panic("dau: DSD variable outside leaf range")
		}
		result = leaves[e.variable]
	case '0':
		result = p.Const0()
	case '1':
		result = p.Const1()
	case '@':
		arguments := make([]aig.Edge, len(e.children))
		for i, child := range e.children {
			arguments[i] = dsdExpressionGraph(p, child, leaves, cover, options)
		}
		result = dsdExpressionGraph(p, e.function, arguments, cover, options)
	case '(', '[':
		if len(e.children) > 12 {
			panic("dau: DSD fanin count exceeds source limit")
		}
		fans := make([]aig.Edge, len(e.children))
		for i, child := range e.children {
			fan := dsdExpressionGraph(p, child, leaves, cover, options)
			if options.Balance {
				Dau_DsdAddToArray(p, fans, int32(i), fan)
			} else {
				fans[i] = fan
			}
		}
		if options.Balance {
			result = dsdBalance(p, fans, e.kind == '(', options.NativeXOR)
		} else {
			result = p.Const0()
			if e.kind == '(' {
				result = p.Const1()
			}
			for _, fan := range fans {
				if e.kind == '(' {
					result = p.AND(result, fan)
				} else if options.NativeXOR {
					result = p.NewXor(result, fan)
				} else {
					result = p.XOR(result, fan)
				}
			}
		}
	case '<':
		if len(e.children) != 3 {
			panic("dau: MUX needs three fanins")
		}
		var fans [3]aig.Edge
		for i, child := range e.children {
			fans[i] = dsdExpressionGraph(p, child, leaves, cover, options)
		}
		result = p.Mux(fans[0], fans[1], fans[2])
	case '{':
		n := len(e.children)
		if n > 12 {
			panic("dau: prime exceeds twelve variables")
		}
		truth := make([]uint64, 1<<max(0, n-6))
		readVars := util.Abc_TtReadHex(truth, e.hex)
		if int(readVars) != n {
			panic("dau: prime truth/argument width mismatch")
		}
		arguments := make([]int32, n)
		for i, child := range e.children {
			arguments[i] = dsdFromEdge(dsdExpressionGraph(p, child, leaves, cover, options))
		}
		words := make([]uint32, 1<<max(0, n-5))
		for i := range words {
			words[i] = uint32(truth[i/2] >> uint(32*(i%2)))
		}
		var scratch []uint32
		if cover != nil {
			for _, entry := range vec.Vec_IntArray(cover) {
				scratch = append(scratch, uint32(entry))
			}
		}
		result = dsdToEdge(kit.Kit_TruthToGia(dsdKitGraph{p}, words, int32(n), &scratch, arguments, 1))
		if cover != nil {
			vec.Vec_IntClear(cover)
			for _, entry := range scratch {
				vec.Vec_IntPush(cover, int32(entry))
			}
		}
		dsmNonOneStep.Add(1)
	default:
		panic("dau: invalid DSD graph expression")
	}
	if e.neg {
		return result.Not()
	}
	return result
}
func Dau_DsdToGiaWithOptions(p *aig.Graph, text string, leaves []aig.Edge, cover *vec.Vec_Int_t, options DSDGraphOptions) aig.Edge {
	return dsdExpressionGraph(p, dsParse(text), leaves, cover, options)
}
func Dau_DsdToGia(p *aig.Graph, text string, leaves []aig.Edge, cover *vec.Vec_Int_t) aig.Edge {
	return Dau_DsdToGiaWithOptions(p, text, leaves, cover, DSDGraphOptions{Balance: true})
}
func Dau_DsdToGia2(p *aig.Graph, text string, leaves []aig.Edge, cover *vec.Vec_Int_t) aig.Edge {
	return Dau_DsdToGiaWithOptions(p, text, leaves, cover, DSDGraphOptions{})
}
func Dsm_ManTruthToGia(p *aig.Graph, truth []uint64, leaves []aig.Edge, cover *vec.Vec_Int_t) aig.Edge {
	return dsmTruthToGraph(p, truth, leaves, cover, DSDGraphOptions{Balance: true})
}
func dsmTruthToGraph(p *aig.Graph, truth []uint64, leaves []aig.Edge, cover *vec.Vec_Int_t, options DSDGraphOptions) aig.Edge {
	dsmCalls.Add(1)
	if len(leaves) > 12 {
		panic("dau: LUT exceeds twelve leaves")
	}
	if len(truth) < 1<<max(0, len(leaves)-6) {
		panic("dau: short LUT truth")
	}
	if len(leaves) == 0 {
		if truth[0]&1 != 0 {
			return p.Const1()
		}
		return p.Const0()
	}
	if len(leaves) == 1 {
		if truth[0]&1 != 0 {
			return leaves[0].Not()
		}
		return leaves[0]
	}
	copyTruth := append([]uint64(nil), truth...)
	var expression []byte
	if Dau_DsdDecompose(copyTruth, int32(len(leaves)), 0, 1, &expression) != 0 {
		dsmNonDSD.Add(1)
	}
	return Dau_DsdToGiaWithOptions(p, string(expression), leaves, cover, options)
}
func Dsm_ManReportStats() {
	fmt.Printf("Calls = %d. NonDSD = %d. Non1Step = %d.\n", dsmCalls.Swap(0), dsmNonDSD.Swap(0), dsmNonOneStep.Swap(0))
}

// DSDMappedNetwork is the live MF mapping contract; no duplicate network is
// stored. LUT truths include the selected cut's output phase.
type DSDMappedNetwork interface {
	mf.Network
	LUT(int32) ([]int32, []uint64, bool)
}

func Dsm_ManDeriveGia(network DSDMappedNetwork, useMuxes int32) *aig.Graph {
	if network == nil {
		panic("dau: nil mapped network")
	}
	p := aig.NewGraph()
	if meta, ok := network.(interface {
		NetworkName() string
		NetworkSpec() string
	}); ok {
		p.Name, p.Spec = meta.NetworkName(), meta.NetworkSpec()
	}
	values := make([]aig.Edge, network.ObjectCount())
	available := make([]bool, len(values))
	if len(values) == 0 {
		panic("dau: empty mapped network")
	}
	values[0] = p.Const0()
	available[0] = true
	for _, input := range network.InputLiterals() {
		id := input >> 1
		if id <= 0 || int(id) >= len(values) {
			panic("dau: invalid mapped input")
		}
		values[id] = p.NewInput()
		available[id] = true
	}
	copyLiteral := func(lit int32) aig.Edge {
		if lit < 0 || int(lit>>1) >= len(values) || !available[lit>>1] {
			panic("dau: mapped node used before derivation")
		}
		edge := values[lit>>1]
		if lit&1 != 0 {
			edge = edge.Not()
		}
		return edge
	}
	cover := vec.Vec_IntAlloc(1 << 16)
	for id := int32(1); int(id) < len(values); id++ {
		a, _, _, kind := network.Gate(id)
		if kind == 5 {
			driver := copyLiteral(a)
			values[id] = aig.Aig_ObjCreate(p, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{driver, aig.InvalidEdge}, Level: p.Node(driver).Level})
			available[id] = true
			continue
		}
		leaves, truth, ok := network.LUT(id)
		if !ok {
			continue
		}
		fanins := make([]aig.Edge, len(leaves))
		for i, leaf := range leaves {
			fanins[i] = copyLiteral(leaf << 1)
		}
		values[id] = dsmTruthToGraph(p, truth, fanins, cover, DSDGraphOptions{Balance: true, NativeXOR: useMuxes != 0})
		available[id] = true
	}
	for _, driver := range network.OutputLiterals() {
		p.AddOutput(copyLiteral(driver))
	}
	if sequential, ok := network.(interface{ RegisterCount() int }); ok {
		aig.Aig_ManSetRegNum(p, int32(sequential.RegisterCount()))
	}
	aig.Aig_ManCleanup(p)
	return p
}

func dsdGraphExpressionEnd(text string, start int, matches []int32) int {
	position := start
	if position < len(text) && text[position] == '!' {
		position++
	}
	if position >= len(text) {
		panic("dau: missing graph expression")
	}
	c := text[position]
	if c >= 'a' && c <= 'z' {
		return position + 1
	}
	if (c == '0' || c == '1') && position+1 == len(text) {
		return position + 1
	}
	for position < len(text) && ((text[position] >= '0' && text[position] <= '9') || (text[position] >= 'A' && text[position] <= 'F')) {
		position++
	}
	if position >= len(matches) || matches[position] <= int32(position) {
		panic("dau: missing graph expression delimiter")
	}
	end := int(matches[position]) + 1
	if end < len(text) && text[end] == '{' {
		if end >= len(matches) || matches[end] <= int32(end) {
			panic("dau: missing composition delimiter")
		}
		end = int(matches[end]) + 1
	}
	return end
}
func Dau_DsdToGia_rec(p *aig.Graph, text string, position *int32, matches []int32, leaves []aig.Edge, cover *vec.Vec_Int_t) aig.Edge {
	end := dsdGraphExpressionEnd(text, int(*position), matches)
	result := Dau_DsdToGia(p, text[*position:end], leaves, cover)
	*position = int32(end - 1)
	return result
}
func Dau_DsdToGia2_rec(p *aig.Graph, text string, position *int32, matches []int32, leaves []aig.Edge, cover *vec.Vec_Int_t) aig.Edge {
	end := dsdGraphExpressionEnd(text, int(*position), matches)
	result := Dau_DsdToGia2(p, text[*position:end], leaves, cover)
	*position = int32(end - 1)
	return result
}
