package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestForcePlacementMatchesOriginalAndGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "force_placement.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C FORCE placement: %v: %s", err, want)
	}
	var records strings.Builder
	perms := [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	for cluster := 0; cluster < 2; cluster++ {
		for perm, indices := range perms {
			g := Gia_ManStart(1)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			inner := Gia_ManAppendAnd(g, b, c)
			x := Gia_ManAppendAnd(g, a, inner)
			y := Gia_ManAppendAnd(g, a, b)
			z := Gia_ManAppendAnd(g, x, y)
			Gia_ManAppendCo(g, z)
			Gia_ManAppendCo(g, y)
			Gia_ManAppendCo(g, x)
			start := Frc_ManStartSimple
			if cluster != 0 {
				start = Frc_ManStart
			}
			p := start(g)
			order := Vec_IntAlloc(3)
			for _, index := range indices {
				Vec_IntPush(order, Frc_ManCo(p, int32(index)).hHandle)
			}
			places := func() {
				for _, obj := range p.objects {
					fmt.Fprintf(&records, " %d", obj.pPlace)
				}
			}
			fmt.Fprintf(&records, "%d %d %d cuts %d %d dfs", cluster, perm, p.nObjs,
				Frc_ManCrossCut(p, order, 0), Frc_ManCrossCut(p, order, 1))
			Frc_ManPlaceDfs(p, order)
			places()
			var other int32
			best := Frc_ManPlaceDfsBoth(p, order, &other)
			fmt.Fprintf(&records, " best %d %d", best, other)
			places()
			Frc_ManPlacementRefine(p, 3, 0)
			records.WriteString(" refined")
			places()
			records.WriteByte('\n')
		}
	}
	if records.String() != string(want) {
		t.Fatalf("Go FORCE:\n%sC FORCE:\n%s", records.String(), want)
	}
}
