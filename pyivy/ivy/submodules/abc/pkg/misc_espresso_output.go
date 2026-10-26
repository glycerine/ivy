package abc

// Deterministic PLA and diagnostic formatting from src/misc/espresso/cvrout.c.

import (
	"slices"
	"strconv"
	"strings"
)

type EspressoPLAOutputType uint16

const (
	EspressoPLAOutputOn EspressoPLAOutputType = 1 << iota
	EspressoPLAOutputDontCare
	EspressoPLAOutputOff
	EspressoPLAOutputPleasure
	EspressoPLAOutputEquation
	_
	_
	EspressoPLAOutputKISS
	EspressoPLAOutputConstraints
	EspressoPLAOutputSymbolicConstraints
)

type EspressoPLAOutputOptions struct {
	Phase                        []bool
	MaxCubes, MaxParts, MaxBytes int
	Stop                         func() bool
}

type espressoOutputWork struct {
	options EspressoPLAOutputOptions
	parts   int
	builder strings.Builder
}

func normalizeEspressoPLAOutputOptions(layout *EspressoCubeLayout, options EspressoPLAOutputOptions) (EspressoPLAOutputOptions, error) {
	if options.MaxCubes < 0 || options.MaxParts < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	output := layout.OutputVariable()
	if len(options.Phase) != 0 && (output < 0 || len(options.Phase) != layout.partSizes[output]) {
		return options, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxParts == 0 {
		options.MaxParts = 16 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	options.Phase = slices.Clone(options.Phase)
	return options, nil
}

func newEspressoOutputWork(layout *EspressoCubeLayout, options EspressoPLAOutputOptions) (*espressoOutputWork, error) {
	options, err := normalizeEspressoPLAOutputOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return &espressoOutputWork{options: options}, nil
}

func (work *espressoOutputWork) step(parts int) error {
	if parts < 0 || work.parts > work.options.MaxParts-parts {
		return ErrLimit
	}
	work.parts += parts
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *espressoOutputWork) write(text string) error {
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	if len(text) > work.options.MaxBytes-work.builder.Len() {
		return ErrLimit
	}
	work.builder.WriteString(text)
	return nil
}

func espressoOutputTypeLetters(outputType EspressoPLAOutputType) string {
	var result strings.Builder
	if outputType&EspressoPLAOutputOn != 0 {
		result.WriteByte('f')
	}
	if outputType&EspressoPLAOutputDontCare != 0 {
		result.WriteByte('d')
	}
	if outputType&EspressoPLAOutputOff != 0 {
		result.WriteByte('r')
	}
	return result.String()
}

func (pla *EspressoPLA) DefaultLabels() ([]string, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	labels := slices.Clone(pla.Labels)
	if len(labels) == 0 {
		labels = make([]string, pla.Layout.bitCount)
	}
	for variable, size := range pla.Layout.partSizes {
		for part := 0; part < size; part++ {
			index := pla.Layout.firstParts[variable] + part
			if labels[index] != "" {
				continue
			}
			if variable < pla.Layout.binaryVariables {
				labels[index] = "v" + strconv.Itoa(variable)
				if part == 0 {
					labels[index] += ".bar"
				}
			} else {
				labels[index] = "v" + strconv.Itoa(variable) + "." + strconv.Itoa(part)
			}
		}
	}
	return labels, nil
}

func (layout *EspressoCubeLayout) formatCubeInto(work *espressoOutputWork, cube *EspressoSet, outputMap string) error {
	if err := layout.checkSet(cube); err != nil {
		return err
	}
	if len(outputMap) != 2 {
		return ErrDimension
	}
	for variable := 0; variable < layout.binaryVariables; variable++ {
		if err := work.step(2); err != nil {
			return err
		}
		first := layout.firstParts[variable]
		low, _ := cube.Has(first)
		high, _ := cube.Has(first + 1)
		code := 0
		if low {
			code++
		}
		if high {
			code += 2
		}
		if err := work.write(string("?01-"[code])); err != nil {
			return err
		}
	}
	output := layout.OutputVariable()
	for variable := layout.binaryVariables; variable < len(layout.partSizes); variable++ {
		if variable == output {
			continue
		}
		if err := work.write(" "); err != nil {
			return err
		}
		for part := 0; part < layout.partSizes[variable]; part++ {
			if err := work.step(1); err != nil {
				return err
			}
			present, _ := cube.Has(layout.firstParts[variable] + part)
			if err := work.write(string("01"[btoi(present)])); err != nil {
				return err
			}
		}
	}
	if output >= 0 {
		if err := work.write(" "); err != nil {
			return err
		}
		for part := 0; part < layout.partSizes[output]; part++ {
			if err := work.step(1); err != nil {
				return err
			}
			present, _ := cube.Has(layout.firstParts[output] + part)
			if err := work.write(string(outputMap[btoi(present)])); err != nil {
				return err
			}
		}
	}
	return nil
}

func btoi(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (layout *EspressoCubeLayout) FormatCube(cube *EspressoSet, outputMap string, options EspressoPLAOutputOptions) (string, error) {
	work, err := newEspressoOutputWork(layout, options)
	if err != nil {
		return "", err
	}
	if err = layout.formatCubeInto(work, cube, outputMap); err != nil {
		return "", err
	}
	return work.builder.String(), nil
}

func (layout *EspressoCubeLayout) formatExpandedCubeInto(work *espressoOutputWork, cube *EspressoSet) error {
	if err := layout.checkSet(cube); err != nil {
		return err
	}
	output := layout.OutputVariable()
	for variable := 0; variable < layout.binaryVariables; variable++ {
		for part := 0; part < 2; part++ {
			if err := work.step(1); err != nil {
				return err
			}
			present, _ := cube.Has(layout.firstParts[variable] + part)
			if err := work.write(string("~1"[btoi(present)])); err != nil {
				return err
			}
		}
	}
	for variable := layout.binaryVariables; variable < len(layout.partSizes); variable++ {
		if variable == output {
			continue
		}
		for part := 0; part < layout.partSizes[variable]; part++ {
			if err := work.step(1); err != nil {
				return err
			}
			present, _ := cube.Has(layout.firstParts[variable] + part)
			if err := work.write(string("1~"[btoi(present)])); err != nil {
				return err
			}
		}
	}
	if output >= 0 {
		if err := work.write(" "); err != nil {
			return err
		}
		for part := 0; part < layout.partSizes[output]; part++ {
			if err := work.step(1); err != nil {
				return err
			}
			present, _ := cube.Has(layout.firstParts[output] + part)
			positive := len(work.options.Phase) == 0 || work.options.Phase[part]
			mapping := "~0"
			if positive {
				mapping = "~1"
			}
			if err := work.write(string(mapping[btoi(present)])); err != nil {
				return err
			}
		}
	}
	return nil
}

func (layout *EspressoCubeLayout) FormatExpandedCube(cube *EspressoSet, options EspressoPLAOutputOptions) (string, error) {
	work, err := newEspressoOutputWork(layout, options)
	if err != nil {
		return "", err
	}
	if err = layout.formatExpandedCubeInto(work, cube); err != nil {
		return "", err
	}
	return work.builder.String(), nil
}

func (pla *EspressoPLA) formatHeaderInto(work *espressoOutputWork, outputType EspressoPLAOutputType) error {
	layout := pla.Layout
	if outputType != EspressoPLAOutputOn {
		if err := work.write(".type " + espressoOutputTypeLetters(outputType) + "\n"); err != nil {
			return err
		}
	}
	output := layout.OutputVariable()
	if layout.MultipleValuedVariableCount() <= 1 {
		if err := work.write(".i " + strconv.Itoa(layout.binaryVariables) + "\n"); err != nil {
			return err
		}
		if output >= 0 {
			if err := work.write(".o " + strconv.Itoa(layout.partSizes[output]) + "\n"); err != nil {
				return err
			}
		}
	} else {
		header := ".mv " + strconv.Itoa(len(layout.partSizes)) + " " + strconv.Itoa(layout.binaryVariables)
		for variable := layout.binaryVariables; variable < len(layout.partSizes); variable++ {
			header += " " + strconv.Itoa(layout.partSizes[variable])
		}
		if err := work.write(header + "\n"); err != nil {
			return err
		}
	}
	if len(pla.Labels) != 0 && layout.binaryVariables > 0 && pla.Labels[1] != "" {
		line := ".ilb"
		for variable := 0; variable < layout.binaryVariables; variable++ {
			label := pla.Labels[layout.firstParts[variable]+1]
			if label == "" {
				label = "(null)"
			}
			line += " " + label
		}
		if err := work.write(line + "\n"); err != nil {
			return err
		}
	}
	if output >= 0 && len(pla.Labels) != 0 && pla.Labels[layout.firstParts[output]] != "" {
		line := ".ob"
		for part := 0; part < layout.partSizes[output]; part++ {
			label := pla.Labels[layout.firstParts[output]+part]
			if label == "" {
				label = "(null)"
			}
			line += " " + label
		}
		if err := work.write(line + "\n"); err != nil {
			return err
		}
	}
	for variable := layout.binaryVariables; variable < len(layout.partSizes); variable++ {
		if variable == output || len(pla.Labels) == 0 || pla.Labels[layout.firstParts[variable]] == "" {
			continue
		}
		line := ".label var=" + strconv.Itoa(variable)
		for part := 0; part < layout.partSizes[variable]; part++ {
			line += " " + pla.Labels[layout.firstParts[variable]+part]
		}
		if err := work.write(line + "\n"); err != nil {
			return err
		}
	}
	if len(work.options.Phase) != 0 {
		line := "#.phase "
		for _, positive := range work.options.Phase {
			line += string("01"[btoi(positive)])
		}
		if err := work.write(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func (pla *EspressoPLA) FormatHeader(outputType EspressoPLAOutputType, options EspressoPLAOutputOptions) (string, error) {
	if err := pla.check(); err != nil {
		return "", err
	}
	work, err := newEspressoOutputWork(pla.Layout, options)
	if err != nil {
		return "", err
	}
	if err = pla.formatHeaderInto(work, outputType); err != nil {
		return "", err
	}
	return work.builder.String(), nil
}

func (pla *EspressoPLA) formatFamilyInto(work *espressoOutputWork, family *EspressoSetFamily, outputMap string) error {
	if len(family.sets) > work.options.MaxCubes {
		return ErrLimit
	}
	for _, cube := range family.sets {
		if err := pla.Layout.formatCubeInto(work, cube, outputMap); err != nil {
			return err
		}
		if err := work.write("\n"); err != nil {
			return err
		}
	}
	return nil
}

func (pla *EspressoPLA) formatPleasureInto(work *espressoOutputWork) error {
	if pla.Layout.OutputVariable() < 0 {
		return ErrDimension
	}
	labels, err := pla.DefaultLabels()
	if err != nil {
		return err
	}
	if len(pla.On.sets) > work.options.MaxCubes {
		return ErrLimit
	}
	if err = work.write(".option unmerged\n.label"); err != nil {
		return err
	}
	column := 6
	for _, label := range labels {
		if column+len(label) > 75 {
			if err = work.write(" \\\n"); err != nil {
				return err
			}
			column = 0
		} else {
			if err = work.write(" "); err != nil {
				return err
			}
			column++
		}
		if err = work.write(label); err != nil {
			return err
		}
		column += len(label)
	}
	if err = work.write("\n.group"); err != nil {
		return err
	}
	column = 6
	for variable := 0; variable < len(pla.Layout.partSizes)-1; variable++ {
		if err = work.write(" ("); err != nil {
			return err
		}
		column += 2
		for part := 0; part < pla.Layout.partSizes[variable]; part++ {
			label := labels[pla.Layout.firstParts[variable]+part]
			if column+len(label) > 75 {
				if err = work.write(" \\\n"); err != nil {
					return err
				}
				column = 0
			} else if pla.Layout.firstParts[variable]+part != 0 {
				if err = work.write(" "); err != nil {
					return err
				}
				column++
			}
			if err = work.write(label); err != nil {
				return err
			}
			column += len(label)
		}
		if err = work.write(")"); err != nil {
			return err
		}
		column++
	}
	if err = work.write("\n.p " + strconv.Itoa(len(pla.On.sets)) + "\n"); err != nil {
		return err
	}
	for _, cube := range pla.On.sets {
		if err = pla.Layout.formatExpandedCubeInto(work, cube); err != nil {
			return err
		}
		if err = work.write("\n"); err != nil {
			return err
		}
	}
	return work.write(".end\n")
}

func (pla *EspressoPLA) formatEquationInto(work *espressoOutputWork) error {
	output := pla.Layout.OutputVariable()
	if output < 0 || pla.Layout.MultipleValuedVariableCount() != 1 {
		return ErrDimension
	}
	if len(pla.On.sets) > work.options.MaxCubes {
		return ErrLimit
	}
	labels, err := pla.DefaultLabels()
	if err != nil {
		return err
	}
	for outputPart := 0; outputPart < pla.Layout.partSizes[output]; outputPart++ {
		outputLabel := labels[pla.Layout.firstParts[output]+outputPart]
		if err = work.write(outputLabel + " = "); err != nil {
			return err
		}
		column := len(outputLabel) + 3
		firstProduct := true
		for _, cube := range pla.On.sets {
			present, _ := cube.Has(pla.Layout.firstParts[output] + outputPart)
			if !present {
				continue
			}
			if firstProduct {
				err = work.write("(")
				column++
			} else {
				err = work.write(" | (")
				column += 4
			}
			if err != nil {
				return err
			}
			firstProduct = false
			firstLiteral := true
			for variable := 0; variable < pla.Layout.binaryVariables; variable++ {
				if err = work.step(2); err != nil {
					return err
				}
				first := pla.Layout.firstParts[variable]
				low, _ := cube.Has(first)
				high, _ := cube.Has(first + 1)
				if low == high {
					continue
				}
				label := labels[first+1]
				if column+len(label) > 72 {
					if err = work.write("\n    "); err != nil {
						return err
					}
					column = 4
				}
				if !firstLiteral {
					if err = work.write("&"); err != nil {
						return err
					}
					column++
				}
				firstLiteral = false
				if low {
					if err = work.write("!"); err != nil {
						return err
					}
					column++
				}
				if err = work.write(label); err != nil {
					return err
				}
				column += len(label)
			}
			if err = work.write(")"); err != nil {
				return err
			}
			column++
		}
		if err = work.write(";\n\n"); err != nil {
			return err
		}
	}
	return nil
}

func (pla *EspressoPLA) formatKISSCubeInto(work *espressoOutputWork, labels []string, cube *EspressoSet, outputMap string) error {
	if len(outputMap) != 2 {
		return ErrDimension
	}
	for variable := 0; variable < pla.Layout.binaryVariables; variable++ {
		first := pla.Layout.firstParts[variable]
		low, _ := cube.Has(first)
		high, _ := cube.Has(first + 1)
		code := btoi(low) + 2*btoi(high)
		if err := work.step(2); err != nil {
			return err
		}
		if err := work.write(string("?01-"[code])); err != nil {
			return err
		}
	}
	output := pla.Layout.OutputVariable()
	for variable := pla.Layout.binaryVariables; variable < len(pla.Layout.partSizes); variable++ {
		if variable == output {
			continue
		}
		if err := work.write(" "); err != nil {
			return err
		}
		selected := -1
		count := 0
		for part := 0; part < pla.Layout.partSizes[variable]; part++ {
			if err := work.step(1); err != nil {
				return err
			}
			present, _ := cube.Has(pla.Layout.firstParts[variable] + part)
			if present {
				selected, count = part, count+1
			}
		}
		switch {
		case count == pla.Layout.partSizes[variable]:
			err := work.write("-")
			if err != nil {
				return err
			}
		case count == 0:
			if err := work.write("~"); err != nil {
				return err
			}
		case count == 1:
			if err := work.write(labels[pla.Layout.firstParts[variable]+selected]); err != nil {
				return err
			}
		default:
			return ErrInvariant
		}
	}
	if output >= 0 {
		if err := work.write(" "); err != nil {
			return err
		}
		for part := 0; part < pla.Layout.partSizes[output]; part++ {
			present, _ := cube.Has(pla.Layout.firstParts[output] + part)
			if err := work.write(string(outputMap[btoi(present)])); err != nil {
				return err
			}
		}
	}
	return work.write("\n")
}

func (pla *EspressoPLA) formatKISSInto(work *espressoOutputWork) error {
	if len(pla.On.sets) > work.options.MaxCubes || len(pla.DontCare.sets) > work.options.MaxCubes {
		return ErrLimit
	}
	labels, err := pla.DefaultLabels()
	if err != nil {
		return err
	}
	for _, cube := range pla.On.sets {
		if err = pla.formatKISSCubeInto(work, labels, cube, "~1"); err != nil {
			return err
		}
	}
	for _, cube := range pla.DontCare.sets {
		if err = pla.formatKISSCubeInto(work, labels, cube, "~2"); err != nil {
			return err
		}
	}
	return nil
}

func (pla *EspressoPLA) formatSymbolicConstraintsInto(work *espressoOutputWork, symbolic bool) error {
	labels, err := pla.DefaultLabels()
	if err != nil {
		return err
	}
	if len(pla.On.sets) > work.options.MaxCubes {
		return ErrLimit
	}
	output := pla.Layout.OutputVariable()
	for variable := pla.Layout.binaryVariables; variable < len(pla.Layout.partSizes); variable++ {
		if variable == output {
			continue
		}
		size := pla.Layout.partSizes[variable]
		unconstrained := 0
		patterns := make([]string, 0)
		weights := make([]int, 0)
		positions := make(map[string]int)
		for _, cube := range pla.On.sets {
			var bits strings.Builder
			count := 0
			for part := 0; part < size; part++ {
				if err = work.step(1); err != nil {
					return err
				}
				present, _ := cube.Has(pla.Layout.firstParts[variable] + part)
				bits.WriteByte("01"[btoi(present)])
				count += btoi(present)
			}
			if count == 1 || count == size {
				unconstrained++
				continue
			}
			pattern := bits.String()
			if index, exists := positions[pattern]; exists {
				weights[index]++
			} else {
				positions[pattern] = len(patterns)
				patterns = append(patterns, pattern)
				weights = append(weights, 1)
			}
		}
		if symbolic {
			if err = work.write("# Symbolic constraints for variable " + strconv.Itoa(variable) + " (Symbolic form)\n"); err != nil {
				return err
			}
			for index, pattern := range patterns {
				line := "#   w=" + strconv.Itoa(weights[index]) + ": ("
				for part, value := range []byte(pattern) {
					if value == '1' {
						line += " " + labels[pla.Layout.firstParts[variable]+part]
					}
				}
				if err = work.write(line + " )\n"); err != nil {
					return err
				}
			}
		} else {
			if err = work.write("# Symbolic constraints for variable " + strconv.Itoa(variable) + " (Numeric form)\n# unconstrained weight = " + strconv.Itoa(unconstrained) + "\nnum_codes=" + strconv.Itoa(size) + "\n"); err != nil {
				return err
			}
			for index, pattern := range patterns {
				line := "weight=" + strconv.Itoa(weights[index]) + ": "
				for part, value := range []byte(pattern) {
					if value == '1' {
						line += " " + strconv.Itoa(part)
					}
				}
				if err = work.write(line + "\n"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validEspressoPLAOutputType(outputType EspressoPLAOutputType) bool {
	const all = EspressoPLAOutputOn | EspressoPLAOutputDontCare | EspressoPLAOutputOff | EspressoPLAOutputPleasure | EspressoPLAOutputEquation | EspressoPLAOutputKISS | EspressoPLAOutputConstraints | EspressoPLAOutputSymbolicConstraints
	if outputType == 0 || outputType&^all != 0 {
		return false
	}
	special := outputType & (EspressoPLAOutputPleasure | EspressoPLAOutputEquation | EspressoPLAOutputKISS)
	return special == 0 || special == EspressoPLAOutputPleasure || special == EspressoPLAOutputEquation || special == EspressoPLAOutputKISS
}

func (pla *EspressoPLA) MarshalPLA(outputType EspressoPLAOutputType, options EspressoPLAOutputOptions) ([]byte, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	if !validEspressoPLAOutputType(outputType) {
		return nil, ErrDimension
	}
	work, err := newEspressoOutputWork(pla.Layout, options)
	if err != nil {
		return nil, err
	}
	remaining := outputType
	if remaining&EspressoPLAOutputConstraints != 0 {
		if err = pla.formatSymbolicConstraintsInto(work, false); err != nil {
			return nil, err
		}
		remaining &^= EspressoPLAOutputConstraints
	}
	if remaining&EspressoPLAOutputSymbolicConstraints != 0 {
		if err = pla.formatSymbolicConstraintsInto(work, true); err != nil {
			return nil, err
		}
		remaining &^= EspressoPLAOutputSymbolicConstraints
	}
	if remaining == 0 {
		return []byte(work.builder.String()), nil
	}
	switch remaining {
	case EspressoPLAOutputPleasure:
		err = pla.formatPleasureInto(work)
	case EspressoPLAOutputEquation:
		err = pla.formatEquationInto(work)
	case EspressoPLAOutputKISS:
		err = pla.formatKISSInto(work)
	default:
		if remaining&^(EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff) != 0 {
			return nil, ErrDimension
		}
		if err = pla.formatHeaderInto(work, remaining); err != nil {
			return nil, err
		}
		count := 0
		if remaining&EspressoPLAOutputOn != 0 {
			count += len(pla.On.sets)
		}
		if remaining&EspressoPLAOutputDontCare != 0 {
			count += len(pla.DontCare.sets)
		}
		if remaining&EspressoPLAOutputOff != 0 {
			count += len(pla.Off.sets)
		}
		if count > work.options.MaxCubes {
			return nil, ErrLimit
		}
		if err = work.write(".p " + strconv.Itoa(count) + "\n"); err != nil {
			return nil, err
		}
		if remaining == EspressoPLAOutputOn {
			if err = pla.formatFamilyInto(work, pla.On, "01"); err == nil {
				err = work.write(".e\n")
			}
		} else {
			if remaining&EspressoPLAOutputOn != 0 {
				err = pla.formatFamilyInto(work, pla.On, "~1")
			}
			if err == nil && remaining&EspressoPLAOutputDontCare != 0 {
				err = pla.formatFamilyInto(work, pla.DontCare, "~2")
			}
			if err == nil && remaining&EspressoPLAOutputOff != 0 {
				err = pla.formatFamilyInto(work, pla.Off, "~0")
			}
			if err == nil {
				err = work.write(".end\n")
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return []byte(work.builder.String()), nil
}

func (layout *EspressoCubeLayout) FormatCubeListDebug(list *EspressoCubeList, name string, level int, verbose bool, options EspressoPLAOutputOptions) (string, error) {
	if level < 0 || errStringToken(name) != nil {
		return "", ErrDimension
	}
	if err := list.check(); err != nil || list.layout.bitCount != layout.bitCount {
		return "", ErrDimension
	}
	work, err := newEspressoOutputWork(layout, options)
	if err != nil {
		return "", err
	}
	if len(list.cubes) > work.options.MaxCubes {
		return "", ErrLimit
	}
	if verbose && level == 0 {
		if err = work.write("\n"); err != nil {
			return "", err
		}
	}
	if err = work.write(name + "[" + strconv.Itoa(level) + "]: ord(T)=" + strconv.Itoa(len(list.cubes)) + "\n"); err != nil {
		return "", err
	}
	if verbose {
		if err = work.write("cofactor="); err != nil {
			return "", err
		}
		if err = layout.formatCubeInto(work, list.cofactor, "01"); err != nil {
			return "", err
		}
		if err = work.write("\n"); err != nil {
			return "", err
		}
		for index, cube := range list.cubes {
			combined, _ := cube.Union(list.cofactor)
			if err = work.write(espressoDebugOrdinal(index + 1)); err != nil {
				return "", err
			}
			if err = layout.formatCubeInto(work, combined, "01"); err != nil {
				return "", err
			}
			if err = work.write("\n"); err != nil {
				return "", err
			}
		}
	}
	return work.builder.String(), nil
}

func espressoDebugOrdinal(value int) string {
	digits := strconv.Itoa(value)
	if len(digits) < 4 {
		digits = strings.Repeat(" ", 4-len(digits)) + digits
	}
	return digits + ". "
}

func errStringToken(value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return ErrDimension
	}
	return nil
}

func (layout *EspressoCubeLayout) FormatCoverDebug(family *EspressoSetFamily, name string, number int, verbose bool, options EspressoPLAOutputOptions) (string, error) {
	if number < 0 || errStringToken(name) != nil {
		return "", ErrDimension
	}
	if err := family.check(); err != nil || family.bits != layout.bitCount {
		return "", ErrDimension
	}
	work, err := newEspressoOutputWork(layout, options)
	if err != nil {
		return "", err
	}
	if len(family.sets) > work.options.MaxCubes {
		return "", ErrLimit
	}
	if verbose && number == 0 {
		if err = work.write("\n"); err != nil {
			return "", err
		}
	}
	if err = work.write(name + "[" + strconv.Itoa(number) + "]: ord(T)=" + strconv.Itoa(len(family.sets)) + "\n"); err != nil {
		return "", err
	}
	if verbose {
		for index, cube := range family.sets {
			if err = work.write(espressoDebugOrdinal(index + 1)); err != nil {
				return "", err
			}
			if err = layout.formatCubeInto(work, cube, "01"); err != nil {
				return "", err
			}
			if err = work.write("\n"); err != nil {
				return "", err
			}
		}
	}
	return work.builder.String(), nil
}

func (layout *EspressoCubeLayout) FormatCoverCubes(family *EspressoSetFamily, options EspressoPLAOutputOptions) (string, error) {
	if err := family.check(); err != nil || family.bits != layout.bitCount {
		return "", ErrDimension
	}
	work, err := newEspressoOutputWork(layout, options)
	if err != nil {
		return "", err
	}
	if err = (&EspressoPLA{Layout: layout, On: family, DontCare: family, Off: family}).formatFamilyInto(work, family, "01"); err != nil {
		return "", err
	}
	return work.builder.String(), nil
}
