package goivy

import "testing"

func TestArtInterpStateAdaptersPreserveUnders(t *testing.T) {
	mod := New()
	pred := NewState(mod, TrueClauses(nil))
	under := NewState(mod, FalseClauses(nil))
	pred.Unders = []*State{under}
	post := NewState(mod, TrueClauses(nil))
	post.Pred = pred

	interpPost := ArtToInterpState(post)
	if interpPost.Pred() == nil {
		t.Fatal("ArtToInterpState dropped predecessor")
	}
	if got := len(interpPost.Pred().Unders()); got != 1 {
		t.Fatalf("ArtToInterpState predecessor unders = %d, want 1", got)
	}

	artPost := InterpToArtState(interpPost)
	if artPost.Pred == nil {
		t.Fatal("InterpToArtState dropped predecessor")
	}
	if got := len(artPost.Pred.Unders); got != 1 {
		t.Fatalf("InterpToArtState predecessor unders = %d, want 1", got)
	}
}
