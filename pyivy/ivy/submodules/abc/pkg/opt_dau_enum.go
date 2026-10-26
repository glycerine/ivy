package abc

// Bounded DSD enumeration from opt/dau/dauEnum.c.

import (
	"strconv"
	"strings"
)

type DauEnumOptions struct {
	MaxResults, MaxFormulaBytes int
	Stop                        func() bool
}

func normalizeDauEnumOptions(o DauEnumOptions) (DauEnumOptions, error) {
	if o.MaxResults < 0 || o.MaxFormulaBytes < 0 {
		return o, ErrDimension
	}
	if o.MaxResults == 0 {
		o.MaxResults = 1 << 20
	}
	if o.MaxFormulaBytes == 0 {
		o.MaxFormulaBytes = 2000
	}
	return o, nil
}

// DauEnumLift shifts every variable letter and returns independent storage.
func DauEnumLift(name string, shift int, maxBytes int) (string, error) {
	if shift < 0 || maxBytes < 0 {
		return "", ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 2000
	}
	if len(name) > maxBytes {
		return "", ErrLimit
	}
	result := []byte(name)
	for i, c := range result {
		if c >= 'a' && c <= 'z' {
			if int(c-'a')+shift >= 26 {
				return "", ErrDimension
			}
			result[i] += byte(shift)
		}
	}
	return string(result), nil
}

func dauEnumUnstar(name string) string {
	if strings.HasPrefix(name, "*") {
		return name[1:]
	}
	return name
}

func DauEnumCombineTwo(star, xor bool, name1, name2 string, shift2 int, complement1, complement2 bool, maxBytes int) (string, error) {
	shifted, err := DauEnumLift(name2, shift2, maxBytes)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	if star {
		result.WriteByte('*')
	}
	if xor {
		result.WriteByte('[')
	} else {
		result.WriteByte('(')
	}
	if complement1 {
		result.WriteByte('!')
	}
	result.WriteString(dauEnumUnstar(name1))
	if complement2 {
		result.WriteByte('!')
	}
	result.WriteString(dauEnumUnstar(shifted))
	if xor {
		result.WriteByte(']')
	} else {
		result.WriteByte(')')
	}
	if maxBytes > 0 && result.Len() > maxBytes {
		return "", ErrLimit
	}
	return result.String(), nil
}

func DauEnumCombineThree(star bool, control, whenTrue, whenFalse string, shiftTrue, shiftFalse int, complementControl, complementTrue, complementFalse bool, maxBytes int) (string, error) {
	shiftedTrue, err := DauEnumLift(whenTrue, shiftTrue, maxBytes)
	if err != nil {
		return "", err
	}
	shiftedFalse, err := DauEnumLift(whenFalse, shiftFalse, maxBytes)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	if star {
		result.WriteByte('*')
	}
	result.WriteByte('<')
	for _, part := range []struct {
		text       string
		complement bool
	}{{control, complementControl}, {shiftedTrue, complementTrue}, {shiftedFalse, complementFalse}} {
		if part.complement {
			result.WriteByte('!')
		}
		result.WriteString(dauEnumUnstar(part.text))
	}
	result.WriteByte('>')
	if maxBytes > 0 && result.Len() > maxBytes {
		return "", ErrLimit
	}
	return result.String(), nil
}

// DauEnumerateDSD returns the source enumeration sets in source order. Index i
// contains formulas using exactly i variables; index zero is empty.
func DauEnumerateDSD(maxVariables int, options DauEnumOptions) ([][]string, error) {
	if maxVariables < 0 || maxVariables > 26 {
		return nil, ErrDimension
	}
	o, err := normalizeDauEnumOptions(options)
	if err != nil {
		return nil, err
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	sets := make([][]string, maxVariables+1)
	total := 0
	if maxVariables >= 1 {
		sets[1] = []string{"*a"}
		total = 1
	}
	appendFormula := func(destination *[]string, formula string, err error) error {
		if err != nil {
			return err
		}
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if total >= o.MaxResults {
			return ErrLimit
		}
		*destination = append(*destination, formula)
		total++
		return nil
	}
	for variables := 2; variables <= maxVariables; variables++ {
		one := make([]string, 0)
		addTwo := func(star, xor bool, first, second string, shift int, complementFirst, complementSecond bool) error {
			formula, combineErr := DauEnumCombineTwo(star, xor, first, second, shift, complementFirst, complementSecond, o.MaxFormulaBytes)
			return appendFormula(&one, formula, combineErr)
		}
		addThree := func(star bool, control, whenTrue, whenFalse string, shiftTrue, shiftFalse int, complementControl, complementTrue, complementFalse bool) error {
			formula, combineErr := DauEnumCombineThree(star, control, whenTrue, whenFalse, shiftTrue, shiftFalse, complementControl, complementTrue, complementFalse, o.MaxFormulaBytes)
			return appendFormula(&one, formula, combineErr)
		}
		for firstSize := 1; firstSize < variables; firstSize++ {
			for secondSize := firstSize; secondSize < variables; secondSize++ {
				if firstSize+secondSize != variables {
					continue
				}
				for firstIndex, first := range sets[firstSize] {
					for secondIndex, second := range sets[secondSize] {
						if err := addTwo(false, false, first, second, firstSize, false, false); err != nil {
							return nil, err
						}
						if !strings.HasPrefix(first, "*") {
							if err := addTwo(false, false, first, second, firstSize, true, false); err != nil {
								return nil, err
							}
						}
						if !strings.HasPrefix(second, "*") && !(firstSize == secondSize && firstIndex == secondIndex) {
							if err := addTwo(false, false, first, second, firstSize, false, true); err != nil {
								return nil, err
							}
						}
						if !strings.HasPrefix(first, "*") && !strings.HasPrefix(second, "*") {
							if err := addTwo(false, false, first, second, firstSize, true, true); err != nil {
								return nil, err
							}
						}
						star := strings.HasPrefix(first, "*") || strings.HasPrefix(second, "*")
						if err := addTwo(star, true, first, second, firstSize, false, false); err != nil {
							return nil, err
						}
					}
				}
			}
		}
		for controlSize := 1; controlSize < variables; controlSize++ {
			for trueSize := 1; trueSize < variables; trueSize++ {
				for falseSize := trueSize; falseSize < variables; falseSize++ {
					if controlSize+trueSize+falseSize != variables {
						continue
					}
					for _, control := range sets[controlSize] {
						for trueIndex, whenTrue := range sets[trueSize] {
							for falseIndex, whenFalse := range sets[falseSize] {
								star := strings.HasPrefix(whenTrue, "*") && strings.HasPrefix(whenFalse, "*")
								combine := func(cc, ct, cf bool) error {
									return addThree(star, control, whenTrue, whenFalse, controlSize, controlSize+trueSize, cc, ct, cf)
								}
								if err := combine(false, false, false); err != nil {
									return nil, err
								}
								if !strings.HasPrefix(whenTrue, "*") {
									if err := combine(false, true, false); err != nil {
										return nil, err
									}
								}
								if !strings.HasPrefix(whenFalse, "*") && !(trueSize == falseSize && trueIndex == falseIndex) {
									if err := combine(false, false, true); err != nil {
										return nil, err
									}
								}
								if !strings.HasPrefix(control, "*") && !(trueSize == falseSize && trueIndex == falseIndex) {
									if err := combine(true, false, false); err != nil {
										return nil, err
									}
									if !strings.HasPrefix(whenTrue, "*") {
										if err := combine(true, true, false); err != nil {
											return nil, err
										}
									}
									if !strings.HasPrefix(whenFalse, "*") {
										if err := combine(true, false, true); err != nil {
											return nil, err
										}
									}
								}
							}
						}
					}
				}
			}
		}
		sets[variables] = one
	}
	return sets, nil
}

func FormatDauEnumeration(sets [][]string) string {
	var output strings.Builder
	for variables, formulas := range sets {
		output.WriteString("VARIABLE NUMBER ")
		output.WriteString(strconv.Itoa(variables))
		output.WriteString(":\n")
		for _, formula := range formulas {
			output.WriteString(formula)
			output.WriteByte('\n')
		}
	}
	return output.String()
}
