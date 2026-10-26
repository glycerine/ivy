package abc

// Semi-canonical AIG recording from src/aig/aig/aigCanon.c.

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

const AigCanonicalMaximumVariables = 12

// AigCanonicalSignature is Aig_VSig_t with owned cofactor storage.  The source
// comparison intentionally uses Ones only; CofactorOnes is retained for
// diagnostics and for checking the complete signature calculation.
type AigCanonicalSignature struct {
	Ones         int
	CofactorOnes []int
}

// AigSemiCanonicalization maps Source to Table. SourceToTarget maps each
// source variable to its position in Table. InputPhases is indexed by source
// variable. TargetToSource is the inverse permutation.
type AigSemiCanonicalization struct {
	Table                          *TruthTable
	SourceToTarget, TargetToSource []int
	InputPhases                    []bool
	Signatures                     []AigCanonicalSignature
	VariablesUnique                bool
}

// Reconstruct reverses this semi-canonical transformation.
func (r AigSemiCanonicalization) Reconstruct() (*TruthTable, error) {
	if r.Table == nil || len(r.TargetToSource) != r.Table.variables || len(r.InputPhases) != r.Table.variables {
		return nil, ErrDimension
	}
	phases := make([]bool, len(r.InputPhases))
	for target, source := range r.TargetToSource {
		if source < 0 || source >= len(phases) {
			return nil, ErrInvariant
		}
		phases[target] = r.InputPhases[source]
	}
	return r.Table.TransformNPN(r.TargetToSource, phases, false)
}

// AigSemiCanonicalize implements Aig_RManComputeVSigs,
// Aig_RManVarsAreUnique and Aig_RManSemiCanonicize without mutating source.
// It preserves the source's stable adjacent swap order and its Ones-only
// signature comparison.
func AigSemiCanonicalize(source *TruthTable, stop func() bool) (AigSemiCanonicalization, error) {
	fail := func(err error) (AigSemiCanonicalization, error) { return AigSemiCanonicalization{}, err }
	if source == nil {
		return fail(ErrDimension)
	}
	if err := source.Check(); err != nil {
		return fail(err)
	}
	variables := source.variables
	if variables > AigCanonicalMaximumVariables {
		return fail(ErrLimit)
	}
	if stop != nil && stop() {
		return fail(ErrCancelled)
	}
	table := source.Clone()
	signatures := make([]AigCanonicalSignature, 2*variables)
	for variable := 0; variable < variables; variable++ {
		for value := 0; value < 2; value++ {
			if stop != nil && stop() {
				return fail(ErrCancelled)
			}
			cofactor, err := source.Cofactor(variable, value != 0)
			if err != nil {
				return fail(err)
			}
			population, err := cofactor.CofactorPopulations()
			if err != nil {
				return fail(err)
			}
			counts := append([]int(nil), population.Zero...)
			// Aig_RManSortNums is a descending selection sort. Equal values
			// need no secondary ordering because only Ones is compared later.
			for i := 0; i+1 < len(counts); i++ {
				best := i
				for j := i + 1; j < len(counts); j++ {
					if counts[j] > counts[best] {
						best = j
					}
				}
				counts[i], counts[best] = counts[best], counts[i]
			}
			signatures[2*variable+value] = AigCanonicalSignature{Ones: population.Ones, CofactorOnes: counts}
		}
	}
	phases := make([]bool, variables)
	for variable := 0; variable < variables; variable++ {
		if signatures[2*variable].Ones <= signatures[2*variable+1].Ones {
			continue
		}
		phases[variable] = true
		signatures[2*variable], signatures[2*variable+1] = signatures[2*variable+1], signatures[2*variable]
		table.flipVariable(variable)
	}
	targetToSource := make([]int, variables)
	for i := range targetToSource {
		targetToSource[i] = i
	}
	for {
		changed := false
		for variable := 0; variable+1 < variables; variable++ {
			if signatures[2*variable].Ones <= signatures[2*(variable+1)].Ones {
				continue
			}
			changed = true
			targetToSource[variable], targetToSource[variable+1] = targetToSource[variable+1], targetToSource[variable]
			signatures[2*variable], signatures[2*(variable+1)] = signatures[2*(variable+1)], signatures[2*variable]
			signatures[2*variable+1], signatures[2*(variable+1)+1] = signatures[2*(variable+1)+1], signatures[2*variable+1]
			table.swapVariables(variable, variable+1)
		}
		if !changed {
			break
		}
		if stop != nil && stop() {
			return fail(ErrCancelled)
		}
	}
	sourceToTarget := make([]int, variables)
	for target, sourceVariable := range targetToSource {
		sourceToTarget[sourceVariable] = target
	}
	unique := true
	for variable := 0; variable+1 < variables; variable++ {
		if signatures[2*variable].Ones == signatures[2*(variable+1)].Ones {
			unique = false
			break
		}
	}
	return AigSemiCanonicalization{
		Table: table, SourceToTarget: sourceToTarget, TargetToSource: targetToSource,
		InputPhases: phases, Signatures: signatures, VariablesUnique: unique,
	}, nil
}

type AigCanonicalOptions struct {
	MaxVariables, MaxEntries, MaxAigObjects, MaxDecompositionNodes int
	Stop                                                           func() bool
}

type AigCanonicalStatistics struct {
	Total, FullDSD, PartialDSD, NonDSD, UniqueVariables int
	VariableFunctions                                   [AigCanonicalMaximumVariables + 1]int
}

type AigCanonicalEntry struct {
	ID, Visits, Variables int
	Table                 *TruthTable
}

type AigCanonicalRecord struct {
	EntryID     int
	New         bool
	SkippedDSD  bool
	PrimeInputs int
	Transform   AigSemiCanonicalization
}

// AigCanonicalSession replaces aigCanon.c's process-global s_pRMan. Entries
// and outputs retain deterministic first-encounter order.
type AigCanonicalSession struct {
	options    AigCanonicalOptions
	graph      *AigManager
	entries    []AigCanonicalEntry
	lookup     map[string]int
	statistics AigCanonicalStatistics
	closed     bool
}

func NewAigCanonicalSession(options AigCanonicalOptions) (*AigCanonicalSession, error) {
	if options.MaxVariables < 0 || options.MaxEntries < 0 || options.MaxAigObjects < 0 || options.MaxDecompositionNodes < 0 {
		return nil, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = AigCanonicalMaximumVariables
	}
	if options.MaxVariables > AigCanonicalMaximumVariables {
		return nil, ErrLimit
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	if options.MaxAigObjects == 0 {
		options.MaxAigObjects = 1 << 24
	}
	if options.MaxDecompositionNodes == 0 {
		options.MaxDecompositionNodes = 1000
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	graph := NewAigManager()
	if options.MaxVariables != 0 {
		if _, err := graph.IthVar(options.MaxVariables - 1); err != nil {
			return nil, err
		}
	}
	return &AigCanonicalSession{options: options, graph: graph, lookup: make(map[string]int)}, nil
}

func aigCanonicalKey(table *TruthTable) string {
	data := make([]byte, 2+8*len(table.words))
	binary.LittleEndian.PutUint16(data, uint16(table.variables))
	for i, word := range table.words {
		binary.LittleEndian.PutUint64(data[2+8*i:], word)
	}
	return string(data)
}

func aigCanonicalPrimeTable(formula *DauDSDFormula, size int) (*TruthTable, error) {
	if formula == nil || size < 0 {
		return nil, ErrDimension
	}
	for _, node := range formula.nodes {
		if node.kind != DauDSDPrime || len(node.children) != size {
			continue
		}
		table, err := NewConstantTruthTable(size, false, 0)
		if err != nil {
			return nil, err
		}
		if len(node.primeBits) != 1<<uint(size) {
			return nil, ErrInvariant
		}
		for assignment, value := range node.primeBits {
			if value != node.complement {
				if err = table.Set(uint64(assignment), true); err != nil {
					return nil, err
				}
			}
		}
		return table, nil
	}
	return nil, ErrInvariant
}

// Record performs aigCanon.c's DSD classification and records the largest
// non-DSD prime. Fully decomposable functions and three-input primes are
// counted but intentionally not stored, matching Aig_RManRecord.
func (s *AigCanonicalSession) Record(source *TruthTable) (AigCanonicalRecord, error) {
	if s == nil || s.closed || source == nil {
		return AigCanonicalRecord{}, ErrDimension
	}
	if err := source.Check(); err != nil {
		return AigCanonicalRecord{}, err
	}
	if source.variables > s.options.MaxVariables {
		return AigCanonicalRecord{}, ErrLimit
	}
	decomposition, err := DauDSDDecompose(source, DauDSDDecomposeOptions{
		MaxWords: len(source.words), MaxNodes: 4096, MaxDepth: 128,
		Stop: s.options.Stop,
	})
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	statistics := s.statistics
	statistics.Total++
	if decomposition.LargestPrime == 0 || decomposition.LargestPrime == 3 {
		statistics.FullDSD++
		s.statistics = statistics
		return AigCanonicalRecord{EntryID: -1, SkippedDSD: true, PrimeInputs: decomposition.LargestPrime}, nil
	}
	formula, err := ParseDauDSD(decomposition.Formula, DauDSDOptions{
		MaxBytes: 1 << 20, MaxNodes: 4096, MaxDepth: 128,
		MaxVariables: s.options.MaxVariables, MaxWords: len(source.words), Stop: s.options.Stop,
	})
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	prime, err := aigCanonicalPrimeTable(formula, decomposition.LargestPrime)
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	result, err := s.recordPrime(prime)
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	statistics.VariableFunctions[prime.variables]++
	if prime.variables < source.variables {
		statistics.PartialDSD++
	} else {
		statistics.NonDSD++
	}
	if result.Transform.VariablesUnique {
		statistics.UniqueVariables++
	}
	s.statistics = statistics
	result.PrimeInputs = prime.variables
	return result, nil
}

// RecordPrime records a supplied prime truth directly. It exposes the source's
// table, semi-canonicalization and SaveOne machinery without DSD classification.
func (s *AigCanonicalSession) RecordPrime(source *TruthTable) (AigCanonicalRecord, error) {
	if s == nil || s.closed || source == nil {
		return AigCanonicalRecord{}, ErrDimension
	}
	return s.recordPrime(source)
}

func (s *AigCanonicalSession) recordPrime(source *TruthTable) (AigCanonicalRecord, error) {
	if err := source.Check(); err != nil {
		return AigCanonicalRecord{}, err
	}
	if source.variables > s.options.MaxVariables {
		return AigCanonicalRecord{}, ErrLimit
	}
	if s.options.Stop != nil && s.options.Stop() {
		return AigCanonicalRecord{}, ErrCancelled
	}
	oriented := source.Clone()
	value, err := oriented.Get(0)
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	if value {
		oriented = oriented.Not()
	}
	transform, err := AigSemiCanonicalize(oriented, s.options.Stop)
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	key := aigCanonicalKey(transform.Table)
	if id, found := s.lookup[key]; found {
		if s.entries[id].Visits == int(^uint(0)>>1) {
			return AigCanonicalRecord{}, ErrLimit
		}
		s.entries[id].Visits++
		return AigCanonicalRecord{EntryID: id, Transform: transform, PrimeInputs: source.variables}, nil
	}
	if len(s.entries) >= s.options.MaxEntries {
		return AigCanonicalRecord{}, ErrLimit
	}
	graph := s.graph.Clone()
	bdc, err := NewBdcManager(BdcParameters{
		MaxVariables: max(2, s.options.MaxVariables), MaxWords: len(transform.Table.words), MaxNodes: s.options.MaxDecompositionNodes,
	})
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	if _, err = bdc.Decompose(transform.Table, nil, nil, BdcDecomposeOptions{MaxNewNodes: s.options.MaxDecompositionNodes, Stop: s.options.Stop}); err != nil {
		return AigCanonicalRecord{}, err
	}
	mapping := make([]AigLit, bdc.NodeCount())
	for id := 0; id < bdc.NodeCount(); id++ {
		if s.options.Stop != nil && s.options.Stop() {
			return AigCanonicalRecord{}, ErrCancelled
		}
		node, nodeErr := bdc.Node(id)
		if nodeErr != nil {
			return AigCanonicalRecord{}, nodeErr
		}
		switch node.Type {
		case BdcNodeConst1:
			mapping[id] = AigConst1
		case BdcNodePI:
			if id < 1 || id > transform.Table.variables {
				return AigCanonicalRecord{}, ErrInvariant
			}
			mapping[id], err = graph.IthVar(id - 1)
		case BdcNodeAND:
			left, right := int(node.Fanin0.ID()), int(node.Fanin1.ID())
			if left >= id || right >= id {
				return AigCanonicalRecord{}, ErrInvariant
			}
			mapping[id], err = graph.And(mapping[left].NotIf(node.Fanin0.Complemented()), mapping[right].NotIf(node.Fanin1.Complemented()))
		default:
			return AigCanonicalRecord{}, ErrInvariant
		}
		if err != nil {
			return AigCanonicalRecord{}, err
		}
		if graph.ObjectCount() > s.options.MaxAigObjects {
			return AigCanonicalRecord{}, ErrLimit
		}
	}
	root, err := bdc.Root()
	if err != nil {
		return AigCanonicalRecord{}, err
	}
	rootID := int(root.ID())
	if rootID < 0 || rootID >= len(mapping) {
		return AigCanonicalRecord{}, ErrInvariant
	}
	if _, err = graph.CreateCO(mapping[rootID].NotIf(root.Complemented())); err != nil {
		return AigCanonicalRecord{}, err
	}
	if graph.ObjectCount() > s.options.MaxAigObjects {
		return AigCanonicalRecord{}, ErrLimit
	}
	if err = graph.Check(); err != nil {
		return AigCanonicalRecord{}, err
	}
	id := len(s.entries)
	s.graph = graph
	s.entries = append(s.entries, AigCanonicalEntry{ID: id, Visits: 1, Variables: transform.Table.variables, Table: transform.Table.Clone()})
	s.lookup[key] = id
	return AigCanonicalRecord{EntryID: id, New: true, Transform: transform, PrimeInputs: source.variables}, nil
}

func (s *AigCanonicalSession) Statistics() AigCanonicalStatistics {
	if s == nil || s.closed {
		return AigCanonicalStatistics{}
	}
	return s.statistics
}

func (s *AigCanonicalSession) Entries() []AigCanonicalEntry {
	if s == nil || s.closed {
		return nil
	}
	result := make([]AigCanonicalEntry, len(s.entries))
	for i, entry := range s.entries {
		result[i] = entry
		result[i].Table = entry.Table.Clone()
	}
	return result
}

func (s *AigCanonicalSession) Graph() *AigManager {
	if s == nil || s.closed {
		return nil
	}
	return s.graph.Clone()
}

// MarshalBinary returns a deterministic, owned table-library encoding. It is
// deliberately Go-owned and does not promise AIGER byte identity.
func (s *AigCanonicalSession) MarshalBinary() ([]byte, error) {
	if s == nil || s.closed {
		return nil, ErrDimension
	}
	var output bytes.Buffer
	output.WriteString("AIGCANON1")
	if err := binary.Write(&output, binary.LittleEndian, uint32(len(s.entries))); err != nil {
		return nil, err
	}
	for _, entry := range s.entries {
		if err := binary.Write(&output, binary.LittleEndian, uint16(entry.Variables)); err != nil {
			return nil, err
		}
		if err := binary.Write(&output, binary.LittleEndian, uint64(entry.Visits)); err != nil {
			return nil, err
		}
		if err := binary.Write(&output, binary.LittleEndian, uint32(len(entry.Table.words))); err != nil {
			return nil, err
		}
		for _, word := range entry.Table.words {
			if err := binary.Write(&output, binary.LittleEndian, word); err != nil {
				return nil, err
			}
		}
	}
	return output.Bytes(), nil
}

func (s *AigCanonicalSession) SHA256() ([32]byte, error) {
	data, err := s.MarshalBinary()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(data), nil
}

func (s *AigCanonicalSession) Close() error {
	if s == nil || s.closed {
		return ErrDimension
	}
	s.closed = true
	s.graph, s.entries, s.lookup = nil, nil, nil
	return nil
}
