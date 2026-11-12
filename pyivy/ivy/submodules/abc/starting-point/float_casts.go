package startingpoint

import "regexp"

var castNumber = regexp.MustCompile(`^(?:0[xX](?:[0-9a-fA-F]+(?:\.[0-9a-fA-F]*)?|\.[0-9a-fA-F]+)(?:[pP][+-]?[0-9]+)?|(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?)[uUlLfF]*`)

// lowerFloatCasts converts only scalar float/double casts. Token boundaries keep
// strings, comments, declarations and pointer casts out of this rewrite.
func lowerFloatCasts(source string) (string, error) {
	all, err := Lex(source)
	if err != nil {
		return "", err
	}
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return "", err
	}
	cast := func(i int) bool {
		return i+2 < len(ts) && ts[i].Text == "(" && (ts[i+1].Text == "double" || ts[i+1].Text == "float") && ts[i+2].Text == ")"
	}
	var unaryEnd func(int) int
	unaryEnd = func(i int) int {
		if i >= len(ts) {
			return -1
		}
		if cast(i) {
			return unaryEnd(i + 3)
		}
		switch ts[i].Text {
		case "+", "-", "!", "~", "*", "&", "++", "--":
			return unaryEnd(i + 1)
		}
		end := i + 1
		if ts[i].Text == "(" {
			end = pairs[i] + 1
		} else if number := castNumber.FindString(source[ts[i].Start:]); number != "" {
			limit := ts[i].Start + len(number)
			for end < len(ts) && ts[end].Start < limit {
				end++
			}
		} else if ts[i].Kind != "word" && ts[i].Kind != "number" && ts[i].Kind != "string" && ts[i].Kind != "char" {
			return -1
		}
		for end < len(ts) {
			switch ts[end].Text {
			case "(", "[":
				end = pairs[end] + 1
			case ".", "->":
				if end+1 >= len(ts) || ts[end+1].Kind != "word" {
					return -1
				}
				end += 2
			case "+", "-":
				if end+1 < len(ts) && ts[end+1].Text == ts[end].Text && ts[end].End == ts[end+1].Start {
					end += 2
				} else {
					return end
				}
			case "++", "--":
				end++
			default:
				return end
			}
		}
		return end
	}
	var edits []Edit
	for i := 0; i+3 < len(ts); i++ {
		if !cast(i) {
			continue
		}
		// sizeof(float) and sizeof(double) name types, not cast expressions.
		if i > 0 && (ts[i-1].Text == "sizeof" || ts[i-1].Text == "_Alignof") {
			continue
		}
		end := unaryEnd(i + 3)
		if end <= i+3 {
			continue
		}
		typ := "float64"
		if ts[i+1].Text == "float" {
			typ = "float32"
		}
		trivia := source[ts[i].End:ts[i+1].Start] + source[ts[i+1].End:ts[i+2].Start]
		edits = append(edits, Edit{ts[i].Start, ts[i+2].End, typ + "(" + trivia})
		edits = append(edits, Edit{ts[end-1].End, ts[end-1].End, ")"})
	}
	return ApplyEdits(source, edits)
}
