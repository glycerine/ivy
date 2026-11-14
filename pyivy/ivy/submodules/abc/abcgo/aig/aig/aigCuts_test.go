package aig

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func testCut(leaves ...int32) *Aig_Cut_t {
	c := &Aig_Cut_t{valid: true}
	cutSetLeaves(c, leaves)
	return c
}
func TestPriorityCutOrderedMerge(t *testing.T) {
	p := &Aig_ManCut_t{nLeafMax: 3}
	for a := 0; a < 64; a++ {
		for b := 0; b < 64; b++ {
			var x, y, want []int32
			for bit := 0; bit < 6; bit++ {
				if a>>bit&1 != 0 {
					x = append(x, int32(bit))
				}
				if b>>bit&1 != 0 {
					y = append(y, int32(bit))
				}
				if (a|b)>>bit&1 != 0 {
					want = append(want, int32(bit))
				}
			}
			out := new(Aig_Cut_t)
			ok := Aig_CutMerge(p, testCut(x...), testCut(y...), out)
			if ok != (len(want) <= 3) {
				t.Fatal(a, b, ok)
			}
			if ok && !reflect.DeepEqual(out.pFanins, want) {
				t.Fatal(out.pFanins, want)
			}
		}
	}
	a, b := testCut(1, 3), testCut(2, 3)
	if !Aig_CutMerge(p, a, b, a) || !reflect.DeepEqual(a.pFanins, []int32{1, 2, 3}) {
		t.Fatal("alias merge", a)
	}
	if !Aig_CutCheckDominance(testCut(2, 3), testCut(1, 2, 3)) || Aig_CutCheckDominance(testCut(4), testCut(1, 2, 3)) {
		t.Fatal("dominance")
	}
}
func TestPriorityCutFilterAndSlots(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	e := g.AND(a, b)
	p := Aig_ManCutStart(g, 3, 4, 0, 0)
	Aig_ObjPrepareCuts(p, e, 1)
	first := p.pCuts[e.NodeID()]
	second, third := first.pNext, first.pNext.pNext
	cutCopy(second, testCut(1, 2, 3))
	second.Cost = 100
	cutCopy(third, testCut(4, 5))
	third.Cost = 200
	candidate := testCut(1, 2)
	if Aig_CutFilter(p, e, candidate) || cutLive(second) {
		t.Fatal("superset retained")
	}
	if Aig_CutFindFree(p, e) != second {
		t.Fatal("free slot not reused")
	}
	cutCopy(second, candidate)
	second.Cost = 50
	if got := Aig_CutFindFree(p, e); got != third || cutLive(third) || !cutLive(first) {
		t.Fatal("wrong replacement")
	}
	if !Aig_CutFilter(p, e, testCut(1, 2, 3)) {
		t.Fatal("contained candidate accepted")
	}
}
func TestPriorityFailedMergePreservesExistingCut(t *testing.T) {
	g := NewGraph()
	inputs := make([]Edge, 40)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	x := g.AND(inputs[0], inputs[1])
	y := g.AND(inputs[8], inputs[9])
	root := g.AND(x, y)
	g.AddOutput(root)
	p := Aig_ComputeCuts(g, 2, 2, 0, 0)
	defer Aig_ManCutStop(p)
	// x/y IDs 41/42 collide in the signature with y's leaves 9/10. Thus
	// an over-limit merge passes the signature precheck after {x,y} is stored.
	cuts := p.Cuts(root)
	if len(cuts) != 2 {
		t.Fatal("failed merge evicted valid cut", cuts)
	}
	if !reflect.DeepEqual(cuts[1].pFanins, []int32{int32(x.NodeID()), int32(y.NodeID())}) {
		t.Fatal(cuts[1].pFanins)
	}
}
func TestPriorityTruthProjectionAndSupport(t *testing.T) {
	p := &Aig_ManCut_t{nLeafMax: 5, nTruthWords: 1, fTruth: 1}
	a, b := testCut(1), testCut(3)
	a.truth = []uint32{0xaaaaaaaa}
	b.truth = []uint32{0xaaaaaaaa}
	out := testCut(1, 2, 3)
	if Aig_CutTruthPhase(out, b) != 4 {
		t.Fatal("phase")
	}
	Aig_CutComputeTruth(p, out, a, b, 0, 1)
	if out.truth[0] != 0x0a0a0a0a {
		t.Fatalf("projection %x", out.truth[0])
	}
	out.truth = []uint32{0x5a5a5a5a}
	if Aig_CutSupportMinimize(p, out) != 2 || !reflect.DeepEqual(out.pFanins, []int32{1, 3}) || out.truth[0] != 0x66666666 {
		t.Fatalf("support %+v %x", out, out.truth)
	}
	out.truth[0] = ^uint32(0)
	if Aig_CutSupportMinimize(p, out) != 0 || !cutLive(out) || len(out.pFanins) != 0 || out.truth[0] != ^uint32(0) {
		t.Fatal("constant cut lost", out)
	}
	wide := &Aig_ManCut_t{nLeafMax: 16, nTruthWords: 2048, fTruth: 1}
	leaves := make([]int32, 16)
	for i := range leaves {
		leaves[i] = int32(i + 1)
	}
	merged := testCut(leaves...)
	a, b = testCut(1), testCut(16)
	a.truth = make([]uint32, 2048)
	b.truth = make([]uint32, 2048)
	for i := range a.truth {
		a.truth[i] = 0xaaaaaaaa
		b.truth[i] = 0xaaaaaaaa
	}
	Aig_CutComputeTruth(wide, merged, a, b, 0, 0)
	for _, index := range []int{0, 1, 32768, 32769, 65535} {
		if cutTruthBit(merged.truth, index) != (index&1 != 0 && index&(1<<15) != 0) {
			t.Fatal("wide phase", index)
		}
	}
}

func independentNodeValues(g *Graph, assignment int) []bool {
	values := make([]bool, len(g.nodes))
	done := make([]bool, len(g.nodes))
	inputs := map[int]bool{}
	for i, e := range g.inputs {
		inputs[e.NodeID()] = assignment>>uint(i)&1 != 0
	}
	var visit func(int) bool
	visit = func(id int) bool {
		if done[id] {
			return values[id]
		}
		n := g.nodes[id]
		edge := func(e Edge) bool { return visit(e.NodeID()) != e.Complemented() }
		switch n.Kind {
		case Constant:
			values[id] = true
		case Input:
			values[id] = inputs[id]
		case And:
			values[id] = edge(n.Fanins[0]) && edge(n.Fanins[1])
		case ExclusiveOr:
			values[id] = edge(n.Fanins[0]) != edge(n.Fanins[1])
		case Buffer, Output:
			values[id] = edge(n.Fanins[0])
		}
		done[id] = true
		return values[id]
	}
	for id := range g.nodes {
		if !g.deleted[id] {
			visit(id)
		}
	}
	return values
}
func TestPriorityCutsTruthMatchesGraph(t *testing.T) {
	random := rand.New(rand.NewSource(17))
	for trial := 0; trial < 40; trial++ {
		g := NewGraph()
		edges := []Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 7; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		g.AddOutput(edges[len(edges)-1].Not())
		p := Aig_ComputeCuts(g, 8, 4, 1, 0)
		for assignment := 0; assignment < 16; assignment++ {
			values := independentNodeValues(g, assignment)
			for id := range g.nodes {
				if g.deleted[id] {
					continue
				}
				for _, c := range p.Cuts(Edge(id) << 1) {
					index := 0
					for pos, leaf := range c.pFanins {
						if values[leaf] {
							index |= 1 << uint(pos)
						}
					}
					if cutTruthBit(c.truth, index) != values[id] {
						t.Fatalf("trial %d node %d leaves %v assignment %d", trial, id, c.pFanins, assignment)
					}
				}
			}
		}
		for id := range g.nodes {
			if got := len(p.Cuts(Edge(id) << 1)); got > 8 {
				t.Fatal("priority budget exceeded", got)
			}
		}
		Aig_ManCutStop(p)
		if g.cutManager != nil {
			t.Fatal("manager attachment leaked")
		}
	}
}
func TestPriorityCutsBuffersConstantsAndRewire(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	old := g.AND(a, b)
	co := g.AddOutput(old)
	replacement := g.AND(a, c)
	Aig_ObjPatchFanin0(g, co, replacement.Not())
	zero := g.AddOutput(g.Const0())
	p := Aig_ComputeCuts(g, 6, 3, 1, 0)
	defer Aig_ManCutStop(p)
	if len(p.Cuts(zero)) != 1 || len(p.Cuts(zero)[0].pFanins) != 0 || p.Cuts(zero)[0].truth[0] != 0 {
		t.Fatal("constant output cut")
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := independentNodeValues(g, assignment)
		for _, cut := range p.Cuts(co) {
			index := 0
			for i, id := range cut.pFanins {
				if values[id] {
					index |= 1 << uint(i)
				}
			}
			if cutTruthBit(cut.truth, index) != values[co.NodeID()] {
				t.Fatal("rewired output truth")
			}
		}
	}
	var full int32
	if Aig_ManCutCount(p, &full) == 0 || full < 0 {
		t.Fatal("cut accounting")
	}
	for _, cut := range p.Cuts(replacement) {
		snapshot := cut.Leaves()
		if len(snapshot) > 0 {
			snapshot[0] = -1
			if cut.pFanins[0] < 0 {
				t.Fatal("leaf snapshot aliased")
			}
		}
		truth := cut.Truth()
		truth[0] ^= 1
		if truth[0] == cut.truth[0] {
			t.Fatal("truth snapshot aliased")
		}
	}
}

func TestPriorityCutsCompleteWhenBudgetSuffices(t *testing.T) {
	random := rand.New(rand.NewSource(103))
	for trial := 0; trial < 30; trial++ {
		g := NewGraph()
		edges := []Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 5; i++ {
			edges = append(edges, g.AND(edges[random.Intn(len(edges))], edges[random.Intn(len(edges))].Not()))
		}
		g.AddOutput(edges[len(edges)-1])
		p := Aig_ComputeCuts(g, 128, 3, 0, 0)
		memo := map[int][][]int32{}
		subset := func(a, b []int32) bool {
			for _, x := range a {
				found := false
				for _, y := range b {
					if x == y {
						found = true
					}
				}
				if !found {
					return false
				}
			}
			return true
		}
		var oracle func(int) [][]int32
		oracle = func(id int) [][]int32 {
			if result, ok := memo[id]; ok {
				return result
			}
			node := g.nodes[id]
			sets := map[string][]int32{}
			add := func(leaves []int32) {
				copy := append([]int32(nil), leaves...)
				sort.Slice(copy, func(i, j int) bool { return copy[i] < copy[j] })
				unique := copy[:0]
				for _, x := range copy {
					if len(unique) == 0 || unique[len(unique)-1] != x {
						unique = append(unique, x)
					}
				}
				if len(unique) <= 3 {
					sets[fmt.Sprint(unique)] = unique
				}
			}
			switch node.Kind {
			case Constant:
				add(nil)
			case Input:
				add([]int32{int32(id)})
			case Buffer, Output:
				for _, c := range oracle(node.Fanins[0].NodeID()) {
					add(c)
				}
				if node.Kind == Buffer {
					add([]int32{int32(id)})
				}
			default:
				add([]int32{int32(id)})
				for _, a := range oracle(node.Fanins[0].NodeID()) {
					for _, b := range oracle(node.Fanins[1].NodeID()) {
						add(append(append([]int32(nil), a...), b...))
					}
				}
			}
			var result [][]int32
			for _, candidate := range sets {
				dominated := false
				if !(len(candidate) == 1 && candidate[0] == int32(id)) {
					for _, other := range sets {
						if len(other) < len(candidate) && subset(other, candidate) {
							dominated = true
							break
						}
					}
				}
				if !dominated {
					result = append(result, candidate)
				}
			}
			memo[id] = result
			return result
		}
		for id := range g.nodes {
			expected := oracle(id)
			actual := p.Cuts(Edge(id) << 1)
			if len(actual) != len(expected) {
				t.Fatalf("trial=%d node=%d count=%d expected=%v", trial, id, len(actual), expected)
			}
			keys := map[string]bool{}
			for _, c := range actual {
				keys[fmt.Sprint(c.pFanins)] = true
			}
			for _, c := range expected {
				if !keys[fmt.Sprint(c)] {
					t.Fatal("missing structural cut", trial, id, c)
				}
			}
		}
		Aig_ManCutStop(p)
	}
}

func TestPriorityCutsOriginalCMergeAndPhase(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	source, err := os.ReadFile("../../../src/aig/aig/aigCuts.c")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	extract := func(name string) string {
		start := strings.Index(text, "static inline "+name)
		if start < 0 {
			t.Fatal("missing C helper", name)
		}
		open := start + strings.IndexByte(text[start:], '{')
		depth, end := 1, open+1
		for depth > 0 {
			if text[end] == '{' {
				depth++
			}
			if text[end] == '}' {
				depth--
			}
			end++
		}
		return text[start:end]
	}
	code := "#include <stdio.h>\n#include <assert.h>\ntypedef struct{int nLeafMax;} Aig_ManCut_t;\ntypedef struct{int nFanins,pFanins[16];} Aig_Cut_t;\n" + extract("int Aig_CutMergeOrdered(") + "\n" + extract("unsigned Aig_CutTruthPhase(") + `
int main(void){Aig_ManCut_t p;Aig_Cut_t a,b,out;
while(scanf("%d%d%d",&a.nFanins,&b.nFanins,&p.nLeafMax)==3){
for(int i=0;i<a.nFanins;i++)scanf("%d",&a.pFanins[i]);for(int i=0;i<b.nFanins;i++)scanf("%d",&b.pFanins[i]);
int ok=a.nFanins>=b.nFanins?Aig_CutMergeOrdered(&p,&a,&b,&out):Aig_CutMergeOrdered(&p,&b,&a,&out);
printf("%d",ok);if(ok){printf(" %d %u %u",out.nFanins,Aig_CutTruthPhase(&out,&a),Aig_CutTruthPhase(&out,&b));for(int i=0;i<out.nFanins;i++)printf(" %d",out.pFanins[i]);}printf("\n");}return 0;}`
	dir := t.TempDir()
	path := filepath.Join(dir, "cuts.c")
	binary := filepath.Join(dir, "cuts")
	if err := os.WriteFile(path, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, path, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("C oracle %v: %s", err, out)
	}
	var input, expected strings.Builder
	random := rand.New(rand.NewSource(12))
	for i := 0; i < 1024; i++ {
		a, b := random.Intn(256), random.Intn(256)
		var x, y []int32
		for bit := 0; bit < 8; bit++ {
			if a>>bit&1 != 0 {
				x = append(x, int32(bit))
			}
			if b>>bit&1 != 0 {
				y = append(y, int32(bit))
			}
		}
		limit := int32(1 + random.Intn(7))
		fmt.Fprintf(&input, "%d %d %d", len(x), len(y), limit)
		for _, id := range append(append([]int32(nil), x...), y...) {
			fmt.Fprintf(&input, " %d", id)
		}
		input.WriteByte('\n')
		p := &Aig_ManCut_t{nLeafMax: limit}
		out := new(Aig_Cut_t)
		c0, c1 := testCut(x...), testCut(y...)
		if Aig_CutMerge(p, c0, c1, out) {
			fmt.Fprintf(&expected, "1 %d %d %d", len(out.pFanins), Aig_CutTruthPhase(out, c0), Aig_CutTruthPhase(out, c1))
			for _, id := range out.pFanins {
				fmt.Fprintf(&expected, " %d", id)
			}
			expected.WriteByte('\n')
		} else {
			expected.WriteString("0\n")
		}
	}
	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader(input.String())
	actual, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(actual))
	}
	if string(actual) != expected.String() {
		t.Fatal("merge/phase differs from original C")
	}
}

func TestPriorityCutXorBufferAndManagerLifecycle(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	xor := Aig_ObjCreate(g, &Node{Kind: ExclusiveOr, Fanins: [2]Edge{a, b.Not()}})
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{xor.Not(), InvalidEdge}})
	co := g.AddOutput(buffer)
	p := Aig_ComputeCuts(g, 6, 4, 1, 0)
	for assignment := 0; assignment < 4; assignment++ {
		values := independentNodeValues(g, assignment)
		for _, e := range []Edge{xor, buffer, co} {
			for _, c := range p.Cuts(e) {
				index := 0
				for i, id := range c.pFanins {
					if values[id] {
						index |= 1 << uint(i)
					}
				}
				if cutTruthBit(c.truth, index) != values[e.NodeID()] {
					t.Fatal("XOR/buffer phase", e, assignment, c.pFanins)
				}
			}
		}
	}
	Aig_ManCutStop(p)
	if p.pAig != nil || g.cutManager != nil {
		t.Fatal("manager not released")
	}
	p = Aig_ComputeCuts(g, 2, 2, 0, 0)
	Aig_ManCutStop(p)
	Aig_ManCutStop(nil)
}

func TestPriorityCutSmallSupportWidePadding(t *testing.T) {
	p := &Aig_ManCut_t{nLeafMax: 16, nTruthWords: 2048, fTruth: 1}
	a, b, out := testCut(1), testCut(2), testCut(1, 2)
	a.truth = make([]uint32, 2048)
	b.truth = make([]uint32, 2048)
	for i := range a.truth {
		a.truth[i] = 0xaaaaaaaa
		b.truth[i] = 0xaaaaaaaa
	}
	Aig_CutComputeTruth(p, out, a, b, 0, 0)
	for _, word := range out.truth {
		if word != 0x88888888 {
			t.Fatalf("truth padding %x", word)
		}
	}
	out = testCut(1, 2, 3)
	out.truth = make([]uint32, 2048)
	for i := range out.truth {
		out.truth[i] = 0x5a5a5a5a
	}
	Aig_CutSupportMinimize(p, out)
	for _, word := range out.truth {
		if word != 0x66666666 {
			t.Fatalf("support padding %x", word)
		}
	}
}

func TestPriorityCutCostsAndParameterValidation(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AND(a, b)
	g.AND(a, c)
	p := Aig_ManCutStart(g, 3, 2, 0, 0)
	cut := testCut(int32(a.NodeID()), int32(b.NodeID()))
	if Aig_CutFindCost(p, cut) != 1500 {
		t.Fatal("average fanout cost")
	}
	if cost := Aig_CutFindCost2(p, cut); cost < 0.6666 || cost > 0.6667 {
		t.Fatal(cost)
	}
	checkPanic := func(f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Error("invalid call accepted")
			}
		}()
		f()
	}
	for _, params := range [][2]int32{{1, 2}, {2, 0}, {2, 17}} {
		checkPanic(func() { Aig_ManCutStart(g, params[0], params[1], 0, 0) })
	}
	checkPanic(func() { Aig_ManCutStart(nil, 2, 2, 0, 0) })
	attached := Aig_ComputeCuts(g, 3, 2, 0, 0)
	checkPanic(func() { Aig_ComputeCuts(g, 3, 2, 0, 0) })
	Aig_ManCutStop(attached)
}

func TestPriorityCutsXorAndBufferTruthAcrossWordBoundary(t *testing.T) {
	g := NewGraph()
	inputs := make([]Edge, 6)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	root := g.NewXor(inputs[0], inputs[1].Not())
	for _, input := range inputs[2:] {
		root = g.NewXor(root, input)
	}
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{root.Not(), InvalidEdge}})
	g.AddOutput(buffer.Not())
	g.AddOutput(g.AND(buffer, inputs[0]))
	manager := Aig_ComputeCuts(g, 16, 6, 1, 0)
	defer Aig_ManCutStop(manager)
	foundWide := false
	for assignment := 0; assignment < 64; assignment++ {
		values := independentNodeValues(g, assignment)
		for _, edge := range g.LiveEdges() {
			for _, cut := range manager.Cuts(edge) {
				foundWide = foundWide || len(cut.pFanins) == 6
				index := 0
				for bit, leaf := range cut.pFanins {
					if values[leaf] {
						index |= 1 << bit
					}
				}
				if got := cutTruthBit(cut.truth, index); got != values[edge.NodeID()] {
					t.Fatalf("assignment=%d node=%d leaves=%v got=%v want=%v", assignment, edge.NodeID(), cut.pFanins, got, values[edge.NodeID()])
				}
			}
		}
	}
	if !foundWide {
		t.Fatal("did not exercise a truth table spanning multiple words")
	}
}
