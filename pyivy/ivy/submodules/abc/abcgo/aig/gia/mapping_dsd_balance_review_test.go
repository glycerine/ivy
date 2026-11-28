package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/internal/testoracle"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"strings"
	"testing"
)

func TestMappingDSDOriginalBalancingContracts(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_balance_review.c")
	output, err := testoracle.Command(t, executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original balancing: %v %s", err, output)
	}
	p := if_.If_DsdManAlloc(8, 0)
	defer if_.If_DsdManFree(p, 0)
	if_.If_DsdManAllocIsops(p, 8)
	var got strings.Builder
	texts := []string{"(abc)", "E8{a(bc)d}", "(ab[cd]<efg>)", "<a(bc)[de]>"}
	sizes := []int32{3, 4, 7, 5}
	for index, text := range texts {
		n := sizes[index]
		truth := dau.Dau_DsdToTruth(text, n)
		permutation := make([]byte, n)
		lit := if_.If_DsdManCompute(p, truth, n, permutation, "")
		for mode := 0; mode < 2; mode++ {
			times := make([]int32, n)
			if mode != 0 {
				for i := range times {
					times[i] = int32(i % 3)
				}
			}
			aig := vec.Vec_IntAlloc(100)
			var area, timingArea int32
			delay := if_.If_CutDsdBalanceEvalInt(p, lit, times, aig, &area, permutation)
			timingDelay := if_.If_CutDsdBalanceEvalInt(p, lit, times, nil, &timingArea, permutation)
			var support int32
			var profile uint64
			profileDelay := if_.If_CutDsdBalancePinDelays_rec(p, lit>>1, times, &profile, &support, n, permutation)
			fmt.Fprintf(&got, "balance %d %d %d %d %d %d %d %016x aig", index, mode, delay, area, timingDelay, timingArea, profileDelay, profile)
			data := vec.Vec_IntArray(aig)
			for _, v := range data {
				fmt.Fprintf(&got, " %d", v)
			}
			got.WriteByte('\n')
			for assignment := 0; assignment < 1<<uint(n); assignment++ {
				values := make([]bool, n, int(n)+(len(data)-1)/2)
				for i := range values {
					values[i] = assignment>>uint(i)&1 != 0
				}
				for i := 0; i < len(data)-1; i += 2 {
					a, b := data[i], data[i+1]
					values = append(values, (values[a>>1] != (a&1 != 0)) && (values[b>>1] != (b&1 != 0)))
				}
				actual := values[len(values)-1] != (data[len(data)-1] != 0)
				expected := truth[assignment>>6]>>uint(assignment&63)&1 != 0
				if actual != expected {
					t.Fatalf("case%d mode%d assignment%d balanced truth mismatch", index, mode, assignment)
				}
			}
			vec.Vec_IntFree(aig)
		}
	}
	if got.String() != string(output) {
		t.Fatalf("Go balancing:\n%soriginal C:\n%s", got.String(), output)
	}
	t.Log("8 exact native balancing graphs, graph/timing costs and full-support profiles; independent complete truth checks")
}
