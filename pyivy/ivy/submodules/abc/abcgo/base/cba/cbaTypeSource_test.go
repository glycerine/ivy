package cba

import (
	"bufio"
	_ "embed"
	"fmt"
	"strings"
	"testing"
)

//go:embed test_vectors/cba_type_predicates.txt
var sourceTypePredicateResults string

func TestAllTypePredicatesAgainstCompiledC(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(sourceTypePredicateResults))
	rows := 0
	for scanner.Scan() {
		var id, box, seq, unary, mux int32
		if _, err := fmt.Sscanf(scanner.Text(), "%d %d %d %d %d", &id, &box, &seq, &unary, &mux); err != nil {
			t.Fatal(err)
		}
		kind := Cba_ObjType_t(id)
		got := [4]int32{Cba_TypeIsBox(kind), Cba_TypeIsSeq(kind), Cba_TypeIsUnary(kind), Cba_TypeIsMux(kind)}
		want := [4]int32{box, seq, unary, mux}
		if got != want {
			t.Fatalf("type %d Go %v C %v", id, got, want)
		}
		rows++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if rows != int(CBA_BOX_LAST)+2 {
		t.Fatalf("incomplete C fixture: %d rows", rows)
	}
}
