package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// Reuse the native archive built for the choice-miter oracle. Both native
// converters receive ordinary AND graphs, their original supported domain.
func compareNativeEquivChoiceOracles(t *testing.T, executable string) {
	t.Helper()
	for name, convert := range map[string]func(*Gia_Man_t, int32) *Gia_Man_t{"gia": Gia_ManEquivToChoices, "hcd": Hcd_ManEquivToChoices} {
		t.Run("equiv_"+name, func(t *testing.T) {
			args := []string{"--equiv"}
			if name == "hcd" {
				args = append(args, "--hcd")
			}
			want, err := exec.Command(executable, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("source choice conversion: %v: %s", err, want)
			}
			var got strings.Builder
			for phase := int32(0); phase < 8; phase++ {
				p := Gia_ManStart(32)
				a, b, c := Gia_ManAppendCi(p)^(phase&1), Gia_ManAppendCi(p)^((phase>>1)&1), Gia_ManAppendCi(p)^((phase>>2)&1)
				ab := Gia_ManAppendAnd(p, a, b)
				head := Gia_ManAppendAnd(p, ab, c)
				ac := Gia_ManAppendAnd(p, a, c)
				tail := Gia_ManAppendAnd(p, ac, b)
				Gia_ManAppendCo(p, head)
				Gia_ManAppendCo(p, tail)
				Gia_ManAppendCo(p, tail^1)
				p.StartRepresentatives()
				Gia_ObjSetRepr(p, tail>>1, head>>1)
				Gia_ObjSetNext(p, head>>1, tail>>1)
				q := convert(p, 1)
				fmt.Fprintf(&got, "eq %d %d %d %d\n", phase, q.nObjs, Gia_ManCoNum(q), Gia_ManCountChoices(q))
				for id := int32(0); id < q.nObjs; id++ {
					a, b, _, kind := q.Gate(id)
					if kind < 2 {
						a = -1
					}
					if kind != 2 {
						b = -1
					}
					repr, next := int32(GIA_VOID), int32(0)
					if q.PReprs != nil {
						repr = Gia_ObjRepr(q, id)
					}
					if q.PNexts != nil {
						next = Gia_ObjNext(q, id)
					}
					fmt.Fprintf(&got, "obj %d %d %d %d %d %d\n", id, kind, a, b, repr, next)
				}
				for mask := 0; mask < 8; mask++ {
					values, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
					if err != nil {
						t.Fatal(err)
					}
					fmt.Fprintf(&got, "truth %d", mask)
					for _, value := range values {
						fmt.Fprintf(&got, " %d", boolInt(value))
					}
					got.WriteByte('\n')
				}
			}
			if got.String() != string(want) {
				t.Fatalf("Go choices:\n%sC choices:\n%s", got.String(), want)
			}
		})
	}
}
