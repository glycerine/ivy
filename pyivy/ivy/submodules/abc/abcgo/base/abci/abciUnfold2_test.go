package abci

import "testing"

func TestParseUnfold2OptionsDefaultsAndValues(t *testing.T) {
	nConfs, nProps, verbose, ok := parseUnfold2Options(1, []string{"unfold2"})
	if !ok || nConfs != 1000 || nProps != 1000 || verbose != 0 {
		t.Fatalf("defaults = (%d,%d,%d,%t)", nConfs, nProps, verbose, ok)
	}
	nConfs, nProps, verbose, ok = parseUnfold2Options(6, []string{"unfold2", "-C", "17", "-P", "19", "-v"})
	if !ok || nConfs != 17 || nProps != 19 || verbose != 1 {
		t.Fatalf("parsed values = (%d,%d,%d,%t)", nConfs, nProps, verbose, ok)
	}
}

func TestParseUnfold2OptionsRejectsMissingNegativeAndUnsupportedValues(t *testing.T) {
	for _, args := range [][]string{
		{"unfold2", "-C"},
		{"unfold2", "-P"},
		{"unfold2", "-C", "-1"},
		{"unfold2", "-P", "-3"},
		{"unfold2", "-h"},
		{"unfold2", "-F", "2"},
	} {
		if _, _, _, ok := parseUnfold2Options(int32(len(args)), args); ok {
			t.Errorf("accepted invalid arguments %v", args)
		}
	}
}
