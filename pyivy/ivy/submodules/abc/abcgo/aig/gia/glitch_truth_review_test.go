package gia

import (
	"math/rand"
	"testing"
)

func TestGlitchScalarAndPackedTruthTables(t *testing.T) {
	tables := 0
	for inputs := 0; inputs <= 6; inputs++ {
		limit := 256
		if inputs < 4 {
			limit = 1 << (1 << inputs)
		}
		r := rand.New(rand.NewSource(int64(inputs + 991)))
		for index := 0; index < limit; index++ {
			truth := r.Uint64()
			if inputs < 4 {
				truth = uint64(index)
			}
			pins := make([]int32, inputs)
			for i := range pins {
				pins[i] = int32(i)
			}
			g := glitchReviewGraph{cis: inputs, nodes: []glitchReviewNode{{pins, truth}}, outputs: []int32{int32(inputs)}}
			p := g.build()
			node := p.objects[inputs]
			for assignment := 0; assignment < 1<<inputs; assignment++ {
				for i := 0; i < inputs; i++ {
					p.objects[i].fPhase = uint32(assignment >> i & 1)
					p.objects[i].fPhase2 = 1 - p.objects[i].fPhase
				}
				if got := Gli_NodeComputeValue(node); got != int32(truth>>assignment&1) {
					t.Fatalf("%d-input table%x assignment%d value%d", inputs, truth, assignment, got)
				}
				inverse := assignment ^ ((1 << inputs) - 1)
				if got := Gli_NodeComputeValue2(node); got != int32(truth>>inverse&1) {
					t.Fatal("second scalar phase table mismatch")
				}
			}
			for base := 0; base < 64; base += 32 {
				var expected uint32
				for i := 0; i < inputs; i++ {
					p.objects[i].uSimInfo = 0
				}
				for bit := 0; bit < 32; bit++ {
					assignment := (base + bit) & ((1 << inputs) - 1)
					for i := 0; i < inputs; i++ {
						p.objects[i].uSimInfo |= uint32(assignment>>i&1) << bit
					}
					expected |= uint32(truth>>assignment&1) << bit
				}
				if got := Gli_ManSimulateSeqNode(p, node); got != expected {
					t.Fatalf("%d-input table%x base%d word%08x want%08x", inputs, truth, base, got, expected)
				}
			}
			Gli_ManStop(p)
			tables++
		}
	}
	t.Logf("checked %d truth tables, including every Boolean function up to three inputs", tables)
}

// C's bitfield/six-element scratch do not safely support these arities.
// Preserve and independently test the existing Go multiword safety extension.
func TestGlitchMultiwordTruthSafetyExtension(t *testing.T) {
	r := rand.New(rand.NewSource(3321))
	for _, inputs := range []int{7, 8, 12, 16} {
		p := Gli_ManAlloc(int32(inputs+1), 0, int32(inputs))
		pins := Vec_IntAlloc(inputs)
		for i := 0; i < inputs; i++ {
			Vec_IntPush(pins, Gli_ManCreateCi(p, 1))
		}
		truth := make([]word, 1<<(inputs-6))
		for i := range truth {
			truth[i] = r.Uint64()
		}
		id := Gli_ManCreateNode(p, pins, 0, truth)
		Gli_ManFinalize(p)
		for iteration := 0; iteration < 16; iteration++ {
			for i := 0; i < inputs; i++ {
				p.objects[i].uSimInfo = r.Uint32()
			}
			var expected uint32
			for bit := 0; bit < 32; bit++ {
				address := 0
				for i := 0; i < inputs; i++ {
					address |= int(p.objects[i].uSimInfo>>bit&1) << i
				}
				expected |= uint32(truth[address/64]>>uint(address%64)&1) << bit
			}
			if got := Gli_ManSimulateSeqNode(p, p.objects[id]); got != expected {
				t.Fatalf("inputs%d packed%08x want%08x", inputs, got, expected)
			}
		}
		Gli_ManStop(p)
	}
}

func glitchReviewEvaluateWords(g glitchReviewGraph, inputs []uint32) []uint32 {
	values := make([]uint32, g.cis+len(g.nodes))
	copy(values, inputs)
	// Evaluate each lane independently, without the simulator's packed operations.
	for bit := 0; bit < 32; bit++ {
		for i, node := range g.nodes {
			assignment := 0
			for k, id := range node.inputs {
				if values[id]&(uint32(1)<<bit) != 0 {
					assignment += 1 << k
				}
			}
			if node.truth&(uint64(1)<<assignment) != 0 {
				values[g.cis+i] |= uint32(1) << bit
			}
		}
	}
	return values
}
func TestGlitchSequentialPrefixIndependent(t *testing.T) {
	for index, g := range glitchReviewGraphs() {
		for _, frames := range []int{0, 1, 2, 16} {
			Gia_ManRandom(1)
			inputs := make([]uint32, g.cis)
			var values []uint32
			for i := 0; i < g.cis-g.regs; i++ {
				inputs[i] = Gia_ManRandom(0)
			}
			for frame := 0; frame < frames; frame++ {
				values = glitchReviewEvaluateWords(g, inputs)
				next := make([]uint32, g.cis)
				for i := 0; i < g.cis-g.regs; i++ {
					next[i] = Gia_ManRandom(0)
				}
				for i := 0; i < g.regs; i++ {
					next[g.cis-g.regs+i] = values[g.outputs[len(g.outputs)-g.regs+i]]
				}
				inputs = next
			}
			nextRandom := Gia_ManRandom(0)
			p := g.build()
			Gia_ManRandom(1)
			Gli_ManSimulateSeqPref(p, int32(frames))
			for i, want := range inputs {
				if p.pSimInfoPrev[i] != want || p.objects[i].uSimInfo != want {
					t.Fatalf("graph%d frames%d CI%d", index, frames, i)
				}
			}
			for i := range g.nodes {
				var want uint32
				if frames > 0 {
					want = values[g.cis+i]
				}
				if p.objects[g.cis+i].uSimInfo != want {
					t.Fatalf("graph%d frames%d node%d", index, frames, i)
				}
			}
			if got := Gia_ManRandom(0); got != nextRandom {
				t.Fatal("prefix RNG schedule changed")
			}
			for _, bit := range []int32{0, 1, 15, 31} {
				Gli_ManSetDataSaved(p, bit)
				expected := glitchReviewEvaluateWords(g, inputs)
				for i, o := range p.objects[:g.cis+len(g.nodes)] {
					want := expected[i] >> bit & 1
					if o.fPhase != want || o.fPhase2 != want {
						t.Fatalf("graph%d frames%d saved bit%d obj%d", index, frames, bit, i)
					}
				}
			}
			Gli_ManStop(p)
		}
	}
}
