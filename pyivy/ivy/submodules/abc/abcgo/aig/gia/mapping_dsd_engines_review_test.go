package gia_test

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/opt/dau"
)

type mappingDSDReviewCase struct {
	variables int32
	truth     []uint64
}

func mappingDSDReviewFullSupport(truth []uint64, n int32) bool {
	for pin := int32(0); pin < n; pin++ {
		dependent := false
		for assignment := 0; assignment < 1<<uint(n); assignment++ {
			other := assignment ^ (1 << uint(pin))
			if truth[assignment>>6]>>uint(assignment&63)&1 != truth[other>>6]>>uint(other&63)&1 {
				dependent = true
				break
			}
		}
		if !dependent {
			return false
		}
	}
	return true
}
func mappingDSDReviewCases() []mappingDSDReviewCase {
	var cases []mappingDSDReviewCase
	for _, text := range []string{"(abc)", "[abc]", "<abc>", "(!a(bc))", "((ab)[cd])", "<a(bc)[de]>", "((ab)(cd)(ef))", "[a(bc)<def>]", "E8{abc}", "<a(bc)(def)>", "(ab[cd]<efg>)", "<a(bcd)[efgh]>", "(abD3C5A79F01234567{cdefgh})", "(ab9E27B4D8603FA15C{cdefgh})", "(ab6D93B71CA8E50F24{cdefgh})"} {
		n := dau.Dau_DsdFindVarNum(text)
		cases = append(cases, mappingDSDReviewCase{n, dau.Dau_DsdToTruth(text, n)})
	}
	random := rand.New(rand.NewSource(2938107))
	for n := int32(3); n <= 8; n++ {
		for i := 0; i < 12; i++ {
			truth := make([]uint64, max(1, 1<<uint(max(n-6, 0))))
			for {
				for j := range truth {
					truth[j] = random.Uint64()
				}
				if n < 6 {
					bits := uint(1) << uint(n)
					truth[0] &= (uint64(1) << bits) - 1
					for shift := bits; shift < 64; shift *= 2 {
						truth[0] |= truth[0] << shift
					}
				}
				if mappingDSDReviewFullSupport(truth, n) {
					break
				}
			}
			cases = append(cases, mappingDSDReviewCase{n, truth})
		}
	}
	originals := append([]mappingDSDReviewCase(nil), cases...)
	for _, c := range originals {
		inverted := append([]uint64(nil), c.truth...)
		for i := range inverted {
			inverted[i] = ^inverted[i]
		}
		cases = append(cases, c, mappingDSDReviewCase{c.variables, inverted})
	}
	return cases
}

func TestMappingDSDNativeCanonicalTreesAndPartitions(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_matrix_review.c")
	cases := mappingDSDReviewCases()
	var input strings.Builder
	for _, c := range cases {
		fmt.Fprintf(&input, "%d", c.variables)
		for _, word := range c.truth {
			fmt.Fprintf(&input, " %016x", word)
		}
		input.WriteByte('\n')
	}
	path := filepath.Join(t.TempDir(), "functions.txt")
	if err := os.WriteFile(path, []byte(input.String()), 0600); err != nil {
		t.Fatal(err)
	}
	source, err := testoracle.Command(t, executable, path).CombinedOutput()
	if err != nil {
		t.Fatalf("native DSD: %v %s", err, source)
	}
	p := if_.If_DsdManAlloc(8, 4)
	defer if_.If_DsdManFree(p, 0)
	var got strings.Builder
	for index, c := range cases {
		permutation := make([]byte, c.variables)
		lit := if_.If_DsdManCompute(p, c.truth, c.variables, permutation, "44")
		reconstructed := if_.If_DsdManComputeTruth(p, lit, permutation)
		for i, word := range c.truth {
			if reconstructed[i] != word {
				t.Fatalf("case%d truth word%d reconstructed incorrectly", index, i)
			}
		}
		fmt.Fprintf(&got, "function %d %d %d permutation", index, lit, if_.If_DsdManObjNum(p))
		for _, pin := range permutation {
			fmt.Fprintf(&got, " %d", pin)
		}
		got.WriteString(" partitions")
		for i, mask := range []uint32{0, 0, 1, ^uint32(0)} {
			derive, effort := int32(0), int32(0)
			if i != 0 {
				derive = 1
			}
			if i >= 2 && c.variables <= 7 {
				effort = 1
			}
			fmt.Fprintf(&got, " %x", if_.If_DsdManCheckXY(p, lit, 4, derive, mask, effort, 0))
		}
		got.WriteByte('\n')
	}
	want, actual := strings.Split(string(source), "\n"), strings.Split(got.String(), "\n")
	if len(want) != len(actual) {
		t.Fatalf("DSD records C%d Go%d", len(want), len(actual))
	}
	for i, line := range want {
		if line != actual[i] {
			t.Fatalf("DSD case%d C%s; Go%s", i, line, actual[i])
		}
	}
	t.Logf("%d native canonical functions, repeated/complemented interning and %d ordered partition checks", len(cases), 4*len(cases))
}
