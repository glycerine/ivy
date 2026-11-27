package vec

import "testing"

// TestPtrMatcherAgreesWithPtrEqual: the pointer fast path of Vec_PtrFind and
// Vec_PtrRemove answers as ptrEqual does for pointers, typed nils, nil,
// values of other types and uncomparable entries.
func TestPtrMatcherAgreesWithPtrEqual(t *testing.T) {
	a, b := new(int), new(int)
	var typedNil *int
	entries := []any{a, b, typedNil, nil, 3, "s", []int{1}, new(string)}
	for _, x := range entries {
		match := ptrMatcher(x)
		for _, e := range entries {
			if got, want := match(e), ptrEqual(e, x); got != want {
				t.Fatalf("ptrMatcher(%v)(%v) = %v, ptrEqual = %v", x, e, got, want)
			}
		}
	}
	v := Vec_PtrAlloc(4)
	Vec_PtrPush(v, a)
	Vec_PtrPush(v, b)
	if Vec_PtrFind(v, b) != 1 {
		t.Fatal("find")
	}
	Vec_PtrRemove(v, a)
	if Vec_PtrSize(v) != 1 || Vec_PtrEntry(v, 0) != b {
		t.Fatal("remove")
	}
}
