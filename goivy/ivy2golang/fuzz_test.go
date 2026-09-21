package ivy2golang

import (
	"strings"
	"testing"
)

func FuzzGoNameDoesNotPanic(f *testing.F) {
	for _, seed := range []string{"", "x", "ext:step", "type", "0bad", "a.b[c]", "__fml:x"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got := goName(s)
		if got == "" {
			t.Fatalf("goName(%q) returned empty string", s)
		}
		if strings.HasPrefix(s, "\"") {
			return
		}
		if got[0] >= '0' && got[0] <= '9' {
			t.Fatalf("goName(%q) starts with a digit: %q", s, got)
		}
		if goKeywords[got] {
			t.Fatalf("goName(%q) returned keyword %q", s, got)
		}
	})
}

func FuzzParseArgsDoesNotPanic(f *testing.F) {
	for _, seed := range []string{"target=test", "classname=X", "bad", "=bad", "trace=true"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _, _ = ParseArgs([]string{s, "x.ivy"})
		_, _, _ = ParseArgs([]string{s})
	})
}
