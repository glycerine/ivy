package goivy

import "testing"

// Python strip_map_lookup (ivy_isolate.py) iterates strip_map (a dict) and
// returns the parameters of the first prefix that matches, so when two
// prefixes match ("a" and "a.b" for "a.b.c") insertion order decides.
// goivy's StripMap was a Go map, so the winner was random.
func TestStripMapLookupFirstMatchingPrefixInOrder(t *testing.T) {
	m := mkModuleWithSig()
	for i := 0; i < 30; i++ {
		ab := StripMapOf(StripPair{"a", []string{"p"}}, StripPair{"a.b", []string{"q"}})
		if got := StripMapLookup("a.b.c", ab, m); len(got) != 1 || got[0] != "p" {
			t.Fatalf("run %d: a first: got %v, want [p]", i, got)
		}
		ba := StripMapOf(StripPair{"a.b", []string{"q"}}, StripPair{"a", []string{"p"}})
		if got := StripMapLookup("a.b.c", ba, m); len(got) != 1 || got[0] != "q" {
			t.Fatalf("run %d: a.b first: got %v, want [q]", i, got)
		}
	}
	if got := StripMapLookup("a.b.c", nil, m); len(got) != 0 {
		t.Fatalf("nil strip map: got %v, want none", got)
	}
}
