package abc

// Six-variable truth decomposition from opt/dau/dauDsd.c:Dau_DsdPerform.

import "strings"

type dauDSD6Decomposer struct {
	options DauDSDOptions
}

// DauDSDFromTruth6 decomposes an ABC-order six-variable truth word into a DSD
// formula. The result is owned, deterministic, bounded, and accepted by
// ParseDauDSD. It preserves the source routine's variable and case order.
func DauDSDFromTruth6(truth uint64, options DauDSDOptions) (string, error) {
	o, err := normalizeDauDSDOptions(options)
	if err != nil {
		return "", err
	}
	if o.Stop != nil && o.Stop() {
		return "", ErrCancelled
	}
	if truth == 0 {
		return "0", nil
	}
	if truth == ^uint64(0) {
		return "1", nil
	}
	d := dauDSD6Decomposer{options: o}
	result, err := d.perform(truth, []int{0, 1, 2, 3, 4, 5}, 0)
	if err != nil {
		return "", err
	}
	if len(result) > o.MaxBytes {
		return "", ErrLimit
	}
	return result, nil
}

func dauDSD6Cofactor(truth uint64, variable int, value bool) uint64 {
	var result uint64
	mask := uint64(1) << uint(variable)
	for assignment := uint64(0); assignment < 64; assignment++ {
		source := assignment &^ mask
		if value {
			source |= mask
		}
		if truth&(uint64(1)<<source) != 0 {
			result |= uint64(1) << assignment
		}
	}
	return result
}

func dauDSD6HasVariable(truth uint64, variable int) bool {
	return dauDSD6Cofactor(truth, variable, false) != dauDSD6Cofactor(truth, variable, true)
}

func (d *dauDSD6Decomposer) checked(text string) (string, error) {
	if len(text) > d.options.MaxBytes {
		return "", ErrLimit
	}
	return text, nil
}

func (d *dauDSD6Decomposer) perform(truth uint64, variables []int, depth int) (string, error) {
	if d.options.Stop != nil && d.options.Stop() {
		return "", ErrCancelled
	}
	if depth >= d.options.MaxDepth {
		return "", ErrLimit
	}
	active := make([]int, 0, len(variables))
	for _, variable := range variables {
		if dauDSD6HasVariable(truth, variable) {
			active = append(active, variable)
		}
	}
	if len(active) == 0 {
		if truth == 0 {
			return "0", nil
		}
		if truth == ^uint64(0) {
			return "1", nil
		}
		return "", ErrInvariant
	}
	if len(active) == 1 {
		variable := active[0]
		pattern := truthVariablePattern(variable)
		if truth == pattern {
			return string(byte('a' + variable)), nil
		}
		if truth == ^pattern {
			return "!" + string(byte('a'+variable)), nil
		}
		return "", ErrInvariant
	}

	cofactor0 := make([]uint64, len(active))
	cofactor1 := make([]uint64, len(active))
	for i, variable := range active {
		cofactor0[i] = dauDSD6Cofactor(truth, variable, false)
		cofactor1[i] = dauDSD6Cofactor(truth, variable, true)
		letter := string(byte('a' + variable))
		var nested string
		var err error
		switch {
		case cofactor0[i] == 0:
			nested, err = d.perform(cofactor1[i], active, depth+1)
			if err == nil {
				nested = "(" + letter + nested + ")"
			}
		case cofactor0[i] == ^uint64(0):
			nested, err = d.perform(^cofactor1[i], active, depth+1)
			if err == nil {
				nested = "!(" + letter + nested + ")"
			}
		case cofactor1[i] == 0:
			nested, err = d.perform(cofactor0[i], active, depth+1)
			if err == nil {
				nested = "(!" + letter + nested + ")"
			}
		case cofactor1[i] == ^uint64(0):
			nested, err = d.perform(^cofactor0[i], active, depth+1)
			if err == nil {
				nested = "!(!" + letter + nested + ")"
			}
		case cofactor0[i] == ^cofactor1[i]:
			nested, err = d.perform(cofactor0[i], active, depth+1)
			if err == nil {
				nested = "[" + letter + nested + "]"
			}
		default:
			continue
		}
		if err != nil {
			return "", err
		}
		return d.checked(nested)
	}

	for vi, v := range active {
		for ui := vi + 1; ui < len(active); ui++ {
			u := active[ui]
			cof := [4]uint64{
				dauDSD6Cofactor(cofactor0[vi], u, false),
				dauDSD6Cofactor(cofactor0[vi], u, true),
				dauDSD6Cofactor(cofactor1[vi], u, false),
				dauDSD6Cofactor(cofactor1[vi], u, true),
			}
			var reduced uint64
			var replacement string
			pattern := truthVariablePattern(u)
			switch {
			case cof[0] == cof[1] && cof[0] == cof[2]:
				replacement = "(" + string(byte('a'+v)) + string(byte('a'+u)) + ")"
				reduced = pattern&cof[3] | ^pattern&cof[0]
			case cof[0] == cof[1] && cof[0] == cof[3]:
				replacement = "(" + string(byte('a'+v)) + "!" + string(byte('a'+u)) + ")"
				reduced = pattern&cof[2] | ^pattern&cof[0]
			case cof[0] == cof[2] && cof[0] == cof[3]:
				replacement = "(!" + string(byte('a'+v)) + string(byte('a'+u)) + ")"
				reduced = pattern&cof[1] | ^pattern&cof[0]
			case cof[1] == cof[2] && cof[1] == cof[3]:
				replacement = "(!" + string(byte('a'+v)) + "!" + string(byte('a'+u)) + ")"
				reduced = pattern&cof[0] | ^pattern&cof[1]
			case cof[0] == cof[3] && cof[1] == cof[2]:
				replacement = "[" + string(byte('a'+v)) + string(byte('a'+u)) + "]"
				reduced = pattern&cof[1] | ^pattern&cof[0]
			default:
				continue
			}
			nested, err := d.perform(reduced, active, depth+1)
			if err != nil {
				return "", err
			}
			nested = strings.ReplaceAll(nested, string(byte('a'+u)), replacement)
			return d.checked(nested)
		}
	}

	best, bestCount := -1, 10
	for i, variable := range active {
		count := 0
		for _, other := range active {
			if other != variable && dauDSD6HasVariable(cofactor0[i], other) && dauDSD6HasVariable(cofactor1[i], other) {
				count++
			}
		}
		if count < bestCount {
			best, bestCount = i, count
		}
		if count == 0 {
			break
		}
	}
	if best < 0 {
		return "", ErrInvariant
	}
	whenTrue, err := d.perform(cofactor1[best], active, depth+1)
	if err != nil {
		return "", err
	}
	whenFalse, err := d.perform(cofactor0[best], active, depth+1)
	if err != nil {
		return "", err
	}
	return d.checked("<" + string(byte('a'+active[best])) + whenTrue + whenFalse + ">")
}
