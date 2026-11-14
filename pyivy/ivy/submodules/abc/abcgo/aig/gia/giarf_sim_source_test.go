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

func TestGiarfSimulationMatchesOriginalAndGraphs(t *testing.T) {
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
	binary := filepath.Join(t.TempDir(), "giarf-simulation")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src,
		filepath.Join(dir, "test_vectors/giarf_simulation.c"), "-Wl,--gc-sections", "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile source simulation: %v: %s", err, out)
	}
	want, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run source simulation: %v: %s", err, want)
	}
	var got strings.Builder
	for signs := int32(0); signs < 8; signs++ {
		for recursive := 0; recursive < 2; recursive++ {
			g := Gia_ManStart(1)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a^(signs&1), b^((signs>>1)&1))
			y := Gia_ManAppendAnd(g, c, x^((signs>>2)&1))
			root := Gia_ManAppendAnd(g, x^1, y)
			p := Gia_ManEquivStart(g, 1000, 0)
			g.StartRepresentatives()
			members := Vec_IntAlloc(3)
			for id := int32(4); id < 7; id++ {
				Vec_IntPush(members, id)
			}
			for round := uint32(0); round < 2; round++ {
				for i, word := range []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0} {
					Hcd_ObjSetSim(p, int32(i+1), word^(round*0x5a5a5a5a))
				}
				if recursive != 0 {
					Gia_ManIncrementTravId(g)
					Gia_Resimulate_rec(p, root>>1)
				} else {
					Hcd_ManSimulateSimple(p)
				}
				fmt.Fprintf(&got, "%d %d %d", signs, recursive, round)
				for _, word := range p.pSimInfo {
					fmt.Fprintf(&got, " %d", word)
				}
				for id := int32(0); id < 7; id++ {
					Gia_ObjUnsetRepr(g, id)
				}
				Hcd_ManClassCreate(g, members)
				fmt.Fprintf(&got, " split %d", Hcd_ManClassRefineOne(p, 4))
				for id := int32(4); id < 7; id++ {
					fmt.Fprintf(&got, " %d:%d", Gia_ObjRepr(g, id), Gia_ObjNext(g, id))
				}
				removed := int32(5)
				if recursive != 0 {
					removed = 6
				}
				fmt.Fprintf(&got, " remove %d", Hcd_ManClassClassRemoveOne(p, removed))
				for id := int32(4); id < 7; id++ {
					fmt.Fprintf(&got, " %d:%d", Gia_ObjRepr(g, id), Gia_ObjNext(g, id))
				}
				got.WriteByte('\n')
			}
		}
	}
	if got.String() != string(want) {
		t.Fatalf("Go simulation:\n%sC simulation:\n%s", got.String(), want)
	}
}
