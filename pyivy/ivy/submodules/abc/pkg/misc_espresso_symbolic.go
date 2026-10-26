package abc

// Symbolic input and output transformations from src/misc/espresso/hack.c.

import (
	"slices"
	"strconv"
)

type EspressoSymbolicOptions struct {
	MaxCubes, MaxOperations, MaxDepth int
	Stop                              func() bool
}

type espressoSymbolicWork struct {
	options    EspressoSymbolicOptions
	operations int
	generated  int
}

func newEspressoSymbolicWork(options EspressoSymbolicOptions) (*espressoSymbolicWork, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxDepth < 0 {
		return nil, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 64 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 16
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return &espressoSymbolicWork{options: options}, nil
}

func (work *espressoSymbolicWork) step(amount int) error {
	if amount < 0 || work.operations > work.options.MaxOperations-amount {
		return ErrLimit
	}
	work.operations += amount
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *espressoSymbolicWork) add(family *EspressoSetFamily, cube *EspressoSet) error {
	work.generated++
	if work.generated > work.options.MaxCubes {
		return ErrLimit
	}
	return family.Add(cube)
}

func espressoSymbolicSize(referenceCount, maxDepth int) (int, error) {
	if referenceCount <= 0 {
		return 0, ErrDimension
	}
	if referenceCount > maxDepth || referenceCount >= strconv.IntSize-1 {
		return 0, ErrLimit
	}
	return 1 << uint(referenceCount), nil
}

func espressoSymbolicLabels(directive EspressoSymbolicDirective, size int) []string {
	labels := make([]string, size)
	for index := range labels {
		if index < len(directive.Labels) {
			labels[index] = directive.Labels[index]
		} else {
			labels[index] = "X" + strconv.Itoa(index)
		}
	}
	return labels
}

func espressoRemapSet(source *EspressoSet, oldLayout, newLayout *EspressoCubeLayout, oldToNew []int, work *espressoSymbolicWork) (*EspressoSet, error) {
	if len(oldToNew) != oldLayout.bitCount {
		return nil, ErrDimension
	}
	result, _ := NewEspressoSet(newLayout.bitCount)
	result.Flags, result.Rank = source.Flags, source.Rank
	for oldPart, newPart := range oldToNew {
		if err := work.step(1); err != nil {
			return nil, err
		}
		if newPart < 0 {
			continue
		}
		present, _ := source.Has(oldPart)
		if present {
			_ = result.Set(newPart, true)
		}
	}
	return result, nil
}

func espressoRemapFamily(source *EspressoSetFamily, oldLayout, newLayout *EspressoCubeLayout, oldToNew []int, work *espressoSymbolicWork) (*EspressoSetFamily, error) {
	if len(source.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	result, _ := NewEspressoSetFamily(len(source.sets), newLayout.bitCount)
	for _, cube := range source.sets {
		mapped, err := espressoRemapSet(cube, oldLayout, newLayout, oldToNew, work)
		if err != nil {
			return nil, err
		}
		if err = work.add(result, mapped); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (pla *EspressoPLA) ApplySymbolicInputs(directives []EspressoSymbolicDirective, options EspressoSymbolicOptions) (*EspressoPLA, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	if len(directives) == 0 {
		return pla.Clone()
	}
	work, err := newEspressoSymbolicWork(options)
	if err != nil {
		return nil, err
	}
	removed := make([]bool, pla.Layout.binaryVariables)
	sizes := make([]int, len(directives))
	for index, directive := range directives {
		sizes[index], err = espressoSymbolicSize(len(directive.References), work.options.MaxDepth)
		if err != nil {
			return nil, err
		}
		for _, reference := range directive.References {
			if reference.Variable < 0 || reference.Variable >= pla.Layout.binaryVariables || reference.Part < 0 || reference.Part > 1 {
				return nil, ErrDimension
			}
			removed[reference.Variable] = true
		}
	}
	newBinary := pla.Layout.binaryVariables
	for _, remove := range removed {
		if remove {
			newBinary--
		}
	}
	output := pla.Layout.OutputVariable()
	partSizes := make([]int, 0, len(pla.Layout.partSizes)-pla.Layout.binaryVariables+newBinary+len(directives))
	for variable := 0; variable < pla.Layout.binaryVariables; variable++ {
		if !removed[variable] {
			partSizes = append(partSizes, 2)
		}
	}
	for variable := pla.Layout.binaryVariables; variable < output; variable++ {
		partSizes = append(partSizes, pla.Layout.partSizes[variable])
	}
	partSizes = append(partSizes, sizes...)
	if output >= 0 {
		partSizes = append(partSizes, pla.Layout.partSizes[output])
	}
	newLayout, err := NewEspressoCubeLayout(partSizes, newBinary)
	if err != nil {
		return nil, err
	}
	oldToNew := make([]int, pla.Layout.bitCount)
	for index := range oldToNew {
		oldToNew[index] = -1
	}
	newVariable := 0
	for variable := 0; variable < pla.Layout.binaryVariables; variable++ {
		if removed[variable] {
			continue
		}
		for part := 0; part < 2; part++ {
			oldToNew[pla.Layout.firstParts[variable]+part] = newLayout.firstParts[newVariable] + part
		}
		newVariable++
	}
	for variable := pla.Layout.binaryVariables; variable < output; variable++ {
		for part := 0; part < pla.Layout.partSizes[variable]; part++ {
			oldToNew[pla.Layout.firstParts[variable]+part] = newLayout.firstParts[newVariable] + part
		}
		newVariable++
	}
	symbolicBaseVariable := newVariable
	newVariable += len(directives)
	if output >= 0 {
		for part := 0; part < pla.Layout.partSizes[output]; part++ {
			oldToNew[pla.Layout.firstParts[output]+part] = newLayout.firstParts[newVariable] + part
		}
	}
	transform := func(source *EspressoSetFamily) (*EspressoSetFamily, error) {
		mapped, mapErr := espressoRemapFamily(source, pla.Layout, newLayout, oldToNew, work)
		if mapErr != nil {
			return nil, mapErr
		}
		for cubeIndex, sourceCube := range source.sets {
			cube := mapped.sets[cubeIndex]
			for directiveIndex, directive := range directives {
				base := newLayout.firstParts[symbolicBaseVariable+directiveIndex]
				for value := 0; value < sizes[directiveIndex]; value++ {
					allowed := true
					for referenceIndex, reference := range directive.References {
						if err := work.step(1); err != nil {
							return nil, err
						}
						bit := value >> uint(len(directive.References)-1-referenceIndex) & 1
						present, _ := sourceCube.Has(pla.Layout.firstParts[reference.Variable] + bit)
						if !present {
							allowed = false
							break
						}
					}
					if allowed {
						_ = cube.Set(base+value, true)
					}
				}
				empty := true
				for value := 0; value < sizes[directiveIndex]; value++ {
					present, _ := cube.Has(base + value)
					empty = empty && !present
				}
				if empty {
					return nil, ErrInvariant
				}
			}
		}
		return mapped, nil
	}
	on, err := transform(pla.On)
	if err != nil {
		return nil, err
	}
	dc, err := transform(pla.DontCare)
	if err != nil {
		return nil, err
	}
	off, err := transform(pla.Off)
	if err != nil {
		return nil, err
	}
	labels := []string(nil)
	if len(pla.Labels) != 0 {
		labels = make([]string, newLayout.bitCount)
		for oldPart, newPart := range oldToNew {
			if newPart >= 0 {
				labels[newPart] = pla.Labels[oldPart]
			}
		}
		for index, directive := range directives {
			copy(labels[newLayout.firstParts[symbolicBaseVariable+index]:], espressoSymbolicLabels(directive, sizes[index]))
		}
	}
	result, err := NewEspressoPLA(newLayout, on, dc, off, labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}

func espressoExpandOutputLayout(pla *EspressoPLA, added int, work *espressoSymbolicWork) (*EspressoPLA, error) {
	output := pla.Layout.OutputVariable()
	if output < 0 || added < 0 || pla.Layout.partSizes[output] > espressoSetMaximumBits-added {
		return nil, ErrDimension
	}
	parts := slices.Clone(pla.Layout.partSizes)
	parts[output] += added
	layout, err := NewEspressoCubeLayout(parts, pla.Layout.binaryVariables)
	if err != nil {
		return nil, err
	}
	oldToNew := make([]int, pla.Layout.bitCount)
	for part := 0; part < pla.Layout.firstParts[output]; part++ {
		oldToNew[part] = part
	}
	for part := 0; part < pla.Layout.partSizes[output]; part++ {
		oldToNew[pla.Layout.firstParts[output]+part] = layout.firstParts[output] + added + part
	}
	on, err := espressoRemapFamily(pla.On, pla.Layout, layout, oldToNew, work)
	if err != nil {
		return nil, err
	}
	dc, err := espressoRemapFamily(pla.DontCare, pla.Layout, layout, oldToNew, work)
	if err != nil {
		return nil, err
	}
	off, err := espressoRemapFamily(pla.Off, pla.Layout, layout, oldToNew, work)
	if err != nil {
		return nil, err
	}
	labels := []string(nil)
	if len(pla.Labels) != 0 {
		labels = make([]string, layout.bitCount)
		for part := 0; part < pla.Layout.firstParts[output]; part++ {
			labels[part] = pla.Labels[part]
		}
		for part := 0; part < pla.Layout.partSizes[output]; part++ {
			labels[layout.firstParts[output]+added+part] = pla.Labels[pla.Layout.firstParts[output]+part]
		}
	}
	result, err := NewEspressoPLA(layout, on, dc, off, labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}

func espressoIntersectSymbolic(layout *EspressoCubeLayout, first, second *EspressoSetFamily, work *espressoSymbolicWork) (*EspressoSetFamily, error) {
	sharp, err := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	result, err := layout.intersectFamilies(first, second, sharp)
	if err != nil {
		return nil, err
	}
	work.operations += sharp.operations
	work.generated += sharp.generated
	if work.operations > work.options.MaxOperations || work.generated > work.options.MaxCubes {
		return nil, ErrLimit
	}
	return result, nil
}

func (pla *EspressoPLA) ApplySymbolicOutputs(directives []EspressoSymbolicDirective, options EspressoSymbolicOptions) (*EspressoPLA, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	if len(directives) == 0 {
		return pla.Clone()
	}
	work, err := newEspressoSymbolicWork(options)
	if err != nil {
		return nil, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 {
		return nil, ErrDimension
	}
	totalAdded := 0
	sizes := make([]int, len(directives))
	remove := make([]bool, pla.Layout.partSizes[output])
	for index, directive := range directives {
		sizes[index], err = espressoSymbolicSize(len(directive.References), work.options.MaxDepth)
		if err != nil || totalAdded > espressoSetMaximumBits-sizes[index] {
			if err != nil {
				return nil, err
			}
			return nil, ErrLimit
		}
		totalAdded += sizes[index]
		for _, reference := range directive.References {
			if reference.Part < 0 || reference.Part >= pla.Layout.partSizes[output] {
				return nil, ErrDimension
			}
			remove[reference.Part] = true
		}
	}
	working, err := pla.Clone()
	if err != nil {
		return nil, err
	}
	if len(working.DontCare.sets) != 0 {
		union, joinErr := JoinEspressoSetFamilies(working.DontCare, working.Off)
		if joinErr != nil {
			return nil, joinErr
		}
		working.On, err = working.Layout.ComplementCover(union, EspressoSharpOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
		if err != nil {
			return nil, err
		}
	}
	expanded, err := espressoExpandOutputLayout(working, totalAdded, work)
	if err != nil {
		return nil, err
	}
	base := 0
	for directiveIndex, directive := range directives {
		newOn, _ := NewEspressoSetFamily(0, expanded.Layout.bitCount)
		for value := 0; value < sizes[directiveIndex]; value++ {
			var region *EspressoSetFamily
			for referenceIndex, reference := range directive.References {
				bit := value >> uint(len(directive.References)-1-referenceIndex) & 1
				source := expanded.Off
				if bit != 0 {
					source = expanded.On
				}
				cofactor, cofactorErr := expanded.Layout.CofactorOutput(source, totalAdded+reference.Part, EspressoCoverManipulationOptions{MaxCubes: work.options.MaxCubes, MaxOperations: work.options.MaxOperations, Stop: work.options.Stop})
				if cofactorErr != nil {
					return nil, cofactorErr
				}
				if region == nil {
					region = cofactor
				} else {
					region, err = espressoIntersectSymbolic(expanded.Layout, region, cofactor, work)
					if err != nil {
						return nil, err
					}
				}
			}
			region, err = espressoIntersectSymbolic(expanded.Layout, region, expanded.On, work)
			if err != nil {
				return nil, err
			}
			for _, cube := range region.sets {
				clone, _ := cube.Clone()
				_ = clone.Set(expanded.Layout.firstParts[output]+base+value, true)
				if err = work.add(newOn, clone); err != nil {
					return nil, err
				}
			}
		}
		expanded.On = newOn
		base += sizes[directiveIndex]
	}
	kept := 0
	for _, removed := range remove {
		if !removed {
			kept++
		}
	}
	parts := slices.Clone(expanded.Layout.partSizes)
	parts[output] = totalAdded + kept
	finalLayout, err := NewEspressoCubeLayout(parts, expanded.Layout.binaryVariables)
	if err != nil {
		return nil, err
	}
	expandedToFinal := make([]int, expanded.Layout.bitCount)
	for index := range expandedToFinal {
		expandedToFinal[index] = -1
	}
	for part := 0; part < expanded.Layout.firstParts[output]; part++ {
		expandedToFinal[part] = part
	}
	for part := 0; part < totalAdded; part++ {
		expandedToFinal[expanded.Layout.firstParts[output]+part] = finalLayout.firstParts[output] + part
	}
	next := totalAdded
	for part, removed := range remove {
		if !removed {
			expandedToFinal[expanded.Layout.firstParts[output]+totalAdded+part] = finalLayout.firstParts[output] + next
			next++
		}
	}
	on, err := espressoRemapFamily(expanded.On, expanded.Layout, finalLayout, expandedToFinal, work)
	if err != nil {
		return nil, err
	}
	dc, err := espressoRemapFamily(expanded.DontCare, expanded.Layout, finalLayout, expandedToFinal, work)
	if err != nil {
		return nil, err
	}
	off, _ := NewEspressoSetFamily(0, finalLayout.bitCount)
	labels := []string(nil)
	if len(pla.Labels) != 0 {
		labels = make([]string, finalLayout.bitCount)
		copy(labels, pla.Labels[:pla.Layout.firstParts[output]])
		base = 0
		for index, directive := range directives {
			copy(labels[finalLayout.firstParts[output]+base:], espressoSymbolicLabels(directive, sizes[index]))
			base += sizes[index]
		}
		next = totalAdded
		for part, removed := range remove {
			if !removed {
				labels[finalLayout.firstParts[output]+next] = pla.Labels[pla.Layout.firstParts[output]+part]
				next++
			}
		}
	}
	contain := EspressoContainOptions{MaxSets: work.options.MaxCubes, MaxChecks: work.options.MaxOperations, Stop: work.options.Stop}
	on, err = on.Contain(contain)
	if err != nil {
		return nil, err
	}
	dc, err = dc.Contain(contain)
	if err != nil {
		return nil, err
	}
	for variable := range finalLayout.partSizes {
		on, err = on.DistanceOneMerge(finalLayout, variable, contain)
		if err != nil {
			return nil, err
		}
		dc, err = dc.DistanceOneMerge(finalLayout, variable, contain)
		if err != nil {
			return nil, err
		}
	}
	on, err = on.Contain(contain)
	if err != nil {
		return nil, err
	}
	dc, err = dc.Contain(contain)
	if err != nil {
		return nil, err
	}
	result, err := NewEspressoPLA(finalLayout, on, dc, off, labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}
