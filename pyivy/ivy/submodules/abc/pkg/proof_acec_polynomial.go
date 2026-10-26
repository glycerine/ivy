package abc

// Port of src/proof/acec/acec.h, acecInt.h, acecPolyn.c, acecNorm.c,
// acecOrder.c and acecUtil.c: arithmetic polynomial and adder-tree foundations.

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"slices"
	"sort"
)

// AcecCECParameters is the concrete parameter record declared by acec.h.
// Solving orchestration belongs to the remaining ACEC source units.
type AcecCECParameters struct {
	ConflictLimit int
	TimeLimit     int
	Miter         bool
	DualOutput    bool
	TwoOutput     bool
	Booth         bool
	Silent        bool
	VeryVerbose   bool
	Verbose       bool
	FailedOutput  int
}

// AcecBox is the owned counterpart of Acec_Box_t. Ranks and literals are
// copied on construction by NewAcecBox, never borrowed from graph scratch.
type AcecBox struct {
	Graph                 *GiaManager
	Adders, Leaves, Roots [][]GiaLit
	Shared, Unique        [][]GiaLit
}

func cloneGiaRanks(ranks [][]GiaLit) [][]GiaLit {
	copyRanks := make([][]GiaLit, len(ranks))
	for i := range ranks {
		copyRanks[i] = slices.Clone(ranks[i])
	}
	return copyRanks
}

func NewAcecBox(graph *GiaManager, adders, leaves, roots, shared, unique [][]GiaLit) (*AcecBox, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	for _, ranks := range [][][]GiaLit{adders, leaves, roots, shared, unique} {
		for _, rank := range ranks {
			for _, lit := range rank {
				if !graph.validSignal(lit) {
					return nil, ErrInvalidLiteral
				}
			}
		}
	}
	return &AcecBox{graph, cloneGiaRanks(adders), cloneGiaRanks(leaves), cloneGiaRanks(roots), cloneGiaRanks(shared), cloneGiaRanks(unique)}, nil
}

// AcecSignBit and AcecSignBit2 preserve the two 16-bit sign fields packed in
// each six-integer adder record by acecInt.h.
func acecSignWordIndex(adders []int, box int) (int, error) {
	if box < 0 || len(adders) < 6 || box > (len(adders)-6)/6 {
		return 0, ErrDimension
	}
	return 6*box + 5, nil
}

func AcecSignBit(adders []int, box, bit int, second bool) (bool, error) {
	index, err := acecSignWordIndex(adders, box)
	if err != nil || bit < 0 || bit >= 16 {
		return false, ErrDimension
	}
	shift := bit
	if second {
		shift += 16
	}
	return uint32(adders[index])&(uint32(1)<<shift) != 0, nil
}

func AcecSetSignBit(adders []int, box, bit int, second, value bool) error {
	index, err := acecSignWordIndex(adders, box)
	if err != nil || bit < 0 || bit >= 16 {
		return ErrDimension
	}
	if !value {
		return nil // The source setters only set bits; they never clear them.
	}
	shift := bit
	if second {
		shift += 16
	}
	adders[index] = int(uint32(adders[index]) | uint32(1)<<shift)
	return nil
}

type AcecPolynomialOptions struct {
	MaxObjects, MaxTerms, MaxDegree int
	Stop                            func() bool
}

type AcecPolynomialTerm struct {
	Coefficient *big.Int
	Variables   []GiaID
}

// AcecPolynomial is a canonical multilinear integer polynomial. Boolean
// variables are GIA object IDs, so multiplication applies x*x == x.
type AcecPolynomial struct{ terms []AcecPolynomialTerm }

func (p AcecPolynomial) Terms() []AcecPolynomialTerm {
	terms := make([]AcecPolynomialTerm, len(p.terms))
	for i := range p.terms {
		terms[i] = AcecPolynomialTerm{new(big.Int).Set(p.terms[i].Coefficient), slices.Clone(p.terms[i].Variables)}
	}
	return terms
}

func (p AcecPolynomial) IsZero() bool { return len(p.terms) == 0 }

func (p AcecPolynomial) Evaluate(objectValues []bool) (*big.Int, error) {
	result := new(big.Int)
	for _, term := range p.terms {
		present := true
		for _, id := range term.Variables {
			if uint64(id) >= uint64(len(objectValues)) {
				return nil, ErrDimension
			}
			if !objectValues[id] {
				present = false
				break
			}
		}
		if present {
			result.Add(result, term.Coefficient)
		}
	}
	return result, nil
}

func acecMonomialKey(variables []GiaID) string {
	bytes := make([]byte, 4*len(variables))
	for i, id := range variables {
		binary.LittleEndian.PutUint32(bytes[4*i:], uint32(id))
	}
	return string(bytes)
}

func acecMergeVariables(a, b []GiaID) []GiaID {
	result := make([]GiaID, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		switch {
		case j == len(b) || i < len(a) && a[i] < b[j]:
			result = append(result, a[i])
			i++
		case i == len(a) || b[j] < a[i]:
			result = append(result, b[j])
			j++
		default:
			result = append(result, a[i])
			i++
			j++
		}
	}
	return result
}

type acecPolynomialBuilder struct {
	options AcecPolynomialOptions
	terms   map[string]AcecPolynomialTerm
}

func normalizeAcecPolynomialOptions(options AcecPolynomialOptions, objects int) (AcecPolynomialOptions, error) {
	if options.MaxObjects < 0 || options.MaxTerms < 0 || options.MaxDegree < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxTerms == 0 {
		options.MaxTerms = 1 << 20
	}
	if options.MaxDegree == 0 {
		options.MaxDegree = 1024
	}
	if objects > options.MaxObjects {
		return options, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return options, ErrCancelled
	}
	return options, nil
}

func newAcecPolynomialBuilder(options AcecPolynomialOptions) *acecPolynomialBuilder {
	return &acecPolynomialBuilder{options: options, terms: make(map[string]AcecPolynomialTerm)}
}

func (b *acecPolynomialBuilder) add(variables []GiaID, coefficient *big.Int) error {
	if coefficient.Sign() == 0 {
		return nil
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	if len(variables) > b.options.MaxDegree {
		return ErrLimit
	}
	key := acecMonomialKey(variables)
	if old, ok := b.terms[key]; ok {
		coefficient = new(big.Int).Add(old.Coefficient, coefficient)
		if coefficient.Sign() == 0 {
			delete(b.terms, key)
			return nil
		}
		b.terms[key] = AcecPolynomialTerm{coefficient, old.Variables}
		return nil
	}
	if len(b.terms) >= b.options.MaxTerms {
		return ErrLimit
	}
	b.terms[key] = AcecPolynomialTerm{new(big.Int).Set(coefficient), slices.Clone(variables)}
	return nil
}

func (b *acecPolynomialBuilder) polynomial() AcecPolynomial {
	terms := make([]AcecPolynomialTerm, 0, len(b.terms))
	for _, term := range b.terms {
		terms = append(terms, AcecPolynomialTerm{new(big.Int).Set(term.Coefficient), slices.Clone(term.Variables)})
	}
	sort.Slice(terms, func(i, j int) bool {
		a, c := terms[i].Variables, terms[j].Variables
		for k := 0; k < len(a) && k < len(c); k++ {
			if a[k] != c[k] {
				return a[k] < c[k]
			}
		}
		return len(a) < len(c)
	})
	return AcecPolynomial{terms}
}

func acecConstant(value int64, options AcecPolynomialOptions) AcecPolynomial {
	if value == 0 {
		return AcecPolynomial{}
	}
	return AcecPolynomial{[]AcecPolynomialTerm{{big.NewInt(value), nil}}}
}

func acecVariable(id GiaID) AcecPolynomial {
	return AcecPolynomial{[]AcecPolynomialTerm{{big.NewInt(1), []GiaID{id}}}}
}

func acecScale(p AcecPolynomial, scale *big.Int, options AcecPolynomialOptions) (AcecPolynomial, error) {
	b := newAcecPolynomialBuilder(options)
	for _, term := range p.terms {
		if err := b.add(term.Variables, new(big.Int).Mul(term.Coefficient, scale)); err != nil {
			return AcecPolynomial{}, err
		}
	}
	return b.polynomial(), nil
}

func acecAdd(a, b AcecPolynomial, options AcecPolynomialOptions) (AcecPolynomial, error) {
	builder := newAcecPolynomialBuilder(options)
	for _, polynomial := range []AcecPolynomial{a, b} {
		for _, term := range polynomial.terms {
			if err := builder.add(term.Variables, term.Coefficient); err != nil {
				return AcecPolynomial{}, err
			}
		}
	}
	return builder.polynomial(), nil
}

func acecSubtract(a, b AcecPolynomial, options AcecPolynomialOptions) (AcecPolynomial, error) {
	minus, err := acecScale(b, big.NewInt(-1), options)
	if err != nil {
		return AcecPolynomial{}, err
	}
	return acecAdd(a, minus, options)
}

func acecMultiply(a, b AcecPolynomial, options AcecPolynomialOptions) (AcecPolynomial, error) {
	builder := newAcecPolynomialBuilder(options)
	for _, left := range a.terms {
		for _, right := range b.terms {
			variables := acecMergeVariables(left.Variables, right.Variables)
			coefficient := new(big.Int).Mul(left.Coefficient, right.Coefficient)
			if err := builder.add(variables, coefficient); err != nil {
				return AcecPolynomial{}, err
			}
		}
	}
	return builder.polynomial(), nil
}

func acecComplement(p AcecPolynomial, options AcecPolynomialOptions) (AcecPolynomial, error) {
	return acecSubtract(acecConstant(1, options), p, options)
}

func acecXOR(a, b AcecPolynomial, options AcecPolynomialOptions) (AcecPolynomial, error) {
	product, err := acecMultiply(a, b, options)
	if err != nil {
		return AcecPolynomial{}, err
	}
	product, err = acecScale(product, big.NewInt(-2), options)
	if err != nil {
		return AcecPolynomial{}, err
	}
	sum, err := acecAdd(a, b, options)
	if err != nil {
		return AcecPolynomial{}, err
	}
	return acecAdd(sum, product, options)
}

// ExtractAcecPolynomial expands the weighted GIA outputs into a canonical
// integer polynomial. Output i has weight 2^i; when signed is true, the last
// output has negative weight, matching Gia_PolynBuild.
func ExtractAcecPolynomial(graph *GiaManager, signed bool, options AcecPolynomialOptions) (AcecPolynomial, error) {
	if graph == nil {
		return AcecPolynomial{}, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return AcecPolynomial{}, err
	}
	options, err := normalizeAcecPolynomialOptions(options, graph.ObjectCount())
	if err != nil {
		return AcecPolynomial{}, err
	}
	values := make([]AcecPolynomial, graph.ObjectCount())
	for _, id := range graph.cis {
		values[id] = acecVariable(id)
	}
	read := func(lit GiaLit) (AcecPolynomial, error) {
		value := values[lit.ID()]
		if lit.Complemented() {
			return acecComplement(value, options)
		}
		return value, nil
	}
	for rawID, object := range graph.objects {
		if options.Stop != nil && options.Stop() {
			return AcecPolynomial{}, ErrCancelled
		}
		if !object.IsAnd() {
			continue
		}
		id := GiaID(rawID)
		first, faninErr := object.Fanin0(id)
		if faninErr != nil {
			return AcecPolynomial{}, faninErr
		}
		a, opErr := read(first)
		if opErr != nil {
			return AcecPolynomial{}, opErr
		}
		if object.IsBuffer() {
			values[id] = a
			continue
		}
		second, faninErr := object.Fanin1(id)
		if faninErr != nil {
			return AcecPolynomial{}, faninErr
		}
		b, opErr := read(second)
		if opErr != nil {
			return AcecPolynomial{}, opErr
		}
		if control, mux := graph.MuxControl(id); mux {
			c, e := read(control)
			if e != nil {
				return AcecPolynomial{}, e
			}
			difference, e := acecSubtract(b, a, options)
			if e != nil {
				return AcecPolynomial{}, e
			}
			selected, e := acecMultiply(c, difference, options)
			if e != nil {
				return AcecPolynomial{}, e
			}
			values[id], opErr = acecAdd(a, selected, options)
		} else if object.IsXor() {
			values[id], opErr = acecXOR(a, b, options)
		} else {
			values[id], opErr = acecMultiply(a, b, options)
		}
		if opErr != nil {
			return AcecPolynomial{}, opErr
		}
	}
	result := AcecPolynomial{}
	for i, id := range graph.cos {
		object := graph.objects[id]
		driver, e := object.Fanin0(id)
		if e != nil {
			return AcecPolynomial{}, e
		}
		output, e := read(driver)
		if e != nil {
			return AcecPolynomial{}, e
		}
		weight := new(big.Int).Lsh(big.NewInt(1), uint(i))
		if signed && i == len(graph.cos)-1 {
			weight.Neg(weight)
		}
		output, e = acecScale(output, weight, options)
		if e != nil {
			return AcecPolynomial{}, e
		}
		result, e = acecAdd(result, output, options)
		if e != nil {
			return AcecPolynomial{}, e
		}
	}
	return result, nil
}

// AppendAcecHalfAdder returns sum and carry. Construction is transactional.
func AppendAcecHalfAdder(graph *GiaManager, inputs [2]GiaLit) ([2]GiaLit, error) {
	var zero [2]GiaLit
	if graph == nil || !graph.validSignal(inputs[0]) || !graph.validSignal(inputs[1]) {
		return zero, ErrInvalidLiteral
	}
	work := graph.Clone()
	carry, err := work.AppendAndSimplified(inputs[0], inputs[1])
	if err != nil {
		return zero, err
	}
	none, err := work.AppendAndSimplified(inputs[0].Not(), inputs[1].Not())
	if err != nil {
		return zero, err
	}
	equal, err := work.AppendOrSimplified(carry, none)
	if err != nil {
		return zero, err
	}
	result := [2]GiaLit{equal.Not(), carry}
	*graph = *work
	return result, nil
}

// AppendAcecFullAdder is the source's two-half-adder construction.
func AppendAcecFullAdder(graph *GiaManager, inputs [3]GiaLit) ([2]GiaLit, error) {
	var zero [2]GiaLit
	if graph == nil {
		return zero, ErrDimension
	}
	work := graph.Clone()
	first, err := AppendAcecHalfAdder(work, [2]GiaLit{inputs[0], inputs[1]})
	if err != nil {
		return zero, err
	}
	second, err := AppendAcecHalfAdder(work, [2]GiaLit{first[0], inputs[2]})
	if err != nil {
		return zero, err
	}
	carry, err := work.AppendOrSimplified(first[1], second[1])
	if err != nil {
		return zero, err
	}
	result := [2]GiaLit{second[0], carry}
	*graph = *work
	return result, nil
}

// BuildAcecAdderTree reduces ranked bit literals with the same first-three
// schedule as Acec_InsertTree. Inputs are copied and the graph changes only on
// complete success.
func BuildAcecAdderTree(graph *GiaManager, ranks [][]GiaLit, options AcecPolynomialOptions) ([]GiaLit, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	options, err := normalizeAcecPolynomialOptions(options, graph.ObjectCount())
	if err != nil {
		return nil, err
	}
	work := graph.Clone()
	levels := cloneGiaRanks(ranks)
	for _, level := range levels {
		for _, lit := range level {
			if !work.validSignal(lit) {
				return nil, ErrInvalidLiteral
			}
		}
	}
	roots := make([]GiaLit, 0, len(levels)+5)
	for rank := 0; rank < len(levels); rank++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		for len(levels[rank]) > 1 {
			if len(levels[rank]) == 2 {
				levels[rank] = append(levels[rank], GiaConst0)
			}
			inputs := [3]GiaLit{levels[rank][0], levels[rank][1], levels[rank][2]}
			levels[rank] = levels[rank][3:]
			outputs, e := AppendAcecFullAdder(work, inputs)
			if e != nil {
				return nil, e
			}
			levels[rank] = append(levels[rank], outputs[0])
			if rank+1 == len(levels) {
				levels = append(levels, nil)
			}
			levels[rank+1] = append(levels[rank+1], outputs[1])
			if work.ObjectCount() > options.MaxObjects {
				return nil, ErrLimit
			}
		}
		if len(levels[rank]) == 0 {
			roots = append(roots, GiaConst0)
		} else {
			roots = append(roots, levels[rank][0])
		}
	}
	*graph = *work
	return roots, nil
}

type AcecFullAdder struct {
	Inputs     [3]GiaID
	Sum, Carry GiaID
}

type AcecHalfAdder struct {
	Inputs     [2]GiaID
	Sum, Carry GiaID
}

// FindAcecPolynomialOrder returns a rank for every object. It recognizes
// supplied adders from the active output frontier, then topologically ranks all
// remaining gates. This is the owned deterministic counterpart of
// Gia_PolynFindOrder plus Gia_PolynReorder.
func FindAcecPolynomialOrder(graph *GiaManager, full []AcecFullAdder, half []AcecHalfAdder, options AcecPolynomialOptions) ([]int, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	options, err := normalizeAcecPolynomialOptions(options, graph.ObjectCount())
	if err != nil {
		return nil, err
	}
	checkID := func(id GiaID) bool { return uint64(id) < uint64(graph.ObjectCount()) }
	for _, adder := range full {
		if !checkID(adder.Sum) || !checkID(adder.Carry) || !checkID(adder.Inputs[0]) || !checkID(adder.Inputs[1]) || !checkID(adder.Inputs[2]) {
			return nil, ErrInvalidLiteral
		}
		if adder.Sum == adder.Carry || !graph.objects[adder.Sum].IsAnd() || !graph.objects[adder.Carry].IsAnd() || adder.Inputs[0] >= adder.Sum || adder.Inputs[0] >= adder.Carry || adder.Inputs[1] >= adder.Sum || adder.Inputs[1] >= adder.Carry || adder.Inputs[2] >= adder.Sum || adder.Inputs[2] >= adder.Carry {
			return nil, ErrDimension
		}
	}
	for _, adder := range half {
		if !checkID(adder.Sum) || !checkID(adder.Carry) || !checkID(adder.Inputs[0]) || !checkID(adder.Inputs[1]) {
			return nil, ErrInvalidLiteral
		}
		if adder.Sum == adder.Carry || !graph.objects[adder.Sum].IsAnd() || !graph.objects[adder.Carry].IsAnd() || adder.Inputs[0] >= adder.Sum || adder.Inputs[0] >= adder.Carry || adder.Inputs[1] >= adder.Sum || adder.Inputs[1] >= adder.Carry {
			return nil, ErrDimension
		}
	}
	active := make([]bool, graph.ObjectCount())
	for _, co := range graph.cos {
		driver, _ := graph.objects[co].Fanin0(co)
		active[driver.ID()] = true
	}
	groups := make([][]GiaID, 0, len(full)+len(half))
	usedFull := make([]bool, len(full))
	usedHalf := make([]bool, len(half))
	changed := true
	for changed {
		changed = false
		for i := len(full) - 1; i >= 0; i-- {
			a := full[i]
			if !usedFull[i] && active[a.Sum] && active[a.Carry] {
				usedFull[i] = true
				active[a.Sum], active[a.Carry] = false, false
				for _, id := range a.Inputs {
					active[id] = true
				}
				groups = append(groups, []GiaID{a.Sum, a.Carry})
				changed = true
			}
		}
		for i := len(half) - 1; i >= 0; i-- {
			a := half[i]
			if !usedHalf[i] && active[a.Sum] && active[a.Carry] {
				usedHalf[i] = true
				active[a.Sum], active[a.Carry] = false, false
				for _, id := range a.Inputs {
					active[id] = true
				}
				groups = append(groups, []GiaID{a.Sum, a.Carry})
				changed = true
			}
		}
	}
	gateOrder := make([]GiaID, 0, graph.AndCount())
	seen := make([]bool, graph.ObjectCount())
	var collect func(GiaID) error
	collect = func(id GiaID) error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		object := graph.objects[id]
		if !object.IsAnd() {
			return nil
		}
		first, e := object.Fanin0(id)
		if e != nil {
			return e
		}
		if e = collect(first.ID()); e != nil {
			return e
		}
		if !object.IsBuffer() {
			second, faninErr := object.Fanin1(id)
			if faninErr != nil {
				return faninErr
			}
			if e = collect(second.ID()); e != nil {
				return e
			}
		}
		if control, yes := graph.MuxControl(id); yes {
			if e = collect(control.ID()); e != nil {
				return e
			}
		}
		gateOrder = append(gateOrder, id)
		return nil
	}
	// Leftover frontier precedes recognized arithmetic groups, as in vRecord2.
	for id, isActive := range active {
		if isActive {
			if err := collect(GiaID(id)); err != nil {
				return nil, err
			}
		}
	}
	for i := len(groups) - 1; i >= 0; i-- {
		for _, id := range groups[i] {
			if err := collect(id); err != nil {
				return nil, err
			}
		}
	}
	// Invalid or incomplete detector input cannot omit otherwise live gates.
	for _, co := range graph.cos {
		driver, _ := graph.objects[co].Fanin0(co)
		if err := collect(driver.ID()); err != nil {
			return nil, err
		}
	}
	order := make([]int, graph.ObjectCount())
	rank := 1
	for _, id := range graph.cis {
		order[id] = rank
		rank++
	}
	for _, id := range gateOrder {
		order[id] = rank
		rank++
	}
	return order, nil
}

func acecReferenceCounts(graph *GiaManager) ([]int, error) {
	refs := make([]int, graph.ObjectCount())
	for rawID, object := range graph.objects {
		id := GiaID(rawID)
		if object.IsAnd() || object.IsCO() {
			first, err := object.Fanin0(id)
			if err != nil {
				return nil, err
			}
			refs[first.ID()]++
			if object.IsAnd() && !object.IsBuffer() {
				second, err := object.Fanin1(id)
				if err != nil {
					return nil, err
				}
				refs[second.ID()]++
			}
			if control, yes := graph.MuxControl(id); yes {
				refs[control.ID()]++
			}
		}
	}
	return refs, nil
}

// CollectAcecLastXOR returns the root-first, single-reference native-XOR cone
// driving the last output, matching Gia_PolynCollectLastXor without mutating
// traversal marks or reference storage.
func CollectAcecLastXOR(graph *GiaManager, options AcecPolynomialOptions) ([]GiaID, error) {
	if graph == nil || graph.OutputCount() == 0 {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	options, err := normalizeAcecPolynomialOptions(options, graph.ObjectCount())
	if err != nil {
		return nil, err
	}
	refs, err := acecReferenceCounts(graph)
	if err != nil {
		return nil, err
	}
	last := graph.cos[len(graph.cos)-1]
	driver, _ := graph.objects[last].Fanin0(last)
	seen := make([]bool, graph.ObjectCount())
	result := make([]GiaID, 0)
	var visit func(GiaID) error
	visit = func(id GiaID) error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		object := graph.objects[id]
		if !object.IsAnd() || !object.IsXor() || refs[id] > 1 {
			return nil
		}
		first, _ := object.Fanin0(id)
		second, _ := object.Fanin1(id)
		if err := visit(first.ID()); err != nil {
			return err
		}
		if err := visit(second.ID()); err != nil {
			return err
		}
		result = append(result, id)
		if len(result) > options.MaxObjects {
			return ErrLimit
		}
		return nil
	}
	if err := visit(driver.ID()); err != nil {
		return nil, err
	}
	slices.Reverse(result)
	return result, nil
}

// AnalyzeAcecXORChains returns the source diagnostic counts as owned data.
func AnalyzeAcecXORChains(graph *GiaManager, options AcecPolynomialOptions) ([]int, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	options, err := normalizeAcecPolynomialOptions(options, graph.ObjectCount())
	if err != nil {
		return nil, err
	}
	refs, err := acecReferenceCounts(graph)
	if err != nil {
		return nil, err
	}
	counts := make([]int, len(graph.cos))
	for output, co := range graph.cos {
		driver, _ := graph.objects[co].Fanin0(co)
		seen := make([]bool, graph.ObjectCount())
		var visit func(GiaID) error
		visit = func(id GiaID) error {
			if options.Stop != nil && options.Stop() {
				return ErrCancelled
			}
			if seen[id] {
				return nil
			}
			seen[id] = true
			object := graph.objects[id]
			if !object.IsAnd() || !object.IsXor() || refs[id] > 1 {
				return nil
			}
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			if err := visit(first.ID()); err != nil {
				return err
			}
			if err := visit(second.ID()); err != nil {
				return err
			}
			counts[output]++
			return nil
		}
		if err := visit(driver.ID()); err != nil {
			return nil, err
		}
	}
	return counts, nil
}

// DuplicateAcecTopMostRange preserves Gia_ManDupTopMostRange's diagnostic
// choice of output ordinals 45 through 51 and fanin depth 100. The returned
// boundary mapping is owned, and short graphs return an error instead of
// reading beyond their output array.
func DuplicateAcecTopMostRange(graph *GiaManager, options GiaCofactorOptions) (GiaCutExtraction, error) {
	if graph == nil {
		return GiaCutExtraction{}, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return GiaCutExtraction{}, err
	}
	if graph.OutputCount() < 52 {
		return GiaCutExtraction{}, ErrDimension
	}
	roots := make([]GiaID, 0, 7)
	for output := 45; output < 52; output++ {
		co := graph.cos[output]
		driver, err := graph.objects[co].Fanin0(co)
		if err != nil {
			return GiaCutExtraction{}, err
		}
		roots = append(roots, driver.ID())
	}
	return graph.ExtractDepthLimitedCones(roots, 100, options)
}

func (p AcecPolynomial) String() string {
	if len(p.terms) == 0 {
		return "0"
	}
	result := ""
	for i, term := range p.terms {
		if i != 0 && term.Coefficient.Sign() >= 0 {
			result += "+"
		}
		result += term.Coefficient.String()
		for _, id := range term.Variables {
			result += fmt.Sprintf("*x%d", id)
		}
	}
	return result
}
