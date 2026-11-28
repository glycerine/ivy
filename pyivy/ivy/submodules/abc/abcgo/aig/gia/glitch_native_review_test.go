package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type glitchReviewNode struct {
	inputs []int32
	truth  uint64
}
type glitchReviewGraph struct {
	cis, regs int
	nodes     []glitchReviewNode
	outputs   []int32
}

func (g glitchReviewGraph) fanouts() []int32 {
	counts := make([]int32, g.cis+len(g.nodes)+len(g.outputs))
	for _, n := range g.nodes {
		for _, id := range n.inputs {
			counts[id]++
		}
	}
	for _, id := range g.outputs {
		counts[id]++
	}
	return counts
}
func (g glitchReviewGraph) build() *Gli_Man_t {
	counts := g.fanouts()
	var edges int32
	for _, n := range counts {
		edges += n
	}
	p := Gli_ManAlloc(int32(len(counts)), int32(g.regs), edges)
	for i := 0; i < g.cis; i++ {
		Gli_ManCreateCi(p, counts[i])
	}
	for i, n := range g.nodes {
		fanins := Vec_IntAlloc(len(n.inputs))
		for _, id := range n.inputs {
			Vec_IntPush(fanins, id)
		}
		Gli_ManCreateNode(p, fanins, counts[g.cis+i], []word{n.truth})
	}
	for _, id := range g.outputs {
		Gli_ManCreateCo(p, id)
	}
	return p
}
func (g glitchReviewGraph) write(t *testing.T, dir string) string {
	t.Helper()
	var text strings.Builder
	fmt.Fprintf(&text, "%d %d %d %d\n", g.cis, g.regs, len(g.nodes), len(g.outputs))
	for _, n := range g.nodes {
		fmt.Fprintf(&text, "%d %016x", len(n.inputs), n.truth)
		for _, id := range n.inputs {
			fmt.Fprintf(&text, " %d", id)
		}
		text.WriteByte('\n')
	}
	for _, id := range g.outputs {
		fmt.Fprintf(&text, "%d\n", id)
	}
	path := filepath.Join(dir, "graph.txt")
	if err := os.WriteFile(path, []byte(text.String()), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func glitchReviewState(w io.Writer, p *Gli_Man_t, name string) {
	fmt.Fprintf(w, "state %s %d %d %d %d %d %d\n", name, Gli_ManCiNum(p), Gli_ManCoNum(p), Gli_ManPiNum(p), Gli_ManPoNum(p), Gli_ManNodeNum(p), Gli_ManRegNum(p))
	for _, o := range p.objects {
		fmt.Fprintf(w, "obj%d %d %d %d %d %d %d %08x %d %d in", o.Handle, o.fTerm, o.fPhase, o.fPhase2, o.fMark, o.nFanins, o.nFanouts, o.uSimInfo, o.nSwitches, o.nGlitches)
		for _, f := range o.fanins {
			fmt.Fprintf(w, " %d", f.Handle)
		}
		fmt.Fprint(w, " out")
		for _, f := range o.fanouts {
			fmt.Fprintf(w, " %d", f.Handle)
		}
		fmt.Fprintln(w)
	}
	for _, v := range []struct {
		name string
		ids  *Vec_Int_t
	}{{"changed", p.vCisChanged}, {"affected", p.vAffected}, {"frontier", p.vFrontier}} {
		fmt.Fprint(w, v.name)
		for _, id := range Vec_IntArray(v.ids) {
			fmt.Fprintf(w, " %d", id)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprint(w, "saved")
	for _, value := range p.pSimInfoPrev {
		fmt.Fprintf(w, " %08x", value)
	}
	fmt.Fprintln(w)
}
func glitchReviewRun(p *Gli_Man_t, mode, steps, prefix int, prob float32) string {
	var w strings.Builder
	if mode == 3 {
		Gli_ManSwitchesAndGlitches(p, int32(steps), prob, 0)
		glitchReviewState(&w, p, "driver")
	} else {
		Gli_ManFinalize(p)
		glitchReviewState(&w, p, "initial")
		Gia_ManRandom(1)
		if mode == 0 {
			for i := 0; i < steps; i++ {
				Gli_ManSetPiRandom(p, prob)
				glitchReviewState(&w, p, "inputs")
				Gli_ManSwitching(p)
				glitchReviewState(&w, p, "switching")
				Gli_ManGlitching(p)
				glitchReviewState(&w, p, "glitching")
			}
		} else {
			Gli_ManSimulateSeqPref(p, int32(prefix))
			glitchReviewState(&w, p, "prefix")
			for bit := int32(0); bit < 32; bit++ {
				Gli_ManSetDataSaved(p, bit)
				glitchReviewState(&w, p, "saved-bit")
				for j := 0; j < steps; j++ {
					if mode == 1 {
						Gli_ManSetPiRandomSeq(p, prob)
					} else {
						for k := int32(0); k < Gli_ManCiNum(p); k++ {
							Gli_ManCi(p, k).uSimInfo ^= 0x9e3779b9 * uint32(int(k)+j+1)
						}
						Gli_ManSetPiFromSaved(p, bit)
					}
					glitchReviewState(&w, p, "inputs")
					Gli_ManSwitching(p)
					glitchReviewState(&w, p, "switching")
					Gli_ManGlitching(p)
					glitchReviewState(&w, p, "glitching")
				}
			}
		}
	}
	fmt.Fprintf(&w, "random %08x\n", Gia_ManRandom(0))
	return w.String()
}
func glitchReviewGraphs() []glitchReviewGraph {
	graphs := []glitchReviewGraph{
		{cis: 0, nodes: []glitchReviewNode{{truth: 1}}, outputs: []int32{0}},
		{cis: 1, nodes: []glitchReviewNode{{[]int32{0}, 0xaaaaaaaaaaaaaaaa}, {[]int32{1}, 0x5555555555555555}, {[]int32{0, 2}, 0x8888888888888888}}, outputs: []int32{3}},
	}
	for seed := int64(1); seed <= 24; seed++ {
		r := rand.New(rand.NewSource(seed))
		g := glitchReviewGraph{cis: 1 + int(seed%6), regs: int(seed % 3)}
		if g.regs > g.cis {
			g.regs = g.cis
		}
		for i := 0; i < 10; i++ {
			n := glitchReviewNode{truth: r.Uint64()}
			for j := 0; j < int((seed+int64(i))%7); j++ {
				n.inputs = append(n.inputs, int32(r.Intn(g.cis+i)))
			}
			g.nodes = append(g.nodes, n)
		}
		for i := 0; i < 3+g.regs; i++ {
			g.outputs = append(g.outputs, int32(r.Intn(g.cis+len(g.nodes))))
		}
		graphs = append(graphs, g)
	}
	return graphs
}
func TestGlitchNativeStateAndRNG(t *testing.T) {
	exe := compileGiaSourceOracle(t, "glitch_review.c")
	dir := t.TempDir()
	count := 0
	probabilities := []float32{1.0 / 65536, 0.17, 0.5, 0.99998474}
	for index, g := range glitchReviewGraphs() {
		file := g.write(t, dir)
		for mode := 0; mode < 4; mode++ {
			prob := probabilities[(index+mode)%len(probabilities)]
			steps := 1 + index%4
			prefix := index % 5
			if mode == 3 {
				steps = []int{0, 1, 31, 32, 33, 65}[index%6]
			}
			source, err := testoracle.Command(t, exe, file, strconv.Itoa(mode), strconv.Itoa(steps), strconv.Itoa(prefix), strconv.FormatFloat(float64(prob), 'g', 9, 32), "0").CombinedOutput()
			if err != nil {
				t.Fatalf("native graph%d mode%d: %v %s", index, mode, err, source)
			}
			p := g.build()
			got := glitchReviewRun(p, mode, steps, prefix, prob)
			Gli_ManStop(p)
			if string(source) != got {
				t.Logf("graph%d mode%d steps%d prefix%d probability%g", index, mode, steps, prefix, prob)
				esopReviewCompare(t, source, got)
			}
			count++
		}
	}
	t.Logf("compared %d complete graph/state/RNG traces", count)
}
func TestGlitchNativeObjectReport(t *testing.T) {
	exe := compileGiaSourceOracle(t, "glitch_review.c")
	g := glitchReviewGraphs()[1]
	file := g.write(t, t.TempDir())
	source, err := testoracle.Command(t, exe, file, "4", "0", "0", "0.5", "0").CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v %s", err, source)
	}
	header, report, ok := strings.Cut(string(source), "\n")
	if !ok {
		t.Fatal("missing layout header")
	}
	var headerWords int
	if _, err := fmt.Sscanf(header, "headerWords %d", &headerWords); err != nil {
		t.Fatal(err)
	}
	counts := g.fanouts()
	offsets := map[int]int{}
	offset := 0
	for id, count := range counts {
		offsets[offset] = id
		faninCount := 0
		if id >= g.cis && id < g.cis+len(g.nodes) {
			faninCount = len(g.nodes[id-g.cis].inputs)
		} else if id >= g.cis+len(g.nodes) {
			faninCount = 1
		}
		offset += headerWords + faninCount + int(count)
	}
	report = regexp.MustCompile(`[0-9]+`).ReplaceAllStringFunc(report, func(s string) string {
		id, _ := strconv.Atoi(s)
		ordinal, ok := offsets[id]
		if !ok {
			t.Fatalf("unknown packed source handle%d", id)
		}
		return strconv.Itoa(ordinal)
	})
	p := g.build()
	defer Gli_ManStop(p)
	got := hashReviewStdout(t, func() { Gli_ManPrintObjects(p) })
	if report != got {
		t.Fatalf("Go object report%q; normalized C%q", got, report)
	}
}

func TestGlitchNativeRandomWordUpdates(t *testing.T) {
	exe := compileGiaSourceOracle(t, "glitch_review.c")
	for _, prob := range []float32{0, 1.0 / 65536, 0.17, 0.49999997, 0.5, 0.50000006, 0.99998474, 1} {
		for _, input := range []uint32{0, 1, 0x80000000, 0x55555555, 0xffffffff} {
			source, err := testoracle.Command(t, exe, "unused", "5", fmt.Sprintf("0x%x", input), "33", strconv.FormatFloat(float64(prob), 'g', 9, 32), "0").CombinedOutput()
			if err != nil {
				t.Fatalf("native %g/%08x: %v %s", prob, input, err, source)
			}
			Gia_ManRandom(1)
			value := input
			var got strings.Builder
			for i := 0; i < 33; i++ {
				value = Gli_ManUpdateRandomInput(value, prob)
				fmt.Fprintf(&got, "word %08x\n", value)
			}
			fmt.Fprintf(&got, "random %08x\n", Gia_ManRandom(0))
			if string(source) != got.String() {
				t.Logf("probability%g initial%08x", prob, input)
				esopReviewCompare(t, source, got.String())
			}
		}
	}
}

func TestGlitchNativeVerboseReport(t *testing.T) {
	exe := compileGiaSourceOracle(t, "glitch_review.c")
	g := glitchReviewGraph{cis: 1, nodes: []glitchReviewNode{{[]int32{0}, 0xaaaaaaaaaaaaaaaa}}, outputs: []int32{1}}
	source, err := testoracle.Command(t, exe, g.write(t, t.TempDir()), "3", "3", "0", "0.5", "1").CombinedOutput()
	if err != nil {
		t.Fatalf("native report: %v %s", err, source)
	}
	want, _, ok := strings.Cut(string(source), "state driver")
	if !ok {
		t.Fatal("missing native state")
	}
	p := g.build()
	defer Gli_ManStop(p)
	got := hashReviewStdout(t, func() { Gli_ManSwitchesAndGlitches(p, 3, 0.5, 1) })
	clock := regexp.MustCompile(`Time = *[0-9]+\.[0-9]{2} sec`)
	if clock.ReplaceAllString(want, "Time = <clock> sec") != clock.ReplaceAllString(got, "Time = <clock> sec") {
		t.Fatalf("Go verbose report%q; C%q", got, want)
	}
	field := regexp.MustCompile(`Time =( *[0-9]+\.[0-9]{2}) sec`).FindStringSubmatch(got)
	if len(field) != 2 || len(field[1]) != 9 {
		t.Fatal("source clock field width not preserved")
	}
}
