package abc

// Owned persistence, exact-class summaries and dumps replacing the file and
// stdout helpers in opt/dau/dauNpn.c.

import (
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/glycerine/insdict"
)

const dauFunctionTableHeaderSize = 12
const dauFunctionTableRecordSize = 16

type DauFunctionTableDecodeOptions struct {
	MaxClasses int
}

func dauTruthMask(inputs int) uint64 {
	if inputs == 6 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(1<<uint(inputs)) - 1
}

func validateDauFunctionClasses(inputs int, classes []DauFunctionClass, requireOrder bool) error {
	if inputs < 1 || inputs > 6 {
		return ErrDimension
	}
	seen := insdict.NewDict[uint64, bool]()
	for index, class := range classes {
		if class.Nodes < 0 || class.Nodes > 0x7fffffff || class.Support < 0 || class.Support > inputs || class.Truth&^dauTruthMask(inputs) != 0 {
			return fmt.Errorf("%w: DAU class %d", ErrFormat, index)
		}
		if len(dauLogicalSupport(class.Truth, inputs)) != class.Support {
			return fmt.Errorf("%w: DAU support %d", ErrFormat, index)
		}
		if requireOrder && index > 0 && class.Nodes < classes[index-1].Nodes {
			return fmt.Errorf("%w: DAU node order %d", ErrFormat, index)
		}
		if _, duplicate := seen.Get2(class.Truth); duplicate {
			return fmt.Errorf("%w: duplicate DAU truth %d", ErrFormat, index)
		}
		seen.Set(class.Truth, true)
	}
	return nil
}

// EncodeDauFunctionTable returns a versioned, little-endian, self-describing
// replacement for Dau_TablesSave's host-dependent concatenated arrays.
func EncodeDauFunctionTable(inputs int, classes []DauFunctionClass) ([]byte, error) {
	if err := validateDauFunctionClasses(inputs, classes, true); err != nil {
		return nil, err
	}
	if uint64(len(classes)) > uint64((int(^uint(0)>>1)-dauFunctionTableHeaderSize)/dauFunctionTableRecordSize) {
		return nil, ErrLimit
	}
	data := make([]byte, dauFunctionTableHeaderSize+len(classes)*dauFunctionTableRecordSize)
	copy(data, "DNP1")
	data[4] = byte(inputs)
	binary.LittleEndian.PutUint32(data[8:], uint32(len(classes)))
	for index, class := range classes {
		offset := dauFunctionTableHeaderSize + index*dauFunctionTableRecordSize
		binary.LittleEndian.PutUint64(data[offset:], class.Truth)
		binary.LittleEndian.PutUint32(data[offset+8:], uint32(class.Nodes))
		data[offset+12] = byte(class.Support)
	}
	return data, nil
}

// DecodeDauFunctionTable validates and owns all decoded records. No partial
// table is returned for malformed or over-budget input.
func DecodeDauFunctionTable(data []byte, options DauFunctionTableDecodeOptions) (int, []DauFunctionClass, error) {
	if options.MaxClasses < 0 {
		return 0, nil, ErrDimension
	}
	if len(data) < dauFunctionTableHeaderSize || string(data[:4]) != "DNP1" || data[5] != 0 || data[6] != 0 || data[7] != 0 {
		return 0, nil, ErrFormat
	}
	inputs := int(data[4])
	if inputs < 1 || inputs > 6 {
		return 0, nil, ErrFormat
	}
	count := uint64(binary.LittleEndian.Uint32(data[8:]))
	limit := options.MaxClasses
	if limit == 0 {
		limit = 1 << 20
	}
	if count > uint64(limit) || count > uint64((int(^uint(0)>>1)-dauFunctionTableHeaderSize)/dauFunctionTableRecordSize) {
		return 0, nil, ErrLimit
	}
	want := dauFunctionTableHeaderSize + int(count)*dauFunctionTableRecordSize
	if len(data) != want {
		return 0, nil, ErrFormat
	}
	classes := make([]DauFunctionClass, int(count))
	for index := range classes {
		offset := dauFunctionTableHeaderSize + index*dauFunctionTableRecordSize
		if data[offset+13] != 0 || data[offset+14] != 0 || data[offset+15] != 0 {
			return 0, nil, ErrFormat
		}
		classes[index] = DauFunctionClass{
			Truth:   binary.LittleEndian.Uint64(data[offset:]),
			Nodes:   int(binary.LittleEndian.Uint32(data[offset+8:])),
			Support: int(data[offset+12]),
		}
	}
	if err := validateDauFunctionClasses(inputs, classes, true); err != nil {
		return 0, nil, err
	}
	return inputs, classes, nil
}

type DauExactNPNSummary struct {
	Nodes, SemiCanonical, Canonical, Total int
}

type DauExactNPNSummaryOptions struct {
	MaxCanonicalCandidates    uint64
	MaxCanonicalScheduleSteps int
	Stop                      func() bool
}

// SummarizeDauExactNPN replaces Dau_ExactNpnForClasses and
// Dau_ExactNpnPrint with owned counts for the exact supported domain.
func SummarizeDauExactNPN(classes []DauFunctionClass, inputs, variables, maxNodes int) ([]DauExactNPNSummary, error) {
	return SummarizeDauExactNPNWithOptions(classes, inputs, variables, maxNodes, DauExactNPNSummaryOptions{})
}

func SummarizeDauExactNPNWithOptions(classes []DauFunctionClass, inputs, variables, maxNodes int, options DauExactNPNSummaryOptions) ([]DauExactNPNSummary, error) {
	if inputs < 1 || inputs > 6 || variables < 0 || variables > inputs || maxNodes < 0 || maxNodes > 0x7fffffff || options.MaxCanonicalScheduleSteps < 0 {
		return nil, ErrDimension
	}
	if err := validateDauFunctionClasses(inputs, classes, true); err != nil {
		return nil, err
	}
	tables := make([]*DauNPNClasses, variables+1)
	for width := 0; width <= min(variables, 4); width++ {
		var err error
		tables[width], err = EnumerateDauNPNClasses(width, DauNPNClassOptions{Stop: options.Stop})
		if err != nil {
			return nil, err
		}
	}
	canonicalizer := dauFunctionCanonicalizer{inputs: variables, tables: tables, maxScheduleSteps: options.MaxCanonicalScheduleSteps, stop: options.Stop}
	if options.MaxCanonicalCandidates != 0 {
		canonicalizer.limited = true
		canonicalizer.remaining = options.MaxCanonicalCandidates
	}
	rows := make([]DauExactNPNSummary, maxNodes+1)
	seen := make([]*insdict.Dict[uint64, bool], len(rows))
	for node := range rows {
		rows[node].Nodes = node
		seen[node] = insdict.NewDict[uint64, bool]()
	}
	for _, class := range classes {
		if class.Nodes > maxNodes || class.Support > variables {
			continue
		}
		rows[class.Nodes].SemiCanonical++
		canonical, _, err := canonicalizer.canonicalize(class.Truth & dauTruthMask(variables))
		if err != nil {
			return nil, err
		}
		if _, found := seen[class.Nodes].Get2(canonical); !found {
			seen[class.Nodes].Set(canonical, true)
			rows[class.Nodes].Canonical++
		}
	}
	total := 0
	for index := range rows {
		total += rows[index].Canonical
		rows[index].Total = total
	}
	return rows, nil
}

type DauFunctionDump struct {
	Nodes int
	Lines []string
}

// FormatDauFunctionDumps replaces Dau_DumpFuncs's family of output files with
// deterministic owned hexadecimal lines.
func FormatDauFunctionDumps(classes []DauFunctionClass, inputs, variables, maxNodes int) ([]DauFunctionDump, error) {
	if inputs < 1 || inputs > 6 || (variables != 4 && variables != 5) || variables > inputs || maxNodes < 0 || maxNodes > 31 {
		return nil, ErrDimension
	}
	if err := validateDauFunctionClasses(inputs, classes, true); err != nil {
		return nil, err
	}
	result := make([]DauFunctionDump, maxNodes+1)
	for node := range result {
		result[node].Nodes = node
	}
	for _, class := range classes {
		if class.Support != variables || class.Nodes > maxNodes {
			continue
		}
		if variables == 4 {
			result[class.Nodes].Lines = append(result[class.Nodes].Lines, fmt.Sprintf("%04x", class.Truth&0xffff))
		} else {
			result[class.Nodes].Lines = append(result[class.Nodes].Lines, fmt.Sprintf("%08x", class.Truth&0xffffffff))
		}
	}
	for index := range result {
		result[index].Lines = slices.Clone(result[index].Lines)
	}
	return result, nil
}
