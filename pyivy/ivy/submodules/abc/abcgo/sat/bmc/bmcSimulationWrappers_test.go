package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestTernarySimulationEntryPoints(t *testing.T) {
	not := map[int32]int32{ABS_ZER: ABS_ONE, ABS_ONE: ABS_ZER, ABS_UND: ABS_UND}
	for a, want := range not {
		if got := Abs_ManSimInfoNot(a); got != want {
			t.Errorf("NOT %d = %d want %d", a, got, want)
		}
		for b := range not {
			want := ABS_UND
			if a == ABS_ZER || b == ABS_ZER {
				want = ABS_ZER
			} else if a == ABS_ONE && b == ABS_ONE {
				want = ABS_ONE
			}
			if got := Abs_ManSimInfoAnd(a, b); got != want {
				t.Errorf("%d AND %d = %d want %d", a, b, got, want)
			}
		}
	}
}
func TestMoSimulationObjectEntryPoints(t *testing.T) {
	for phase := int32(0); phase < 4; phase++ {
		p := gia.Gia_ManStart(10)
		a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		node := gia.Gia_ManAppendAnd(p, a^(phase&1), b^((phase>>1)&1))
		gia.Gia_ManAppendCo(p, node^1)
		Gia_ManMoFindSimulate(p, 2)
		gia.Gia_ManRandomW(1)
		expected0, expected1 := gia.Gia_ManRandomW(0), gia.Gia_ManRandomW(0)
		gia.Gia_ManRandomW(1)
		Gia_ManMoSetCi(p, a>>1)
		av, bv := Gia_ManMoObj(p, a>>1), Gia_ManMoObj(p, b>>1)
		if av[0] != expected0 || av[1] != expected1 {
			t.Fatal("input words not filled in deterministic sequence")
		}
		bv[0], bv[1] = 0xaaaa5555aaaa5555, 0x00ff00ff00ff00ff
		Gia_ManMoSimAnd(p, node>>1)
		Gia_ManMoSetCo(p, gia.Gia_ObjId(p, gia.Gia_ManCo(p, 0)))
		got, co := Gia_ManMoObj(p, node>>1), Gia_ManMoObj(p, gia.Gia_ObjId(p, gia.Gia_ManCo(p, 0)))
		for i := range got {
			x, y := av[i], bv[i]
			if phase&1 != 0 {
				x = ^x
			}
			if phase&2 != 0 {
				y = ^y
			}
			if got[i] != x&y || co[i] != ^(x&y) {
				t.Fatalf("phase %d word %d got %x out %x", phase, i, got[i], co[i])
			}
		}
	}
}
func TestMesh2ModelOneHotnessEntryPoint(t *testing.T) {
	s := bsat.NewSolver()
	for i := 0; i < 4; i++ {
		s.NewVariable()
	}
	if s.Solve(bsat.Positive(0), bsat.Positive(1), bsat.Positive(2), bsat.Negative(3)) != bsat.Satisfiable {
		t.Fatal("model")
	}
	if got := Bmc_MeshAddOneHotness2(s, 0, 4); got != 3 {
		t.Fatalf("added %d clauses", got)
	}
	for mask := 0; mask < 16; mask++ {
		var assumptions []bsat.Literal
		count := 0
		for i := int32(0); i < 4; i++ {
			if mask&(1<<i) != 0 {
				assumptions = append(assumptions, bsat.Positive(i))
				if i < 3 {
					count++
				}
			} else {
				assumptions = append(assumptions, bsat.Negative(i))
			}
		}
		if got := s.Solve(assumptions...) == bsat.Satisfiable; got != (count <= 1) {
			t.Fatalf("mask %04b got %v", mask, got)
		}
	}
}
