package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func referenceNetwork(seed int) *Nwk_Man_t {
	p := Nwk_ManAlloc()
	for i := 0; i < 6; i++ {
		Nwk_ManCreateCi(p, 0)
	}
	state := uint32(seed + 1)
	for i := 0; i < 12; i++ {
		state = 1664525*state + 1013904223
		a := p.Objects[int(state)%len(p.Objects)]
		state = 1664525*state + 1013904223
		b := p.Objects[int(state)%len(p.Objects)]
		o := Nwk_ManCreateNode(p, 2, 0)
		Nwk_ObjAddFanin(o, a)
		Nwk_ObjAddFanin(o, b)
	}
	for i := 0; i < 4; i++ {
		o := Nwk_ManCreateCo(p)
		Nwk_ObjAddFanin(o, p.Objects[17-i])
	}
	return p
}
func TestAlgorithmsAgainstOriginalC(t *testing.T) {
	compiler, e := exec.LookPath("cc")
	if e != nil {
		t.Skip("C compiler unavailable")
	}
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(t.TempDir(), "nwk-reference")
	args := []string{"-O1", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I" + filepath.Join(root, "src"), "test_vectors/reference.c"}
	for _, name := range []string{"Obj", "Fanio", "Dfs", "Flow", "Timing", "Merge", "Util"} {
		args = append(args, filepath.Join(root, "src/opt/nwk/nwk"+name+".c"))
	}
	args = append(args, filepath.Join(root, "src/aig/aig/aigMem.c"))
	for _, name := range []string{"Man", "Box", "Trav", "Time"} {
		args = append(args, filepath.Join(root, "src/misc/tim/tim"+name+".c"))
	}
	args = append(args, "-Wl,--gc-sections", "-lm", "-o", binary)
	if log, e := exec.Command(compiler, args...).CombinedOutput(); e != nil {
		t.Fatalf("C compile: %v\n%s", e, log)
	}
	for kind := 0; kind < 4; kind++ {
		for seed := 0; seed < 16; seed++ {
			log, e := exec.Command(binary, strconv.Itoa(kind), strconv.Itoa(seed)).CombinedOutput()
			if e != nil {
				t.Fatalf("C kind%d seed%d: %v\n%s", kind, seed, e, log)
			}
			index := strings.Index(string(log), "RESULT")
			if index < 0 {
				t.Fatalf("missing result %s", log)
			}
			fields := strings.Fields(string(log)[index:])[1:]
			if kind == 3 {
				p := referenceNetwork(seed)
				var arrival float32
				var level int32
				if _, e := fmt.Sscan(strings.Join(fields, " "), &arrival, &level); e != nil {
					t.Fatal(e)
				}
				if got := Nwk_ManDelayTraceLut(p); got != arrival || Nwk_ManLevel(p) != level {
					t.Fatalf("timing seed%d: native%g/%d source%g/%d", seed, got, Nwk_ManLevel(p), arrival, level)
				}
				continue
			}
			want := []int32{}
			for _, field := range fields {
				v, e := strconv.Atoi(field)
				if e != nil {
					t.Fatal(e)
				}
				want = append(want, int32(v))
			}
			got := []int32{}
			if kind == 2 {
				g := Nwk_ManGraphAlloc(16)
				state := uint32(seed + 1)
				for i := 0; i < 36; i++ {
					state = 1664525*state + 1013904223
					x := int32(state % 16)
					state = 1664525*state + 1013904223
					y := int32(state % 16)
					Nwk_ManGraphHashEdge(g, x, y)
				}
				Nwk_ManGraphSolve(g)
				got = append(got, vec.Vec_IntArray(g.Pairs)...)
			} else {
				p := referenceNetwork(seed)
				var cut *Vec_Ptr_t
				if kind == 0 {
					cut = Nwk_ManRetimeCutForward(p, 4, 0)
				} else {
					cut = Nwk_ManRetimeCutBackward(p, 4, 0)
				}
				for _, o := range vectorNodes(cut) {
					got = append(got, o.ID)
				}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("kind%d seed%d native%v source%v", kind, seed, got, want)
			}
		}
	}
}
