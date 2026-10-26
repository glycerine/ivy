package abc

// Concrete manager, cache, table, reference, and API foundations from
// src/bdd/cudd/cudd.h, cuddInt.h, cuddAPI.c, cuddInit.c, cuddRef.c,
// cuddCache.c, cuddLCache.c, cuddTable.c, cuddCheck.c, and cuddUtil.c.

import (
	"math"
	"strconv"
)

type CuddCacheDomain uint8

const (
	CuddCacheBDD CuddCacheDomain = iota
	CuddCacheADD
)

type CuddVariableRole uint8

const (
	CuddVariablePI CuddVariableRole = iota
	CuddVariablePresentState
	CuddVariableNextState
	// CuddVariableOrdinary is retained as a descriptive alias for CUDD's
	// default primary-input role.
	CuddVariableOrdinary = CuddVariablePI
)

type CuddGroupFlag uint8

const (
	CuddGroupNone CuddGroupFlag = iota
	CuddGroupSoft
	CuddGroupHard
	CuddGroupUngrouped
)

type CuddFoundationOptions struct {
	Variables     int
	BDDCapacity   int
	ADDCapacity   int
	CacheSlots    int
	MaxOperations int
	MaxBytes      int
	Stop          func() bool
}

type CuddFoundationVariable struct {
	Level     int
	Role      CuddVariableRole
	Bound     bool
	PairIndex int
	Group     CuddGroupFlag
}

type CuddCacheKey struct {
	Domain    CuddCacheDomain
	Operation uint16
	Arity     uint8
	Operands  [3]uint32
}

type CuddCacheRecord struct {
	Key    CuddCacheKey
	Result uint32
	Valid  bool
}

type CuddFoundationStatistics struct {
	Variables          int
	BDDNodes, ADDNodes int
	BDDExternalRefs    uint64
	ADDExternalRefs    uint64
	CacheSlots         int
	CacheUsed          int
	CacheLookups       uint64
	CacheHits          uint64
	RecursiveCalls     uint64
	MemoryBytes        uint64
	Background         float64
	Epsilon            float64
}

type CuddFoundationSnapshot struct {
	Variables                       []CuddFoundationVariable
	Permutation, InversePermutation []int
	Cache                           []CuddCacheRecord
	Background                      CuddADDEdge
	Epsilon                         float64
}

type CuddFoundation struct {
	bdd                                     *ExtraBDDManager
	add                                     *CuddADDManager
	options                                 CuddFoundationOptions
	variables                               []CuddFoundationVariable
	permutation, inversePermutation         []int
	bddReferences                           map[uint32]uint32
	addReferences                           map[uint32]uint32
	cache                                   []CuddCacheRecord
	cacheLookups, cacheHits, recursiveCalls uint64
	background                              CuddADDEdge
	epsilon                                 float64
	closed                                  bool
}

func normalizeCuddFoundationOptions(options CuddFoundationOptions) (CuddFoundationOptions, error) {
	if options.Variables < 0 || options.Variables > 30 || options.BDDCapacity < 0 || options.ADDCapacity < 0 || options.CacheSlots < 0 || options.MaxOperations < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.BDDCapacity == 0 {
		options.BDDCapacity = max(options.Variables+32, 256)
	}
	if options.ADDCapacity == 0 {
		options.ADDCapacity = max(options.Variables+32, 256)
	}
	if options.BDDCapacity < options.Variables+1 || options.ADDCapacity < options.Variables+2 {
		return options, ErrDimension
	}
	if options.CacheSlots == 0 {
		options.CacheSlots = 256
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 26
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	slots := 1
	for slots < options.CacheSlots {
		if slots > 1<<23 {
			return options, ErrLimit
		}
		slots <<= 1
	}
	options.CacheSlots = slots
	// Account for the bounded node and table capacities before any
	// allocation. This is deliberately conservative and architecture-neutral.
	estimate := uint64(options.BDDCapacity)*24 + uint64(options.ADDCapacity)*40 + uint64(options.CacheSlots)*40 + uint64(options.Variables)*48
	if estimate > uint64(options.MaxBytes) {
		return options, ErrLimit
	}
	return options, nil
}

func NewCuddFoundation(options CuddFoundationOptions) (*CuddFoundation, error) {
	options, err := normalizeCuddFoundationOptions(options)
	if err != nil {
		return nil, err
	}
	bdd, err := NewExtraBDDManager(options.Variables, options.BDDCapacity)
	if err != nil {
		return nil, err
	}
	add, err := NewCuddADDManager(options.Variables, options.ADDCapacity)
	if err != nil {
		_, _ = bdd.Close()
		return nil, err
	}
	foundation := &CuddFoundation{
		bdd: bdd, add: add, options: options,
		variables:   make([]CuddFoundationVariable, options.Variables),
		permutation: make([]int, options.Variables), inversePermutation: make([]int, options.Variables),
		bddReferences: make(map[uint32]uint32), addReferences: make(map[uint32]uint32),
		cache: make([]CuddCacheRecord, options.CacheSlots), epsilon: 1e-12,
	}
	for index := range foundation.variables {
		foundation.variables[index] = CuddFoundationVariable{Level: index, PairIndex: -1}
		foundation.permutation[index], foundation.inversePermutation[index] = index, index
	}
	if err := foundation.Validate(); err != nil {
		_, _ = bdd.Close()
		_ = add.Close()
		return nil, err
	}
	return foundation, nil
}

func (foundation *CuddFoundation) poll() error {
	if foundation == nil || foundation.closed {
		return ErrInvariant
	}
	if foundation.recursiveCalls >= uint64(foundation.options.MaxOperations) {
		return ErrLimit
	}
	foundation.recursiveCalls++
	if foundation.options.Stop != nil && foundation.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (foundation *CuddFoundation) BDDManager() (*ExtraBDDManager, error) {
	if err := foundation.poll(); err != nil {
		return nil, err
	}
	return foundation.bdd, nil
}

func (foundation *CuddFoundation) ADDManager() (*CuddADDManager, error) {
	if err := foundation.poll(); err != nil {
		return nil, err
	}
	return foundation.add, nil
}

func (foundation *CuddFoundation) BDDVariable(index int) (ExtraBDDLit, error) {
	if err := foundation.poll(); err != nil {
		return 0, err
	}
	return foundation.bdd.Variable(index)
}

func (foundation *CuddFoundation) ADDVariable(index int) (CuddADDEdge, error) {
	if err := foundation.poll(); err != nil {
		return 0, err
	}
	return foundation.add.Variable(index)
}

func (foundation *CuddFoundation) ADDConstant(value float64) (CuddADDEdge, error) {
	if err := foundation.poll(); err != nil {
		return 0, err
	}
	return foundation.add.Constant(value)
}

func (foundation *CuddFoundation) validateVariable(index int) error {
	if index < 0 || index >= len(foundation.variables) {
		return ErrDimension
	}
	return nil
}

func (foundation *CuddFoundation) Variable(index int) (CuddFoundationVariable, error) {
	if err := foundation.poll(); err != nil {
		return CuddFoundationVariable{}, err
	}
	if err := foundation.validateVariable(index); err != nil {
		return CuddFoundationVariable{}, err
	}
	return foundation.variables[index], nil
}

func (foundation *CuddFoundation) SetVariable(index int, value CuddFoundationVariable) error {
	if err := foundation.poll(); err != nil {
		return err
	}
	if err := foundation.validateVariable(index); err != nil {
		return err
	}
	if value.Level != foundation.permutation[index] || value.Role > CuddVariableNextState || value.Group > CuddGroupUngrouped || value.PairIndex < -1 || value.PairIndex >= len(foundation.variables) || value.PairIndex == index {
		return ErrDimension
	}
	foundation.variables[index] = value
	return nil
}

func (foundation *CuddFoundation) SetPermutation(permutation []int) error {
	if err := foundation.poll(); err != nil {
		return err
	}
	if len(permutation) != len(foundation.variables) {
		return ErrDimension
	}
	inverse := make([]int, len(permutation))
	seen := make([]bool, len(permutation))
	for variable, level := range permutation {
		if level < 0 || level >= len(permutation) || seen[level] {
			return ErrDimension
		}
		seen[level], inverse[level] = true, variable
	}
	foundation.permutation = append(foundation.permutation[:0], permutation...)
	foundation.inversePermutation = inverse
	for variable, level := range permutation {
		foundation.variables[variable].Level = level
	}
	return nil
}

func (foundation *CuddFoundation) Permutation() ([]int, []int, error) {
	if err := foundation.poll(); err != nil {
		return nil, nil, err
	}
	return append([]int(nil), foundation.permutation...), append([]int(nil), foundation.inversePermutation...), nil
}

func (foundation *CuddFoundation) SetEpsilon(epsilon float64) error {
	if err := foundation.poll(); err != nil {
		return err
	}
	if math.IsNaN(epsilon) || math.IsInf(epsilon, 0) || epsilon < 0 {
		return ErrDimension
	}
	foundation.epsilon = epsilon
	return nil
}

func (foundation *CuddFoundation) SetBackground(value float64) (CuddADDEdge, error) {
	if err := foundation.poll(); err != nil {
		return 0, err
	}
	edge, err := foundation.add.Constant(value)
	if err != nil {
		return 0, err
	}
	foundation.background = edge
	return edge, nil
}

func (foundation *CuddFoundation) ref(domain CuddCacheDomain, operand uint32, increment bool) error {
	if err := foundation.poll(); err != nil {
		return err
	}
	if err := foundation.validateOperand(domain, operand); err != nil {
		return err
	}
	key := operand
	if domain == CuddCacheBDD {
		key = uint32(ExtraBDDLit(operand).ID())
	}
	references := foundation.bddReferences
	if domain == CuddCacheADD {
		references = foundation.addReferences
	}
	current := references[key]
	if increment {
		if current == math.MaxUint32 {
			return ErrLimit
		}
		references[key] = current + 1
		return nil
	}
	if current == 0 {
		return ErrInvariant
	}
	if current == 1 {
		delete(references, key)
	} else {
		references[key] = current - 1
	}
	return nil
}

func (foundation *CuddFoundation) RefBDD(literal ExtraBDDLit) error {
	return foundation.ref(CuddCacheBDD, uint32(literal), true)
}
func (foundation *CuddFoundation) DerefBDD(literal ExtraBDDLit) error {
	return foundation.ref(CuddCacheBDD, uint32(literal), false)
}
func (foundation *CuddFoundation) RefADD(edge CuddADDEdge) error {
	return foundation.ref(CuddCacheADD, uint32(edge), true)
}
func (foundation *CuddFoundation) DerefADD(edge CuddADDEdge) error {
	return foundation.ref(CuddCacheADD, uint32(edge), false)
}

func (foundation *CuddFoundation) ReferenceCount(domain CuddCacheDomain, operand uint32) (uint32, error) {
	if err := foundation.poll(); err != nil {
		return 0, err
	}
	if err := foundation.validateOperand(domain, operand); err != nil {
		return 0, err
	}
	if domain == CuddCacheBDD {
		return foundation.bddReferences[uint32(ExtraBDDLit(operand).ID())], nil
	}
	return foundation.addReferences[operand], nil
}

func (foundation *CuddFoundation) CheckZeroReferences() (int, error) {
	if err := foundation.poll(); err != nil {
		return 0, err
	}
	return len(foundation.bddReferences) + len(foundation.addReferences), nil
}

func (foundation *CuddFoundation) validateOperand(domain CuddCacheDomain, operand uint32) error {
	switch domain {
	case CuddCacheBDD:
		if foundation.bdd == nil || !foundation.bdd.valid(ExtraBDDLit(operand)) {
			return ErrInvalidLiteral
		}
	case CuddCacheADD:
		if foundation.add == nil || !foundation.add.valid(CuddADDEdge(operand)) {
			return ErrInvalidLiteral
		}
	default:
		return ErrDimension
	}
	return nil
}

func validateCuddCacheKey(key CuddCacheKey) error {
	if key.Domain > CuddCacheADD || key.Arity == 0 || key.Arity > 3 {
		return ErrDimension
	}
	for index := int(key.Arity); index < len(key.Operands); index++ {
		if key.Operands[index] != 0 {
			return ErrDimension
		}
	}
	return nil
}

func cuddCacheHash(key CuddCacheKey) uint32 {
	hash := uint32(key.Domain)*0x9e3779b1 ^ uint32(key.Operation)*0x85ebca6b ^ uint32(key.Arity)*0xc2b2ae35
	for index := 0; index < int(key.Arity); index++ {
		hash = (hash ^ key.Operands[index]*0x27d4eb2d) * 0x165667b1
	}
	return hash
}

func (foundation *CuddFoundation) validateCacheRecord(record CuddCacheRecord) error {
	if !record.Valid {
		return nil
	}
	if err := validateCuddCacheKey(record.Key); err != nil {
		return err
	}
	for index := 0; index < int(record.Key.Arity); index++ {
		if err := foundation.validateOperand(record.Key.Domain, record.Key.Operands[index]); err != nil {
			return err
		}
	}
	return foundation.validateOperand(record.Key.Domain, record.Result)
}

func (foundation *CuddFoundation) CacheLookup(key CuddCacheKey) (uint32, bool, error) {
	if err := foundation.poll(); err != nil {
		return 0, false, err
	}
	if err := validateCuddCacheKey(key); err != nil {
		return 0, false, err
	}
	for index := 0; index < int(key.Arity); index++ {
		if err := foundation.validateOperand(key.Domain, key.Operands[index]); err != nil {
			return 0, false, err
		}
	}
	foundation.cacheLookups++
	record := foundation.cache[int(cuddCacheHash(key))&(len(foundation.cache)-1)]
	if record.Valid && record.Key == key {
		foundation.cacheHits++
		return record.Result, true, nil
	}
	return 0, false, nil
}

func (foundation *CuddFoundation) CacheInsert(key CuddCacheKey, result uint32) error {
	if err := foundation.poll(); err != nil {
		return err
	}
	record := CuddCacheRecord{Key: key, Result: result, Valid: true}
	if err := foundation.validateCacheRecord(record); err != nil {
		return err
	}
	foundation.cache[int(cuddCacheHash(key))&(len(foundation.cache)-1)] = record
	return nil
}

func (foundation *CuddFoundation) CacheFlush() error {
	if err := foundation.poll(); err != nil {
		return err
	}
	clear(foundation.cache)
	return nil
}

func (foundation *CuddFoundation) Snapshot() (CuddFoundationSnapshot, error) {
	if err := foundation.poll(); err != nil {
		return CuddFoundationSnapshot{}, err
	}
	return CuddFoundationSnapshot{
		Variables:          append([]CuddFoundationVariable(nil), foundation.variables...),
		Permutation:        append([]int(nil), foundation.permutation...),
		InversePermutation: append([]int(nil), foundation.inversePermutation...),
		Cache:              append([]CuddCacheRecord(nil), foundation.cache...), Background: foundation.background, Epsilon: foundation.epsilon,
	}, nil
}

func (foundation *CuddFoundation) validateBDDTable() error {
	if err := foundation.bdd.check(); err != nil {
		return err
	}
	seen := make([]bool, len(foundation.bdd.nodes))
	seen[0] = true
	for slot, head := range foundation.bdd.uniqueHeads {
		steps := 0
		for current := head; current != 0; current = foundation.bdd.next[current] {
			if int(current) >= len(foundation.bdd.nodes) || seen[current] || steps >= len(foundation.bdd.nodes) {
				return ErrInvariant
			}
			node := foundation.bdd.nodes[current]
			expected := int(extraBDDHash(ExtraBDDLit(node.variable), node.then, node.otherwise) & uint32(len(foundation.bdd.uniqueHeads)-1))
			if expected != slot {
				return ErrInvariant
			}
			seen[current], steps = true, steps+1
		}
	}
	for index := 1; index < len(seen); index++ {
		if !seen[index] {
			return ErrInvariant
		}
	}
	for _, entry := range foundation.bdd.cache {
		if entry.valid && (!foundation.bdd.valid(entry.first) || !foundation.bdd.valid(entry.second) || !foundation.bdd.valid(entry.result)) {
			return ErrInvariant
		}
	}
	return nil
}

func (foundation *CuddFoundation) validateADDTable() error {
	if err := foundation.add.check(); err != nil {
		return err
	}
	if len(foundation.add.terminals)+len(foundation.add.unique) != len(foundation.add.nodes) {
		return ErrInvariant
	}
	for bits, edge := range foundation.add.terminals {
		if !foundation.add.valid(edge) {
			return ErrInvariant
		}
		node := foundation.add.nodes[edge]
		valueBits, err := cuddADDValueBits(node.Value)
		if err != nil || !node.Terminal || valueBits != bits {
			return ErrInvariant
		}
	}
	for key, edge := range foundation.add.unique {
		if !foundation.add.valid(edge) {
			return ErrInvariant
		}
		node := foundation.add.nodes[edge]
		if node.Terminal || node.Variable != key.variable || node.Then != key.high || node.Else != key.low {
			return ErrInvariant
		}
	}
	return nil
}

func (foundation *CuddFoundation) Validate() error {
	if foundation == nil || foundation.closed || foundation.bdd == nil || foundation.add == nil || len(foundation.variables) != foundation.options.Variables || len(foundation.permutation) != len(foundation.variables) || len(foundation.inversePermutation) != len(foundation.variables) || len(foundation.cache) == 0 || len(foundation.cache)&(len(foundation.cache)-1) != 0 || math.IsNaN(foundation.epsilon) || math.IsInf(foundation.epsilon, 0) || foundation.epsilon < 0 {
		return ErrInvariant
	}
	if err := foundation.validateBDDTable(); err != nil {
		return err
	}
	if err := foundation.validateADDTable(); err != nil {
		return err
	}
	if !foundation.add.valid(foundation.background) {
		return ErrInvariant
	}
	for variable, metadata := range foundation.variables {
		level := foundation.permutation[variable]
		if level < 0 || level >= len(foundation.variables) || foundation.inversePermutation[level] != variable || metadata.Level != level || metadata.Role > CuddVariableNextState || metadata.Group > CuddGroupUngrouped || metadata.PairIndex < -1 || metadata.PairIndex >= len(foundation.variables) || metadata.PairIndex == variable {
			return ErrInvariant
		}
	}
	for node, count := range foundation.bddReferences {
		if count == 0 || int(node) >= len(foundation.bdd.nodes) {
			return ErrInvariant
		}
	}
	for node, count := range foundation.addReferences {
		if count == 0 || int(node) >= len(foundation.add.nodes) {
			return ErrInvariant
		}
	}
	for _, record := range foundation.cache {
		if err := foundation.validateCacheRecord(record); err != nil {
			return ErrInvariant
		}
	}
	return nil
}

func (foundation *CuddFoundation) Statistics() (CuddFoundationStatistics, error) {
	if err := foundation.Validate(); err != nil {
		return CuddFoundationStatistics{}, err
	}
	bddStats, err := foundation.bdd.Statistics()
	if err != nil {
		return CuddFoundationStatistics{}, err
	}
	var bddRefs, addRefs uint64
	for _, count := range foundation.bddReferences {
		bddRefs += uint64(count)
	}
	for _, count := range foundation.addReferences {
		addRefs += uint64(count)
	}
	used := 0
	for _, record := range foundation.cache {
		if record.Valid {
			used++
		}
	}
	memory := bddStats.MemoryBytes + uint64(len(foundation.add.nodes))*40 + uint64(len(foundation.cache))*40 + uint64(len(foundation.variables))*48
	background := foundation.add.nodes[foundation.background].Value
	return CuddFoundationStatistics{Variables: len(foundation.variables), BDDNodes: len(foundation.bdd.nodes), ADDNodes: len(foundation.add.nodes), BDDExternalRefs: bddRefs, ADDExternalRefs: addRefs, CacheSlots: len(foundation.cache), CacheUsed: used, CacheLookups: foundation.cacheLookups, CacheHits: foundation.cacheHits, RecursiveCalls: foundation.recursiveCalls, MemoryBytes: memory, Background: background, Epsilon: foundation.epsilon}, nil
}

func (statistics CuddFoundationStatistics) Format() string {
	return "CUDD stats: vars=" + strconv.Itoa(statistics.Variables) + " bdd=" + strconv.Itoa(statistics.BDDNodes) + " add=" + strconv.Itoa(statistics.ADDNodes) + " refs=" + strconv.FormatUint(statistics.BDDExternalRefs+statistics.ADDExternalRefs, 10) + " cache=" + strconv.Itoa(statistics.CacheUsed) + "/" + strconv.Itoa(statistics.CacheSlots) + " hits=" + strconv.FormatUint(statistics.CacheHits, 10) + " lookups=" + strconv.FormatUint(statistics.CacheLookups, 10) + " bytes=" + strconv.FormatUint(statistics.MemoryBytes, 10) + "\n"
}

func (foundation *CuddFoundation) Close() (CuddFoundationStatistics, error) {
	statistics, err := foundation.Statistics()
	if err != nil {
		return CuddFoundationStatistics{}, err
	}
	_, err = foundation.bdd.Close()
	if err != nil {
		return CuddFoundationStatistics{}, err
	}
	if err = foundation.add.Close(); err != nil {
		return CuddFoundationStatistics{}, err
	}
	foundation.closed = true
	foundation.variables, foundation.permutation, foundation.inversePermutation, foundation.cache = nil, nil, nil, nil
	foundation.bddReferences, foundation.addReferences = nil, nil
	return statistics, nil
}

var cuddFoundationSources = []string{
	"cuddAPI.c", "cuddAddAbs.c", "cuddAddApply.c", "cuddAddFind.c", "cuddAddInv.c", "cuddAddIte.c", "cuddAddNeg.c", "cuddAddWalsh.c", "cuddAndAbs.c", "cuddAnneal.c", "cuddApa.c", "cuddApprox.c", "cuddBddAbs.c", "cuddBddCorr.c", "cuddBddIte.c", "cuddBridge.c", "cuddCache.c", "cuddCheck.c", "cuddClip.c", "cuddCof.c", "cuddCompose.c", "cuddDecomp.c", "cuddEssent.c", "cuddExact.c", "cuddExport.c", "cuddGenCof.c", "cuddGenetic.c", "cuddGroup.c", "cuddHarwell.c", "cuddInit.c", "cuddInteract.c", "cuddLCache.c", "cuddLevelQ.c", "cuddLinear.c", "cuddLiteral.c", "cuddMatMult.c", "cuddPriority.c", "cuddRead.c", "cuddRef.c", "cuddReorder.c", "cuddSat.c", "cuddSign.c", "cuddSolve.c", "cuddSplit.c", "cuddSubsetHB.c", "cuddSubsetSP.c", "cuddSymmetry.c", "cuddTable.c", "cuddUtil.c", "cuddWindow.c", "cuddZddCount.c", "cuddZddFuncs.c", "cuddZddGroup.c", "cuddZddIsop.c", "cuddZddLin.c", "cuddZddMisc.c", "cuddZddPort.c", "cuddZddReord.c", "cuddZddSetop.c", "cuddZddSymm.c", "cuddZddUtil.c",
}

func CuddFoundationBuildSources() []string { return append([]string(nil), cuddFoundationSources...) }
