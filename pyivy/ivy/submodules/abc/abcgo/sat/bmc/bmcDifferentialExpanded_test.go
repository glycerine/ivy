package bmc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

// abcFixture uses AIGER literals: 0 is false, 1 true, and bit zero
// complements each nonconstant signal. AND gates are topologically ordered.
type abcFixture struct {
	name       string
	pis, regs  int
	ands       [][2]int
	next, outs []int
}

func (f abcFixture) aiger() []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "aig %d %d %d %d %d\n", f.pis+f.regs+len(f.ands), f.pis, f.regs, len(f.outs), len(f.ands))
	for _, lit := range f.next {
		fmt.Fprintf(&b, "%d\n", lit)
	}
	for _, lit := range f.outs {
		fmt.Fprintf(&b, "%d\n", lit)
	}
	put := func(x int) {
		for x >= 128 {
			b.WriteByte(byte(x&127) | 128)
			x >>= 7
		}
		b.WriteByte(byte(x))
	}
	for i, pair := range f.ands {
		lhs := 2 * (1 + f.pis + f.regs + i)
		a, c := pair[0], pair[1]
		if a < c {
			a, c = c, a
		}
		if a >= lhs || c > a {
			panic("invalid AIGER gate")
		}
		put(lhs - a)
		put(a - c)
	}
	return b.Bytes()
}
func (f abcFixture) graph() *aig.Graph {
	g := aig.NewGraph()
	edges := []aig.Edge{g.Const0()}
	for i := 0; i < f.pis+f.regs; i++ {
		edges = append(edges, g.NewInput())
	}
	lit := func(n int) aig.Edge {
		e := edges[n>>1]
		if n&1 != 0 {
			e = e.Not()
		}
		return e
	}
	for _, a := range f.ands {
		edges = append(edges, g.AND(lit(a[0]), lit(a[1])))
	}
	for _, v := range f.outs {
		g.AddOutput(lit(v))
	}
	for _, v := range f.next {
		g.AddOutput(lit(v))
	}
	aig.Aig_ManSetRegNum(g, int32(f.regs))
	return g
}

// step is deliberately separate from GIA/AIG simulation and uses only the
// fixture literals, so replay checks the returned witness independently.
func (f abcFixture) step(state int, input int) ([]bool, int) {
	vals := make([]bool, 1+f.pis+f.regs+len(f.ands))
	for i := 0; i < f.pis; i++ {
		vals[1+i] = input&(1<<i) != 0
	}
	for i := 0; i < f.regs; i++ {
		vals[1+f.pis+i] = state&(1<<i) != 0
	}
	lit := func(n int) bool { return vals[n>>1] != (n&1 != 0) }
	for i, p := range f.ands {
		vals[1+f.pis+f.regs+i] = lit(p[0]) && lit(p[1])
	}
	outs := make([]bool, len(f.outs))
	for i, v := range f.outs {
		outs[i] = lit(v)
	}
	next := 0
	for i, v := range f.next {
		if lit(v) {
			next |= 1 << i
		}
	}
	return outs, next
}
func (f abcFixture) replay(c *util.Abc_Cex_t) bool {
	if c == nil || c.NRegs != int32(f.regs) || c.NPis != int32(f.pis) || c.IPo < 0 || int(c.IPo) >= len(f.outs) {
		return false
	}
	state := 0
	for i := 0; i < f.regs; i++ {
		if util.Abc_InfoHasBit(c.PData, int32(i)) {
			state |= 1 << i
		}
	}
	if state != 0 {
		return false
	}
	for frame := int32(0); frame <= c.IFrame; frame++ {
		input := 0
		for i := 0; i < f.pis; i++ {
			if util.Abc_InfoHasBit(c.PData, int32(f.regs)+frame*int32(f.pis)+int32(i)) {
				input |= 1 << i
			}
		}
		outs, next := f.step(state, input)
		if frame == c.IFrame {
			return outs[c.IPo]
		}
		state = next
	}
	return false
}
func (f abcFixture) earliest(bound int) (int, int) {
	states := map[int]bool{0: true}
	for frame := 0; frame < bound; frame++ {
		nextStates := map[int]bool{}
		first := -1
		for state := range states {
			for input := 0; input < 1<<f.pis; input++ {
				outs, next := f.step(state, input)
				for o, hit := range outs {
					if hit && (first < 0 || o < first) {
						first = o
					}
				}
				nextStates[next] = true
			}
		}
		if first >= 0 {
			return first, frame
		}
		states = nextStates
	}
	return -1, -1
}
func TestBMCAgainstCABCExpanded(t *testing.T) {
	binary := os.Getenv("ABC_BINARY")
	if binary == "" {
		binary = filepath.Join("..", "..", "..", "abc")
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(binary); err != nil || fi.IsDir() || fi.Mode()&0111 == 0 {
		t.Skip("C ABC executable unavailable; set ABC_BINARY")
	}
	fixtures := []abcFixture{
		{name: "two_inputs_complemented_latch", pis: 2, regs: 2, ands: [][2]int{{2, 5}}, next: []int{10, 6}, outs: []int{8}},
		{name: "second_output_delay", pis: 2, regs: 2, ands: [][2]int{{2, 4}}, next: []int{10, 6}, outs: []int{0, 8}},
		{name: "two_gate_safe_invariant", pis: 2, regs: 1, ands: [][2]int{{2, 4}, {6, 8}}, next: []int{10}, outs: []int{6}},
		{name: "complemented_output", pis: 2, regs: 1, ands: [][2]int{{2, 5}}, next: []int{8}, outs: []int{7}},
		{name: "two_step_conjunctive_failure", pis: 2, regs: 2, ands: [][2]int{{2, 4}, {6, 8}}, next: []int{10, 6}, outs: []int{12}},
	}
	re := regexp.MustCompile(`Output\s+([0-9]+).*asserted in frame\s+([0-9]+)`)
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			expectedOutput, expectedFrame := f.earliest(4)
			file := filepath.Join(t.TempDir(), "fixture.aig")
			if err := os.WriteFile(file, f.aiger(), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, binary, "-c", fmt.Sprintf("read_aiger %s; bmc3 -F 4", file)).CombinedOutput()
			if err != nil {
				t.Fatalf("C ABC: %v\n%s", err, out)
			}
			text := string(out)
			match := re.FindStringSubmatch(text)
			output, frame := -1, -1
			if match != nil {
				output, _ = strconv.Atoi(match[1])
				frame, _ = strconv.Atoi(match[2])
			} else if !strings.Contains(text, "Explored all reachable states") && !strings.Contains(text, "No output asserted") {
				t.Fatalf("unrecognized C result: %s", text)
			}
			if output != expectedOutput || frame != expectedFrame {
				t.Fatalf("C got output=%d frame=%d, oracle output=%d frame=%d: %s", output, frame, expectedOutput, expectedFrame, text)
			}
			graph := f.graph()
			p := Saig_ParBmc_t{NFramesMax: 4}
			result := Saig_ManBmcScalable(graph, &p)
			if frame < 0 {
				if result != 1 {
					t.Fatalf("Go=%d want safe", result)
				}
				return
			}
			cex := graph.SequentialCounterexample()
			if result != 0 || cex == nil || int(cex.IPo) != output || int(cex.IFrame) != frame || !f.replay(cex) {
				t.Fatalf("Go result=%d witness=%+v, C output=%d frame=%d", result, cex, output, frame)
			}
		})
	}
}
