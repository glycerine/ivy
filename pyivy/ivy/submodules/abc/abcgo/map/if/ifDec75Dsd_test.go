package if_

import "testing"

func TestDsdDecExistBoundsets(t *testing.T) {
	tests := []struct {
		text string
		want int32
	}{
		{"a", 0},
		{"(ab)", 1 << 2},
		{"(abc)", (1 << 2) | (1 << 3)},
		{"[abc]", (1 << 2) | (1 << 3)},
		{"<abc>", 1 << 3},
		{"(<ab>c)", (1 << 2) | (1 << 3)},
		{"(abc)\x00ignored", (1 << 2) | (1 << 3)},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			if got := Dau_DsdCheckDecExist(test.text); got != test.want {
				t.Fatalf("Dau_DsdCheckDecExist(%q) = %#x, want %#x", test.text, got, test.want)
			}
		})
	}
}

func TestDsdAndDecExistBoundsets(t *testing.T) {
	tests := []struct {
		text string
		want int32
	}{
		{"a", 1},
		{"(abc)", 1 << 2},
		{"(a(bc))", 1 << 2},
		{"<abc>", 0},
		{"[abc]", 0},
		{"((ab)c)", (1 << 2)},
		{"(a(bc)d)", (1 << 2) | (1 << 3)},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			if got := Dau_DsdCheckDecAndExist(test.text); got != test.want {
				t.Fatalf("Dau_DsdCheckDecAndExist(%q) = %#x, want %#x", test.text, got, test.want)
			}
		})
	}
}

func TestDsdCheckRejectsBadExpressions(t *testing.T) {
	for _, text := range []string{"(a", "(a]", "()", "(a)"} {
		t.Run(text, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("invalid DSD expression %q did not panic", text)
				}
			}()
			Dau_DsdCheckDecExist(text)
		})
	}
}
