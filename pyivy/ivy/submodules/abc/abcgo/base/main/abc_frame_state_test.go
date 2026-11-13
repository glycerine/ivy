package abcmain

import "testing"

func TestABCFrameStateReplacementOwnershipAndClearing(t *testing.T) {
	s := &ABCFrameState{}
	cex := any(&struct{ id int }{1})
	s.ReplaceCex(&cex)
	if cex != nil || s.Cex() == nil {
		t.Fatal("CEX replacement did not transfer ownership")
	}

	vec := any(&struct{ id int }{2})
	s.ReplaceCexVec(&vec)
	if vec != nil || s.Cex() != nil || s.CexVec() == nil {
		t.Fatal("CEX-vector replacement did not clear CEX and transfer ownership")
	}

	cex = any(&struct{ id int }{3})
	s.ReplaceCex(&cex)
	if cex != nil || s.CexVec() != nil || s.Cex() == nil {
		t.Fatal("CEX replacement did not clear CEX vector")
	}
}

func TestABCFrameStateRetainsOpaqueVectorValues(t *testing.T) {
	s := &ABCFrameState{}
	equivs := any(new(struct{ entries []any }))
	want := equivs
	s.ReplacePoEquivs(&equivs)
	if equivs != nil || s.PoEquivs() == nil {
		t.Fatal("PO-equivalence replacement did not transfer ownership")
	}

	statuses := any(new(struct{}))
	s.ReplaceStatuses(&statuses)
	if statuses != nil || s.Statuses() == nil {
		t.Fatal("status replacement did not transfer ownership")
	}
	if s.PoEquivs() != want {
		t.Fatal("opaque value identity was not preserved")
	}
}
