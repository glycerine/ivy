package startingpoint

import (
	"regexp"
	"strings"
)

var typedefName = regexp.MustCompile(`\btypedef\s+[^;{}]+\b([A-Za-z_]\w*)\s*;`)

// declarationTypes learns ordinary named types from signatures and simple
// typedefs, without treating arbitrary multiplication expressions as declarations.
func declarationTypes(mask string, signatures [][]int) map[string]bool {
	types := map[string]bool{"FILE": true, "int": true, "unsigned": true, "char": true, "short": true, "long": true, "float": true, "double": true, "bool": true, "sdbl_t": true}
	add := func(text string) {
		text = strings.ReplaceAll(text, "*", " ")
		for _, word := range strings.Fields(text) {
			switch word {
			case "static", "inline", "extern", "const", "volatile", "void", "struct", "union", "enum":
				continue
			}
			types[word] = true
		}
	}
	for _, loc := range signatures {
		add(mask[loc[2]:loc[3]])
		args, err := stripConstQualifiers(mask[loc[6]:loc[7]])
		if err != nil {
			continue
		}
		arguments, _ := splitExpressions(args)
		for _, arg := range arguments {
			if match := namedType.FindStringSubmatch(arg); match != nil {
				add(match[1])
			}
		}
	}
	for _, match := range typedefName.FindAllStringSubmatch(mask, -1) {
		types[match[1]] = true
	}
	return types
}

// Header-defined UpperCamelCase names (for example ProgressBar) are not always
// visible in this translation unit. Recognize that common type spelling as a
// draft heuristic, in addition to known names and C's conventional _t suffix.
// Requiring mixed case and another capital avoids treating ordinary lowercase
// operands or ALL_CAPS constants as types. Full C name resolution is out of scope.
func isDeclarationType(name string, types map[string]bool) bool {
	if types[name] || strings.HasSuffix(name, "_t") {
		return true
	}
	if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
		return false
	}
	lower, capital := false, false
	for i := 1; i < len(name); i++ {
		lower = lower || (name[i] >= 'a' && name[i] <= 'z')
		capital = capital || (name[i] >= 'A' && name[i] <= 'Z')
	}
	return lower && capital
}
