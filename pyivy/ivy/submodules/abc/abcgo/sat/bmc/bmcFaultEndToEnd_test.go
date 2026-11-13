package bmc

import (
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Exercise the actual test generator and compare its emitted patterns with an
// independent truth table, rather than the fault-unfolding implementation.
func TestFaultDriverDetectsSingleGateFaults(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	for _, tc := range []struct {
		name      string
		algo      int32
		form      string
		basic     int32
		card      int32
		nonstrict int32
		masks     []int
	}{
		{"formula", 0, "((a&b)^p)", 0, 0, 0, []int{15}},
		{"stuck", 2, "", 1, 0, 0, []int{8, 7}},
		{"stuck exact one", 2, "", 0, 1, 0, []int{8, 7}},
		{"stuck at most one", 2, "", 0, 1, 1, []int{8, 7}},
		{"flip", 3, "", 0, 0, 0, []int{15}},
		{"function", 4, "", 1, 0, 0, []int{1, 2, 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Remove("tests.txt")
			Gia_ManFaultTest(p, p, &Bmc_ParFf_t{Algo: tc.algo, pFormStr: tc.form, fBasic: tc.basic, nCardConstr: tc.card, fNonStrict: tc.nonstrict, fDump: 1})
			data, err := os.ReadFile("tests.txt")
			if err != nil {
				t.Fatalf("generated suite: %v", err)
			}
			patterns := strings.Fields(string(data))
			coverage := 0
			for _, row := range patterns {
				if len(row) != 2 || (row[0] != '0' && row[0] != '1') || (row[1] != '0' && row[1] != '1') {
					t.Fatalf("invalid pattern %q", row)
				}
				index := int(row[0]-'0') + 2*int(row[1]-'0')
				coverage |= 1 << index
			}
			for _, detectable := range tc.masks {
				if coverage&detectable == 0 {
					t.Fatalf("suite %q fails to detect fault; patterns=%04b detecting=%04b", data, coverage, detectable)
				}
			}
		})
	}
}

func TestFaultDriverDetectsDelayFault(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	p := gia.Gia_ManStart(8)
	x, r := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, x, r)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManSetRegNum(p, 1)
	for _, ffOnly := range []int32{0, 1} {
		_ = os.Remove("tests.txt")
		Gia_ManFaultTest(p, p, &Bmc_ParFf_t{Algo: 1, fFfOnly: ffOnly, fDump: 1})
		data, err := os.ReadFile("tests.txt")
		if err != nil {
			t.Fatalf("ffOnly=%d: %v", ffOnly, err)
		}
		detected := false
		for _, row := range strings.Fields(string(data)) {
			if len(row) != 3 {
				t.Fatalf("ffOnly=%d invalid pattern %q", ffOnly, row)
			}
			// Input ordering is initial register, first-frame PI, second-frame PI.
			r0, x0, x1 := row[0] == '1', row[1] == '1', row[2] == '1'
			first := r0 && x0
			second := first && x1
			if first != second {
				detected = true
			}
		}
		if !detected {
			t.Fatalf("ffOnly=%d suite %q misses the delay fault", ffOnly, data)
		}
	}
}

func TestFaultDriverExtendsPriorPatterns(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	if err = os.WriteFile("prior.txt", []byte("11\n"), 0600); err != nil {
		t.Fatal(err)
	}
	Gia_ManFaultTest(p, p, &Bmc_ParFf_t{Algo: 2, fBasic: 1, pFileName: "prior.txt", fDump: 1})
	data, err := os.ReadFile("tests.txt")
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Fields(string(data))
	if len(rows) < 2 || rows[0] != "11" {
		t.Fatalf("prior pattern not preserved and extended: %q", data)
	}
	found := false
	for _, row := range rows[1:] {
		if row != "11" {
			found = true
		}
	}
	if !found {
		t.Fatalf("no pattern detects stuck-at-one: %q", data)
	}
}
