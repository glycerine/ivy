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

func domTestFanouts(g *Graph) *Graph {
	if g.fanouts == nil {
		Aig_ManFanoutStart(g)
	}
	return g
}

// Check graph structure and values without imposing a non-source promise that
// analysis leaves the traversal generation unchanged.
func domTestWithoutTraversal(nodes []Node) []Node {
	for i := range nodes {
		nodes[i].TravID = 0
	}
	return nodes
}

func TestDominatorOrderedUnion(t *testing.T) {
	for a := 0; a < 64; a++ {
		for b := 0; b < 64; b++ {
			var x, y, union []int
			for bit := 0; bit < 6; bit++ {
				if a>>bit&1 != 0 {
					x = append(x, bit)
				}
				if b>>bit&1 != 0 {
					y = append(y, bit)
				}
				if (a|b)>>bit&1 != 0 {
					union = append(union, bit)
				}
			}
			for limit := 1; limit <= 4; limit++ {
				out := newAigDom([]int{99})
				ok := Aig_ObjDomMergeTwo(newAigDom(x), newAigDom(y), out, limit)
				if ok != (len(union) <= limit) {
					t.Fatal(a, b, limit, ok)
				}
				if ok && (!reflect.DeepEqual(out.Nodes, union) || !Aig_ObjDomCheck(out)) {
					t.Fatal(out, union)
				}
			}
		}
	}
	a := newAigDom([]int{0, 63})
	b := newAigDom([]int{63, 126})
	if !Aig_ObjDomMergeTwo(a, b, a, 3) || !reflect.DeepEqual(a.Nodes, []int{0, 63, 126}) {
		t.Fatal(a)
	}
	if Aig_ObjDomCheck(&Aig_Dom_t{Nodes: []int{2, 1}}) || Aig_ObjDomCheck(&Aig_Dom_t{Nodes: []int{1, 1}}) {
		t.Fatal("invalid order accepted")
	}
}
func TestDominatorFilterAndCopy(t *testing.T) {
	s := &Aig_Sto_t{Limit: 4}
	first := newAigDom([]int{10})
	doms := []*Aig_Dom_t{first, newAigDom([]int{1, 2, 3}), newAigDom([]int{4, 5})}
	if Aig_ObjDomFilter(s, &doms, newAigDom([]int{1, 2})) || len(doms) != 2 || doms[0] != first || s.FilteredSupersets != 1 {
		t.Fatal(doms, s)
	}
	if !Aig_ObjDomFilter(s, &doms, newAigDom([]int{4, 5, 6})) || s.FilteredCandidates != 1 {
		t.Fatal(s)
	}
	copy := Aig_ObjDomVecDup(s, doms, 1)
	copy[0].Nodes[0] = 99
	if doms[1].Nodes[0] == 99 {
		t.Fatal("copy aliased")
	}
	// The first cut is protected even when it is a strict superset.
	protected := []*Aig_Dom_t{newAigDom([]int{1, 2, 3})}
	Aig_ObjDomFilter(s, &protected, newAigDom([]int{1, 2}))
	if len(protected) != 1 {
		t.Fatal(protected)
	}
}
func domSetKey(nodes []int) string { return fmt.Sprint(nodes) }
func TestDominatorReconvergence(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x, y := g.AND(a, b), g.AND(a, c)
	z := g.AND(x, y)
	g.AddOutput(z)
	original := g.Nodes()
	s := Aig_ManComputeDomsNodes(domTestFanouts(g), 2)
	got := map[string]bool{}
	for _, d := range s.Dominators[a.NodeID()] {
		got[domSetKey(d.Nodes)] = true
	}
	for _, want := range [][]int{{a.NodeID()}, {x.NodeID(), y.NodeID()}, {z.NodeID()}, {s.OutputIDs[0]}} {
		if !got[domSetKey(want)] {
			t.Fatal(want, s.DominatorSets(a.NodeID()))
		}
	}
	if Aig_ObjDomVolume(s, newAigDom([]int{z.NodeID()})) != 3 {
		t.Fatal("wrong boundary volume")
	}
	if !reflect.DeepEqual(domTestWithoutTraversal(original), domTestWithoutTraversal(g.Nodes())) {
		t.Fatal("analysis mutated graph")
	}
	sets := s.DominatorSets(a.NodeID())
	sets[0][0] = 999
	if reflect.DeepEqual(sets, s.DominatorSets(a.NodeID())) {
		t.Fatal("exported sets aliased")
	}
}

// Enumerate cut sets independently by checking all directed paths to output
// boundaries. Every inclusion-minimal bounded cut must be reported.
func TestDominatorBruteForcePaths(t *testing.T) {
	random := rand.New(rand.NewSource(37))
	for trial := 0; trial < 40; trial++ {
		g := NewGraph()
		edges := []Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 4; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			e := g.AND(a, b)
			edges = append(edges, e)
		}
		g.AddOutput(edges[len(edges)-1])
		g.AddOutput(edges[len(edges)-2])
		Aig_ManCleanup(g)
		s := Aig_ManComputeDomsNodes(domTestFanouts(g), 2)
		active := make([]bool, len(s.fanins))
		var activate func(int)
		activate = func(id int) {
			if active[id] {
				return
			}
			active[id] = true
			for _, f := range s.fanins[id] {
				activate(f)
			}
		}
		for _, id := range s.OutputIDs {
			activate(id)
		}

		terminals := map[int]bool{}
		for _, id := range s.OutputIDs {
			terminals[id] = true
		}
		for _, start := range g.inputs {
			id := start.NodeID()
			if !active[id] {
				continue
			}
			candidates := []int{}
			seen := map[int]bool{}
			var reach func(int)
			reach = func(n int) {
				if seen[n] || !active[n] {
					return
				}
				seen[n] = true
				candidates = append(candidates, n)
				for _, f := range s.fanouts[n] {
					reach(f)
				}
			}
			reach(id)
			sort.Ints(candidates)
			hits := func(cut map[int]bool) bool {
				memo := map[int]bool{}
				var walk func(int) bool
				walk = func(n int) bool {
					if cut[n] {
						return true
					}
					if terminals[n] {
						return false
					}
					if value, ok := memo[n]; ok {
						return value
					}
					for _, f := range s.fanouts[n] {
						if active[f] && !walk(f) {
							memo[n] = false
							return false
						}
					}
					memo[n] = true
					return true
				}
				return walk(id)
			}
			actual := map[string]bool{}
			for _, d := range s.Dominators[id] {
				cut := map[int]bool{}
				for _, n := range d.Nodes {
					cut[n] = true
				}
				if !hits(cut) {
					t.Fatalf("invalid dominator trial=%d root=%d cut=%v", trial, id, d.Nodes)
				}
				actual[domSetKey(d.Nodes)] = true
			}
			for i, a := range candidates {
				if hits(map[int]bool{a: true}) && !actual[domSetKey([]int{a})] {
					t.Fatalf("missing singleton %d", a)
				}
				for _, b := range candidates[i+1:] {
					cut := map[int]bool{a: true, b: true}
					if hits(cut) && !hits(map[int]bool{a: true}) && !hits(map[int]bool{b: true}) && !actual[domSetKey([]int{a, b})] {
						t.Fatalf("missing minimal pair %d %d in %v", a, b, s.DominatorSets(id))
					}
				}
			}
		}
	}
}

func TestDominatorFlopConeAndRefRestoration(t *testing.T) {
	g := NewGraph()
	a, b, lo := g.NewInput(), g.NewInput(), g.NewInput()
	po, li := g.AND(a, lo), g.AND(b, lo)
	g.AddOutput(po)
	g.AddOutput(li)
	g.nRegs = 1
	nodes := g.Nodes()
	s := Aig_ManComputeDomsFlops(domTestFanouts(g), 2)
	if len(s.Dominators[a.NodeID()]) != 0 || len(s.Dominators[b.NodeID()]) == 0 || len(s.Dominators[lo.NodeID()]) == 0 {
		t.Fatal("flop cone filtering")
	}
	refs := append([]int(nil), s.refs...)
	reports := Aig_ObjDomFindGood(s)
	if len(reports) != 2 {
		t.Fatal(reports)
	}
	for _, r := range reports {
		if r.Domed != 1 || r.Volume != 2 {
			t.Fatal(r)
		}
	}
	if !reflect.DeepEqual(refs, s.refs) || !reflect.DeepEqual(domTestWithoutTraversal(nodes), domTestWithoutTraversal(g.Nodes())) {
		t.Fatal("references or graph mutated")
	}
	for _, mark := range s.marked {
		if mark {
			t.Fatal("mark leaked")
		}
	}
	pis := Aig_ManComputeDomsPis(domTestFanouts(g), 1)
	if len(pis.Dominators[lo.NodeID()]) != 0 || len(pis.Dominators[b.NodeID()]) == 0 {
		t.Fatal("PI filtering")
	}
	Aig_ManMarkFlopTfi(g)
	if g.nodes[a.NodeID()].TravID == g.travID || g.nodes[b.NodeID()].TravID != g.travID {
		t.Fatal("wrong flop TFI marks")
	}
}

func TestDominatorDanglingAndDuplicateOutputs(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	live := g.AND(a, b)
	_ = g.AND(a, c)
	g.AddOutput(live)
	g.AddOutput(live)
	Aig_ManCleanup(g)
	s := Aig_ManComputeDomsNodes(domTestFanouts(g), 2)
	if len(s.Dominators[c.NodeID()]) != 0 {
		t.Fatal("dangling input analyzed")
	}
	if s.OutputIDs[0] == s.OutputIDs[1] {
		t.Fatal("output boundaries merged")
	}
	found := false
	for _, d := range s.Dominators[a.NodeID()] {
		if reflect.DeepEqual(d.Nodes, []int{live.NodeID()}) {
			found = true
		}
	}
	if !found {
		t.Fatal(s.DominatorSets(a.NodeID()))
	}
}

func TestDominatorCofactoringCounts(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	var fans []Edge
	for i := 0; i < 10; i++ {
		fans = append(fans, g.AND(a, g.NewInput()))
	}
	root := fans[0]
	for _, f := range fans[1:] {
		root = g.AND(root, f)
	}
	g.AddOutput(root)
	s := Aig_ManComputeDomsNodes(domTestFanouts(g), 1)
	counts := Aig_ObjDomCount(s, a)
	if counts[root.NodeID()] != 10 || counts[s.OutputIDs[0]] != 10 {
		t.Fatal(counts)
	}
	for _, n := range s.Times {
		if n != 0 {
			t.Fatal("counting scratch not reset")
		}
	}
	Aig_ManFanoutStop(g)
	result := Aig_ManComputeDomsForCofactoring(g)
	if result[a.NodeID()][root.NodeID()] != 10 {
		t.Fatal(result)
	}
}

func TestDominatorExplicitOutputAndSignatures(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	co := Aig_ObjCreateCo(g, x)
	s := Aig_ManComputeDomsNodes(domTestFanouts(g), 1)
	if len(s.OutputIDs) != 1 || s.OutputIDs[0] != co.NodeID() {
		t.Fatal(s.OutputIDs, co)
	}
	found := false
	for _, d := range s.Dominators[a.NodeID()] {
		if reflect.DeepEqual(d.Nodes, []int{co.NodeID()}) {
			found = true
		}
	}
	if !found {
		t.Fatal(s.DominatorSets(a.NodeID()))
	}
	d := NewDominator(63, 1, 63, 2)
	if !reflect.DeepEqual(d.Nodes, []int{1, 2, 63}) {
		t.Fatal(d)
	}
	// Signature is a cache, not an alternative source of set membership.
	broken := &Aig_Dom_t{Signature: ^uint64(0), Nodes: []int{1}}
	got := Aig_ObjDomMerge(&Aig_Sto_t{Limit: 1}, []*Aig_Dom_t{broken}, []*Aig_Dom_t{newAigDom([]int{1})})
	if len(got) != 1 || !reflect.DeepEqual(got[0].Nodes, []int{1}) {
		t.Fatal(got)
	}
}

func TestDominatorRelatedCutRefRestoration(t *testing.T) {
	g := NewGraph()
	a, lo := g.NewInput(), g.NewInput()
	x := g.AND(a, lo)
	y := g.AND(x, a)
	g.AddOutput(y)
	g.nRegs = 1
	s := Aig_ManComputeDomsFlops(domTestFanouts(g), 2)
	Aig_ObjDomCollectLos(s)
	before := append([]int(nil), s.refs...)
	if got := Aig_ObjDomDomed(s, newAigDom([]int{x.NodeID(), y.NodeID()})); got != 1 {
		t.Fatal(got)
	}
	if !reflect.DeepEqual(before, s.refs) {
		t.Fatal("references not restored", before, s.refs)
	}
}

func TestDominatorOriginalCOrderedMerge(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for oracle")
	}
	source, err := os.ReadFile("../../../src/aig/aig/aigDoms.c")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	start := strings.Index(text, "static inline int Aig_ObjDomMergeOrdered(")
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
	code := "#include <stdio.h>\n#include <assert.h>\ntypedef struct {int uSign,nNodes,pNodes[16];} Aig_Dom_t;\n" + text[start:end] + `
int main(void){Aig_Dom_t a,b,out;int limit;
while(scanf("%d%d%d",&a.nNodes,&b.nNodes,&limit)==3){
 for(int i=0;i<a.nNodes;i++)scanf("%d",&a.pNodes[i]);
 for(int i=0;i<b.nNodes;i++)scanf("%d",&b.pNodes[i]);
 int ok=a.nNodes>=b.nNodes?Aig_ObjDomMergeOrdered(&a,&b,&out,limit):Aig_ObjDomMergeOrdered(&b,&a,&out,limit);
 printf("%d",ok);if(ok){printf(" %d",out.nNodes);for(int i=0;i<out.nNodes;i++)printf(" %d",out.pNodes[i]);}printf("\n");
}return 0;}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "dom.c")
	binary := filepath.Join(dir, "dom")
	if err := os.WriteFile(path, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, path, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("C oracle: %v: %s", err, out)
	}
	var input, expected strings.Builder
	random := rand.New(rand.NewSource(71))
	for i := 0; i < 2048; i++ {
		a, b := random.Intn(128), random.Intn(128)
		var x, y []int
		for bit := 0; bit < 7; bit++ {
			if a>>bit&1 != 0 {
				x = append(x, bit)
			}
			if b>>bit&1 != 0 {
				y = append(y, bit)
			}
		}
		limit := 1 + random.Intn(5)
		fmt.Fprintf(&input, "%d %d %d", len(x), len(y), limit)
		for _, id := range append(append([]int(nil), x...), y...) {
			fmt.Fprintf(&input, " %d", id)
		}
		input.WriteByte('\n')
		out := new(Aig_Dom_t)
		if Aig_ObjDomMergeTwo(newAigDom(x), newAigDom(y), out, limit) {
			fmt.Fprintf(&expected, "1 %d", len(out.Nodes))
			for _, id := range out.Nodes {
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
		t.Fatal("Go ordered merge differs from original C")
	}
}

func TestDominatorRepeatedExplicitOutput(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	co := Aig_ObjCreateCo(g, x)
	g.AddOutput(co)
	s := Aig_ManComputeDomsNodes(domTestFanouts(g), 2)
	if len(s.OutputIDs) != 2 || s.OutputIDs[0] == s.OutputIDs[1] {
		t.Fatal(s.OutputIDs)
	}
	if s.refs[x.NodeID()] != 2 {
		t.Fatal("repeated output reference not represented", s.refs)
	}
}

func TestDominatorAfterOutputRewire(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	old := g.AND(a, b)
	co := g.AddOutput(old)
	replacement := g.AND(a, c)
	Aig_ObjPatchFanin0(g, co, replacement)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	s := Aig_ManComputeDomsNodes(domTestFanouts(g), 1)
	if len(s.Dominators[b.NodeID()]) != 0 {
		t.Fatal("deleted cone remained active")
	}
	found := false
	for _, d := range s.Dominators[a.NodeID()] {
		if reflect.DeepEqual(d.Nodes, []int{replacement.NodeID()}) {
			found = true
		}
	}
	if !found {
		t.Fatal("reverse numeric IDs used as topological order", s.DominatorSets(a.NodeID()))
	}
}
