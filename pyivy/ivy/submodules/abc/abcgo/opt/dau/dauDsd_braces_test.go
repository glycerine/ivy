package dau

import (
	"math/rand"
	"testing"
)

// dsRandomStructured returns a random function of nVars built from ANDs,
// XORs and MUXes of literals, so that it has nontrivial DSD structure.
func dsRandomStructured(r *rand.Rand, nVars int, depth int) []uint64 {
	words := 1
	if nVars > 6 {
		words = 1 << (nVars - 6)
	}
	out := make([]uint64, words)
	if depth == 0 || r.Intn(4) == 0 {
		v := r.Intn(nVars)
		for w := range out {
			for b := 0; b < 64; b++ {
				if ((w*64+b)>>v)&1 != 0 {
					out[w] |= 1 << b
				}
			}
			if r.Intn(2) == 0 {
				out[w] = ^out[w]
			}
		}
		return out
	}
	a, b := dsRandomStructured(r, nVars, depth-1), dsRandomStructured(r, nVars, depth-1)
	op := r.Intn(3)
	var c []uint64
	if op == 2 {
		c = dsRandomStructured(r, nVars, depth-1)
	}
	for w := range out {
		switch op {
		case 0:
			out[w] = a[w] & b[w]
		case 1:
			out[w] = a[w] ^ b[w]
		default:
			out[w] = a[w]&c[w] | b[w]&^c[w]
		}
	}
	return out
}

// The C-style brace removal must agree with flattening the parsed tree.
func TestRemoveBracesMatchesTreeFlattening(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	changed := 0
	defer func() { t.Logf("%d cases had braces removed", changed) }()
	for it := 0; it < 5000; it++ {
		nVars := 3 + r.Intn(6)
		truth := dsRandomStructured(r, nVars, 2+r.Intn(4))
		p := &Dau_Dsd_t{fWriteTruth: 1}
		dsDecomposeInternal(p, truth, int32(nVars))
		raw := p.output.String()
		if len(raw) <= 1 {
			continue
		}
		want := dsFlatten(dsParse(raw)).text()
		if raw != want {
			changed++
		}
		if got := string(dsRemoveBraces([]byte(raw))); got != want {
			t.Fatalf("%s: got %s want %s", raw, got, want)
		}
	}
}
