package startingpoint

import (
	"strings"
	"testing"
)

func TestLowerFreeAndNull(t *testing.T) {
	got, err := lowerFreeAndNull(`((pGia.pRefs) ? (free((char *) (pGia.pRefs)), (pGia.pRefs) = 0) : 0);`)
	if err != nil {
		t.Fatal(err)
	}
	if got != `pGia.pRefs = nil;` {
		t.Fatalf("got %q", got)
	}
}

func TestLowerFreeAndNullIndexedAndFormatting(t *testing.T) {
	got, err := lowerFreeAndNull(`((p.chunks[i]) ? (free((void *)(p.chunks[i])), (p.chunks[i]) = NULL) : 0);`)
	if err != nil {
		t.Fatal(err)
	}
	if got != `p.chunks[i] = nil;` {
		t.Fatalf("got %q", got)
	}
}

func TestLowerFreeAndNullLeavesNonMatchingCodeAlone(t *testing.T) {
	source := `((p) ? (free((char *)(q)), (p) = 0) : 0); // ((x) ? (free((char *)(x)), (x) = 0) : 0)`
	got, err := lowerFreeAndNull(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != source || strings.Contains(got, `p = nil`) {
		t.Fatalf("unexpected lowering: %q", got)
	}
}

func TestMechanicalDraftLowersFreeAndNull(t *testing.T) {
	got, err := mechanicalDraft(`void cleanup() { // inlining of macro ABC_FREE
    ((pGia.pRefs) ? (free((char *) (pGia.pRefs)), (pGia.pRefs) = 0) : 0);
}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "pGia.pRefs = nil;") || strings.Contains(got, "free((char *)") {
		t.Fatalf("free-and-null expansion was not lowered: %q", got)
	}
}
