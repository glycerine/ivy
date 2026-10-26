package abc

// Binary-BLIF network representation from src/misc/bbl/bblif.c and bblif.h.

import (
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
)

type BblObjectType uint8

const (
	BblObjectNone BblObjectType = iota
	BblObjectCI
	BblObjectCO
	BblObjectNode
	BblObjectVoid
)

type BblLimits struct {
	MaxObjects, MaxObjectID, MaxFanins int
	MaxSOPBytes, MaxBinaryBytes        int
	MaxTruthWords                      int
	Stop                               func() bool
}

type bblObject struct {
	id, originalID int
	kind           BblObjectType
	function       int
	fanins         []int
	declaredFanins int
	marked, box    bool
}

type bblFunction struct {
	sop string
}

type BblManager struct {
	name, nameRaw  string
	objects        []bblObject
	functions      []bblFunction
	idToObject     map[int]int
	functionBySOP  map[string]int
	limits         BblLimits
	sealed, closed bool
}

type BblObject struct {
	ID, OriginalID int
	Type           BblObjectType
	Fanins         []int
	SOP            string
	Marked, Box    bool
	FunctionHandle int
}

type BblStats struct {
	Objects, Nodes, UniqueFunctions       int
	NameBytes, ObjectBytes, FunctionBytes int
}

func normalizeBblLimits(limits BblLimits) (BblLimits, error) {
	if limits.MaxObjects < 0 || limits.MaxObjectID < 0 || limits.MaxFanins < 0 || limits.MaxSOPBytes < 0 || limits.MaxBinaryBytes < 0 || limits.MaxTruthWords < 0 {
		return limits, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if limits.MaxObjectID == 0 {
		limits.MaxObjectID = 1 << 24
	}
	if limits.MaxFanins == 0 {
		limits.MaxFanins = 1 << 20
	}
	if limits.MaxSOPBytes == 0 {
		limits.MaxSOPBytes = 64 << 20
	}
	if limits.MaxBinaryBytes == 0 {
		limits.MaxBinaryBytes = 256 << 20
	}
	if limits.MaxTruthWords == 0 {
		limits.MaxTruthWords = 1 << 20
	}
	return limits, nil
}

func NewBblManager(name string, limits BblLimits) (*BblManager, error) {
	limits, err := normalizeBblLimits(limits)
	if err != nil {
		return nil, err
	}
	if strings.IndexByte(name, 0) >= 0 || len(name) > limits.MaxSOPBytes {
		return nil, ErrDimension
	}
	raw := ""
	if name != "" {
		size := 4 * ((len(name)+1)/4 + 1)
		data := make([]byte, size)
		copy(data, name)
		raw = string(data)
	}
	return &BblManager{name: name, nameRaw: raw, idToObject: make(map[int]int), functionBySOP: make(map[string]int), limits: limits}, nil
}

func (manager *BblManager) checkOpen() error {
	if manager == nil || manager.closed || manager.idToObject == nil || manager.functionBySOP == nil {
		return ErrInvariant
	}
	return nil
}

func validateBblSOP(sop string, variables, maxBytes int) error {
	if variables < 0 || len(sop) == 0 || len(sop) > maxBytes || strings.IndexByte(sop, 0) >= 0 {
		return ErrFormat
	}
	width := variables + 3
	if width <= 0 || len(sop)%width != 0 {
		return ErrFormat
	}
	for offset := 0; offset < len(sop); offset += width {
		for index := 0; index < variables; index++ {
			if strings.IndexByte("01-", sop[offset+index]) < 0 {
				return ErrFormat
			}
		}
		if sop[offset+variables] != ' ' || !strings.ContainsRune("01xn", rune(sop[offset+variables+1])) || sop[offset+variables+2] != '\n' {
			return ErrFormat
		}
	}
	return nil
}

func canonicalBblSOP(sop string, variables int) string {
	width, cubes := variables+3, len(sop)/(variables+3)
	if cubes < 2 || cubes >= 300 {
		return strings.Clone(sop)
	}
	records := make([]string, cubes)
	for index := range records {
		records[index] = sop[index*width : (index+1)*width]
	}
	for index := 0; index < len(records)-1; index++ {
		best := index
		for candidate := index + 1; candidate < len(records); candidate++ {
			if records[candidate][:variables] < records[best][:variables] {
				best = candidate
			}
		}
		records[index], records[best] = records[best], records[index]
	}
	return strings.Join(records, "")
}

func (manager *BblManager) saveSOP(sop string, variables int) int {
	sop = canonicalBblSOP(sop, variables)
	if index, found := manager.functionBySOP[sop]; found {
		return index
	}
	index := len(manager.functions)
	manager.functions = append(manager.functions, bblFunction{sop: sop})
	manager.functionBySOP[sop] = index
	return index
}

func (manager *BblManager) AddObject(kind BblObjectType, id, fanins int, sop string) error {
	if err := manager.checkOpen(); err != nil {
		return err
	}
	if manager.sealed {
		return ErrInvariant
	}
	if manager.limits.Stop != nil && manager.limits.Stop() {
		return ErrCancelled
	}
	if id < 0 || id > manager.limits.MaxObjectID || fanins < 0 || fanins > manager.limits.MaxFanins {
		return ErrDimension
	}
	if len(manager.objects) == manager.limits.MaxObjects {
		return ErrLimit
	}
	if _, exists := manager.idToObject[id]; exists {
		return ErrFormat
	}
	switch kind {
	case BblObjectCI:
		if fanins != 0 || sop != "" {
			return ErrFormat
		}
	case BblObjectCO:
		if fanins != 1 || sop != "" {
			return ErrFormat
		}
	case BblObjectNode:
		if err := validateBblSOP(sop, fanins, manager.limits.MaxSOPBytes); err != nil {
			return err
		}
	default:
		return ErrDimension
	}
	function := -1
	extraFunctionBytes := 0
	if kind == BblObjectNode {
		canonical := canonicalBblSOP(sop, fanins)
		if _, found := manager.functionBySOP[canonical]; !found {
			extraFunctionBytes = 4 + 4*((len(canonical)+1+3)/4)
		}
	}
	_, objectBytes := bblObjectOffsets(manager.objects)
	_, functionBytes := bblFunctionOffsets(manager.functions)
	if functionBytes+extraFunctionBytes > manager.limits.MaxSOPBytes || 12+len(manager.nameRaw)+objectBytes+12+4*fanins+functionBytes+extraFunctionBytes > manager.limits.MaxBinaryBytes {
		return ErrLimit
	}
	if kind == BblObjectNode {
		function = manager.saveSOP(sop, fanins)
	}
	index := len(manager.objects)
	manager.objects = append(manager.objects, bblObject{id: id, originalID: id, kind: kind, function: function, declaredFanins: fanins})
	manager.idToObject[id] = index
	return nil
}

func (manager *BblManager) AddFanin(objectID, faninID int) error {
	if err := manager.checkOpen(); err != nil {
		return err
	}
	if manager.sealed {
		return ErrInvariant
	}
	if manager.limits.Stop != nil && manager.limits.Stop() {
		return ErrCancelled
	}
	objectIndex, objectFound := manager.idToObject[objectID]
	faninIndex, faninFound := manager.idToObject[faninID]
	if !objectFound || !faninFound {
		return ErrFormat
	}
	object := &manager.objects[objectIndex]
	if object.kind == BblObjectCI || manager.objects[faninIndex].kind == BblObjectCO || len(object.fanins) == object.declaredFanins {
		return ErrFormat
	}
	object.fanins = append(object.fanins, faninIndex)
	return nil
}

func (manager *BblManager) Check() error {
	if err := manager.checkOpen(); err != nil {
		return err
	}
	if len(manager.objects) > manager.limits.MaxObjects {
		return ErrInvariant
	}
	for index := range manager.objects {
		if manager.limits.Stop != nil && manager.limits.Stop() {
			return ErrCancelled
		}
		object := &manager.objects[index]
		if !slices.Contains([]BblObjectType{BblObjectCI, BblObjectCO, BblObjectNode}, object.kind) || object.id < 0 || object.id > manager.limits.MaxObjectID || len(object.fanins) != object.declaredFanins {
			return ErrInvariant
		}
		if object.kind == BblObjectCI && (len(object.fanins) != 0 || object.function != -1) || object.kind == BblObjectCO && (len(object.fanins) != 1 || object.function != -1) {
			return ErrInvariant
		}
		if object.kind == BblObjectNode {
			if object.function < 0 || object.function >= len(manager.functions) || validateBblSOP(manager.functions[object.function].sop, len(object.fanins), manager.limits.MaxSOPBytes) != nil {
				return ErrInvariant
			}
		}
		for _, fanin := range object.fanins {
			if fanin < 0 || fanin >= len(manager.objects) || manager.objects[fanin].kind == BblObjectCO {
				return ErrInvariant
			}
		}
	}
	return nil
}

func bblFunctionOffsets(functions []bblFunction) ([]int, int) {
	offsets := make([]int, len(functions))
	size := 0
	for index, function := range functions {
		offsets[index] = size
		words := (len(function.sop) + 1 + 3) / 4
		size += 4 + words*4
	}
	return offsets, size
}

func bblObjectOffsets(objects []bblObject) ([]int, int) {
	offsets := make([]int, len(objects))
	size := 0
	for index, object := range objects {
		offsets[index] = size
		size += 12 + 4*len(object.fanins)
	}
	return offsets, size
}

func (manager *BblManager) object(index int) (BblObject, error) {
	if err := manager.checkOpen(); err != nil {
		return BblObject{}, err
	}
	if index < 0 || index >= len(manager.objects) {
		return BblObject{}, ErrDimension
	}
	object := manager.objects[index]
	result := BblObject{ID: object.id, OriginalID: object.originalID, Type: object.kind, Marked: object.marked, Box: object.box, FunctionHandle: -1}
	result.Fanins = make([]int, len(object.fanins))
	for i, fanin := range object.fanins {
		result.Fanins[i] = manager.objects[fanin].id
	}
	if object.function >= 0 {
		result.SOP = strings.Clone(manager.functions[object.function].sop)
		offsets, _ := bblFunctionOffsets(manager.functions)
		result.FunctionHandle = offsets[object.function]
	}
	return result, nil
}

func (manager *BblManager) Object(index int) (BblObject, error) { return manager.object(index) }

func (manager *BblManager) ObjectByID(id int) (BblObject, bool, error) {
	if err := manager.checkOpen(); err != nil {
		return BblObject{}, false, err
	}
	index, found := manager.idToObject[id]
	if !found {
		return BblObject{}, false, nil
	}
	object, err := manager.object(index)
	return object, err == nil, err
}

func (manager *BblManager) Objects() ([]BblObject, error) {
	if err := manager.checkOpen(); err != nil {
		return nil, err
	}
	result := make([]BblObject, len(manager.objects))
	for index := range result {
		if manager.limits.Stop != nil && manager.limits.Stop() {
			return nil, ErrCancelled
		}
		object, err := manager.object(index)
		if err != nil {
			return nil, err
		}
		result[index] = object
	}
	return result, nil
}

func (manager *BblManager) Name() (string, error) {
	if err := manager.checkOpen(); err != nil {
		return "", err
	}
	return strings.Clone(manager.name), nil
}

func (manager *BblManager) FunctionBytes() (int, error) {
	if err := manager.checkOpen(); err != nil {
		return 0, err
	}
	_, size := bblFunctionOffsets(manager.functions)
	return size, nil
}

func (manager *BblManager) Mark(index int) error {
	if err := manager.checkOpen(); err != nil {
		return err
	}
	if index < 0 || index >= len(manager.objects) {
		return ErrDimension
	}
	manager.objects[index].marked = true
	return nil
}

func (manager *BblManager) Stats() (BblStats, error) {
	if err := manager.checkOpen(); err != nil {
		return BblStats{}, err
	}
	_, objectBytes := bblObjectOffsets(manager.objects)
	_, functionBytes := bblFunctionOffsets(manager.functions)
	stats := BblStats{Objects: len(manager.objects), UniqueFunctions: len(manager.functions), NameBytes: len(manager.nameRaw), ObjectBytes: objectBytes, FunctionBytes: functionBytes}
	for _, object := range manager.objects {
		if object.kind == BblObjectNode {
			stats.Nodes++
		}
	}
	return stats, nil
}

func (manager *BblManager) FormatStats() (string, error) {
	stats, err := manager.Stats()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Total objects = %7d.  Total nodes = %7d. Unique functions = %7d.\nName manager = %5.2f MB\nObjs manager = %5.2f MB\nFncs manager = %5.2f MB\n", stats.Objects, stats.Nodes, stats.UniqueFunctions, float64(stats.NameBytes)/(1<<20), float64(stats.ObjectBytes)/(1<<20), float64(stats.FunctionBytes)/(1<<20)), nil
}

func (manager *BblManager) FormatBLIF() (string, error) {
	if err := manager.Check(); err != nil {
		return "", err
	}
	var output strings.Builder
	output.WriteString("# Test file written by Bbl_ManDumpBlif() in ABC.\n.model ")
	output.WriteString(manager.name)
	output.WriteByte('\n')
	for index, object := range manager.objects {
		switch object.kind {
		case BblObjectCI:
			fmt.Fprintf(&output, ".inputs %d\n", object.id)
		case BblObjectCO:
			fmt.Fprintf(&output, ".outputs %d\n", object.id)
		case BblObjectNode:
			output.WriteString(".names")
			for _, fanin := range object.fanins {
				fmt.Fprintf(&output, " %d", manager.objects[fanin].id)
			}
			fmt.Fprintf(&output, " %d\n%s", manager.objects[index].id, manager.functions[object.function].sop)
		}
	}
	for _, object := range manager.objects {
		if object.kind != BblObjectCO {
			continue
		}
		output.WriteString(".names")
		for _, fanin := range object.fanins {
			fmt.Fprintf(&output, " %d", manager.objects[fanin].id)
		}
		fmt.Fprintf(&output, " %d\n1 1\n", object.id)
	}
	output.WriteString(".end\n")
	return output.String(), nil
}

func appendBblVector(destination, data []byte) []byte {
	var length [4]byte
	binary.LittleEndian.PutUint32(length[:], uint32(len(data)))
	destination = append(destination, length[:]...)
	return append(destination, data...)
}

func (manager *BblManager) MarshalBinary() ([]byte, error) {
	if err := manager.Check(); err != nil {
		return nil, err
	}
	objectOffsets, objectBytes := bblObjectOffsets(manager.objects)
	functionOffsets, functionBytes := bblFunctionOffsets(manager.functions)
	total := 12 + len(manager.nameRaw) + objectBytes + functionBytes
	if total > manager.limits.MaxBinaryBytes {
		return nil, ErrLimit
	}
	objects := make([]byte, objectBytes)
	for index, object := range manager.objects {
		offset := objectOffsets[index]
		binary.LittleEndian.PutUint32(objects[offset:], uint32(int32(object.id)))
		functionHandle := int32(-1)
		if object.function >= 0 {
			functionHandle = int32(functionOffsets[object.function])
		}
		binary.LittleEndian.PutUint32(objects[offset+4:], uint32(functionHandle))
		flags := uint32(len(object.fanins)) << 4
		if object.kind == BblObjectCI {
			flags |= 1
		}
		if object.kind == BblObjectCO {
			flags |= 2
		}
		if object.box {
			flags |= 4
		}
		if object.marked {
			flags |= 8
		}
		binary.LittleEndian.PutUint32(objects[offset+8:], flags)
		for faninIndex, fanin := range object.fanins {
			relative := int32(offset - objectOffsets[fanin])
			binary.LittleEndian.PutUint32(objects[offset+12+4*faninIndex:], uint32(relative))
		}
	}
	functions := make([]byte, functionBytes)
	for index, function := range manager.functions {
		offset := functionOffsets[index]
		words := (len(function.sop) + 1 + 3) / 4
		binary.LittleEndian.PutUint32(functions[offset:], uint32(words))
		copy(functions[offset+4:], function.sop)
	}
	result := make([]byte, 0, total)
	result = appendBblVector(result, []byte(manager.nameRaw))
	result = appendBblVector(result, objects)
	result = appendBblVector(result, functions)
	return result, nil
}

func (manager *BblManager) Close() error {
	if err := manager.checkOpen(); err != nil {
		return err
	}
	manager.closed = true
	manager.objects, manager.functions = nil, nil
	manager.idToObject, manager.functionBySOP = nil, nil
	manager.name, manager.nameRaw = "", ""
	return nil
}
