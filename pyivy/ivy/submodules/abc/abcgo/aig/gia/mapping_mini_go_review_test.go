package gia

import (
	"fmt"
	"slices"
	"testing"
)

type mappingReviewMini struct {
	inputs       int
	leaves, mini []int32
	hash         int32
}

func (c mappingReviewMini) want(mask int) bool {
	value := func(lit int32) bool {
		v := false
		if lit>>1 > 0 {
			v = mask>>uint((lit>>1)-1)&1 != 0
		}
		return v != (lit&1 != 0)
	}
	if len(c.mini) == 1 {
		return c.mini[0] != 0
	}
	if len(c.mini) == 2 {
		return value(c.leaves[0]) != (c.mini[1] != 0)
	}
	values := make([]bool, 0, len(c.leaves)+len(c.mini)/2)
	for _, leaf := range c.leaves {
		values = append(values, value(leaf))
	}
	for i := 0; i < len(c.mini)-1; i += 2 {
		a, b := c.mini[i], c.mini[i+1]
		values = append(values, (values[a>>1] != (a&1 != 0)) && (values[b>>1] != (b&1 != 0)))
	}
	return values[len(values)-1] != (c.mini[len(c.mini)-1] != 0)
}
func mappingReviewMiniCases() []mappingReviewMini {
	cases := []mappingReviewMini{{mini: []int32{0}}, {mini: []int32{1}}}
	for leaf := int32(0); leaf < 4; leaf++ {
		for phase := int32(0); phase < 2; phase++ {
			cases = append(cases, mappingReviewMini{inputs: 1, leaves: []int32{leaf}, mini: []int32{0, phase}})
		}
	}
	for seed := uint32(1); seed <= 64; seed++ {
		state := seed
		next := func() uint32 { state = state*1664525 + 1013904223; return state }
		inputs := 2 + int(seed%5)
		for hash := int32(0); hash < 2; hash++ {
			c := mappingReviewMini{inputs: inputs, hash: hash}
			nLeaves := inputs
			for i := 0; i < nLeaves; i++ {
				leaf := int32(2*(i+1)) + int32(next()&1)
				if hash != 0 {
					leaf = int32(next() % uint32(2*(inputs+1)))
				}
				c.leaves = append(c.leaves, leaf)
			}
			gates := 1 + int(seed%12)
			for i := 0; i < gates; i++ {
				a, b := int32(0), int32(2)
				if i > 0 {
					a = int32(2 * (nLeaves + i - 1))
					b = int32(2 * ((i + 1) % nLeaves))
				}
				if hash != 0 {
					a = int32(next()%uint32(nLeaves+i)) * 2
					b = int32(next()%uint32(nLeaves+i)) * 2
				}
				a ^= int32(next() & 1)
				b ^= int32(next() & 1)
				c.mini = append(c.mini, a, b)
			}
			c.mini = append(c.mini, int32(seed&1))
			cases = append(cases, c)
		}
	}
	for phase := int32(0); phase < 2; phase++ {
		cases = append(cases, mappingReviewMini{inputs: 1, leaves: []int32{2}, mini: []int32{0, 0, phase}})
	}
	return cases
}

// Run the full scalar matrix independently of the optional source comparison.
// No C compiler, archive, executable or .c input is used by this test.
func TestMappingMiniGoOnlyReconstruction(t *testing.T) {
	cases := mappingReviewMiniCases()
	for index, c := range cases {
		t.Run(fmt.Sprintf("case%d/hash%d", index, c.hash), func(t *testing.T) {
			p := Gia_ManStart(64)
			defer Gia_ManStop(p)
			for i := 0; i < c.inputs; i++ {
				Gia_ManAppendCi(p)
			}
			leaves := Vec_IntAllocArrayCopy(c.leaves, len(c.leaves))
			defer Vec_IntFree(leaves)
			mini := Vec_IntAllocArrayCopy(c.mini, len(c.mini))
			defer Vec_IntFree(mini)
			if c.hash != 0 {
				Gia_ManHashAlloc(p)
			}
			root := Gia_ManBuildFromMiniInt(p, leaves, mini, c.hash)
			if !slices.Equal(Vec_IntArray(leaves), c.leaves) {
				t.Fatal("mini conversion modified supplied leaves")
			}
			if len(c.mini) <= 2 {
				if !slices.Equal(Vec_IntArray(mini), c.mini) {
					t.Fatal("constant/variable scratch changed")
				}
			} else if Vec_IntSize(mini) != 0 {
				t.Fatal("AND scratch was not cleared")
			}
			Gia_ManAppendCo(p, root)
			for assignment := 0; assignment < 1<<c.inputs; assignment++ {
				input := make([]bool, c.inputs)
				for i := range input {
					input[i] = assignment>>i&1 != 0
				}
				got, err := p.Evaluate(input)
				want := c.want(assignment)
				if err != nil || !slices.Equal(got, []bool{want}) {
					t.Fatalf("assignment%d got%v want%t err%v", assignment, got, want, err)
				}
			}
		})
	}
	if len(cases) != 140 {
		t.Fatalf("mini matrix changed: %d cases, update the coverage inventory", len(cases))
	}
}
