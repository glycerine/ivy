package acec

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// The source fixture includes a complemented high bit so signed and unsigned
// coefficient signs differ. C prints the three resulting monomials below.
func TestSourcePriorityPolynomialSignedAndUnsigned(t *testing.T) {
	p := gia.Gia_ManStart(32)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendAnd(p, a, b)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManAppendCo(p, a^1)
	order := Gia_PolynReorder(p, 0, 0)
	for _, tc := range []struct {
		signed int32
		terms  []string
	}{
		{0, []string{"0      : -2 * 1\n", "1      : +1 * 1 * 2\n", "2      : +2\n"}},
		{1, []string{"0      : -2\n", "1      : +1 * 1 * 2\n", "2      : +2 * 1\n"}},
	} {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdout
		os.Stdout = w
		func() {
			defer func() { os.Stdout = old; w.Close() }()
			Gia_PolynBuild(p, order, tc.signed, 1, 1)
		}()
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		for _, term := range tc.terms {
			if !strings.Contains(string(data), term) {
				t.Fatalf("signed=%d missing source monomial %q in %q", tc.signed, term, data)
			}
		}
	}
}
