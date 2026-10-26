package abc

// Ternary, structural and mapped-cut analysis from src/sat/bmc/bmcBmc3.c.

import (
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
)

type AigBMC3TernaryMode uint8

const (
	AigBMC3StableRegisters AigBMC3TernaryMode = iota
	AigBMC3FirstUnknownOutput
)

type AigBMC3TernaryOptions struct {
	Mode                              AigBMC3TernaryMode
	MaxFrames, MaxObjects, MaxEntries int
	Stop                              func() bool
}

type AigBMC3TernaryHistory struct {
	Values                                   [][]AigTernary
	RegisterBinaryCounts, OutputBinaryCounts []int
	FirstUnknownOutputFrame                  int
	Complete                                 bool
	graph                                    *AigManager
}

func normalizeAigBMC3TernaryOptions(options AigBMC3TernaryOptions) (AigBMC3TernaryOptions, error) {
	if options.Mode > AigBMC3FirstUnknownOutput || options.MaxFrames < 0 || options.MaxObjects < 0 || options.MaxEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1000
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 16 << 20
	}
	return options, nil
}

func checkAigBMC3TernaryGraph(graph *AigManager) error {
	if graph == nil || graph.RegisterCount() == 0 || graph.PrimaryOutputCount() == 0 {
		return ErrDimension
	}
	if err := graph.Check(); err != nil {
		return err
	}
	for _, object := range graph.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return ErrUnsupported
		}
	}
	return nil
}

func aigBMC3Read(values []AigTernary, literal AigLit) AigTernary {
	value := values[literal.ID()]
	if literal.Complemented() {
		return value.Not()
	}
	return value
}

func aigBMC3BinaryCounts(graph *AigManager, values []AigTernary) (registers, outputs int) {
	for _, id := range graph.cos[:graph.PrimaryOutputCount()] {
		if values[id] != AigTernaryUnknown {
			outputs++
		}
	}
	for _, id := range graph.cos[graph.PrimaryOutputCount():] {
		if graph.objects[id].Fanin0.ID() != 0 && values[id] != AigTernaryUnknown {
			registers++
		}
	}
	return registers, outputs
}

// AnalyzeBMC3Ternary preserves both source drivers: convergence after five
// non-decreasing register-binary counts, or the first frame containing an
// unknown PO. Full owned object rows replace packed two-bit mutation.
func (graph *AigManager) AnalyzeBMC3Ternary(options AigBMC3TernaryOptions) (AigBMC3TernaryHistory, error) {
	result := AigBMC3TernaryHistory{FirstUnknownOutputFrame: -1}
	if err := checkAigBMC3TernaryGraph(graph); err != nil {
		return result, err
	}
	options, err := normalizeAigBMC3TernaryOptions(options)
	if err != nil {
		return result, err
	}
	if graph.ObjectCount() > options.MaxObjects {
		return result, ErrLimit
	}
	if _, err = checkedSize(0, options.MaxFrames, graph.ObjectCount(), options.MaxEntries); err != nil {
		return result, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return result, ErrCancelled
	}
	state := make([]AigTernary, graph.RegisterCount())
	for index := range state {
		state[index] = AigTernaryZero
	}
	previousBinary := math.MaxInt
	nonDecreasing := 0
	currentBinary := graph.RegisterCount()
	for frame := 0; frame < options.MaxFrames; frame++ {
		if options.Mode == AigBMC3StableRegisters {
			if nonDecreasing >= 5 || previousBinary == 0 {
				result.Complete = true
				break
			}
			if currentBinary >= previousBinary {
				nonDecreasing++
			}
			previousBinary = currentBinary
		}
		if stopped() {
			return AigBMC3TernaryHistory{}, ErrCancelled
		}
		values := make([]AigTernary, graph.ObjectCount())
		values[0] = AigTernaryOne
		for _, id := range graph.cis[:graph.PrimaryInputCount()] {
			values[id] = AigTernaryUnknown
		}
		for register, id := range graph.cis[graph.PrimaryInputCount():] {
			values[id] = state[register]
		}
		for id, object := range graph.objects {
			if stopped() {
				return AigBMC3TernaryHistory{}, ErrCancelled
			}
			switch object.Type {
			case AigAND:
				values[id] = aigTernaryAnd(aigBMC3Read(values, object.Fanin0), aigBMC3Read(values, object.Fanin1))
			case AigCO:
				values[id] = aigBMC3Read(values, object.Fanin0)
			}
		}
		registerBinary, outputBinary := aigBMC3BinaryCounts(graph, values)
		result.Values = append(result.Values, values)
		result.RegisterBinaryCounts = append(result.RegisterBinaryCounts, registerBinary)
		result.OutputBinaryCounts = append(result.OutputBinaryCounts, outputBinary)
		for register, id := range graph.cos[graph.PrimaryOutputCount():] {
			state[register] = values[id]
		}
		currentBinary = registerBinary
		if options.Mode == AigBMC3FirstUnknownOutput && outputBinary < graph.PrimaryOutputCount() {
			result.FirstUnknownOutputFrame = frame
			result.Complete = true
			break
		}
	}
	if options.Mode == AigBMC3StableRegisters && !result.Complete && (nonDecreasing >= 5 || previousBinary == 0) {
		result.Complete = true
	}
	result.graph = graph
	return result, nil
}

// UnknownConeCounts returns the source recursion's per-frame count of unknown
// objects in one PO cone, following latch outputs into the preceding frame.
func (history AigBMC3TernaryHistory) UnknownConeCounts(frame, output int) ([]int, error) {
	graph := history.graph
	if graph == nil || frame < 0 || frame >= len(history.Values) || output < 0 || output >= graph.PrimaryOutputCount() {
		return nil, ErrDimension
	}
	root := graph.objects[graph.cos[output]].Fanin0.ID()
	counts := make([]int, frame+1)
	seen := make([]bool, (frame+1)*graph.ObjectCount())
	type visit struct {
		frame int
		id    AigID
	}
	stack := []visit{{frame: frame, id: root}}
	for len(stack) != 0 {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		mark := item.frame*graph.ObjectCount() + int(item.id)
		if seen[mark] {
			continue
		}
		seen[mark] = true
		value := history.Values[item.frame][item.id]
		if value == AigTernaryUnknown {
			counts[item.frame]++
		}
		object := graph.objects[item.id]
		switch object.Type {
		case AigConstant1, AigCI:
			terminal := slices.Index(graph.cis, item.id)
			if terminal >= graph.PrimaryInputCount() && item.frame > 0 {
				register := terminal - graph.PrimaryInputCount()
				stack = append(stack, visit{frame: item.frame - 1, id: graph.cos[graph.PrimaryOutputCount()+register]})
			}
		case AigCO, AigBuffer:
			stack = append(stack, visit{frame: item.frame, id: object.Fanin0.ID()})
		case AigAND:
			stack = append(stack, visit{frame: item.frame, id: object.Fanin1.ID()}, visit{frame: item.frame, id: object.Fanin0.ID()})
		}
	}
	return counts, nil
}

type AigBMC3SectionOptions struct {
	MaxSections, MaxVisits int
	Stop                   func() bool
}

// BMC3Sections returns PO-to-register-input sequential cone layers in native
// postorder without borrowing traversal IDs.
func (graph *AigManager) BMC3Sections(options AigBMC3SectionOptions) ([][]AigID, error) {
	if graph == nil || options.MaxSections < 0 || options.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if options.MaxSections == 0 {
		options.MaxSections = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 8 << 20
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	seen := make([]bool, graph.ObjectCount())
	seen[0] = true
	roots := slices.Clone(graph.cos[:graph.PrimaryOutputCount()])
	for _, id := range roots {
		seen[id] = true
	}
	sections := make([][]AigID, 0)
	visits := 0
	type event struct {
		id   AigID
		exit bool
	}
	for len(roots) != 0 {
		if len(sections) == options.MaxSections {
			return nil, ErrLimit
		}
		section := make([]AigID, 0)
		for _, root := range roots {
			driver := graph.objects[root].Fanin0.ID()
			stack := []event{{id: driver}}
			for len(stack) != 0 {
				if stopped() {
					return nil, ErrCancelled
				}
				e := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if e.exit {
					section = append(section, e.id)
					continue
				}
				if seen[e.id] {
					continue
				}
				if visits == options.MaxVisits {
					return nil, ErrLimit
				}
				visits++
				seen[e.id] = true
				stack = append(stack, event{id: e.id, exit: true})
				object := graph.objects[e.id]
				if object.Type == AigAND {
					stack = append(stack, event{id: object.Fanin1.ID()}, event{id: object.Fanin0.ID()})
				}
			}
		}
		sections = append(sections, section)
		roots = roots[:0]
		for _, id := range section {
			terminal := slices.Index(graph.cis, id)
			if terminal < graph.PrimaryInputCount() {
				continue
			}
			register := terminal - graph.PrimaryInputCount()
			li := graph.cos[graph.PrimaryOutputCount()+register]
			if !seen[li] {
				seen[li] = true
				roots = append(roots, li)
			}
		}
	}
	return sections, nil
}

type AigBMC3Supergate struct {
	Leaves     []AigID
	Referenced int
}

// BMC3ReferenceOrder implements Aig_NodeCompareRefsIncrease over an owned ID
// list: increasing structural reference count, then increasing object ID.
func (graph *AigManager) BMC3ReferenceOrder(ids []AigID) ([]AigID, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	result := slices.Clone(ids)
	for _, id := range result {
		if uint64(id) >= uint64(len(graph.objects)) || graph.objects[id].Type == AigNone {
			return nil, ErrInvalidLiteral
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		a, b := result[left], result[right]
		if graph.objects[a].RefCount != graph.objects[b].RefCount {
			return graph.objects[a].RefCount < graph.objects[b].RefCount
		}
		return a < b
	})
	return result, nil
}

func (graph *AigManager) BMC3Supergate(output, maxLeaves int) (AigBMC3Supergate, error) {
	var result AigBMC3Supergate
	if graph == nil || output < 0 || output >= graph.PrimaryOutputCount() || maxLeaves < 0 {
		return result, ErrDimension
	}
	if maxLeaves == 0 {
		maxLeaves = 1 << 20
	}
	if err := graph.Check(); err != nil {
		return result, err
	}
	root := graph.objects[graph.cos[output]].Fanin0
	if !root.Complemented() || graph.objects[root.ID()].Type != AigAND {
		result.Leaves = []AigID{root.ID()}
	} else {
		object := graph.objects[root.ID()]
		stack := []AigLit{object.Fanin1, object.Fanin0}
		seen := make([]bool, graph.ObjectCount())
		for len(stack) != 0 {
			literal := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			object = graph.objects[literal.ID()]
			if literal.Complemented() || object.Type == AigCI {
				if !seen[literal.ID()] {
					if len(result.Leaves) == maxLeaves {
						return AigBMC3Supergate{}, ErrLimit
					}
					seen[literal.ID()] = true
					result.Leaves = append(result.Leaves, literal.ID())
				}
				continue
			}
			if object.Type != AigAND {
				return AigBMC3Supergate{}, ErrUnsupported
			}
			stack = append(stack, object.Fanin1, object.Fanin0)
		}
	}
	for _, id := range result.Leaves {
		if graph.objects[id].RefCount > 1 {
			result.Referenced++
		}
	}
	return result, nil
}

func BMC3TruthCofactor(truth uint16, variable int, value bool) (uint16, error) {
	if variable < 0 || variable > 3 {
		return 0, ErrDimension
	}
	mask := [4]uint16{0xaaaa, 0xcccc, 0xf0f0, 0xff00}[variable]
	shift := 1 << variable
	if value {
		return (truth & mask) | (truth&mask)>>shift, nil
	}
	return (truth &^ mask) | (truth&^mask)<<shift, nil
}

func bmc3TruthIndependent(truth uint16, variable int) bool {
	zero, _ := BMC3TruthCofactor(truth, variable, false)
	one, _ := BMC3TruthCofactor(truth, variable, true)
	return zero == one
}

// BMC3ReduceTruth propagates literal constants and marks functionally absent
// cut inputs with -1. Literal zero/one are constants; values above one are SAT
// literals, matching the source representation.
func BMC3ReduceTruth(truth uint16, literals [4]int) (uint16, [4]int, error) {
	for variable, literal := range literals {
		switch literal {
		case -1:
		case 0, 1:
			var err error
			truth, err = BMC3TruthCofactor(truth, variable, literal == 1)
			if err != nil {
				return 0, literals, err
			}
			literals[variable] = -1
		default:
			if literal < 2 {
				return 0, literals, ErrDimension
			}
		}
	}
	for variable := range literals {
		if literals[variable] == -1 && !bmc3TruthIndependent(truth, variable) {
			return 0, literals, ErrInvariant
		}
		if literals[variable] != -1 && bmc3TruthIndependent(truth, variable) {
			literals[variable] = -1
		}
	}
	return truth, literals, nil
}

type AigBMC3Mapping struct {
	Packed     []int32
	Cuts       []AigCNFCut4
	References []uint32
}

func (graph *AigManager) BMC3Mapping(cuts DARCutOptions, options CNFMappingArrayOptions) (*AigBMC3Mapping, error) {
	packed, err := DeriveAigCNFMappingArray4(graph, cuts, options)
	if err != nil {
		return nil, err
	}
	result := &AigBMC3Mapping{Packed: packed, References: make([]uint32, graph.ObjectCount())}
	for _, id := range graph.cos {
		result.References[graph.objects[id].Fanin0.ID()]++
	}
	for id := range graph.objects {
		offset := int(packed[id])
		if offset == 0 {
			continue
		}
		cut := AigCNFCut4{Root: AigID(id), Truth: uint16(uint32(packed[offset]))}
		for leaf := 0; leaf < 4 && packed[offset+1+leaf] >= 0; leaf++ {
			leafID := AigID(packed[offset+1+leaf])
			cut.Leaves = append(cut.Leaves, leafID)
			result.References[leafID]++
		}
		result.Cuts = append(result.Cuts, cut)
	}
	return result, nil
}

func (mapping *AigBMC3Mapping) FormatBLIF(graph *AigManager) (string, error) {
	if mapping == nil || graph == nil || len(mapping.Packed) < graph.ObjectCount() {
		return "", ErrDimension
	}
	if err := graph.Check(); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString(".model test\n.inputs")
	for _, id := range graph.cis {
		fmt.Fprintf(&builder, " n%d", id)
	}
	builder.WriteString("\n.outputs")
	for _, id := range graph.cos {
		fmt.Fprintf(&builder, " n%d", id)
	}
	builder.WriteString("\n.names n0\n 1\n")
	for _, cut := range mapping.Cuts {
		builder.WriteString(".names")
		for _, leaf := range cut.Leaves {
			fmt.Fprintf(&builder, " n%d", leaf)
		}
		fmt.Fprintf(&builder, " n%d\n", cut.Root)
		cover, err := CNFSOP4Cover(cut.Truth)
		if err != nil {
			return "", err
		}
		for _, encoded := range cover {
			values := [4]byte{'-', '-', '-', '-'}
			value := int(encoded)
			for bit := 3; bit >= 0; bit-- {
				digit := value % 3
				value /= 3
				if digit == 0 {
					values[bit] = '0'
				} else if digit == 1 {
					values[bit] = '1'
				}
			}
			builder.Write(values[:len(cut.Leaves)])
			builder.WriteString(" 1\n")
		}
	}
	for _, id := range graph.cos {
		fanin := graph.objects[id].Fanin0
		fmt.Fprintf(&builder, ".names n%d n%d\n", fanin.ID(), id)
		if fanin.Complemented() {
			builder.WriteString("0 1\n")
		} else {
			builder.WriteString("1 1\n")
		}
	}
	builder.WriteString(".end\n")
	return builder.String(), nil
}

func (mapping *AigBMC3Mapping) WriteBLIFFile(graph *AigManager, path string) error {
	if path == "" {
		return ErrDimension
	}
	text, err := mapping.FormatBLIF(graph)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}
