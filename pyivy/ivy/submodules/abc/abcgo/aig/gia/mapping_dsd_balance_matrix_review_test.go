package gia_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
)

func mappingDSDReviewCheckTruth(t *testing.T, c mappingDSDReviewCase, data []int32) {
	t.Helper()
	if len(data) < 3 || len(data)&1 == 0 {
		t.Fatalf("invalid full-support balanced graph%v", data)
	}
	for assignment := 0; assignment < 1<<uint(c.variables); assignment++ {
		values := make([]bool, c.variables, int(c.variables)+(len(data)-1)/2)
		for i := range values {
			values[i] = assignment>>uint(i)&1 != 0
		}
		for i := 0; i < len(data)-1; i += 2 {
			a, b := data[i], data[i+1]
			values = append(values, (values[a>>1] != (a&1 != 0)) && (values[b>>1] != (b&1 != 0)))
		}
		actual := values[len(values)-1] != (data[len(data)-1] != 0)
		expected := c.truth[assignment>>6]>>uint(assignment&63)&1 != 0
		if actual != expected {
			t.Fatalf("vars%d assignment%d balanced truth mismatch", c.variables, assignment)
		}
	}
}
func TestMappingDSDNativeBalancingMatrix(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_balance_matrix_review.c")
	cases := mappingDSDReviewCases()
	var input strings.Builder
	for index, c := range cases {
		for mode := 0; mode < 2; mode++ {
			fmt.Fprintf(&input, "%d", c.variables)
			for _, word := range c.truth {
				fmt.Fprintf(&input, " %016x", word)
			}
			for pin := int32(0); pin < c.variables; pin++ {
				arrival := int32(0)
				if mode != 0 {
					arrival = (3*pin + int32(index)) % 5
				}
				fmt.Fprintf(&input, " %d", arrival)
			}
			input.WriteByte('\n')
		}
	}
	path := filepath.Join(t.TempDir(), "balanced-functions.txt")
	if err := os.WriteFile(path, []byte(input.String()), 0600); err != nil {
		t.Fatal(err)
	}
	source, err := exec.Command(executable, path).CombinedOutput()
	if err != nil {
		t.Fatalf("original balancing matrix: %v %s", err, source)
	}
	p := if_.If_DsdManAlloc(8, 0)
	defer if_.If_DsdManFree(p, 0)
	if_.If_DsdManAllocIsops(p, 8)
	var got strings.Builder
	supported, unsupported, partial := 0, 0, 0
	for index, c := range cases {
		for mode := 0; mode < 2; mode++ {
			times := make([]int32, c.variables)
			if mode != 0 {
				for pin := range times {
					times[pin] = (3*int32(pin) + int32(index)) % 5
				}
			}
			permutation := make([]byte, c.variables)
			lit := if_.If_DsdManCompute(p, c.truth, c.variables, permutation, "")
			aig := vec.Vec_IntAlloc(100)
			var area, timingArea, support int32
			delay := if_.If_CutDsdBalanceEvalInt(p, lit, times, aig, &area, permutation)
			timingDelay := if_.If_CutDsdBalanceEvalInt(p, lit, times, nil, &timingArea, permutation)
			profileDelay := int32(-1)
			var profile uint64
			data := vec.Vec_IntArray(aig)
			if delay >= 0 {
				profileDelay = if_.If_CutDsdBalancePinDelays_rec(p, lit>>1, times, &profile, &support, c.variables, permutation)
				mappingDSDReviewCheckTruth(t, c, data)
				supported++
			} else {
				unsupported++
				if len(data) > 0 {
					partial++
				}
			}
			fmt.Fprintf(&got, "balance %d %d %d %d %d %d %016x aig", 2*index+mode, delay, area, timingDelay, timingArea, profileDelay, profile)
			for _, value := range data {
				fmt.Fprintf(&got, " %d", value)
			}
			got.WriteByte('\n')
			vec.Vec_IntFree(aig)
		}
	}
	expected, actual := strings.Split(string(source), "\n"), strings.Split(got.String(), "\n")
	if len(expected) != len(actual) {
		t.Fatalf("balance records C%d Go%d", len(expected), len(actual))
	}
	for i, line := range expected {
		if actual[i] != line {
			c := cases[i/2]
			t.Fatalf("balance case%d vars%d truth%x\nC:%s\nGo:%s", i, c.variables, c.truth, line, actual[i])
		}
	}
	if supported == 0 || unsupported == 0 || partial == 0 {
		t.Fatalf("missing coverage: supported%d unsupported%d partial%d", supported, unsupported, partial)
	}
	t.Logf("%d complete native balance decisions: %d graphs/%d unsupported, %d retain partial construction; exact graphs/costs/profiles and independent full truth", 2*len(cases), supported, unsupported, partial)
}
