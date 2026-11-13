package bmc

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestMajorityValueV2(t *testing.T) {
	for n := int32(0); n <= 8; n++ {
		for v := int32(0); v < (1 << n); v++ {
			var ones int32
			for i := int32(0); i < n; i++ {
				ones += (v >> i) & 1
			}
			want := int32(0)
			if ones > n/2 {
				want = 1
			}
			if got := Maj_ManValueV2(v, n); got != want {
				t.Fatalf("n=%d value=%d got=%d want=%d", n, v, got, want)
			}
		}
	}
}
func TestMajorityConnectV2(t *testing.T) {
	c := make([][3]int32, 12)
	Maj_ManConnect(c, 3, 12, 4, 0)
	seen := map[int32]bool{}
	for dst, row := range c {
		unique := map[int32]bool{}
		for _, v := range row {
			if v == 0 {
				continue
			}
			if v >= int32(dst) || unique[v] {
				t.Fatalf("invalid edge %d -> %d", v, dst)
			}
			unique[v] = true
			seen[v] = true
		}
	}
	for v := int32(5); v < 11; v++ {
		if !seen[v] {
			t.Fatalf("node %d has no fanout", v)
		}
	}
}

func TestMajorityTruthTablesV2(t *testing.T) {
	for n := int32(3); n <= 8; n++ {
		p := &Maj_Man_tV2{nVars: n, nObjs: n + 3, nWords: util.Abc_TtWordNum(n)}
		Maj_ManTruthTablesV2(p)
		for i := int32(0); i < (1 << uint(n)); i++ {
			if got, want := util.Abc_TtGetBit(Maj_ManTruthV2(p, p.nObjs), i), Maj_ManValueV2(i, n); got != want {
				t.Fatalf("n=%d minterm=%d got=%d want=%d", n, i, got, want)
			}
			for v := int32(0); v < n; v++ {
				if got, want := util.Abc_TtGetBit(Maj_ManTruthV2(p, v+2), i), (i>>uint(v))&1; got != want {
					t.Fatalf("n=%d variable=%d minterm=%d", n, v, i)
				}
			}
		}
	}
}

func TestMajorityManagerV2Markup(t *testing.T) {
	p := Maj_ManAllocV2(3, 2, 0, 0, 0, 0, 0)
	if p.nObjs != 7 || p.iVar <= 1 || p.pSat.NumVars() != p.iVar {
		t.Fatalf("invalid manager: %+v", p)
	}
	for k := 0; k < 3; k++ {
		if p.VarMarks[5][k][4-k] == 0 {
			t.Fatalf("missing initial selector %d", k)
		}
	}
	if p.VarVals[1] != 1 {
		t.Fatal("constant one missing")
	}
}

func TestMajoritySynthesisV2(t *testing.T) {
	if got := Maj_ManExactSynthesis2(3, 1, 0, 0, 0, 0, 0); got != 1 {
		t.Fatalf("3-input majority result %d", got)
	}
}

func TestMajoritySynthesisV2Unsatisfiable(t *testing.T) {
	if got := Maj_ManExactSynthesis2(4, 1, 0, 0, 0, 0, 0); got != 0 {
		t.Fatalf("four inputs cannot fit one MAJ3 gate: %d", got)
	}
}
