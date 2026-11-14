package abci

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func outputInteger(t *testing.T, output, pattern string) int {
	t.Helper()
	match := regexp.MustCompile(pattern).FindStringSubmatch(output)
	if len(match) != 2 {
		t.Fatalf("missing %q in %q", pattern, output)
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestQbfExhaustiveParameterAndInputFunctions(t *testing.T) {
	for truth := uint64(0); truth < 16; truth++ {
		n := truthNetwork(t, 2, truth)
		output := capturePrintStdout(t, func() { Abc_NtkQbf(n, 1, 4, 0, 0) })
		abc.Abc_NtkDelete(n)
		want := truth&5 == 5 || truth&10 == 10 // exists x0, forall x1
		if strings.Contains(output, "Solved after") != want || strings.Contains(output, "Implementation does not exist") == want {
			t.Fatalf("truth=%x synthesis result: %s", truth, output)
		}
		if want {
			parameter := outputInteger(t, output, `Parameters:\s*([01])`)
			for input := 0; input < 2; input++ {
				if truth>>(parameter+2*input)&1 == 0 {
					t.Fatalf("truth=%x invalid parameter %d", truth, parameter)
				}
			}
		}
	}
}

func TestUnatenessAndSymmetryReporting(t *testing.T) {
	for _, tc := range []struct {
		truth            uint64
		unate, symmetric int
	}{{0x80, 3, 3}, {0x96, 0, 3}, {0xac, 2, 0}} {
		for _, naive := range []int32{0, 1} {
			t.Run(fmt.Sprintf("%x/naive=%d", tc.truth, naive), func(t *testing.T) {
				n := truthNetwork(t, 3, tc.truth)
				defer abc.Abc_NtkDelete(n)
				output := capturePrintStdout(t, func() { Abc_NtkPrintUnate(n, 1, naive, 1) })
				if got := outputInteger(t, output, `Total unate =\s*(\d+)`); got != tc.unate {
					t.Fatalf("unate=%d want=%d: %s", got, tc.unate, output)
				}
				output = capturePrintStdout(t, func() { Abc_NtkSymmetries(n, 1, naive, 1, 1) })
				if got := outputInteger(t, output, `two-variable symmetries\s*=\s*(\d+)`); got != tc.symmetric {
					t.Fatalf("symmetries=%d want=%d: %s", got, tc.symmetric, output)
				}
				if Abc_NtkGlobalBddMan(n) != nil {
					t.Fatal("analysis retained temporary global BDDs")
				}
				for bits := 0; bits < 8; bits++ {
					if got := evaluateNetwork(t, n, bits)[0]; got != (tc.truth>>bits&1 != 0) {
						t.Fatal("analysis changed original network")
					}
				}
			})
		}
	}
}

func TestAutosymmetryReportsParityDimension(t *testing.T) {
	for _, output := range []int32{-1, 0} {
		n := truthNetwork(t, 3, 0x96)
		text := capturePrintStdout(t, func() { Abc_NtkAutoPrint(n, output, 0, 0) })
		if got := outputInteger(t, text, `AutoK =\s*(\d+)`); got != 2 {
			t.Fatalf("parity autosymmetry dimension=%d: %s", got, text)
		}
		if strings.Contains(text, "NOT EQUAL") {
			t.Fatal(text)
		}
		if Abc_NtkGlobalBddMan(n) != nil {
			t.Fatal("autosymmetry retained BDD manager")
		}
		abc.Abc_NtkDelete(n)
	}
}
