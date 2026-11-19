package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

// Python DomainSetup.named compiles the name's arguments inside
// ivy_logic.UnsortedContext(), so a variable of an undeclared sort S is
// accepted (test/proving7a.ivy: "property exists Y. succ(X,Y) named next(X)").
func TestNamedArgsCompileUnsorted(t *testing.T) {
	src := "#lang ivy1.7\n\ntype t\nrelation succ(X:t,Y:t)\n\naxiom exists Y. succ(X,Y)\n\nproperty exists Y. succ(X,Y) named next(X)\n"
	fn := filepath.Join(t.TempDir(), "named.ivy")
	if err := os.WriteFile(fn, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := IvyLoadFile(fn); err != nil {
		t.Fatal(err)
	}
}
