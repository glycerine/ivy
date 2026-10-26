package abc

// Variable permutation from opt/dau/dauDsd.c:Dau_DsdPermute.

import "strings"

// DauDSDVariableCount preserves Dau_DsdFindVarNum's lexical convention: the
// result is one plus the largest variable letter and is one when none occurs.
func DauDSDVariableCount(text string, maxBytes int) (int, error) {
	if maxBytes < 0 {
		return 0, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 2000
	}
	if len(text) > maxBytes {
		return 0, ErrLimit
	}
	count := 1
	for i := 0; i < len(text); i++ {
		if text[i] >= 'a' && text[i] <= 'z' && int(text[i]-'a')+1 > count {
			count = int(text[i]-'a') + 1
		}
	}
	return count, nil
}

// DauDSDPermute applies old-variable-to-new-variable mapping exactly once to
// every variable symbol. Input and result are independently owned strings.
func DauDSDPermute(text string, permutation []int, options DauDSDOptions) (string, error) {
	o, err := normalizeDauDSDOptions(options)
	if err != nil {
		return "", err
	}
	if o.Stop != nil && o.Stop() {
		return "", ErrCancelled
	}
	if _, err = ParseDauDSD(text, o); err != nil {
		return "", err
	}
	variables, err := DauDSDVariableCount(text, o.MaxBytes)
	if err != nil {
		return "", err
	}
	if len(permutation) != variables {
		return "", ErrDimension
	}
	seen := make([]bool, variables)
	for _, target := range permutation {
		if target < 0 || target >= variables || seen[target] {
			return "", ErrDimension
		}
		seen[target] = true
	}
	var result strings.Builder
	result.Grow(len(text))
	for i := 0; i < len(text); i++ {
		if o.Stop != nil && o.Stop() {
			return "", ErrCancelled
		}
		c := text[i]
		if c >= 'a' && c < 'a'+byte(variables) {
			c = 'a' + byte(permutation[int(c-'a')])
		}
		result.WriteByte(c)
	}
	return result.String(), nil
}

func dauDSDSplitMix64(state *uint64) uint64 {
	*state += 0x9e3779b97f4a7c15
	z := *state
	z = (z ^ z>>30) * 0xbf58476d1ce4e5b9
	z = (z ^ z>>27) * 0x94d049bb133111eb
	return z ^ z>>31
}

// DauDSDRandomPermutation uses an explicit seed and a specified SplitMix64
// stream, avoiding the source routine's process-global, libc-dependent rand().
// Its swap schedule otherwise matches Dau_DsdGenRandPerm.
func DauDSDRandomPermutation(text string, seed uint64, options DauDSDOptions) (string, []int, error) {
	o, err := normalizeDauDSDOptions(options)
	if err != nil {
		return "", nil, err
	}
	variables, err := DauDSDVariableCount(text, o.MaxBytes)
	if err != nil {
		return "", nil, err
	}
	permutation := make([]int, variables)
	for i := range permutation {
		permutation[i] = i
	}
	state := seed
	for i := range permutation {
		if o.Stop != nil && o.Stop() {
			return "", nil, ErrCancelled
		}
		target := int(dauDSDSplitMix64(&state) % uint64(variables))
		permutation[i], permutation[target] = permutation[target], permutation[i]
	}
	result, err := DauDSDPermute(text, permutation, o)
	if err != nil {
		return "", nil, err
	}
	return result, permutation, nil
}
