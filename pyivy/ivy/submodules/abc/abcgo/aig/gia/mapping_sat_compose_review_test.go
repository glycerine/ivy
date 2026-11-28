package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/internal/testoracle"
	if_ "github.com/glycerine/abcgo/map/if"
	"strings"
	"testing"
)

func TestMappingSATNativeCompositionHelpers(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_sat_compose_review.c")
	want, err := testoracle.Command(t, executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original composition: %v: %s", err, want)
	}
	state := uint64(1234567)
	next := func() uint64 { state = state*6364136223846793005 + 1442695040888963407; return state }
	var got strings.Builder
	for k := int32(0); k <= 4; k++ {
		for trial := 0; trial < 128; trial++ {
			table := uint32(next())
			functions := make([]uint64, 4)
			for i := range functions {
				functions[i] = next()
			}
			fmt.Fprintf(&got, "compose %016x\n", if_.If_ManSat6ComposeLut4(table, functions, k))
		}
	}
	for trial := 0; trial < 512; trial++ {
		order := []int32{0, 1, 2, 3, 4, 5}
		for i := 5; i > 0; i-- {
			j := int(next() % uint64(i+1))
			order[i], order[j] = order[j], order[i]
		}
		ns := int(next() % 4)
		nb := int(next() % uint64(5-ns))
		nf := int(next() % uint64(4-ns))
		if nb+ns+nf > 6 {
			nf = 6 - nb - ns
		}
		bound, free := next(), next()
		fmt.Fprintf(&got, "truth %016x\n", if_.If_ManSat6Truth(bound, free, order[:nb], order[nb:nb+ns], order[nb+ns:nb+ns+nf]))
	}
	actual, expected := strings.Split(got.String(), "\n"), strings.Split(string(want), "\n")
	if len(actual) != len(expected) {
		t.Fatalf("composition record counts C%d Go%d", len(expected), len(actual))
	}
	for i, w := range expected {
		if actual[i] != w {
			t.Fatalf("composition case%d Go%s C%s", i, actual[i], w)
		}
	}
}

var mappingSATCompositionResult uint64

func BenchmarkMappingSATCompositionHelpers(b *testing.B) {
	bound, shared, free := []int32{0, 2, 3}, []int32{1}, []int32{4, 5}
	var result uint64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result = if_.If_ManSat6Truth(uint64(i), 0x159d159d159d159d, bound, shared, free)
	}
	mappingSATCompositionResult = result
}
