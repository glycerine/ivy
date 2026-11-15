package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestFxNativeMultiwordRetrieval(t *testing.T) {
	exe := compileGiaSourceOracle(t, "fx_wide_retrieve_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C multiword retrieve: %v\n%s", err, native)
	}
	var got strings.Builder
	for n := 2; n <= 9; n++ {
		for seed := 0; seed < 8; seed++ {
			for reverse := int32(0); reverse < 2; reverse++ {
				p := Gia_ManStart(64)
				ins := make([]int32, n)
				for i := range ins {
					ins[i] = Gia_ManAppendCi(p)
				}
				lit := ins[0] ^ int32(seed&1)
				for i := 1; i < n; i++ {
					lit = Gia_ManAppendAnd(p, lit, ins[i]^int32(seed>>i&1))
				}
				Gia_ManAppendCo(p, lit)
				p.VMapping = Vec_IntStart(int(p.nObjs))
				Vec_IntWriteEntry(p.VMapping, int(lit>>1), int32(Vec_IntSize(p.VMapping)))
				Vec_IntPush(p.VMapping, int32(n))
				for i := n - 1; i >= 0; i-- {
					Vec_IntPush(p.VMapping, ins[i]>>1)
				}
				Vec_IntPush(p.VMapping, lit>>1)
				var phases *Vec_Str_t
				cubes := Gia_ManFxRetrieve(p, &phases, reverse)
				fmt.Fprintf(&got, "n%d seed%d reverse%d phases", n, seed, reverse)
				for i := 0; i < Vec_StrSize(phases); i++ {
					fmt.Fprint(&got, Vec_StrEntry(phases, i))
				}
				for i := 0; i < Vec_WecSize(cubes); i++ {
					row := Vec_IntArray(Vec_WecEntry(cubes, i))
					got.WriteString(" [")
					for j, x := range row {
						if j > 0 {
							got.WriteByte(',')
						}
						fmt.Fprint(&got, x)
					}
					got.WriteByte(']')
				}
				got.WriteByte('\n')
				// This conjunction must retrieve exactly one cube of all n literals.
				if Vec_WecSize(cubes) != 1 || Vec_IntSize(Vec_WecEntry(cubes, 0)) != n+1 {
					t.Fatal("conjunction cover shape changed")
				}
				row := Vec_IntArray(Vec_WecEntry(cubes, 0))
				for input := 0; input < 1<<n; input++ {
					truth := true
					for _, literal := range row[1:] {
						truth = truth && ((input>>uint(literal>>1)&1 != 0) != (literal&1 != 0))
					}
					truth = truth != (Vec_StrEntry(phases, n) != 0)
					if want := (input ^ (seed & ((1 << n) - 1))) == (1<<n)-1; truth != want {
						t.Fatalf("n%d seed%d input%d", n, seed, input)
					}
				}
				Vec_WecFree(cubes)
				Vec_StrFree(phases)
				Gia_ManStop(p)
			}
		}
	}
	for value := 0; value < 2; value++ {
		p := Gia_ManStart(16)

		a := Gia_ManAppendCi(p)
		p.fGiaSimple = 1
		zero := Gia_ManAppendAnd(p, a, a^1)
		lit := zero
		if value != 0 {
			b := Gia_ManAppendCi(p)
			other := Gia_ManAppendAnd(p, b, b^1)
			lit = Gia_ManAppendAnd(p, zero^1, other^1)
		}
		Gia_ManAppendCo(p, lit)
		p.VMapping = Vec_IntStart(int(p.nObjs))
		Vec_IntWriteEntry(p.VMapping, int(lit>>1), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(1+value))
		for i := 0; i < 1+value; i++ {
			Vec_IntPush(p.VMapping, Gia_ManCi(p, int32(i)).id)
		}
		Vec_IntPush(p.VMapping, lit>>1)
		var phases *Vec_Str_t
		cubes := Gia_ManFxRetrieve(p, &phases, 0)
		if Vec_WecSize(cubes) != 1 || Vec_IntSize(Vec_WecEntry(cubes, 0)) != 1 || Vec_StrEntry(phases, 1+value) != byte(1-value) {
			t.Fatal("constant cover/phase changed")
		}
		fmt.Fprintf(&got, "constant%d phases", value)
		for i := 0; i < Vec_StrSize(phases); i++ {
			fmt.Fprint(&got, Vec_StrEntry(phases, i))
		}
		for i := 0; i < Vec_WecSize(cubes); i++ {
			got.WriteString(" [")
			for j, v := range Vec_IntArray(Vec_WecEntry(cubes, i)) {
				if j > 0 {
					got.WriteByte(',')
				}
				fmt.Fprint(&got, v)
			}
			got.WriteByte(']')
		}
		got.WriteByte('\n')
		Vec_WecFree(cubes)
		Vec_StrFree(phases)
		Gia_ManStop(p)
	}
	esopReviewCompare(t, native, got.String())
}
