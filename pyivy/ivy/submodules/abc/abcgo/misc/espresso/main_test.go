package espresso

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func plaTruthText(truth uint32) string {
	var b strings.Builder
	b.WriteString(".i 3\n.o 1\n")
	for value := uint32(0); value < 8; value++ {
		if truth&(1<<value) != 0 {
			fmt.Fprintf(&b, "%d%d%d 1\n", value&1, (value>>1)&1, (value>>2)&1)
		}
	}
	b.WriteString(".e\n")
	return b.String()
}
func minimumThreeInputCubes(truth uint32) int32 {
	if truth == 0 {
		return 0
	}
	var products []uint32
	for pattern := 0; pattern < 27; pattern++ {
		digits := [3]int{}
		n := pattern
		for v := range digits {
			digits[v] = n % 3
			n /= 3
		}
		var covered uint32
		for value := 0; value < 8; value++ {
			match := true
			for v, digit := range digits {
				if digit < 2 && ((value>>v)&1) != digit {
					match = false
				}
			}
			if match {
				covered |= 1 << value
			}
		}
		if covered&^truth == 0 {
			products = append(products, covered)
		}
	}
	dp := [256]int32{}
	for i := 1; i < len(dp); i++ {
		dp[i] = 99
	}
	for subset := uint32(0); subset < 256; subset++ {
		if subset&^truth != 0 {
			continue
		}
		for _, p := range products {
			union := subset | p
			if dp[subset]+1 < dp[union] {
				dp[union] = dp[subset] + 1
			}
		}
	}
	return dp[truth]
}
func TestRunAllThreeInputFunctionsAndExactOptimality(t *testing.T) {
	for truth := uint32(0); truth < 256; truth++ {
		for _, command := range []string{"ESPRESSO", "exact"} {
			var output bytes.Buffer
			if err := Run([]string{"-D" + command}, strings.NewReader(plaTruthText(truth)), &output); err != nil {
				t.Fatalf("truth%02x %s: %v", truth, command, err)
			}
			pla := parsePLATest(output.String())
			for value := int32(0); value < 8; value++ {
				got := false
				for row := int32(0); row < pla.F.count; row++ {
					got = got || matchesBinary(GETSET(pla.F, row), value, 3)
				}
				if got != (truth&(1<<value) != 0) {
					t.Fatalf("truth%02x %s value%d output%s", truth, command, value, output.String())
				}
			}
			if command == "exact" {
				if want := minimumThreeInputCubes(truth); pla.F.count != want {
					t.Fatalf("truth%02x exact cubes%d want%d", truth, pla.F.count, want)
				}
			}
		}
	}
}
func TestRunCompatibilityRoundTripAndMany(t *testing.T) {
	input := ".i 1\n.o 1\n0 1\n.e\n"
	var out bytes.Buffer
	if err := Run([]string{"-do", "echo", "-out", "fdr", "-eat"}, strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	pla := parsePLATest(out.String())
	if pla.F.count != 1 || pla.R.count != 1 {
		t.Fatal(out.String())
	}
	out.Reset()
	if err := Run([]string{"-Dmany"}, strings.NewReader(input+input), &out); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), ".e\n") != 2 {
		t.Fatal(out.String())
	}
}
func TestRunErrorsAndContextRestoration(t *testing.T) {
	ConfigureCube(2, []int32{3})
	saved := cube.fullset
	for _, args := range [][]string{{"-Dmissing"}, {"-omissing"}, {"-emissing"}, {"-Sx"}, {"-rno"}, {"-D"}, {"-z"}} {
		if err := Run(args, strings.NewReader(""), &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted bad options %v", args)
		}
		if cube.size != 7 || !setp_equal(cube.fullset, saved) {
			t.Fatal("Run changed existing cube context")
		}
	}
	var out bytes.Buffer
	if err := Run([]string{"--help"}, nil, &out); err != nil || !strings.Contains(out.String(), "SYNOPSIS") {
		t.Fatal(err, out.String())
	}
}
func TestVerifySharedRowsAndNamedPermutation(t *testing.T) {
	ConfigureCube(2, []int32{1})
	f := maskFamily(cube.size, 21, 26)
	empty := new_cover(0)
	if verify(f, f, empty) {
		t.Fatal("a cover must verify against itself despite shared row identities")
	}
	first := &PLA_t{F: f, D: empty, R: new_cover(0), label: []string{"a.bar", "a", "b.bar", "b", "y"}}
	second := &PLA_t{F: maskFamily(cube.size, 21, 26), D: new_cover(0), R: new_cover(0), label: []string{"b.bar", "b", "a.bar", "a", "y"}}
	if PLA_verify(first, second) {
		t.Fatal("named symmetric function permutation failed")
	}
}
