package gia_test

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/sat/bsat"
)

type mappingSATReviewCase struct {
	size, variables int32
	set             uint32
	outputs         int
	truth           []uint64
}

func mappingSATReviewGroups(c mappingSATReviewCase) (bound, shared, free []int) {
	for i := 0; i < int(c.variables); i++ {
		switch c.set >> uint(2*i) & 3 {
		case 0:
			free = append(free, i)
		case 1:
			bound = append(bound, i)
		case 3:
			shared = append(shared, i)
		}
	}
	return
}
func mappingSATReviewAddress(value int, pins []int) int {
	var result int
	for i, pin := range pins {
		result |= (value >> pin & 1) << i
	}
	return result
}
func mappingSATReviewCompose(c mappingSATReviewCase, bound, free uint64, assignment int) bool {
	b, s, f := mappingSATReviewGroups(c)
	shared := mappingSATReviewAddress(assignment, s)
	bi := shared | mappingSATReviewAddress(assignment, b)<<len(s)
	q := int(bound >> bi & 1)
	fi := q | shared<<1 | mappingSATReviewAddress(assignment, f)<<(1+len(s))
	return free>>fi&1 != 0
}

// A test-only mathematical oracle: each shared assignment admits at most two
// distinct rows. Production uses the original SAT construction and solver.
func mappingSATReviewSatisfiable(c mappingSATReviewCase) bool {
	b, s, f := mappingSATReviewGroups(c)
	for shared := 0; shared < 1<<len(s); shared++ {
		patterns := map[uint64]bool{}
		for bound := 0; bound < 1<<len(b); bound++ {
			var pattern uint64
			for free := 0; free < 1<<len(f); free++ {
				var assignment int
				for i, pin := range s {
					assignment |= (shared >> i & 1) << pin
				}
				for i, pin := range b {
					assignment |= (bound >> i & 1) << pin
				}
				for i, pin := range f {
					assignment |= (free >> i & 1) << pin
				}
				pattern |= (c.truth[assignment>>6] >> uint(assignment&63) & 1) << free
			}
			patterns[pattern] = true
		}
		if len(patterns) > 2 {
			return false
		}
	}
	return true
}
func mappingSATReviewCases() []mappingSATReviewCase {
	r := rand.New(rand.NewSource(902101))
	var result []mappingSATReviewCase
	for size := int32(2); size <= 6; size++ {
		for shared := int32(0); shared < size; shared++ {
			for bound := int32(0); bound <= size-shared; bound++ {
				freeCounts := []int32{0, size - shared - 1}
				if freeCounts[1] == 0 {
					freeCounts = freeCounts[:1]
				}
				for _, free := range freeCounts {
					c := mappingSATReviewCase{size: size, variables: bound + shared + free, outputs: 3}
					permutation := r.Perm(int(c.variables))
					for i := int32(0); i < shared; i++ {
						c.set |= 3 << uint(2*permutation[i])
					}
					for i := shared; i < shared+bound; i++ {
						c.set |= 1 << uint(2*permutation[i])
					}
					words := (1<<uint(c.variables) + 63) / 64
					c.truth = make([]uint64, words)
					bt, ft := r.Uint64(), r.Uint64()
					for bit := 0; bit < words*64; bit++ {
						assignment := bit & ((1 << uint(c.variables)) - 1)
						if mappingSATReviewCompose(c, bt, ft, assignment) {
							c.truth[bit>>6] |= 1 << uint(bit&63)
						}
					}
					result = append(result, c)
					random := c
					random.outputs = 1 + len(result)%3
					random.truth = make([]uint64, words)
					for i := range random.truth {
						random.truth[i] = r.Uint64()
					}
					result = append(result, random)
					complement := c
					complement.truth = make([]uint64, words)
					for i, v := range c.truth {
						complement.truth[i] = ^v
					}
					result = append(result, complement)
				}
			}
		}
	}
	return result
}
func TestMappingSATNativeConstructionReuseAndModels(t *testing.T) {
	exe := gia.CompileSourceOracleForTest(t, "if_sat_matrix_review.c")
	cases := mappingSATReviewCases()
	var input strings.Builder
	for size := 2; size <= 6; size++ {
		fmt.Fprintf(&input, "B X %d\nB Z %d\n", size, size)
	}
	for _, c := range cases {
		fmt.Fprintf(&input, "C %d %d %x %d", c.size, c.variables, c.set, c.outputs)
		for _, word := range c.truth {
			fmt.Fprintf(&input, " %016x", word)
		}
		input.WriteByte('\n')
	}
	path := filepath.Join(t.TempDir(), "sat-cases.txt")
	if err := os.WriteFile(path, []byte(input.String()), 0600); err != nil {
		t.Fatal(err)
	}
	source, err := exec.Command(exe, path).CombinedOutput()
	if err != nil {
		t.Fatalf("native SAT matrix: %v %s", err, source)
	}
	var got strings.Builder
	solvers := map[int32]any{}
	defer func() {
		for _, solver := range solvers {
			if_.If_ManSatUnbuild(solver)
		}
	}()
	for size := int32(2); size <= 6; size++ {
		for _, kind := range []byte{'X', 'Z'} {
			var handle any
			if kind == 'X' {
				handle = if_.If_ManSatBuildXY(size)
				solvers[size] = handle
			} else {
				handle = if_.If_ManSatBuildXYZ(size)
			}
			solver, ok := handle.(*bsat.Solver)
			if !ok {
				t.Fatalf("factory returned %T instead of SAT", handle)
			}
			fmt.Fprintf(&got, "build %c %d %d %d\n", kind, size, solver.NumVars(), solver.Statistics().Clauses)
			if kind == 'Z' {
				if_.If_ManSatUnbuild(handle)
			}
		}
	}
	scratch := gia.Vec_IntAlloc(2048)
	sat, unsat := 0, 0
	for index, c := range cases {
		bound, free := uint64(0x0123456789abcdef), uint64(0xfedcba9876543210)
		var bp, fp *uint64
		if c.outputs&1 != 0 {
			bp = &bound
		}
		if c.outputs&2 != 0 {
			fp = &free
		}
		result := if_.If_ManSatCheckXY(solvers[c.size], c.size, c.truth, c.variables, c.set, bp, fp, scratch)
		if (result == 1) != mappingSATReviewSatisfiable(c) {
			t.Fatalf("case%d K%d vars%d set%x SAT%d disagrees with independent relation", index, c.size, c.variables, c.set, result)
		}
		if result == 1 {
			sat++
		} else {
			unsat++
		}
		if result == 1 && c.outputs == 3 {
			for assignment := 0; assignment < 1<<uint(c.variables); assignment++ {
				want := c.truth[assignment>>6]>>uint(assignment&63)&1 != 0
				if mappingSATReviewCompose(c, bound, free, assignment) != want {
					t.Fatalf("case%d assignment%d extracted model truth mismatch", index, assignment)
				}
			}
		}
		fmt.Fprintf(&got, "check %d %d %016x %016x assumptions", index, result, bound, free)
		for _, lit := range gia.Vec_IntArray(scratch) {
			fmt.Fprintf(&got, " %d", lit)
		}
		got.WriteByte('\n')
	}
	a, b := bufio.NewScanner(strings.NewReader(string(source))), bufio.NewScanner(strings.NewReader(got.String()))
	line := 0
	for a.Scan() {
		line++
		if !b.Scan() {
			t.Fatalf("missing Go line%d", line)
		}
		if a.Text() != b.Text() {
			want, actual := strings.Fields(a.Text()), strings.Fields(b.Text())
			at := 0
			for at < min(len(want), len(actual)) && want[at] == actual[at] {
				at++
			}
			t.Fatalf("native SAT line%d first differing token%d\nC: %.220s\nGo: %.220s", line, at, a.Text(), b.Text())
		}
	}
	if err := a.Err(); err != nil {
		t.Fatal(err)
	}
	if b.Scan() {
		t.Fatal("extra Go SAT records")
	}
	if err := b.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d native SAT checks (%d SAT/%d UNSAT), K2..6 XY/XYZ shapes, reused models and complete assumptions", len(cases), sat, unsat)
}
