package goivy

import "testing"

// Python's Redefinition, Circular, NoMatch, ProofError and CaptureError
// (ivy_proof.py) are plain iu.IvyError subclasses: str(e) is
// str(lineno) + 'error: ' + msg, with no type prefix. goivy printed
// "proof error: msg (at node)" etc. (test/conjpf1.ivy).
func TestProofErrorsFormatLikeIvyError(t *testing.T) {
	cfg := NewAstConfig()
	at := cfg.NewAtom("foo.inv1")
	at.SetLineno(Location{Filename: "f.ivy", Line: 34})
	errs := []interface {
		error
		HasLocation() bool
	}{
		&ProofError{Node: at, Msg: "m"},
		&NoMatch{Node: at, Msg: "m"},
		&CaptureError{Node: at, Msg: "m"},
		&Redefinition{Node: at, Msg: "m"},
		&Circular{Node: at, Msg: "m"},
	}
	for _, e := range errs {
		if e.Error() != "f.ivy: line 34: error: m" || !e.HasLocation() {
			t.Fatalf("%T: %q", e, e.Error())
		}
	}
	if e := (&ProofError{Msg: "m"}); e.Error() != "m" || e.HasLocation() {
		t.Fatalf("unlocated: %q", e.Error())
	}
}
