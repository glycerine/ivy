package bmc

import (
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

// TestBMCAgainstCABC compares independently constructed Go graphs to tiny
// binary AIGER fixtures consumed by ABC's original bmc3 implementation. Set
// ABC_BINARY to an ABC executable; a local source-tree build is also detected.
func TestBMCAgainstCABC(t *testing.T) {
	binary := os.Getenv("ABC_BINARY")
	if binary == "" {
		binary = filepath.Join("..", "..", "..", "abc")
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(absolute); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("C ABC executable unavailable; set ABC_BINARY to enable differential test")
	}
	tests := []struct {
		name, aiger string
		build       func() *aig.Graph
		wantFrame   int32
		replay      func(*util.Abc_Cex_t) bool
	}{
		{"two_register_delay", "aig 2 0 2 1 0\n1\n2\n4\n", func() *aig.Graph {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(b)
			g.AddOutput(g.Const1())
			g.AddOutput(a)
			aig.Aig_ManSetRegNum(g, 2)
			return g
		}, 2, func(c *util.Abc_Cex_t) bool {
			a, b := false, false
			for f := int32(0); f <= c.IFrame; f++ {
				if f == c.IFrame {
					return b
				}
				a, b = true, a
			}
			return false
		}},
		{"input_to_register", "aig 2 1 1 1 0\n2\n4\n", func() *aig.Graph {
			g := aig.NewGraph()
			x, r := g.NewInput(), g.NewInput()
			g.AddOutput(r)
			g.AddOutput(x)
			aig.Aig_ManSetRegNum(g, 1)
			return g
		}, 1, func(c *util.Abc_Cex_t) bool {
			r := false
			for f := int32(0); f <= c.IFrame; f++ {
				if f == c.IFrame {
					return r
				}
				r = util.Abc_InfoHasBit(c.PData, c.NRegs+f*c.NPis)
			}
			return false
		}},
		{"stable_zero", "aig 1 0 1 1 0\n2\n2\n", func() *aig.Graph {
			g := aig.NewGraph()
			r := g.NewInput()
			g.AddOutput(r)
			g.AddOutput(r)
			aig.Aig_ManSetRegNum(g, 1)
			return g
		}, -1, nil},
	}
	re := regexp.MustCompile(`Output\s+([0-9]+).*asserted in frame\s+([0-9]+)`)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "fixture.aig")
			if err := os.WriteFile(path, []byte(tc.aiger), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			output, err := exec.CommandContext(ctx, absolute, "-c", fmt.Sprintf("read_aiger %s; bmc3 -F 4", path)).CombinedOutput()
			if err != nil {
				t.Fatalf("C ABC failed: %v\n%s", err, output)
			}
			text := string(output)
			match := re.FindStringSubmatch(text)
			cframe := int32(-1)
			if match != nil {
				outputID, _ := strconv.Atoi(match[1])
				frame, _ := strconv.Atoi(match[2])
				if outputID != 0 {
					t.Fatalf("unexpected C output: %s", text)
				}
				cframe = int32(frame)
			} else if !strings.Contains(text, "Explored all reachable states") && !strings.Contains(text, "No output asserted") {
				t.Fatalf("could not identify C result: %s", text)
			}
			graph := tc.build()
			params := Saig_ParBmc_t{NFramesMax: 4}
			result := Saig_ManBmcScalable(graph, &params)
			if cframe != tc.wantFrame {
				t.Fatalf("C frame=%d want=%d: %s", cframe, tc.wantFrame, text)
			}
			if cframe < 0 {
				if result != 1 {
					t.Fatalf("Go result=%d, C safe", result)
				}
				return
			}
			cex := graph.SequentialCounterexample()
			if result != 0 || cex == nil || cex.IFrame != cframe || cex.IPo != 0 || !tc.replay(cex) {
				t.Fatalf("C frame=%d Go result=%d witness=%+v", cframe, result, cex)
			}
		})
	}
}
