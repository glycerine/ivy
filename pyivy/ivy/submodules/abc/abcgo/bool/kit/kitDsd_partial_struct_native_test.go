package kit

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

// kitTestRand is the LCG of the native drivers in test_vectors.
type kitTestRand uint64

func (s *kitTestRand) next() uint32 {
	*s = *s*6364136223846793005 + 1442695040888963407
	return uint32(*s >> 33)
}

// gen mirrors gen() in test_vectors/dsd_struct_native.c.
func (s *kitTestRand) gen(t []uint32, nVars int32, depth int) {
	nW := int(Kit_TruthWordNum(nVars))
	op := s.next() % 6
	if depth == 0 || op == 0 {
		Kit_TruthIthVar(t, nVars, int32(s.next()%uint32(nVars)))
		if s.next()&1 != 0 {
			Kit_TruthNot(t, t, nVars)
		}
		return
	}
	a, b, c := make([]uint32, 8), make([]uint32, 8), make([]uint32, 8)
	s.gen(a, nVars, depth-1)
	s.gen(b, nVars, depth-1)
	for k := 0; k < nW; k++ {
		switch {
		case op == 1 || op == 2:
			t[k] = a[k] & b[k]
		case op == 3:
			t[k] = a[k] ^ b[k]
		default:
			if k == 0 {
				s.gen(c, nVars, depth-1)
			}
			if op == 4 {
				t[k] = a[k]&c[k] | b[k]&^c[k]
			} else {
				t[k] = a[k]&b[k] | a[k]&c[k] | b[k]&c[k]
			}
		}
	}
	if s.next()&1 != 0 {
		Kit_TruthNot(t, t, nVars)
	}
}

// test_vectors/dsd_partial_struct_native.txt holds the first 5000 lines
// printed by dsd_partial_struct_native.c against native ABC: decomposable
// functions with bound sets made of whole DSD components, as lutpack uses,
// and the results of Kit_DsdTruthPartialTwo and Kit_DsdTruthPartial.
func TestDsdTruthPartialStructuredMatchesNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/dsd_partial_struct_native.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := bufio.NewScanner(strings.NewReader(string(data)))
	rng := kitTestRand(12345)
	m := Kit_DsdManAlloc(8, 64)
	tt, co, dec, part := make([]uint32, 8), make([]uint32, 8), make([]uint32, 8), make([]uint32, 8)
	lines := 0
	for it := 0; lines < 5000; it++ {
		nVars := 4 + int32(rng.next()%5)
		nW := int(Kit_TruthWordNum(nVars))
		rng.gen(tt, nVars, 2+int(rng.next()%4))
		ntk := Kit_DsdDecomposeExpand(tt, nVars)
		Kit_DsdGetSupports(ntk)
		bound := uint32(0)
		if len(ntk.pNodes) > 0 {
			o := ntk.pNodes[rng.next()%uint32(len(ntk.pNodes))]
			pick := rng.next()
			if o.Type == KIT_DSD_PRIME {
				bound = Kit_DsdLitSupport(ntk, int32(o.pFans[pick%uint32(o.nFans)]))
			} else if o.Type == KIT_DSD_AND || o.Type == KIT_DSD_XOR {
				for i := uint32(0); i < uint32(o.nFans); i++ {
					if pick>>i&1 != 0 {
						bound |= Kit_DsdLitSupport(ntk, int32(o.pFans[i]))
					}
				}
			}
		}
		if Kit_WordCountOnes(bound) < 2 {
			continue
		}
		Kit_DsdTruthPartialTwo(m, ntk, bound, Kit_WordFindFirstBit(bound), co, dec)
		Kit_DsdTruthPartial(m, Kit_DsdDecompose(tt, nVars), part, bound)
		var b strings.Builder
		fmt.Fprintf(&b, "%d %d %x", it, nVars, bound)
		for j, x := range [][]uint32{tt, co, dec, part} {
			if j > 0 {
				b.WriteString(" |")
			}
			for w := 0; w < nW; w++ {
				fmt.Fprintf(&b, " %08x", x[w])
			}
		}
		if !want.Scan() {
			t.Fatal("native vectors ended early")
		}
		if got := b.String(); got != want.Text() {
			t.Fatalf("got  %s\nwant %s", got, want.Text())
		}
		lines++
	}
}
