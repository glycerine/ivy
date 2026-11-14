package fra

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
)

// ABC_REFERENCE opts into a comparison against an independently built ABC.
// ABC's network wrapper may clean registers, so compare output traces rather
// than requiring the internal state counts to match.
func TestABCInductionDifferential(t *testing.T) {
	binary := os.Getenv("ABC_REFERENCE")
	if binary == "" {
		t.Skip("set ABC_REFERENCE to an ABC executable")
	}
	dir := t.TempDir()
	random := rand.New(rand.NewSource(20261113))
	compared, unknown := 0, 0
	for trial := 0; trial < 12; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 8; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		for i := 0; i < 5; i++ {
			e := edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				e = e.Not()
			}
			g.AddOutput(e)
		}
		aig.Aig_ManSetRegNum(g, 3)
		aig.Aig_ManCleanup(g)
		input := filepath.Join(dir, "input.aig")
		if err := ioa.Ioa_WriteAiger(g, input, 0, 0); err != nil {
			t.Fatal(err)
		}
	configs:
		for _, config := range []struct {
			args                                                      string
			frames, partition, latch, rewrite, prefix, levels, onehot int32
		}{
			{"-F 1", 1, 0, 0, 0, 0, 0, 0}, {"-F 2", 2, 0, 0, 0, 0, 0, 0}, {"-F 1 -P 2 -l", 1, 2, 1, 0, 0, 0, 0}, {"-F 1 -r", 1, 0, 0, 1, 0, 0, 0},
			{"-F 1 -N 1", 1, 0, 0, 0, 1, 0, 0}, {"-F 1 -L 1", 1, 0, 0, 0, 0, 1, 0}, {"-F 1 -t", 1, 0, 0, 0, 0, 0, 1},
		} {
			output := filepath.Join(dir, "output.aig")
			os.Remove(output)
			command := fmt.Sprintf("read_aiger %s; strash; ssweep %s; write_aiger %s", input, config.args, output)
			log, err := exec.Command(binary, "-c", command).CombinedOutput()
			if err != nil {
				t.Fatalf("trial %d %s: %v: %s", trial, config.args, err, log)
			}
			model, err := ioa.Ioa_ReadAiger(output, 1)
			if err != nil {
				t.Fatalf("trial %d %s: %v: %s", trial, config.args, err, log)
			}
			for _, l := range model.Latches {
				if l.Init != 0 && l.Init != 1 {
					unknown++
					continue configs
				}
			}
			reduced := Fra_FraigInduction(aig.Aig_ManDupOrdered(g), &Fra_Ssw_t{NFramesK: config.frames, NPartSize: config.partition, FLatchCorr: config.latch, FRewrite: config.rewrite, NFramesP: config.prefix, NMaxLevs: config.levels, FUse1Hot: config.onehot})
			if reduced == nil {
				t.Fatalf("trial %d %s: Go reduction failed", trial, config.args)
			}
			compared++
			for seq := 0; seq < 64; seq++ {
				originalState, goState, cState := make([]bool, 3), make([]bool, reduced.RegisterCount()), make([]bool, len(model.Latches))
				for i, latch := range model.Latches {
					cState[i] = latch.Init == 1
				}
				for frame := 0; frame < 3; frame++ {
					inputs := []bool{seq&(1<<(2*frame)) != 0, seq&(1<<(2*frame+1)) != 0}
					orig, err := g.Evaluate(append(append([]bool{}, inputs...), originalState...))
					if err != nil {
						t.Fatal(err)
					}
					got, err := reduced.Evaluate(append(append([]bool{}, inputs...), goState...))
					if err != nil {
						t.Fatal(err)
					}
					ref, next := fraEvalReference(model, inputs, cState)
					for i := 0; i < 2; i++ {
						if orig[i] != got[i] || orig[i] != ref[i] {
							t.Fatalf("trial=%d %s sequence=%d frame=%d output=%d original=%v go=%v C=%v", trial, config.args, seq, frame, i, orig, got, ref)
						}
					}
					copy(originalState, orig[len(orig)-len(originalState):])
					copy(goState, got[len(got)-len(goState):])
					copy(cState, next)
				}
			}
		}
	}
	if compared == 0 {
		t.Fatal("no ABC output had a defined initial state")
	}
	t.Logf("compared %d cases, skipped %d with unknown C initial state", compared, unknown)
}

func fraEvalReference(m *ioa.Model, inputs, state []bool) ([]bool, []bool) {
	values := make([]bool, int(m.MaxVar)+1)
	for i, l := range m.Inputs {
		values[l>>1] = inputs[i]
	}
	for i, l := range m.Latches {
		values[l.Literal>>1] = state[i]
	}
	value := func(l uint32) bool { return values[l>>1] != (l&1 != 0) }
	for _, n := range m.ANDs {
		values[n.Literal>>1] = value(n.Left) && value(n.Right)
	}
	outputs := make([]bool, len(m.Outputs))
	for i, l := range m.Outputs {
		outputs[i] = value(l)
	}
	next := make([]bool, len(m.Latches))
	for i, l := range m.Latches {
		next[i] = value(l.Next)
	}
	return outputs, next
}
