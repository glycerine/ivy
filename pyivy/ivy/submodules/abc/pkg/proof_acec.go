package abc

// Port of the standalone arithmetic semantics in src/proof/acec/acec2Mult.c,
// acecBo.c, acecCl.c, acecCo.c, acecCore.c, acecCover.c, acecFadds.c,
// acecMult.c, acecPa.c, acecPo.c, acecPool.c, acecRe.c, acecSt.c,
// acecStruct.c, acecTree.c, acecXor.c, and their module.make build unit.

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"slices"
	"sort"
	"sync"
)

// AcecPortUnit describes the disposition of an ACEC build unit. It is kept in
// source order so reports and tests do not depend on map iteration order.
type AcecPortUnit struct {
	Source, Definitions, Disposition string
}

// AcecPortInventory returns owned records for every previously unported ACEC
// build boundary. Exact arithmetic, normalization, ordering, cuts, adders,
// trees, signatures, and counterexamples are implemented in Go. ABC command,
// SAT, timing-box, truth-canonization, and experimental detector entry points
// remain explicitly identified rather than silently approximated.
func AcecPortInventory() []AcecPortUnit {
	units := []AcecPortUnit{
		{"acec2Mult.c", "Sdb_Cut*, Sdb_Sto* cut storage, merge, truth, iteration and detection", "bounded deterministic Go polynomial normalization supersedes storage; specialized six-variable cut detector remains a heuristic gap"},
		{"acecBo.c", "Acec_DetectBoothXorMux/Fanin/One/Two/Test", "Booth arithmetic is covered exactly by normalized Boolean polynomials; pattern-name detector remains a heuristic gap"},
		{"acecCl.c", "Acec_ManDerive, XOR support/collection, rewrite, de-clause", "semantic normalization and owned graph validation ported; experimental rewrite heuristic remains a gap"},
		{"acecCo.c", "Gia_PolynCore map/order/collect/duplicate/detect", "canonical order and arithmetic comparison ported; cone-printing and detector driver omitted"},
		{"acecCore.c", "parameters, classes, common graph, matching, Acec_Solve", "exact polynomial proof and bounded counterexamples ported; SAT conflict/time command orchestration remains a solver gap"},
		{"acecCover.c", "adder marking, XOR collection/exploration/cover", "exact cone semantics ported; diagnostic cover printing omitted"},
		{"acecFadds.c", "half/full-adder detection, cuts, chains, boxed duplication, delay paths", "adder construction and exact arithmetic ported; timing-manager boxes and detector heuristics remain gaps"},
		{"acecMult.c", "NPN canonization, multiplier inputs/partial products/Booth PPG", "multiplication semantics are exact through polynomial expansion; NPN tables and pattern detectors remain heuristic gaps"},
		{"acecPa.c", "phase verification/propagation and cut computation", "polarity is represented exactly by complemented literals; phase detector remains a structural gap"},
		{"acecPo.c", "signature parse/distribute/print and polynomial build/result", "canonical coefficient/monomial result and stable digest ported"},
		{"acecPool.c", "carry maps/roots, box sets/ranks/profile/top-most pool", "owned ranked adders and deterministic ordering ported; profiling output omitted"},
		{"acecRe.c", "truth/cut computation, derive/remove/print adders", "exact polynomial normalization ported; local cut-recognition heuristic remains a gap"},
		{"acecSt.c", "NPN table generation and statistics collection", "stable semantic signature ported; NPN statistics remain a diagnostic gap"},
		{"acecStruct.c", "XOR roots/ranks/tree leaves/shadows/support/nexts", "validated graph traversal and canonical order ported; shadow heuristic remains a gap"},
		{"acecTree.c", "box lifecycle, filters, phases, points, trees, box derivation", "owned AcecBox lifecycle, adder trees, normalization and comparison ported"},
		{"acecXor.c", "XOR checks/maps/roots/ranks/leaves and box production", "XOR Boolean semantics and canonical polynomial form ported; box-production heuristic remains a gap"},
		{"module.make", "19 ACEC compilation units", "all build units inventoried; acecNorm/Order/Polyn/Util are implemented in proof_acec_polynomial.go"},
	}
	return slices.Clone(units)
}

type AcecEquivalenceOptions struct {
	Signed                                          bool
	MaxObjects, MaxTerms, MaxDegree, MaxDepth       int
	MaxCoefficientBits                              int
	MaxInputBits, MaxAssignments, MaxWork, MaxBytes int
	Stop                                            func() bool
}

// AcecNormalizedTerm uses input positions rather than graph object IDs, so
// independently constructed graphs have directly comparable signatures.
type AcecNormalizedTerm struct {
	Coefficient *big.Int
	Inputs      []int
}

type AcecArithmeticSignature struct {
	Signed bool
	Width  int
	Terms  []AcecNormalizedTerm
	Digest [32]byte
}

func (s AcecArithmeticSignature) Clone() AcecArithmeticSignature {
	r := s
	r.Terms = make([]AcecNormalizedTerm, len(s.Terms))
	for i, t := range s.Terms {
		r.Terms[i] = AcecNormalizedTerm{new(big.Int).Set(t.Coefficient), slices.Clone(t.Inputs)}
	}
	return r
}

type AcecEquivalenceResult struct {
	Equivalent, Proven bool
	Witness            []bool
	LeftValue          *big.Int
	RightValue         *big.Int
	Left, Right        AcecArithmeticSignature
	Assignments, Work  uint64
}

func acecNormalizeEquivalenceOptions(o AcecEquivalenceOptions) (AcecEquivalenceOptions, error) {
	values := []int{o.MaxObjects, o.MaxTerms, o.MaxDegree, o.MaxDepth, o.MaxCoefficientBits, o.MaxInputBits, o.MaxAssignments, o.MaxWork, o.MaxBytes}
	for _, value := range values {
		if value < 0 {
			return o, ErrDimension
		}
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxTerms == 0 {
		o.MaxTerms = 1 << 20
	}
	if o.MaxDegree == 0 {
		o.MaxDegree = 1024
	}
	if o.MaxDepth == 0 {
		o.MaxDepth = 1 << 20
	}
	if o.MaxCoefficientBits == 0 {
		o.MaxCoefficientBits = 1 << 20
	}
	if o.MaxInputBits == 0 {
		o.MaxInputBits = 20
	}
	if o.MaxAssignments == 0 {
		o.MaxAssignments = 1 << 20
	}
	if o.MaxWork == 0 {
		o.MaxWork = 1 << 28
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 1 << 30
	}
	if o.Stop != nil && o.Stop() {
		return o, ErrCancelled
	}
	return o, nil
}

func acecGraphDepth(g *GiaManager, o AcecEquivalenceOptions) (int, int, error) {
	depth := make([]int, g.ObjectCount())
	work := 0
	for raw, object := range g.objects {
		if o.Stop != nil && o.Stop() {
			return 0, work, ErrCancelled
		}
		if !object.IsAnd() {
			continue
		}
		id := GiaID(raw)
		a, err := object.Fanin0(id)
		if err != nil {
			return 0, work, err
		}
		d := depth[a.ID()]
		if !object.IsBuffer() {
			b, e := object.Fanin1(id)
			if e != nil {
				return 0, work, e
			}
			if depth[b.ID()] > d {
				d = depth[b.ID()]
			}
		}
		if c, ok := g.MuxControl(id); ok && depth[c.ID()] > d {
			d = depth[c.ID()]
		}
		depth[id] = d + 1
		work++
		if depth[id] > o.MaxDepth || work > o.MaxWork {
			return 0, work, ErrLimit
		}
	}
	maximum := 0
	for _, d := range depth {
		if d > maximum {
			maximum = d
		}
	}
	return maximum, work, nil
}

// NormalizeAcecArithmetic computes the exact multilinear integer polynomial
// of the output bit-vector, with two's-complement weight on the high bit when
// Signed is set. A failure returns no partial signature.
func NormalizeAcecArithmetic(g *GiaManager, options AcecEquivalenceOptions) (AcecArithmeticSignature, error) {
	var zero AcecArithmeticSignature
	if g == nil {
		return zero, ErrDimension
	}
	if err := g.Check(); err != nil {
		return zero, err
	}
	o, err := acecNormalizeEquivalenceOptions(options)
	if err != nil {
		return zero, err
	}
	if g.ObjectCount() > o.MaxObjects {
		return zero, ErrLimit
	}
	_, work, err := acecGraphDepth(g, o)
	if err != nil {
		return zero, err
	}
	poly, err := ExtractAcecPolynomial(g, o.Signed, AcecPolynomialOptions{MaxObjects: o.MaxObjects, MaxTerms: o.MaxTerms, MaxDegree: o.MaxDegree, Stop: o.Stop})
	if err != nil {
		return zero, err
	}
	inputIndex := make(map[GiaID]int, len(g.cis))
	for i, id := range g.cis {
		inputIndex[id] = i
	}
	result := AcecArithmeticSignature{Signed: o.Signed, Width: g.OutputCount(), Terms: make([]AcecNormalizedTerm, len(poly.terms))}
	bytesUsed := 0
	for i, term := range poly.terms {
		if term.Coefficient.BitLen() > o.MaxCoefficientBits {
			return zero, ErrLimit
		}
		inputs := make([]int, len(term.Variables))
		for j, id := range term.Variables {
			index, ok := inputIndex[id]
			if !ok {
				return zero, ErrInvariant
			}
			inputs[j] = index
		}
		sort.Ints(inputs)
		result.Terms[i] = AcecNormalizedTerm{new(big.Int).Set(term.Coefficient), inputs}
		bytesUsed += len(term.Coefficient.Bytes()) + 8*len(inputs) + 16
		work++
		if bytesUsed > o.MaxBytes || work > o.MaxWork {
			return zero, ErrLimit
		}
	}
	sort.Slice(result.Terms, func(i, j int) bool {
		a, b := result.Terms[i].Inputs, result.Terms[j].Inputs
		for k := 0; k < len(a) && k < len(b); k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return len(a) < len(b)
	})
	h := sha256.New()
	var buffer [8]byte
	if o.Signed {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
	binary.LittleEndian.PutUint64(buffer[:], uint64(result.Width))
	h.Write(buffer[:])
	for _, term := range result.Terms {
		sign := byte(0)
		if term.Coefficient.Sign() < 0 {
			sign = 1
		}
		h.Write([]byte{sign})
		magnitude := new(big.Int).Abs(term.Coefficient).Bytes()
		binary.LittleEndian.PutUint64(buffer[:], uint64(len(magnitude)))
		h.Write(buffer[:])
		h.Write(magnitude)
		binary.LittleEndian.PutUint64(buffer[:], uint64(len(term.Inputs)))
		h.Write(buffer[:])
		for _, input := range term.Inputs {
			binary.LittleEndian.PutUint64(buffer[:], uint64(input))
			h.Write(buffer[:])
		}
	}
	copy(result.Digest[:], h.Sum(nil))
	return result, nil
}

func acecSignaturesEqual(a, b AcecArithmeticSignature) bool {
	if a.Signed != b.Signed || a.Width != b.Width || a.Digest != b.Digest || len(a.Terms) != len(b.Terms) {
		return false
	}
	for i := range a.Terms {
		if a.Terms[i].Coefficient.Cmp(b.Terms[i].Coefficient) != 0 || !slices.Equal(a.Terms[i].Inputs, b.Terms[i].Inputs) {
			return false
		}
	}
	return true
}

func acecSignatureBytes(s AcecArithmeticSignature) int {
	total := 33
	for _, term := range s.Terms {
		total += len(term.Coefficient.Bytes()) + 8*len(term.Inputs) + 16
	}
	return total
}

func acecOutputValue(bits []bool, signed bool) *big.Int {
	r := new(big.Int)
	for i, bit := range bits {
		if bit {
			w := new(big.Int).Lsh(big.NewInt(1), uint(i))
			if signed && i == len(bits)-1 {
				w.Neg(w)
			}
			r.Add(r, w)
		}
	}
	return r
}

// CheckAcecArithmeticEquivalent first proves equality by canonical polynomial
// identity. For a mismatch it searches assignments in stable binary order and
// returns the first owned witness when within the configured exhaustive bound.
func CheckAcecArithmeticEquivalent(left, right *GiaManager, options AcecEquivalenceOptions) (AcecEquivalenceResult, error) {
	var zero AcecEquivalenceResult
	if left == nil || right == nil {
		return zero, ErrDimension
	}
	o, err := acecNormalizeEquivalenceOptions(options)
	if err != nil {
		return zero, err
	}
	if left.InputCount() != right.InputCount() || left.OutputCount() != right.OutputCount() {
		return zero, ErrDimension
	}
	// MaxObjects is a budget for the complete comparison, not an allowance for
	// each side independently.  Reject before polynomial expansion so a caller
	// cannot make us do twice the configured work and only then receive a limit
	// error.
	if left.ObjectCount() > o.MaxObjects-right.ObjectCount() {
		return zero, ErrLimit
	}
	ls, err := NormalizeAcecArithmetic(left, o)
	if err != nil {
		return zero, err
	}
	rs, err := NormalizeAcecArithmetic(right, o)
	if err != nil {
		return zero, err
	}
	result := AcecEquivalenceResult{Left: ls.Clone(), Right: rs.Clone(), Work: uint64(left.ObjectCount() + right.ObjectCount() + len(ls.Terms) + len(rs.Terms))}
	if left.ObjectCount()+right.ObjectCount() > o.MaxObjects || len(ls.Terms)+len(rs.Terms) > o.MaxTerms || acecSignatureBytes(ls)+acecSignatureBytes(rs) > o.MaxBytes || result.Work > uint64(o.MaxWork) {
		return zero, ErrLimit
	}
	if acecSignaturesEqual(ls, rs) {
		result.Equivalent = true
		result.Proven = true
		return result, nil
	}
	n := left.InputCount()
	if n > o.MaxInputBits || n >= 63 {
		return result, nil
	}
	total := uint64(1) << uint(n)
	if total > uint64(o.MaxAssignments) {
		return result, nil
	}
	inputs := make([]bool, n)
	for assignment := uint64(0); assignment < total; assignment++ {
		if o.Stop != nil && o.Stop() {
			return zero, ErrCancelled
		}
		if result.Work+uint64(left.ObjectCount()+right.ObjectCount()) > uint64(o.MaxWork) {
			return zero, ErrLimit
		}
		for i := range inputs {
			inputs[i] = assignment&(uint64(1)<<uint(i)) != 0
		}
		lo, e := left.Evaluate(inputs)
		if e != nil {
			return zero, e
		}
		ro, e := right.Evaluate(inputs)
		if e != nil {
			return zero, e
		}
		result.Assignments++
		result.Work += uint64(left.ObjectCount() + right.ObjectCount())
		lv, rv := acecOutputValue(lo, o.Signed), acecOutputValue(ro, o.Signed)
		if lv.Cmp(rv) != 0 {
			result.Witness = slices.Clone(inputs)
			result.LeftValue = lv
			result.RightValue = rv
			result.Proven = true
			return result, nil
		}
	}
	// Distinct canonical polynomials must differ on some Boolean assignment;
	// reaching here would indicate a broken normalizer rather than equivalence.
	return zero, ErrInvariant
}

// AcecEquivalenceChecker provides an explicit, concurrency-safe lifecycle.
// Checks are independent and return owned results.
type AcecEquivalenceChecker struct {
	mu      sync.Mutex
	options AcecEquivalenceOptions
	closed  bool
}

func NewAcecEquivalenceChecker(options AcecEquivalenceOptions) (*AcecEquivalenceChecker, error) {
	o, err := acecNormalizeEquivalenceOptions(options)
	if err != nil {
		return nil, err
	}
	return &AcecEquivalenceChecker{options: o}, nil
}
func (c *AcecEquivalenceChecker) Check(left, right *GiaManager) (AcecEquivalenceResult, error) {
	if c == nil {
		return AcecEquivalenceResult{}, ErrDimension
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return AcecEquivalenceResult{}, ErrInvariant
	}
	// Options are immutable after construction.  Snapshot them while holding
	// the lifecycle lock, then release it before invoking Stop or doing lengthy
	// graph work.  In particular, a Stop callback may safely call Close.
	options := c.options
	c.mu.Unlock()
	return CheckAcecArithmeticEquivalent(left, right, options)
}
func (c *AcecEquivalenceChecker) Close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
}
