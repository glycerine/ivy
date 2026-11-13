package acec

import (
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func sourceFixture(t *testing.T) map[string][]int32 {
	t.Helper()
	data, err := os.ReadFile("test_vectors/acec_source_results.txt")
	if err != nil {
		t.Fatal(err)
	}
	r := map[string][]int32{}
	for _, line := range strings.Split(string(data), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		for _, word := range strings.Fields(v) {
			n, e := strconv.Atoi(word)
			if e != nil {
				t.Fatal(e)
			}
			r[k] = append(r[k], int32(n))
		}
	}
	return r
}

// This graph contains a Booth XOR/AND/XOR recoder and a positive partial
// product. Both feed a two-adder chain; its input order matches the C fixture.
func TestSourceBoothIgnoreAndNormalize(t *testing.T) {
	expected := sourceFixture(t)
	p := gia.Gia_ManStart(200)
	in := make([]int32, 7)
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(p)
	}
	gia.Gia_ManHashStart(p)
	product := gia.Gia_ManHashAnd(p, in[0], in[1])
	digit := gia.Gia_ManHashXor(p, in[2], in[3])
	pp := gia.Gia_ManHashXor(p, gia.Gia_ManHashAnd(p, digit, in[4]), in[5])
	gia.Gia_ManHashStop(p)
	var lo, hi [2]int32
	Acec_InsertFadd(p, []int32{product, pp, in[6]}, lo[:])
	Acec_InsertFadd(p, []int32{lo[1], in[0], in[5]}, hi[:])
	for _, lit := range []int32{lo[0], hi[0], hi[1]} {
		gia.Gia_ManAppendCo(p, lit)
	}
	ignore := Acec_BoothFindPPG(p)
	var got []int32
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		if vec.Vec_BitEntry(ignore, int(i)) != 0 {
			got = append(got, i)
		}
	}
	if !reflect.DeepEqual(got, expected["booth_mask"]) {
		t.Fatalf("mask %v, ABC %v", got, expected["booth_mask"])
	}
	orig := make([][]bool, 128)
	for i := 0; i < 128; i++ {
		in := make([]bool, 7)
		for j := range in {
			in[j] = i&(1<<j) != 0
		}
		var err error
		orig[i], err = p.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
	}
	n := Acec_Normalize(p, 1, 0)
	if gia.Gia_ManCiNum(n) != expected["booth_normalized_inputs"][0] || gia.Gia_ManCoNum(n) != expected["booth_normalized_outputs"][0] {
		t.Fatal("normalized interface differs from ABC")
	}
	for i := 0; i < 128; i++ {
		in := make([]bool, 7)
		for j := range in {
			in[j] = i&(1<<j) != 0
		}
		v, err := n.Evaluate(in)
		if err != nil || !reflect.DeepEqual(v, orig[i]) {
			t.Fatalf("input %d: %v vs %v: %v", i, v, orig[i], err)
		}
	}
}

func TestSourceTiedXorMajorityCuts(t *testing.T) {
	expected := sourceFixture(t)
	x := [][4]int32{{1, 2, 3, 10}, {1, 2, 3, 11}, {1, 2, 3, 12}}
	m := [][4]int32{{1, 2, 3, 20}, {1, 2, 3, 21}}
	sort.SliceStable(x, func(i, j int) bool { return Dtc_ManCompare(x[i][:], x[j][:]) < 0 })
	sort.SliceStable(m, func(i, j int) bool { return Dtc_ManCompare(m[i][:], m[j][:]) < 0 })
	a, b := vec.Vec_IntAlloc(12), vec.Vec_IntAlloc(8)
	for _, row := range x {
		for _, v := range row {
			vec.Vec_IntPush(a, v)
		}
	}
	for _, row := range m {
		for _, v := range row {
			vec.Vec_IntPush(b, v)
		}
	}
	got := vec.Vec_IntArray(Dtc_ManFindCommonCuts(nil, a, b))
	if !reflect.DeepEqual(got, expected["tied_cuts"]) {
		t.Fatalf("tied cuts %v, ABC %v", got, expected["tied_cuts"])
	}
}
