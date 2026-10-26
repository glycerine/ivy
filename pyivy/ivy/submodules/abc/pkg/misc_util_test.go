package abc

import (
	"errors"
	"math"
	"testing"
)

func TestLiteralDomains(t *testing.T) {
	for i := uint32(0); i < 1024; i++ {
		for _, neg := range []bool{false, true} {
			l, err := MakeAigLit(AigID(i), neg)
			if err != nil {
				t.Fatal(err)
			}
			if l.ID() != AigID(i) || l.Complemented() != neg || l.Not().Not() != l || l.NotIf(neg) != l.Regular() || !l.Valid() {
				t.Fatalf("bad AIG literal %d", l)
			}
			s, err := MakeSatLit(SatVar(i), neg)
			if err != nil {
				t.Fatal(err)
			}
			if s.Var() != SatVar(i) || s.Negative() != neg || s.Not().Not() != s || !s.Valid() {
				t.Fatalf("bad SAT literal %d", s)
			}
		}
	}
	for _, id := range []uint32{MaxLiteralID, MaxLiteralID + 1, math.MaxUint32} {
		a, err := MakeAigLit(AigID(id), true)
		if id == MaxLiteralID {
			if err != nil || a != math.MaxInt32 {
				t.Fatal(a, err)
			}
		} else if !errors.Is(err, ErrInvalidLiteral) || a != InvalidAigLit {
			t.Fatal(a, err)
		}
		_, err = MakeSatLit(SatVar(id), false)
		if (err == nil) != (id == MaxLiteralID) {
			t.Fatal(id, err)
		}
	}
	if InvalidAigLit.Valid() || AigLit(math.MaxUint32-1).Valid() {
		t.Fatal("invalid sentinel accepted")
	}
	if AigConst1.Not() != AigConst0 || AigConst0.Not() != AigConst1 {
		t.Fatal("constant polarity")
	}
}
func TestBitWordCountAndSize(t *testing.T) {
	for _, width := range []int{8, 32, 64} {
		for n := 0; n < 140; n++ {
			got, err := BitWordCount(n, width)
			want := (n + width - 1) / width
			if err != nil || got != want {
				t.Fatal(n, width, got, err)
			}
		}
		got, err := BitWordCount(math.MaxInt, width)
		if err != nil || got != math.MaxInt/width+1 {
			t.Fatal(got, err)
		}
	}
	for _, n := range []int{-1, 31, 0, 128} {
		if _, err := BitWordCount(1, n); err == nil {
			t.Fatal("accepted width", n)
		}
	}
	if _, err := BitWordCount(-1, 32); err == nil {
		t.Fatal("accepted negative bits")
	}
	for _, c := range []struct {
		base, count, width, limit, want int
		ok                              bool
	}{
		{0, 0, 0, 0, 0, true}, {3, 2, 4, 11, 11, true}, {3, 2, 5, 11, 0, false},
		{math.MaxInt, 0, math.MaxInt, math.MaxInt, math.MaxInt, true},
		{1, math.MaxInt, math.MaxInt, math.MaxInt, 0, false}, {-1, 0, 0, 2, 0, false},
	} {
		got, err := checkedSize(c.base, c.count, c.width, c.limit)
		if (err == nil) != c.ok || (c.ok && got != c.want) {
			t.Fatal(c, got, err)
		}
	}
}
