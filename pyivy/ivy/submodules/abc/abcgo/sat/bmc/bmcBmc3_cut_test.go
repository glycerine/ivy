package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestBmcCutEncodingTruthTables(t *testing.T) {
	leaves := [4]int32{2, 4, 6, 8}
	const output = int32(10)
	for raw := 0; raw < 1<<16; raw++ {
		truth := uint16(raw)
		encoded, err := bmcEncodeCut(truth, leaves, output)
		if err != nil {
			t.Fatalf("truth=%04x: %v", truth, err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			want := truth&(1<<assignment) != 0
			canonicalAssignment := bmcCutTestCanonicalAssignment(encoded.leaves, assignment)
			normalized := encoded.truth&(1<<canonicalAssignment) != 0
			if got := bmcLitValue(encoded.resultLiteral, assignment, normalized); got != want {
				t.Fatalf("truth=%04x canonical=%04x assignment=%04b normalized=%t result literal=%d got %t want %t", truth, encoded.truth, assignment, normalized, encoded.resultLiteral, got, want)
			}
			if encoded.constant || encoded.resultLiteral>>1 != output>>1 {
				continue
			}
			for outputValue := 0; outputValue <= 1; outputValue++ {
				got := bmcClausesValue(encoded.clauses, assignment, outputValue != 0)
				if got != (normalized == (outputValue != 0)) {
					t.Fatalf("truth=%04x assignment=%04b output=%d normalized=%t clause result=%t", truth, assignment, outputValue, normalized, got)
				}
			}
		}
	}
}

func TestBmcCutEncodingMatchesSourceOrdering(t *testing.T) {
	// C first regularizes complemented SAT inputs, then uses Dar_CutSortVars'
	// stable adjacent swaps to order their variable IDs.
	truth := uint16(0x6996) // four-input parity
	leaves := [4]int32{9, 2, 7, 4}
	encoded, err := bmcEncodeCut(truth, leaves, 10)
	if err != nil {
		t.Fatal(err)
	}
	wantLeaves := [4]int32{2, 4, 6, 8}
	if encoded.leaves != wantLeaves {
		t.Fatalf("canonical leaves=%v want %v", encoded.leaves, wantLeaves)
	}
	for assignment := 0; assignment < 16; assignment++ {
		original := bmcLitValue(2, assignment, false) != bmcLitValue(9, assignment, false)
		original = original != bmcLitValue(7, assignment, false)
		original = original != bmcLitValue(4, assignment, false)
		canonicalAssignment := bmcCutTestCanonicalAssignment(encoded.leaves, assignment)
		normalized := encoded.truth&(1<<canonicalAssignment) != 0
		if got := bmcLitValue(encoded.resultLiteral, assignment, normalized); got != original {
			t.Fatalf("assignment=%04b got=%t want=%t", assignment, got, original)
		}
	}
}

func TestBmcCutEncodingCofactorsConstantsAndWires(t *testing.T) {
	// x0 AND x1, with x0 fixed to one, reduces to the wire x1.
	and := uint16(0x8888)
	got, err := bmcEncodeCut(and, [4]int32{1, 4, -1, -1}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got.resultLiteral != 4 || got.constant || len(got.clauses) != 0 {
		t.Fatalf("fixed cofactor=%+v; want wire literal 4", got)
	}

	got, err = bmcEncodeCut(and, [4]int32{0, 4, -1, -1}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !got.constant || got.resultLiteral != 0 || len(got.clauses) != 0 {
		t.Fatalf("false cofactor=%+v; want constant false", got)
	}

	got, err = bmcEncodeCut(0xaaaa, [4]int32{9, -1, -1, -1}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got.resultLiteral != 9 || got.truth != 0x5555 || len(got.clauses) != 0 {
		t.Fatalf("complemented wire=%+v; want source literal 9", got)
	}
}

func bmcCutTestCanonicalAssignment(leaves [4]int32, assignment int) int {
	var result int
	for variable, literal := range leaves {
		if literal < 2 {
			continue
		}
		if bmcLitValue(literal, assignment, false) {
			result |= 1 << variable
		}
	}
	return result
}

func TestBmcCutEncodingUsesABCMSOPClauseOrder(t *testing.T) {
	// For x0 AND x1, C emits the positive MSOP first, with input literals in
	// descending leaf position, followed by the two cubes for the complement.
	encoded, err := bmcEncodeCut(0x8888, [4]int32{2, 4, 6, 8}, 10)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int32{{10, 5, 3}, {11, 2}, {11, 4}}
	if !sameClauses(encoded.clauses, want) {
		t.Fatalf("clauses differ from C MSOP ordering: got %v want %v", encoded.clauses, want)
	}
}

func TestBmcCutTernaryMatchesExhaustiveTwoLeafCofactors(t *testing.T) {
	states := []bmcTernary{bmcTernaryZero, bmcTernaryOne, bmcTernaryUnknown}
	for raw := 0; raw < 1<<4; raw++ {
		truth := uint16(0)
		for assignment := 0; assignment < 4; assignment++ {
			if raw&(1<<assignment) != 0 {
				for high := 0; high < 4; high++ {
					truth |= 1 << (assignment | high<<2)
				}
			}
		}
		for a, s0 := range states {
			for b, s1 := range states {
				values := [4]bmcTernary{s0, s1, bmcTernaryUnknown, bmcTernaryUnknown}
				got := bmcCutTernary(truth, values)
				want := bmcTernaryUnknown
				if s0 != bmcTernaryUnknown && s1 != bmcTernaryUnknown {
					if raw&(1<<((a&1)|((b&1)<<1))) == 0 {
						want = bmcTernaryZero
					} else {
						want = bmcTernaryOne
					}
				} else {
					canZero, canOne := false, false
					for assignment := 0; assignment < 4; assignment++ {
						if s0 != bmcTernaryUnknown && assignment&1 != int(s0) || s1 != bmcTernaryUnknown && (assignment>>1)&1 != int(s1) {
							continue
						}
						if raw&(1<<assignment) != 0 {
							canOne = true
						} else {
							canZero = true
						}
					}
					if canZero && !canOne {
						want = bmcTernaryZero
					} else if canOne && !canZero {
						want = bmcTernaryOne
					}
				}
				if got != want {
					t.Fatalf("truth=%04x states=%v got=%d want=%d", truth, values, got, want)
				}
			}
		}
	}
}

func TestBmcCutRejectsMalformedMappings(t *testing.T) {
	if _, err := bmcEncodeCut(0x8888, [4]int32{2, -2, -1, -1}, 10); err == nil {
		t.Fatal("accepted invalid negative leaf literal")
	}
	if _, err := bmcEncodeCut(0x8888, [4]int32{2, -1, -1, -1}, 10); err == nil {
		t.Fatal("accepted a truth table that depends on an absent leaf")
	}
	if _, err := bmcEncodeCut(0x8888, [4]int32{2, 4, -1, -1}, 11); err == nil {
		t.Fatal("accepted complemented output variable")
	}
}

func TestBmcCutMappingRecordsAndPerFrameHashReuse(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	root := graph.AND(a, b)
	graph.AddOutput(root)
	mapping := cnf.Cnf_DeriveMappingArray(graph)
	truth, leaves, ok, err := bmcCutMappingRecord(vec.Vec_IntArray(mapping), root.NodeID())
	if err != nil || !ok {
		t.Fatalf("mapping record found=%t err=%v", ok, err)
	}
	if truth == 0 || truth == 0xffff {
		t.Fatalf("AND cut has constant truth %04x", truth)
	}
	leafSet := map[int32]bool{}
	for _, id := range leaves {
		if id >= 0 {
			leafSet[id] = true
		}
	}
	if !leafSet[int32(a.NodeID())] || !leafSet[int32(b.NodeID())] {
		t.Fatalf("packed cut leaves=%v omit source inputs %d,%d", leaves, a.NodeID(), b.NodeID())
	}
	if _, _, found, err := bmcCutMappingRecord(vec.Vec_IntArray(mapping), a.NodeID()); err != nil || found {
		t.Fatalf("primary input mapping found=%t err=%v; want no cut record", found, err)
	}

	// Append a second object-indexed record with the same source mapping. This
	// exercises the exact five-int layout without relying on AIG strashing to
	// retain duplicate nodes.
	raw := append([]int32(nil), vec.Vec_IntArray(mapping)...)
	duplicateID := len(raw)
	raw = append(raw, 0)
	duplicateOffset := len(raw)
	raw = append(raw, int32(truth), leaves[0], leaves[1], leaves[2], leaves[3])
	raw[duplicateID] = int32(duplicateOffset)
	withDuplicate := vec.Vec_IntAllocArray(raw, len(raw))

	encoder := newBmcCutFrameEncoder(10)
	literalAt := func(frame, objectID int) (int32, bool) {
		return int32(2 * (100 + 20*frame + objectID)), true
	}
	first, mapped, err := encoder.EncodeObject(mapping, root.NodeID(), 0, literalAt)
	if err != nil || !mapped || len(first.clauses) == 0 || first.resultLiteral != 20 {
		t.Fatalf("first frame encoding=%+v mapped=%t err=%v", first, mapped, err)
	}
	second, mapped, err := encoder.EncodeObject(withDuplicate, duplicateID, 0, literalAt)
	if err != nil || !mapped || len(second.clauses) != 0 || second.resultLiteral != first.resultLiteral {
		t.Fatalf("same-frame hash result=%+v mapped=%t err=%v", second, mapped, err)
	}
	if encoder.nextVariable != 11 {
		t.Fatalf("hash hit allocated another variable: next=%d want 11", encoder.nextVariable)
	}
	third, mapped, err := encoder.EncodeObject(mapping, root.NodeID(), 1, literalAt)
	if err != nil || !mapped || len(third.clauses) == 0 || third.resultLiteral != 22 {
		t.Fatalf("next frame should use frame-local leaves and allocate: %+v mapped=%t err=%v", third, mapped, err)
	}

	bad := append([]int32(nil), raw...)
	bad[duplicateID] = int32(len(bad) - 2)
	if _, _, _, err := bmcCutMappingRecord(bad, duplicateID); err == nil {
		t.Fatal("accepted a truncated packed cut record")
	}
}

func bmcLitValue(literal int32, assignment int, output bool) bool {
	if literal == 0 {
		return false
	}
	if literal == 1 {
		return true
	}
	variable := literal >> 1
	value := output
	if variable < 5 {
		value = assignment&(1<<(variable-1)) != 0
	}
	if literal&1 != 0 {
		return !value
	}
	return value
}

func bmcClausesValue(clauses [][]int32, assignment int, output bool) bool {
	for _, clause := range clauses {
		satisfied := false
		for _, literal := range clause {
			if bmcLitValue(literal, assignment, output) {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

func sameClauses(a, b [][]int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
