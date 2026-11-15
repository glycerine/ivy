package gia

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func esopReviewCompare(t *testing.T, want []byte, got string) {
	t.Helper()
	a, b := strings.Split(strings.TrimSpace(string(want)), "\n"), strings.Split(strings.TrimSpace(got), "\n")
	if len(a) != len(b) {
		t.Fatalf("native rows C%d Go%d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("record%d\nGo %s\n C %s", i, b[i], a[i])
		}
	}
}
func TestEsopNativeExtractedCovers(t *testing.T) {
	exe := compileGiaSourceOracle(t, "esop_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C ESOP: %v\n%s", err, native)
	}
	var records strings.Builder
	for seed := uint32(0); seed < 64; seed++ {
		src := cofReviewGraph(seed)
		Gia_ManHashStop(src)
		var cover *Vec_Wec_t
		Eso_ManCompute(src, 0, &cover)
		esopReviewCheckCover(t, src, cover)
		fmt.Fprintf(&records, "cover %d levels%d\n", seed, vec.Vec_WecSize(cover))
		for i := 0; i < vec.Vec_WecSize(cover); i++ {
			records.WriteString("cube")
			for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(cover, i)) {
				fmt.Fprintf(&records, " %d", lit)
			}
			records.WriteByte('\n')
		}
		out := Eso_ManCompute(src, 0, nil)
		cofReviewCheckRebuild(t, src, out)
		Gia_ManStop(out)
		Gia_ManStop(src)
	}
	esopReviewCompare(t, native, records.String())
}

func esopReviewCube(code, nvars int) []int32 {
	var cube []int32
	for i := 0; i < nvars; i++ {
		x := code % 3
		code /= 3
		if x != 0 {
			cube = append(cube, int32(2*i)+giaBool(x == 2))
		}
	}
	return cube
}
func TestEsopNativeCubeOperations(t *testing.T) {
	exe := compileGiaSourceOracle(t, "esop_review.c")
	native, err := exec.Command(exe, "--cubes").CombinedOutput()
	if err != nil {
		t.Fatalf("C cube operations: %v\n%s", err, native)
	}
	g := Gia_ManStart(16)
	defer Gia_ManStop(g)
	for i := 0; i < 4; i++ {
		Gia_ManAppendCi(g)
	}
	p := Eso_ManAlloc(g)
	defer Eso_ManStop(p)
	cubes := make([]*Vec_Int_t, 81)
	ids := make([]int32, 81)
	for code := range cubes {
		lits := esopReviewCube(code, 4)
		cubes[code] = Vec_IntAlloc(len(lits))
		for _, lit := range lits {
			Vec_IntPush(cubes[code], lit)
		}
		ids[code] = p.Cube1
		if code != 0 {
			ids[code] = p.intern(lits)
		}
	}
	out, esop := Vec_IntAlloc(4), Vec_IntAlloc(8)
	var records strings.Builder
	for a := range cubes {
		for b := range cubes {
			x, y := Vec_IntArray(cubes[a]), Vec_IntArray(cubes[b])
			equal, unequal := int32(-2), int32(-2)
			if len(x) == len(y) {
				equal = Eso_ManFindDistOneLitEqual(x, y, int32(len(x)))
			}
			if len(x)+1 == len(y) {
				unequal = Eso_ManFindDistOneLitNotEqual(x, y, int32(len(x)))
			}
			product := ids[b]
			if a != 0 {
				product = ids[a]
				if b != 0 {
					product = Eso_ManComputeAnd(p, cubes[a], cubes[b], out)
				}
			}
			Eso_ManMinimizeAdd(p, ids[a])
			Eso_ManMinimizeAdd(p, ids[b])
			Eso_ManMinimizeCopy(p, esop)
			fmt.Fprintf(&records, "pair %d %d eq%d neq%d and%d xor", a, b, equal, unequal, product)
			for _, id := range Vec_IntArray(esop) {
				fmt.Fprintf(&records, " %d", id)
			}
			records.WriteByte('\n')
		}
	}
	esopReviewCompare(t, native, records.String())
}

func TestEsopNativeReports(t *testing.T) {
	exe := compileGiaSourceOracle(t, "esop_review.c")
	for _, mode := range []string{"--verbose", "--print"} {
		t.Run(mode, func(t *testing.T) {
			native, err := exec.Command(exe, mode).CombinedOutput()
			if err != nil {
				t.Fatalf("C ESOP report: %v\n%s", err, native)
			}
			got := hashReviewStdout(t, func() {
				if mode == "--verbose" {
					g := cofReviewGraph(3)
					Gia_ManHashStop(g)
					out := Eso_ManCompute(g, 1, nil)
					Gia_ManStop(out)
					Gia_ManStop(g)
					return
				}
				g := Gia_ManStart(8)
				for i := 0; i < 4; i++ {
					Gia_ManAppendCi(g)
				}
				p := Eso_ManAlloc(g)
				cover := Vec_IntAlloc(4)
				Eso_ManCoverPrint(p, cover)
				for _, id := range []int32{p.Cube1, 0, 3} {
					Vec_IntPush(cover, id)
				}
				Eso_ManCoverPrint(p, cover)
				Eso_ManStop(p)
				Gia_ManStop(g)
			})
			// Clock values differ; retain the source field and its column width.
			clock := regexp.MustCompile(`Time = {5}[0-9]+\.[0-9]{2} sec`)
			if mode == "--verbose" {
				if !clock.MatchString(got) || !clock.Match(native) {
					t.Fatalf("missing source time field\nGo %s\n C %s", got, native)
				}
				got = clock.ReplaceAllString(got, "<time>")
				native = clock.ReplaceAll(native, []byte("<time>"))
			}
			if got != string(native) {
				t.Fatalf("report differs\nGo %s\n C %s", got, native)
			}
		})
	}
}
