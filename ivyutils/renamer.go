package ivyutils

import "strings"

// UniqueRenamer generates unique names by appending suffixes when needed.
type UniqueRenamer struct {
	Prefix string
	Used   map[string]struct{}
}

// NewUniqueRenamer creates a UniqueRenamer with optional prefix and used names.
func NewUniqueRenamer(prefix string, used []string) *UniqueRenamer {
	u := make(map[string]struct{}, len(used))
	for _, s := range used {
		u[s] = struct{}{}
	}
	return &UniqueRenamer{Prefix: prefix, Used: u}
}

// Rename returns a unique name based on the given name.
// If name is non-empty and prefix+name is unused, returns it directly.
// Otherwise generates a new unique name using suffixes.
func (r *UniqueRenamer) Rename(name string) string {
	thing := r.Prefix + name
	var res string
	if name != "" {
		if _, used := r.Used[thing]; !used {
			res = thing
		} else {
			res = UnusedNameWithBase(thing, r.Used)
		}
	} else {
		res = UnusedNameWithBase(thing, r.Used)
	}
	r.Used[res] = struct{}{}
	return res
}

// VariableGenerator generates unique variable names (A, B, ..., Z, AA, BB, ...).
type VariableGenerator struct {
	Used map[string]struct{}
}

// NewVariableGenerator creates a new VariableGenerator.
func NewVariableGenerator() *VariableGenerator {
	return &VariableGenerator{Used: make(map[string]struct{})}
}

// Generate returns a unique variable name. If name is provided, starts from
// its first character uppercased.
func (g *VariableGenerator) Generate(name string) string {
	firstChar := 'A'
	if len(name) > 0 {
		c := rune(name[0])
		if c >= 'a' && c <= 'z' {
			firstChar = c - 32
		} else if c >= 'A' && c <= 'Z' {
			firstChar = c
		}
	}
	nchars := 1
	ch := firstChar
	for {
		for ch <= 'Z' {
			guess := strings.Repeat(string(ch), nchars)
			if ch != 'O' {
				if _, used := g.Used[guess]; !used {
					g.Used[guess] = struct{}{}
					return guess
				}
			}
			ch++
		}
		ch = 'A'
		nchars++
	}
}

// ConstantNameGenerator yields names: a-z, then a0-z0, a1-z1, etc.
func ConstantNameGenerator() func() string {
	phase := 0 // 0 = first pass (a-z), 1+ = suffixed passes
	idx := 0
	return func() string {
		if phase == 0 {
			c := 'a' + rune(idx)
			idx++
			if idx >= 26 {
				phase = 1
				idx = 0
			}
			return string(c)
		}
		suffix := phase - 1
		c := 'a' + rune(idx)
		idx++
		if idx >= 26 {
			phase++
			idx = 0
		}
		return string(c) + itoa(suffix)
	}
}

// UnusedNameWithBase returns base + "_" + suffix that's not in usedNames.
func UnusedNameWithBase(base string, usedNames map[string]struct{}) string {
	gen := ConstantNameGenerator()
	for {
		name := base + "_" + gen()
		if _, used := usedNames[name]; !used {
			return name
		}
	}
}

// DistinctRenaming maps names1 to distinct names avoiding names2.
func DistinctRenaming(names1, names2 []string) map[string]string {
	rn := NewUniqueRenamer("", names2)
	result := make(map[string]string, len(names1))
	for _, s := range names1 {
		result[s] = rn.Rename(s)
	}
	return result
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
