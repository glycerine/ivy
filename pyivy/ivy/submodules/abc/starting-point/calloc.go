package startingpoint

import (
	"strconv"
	"strings"
)

// lowerSingleObjectCalloc recognizes a cast that supplies the allocation type,
// a literal count of one, and a matching sizeof type. Run after macro expansion
// so ABC_CALLOC and the explicit C expression share this rule.
func lowerSingleObjectCalloc(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	covered := 0
	for i, t := range ts {
		if i < covered || t.Text != "(" || i+4 >= len(ts) {
			continue
		}
		// (T *) calloc(...), with precisely one pointer level.
		if ts[i+1].Kind != "word" || ts[i+2].Text != "*" || ts[i+3].Text != ")" || pairs[i] != i+3 || ts[i+4].Text != "calloc" {
			continue
		}
		typ := ts[i+1].Text
		call := i + 5
		if call >= len(ts) || ts[call].Text != "(" {
			continue
		}
		end, ok := pairs[call]
		if !ok {
			continue
		}
		comma := -1
		for j := call + 1; j < end; j++ {
			if ts[j].Text == "," {
				if comma >= 0 {
					comma = -1
					break
				}
				comma = j
			}
			if close, ok := pairs[j]; ok && close > j {
				j = close
			}
		}
		if comma < 0 || !callocOne(ts, pairs, call+1, comma) {
			continue
		}
		lo, hi := callocUngroup(ts, pairs, comma+1, end)
		if hi-lo != 4 || ts[lo].Text != "sizeof" || ts[lo+1].Text != "(" || ts[lo+2].Text != typ || ts[lo+3].Text != ")" {
			continue
		}
		start := i
		// Consume only parentheses enclosing this complete cast/call expression.
		for start > 0 && end+1 < len(ts) && ts[start-1].Text == "(" && pairs[start-1] == end+1 {
			start--
			end++
		}
		goTyp := goType(typ)
		value := "&" + goTyp + "{}"
		switch goTyp {
		case "int", "int16", "int32", "int64", "uint", "uint16", "uint32", "uint64", "byte", "bool", "float32", "float64":
			value = "new(" + goTyp + ")"
		}
		if end+1 < len(ts) && (ts[end+1].Text == "." || ts[end+1].Text == "->" || ts[end+1].Text == "[" || ts[end+1].Text == "(") {
			value = "(" + value + ")"
		}
		original := source[ts[start].Start:ts[end].End]
		value = commentsIn(original) + value
		edits = append(edits, Edit{ts[start].Start, ts[end].End, value})
		covered = end + 1
	}
	return ApplyEdits(source, edits)
}
func callocUngroup(ts []Token, pairs map[int]int, lo, hi int) (int, int) {
	for lo < hi && ts[lo].Text == "(" && pairs[lo] == hi-1 {
		lo++
		hi--
	}
	return lo, hi
}
func callocOne(ts []Token, pairs map[int]int, lo, hi int) bool {
	lo, hi = callocUngroup(ts, pairs, lo, hi)
	if hi-lo >= 3 && ts[lo].Text == "(" && ts[lo+1].Text == "size_t" && ts[lo+2].Text == ")" {
		lo, hi = callocUngroup(ts, pairs, lo+3, hi)
	}
	var text strings.Builder
	for i := lo; i < hi; i++ {
		text.WriteString(ts[i].Text)
	}
	value := strings.TrimRight(text.String(), "uUlL")
	n, err := strconv.ParseInt(value, 0, 64)
	return err == nil && n == 1
}
