package startingpoint

import (
	"regexp"
	"strconv"
	"strings"
)

var objectDefine = regexp.MustCompile(`^#\s*define[ \t]+([A-Za-z_]\w*)([ \t]+)([\s\S]*)$`)
var integerLiteral = regexp.MustCompile(`^\s*([+-]?\s*(?:0[xX][0-9a-fA-F]+|0[bB][01]+|0[0-7]*|[1-9][0-9]*))((?:[uU](?:ll|LL|[lL])?|(?:ll|LL|[lL])[uU]?)?)\s*$`)

var decimalFloatLiteral = regexp.MustCompile(`^\s*([+-]?\s*(?:(?:[0-9]+\.[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?|[0-9]+[eE][+-]?[0-9]+))([fF]?)\s*$`)

// integerDefines runs after C qualifier stripping so generated Go const keywords
// survive. Complete integer, decimal, scientific and string literals convert;
// expressions and function-like macros are left for their respective passes.
func integerDefines(source string) (string, error) {
	if !strings.Contains(source, "#") {
		return source, nil
	}
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for _, t := range ts {
		if t.Kind != "directive" {
			continue
		}
		parts := objectDefine.FindStringSubmatch(t.Text)
		if parts == nil {
			continue
		}
		value := parts[3]
		tokens, err := Lex(value)
		if err != nil {
			continue
		}
		if stringValue, ok := stringDefineValue(value, tokens); ok {
			edits = append(edits, Edit{t.Start, t.End, "const " + parts[1] + " =" + parts[2] + stringValue})
			continue
		}
		literal := integerLiteral.FindStringSubmatchIndex(Mask(tokens))
		typ, replacement := "", ""
		if literal != nil {
			var ok bool
			typ, replacement, ok = suffixedInteger(value[literal[2]:literal[3]], value[literal[4]:literal[5]])
			if !ok {
				continue
			}
		} else {
			literal = decimalFloatLiteral.FindStringSubmatchIndex(Mask(tokens))
			if literal == nil {
				continue
			}
			replacement = value[literal[2]:literal[3]]
			if value[literal[4]:literal[5]] != "" {
				typ = "float32"
			}
		}

		value = value[:literal[2]] + replacement + value[literal[5]:]
		if typ != "" {
			typ = " " + typ
		}
		edits = append(edits, Edit{t.Start, t.End, "const " + parts[1] + typ + " =" + parts[2] + value})
	}
	return ApplyEdits(source, edits)
}

// The current ABC target is LP64: int=32, long=64, long long=64.
// Untyped Go constants remain the default for literals without a C suffix.
func suffixedInteger(number, suffix string) (typ, replacement string, ok bool) {
	if suffix == "" {
		return "", number, true
	}
	text := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(number), " ", ""), "\t", "")
	negative := strings.HasPrefix(text, "-")
	digits := strings.TrimLeft(text, "+-")
	n, err := strconv.ParseUint(digits, 0, 64)
	if err != nil {
		return "", "", false
	}
	suffix = strings.ToLower(suffix)
	unsigned := strings.Contains(suffix, "u")
	if unsigned {
		typ = "uint64"
		if !strings.Contains(suffix, "l") && n <= uint64(^uint32(0)) {
			typ = "uint32"
		}
	} else {
		typ = "int64"
		if n > uint64(1<<63-1) {
			// Hex/octal/binary L literals may select unsigned long on LP64.
			if len(digits) > 1 && digits[0] == '0' {
				typ = "uint64"
				unsigned = true
			} else {
				return "", "", false
			}
		}
	}
	replacement = number
	if negative && unsigned {
		wrapped := uint64(0) - n
		if typ == "uint32" {
			wrapped = uint64(uint32(wrapped))
		}
		replacement = strconv.FormatUint(wrapped, 10)
	}
	return typ, replacement, true
}

// Adjacent C string literals concatenate. Insert + between tokens, leaving
// original literal spelling and intervening comments untouched.
func stringDefineValue(source string, tokens []Token) (string, bool) {
	var literals []Token
	for _, t := range tokens {
		if t.Kind == "space" || t.Kind == "comment" {
			continue
		}
		if t.Kind != "string" || !strings.HasPrefix(t.Text, `"`) {
			return "", false
		}
		literals = append(literals, t)
	}
	if len(literals) == 0 {
		return "", false
	}
	var edits []Edit
	for _, t := range literals[:len(literals)-1] {
		edits = append(edits, Edit{t.End, t.End, " +"})
	}
	value, err := ApplyEdits(source, edits)
	return value, err == nil
}
