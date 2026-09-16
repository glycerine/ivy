package goivy

import (
	"strings"
	"testing"
)

func TestIvyErrorFormatsReferenceChainLikePython(t *testing.T) {
	err := &IvyError{
		Msg: "broken definition",
		Loc: Location{
			Filename: "caller.ivy",
			Line:     7,
			Reference: &Location{
				Filename: "library.ivy",
				Line:     3,
			},
		},
		HasLoc: true,
	}

	got := err.Error()
	wantFirst := "library.ivy: line 3: error: broken definition"
	wantSecond := "caller.ivy: line 7: error: instantiated here"
	if !strings.Contains(got, wantFirst) || !strings.Contains(got, wantSecond) {
		t.Fatalf("IvyError.Error() = %q, want %q and %q", got, wantFirst, wantSecond)
	}
}
