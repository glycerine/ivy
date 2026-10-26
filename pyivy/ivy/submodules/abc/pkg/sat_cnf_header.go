package abc

// Public CNF header helpers from src/sat/cnf/cnf.h.

import "slices"

// CNFHeaderOptions bounds the owned replacements for the pointer iterators and
// object scratch slots declared by cnf.h. Zero limits select conservative
// package defaults.
type CNFHeaderOptions struct {
	MaxObjects, MaxClauses, MaxLiterals int
	Stop                                func() bool
}

func checkCNFHeaderOptions(o *CNFHeaderOptions) error {
	if o.MaxObjects < 0 || o.MaxClauses < 0 || o.MaxLiterals < 0 {
		return ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 1 << 20
	}
	if o.MaxLiterals == 0 {
		o.MaxLiterals = 4 << 20
	}
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	return nil
}

// CNFCutSOPCost4 is Cnf_CutSopCost over the immutable four-input SOP table.
// It counts cubes for both polarities, exactly as the inline C helper does.
func CNFCutSOPCost4(cut AigCNFCut4) (int, error) {
	positive, err := CNFSOP4Cover(cut.Truth)
	if err != nil {
		return 0, err
	}
	negative, err := CNFSOP4Cover(^cut.Truth)
	if err != nil {
		return 0, err
	}
	return len(positive) + len(negative), nil
}

// CNFCutLeafCount, CNFCutLeaves and CNFCutTruth replace the three raw-pointer
// cut accessors. Returned data is owned and cannot mutate the source cut.
func CNFCutLeafCount(cut AigCNFCut) int { return len(cut.Leaves) }

func CNFCutLeaves(cut AigCNFCut) []AigID { return slices.Clone(cut.Leaves) }

func CNFCutTruth(cut AigCNFCut) *TruthTable { return cut.Truth.Clone() }

// BestCut4 replaces Dar_ObjBestCut after area-flow selection. Selection already
// resolves fBest to one cut per AND; lookup never depends on map iteration.
func (mapping *CNFAreaFlowMapping) BestCut4(root AigID) (AigCNFCut4, bool) {
	if mapping == nil {
		return AigCNFCut4{}, false
	}
	for _, cut := range mapping.Cuts {
		if cut.Root == root {
			cut.Leaves = slices.Clone(cut.Leaves)
			return cut, true
		}
	}
	return AigCNFCut4{}, false
}

// CNFBestCutTable is an owned replacement for storing a Cnf_Cut_t pointer in
// Aig_Obj_t.pData. It is deliberately separate from AigManager so CNF mapping
// cannot mutate graph scratch state.
type CNFBestCutTable struct {
	cuts    []AigCNFCut
	present []bool
	stop    func() bool
}

func NewCNFBestCutTable(objects int, o CNFHeaderOptions) (*CNFBestCutTable, error) {
	if objects < 0 {
		return nil, ErrDimension
	}
	if err := checkCNFHeaderOptions(&o); err != nil {
		return nil, err
	}
	if objects > o.MaxObjects {
		return nil, ErrLimit
	}
	return &CNFBestCutTable{cuts: make([]AigCNFCut, objects), present: make([]bool, objects), stop: o.Stop}, nil
}

func cloneCNFHeaderCut(cut AigCNFCut) (AigCNFCut, error) {
	if cut.Truth == nil || cut.Truth.Variables() != len(cut.Leaves) || len(cut.Leaves) > 16 {
		return AigCNFCut{}, ErrDimension
	}
	if err := cut.Truth.Check(); err != nil {
		return AigCNFCut{}, err
	}
	for i, leaf := range cut.Leaves {
		if leaf >= cut.Root || i > 0 && leaf <= cut.Leaves[i-1] {
			return AigCNFCut{}, ErrDimension
		}
	}
	cut.Leaves = slices.Clone(cut.Leaves)
	cut.Truth = cut.Truth.Clone()
	return cut, nil
}

// SetBestCut implements Cnf_ObjSetBestCut. A nil cut clears the slot. Failed
// validation and cancellation leave the prior slot unchanged.
func (table *CNFBestCutTable) SetBestCut(root AigID, cut *AigCNFCut) error {
	if table == nil || uint64(root) >= uint64(len(table.cuts)) {
		return ErrDimension
	}
	if table.stop != nil && table.stop() {
		return ErrCancelled
	}
	if cut == nil {
		table.cuts[root] = AigCNFCut{}
		table.present[root] = false
		return nil
	}
	if cut.Root != root {
		return ErrDimension
	}
	owned, err := cloneCNFHeaderCut(*cut)
	if err != nil {
		return err
	}
	if table.stop != nil && table.stop() {
		return ErrCancelled
	}
	table.cuts[root] = owned
	table.present[root] = true
	return nil
}

// BestCut implements Cnf_ObjBestCut and returns a fully owned snapshot.
func (table *CNFBestCutTable) BestCut(root AigID) (AigCNFCut, bool, error) {
	if table == nil || uint64(root) >= uint64(len(table.cuts)) {
		return AigCNFCut{}, false, ErrDimension
	}
	if table.stop != nil && table.stop() {
		return AigCNFCut{}, false, ErrCancelled
	}
	if !table.present[root] {
		return AigCNFCut{}, false, nil
	}
	cut, err := cloneCNFHeaderCut(table.cuts[root])
	return cut, err == nil, err
}

// CNFClauseRange is the index-safe replacement for the begin/end pointers
// yielded by Cnf_CnfForClause. End is exclusive in the flattened stream.
type CNFClauseRange struct {
	Begin, End int
}

type CNFClauseStream struct {
	Literals []SatLit
	Ranges   []CNFClauseRange
}

// BuildCNFClauseStream owns and flattens clauses while retaining every empty
// clause and native clause boundary. Input storage is never retained.
func BuildCNFClauseStream(clauses [][]SatLit, o CNFHeaderOptions) (*CNFClauseStream, error) {
	if err := checkCNFHeaderOptions(&o); err != nil {
		return nil, err
	}
	if len(clauses) > o.MaxClauses {
		return nil, ErrLimit
	}
	result := &CNFClauseStream{Ranges: make([]CNFClauseRange, 0, len(clauses))}
	for _, clause := range clauses {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		if len(clause) > o.MaxLiterals-len(result.Literals) {
			return nil, ErrLimit
		}
		begin := len(result.Literals)
		result.Literals = append(result.Literals, clause...)
		result.Ranges = append(result.Ranges, CNFClauseRange{Begin: begin, End: len(result.Literals)})
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

// CNFCutLeafObjects replaces Cnf_CutForEachLeaf with checked owned object
// snapshots in cut order.
func CNFCutLeafObjects(graph *AigManager, cut AigCNFCut, o CNFHeaderOptions) ([]AigObject, error) {
	if graph == nil || cut.Truth == nil || cut.Truth.Variables() != len(cut.Leaves) {
		return nil, ErrDimension
	}
	if err := checkCNFHeaderOptions(&o); err != nil {
		return nil, err
	}
	if len(cut.Leaves) > o.MaxObjects {
		return nil, ErrLimit
	}
	objects := make([]AigObject, 0, len(cut.Leaves))
	for _, leaf := range cut.Leaves {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		object, err := graph.Object(leaf)
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return objects, nil
}
