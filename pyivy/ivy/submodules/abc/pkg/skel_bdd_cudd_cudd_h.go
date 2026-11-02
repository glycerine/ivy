package abc

// Hand-written skeleton of the exported types in src/bdd/cudd/cudd.h that
// have no literal Go counterpart (the Go CUDD port uses edge handles).

// DdChildren is the DdChildren struct.
type DdChildren struct {
	T *DdNode
	E *DdNode
}

// DdNode is struct DdNode. The C `type` member is a union of `value`
// (constants) and `kids` (internal nodes); Go has no unions, so both members
// are present and only one is meaningful for a given node.
type DdNode struct {
	index uint32  // DdHalfWord
	ref   uint32  // DdHalfWord : reference count
	next  *DdNode // next pointer for unique table
	// union { CUDD_VALUE_TYPE value; DdChildren kids; } type;
	typeValue float64    // type.value : for constant nodes (CUDD_VALUE_TYPE is double)
	typeKids  DdChildren // type.kids : for internal nodes
	Id        int64      // ABC_INT64_T
}
