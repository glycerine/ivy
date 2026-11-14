package gia

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestChoiceTfiMatchesOriginalSource(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture uses GNU linker section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "choice-tfi")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src,
		filepath.Join(dir, "test_vectors/choice_tfi.c"), filepath.Join(src, "aig/gia/giaEquiv.c"),
		filepath.Join(src, "aig/gia/giaHcd.c"), filepath.Join(src, "aig/gia/giaUtil.c"), "-Wl,--gc-sections", "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile original TFI routines: %v: %s", err, out)
	}
	want, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run source TFI fixture: %v: %s", err, want)
	}
	var got strings.Builder
	for mode := 0; mode < 4; mode++ {
		p := choiceTfiSourceGraph()
		p.PNexts = make([]int32, p.nObjs)
		if mode&1 != 0 {
			p.PNexts[6] = 8
		}
		if mode&2 != 0 {
			p.PNexts[9] = 11
		}
		if mode == 3 {
			p.PNexts[8], p.PNexts[11] = 14, 15
		}
		for api, check := range []func(*Gia_Man_t, *Gia_Obj_t, *Gia_Obj_t) int32{Gia_ObjCheckTfi, Hcd_ObjCheckTfi} {
			for old := int32(5); old < 16; old++ {
				for node := int32(0); node < 16; node++ {
					var root *Gia_Obj_t
					if node != 0 {
						root = Gia_ManObj(p, node)
					}
					result := check(p, Gia_ManObj(p, old), root)
					var dirty uint32
					for _, obj := range p.pObjs {
						dirty |= obj.fMark0
					}
					fmt.Fprintf(&got, "%d %d %d %d %d %d\n", mode, api, old, node, result, dirty)
				}
			}
		}
	}
	for api, remove := range []func(*Gia_Man_t){Gia_ManRemoveBadChoices, Hcd_ManRemoveBadChoices} {
		for limit, size := range []int32{9, 15, 16, 16} {
			p := choiceTfiSourceGraph()
			p.pObjs, p.nObjs = p.pObjs[:size], size
			if limit == 3 {
				// Replace node 11's use of member 8 by the ordinary AND(a,d),
				// leaving only the last member (14) with a physical fanout.
				p.pObjs[11].iDiff0, p.pObjs[11].iDiff1 = 11-1, 11-4
			}
			p.StartRepresentatives()
			p.PNexts = make([]int32, size)
			Gia_ObjSetRepr(p, 8, 6)
			Gia_ObjSetNext(p, 6, 8)
			if size > 14 {
				Gia_ObjSetRepr(p, 14, 6)
				Gia_ObjSetNext(p, 8, 14)
			}
			remove(p)
			fmt.Fprintf(&got, "remove %d %d %d", api, limit, size)
			for id := int32(0); id < size; id++ {
				fmt.Fprintf(&got, " %d:%d:%d", Gia_ObjRepr(p, id), Gia_ObjNext(p, id), p.pObjs[id].fMark0)
			}
			got.WriteByte('\n')
		}
	}
	if got.String() != string(want) {
		t.Fatalf("Go TFI matrix:\n%sC matrix:\n%s", got.String(), want)
	}
}

func choiceTfiSourceGraph() *Gia_Man_t {
	p := Gia_ManStart(1)
	for i := 0; i < 4; i++ {
		Gia_ManAppendCi(p)
	}
	for _, fanins := range [][2]int32{{1, 2}, {3, 5}, {1, 3}, {2, 7}, {4, 6}, {2, 9}, {4, 8}, {4, 5}, {2, 3}, {1, 13}, {4, 14}} {
		Gia_ManAppendAnd(p, fanins[0]<<1, fanins[1]<<1)
	}
	return p
}
