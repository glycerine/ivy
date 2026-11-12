package startingpoint

import (
	"strconv"
	"strings"
)

func significantTokens(source string) ([]Token, map[int]int, error) {
	all, err := Lex(source)
	if err != nil {
		return nil, nil, err
	}
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	pairs, err := MatchDelimiters(ts)
	return ts, pairs, err
}

func mechanicalSyntax(source string) (string, error) {
	ts, _, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text == "->" {
			edits = append(edits, Edit{t.Start, t.End, "."})
		}
		if (t.Text == "if" || t.Text == "for") && i+2 < len(ts) && ts[i+2].Text == "{" {
			v, err := strconv.ParseInt(ts[i+1].Text, 0, 64)
			if err == nil {
				replacement := "true"
				if v == 0 {
					replacement = "false"
				}
				if t.Text == "for" && v != 0 {
					if strings.TrimSpace(source[t.End:ts[i+1].Start]) == "" && strings.TrimSpace(source[ts[i+1].End:ts[i+2].Start]) == "" {
						edits = append(edits, Edit{t.End, ts[i+2].Start, " "})
						continue
					}
					replacement = ""
				}
				edits = append(edits, Edit{ts[i+1].Start, ts[i+1].End, replacement})
			}
		}
		// Go's if initializer preserves evaluation order even in else-if chains.
		if t.Text == "if" && i+3 < len(ts) && (ts[i+1].Text == "+" || ts[i+1].Text == "-") && ts[i+2].Text == ts[i+1].Text && ts[i+1].End == ts[i+2].Start && ts[i+3].Kind == "word" && (i+4 >= len(ts) || (ts[i+4].Text != "." && ts[i+4].Text != "->" && ts[i+4].Text != "[" && ts[i+4].Text != "(")) {
			name := ts[i+3].Text
			edits = append(edits, Edit{ts[i+1].Start, ts[i+3].End, name + ts[i+1].Text + ts[i+2].Text + "; " + name})
		}
	}
	return ApplyEdits(source, edits)
}

func lowerDoWhile(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "do" || i+1 >= len(ts) || ts[i+1].Text != "{" {
			continue
		}
		close := pairs[i+1]
		if close+2 >= len(ts) || ts[close+1].Text != "while" || ts[close+2].Text != "(" {
			continue
		}
		end := pairs[close+2]
		if end+1 >= len(ts) || ts[end+1].Text != ";" {
			continue
		}
		hasContinue := false
		for _, inner := range ts[i+2 : close] {
			if inner.Text == "continue" {
				hasContinue = true
			}
		}
		if hasContinue {
			continue
		}
		cond := source[ts[close+2].End:ts[end].Start]
		edits = append(edits, Edit{t.Start, t.End, "for"}, Edit{ts[close].Start, ts[close].Start, "\nif !(" + cond + ") { break }\n"}, Edit{ts[close+1].Start, ts[end+1].End, ""})
	}
	return ApplyEdits(source, edits)
}

func lowerSwitchCases(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "switch" {
			continue
		}
		open := i + 1
		for open < len(ts) && ts[open].Text != "{" {
			if ts[open].Text == ";" {
				break
			}
			if close, ok := pairs[open]; ok && close > open {
				open = close
			}
			open++
		}
		if open >= len(ts) || ts[open].Text != "{" {
			continue
		}
		end := pairs[open]
		var labels, colons []int
		for j := open + 1; j < end; j++ {
			if close, ok := pairs[j]; ok && close > j {
				j = close
				continue
			}
			if ts[j].Text != "case" && ts[j].Text != "default" {
				continue
			}
			k := j + 1
			for k < end && ts[k].Text != ":" {
				if close, ok := pairs[k]; ok && close > k {
					k = close
				}
				k++
			}
			if k == end {
				break
			}
			labels = append(labels, j)
			colons = append(colons, k)
			j = k
		}
		for n := 0; n+1 < len(labels); n++ {
			left, right := colons[n], labels[n+1]
			if left+1 == right && ts[labels[n]].Text == "case" && ts[right].Text == "case" {
				edits = append(edits, Edit{ts[left].Start, ts[left].End, ","}, Edit{ts[right].Start, ts[right].End, ""})
				continue
			}
			// Inspect the final direct statement, not nested return/break statements.
			last := left + 1
			for j := left + 1; j < right; j++ {
				if ts[j].Text == ";" {
					continue
				}
				last = j
				for j+1 < right && ts[j].Text != ";" {
					if close, ok := pairs[j]; ok && close > j {
						j = close
					}
					if j+1 >= right || ts[j].Text == "}" {
						break
					}
					j++
				}
			}
			if last < right {
				switch ts[last].Text {
				case "break", "return", "goto", "continue", "fallthrough":
					continue
				}
			}
			edits = append(edits, Edit{ts[right].Start, ts[right].Start, "fallthrough\n"})
		}
	}
	return ApplyEdits(source, edits)
}

func mechanicalDraft(source string) (string, error) {
	var err error
	for _, pass := range []func(string) (string, error){liftStaticLocals, lowerAggregateInitializers, lowerBitfields, lowerRemainingStructFields, lowerDoWhile, joinControlBraces, unwrapControlHeaders, simplifyConditionOnlyFor, mechanicalSyntax, lowerSingleObjectCalloc, combineAllocationDeclarations, lowerCommaExpressions, lowerPostfixIndexAssignments, lowerSwitchCases, separateBlockStatements} {
		source, err = pass(source)
		if err != nil {
			return "", err
		}
	}
	return source, nil
}

// C permits a statement immediately after a closing block on the same line.
// Go needs a separator there; keep else attached to its preceding if block.
func separateBlockStatements(source string) (string, error) {
	ts, _, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "}" || i+1 >= len(ts) || ts[i+1].Kind != "word" || (ts[i+1].Text == "else" || ts[i+1].Text == "while") {
			continue
		}
		if !strings.ContainsAny(source[t.End:ts[i+1].Start], "\r\n") {
			edits = append(edits, Edit{t.End, t.End, "\n"})
		}
	}
	return ApplyEdits(source, edits)
}
