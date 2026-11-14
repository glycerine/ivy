package wlc

import "testing"

func TestSmtParserInlineCommentAdjacentToToken(t *testing.T) {
	input := "(assert true;comment with ) ignored\n)"
	if got := Smt_PrsRemoveComments(input, len(input)); got != 1 {
		t.Fatalf("parenthesis count = %d, want 1", got)
	}
	// The C parser pre-counts opening parentheses and reserves exactly one
	// Vec_Wec level for the top-level list plus one per opening parenthesis.
	// Keep that exact count here so ReadLines' object-capacity invariant is
	// exercised through the same lifecycle as Wlc_ReadSmtBuffer.
	nObjs := Smt_PrsRemoveComments(input, len(input))
	p := Smt_PrsAlloc("inline.smt2", input, len(input), nObjs)
	if p == nil {
		t.Fatal("parser allocation failed")
	}
	defer Smt_PrsFree(p)
	Smt_PrsReadLines(p)
	root := Vec_WecEntry(&p.vObjs, 0)
	if got := Vec_IntSize(root); got != 1 {
		t.Fatalf("top-level forms = %d, want 1", got)
	}
	assertion := Vec_WecEntry(&p.vObjs, Abc_Lit2Var(Vec_IntEntry(root, 0)))
	if got := Vec_IntSize(assertion); got != 2 {
		t.Fatalf("assertion tokens = %d, want 2", got)
	}
	for i, want := range []string{"assert", "true"} {
		if got := Smt_EntryName(p, Vec_IntEntry(assertion, int32(i))); got != want {
			t.Errorf("token %d = %q, want %q", i, got, want)
		}
	}
}

func TestSmtReadHexWordsPreservesWordOrderAndLeadingZeros(t *testing.T) {
	words := make([]int32, 2)
	if got := smtReadHexWords(words, "0123456789abcdef"); got != 16 {
		t.Fatalf("hex digit count = %d, want 16", got)
	}
	if uint32(words[0]) != 0x89abcdef || uint32(words[1]) != 0x01234567 {
		t.Fatalf("little-endian words = %#08x %#08x", uint32(words[0]), uint32(words[1]))
	}
	if got := smtReadHexWords(words[:1], "0000"); got != 4 || words[0] != 0 {
		t.Fatalf("leading-zero parse = (%d, %#x), want (4, 0)", got, uint32(words[0]))
	}
}
