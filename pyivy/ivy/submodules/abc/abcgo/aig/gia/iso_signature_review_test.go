package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

type isoSignatureReviewEdge struct {
	id    int
	phase uint64
}

type isoSignatureReviewNode struct {
	kind  byte // constant, PI, RO, AND, CO
	edges []isoSignatureReviewEdge
	level int32
}

type isoSignatureReviewModel struct {
	nodes []isoSignatureReviewNode
	regs  [][2]int // RI, RO
}

func isoSignatureReviewGraph(t *testing.T, registers int, phases uint32) (*Gia_Man_t, isoSignatureReviewModel) {
	t.Helper()
	graph := Gia_ManStart(24)
	model := isoSignatureReviewModel{nodes: []isoSignatureReviewNode{{kind: 'c'}}}
	add := func(literal int32, kind byte, fanins ...int32) int32 {
		id := int(literal >> 1)
		if id != len(model.nodes) {
			t.Fatal("fixture constructor did not append exactly one object")
		}
		node := isoSignatureReviewNode{kind: kind}
		for _, fanin := range fanins {
			node.edges = append(node.edges, isoSignatureReviewEdge{int(fanin >> 1), uint64(fanin & 1)})
			node.level = max(node.level, model.nodes[fanin>>1].level+1)
		}
		model.nodes = append(model.nodes, node)
		return literal
	}
	a := add(Gia_ManAppendCi(graph), 'p')
	b := add(Gia_ManAppendCi(graph), 'p')
	var ros []int32
	for range registers {
		ros = append(ros, add(Gia_ManAppendCi(graph), 'r'))
	}
	phase := func(literal int32, bit uint) int32 { return literal ^ int32((phases>>bit)&1) }
	and := func(x, y int32) int32 { return add(Gia_ManAppendAnd(graph, x, y), 'a', x, y) }
	first, last := b, a
	if len(ros) != 0 {
		first, last = ros[0], ros[len(ros)-1]
	}
	x := and(phase(a, 0), phase(first, 1))
	y := and(phase(b, 2), phase(a, 3))
	z := and(phase(x, 4), phase(y, 5))
	root := and(phase(z, 6), phase(last, 7))
	and(phase(0, 8), phase(y, 9)) // Unused gate still participates in signatures.
	co := func(literal int32) int32 { return add(Gia_ManAppendCo(graph, literal), 'o', literal) }
	co(phase(root, 10))
	co(phase(0, 11))
	for i, ro := range ros {
		fanin := z
		if i == 1 {
			fanin = ro // Exercise the source's self-feedback bias in Iso2.
		}
		ri := co(phase(fanin, uint(12+i)))
		model.regs = append(model.regs, [2]int{int(ri >> 1), int(ro >> 1)})
	}
	Gia_ManSetRegNum(graph, int32(registers))
	return graph, model
}

func isoSignatureReviewAdd(values []uint64, id int, extra uint64) {
	values[id] = (values[id] + extra) & 0xffffffff
}

// Wider accumulators plus explicit masks provide an independent check on the
// production uint32 overflow, using the specification graph rather than GIA walks.
func (model isoSignatureReviewModel) firstPass(values []uint64, labels []int32, iteration int32, backward bool) {
	salt := func(id int, phase uint64) uint64 {
		value := uint64(model.nodes[id].level)
		if iteration != 0 {
			if labels[id] == 0 {
				return 0
			}
			value = uint64(labels[id])
		}
		return ((value + 1) * uint64(s_256Primes[((value<<1)|phase)&255])) & 0xffffffff
	}
	if !backward {
		isoSignatureReviewAdd(values, 0, uint64(s_256Primes[255]))
		for id, node := range model.nodes {
			if node.kind == 'p' {
				isoSignatureReviewAdd(values, id, uint64(s_256Primes[254]))
			} else if node.kind == 'r' && iteration == 0 {
				isoSignatureReviewAdd(values, id, uint64(s_256Primes[253]))
			}
		}
		for _, kind := range []byte{'a', 'o'} {
			for id, node := range model.nodes {
				if node.kind == kind {
					for _, edge := range node.edges {
						isoSignatureReviewAdd(values, id, values[edge.id]+salt(edge.id, edge.phase))
					}
				}
			}
		}
		for _, pair := range model.regs {
			isoSignatureReviewAdd(values, pair[1], values[pair[0]])
		}
		return
	}
	for id, node := range model.nodes {
		if node.kind == 'o' {
			edge := node.edges[0]
			isoSignatureReviewAdd(values, edge.id, values[id]+salt(id, edge.phase))
		}
	}
	for id := len(model.nodes) - 1; id >= 0; id-- {
		if node := model.nodes[id]; node.kind == 'a' {
			for _, edge := range node.edges {
				isoSignatureReviewAdd(values, edge.id, values[id]+salt(id, edge.phase))
			}
		}
	}
	for _, pair := range model.regs {
		isoSignatureReviewAdd(values, pair[0], values[pair[1]])
	}
}

func (model isoSignatureReviewModel) secondInit() []uint64 {
	values := make([]uint64, len(model.nodes))
	values[0] = uint64(s_256Primes[255])
	for id, node := range model.nodes {
		switch node.kind {
		case 'p':
			values[id] = uint64(s_256Primes[254])
		case 'r':
			values[id] = uint64(s_256Primes[253])
		case 'a':
			negative := node.edges[0].phase + node.edges[1].phase
			values[id] = (uint64(s_256Primes[node.level&255]) + uint64(s_256Primes[245+negative])) & 0xffffffff
		}
	}
	return values
}

func (model isoSignatureReviewModel) secondPass(values []uint64) {
	for id, node := range model.nodes {
		if node.kind == 'a' || node.kind == 'o' {
			for _, edge := range node.edges {
				isoSignatureReviewAdd(values, id, (edge.phase+49)*values[edge.id])
			}
		}
		if node.kind == 'a' && node.edges[0].phase == node.edges[1].phase && values[node.edges[0].id] == values[node.edges[1].id] {
			isoSignatureReviewAdd(values, id, uint64(s_256Primes[244]))
		}
	}
	for _, pair := range model.regs {
		isoSignatureReviewAdd(values, pair[1], values[pair[0]])
		if model.nodes[pair[0]].edges[0].id == pair[1] {
			isoSignatureReviewAdd(values, pair[1], uint64(s_256Primes[243]))
		}
	}
}

var isoSignatureReviewKinds = [6]uint64{0x04892ed6, 0xc2523d7d, 0xdc36cd2e, 0xf2db64f8, 0xde3126bb, 0xdebbdff0}

func (model isoSignatureReviewModel) thirdInit() []uint64 {
	values := make([]uint64, len(model.nodes))
	for id, node := range model.nodes {
		switch node.kind {
		case 'c':
			values[id] = isoSignatureReviewKinds[5]
		case 'p', 'r':
			values[id] = isoSignatureReviewKinds[3]
		case 'o':
			values[id] = isoSignatureReviewKinds[4]
		case 'a':
			values[id] = isoSignatureReviewKinds[node.edges[0].phase+node.edges[1].phase]
		}
	}
	return values
}

func (model isoSignatureReviewModel) thirdPass(old []uint64) []uint64 {
	values := slices.Clone(old)
	for id, node := range model.nodes {
		for _, edge := range node.edges {
			bias := []uint64{0x8ba63e50, 0x14d87f02}[edge.phase]
			isoSignatureReviewAdd(values, id, old[edge.id]+bias+0x855ee0cf)
			isoSignatureReviewAdd(values, edge.id, old[id]+bias+0x946e1b5f)
		}
	}
	return values
}

func isoSignatureReviewFields(obj *Gia_Obj_t) [8]uint32 {
	return [8]uint32{obj.iDiff0, obj.fCompl0, obj.fMark0, obj.fTerm, obj.iDiff1, obj.fCompl1, obj.fMark1, obj.fPhase}
}

func checkIsoSignatureReview(t *testing.T, graph *Gia_Man_t, expected []uint64, fields [][8]uint32, stage string) {
	t.Helper()
	if int(Gia_ManObjNum(graph)) != len(expected) {
		t.Fatalf("%s changed the graph object count", stage)
	}
	for id, want := range expected {
		obj := Gia_ManObj(graph, int32(id))
		if obj.Value != uint32(want) || isoSignatureReviewFields(obj) != fields[id] {
			t.Fatalf("%s object%d got%08x want%08x or graph/marks changed", stage, id, obj.Value, want)
		}
	}
}

func checkIsoSignaturePasses(t *testing.T, registers int, phases uint32, seed, labelProfile int) {
	t.Helper()
	graph, model := isoSignatureReviewGraph(t, registers, phases)
	expected := make([]uint64, len(model.nodes))
	labels := make([]int32, len(model.nodes))
	fields := make([][8]uint32, len(model.nodes))
	state := uint32(19)
	for id := range expected {
		state = state*1664525 + 1013904223
		if seed == 1 {
			expected[id] = (0xfffffff0 + uint64(id)*37) & 0xffffffff
		} else if seed == 2 {
			expected[id] = uint64(state)
		}
		if labelProfile == 1 || labelProfile == 2 && id%2 == 0 {
			labels[id] = int32(id + 1)
		}
		obj := Gia_ManObj(graph, int32(id))
		obj.Value, obj.fMark0, obj.fMark1 = uint32(expected[id]), uint32(id%2), uint32((id+1)%2)
		fields[id] = isoSignatureReviewFields(obj)
	}
	p := Gia_IsoManStart(graph)
	defer Gia_IsoManStop(p)
	Gia_IsoPrepare(p)
	copy(p.pUniques, labels)
	for _, iteration := range []int32{0, 1, 7} {
		model.firstPass(expected, labels, iteration, false)
		Gia_IsoSimulate(p, iteration)
		checkIsoSignatureReview(t, graph, expected, fields, "forward")
		model.firstPass(expected, labels, iteration, true)
		Gia_IsoSimulateBack(p, iteration)
		checkIsoSignatureReview(t, graph, expected, fields, "backward")
	}
	expected = model.secondInit()
	Gia_Iso2ManPrepare(graph)
	checkIsoSignatureReview(t, graph, expected, fields, "second initialization")
	for range 3 {
		model.secondPass(expected)
		Gia_Iso2ManPropagate(graph)
		checkIsoSignatureReview(t, graph, expected, fields, "second propagation")
	}
	expected = model.thirdInit()
	Gia_Iso3Init(graph)
	checkIsoSignatureReview(t, graph, expected, fields, "third initialization")
	for range 3 {
		snapshot := Gia_Iso3Save(graph)
		before := slices.Clone(vec.Vec_IntArray(snapshot))
		if len(before) != len(expected) {
			t.Fatal("snapshot object count changed")
		}
		var distinct int32
		for i, value := range before {
			if uint32(value) != uint32(expected[i]) {
				t.Fatalf("snapshot%d got%08x want%08x", i, uint32(value), expected[i])
			}
			if !slices.Contains(before[:i], value) {
				distinct++
			}
		}
		if Gia_Iso3Unique(snapshot) != distinct {
			t.Fatal("snapshot distinct-value count differs from scalar equality")
		}
		expected = model.thirdPass(expected)
		Gia_Iso3Compute(graph, snapshot)
		checkIsoSignatureReview(t, graph, expected, fields, "third synchronous propagation")
		if !slices.Equal(vec.Vec_IntArray(snapshot), before) {
			t.Fatal("third-generation snapshot changed")
		}
	}
}

func TestIsoSignaturePassesAgainstIndependentWordModel(t *testing.T) {
	phases := []uint32{0, 0x3fff, 0x1555, 0x2aaa}
	for bit := 0; bit < 14; bit++ {
		phases = append(phases, 1<<bit)
	}
	for registers := 0; registers <= 2; registers++ {
		for _, phase := range phases {
			for seed := 0; seed < 3; seed++ {
				for labels := 0; labels < 3; labels++ {
					t.Run(fmt.Sprintf("registers%d/phases%04x/seed%d/labels%d", registers, phase, seed, labels), func(t *testing.T) {
						checkIsoSignaturePasses(t, registers, phase, seed, labels)
					})
				}
			}
		}
	}
}

func TestIsoUpdateValueIndexAndUnsignedProduct(t *testing.T) {
	for _, value := range []int32{0, 1, 127, 128, 255, 256, 65535, 1 << 20, 1 << 29, (1 << 30) - 1} {
		for phase := int32(0); phase < 2; phase++ {
			index := (uint64(value)*2 + uint64(phase)) % 256
			want := uint32((uint64(value) + 1) * uint64(s_256Primes[index]))
			if got := Gia_IsoUpdateValue(value, phase); got != want {
				t.Fatalf("value%d phase%d got%08x want%08x", value, phase, got, want)
			}
		}
	}
}
