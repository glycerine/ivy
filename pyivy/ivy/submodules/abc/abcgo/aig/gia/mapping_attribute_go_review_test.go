package gia_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
)

func TestMappingDriverOutputAttributeOrder(t *testing.T) {
	pairs := [][2]int32{{0, 0}, {1, 0}, {0, 1}, {-1, 0}, {0, -2}, {1, -3}}
	for _, first := range pairs {
		for _, second := range pairs {
			for mode := 0; mode < 6; mode++ {
				t.Run(fmt.Sprintf("first%v/second%v/mode%d", first, second, mode), func(t *testing.T) {
					src := gia.Gia_ManStart(16)
					defer gia.Gia_ManStop(src)
					a, b, c := gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src)
					left := gia.Gia_ManAppendAnd(src, gia.Gia_ManAppendAnd(src, a, b), c)
					right := gia.Gia_ManAppendAnd(src, gia.Gia_ManAppendAnd(src, a, b^1), c)
					for _, lit := range []int32{left, left ^ 1, right, right ^ 1, a} {
						gia.Gia_ManAppendCo(src, lit)
					}
					attrs := []int32{first[0], first[1], second[0], second[1], -1}
					gia.SetMappingOutputAttributesForTest(src, attrs)
					var pars if_.If_Par_t
					if_.Gia_ManSetIfParsDefault(&pars)
					pars.NLutSize = 3
					pars.FHashMapping = int32(mode % 3 & 1)
					pars.FExpRed = int32(mode / 3)
					var result *gia.Gia_Man_t
					mappingGoReviewCapture(t, func() {
						if mode%3 == 0 {
							result = if_.Gia_ManPerformMappingInt(src, &pars)
						} else {
							result = if_.Gia_ManPerformMapping(src, &pars)
						}
					})
					if result == nil {
						t.Fatal("attribute mapping failed")
					}
					defer gia.Gia_ManStop(result)
					for i, pair := range [][2]int32{first, second} {
						want := int32(3)
						// The later CO overwrites fSpec on the shared regular driver.
						// Source area recovery subsequently expands to the full limit
						// without consulting fSpec; retain that behavior as well.
						if pair[1] != 0 && pars.FExpRed == 0 {
							want = 2
						}
						for j := 0; j < 2; j++ {
							id := gia.Gia_ObjFaninId0p(result, gia.Gia_ManCo(result, int32(2*i+j)))
							if got := gia.Gia_ObjLutSize(result, id); got != want {
								t.Errorf("CO%d LUT width %d, want %d", 2*i+j, got, want)
							}
						}
					}
					_, _, sourceAttrs, _ := src.MappingInputs()
					_, _, resultAttrs, _ := result.MappingInputs()
					if mode%3 == 0 {
						if !slices.Equal(sourceAttrs, attrs) || resultAttrs != nil {
							t.Fatal("inner driver changed attribute ownership")
						}
					} else if sourceAttrs != nil || !slices.Equal(resultAttrs, attrs) {
						t.Fatal("outer driver did not transfer attributes")
					}
					for x := 0; x < 8; x++ {
						a, b, c := x&1 != 0, x&2 != 0, x&4 != 0
						left, right := a && b && c, a && !b && c
						want := []bool{left, !left, right, !right, a}
						got, err := result.Evaluate([]bool{a, b, c})
						if err != nil || !slices.Equal(got, want) {
							t.Fatalf("input%d: got%v want%v err%v", x, got, want, err)
						}
					}
				})
			}
		}
	}
}
