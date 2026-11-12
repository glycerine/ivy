package sbd

import (
	"bytes"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"testing"
)

func TestTruthCnfExhaustiveFourInput(t *testing.T) {
	cover, cnf := vec.Vec_IntAlloc(32), vec.Vec_StrAlloc(128)
	for function := uint64(0); function < 65536; function++ {
		truth := function * 0x0001000100010001
		n := int32(4)

		count := Sbd_TruthToCnf(truth, n, cover, cnf)
		clauses := 0
		for _, b := range vec.Vec_StrArray(cnf) {
			if b == 255 {
				clauses++
			}
		}
		if clauses != int(count) {
			t.Fatal("clause count")
		}
		for assignment := uint32(0); assignment < 1<<uint(n+1); assignment++ {
			satisfies, clause := true, false
			for _, entry := range vec.Vec_StrArray(cnf) {
				if entry == 255 {
					satisfies = satisfies && clause
					clause = false
				} else {
					value := assignment >> (entry >> 1) & 1
					clause = clause || (value^uint32(entry&1) != 0)
				}
			}
			want := uint64(assignment>>n&1) == (truth >> uint(assignment&uint32(1<<n-1)) & 1)
			if satisfies != want {
				t.Fatalf("truth%x assignment%x n%d", function, assignment, n)
			}
		}
	}
}
func TestCnfTranslationPreservesPolarity(t *testing.T) {
	cnf := vec.Vec_StrAlloc(8)
	for _, v := range []byte{0, 3, 4, 255, 1, 255} {
		vec.Vec_StrPush(cnf, v)
	}
	mapping := vec.Vec_IntAlloc(3)
	for _, v := range []int32{8, 13, 20} {
		vec.Vec_IntPush(mapping, v)
	}
	out := vec.Vec_WecAlloc(3)
	Sbd_TranslateCnf(out, cnf, mapping, 10)
	if vec.Vec_WecSize(out) != 3 || !slices.Equal(vec.Vec_IntArray(vec.Vec_WecEntry(out, 0)), []int32{8, 12, 21}) || !slices.Equal(vec.Vec_IntArray(vec.Vec_WecEntry(out, 1)), []int32{9}) || vec.Vec_IntSize(vec.Vec_WecEntry(out, 2)) != 0 {
		t.Fatal("literal translation")
	}
	var printed bytes.Buffer
	sbdPrintCnf(&printed, cnf)
	if printed.String() != "0 -1 2 \n-0 \n" {
		t.Fatal(printed.String())
	}
}
