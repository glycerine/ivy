package abc

// Named don't-care extraction from src/misc/espresso/hack.c.

import (
	"slices"
	"strconv"
	"strings"
)

func espressoHackVariableCofactor(layout *EspressoCubeLayout, family *EspressoSetFamily, variable, value int, work *espressoSymbolicWork) (*EspressoSetFamily, error) {
	if variable < 0 || variable >= layout.binaryVariables || value < 0 || value > 1 {
		return nil, ErrDimension
	}
	if len(family.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	result, _ := NewEspressoSetFamily(len(family.sets), layout.bitCount)
	start := layout.firstParts[variable]
	for _, cube := range family.sets {
		if err := work.step(2); err != nil {
			return nil, err
		}
		present, _ := cube.Has(start + value)
		if !present {
			continue
		}
		clone, _ := cube.Clone()
		_ = clone.Set(start, true)
		_ = clone.Set(start+1, true)
		if err := work.add(result, clone); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func espressoDontCareLabel(label string) bool {
	return strings.HasPrefix(label, "DONT_CARE") || strings.HasPrefix(label, "DONTCARE") || strings.HasPrefix(label, "dont_care") || strings.HasPrefix(label, "dontcare")
}

// MapDontCareVariable converts the first binary DONT_CARE control into an
// explicit DC cover and removes that variable from the layout.
func (pla *EspressoPLA) MapDontCareVariable(options EspressoSymbolicOptions) (*EspressoPLA, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	if len(pla.Labels) == 0 {
		return pla.Clone()
	}
	variable := -1
	for current := 0; current < pla.Layout.binaryVariables && variable < 0; current++ {
		for part := 0; part < 2; part++ {
			if espressoDontCareLabel(pla.Labels[pla.Layout.firstParts[current]+part]) {
				variable = current
				break
			}
		}
	}
	if variable < 0 {
		return pla.Clone()
	}
	work, err := newEspressoSymbolicWork(options)
	if err != nil {
		return nil, err
	}
	zero, err := espressoHackVariableCofactor(pla.Layout, pla.On, variable, 0, work)
	if err != nil {
		return nil, err
	}
	one, err := espressoHackVariableCofactor(pla.Layout, pla.On, variable, 1, work)
	if err != nil {
		return nil, err
	}
	zeroBar, err := pla.Layout.ComplementCover(zero, EspressoSharpOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	oneBar, err := pla.Layout.ComplementCover(one, EspressoSharpOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	zeroOnly, err := espressoIntersectSymbolic(pla.Layout, zero, oneBar, work)
	if err != nil {
		return nil, err
	}
	oneOnly, err := espressoIntersectSymbolic(pla.Layout, one, zeroBar, work)
	if err != nil {
		return nil, err
	}
	dc, err := JoinEspressoSetFamilies(zeroOnly, oneOnly)
	if err != nil {
		return nil, err
	}
	dc, err = pla.Layout.SimplifyCover(dc, EspressoSharpOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	dcBar, err := pla.Layout.ComplementCover(dc, EspressoSharpOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	on, err := espressoIntersectSymbolic(pla.Layout, pla.On, dcBar, work)
	if err != nil {
		return nil, err
	}
	independent, _ := NewEspressoSetFamily(len(on.sets), pla.Layout.bitCount)
	for _, cube := range on.sets {
		if err = work.step(2); err != nil {
			return nil, err
		}
		if pla.Layout.variableFull(cube, variable) {
			if err = work.add(independent, cube); err != nil {
				return nil, err
			}
		}
	}
	parts := append(slices.Clone(pla.Layout.partSizes[:variable]), pla.Layout.partSizes[variable+1:]...)
	layout, err := NewEspressoCubeLayout(parts, pla.Layout.binaryVariables-1)
	if err != nil {
		return nil, err
	}
	oldToNew := make([]int, pla.Layout.bitCount)
	for part := range oldToNew {
		switch {
		case part < pla.Layout.firstParts[variable]:
			oldToNew[part] = part
		case part < pla.Layout.firstParts[variable]+2:
			oldToNew[part] = -1
		default:
			oldToNew[part] = part - 2
		}
	}
	on, err = espressoRemapFamily(independent, pla.Layout, layout, oldToNew, work)
	if err != nil {
		return nil, err
	}
	dc, err = espressoRemapFamily(dc, pla.Layout, layout, oldToNew, work)
	if err != nil {
		return nil, err
	}
	union, err := JoinEspressoSetFamilies(on, dc)
	if err != nil {
		return nil, err
	}
	off, err := layout.ComplementCover(union, EspressoSharpOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	labels := make([]string, 0)
	if len(pla.Labels) != 0 {
		labels = append(labels, pla.Labels[:pla.Layout.firstParts[variable]]...)
		labels = append(labels, pla.Labels[pla.Layout.firstParts[variable]+2:]...)
	}
	result, err := NewEspressoPLA(layout, on, dc, off, labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}

type EspressoFSMArcKind uint8

const (
	EspressoFSMEveryState EspressoFSMArcKind = iota
	EspressoFSMNowhere
	EspressoFSMSpecific
)

type EspressoFSMArcStatistic struct {
	Kind          EspressoFSMArcKind
	From, To      int
	Before, After int
}

type EspressoFSMDisassembleOptions struct {
	Verbose                                    bool
	MaxCubes, MaxOperations, MaxRows, MaxDepth int
	MaxIterations, MaxBytes                    int
	Stop                                       func() bool
	Minimize                                   EspressoMinimizeOptions
}

type EspressoFSMDisassembly struct {
	PLA         *EspressoPLA
	Statistics  []EspressoFSMArcStatistic
	Diagnostics string
	KISS        []byte
}

func normalizeEspressoFSMDisassembleOptions(options EspressoFSMDisassembleOptions) (EspressoFSMDisassembleOptions, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxRows < 0 || options.MaxDepth < 0 || options.MaxIterations < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 64 << 20
	}
	if options.MaxRows == 0 {
		options.MaxRows = 1 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 16
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 64
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	if options.Minimize.MaxCubes == 0 {
		options.Minimize.MaxCubes = options.MaxCubes
	}
	if options.Minimize.MaxOperations == 0 {
		options.Minimize.MaxOperations = options.MaxOperations
	}
	if options.Minimize.MaxRows == 0 {
		options.Minimize.MaxRows = options.MaxRows
	}
	if options.Minimize.MaxDepth == 0 {
		options.Minimize.MaxDepth = options.MaxDepth
	}
	if options.Minimize.MaxIterations == 0 {
		options.Minimize.MaxIterations = options.MaxIterations
	}
	if options.Minimize.Stop == nil {
		options.Minimize.Stop = options.Stop
	}
	return options, nil
}

func espressoFSMSimplify(layout *EspressoCubeLayout, family *EspressoSetFamily, options EspressoFSMDisassembleOptions) (*EspressoSetFamily, error) {
	if len(family.sets) == 0 {
		return family.Clone()
	}
	empty, _ := NewEspressoSetFamily(0, layout.bitCount)
	off, err := layout.ComplementCover(family, EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	result, err := layout.MinimizeEspressoCover(family, empty, off, options.Minimize)
	if err != nil {
		return nil, err
	}
	return result.Cover, nil
}

func espressoFSMAppend(destination, source *EspressoSetFamily, work *espressoSymbolicWork) error {
	for _, cube := range source.sets {
		if err := work.add(destination, cube); err != nil {
			return err
		}
	}
	return nil
}

func espressoFSMSymbol(labels []string, start, size int, cube *EspressoSet) (string, error) {
	count, selected := 0, -1
	for part := 0; part < size; part++ {
		present, _ := cube.Has(start + part)
		if present {
			count++
			selected = part
		}
	}
	switch {
	case count == size:
		return "-", nil
	case count == 0:
		return "~", nil
	case count == 1:
		if len(labels) == 0 || labels[start+selected] == "" {
			return "v." + strconv.Itoa(selected), nil
		}
		return labels[start+selected], nil
	default:
		return "", ErrInvariant
	}
}

func espressoFSMKISS(pla *EspressoPLA, states, external int, maxBytes int, stop func() bool) ([]byte, error) {
	layout := pla.Layout
	labels, err := pla.DefaultLabels()
	if err != nil {
		return nil, err
	}
	present := layout.binaryVariables
	next := present + 1
	if present >= len(layout.partSizes) || next >= len(layout.partSizes) {
		return nil, ErrInvariant
	}
	var output int
	if external > 0 {
		output = next + 1
	}
	var builder strings.Builder
	write := func(text string) error {
		if stop != nil && stop() {
			return ErrCancelled
		}
		if len(text) > maxBytes-builder.Len() {
			return ErrLimit
		}
		builder.WriteString(text)
		return nil
	}
	for _, cube := range pla.On.sets {
		for variable := 0; variable < layout.binaryVariables; variable++ {
			start := layout.firstParts[variable]
			low, _ := cube.Has(start)
			high, _ := cube.Has(start + 1)
			if err := write(string("?01-"[btoi(low)+2*btoi(high)])); err != nil {
				return nil, err
			}
		}
		presentText, err := espressoFSMSymbol(labels, layout.firstParts[present], states, cube)
		if err != nil {
			return nil, err
		}
		nextText, err := espressoFSMSymbol(labels, layout.firstParts[next], states, cube)
		if err != nil {
			return nil, err
		}
		if err = write(" " + presentText + " " + nextText + " "); err != nil {
			return nil, err
		}
		if external > 0 {
			for part := 0; part < external; part++ {
				presentPart, _ := cube.Has(layout.firstParts[output] + part)
				if err = write(string("~1"[btoi(presentPart)])); err != nil {
					return nil, err
				}
			}
		}
		if err = write("\n"); err != nil {
			return nil, err
		}
	}
	return []byte(builder.String()), nil
}

// DisassembleFSM partitions symbolic-state cubes by transition, minimizes each
// partition, and returns the source KISS rendering without writing stdout.
func (pla *EspressoPLA) DisassembleFSM(options EspressoFSMDisassembleOptions) (EspressoFSMDisassembly, error) {
	if err := pla.check(); err != nil {
		return EspressoFSMDisassembly{}, err
	}
	options, err := normalizeEspressoFSMDisassembleOptions(options)
	if err != nil {
		return EspressoFSMDisassembly{}, err
	}
	if options.Stop != nil && options.Stop() {
		return EspressoFSMDisassembly{}, ErrCancelled
	}
	layout := pla.Layout
	if layout.MultipleValuedVariableCount() != 2 {
		return EspressoFSMDisassembly{}, ErrDimension
	}
	present := layout.binaryVariables
	output := layout.OutputVariable()
	states := layout.partSizes[present]
	if states <= 0 || layout.partSizes[output] < states {
		return EspressoFSMDisassembly{}, ErrDimension
	}
	external := layout.partSizes[output] - states
	work, err := newEspressoSymbolicWork(EspressoSymbolicOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
	if err != nil {
		return EspressoFSMDisassembly{}, err
	}
	combined, _ := NewEspressoSetFamily(0, layout.bitCount)
	statistics := make([]EspressoFSMArcStatistic, 0)
	var diagnostics strings.Builder
	writeDiagnostic := func(text string) error {
		if !options.Verbose {
			return nil
		}
		if len(text) > options.MaxBytes-diagnostics.Len() {
			return ErrLimit
		}
		diagnostics.WriteString(text)
		return nil
	}
	presentStart := layout.firstParts[present]
	nextStart := layout.firstParts[output]
	for to := 0; to < states; to++ {
		family, _ := NewEspressoSetFamily(0, layout.bitCount)
		for _, cube := range pla.On.sets {
			if err = work.step(states + 1); err != nil {
				return EspressoFSMDisassembly{}, err
			}
			nextPresent, _ := cube.Has(nextStart + to)
			if layout.variableFull(cube, present) && nextPresent {
				if err = work.add(family, cube); err != nil {
					return EspressoFSMDisassembly{}, err
				}
			}
		}
		before := len(family.sets)
		if before == 0 {
			continue
		}
		family, err = espressoFSMSimplify(layout, family, options)
		if err != nil {
			return EspressoFSMDisassembly{}, err
		}
		for _, cube := range family.sets {
			_ = cube.Set(nextStart+to, true)
		}
		stat := EspressoFSMArcStatistic{Kind: EspressoFSMEveryState, From: -1, To: to, Before: before, After: len(family.sets)}
		statistics = append(statistics, stat)
		if err = writeDiagnostic("# state EVERY to " + strconv.Itoa(to) + ", before=" + strconv.Itoa(stat.Before) + " after=" + strconv.Itoa(stat.After) + "\n"); err != nil {
			return EspressoFSMDisassembly{}, err
		}
		if err = espressoFSMAppend(combined, family, work); err != nil {
			return EspressoFSMDisassembly{}, err
		}
	}
	goNowhere, _ := NewEspressoSetFamily(0, layout.bitCount)
	for _, cube := range pla.On.sets {
		hasNext := false
		for state := 0; state < states; state++ {
			if err = work.step(1); err != nil {
				return EspressoFSMDisassembly{}, err
			}
			presentPart, _ := cube.Has(nextStart + state)
			hasNext = hasNext || presentPart
		}
		if !hasNext {
			if err = work.add(goNowhere, cube); err != nil {
				return EspressoFSMDisassembly{}, err
			}
		}
	}
	nowhereBefore := len(goNowhere.sets)
	goNowhere, err = layout.UnravelRange(goNowhere, present, present, EspressoCoverManipulationOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return EspressoFSMDisassembly{}, err
	}
	nowhereStat := EspressoFSMArcStatistic{Kind: EspressoFSMNowhere, From: -1, To: -1, Before: nowhereBefore, After: len(goNowhere.sets)}
	statistics = append(statistics, nowhereStat)
	if err = writeDiagnostic("# state ANY to NOWHERE, before=" + strconv.Itoa(nowhereStat.Before) + " after=" + strconv.Itoa(nowhereStat.After) + "\n"); err != nil {
		return EspressoFSMDisassembly{}, err
	}
	if err = espressoFSMAppend(combined, goNowhere, work); err != nil {
		return EspressoFSMDisassembly{}, err
	}
	for from := 0; from < states; from++ {
		for to := 0; to < states; to++ {
			family, _ := NewEspressoSetFamily(0, layout.bitCount)
			for _, cube := range pla.On.sets {
				if err = work.step(2); err != nil {
					return EspressoFSMDisassembly{}, err
				}
				fromPresent, _ := cube.Has(presentStart + from)
				toPresent, _ := cube.Has(nextStart + to)
				if layout.variableFull(cube, present) || !fromPresent || !toPresent {
					continue
				}
				clone, _ := cube.Clone()
				for state := 0; state < states; state++ {
					_ = clone.Set(presentStart+state, state == from)
					_ = clone.Set(nextStart+state, state == to)
				}
				if err = work.add(family, clone); err != nil {
					return EspressoFSMDisassembly{}, err
				}
			}
			before := len(family.sets)
			if before == 0 {
				continue
			}
			family, err = espressoFSMSimplify(layout, family, options)
			if err != nil {
				return EspressoFSMDisassembly{}, err
			}
			for _, cube := range family.sets {
				_ = cube.Set(nextStart+to, true)
			}
			stat := EspressoFSMArcStatistic{Kind: EspressoFSMSpecific, From: from, To: to, Before: before, After: len(family.sets)}
			statistics = append(statistics, stat)
			if err = writeDiagnostic("# state " + strconv.Itoa(from) + " to " + strconv.Itoa(to) + ", before=" + strconv.Itoa(stat.Before) + " after=" + strconv.Itoa(stat.After) + "\n"); err != nil {
				return EspressoFSMDisassembly{}, err
			}
			if err = espressoFSMAppend(combined, family, work); err != nil {
				return EspressoFSMDisassembly{}, err
			}
		}
	}
	parts := slices.Clone(layout.partSizes[:output])
	parts = append(parts, states)
	if external > 0 {
		parts = append(parts, external)
	}
	finalLayout, err := NewEspressoCubeLayout(parts, layout.binaryVariables)
	if err != nil || finalLayout.bitCount != layout.bitCount {
		return EspressoFSMDisassembly{}, ErrInvariant
	}
	empty, _ := NewEspressoSetFamily(0, finalLayout.bitCount)
	resultPLA, err := NewEspressoPLA(finalLayout, combined, empty, pla.Off, pla.Labels)
	if err != nil {
		return EspressoFSMDisassembly{}, err
	}
	resultPLA.Filename, resultPLA.Type = pla.Filename, pla.Type
	kiss, err := espressoFSMKISS(resultPLA, states, external, options.MaxBytes, options.Stop)
	if err != nil {
		return EspressoFSMDisassembly{}, err
	}
	return EspressoFSMDisassembly{PLA: resultPLA, Statistics: statistics, Diagnostics: diagnostics.String(), KISS: kiss}, nil
}
