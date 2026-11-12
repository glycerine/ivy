package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ManDetectConstr preserves the original C routine's three-way status:
// zero means no structural match, -1 means the output is constant false, and
// a positive value is the number of detected constraint factors. The returned
// edge slices retain the original complemented-supergate representation.
func Saig_ManDetectConstr(g *aig.Graph, output int32) (propertyNegations, constraintNegations []aig.Edge, status int32) {
	split, ok, err := DetectStructuralConstraint(g, int(output))
	if err != nil {
		panic(fmt.Errorf("saig: detect constraints: %w", err))
	}
	if !ok {
		return nil, nil, 0
	}
	if split.ConstantFalse {
		return split.PropertyNegations, []aig.Edge{}, -1
	}
	return split.PropertyNegations, split.ConstraintNegations, int32(len(split.ConstraintNegations))
}

// Saig_ManDupUnfoldConstrs duplicates a sequential AIG while exposing
// structurally detected constraints as trailing outputs. When the source
// algorithm cannot safely split the outputs, it returns a behavior-preserving
// copy instead.
func Saig_ManDupUnfoldConstrs(g *aig.Graph) *aig.Graph {
	copy, _, err := UnfoldStructuralConstraints(g)
	if err != nil {
		panic(fmt.Errorf("saig: unfold constraints: %w", err))
	}
	return copy
}

// Saig_ManDupFoldConstrs adds the original algorithm's sticky constraint
// register and gates outputs with its healthy-state value. Constraint indices
// use ordinary Go slice indexing and preserve the C routine's rule that output
// zero is reserved for the property.
func Saig_ManDupFoldConstrs(g *aig.Graph, constraintOutputs []int) *aig.Graph {
	copy, err := FoldStructuralConstraints(g, constraintOutputs)
	if err != nil {
		panic(fmt.Errorf("saig: fold constraints: %w", err))
	}
	return copy
}
