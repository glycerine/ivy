package gia

import (
	"bytes"
	"testing"
)

type partitionReviewCopy struct {
	name string
	copy func(*Gia_Man_t) *Gia_Man_t
	regs []int32
}

func partitionReviewCopies() []partitionReviewCopy {
	return []partitionReviewCopy{
		{"normalize", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupNormalize(g, 0) }, []int32{0, 1}},
		{"normalize_hash", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupNormalize(g, 1) }, []int32{0, 1}},
		{"levelized", Gia_ManDupLevelized, []int32{0, 1}},
		{"ordered_dfs", Gia_ManDupOrderDfs, []int32{0, 1}},
		{"reverse_dfs", Gia_ManDupOrderDfsReverse, []int32{0, 1}},
		{"aiger_order", Gia_ManDupOrderAiger, []int32{0, 1}},
		{"dfs", Gia_ManDupDfs, []int32{0, 1}},
		{"dfs2", Gia_ManDupDfs2, []int32{0, 1}},
		{"classes", func(g *Gia_Man_t) *Gia_Man_t {
			g.pReprsOld = make([]int32, g.nObjs)
			for i := range g.pReprsOld {
				g.pReprsOld[i] = -1
			}
			return Gia_ManDupDfsClasses(g)
		}, []int32{0, 1}},
		{"abstract_none", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(), permutation(-1, -1)) }, []int32{0, 1}},
		{"abstract_first", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(0), permutation(0, -1)) }, []int32{1}},
		{"abstract_second", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(1), permutation(-1, 0)) }, []int32{0}},
		{"abstract_both_reversed", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(1, 0), permutation(1, 0)) }, nil},
		{"cut", Gia_ManDupReplaceCut, []int32{0, 1}},
	}
}

func partitionReviewGraph(signs int, typed, layout bool) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, c, r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	var x, y, z int32
	if typed {
		x = Gia_ManAppendXorReal(g, a^phase(0), b^phase(1))
		g.EnableMuxes()
		y = Gia_ManAppendMuxReal(g, r0^phase(2), x, c^phase(3))
		z = Gia_ManAppendBuf(g, Gia_ManAppendAnd(g, r1, c)^1)
	} else {
		x = Gia_ManAppendAnd(g, a^phase(0), b^phase(1))
		y = Gia_ManAppendAnd(g, x, r0^phase(2))
		z = Gia_ManAppendAnd(g, r1, c)
	}
	for _, lit := range []int32{x, y, y ^ 1, z, a ^ 1, b, r1 ^ 1} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, 2)
	g.SetIdentity("partition", "partition.aig")
	if layout {
		g = Gia_ManDupOrderDfsReverse(g)
	}
	return g
}

func TestFullOutputCopiesRetainBadPartition(t *testing.T) {
	for _, tc := range partitionReviewCopies() {
		t.Run(tc.name, func(t *testing.T) {
			src := partitionReviewGraph(0, false, false)
			if err := src.SetAigerProperties(2, 2); err != nil {
				t.Fatal(err)
			}
			out := tc.copy(src)
			if out.nConstrs != 2 || out.aigerBadOutputs != 2 {
				t.Errorf("copied roles: constraints%d bad%d, want 2/2", out.nConstrs, out.aigerBadOutputs)
			}
			data, err := MarshalAiger(out, AigerWriteOptions{OmitExtensions: true, OmitSymbols: true})
			if err != nil {
				t.Fatal(err)
			}
			restored, err := ParseAiger(data, AigerReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if restored.nConstrs != 2 || restored.aigerBadOutputs != 2 {
				t.Errorf("AIGER round trip lost bad outputs: header %s", bytes.SplitN(data, []byte{'\n'}, 2)[0])
			}
		})
	}
}
