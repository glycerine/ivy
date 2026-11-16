package gia

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type mappingReviewMini struct {
	inputs       int
	leaves, mini []int32
	hash         int32
}

func (c mappingReviewMini) write(t *testing.T, path string) {
	t.Helper()
	var w strings.Builder
	fmt.Fprintf(&w, "%d %d %d %d\n", c.inputs, len(c.leaves), len(c.mini), c.hash)
	for _, v := range c.leaves {
		fmt.Fprintf(&w, "%d ", v)
	}
	w.WriteByte('\n')
	for _, v := range c.mini {
		fmt.Fprintf(&w, "%d ", v)
	}
	w.WriteByte('\n')
	if err := os.WriteFile(path, []byte(w.String()), 0600); err != nil {
		t.Fatal(err)
	}
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
func TestMappingNativeMiniReconstruction(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_mini_review.c")
	path := filepath.Join(t.TempDir(), "mini.txt")
	cases := mappingReviewMiniCases()
	for index, c := range cases {
		c.write(t, path)
		source, err := exec.Command(exe, path).CombinedOutput()
		if err != nil {
			t.Fatalf("native case%d: %v %s", index, err, source)
		}
		p := Gia_ManStart(64)
		for i := 0; i < c.inputs; i++ {
			Gia_ManAppendCi(p)
		}
		leaves, mini := Vec_IntAlloc(len(c.leaves)), Vec_IntAlloc(len(c.mini))
		for _, v := range c.leaves {
			Vec_IntPush(leaves, v)
		}
		for _, v := range c.mini {
			Vec_IntPush(mini, v)
		}
		if c.hash != 0 {
			Gia_ManHashAlloc(p)
		}
		root := Gia_ManBuildFromMiniInt(p, leaves, mini, c.hash)
		var got strings.Builder
		fmt.Fprintf(&got, "root %d\nscratch %d", root, Vec_IntSize(mini))
		for _, v := range Vec_IntArray(mini) {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
		if !reflect.DeepEqual(Vec_IntArray(leaves), c.leaves) && len(c.leaves) != 0 {
			t.Fatal("mini construction changed leaves")
		}
		Gia_ManAppendCo(p, root)
		fmt.Fprintf(&got, "objects %d\n", p.nObjs)
		for _, o := range p.pObjs {
			switch {
			case o.id == 0:
				got.WriteString("const0\n")
			case Gia_ObjIsCi(o) != 0:
				fmt.Fprintf(&got, "ci %d\n", Gia_ObjCioId(o))
			case Gia_ObjIsCo(o) != 0:
				fmt.Fprintf(&got, "co %d\n", Gia_ObjFaninLit0(o, o.id))
			default:
				fmt.Fprintf(&got, "and %d %d\n", Gia_ObjFaninLit0(o, o.id), Gia_ObjFaninLit1(o, o.id))
			}
		}
		got.WriteString("truth ")
		for mask := 0; mask < 1<<c.inputs; mask++ {
			inputs := make([]bool, c.inputs)
			for i := range inputs {
				inputs[i] = mask>>i&1 != 0
			}
			values, err := p.Evaluate(inputs)
			if err != nil || values[0] != c.want(mask) {
				t.Fatalf("case%d input%d truth%v want%v err%v", index, mask, values, c.want(mask), err)
			}
			if values[0] {
				got.WriteByte('1')
			} else {
				got.WriteByte('0')
			}
		}
		got.WriteByte('\n')
		if string(source) != got.String() {
			t.Logf("case%d hash%d leaves%v mini%v", index, c.hash, c.leaves, c.mini)
			esopReviewCompare(t, source, got.String())
		}
		Gia_ManStop(p)
	}
	t.Logf("compared %d exact mini graphs and independent complete truth tables", len(cases))
}
func TestMappingMiniInvalidInputs(t *testing.T) {
	for _, c := range []mappingReviewMini{{mini: nil}, {mini: []int32{2}}, {mini: []int32{-1}}, {leaves: []int32{2}, mini: []int32{1, 0}}, {mini: []int32{0, 0}}, {leaves: []int32{2}, mini: []int32{0, 2, 0}}, {leaves: []int32{2}, mini: []int32{-1, 0, 0}}, {leaves: []int32{2}, mini: []int32{0, 0, 0, 0}}} {
		t.Run(strconv.Itoa(len(c.mini))+fmt.Sprint(c.mini), func(t *testing.T) {
			p := Gia_ManStart(8)
			defer Gia_ManStop(p)
			Gia_ManAppendCi(p)
			leaves, mini := Vec_IntAlloc(len(c.leaves)), Vec_IntAlloc(len(c.mini))
			for _, v := range c.leaves {
				Vec_IntPush(leaves, v)
			}
			for _, v := range c.mini {
				Vec_IntPush(mini, v)
			}
			cexReviewMustPanic(t, func() { Gia_ManBuildFromMiniInt(p, leaves, mini, c.hash) })
		})
	}
}
