package startingpoint

import "regexp"

var expressionIntegerSuffix = regexp.MustCompile(`\b(0[xX][0-9a-fA-F]+|0[bB][01]+|[0-9]+)([uU](?:ll|LL|[lL])?|(?:ll|LL|[lL])[uU]?)\b`)

func lowerIntegerSuffixes(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	mask := Mask(ts)
	var edits []Edit
	for _, m := range expressionIntegerSuffix.FindAllStringSubmatchIndex(mask, -1) {
		if m[0] > 0 && source[m[0]-1] == '.' {
			continue
		}
		typ, value, ok := suffixedInteger(source[m[2]:m[3]], source[m[4]:m[5]])
		if ok {
			edits = append(edits, Edit{m[0], m[1], typ + "(" + value + ")"})
		}
	}
	return ApplyEdits(source, edits)
}
