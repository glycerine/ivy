package bmc

import "testing"

func TestPacked2FieldsRoundTripAndIsolation(t *testing.T) {
	var word uint32
	for id := uint32(0); id < 16; id++ {
		value := (id*3 + 1) & 3
		word = packed2Set(word, id, value)
	}
	for id := uint32(0); id < 16; id++ {
		want := (id*3 + 1) & 3
		if got := packed2Get(word, id); got != want {
			t.Fatalf("field %d = %d, want %d", id, got, want)
		}
	}
	word = packed2Set(word, 7, 2)
	for id := uint32(0); id < 16; id++ {
		want := (id*3 + 1) & 3
		if id == 7 {
			want = 2
		}
		if got := packed2Get(word, id); got != want {
			t.Fatalf("after replacing field 7, field %d = %d, want %d", id, got, want)
		}
	}
}

func TestPacked2AddUsesBitwiseOr(t *testing.T) {
	var word uint32
	word = packed2Set(word, 0, 1)
	word = packed2Add(word, 0, 2)
	if got := packed2Get(word, 0); got != 3 {
		t.Fatalf("OR addition yielded %d, want 3", got)
	}
	word = packed2Add(word, 1, 3)
	if got := packed2Get(word, 1); got != 3 {
		t.Fatalf("new field yielded %d, want 3", got)
	}
	if got := packed2Get(word, 2); got != 0 {
		t.Fatalf("unrelated field changed to %d", got)
	}
}
