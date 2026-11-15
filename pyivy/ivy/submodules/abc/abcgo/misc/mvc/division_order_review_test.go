package mvc

import "testing"

// Source division sorts both operands before grouping. Factoring reuses the
// dividend, so preserving only the quotient's Boolean value is insufficient.
func TestDivisionSortsReusedOperands(t *testing.T) {
	c, d := Mvc_CoverAlloc(nil, 6), Mvc_CoverAlloc(nil, 6)
	addBits(c, 0, 3, 4) // 25
	addBits(c, 1, 2, 4) // 22
	addBits(d, 0, 3)    // 9
	addBits(d, 1, 2)    // 6
	q, r := Mvc_CoverDivideInternal(c, d)
	if c.Cubes.Head.Data[0] != 22 || d.Cubes.Head.Data[0] != 6 {
		t.Fatalf("source operand order lost: dividend %v, divisor %v", c.Cubes.Head.Data, d.Cubes.Head.Data)
	}
	if q.Cubes.Items != 1 || q.Cubes.Head.Data[0] != 16 || r.Cubes.Items != 0 {
		t.Fatalf("unexpected division: quotient %v, remainder %v", terms(q), terms(r))
	}
}
