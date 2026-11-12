package aig

import "testing"

func TestStructuralTableOperations(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	ghost := Node{Kind: And, Fanins: [2]Edge{a, b}}
	if Aig_TableLookup(p, &ghost) != InvalidEdge {
		t.Fatal("unexpected lookup")
	}
	e := p.AND(a, b)
	if Aig_TableLookup(p, &ghost) != e || Aig_TableCountEntries(p) != 1 {
		t.Fatal("missing entry")
	}
	Aig_TableResize(p)
	if Aig_TableLookupTwo(p, b, a) != e {
		t.Fatal("resize lost entry")
	}
	Aig_TableDelete(p, e)
	if Aig_TableLookup(p, &ghost) != InvalidEdge {
		t.Fatal("delete failed")
	}
	Aig_TableInsert(p, e)
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	if Aig_Hash(&ghost, 17) >= 17 {
		t.Fatal("hash outside table")
	}
	Aig_TableClear(p)
	if Aig_TableCountEntries(p) != 0 {
		t.Fatal("clear failed")
	}
}
