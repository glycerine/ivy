package bmc

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

var bmcCutMSOPOnce sync.Once
var bmcCutMSOPTable [][]byte

// bmcTernary is the value domain used by Saig_ManBmcRunTerSim_rec.
type bmcTernary uint8

const (
	bmcTernaryZero bmcTernary = iota
	bmcTernaryOne
	bmcTernaryUnknown
)

// bmcCutEncoding is the result of reducing and encoding one source-style
// four-leaf cut. Clauses define the normalized (positive-phase) cut function
// in outputLiteral. resultLiteral is the cut's actual function and may be a
// constant, an input literal, or outputLiteral with its polarity adjusted.
// truth and leaves are the canonical hash key used by Saig_ManBmcCreateCnf_rec.
type bmcCutEncoding struct {
	truth         uint16
	leaves        [4]int32
	resultLiteral int32
	constant      bool
	clauses       [][]int32
}

// bmcCutFrameEncoder mirrors the truth/leaf hash in Saig_Bmc3ManStart. The
// packed mapping stores object IDs; EncodeObject resolves those IDs to the
// SAT literals belonging to the requested frame before reducing or hashing.
type bmcCutFrameEncoder struct {
	nextVariable int32
	hash         map[[5]int32]int32
}

func newBmcCutFrameEncoder(firstVariable int32) *bmcCutFrameEncoder {
	if firstVariable < 1 {
		firstVariable = 1
	}
	return &bmcCutFrameEncoder{nextVariable: firstVariable, hash: make(map[[5]int32]int32)}
}

// bmcCutMappingRecord decodes Cnf_DeriveMappingArray's object-indexed offsets.
// A nonzero entry points to five integers: 16-bit truth, then four leaf IDs.
func bmcCutMappingRecord(mapping []int32, objectID int) (uint16, [4]int32, bool, error) {
	var leaves [4]int32
	for i := range leaves {
		leaves[i] = -1
	}
	if objectID < 0 || objectID >= len(mapping) {
		return 0, leaves, false, fmt.Errorf("bmc: object ID %d outside cut mapping", objectID)
	}
	offset := mapping[objectID]
	if offset == 0 {
		return 0, leaves, false, nil
	}
	if offset < 0 || int64(offset)+5 > int64(len(mapping)) {
		return 0, leaves, false, fmt.Errorf("bmc: malformed cut mapping offset %d", offset)
	}
	truth := uint16(mapping[offset])
	copy(leaves[:], mapping[offset+1:offset+5])
	return truth, leaves, true, nil
}

// EncodeObject decodes one mapping record, resolves its leaves with
// literalAt(frame, objectID), and uses the normalized truth/leaves as the
// source hash key. It returns mapped=false only when the object has no cut
// record; a mapped constant, wire, or hash hit has no new clauses. literalAt
// must return a nonnegative ABC literal.
func (e *bmcCutFrameEncoder) EncodeObject(mapping *vec.Vec_Int_t, objectID, frame int, literalAt func(frame, objectID int) (int32, bool)) (bmcCutEncoding, bool, error) {
	if e == nil || mapping == nil || literalAt == nil {
		return bmcCutEncoding{}, false, fmt.Errorf("bmc: nil cut encoder input")
	}
	truth, objectLeaves, exists, err := bmcCutMappingRecord(vec.Vec_IntArray(mapping), objectID)
	if err != nil {
		return bmcCutEncoding{}, false, err
	}
	if !exists {
		return bmcCutEncoding{}, false, nil
	}
	literals := objectLeaves
	for i, leafID := range objectLeaves {
		if leafID < 0 {
			continue
		}
		literal, ok := literalAt(frame, int(leafID))
		if !ok || literal < 0 {
			return bmcCutEncoding{}, false, fmt.Errorf("bmc: no SAT literal for frame %d leaf object %d", frame, leafID)
		}
		literals[i] = literal
	}
	output := e.nextVariable << 1
	encoded, err := bmcEncodeCut(truth, literals, output)
	if err != nil {
		return bmcCutEncoding{}, false, err
	}
	if encoded.constant || len(encoded.clauses) == 0 {
		return encoded, true, nil
	}
	key := [5]int32{int32(encoded.truth), encoded.leaves[0], encoded.leaves[1], encoded.leaves[2], encoded.leaves[3]}
	if previous, found := e.hash[key]; found {
		if encoded.resultLiteral == output^1 {
			encoded.resultLiteral = previous ^ 1
		} else {
			encoded.resultLiteral = previous
		}
		encoded.clauses = nil
		return encoded, true, nil
	}
	e.hash[key] = output
	e.nextVariable++
	return encoded, true, nil
}

// bmcCutTernary evaluates a four-input truth table under zero, one, or
// unknown leaf values. It returns a known value exactly when all compatible
// assignments have the same result, matching C's cofactor reduction.
func bmcCutTernary(truth uint16, values [4]bmcTernary) bmcTernary {
	var result bmcTernary
	haveResult := false
	for assignment := 0; assignment < 16; assignment++ {
		compatible := true
		for variable, value := range values {
			switch value {
			case bmcTernaryZero:
				compatible = compatible && assignment&(1<<variable) == 0
			case bmcTernaryOne:
				compatible = compatible && assignment&(1<<variable) != 0
			case bmcTernaryUnknown:
			default:
				return bmcTernaryUnknown
			}
		}
		if !compatible {
			continue
		}
		value := bmcTernaryZero
		if truth&(1<<assignment) != 0 {
			value = bmcTernaryOne
		}
		if !haveResult {
			result, haveResult = value, true
		} else if result != value {
			return bmcTernaryUnknown
		}
	}
	if !haveResult {
		return bmcTernaryUnknown
	}
	return result
}

// bmcEncodeCut follows Saig_ManBmcReduceTruth, Dar_CutSortVars,
// Saig_ManBmcAddClauses, and the constant/wire cases in
// Saig_ManBmcCreateCnf_rec. Leaves are ABC SAT literals; 0 and 1 are Boolean
// constants, -1 means an unused cut input, and ordinary literals are >= 2.
// outputLiteral is a positive SAT literal reserved for the normalized cut.
func bmcEncodeCut(truth uint16, leaves [4]int32, outputLiteral int32) (bmcCutEncoding, error) {
	result := bmcCutEncoding{leaves: leaves}
	for _, literal := range leaves {
		if literal < -1 {
			return result, fmt.Errorf("bmc: invalid cut literal %d", literal)
		}
	}
	if outputLiteral < 2 || outputLiteral&1 != 0 {
		return result, fmt.Errorf("bmc: output literal must be a positive variable, got %d", outputLiteral)
	}

	for variable, literal := range result.leaves {
		switch literal {
		case 0:
			truth = bmcCutCofactor(truth, variable, false)
			result.leaves[variable] = -1
		case 1:
			truth = bmcCutCofactor(truth, variable, true)
			result.leaves[variable] = -1
		}
	}
	for variable, literal := range result.leaves {
		if literal == -1 {
			if !bmcCutCofactorsEqual(truth, variable) {
				return result, fmt.Errorf("bmc: truth table depends on absent cut input %d", variable)
			}
			continue
		}
		if bmcCutCofactorsEqual(truth, variable) {
			result.leaves[variable] = -1
		}
	}
	if truth == 0 || truth == 0xffff {
		result.truth = truth
		result.constant = true
		result.resultLiteral = int32(truth >> 15)
		return result, nil
	}

	// Dar_CutSortVars first makes every present input positive, changing the
	// truth table phase, then bubble-sorts input IDs with adjacent truth swaps.
	for variable, literal := range result.leaves {
		if literal >= 2 && literal&1 != 0 {
			result.leaves[variable] = literal ^ 1
			truth = bmcCutSwapPolarity(truth, variable)
		}
	}
	for changed := true; changed; {
		changed = false
		for variable := 0; variable < 3; variable++ {
			left, right := result.leaves[variable], result.leaves[variable+1]
			if cutVariableID(left) <= cutVariableID(right) {
				continue
			}
			result.leaves[variable], result.leaves[variable+1] = right, left
			truth = bmcCutSwapAdjacent(truth, variable)
			changed = true
		}
	}

	if truth == 0xaaaa || truth == 0x5555 {
		result.truth = truth
		result.resultLiteral = result.leaves[0]
		if truth == 0x5555 {
			result.resultLiteral ^= 1
		}
		return result, nil
	}

	complemented := truth&1 != 0
	normalized := truth
	if complemented {
		normalized = ^truth
	}
	result.truth = normalized
	result.resultLiteral = outputLiteral
	if complemented {
		result.resultLiteral ^= 1
	}
	result.clauses = bmcCutMSOPClauses(normalized, result.leaves, outputLiteral)
	return result, nil
}

func cutVariableID(literal int32) int32 {
	if literal == -1 {
		return 0x3fffffff
	}
	return literal >> 1
}

func bmcCutCofactor(truth uint16, variable int, value bool) uint16 {
	var result uint16
	for assignment := 0; assignment < 16; assignment++ {
		source := assignment &^ (1 << variable)
		if value {
			source |= 1 << variable
		}
		if truth&(1<<source) != 0 {
			result |= 1 << assignment
		}
	}
	return result
}

func bmcCutCofactorsEqual(truth uint16, variable int) bool {
	return bmcCutCofactor(truth, variable, false) == bmcCutCofactor(truth, variable, true)
}

func bmcCutSwapPolarity(truth uint16, variable int) uint16 {
	var result uint16
	for assignment := 0; assignment < 16; assignment++ {
		source := assignment ^ (1 << variable)
		if truth&(1<<source) != 0 {
			result |= 1 << assignment
		}
	}
	return result
}

func bmcCutSwapAdjacent(truth uint16, variable int) uint16 {
	var result uint16
	for assignment := 0; assignment < 16; assignment++ {
		source := assignment
		left, right := assignment&(1<<variable) != 0, assignment&(1<<(variable+1)) != 0
		if left != right {
			source ^= (1 << variable) | (1 << (variable + 1))
		}
		if truth&(1<<source) != 0 {
			result |= 1 << assignment
		}
	}
	return result
}

func bmcCutMSOPClauses(truth uint16, leaves [4]int32, outputLiteral int32) [][]int32 {
	bmcCutMSOPOnce.Do(func() { bmcCutMSOPTable = cnf.Cnf_ReadMsops() })
	covers := bmcCutMSOPTable
	clauses := make([][]int32, 0, len(covers[truth])+len(covers[^truth]))
	for polarity := 0; polarity < 2; polarity++ {
		coverTruth := truth
		if polarity != 0 {
			coverTruth = ^coverTruth
		}
		for _, encodedCube := range covers[coverTruth] {
			clause := []int32{outputLiteral}
			if polarity != 0 {
				clause[0] ^= 1
			}
			cube := int(encodedCube)
			for variable := 3; variable >= 0; variable-- {
				digit := cube % 3
				if digit == 0 {
					clause = append(clause, leaves[variable])
				} else if digit == 1 {
					clause = append(clause, leaves[variable]^1)
				}
				cube /= 3
			}
			clauses = append(clauses, clause)
		}
	}
	return clauses
}
