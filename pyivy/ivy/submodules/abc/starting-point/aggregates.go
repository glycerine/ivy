package startingpoint

import (
	"regexp"
	"strings"
)

var bitfieldDecl = regexp.MustCompile(`\b(unsigned(?:\s+int)?|int)\s+([A-Za-z_]\w*)\s*:\s*([0-9]+)\s*;`)

func lowerBitfields(source string) (string, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for _, m := range bitfieldDecl.FindAllStringSubmatchIndex(Mask(ts), -1) {
		typ := "int32"
		if strings.HasPrefix(source[m[2]:m[3]], "unsigned") {
			typ = "uint32"
		}
		original := source[m[0]:m[1]]
		replacement := source[m[4]:m[5]] + " " + typ + "; /* C bitfield: " + source[m[6]:m[7]] + " bits; review packing and truncation */"
		if commentsRetained(original, replacement) {
			edits = append(edits, Edit{m[0], m[1], replacement})
		} else if !strings.Contains(original, "/*") && !strings.Contains(original, "//") {
			edits = append(edits, Edit{m[0], m[1], replacement})
		}
	}
	return ApplyEdits(source, edits)
}

// lowerAggregateInitializers uses an explicit Go declaration's type. It never
// guesses a struct/slice type from a bare assignment to an unknown variable.
func lowerAggregateInitializers(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "var" || i+3 >= len(ts) || ts[i+1].Kind != "word" {
			continue
		}
		equal := i + 2
		for equal < len(ts) && ts[equal].Text != "=" && ts[equal].Text != ";" && ts[equal].Text != "}" {
			if close, ok := pairs[equal]; ok && close > equal {
				equal = close
			}
			equal++
		}
		if equal+1 >= len(ts) || ts[equal].Text != "=" || ts[equal+1].Text != "{" {
			continue
		}
		typ := strings.TrimSpace(source[ts[i+1].End:ts[equal].Start])
		if typ == "" || strings.Contains(typ, ",") || strings.HasPrefix(typ, "*") {
			continue
		}
		open := equal + 1
		close := pairs[open]
		if close == open+2 && ts[open+1].Text == "0" {
			original := source[ts[equal].Start:ts[close].End]
			// Retain comments within the removed zero initializer.
			edits = append(edits, Edit{ts[equal].Start, ts[close].End, commentsIn(original)})
			continue
		}
		edits = append(edits, Edit{ts[open].Start, ts[open].Start, typ})
		if close > open+1 && ts[close-1].Text != "," && strings.Contains(source[ts[close-1].End:ts[close].Start], "\n") {
			edits = append(edits, Edit{ts[close-1].End, ts[close-1].End, ","})
		}
	}
	return ApplyEdits(source, edits)
}

var goVariableNames = regexp.MustCompile(`\bvar\s+([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)*)`)

func liftStaticLocals(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	used := map[string]bool{}
	for _, t := range ts {
		if t.Kind == "word" {
			used[t.Text] = true
		}
	}
	globals := map[string]bool{}
	depth := 0
	for j, t := range ts {
		if t.Text == "{" {
			depth++
		}
		if t.Text == "}" {
			depth--
		}
		if depth == 0 && t.Text == "var" && j+1 < len(ts) {
			globals[ts[j+1].Text] = true
		}
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "func" || i+2 >= len(ts) || ts[i+2].Text != "(" {
			continue
		}
		open := pairs[i+2] + 1
		for open < len(ts) && ts[open].Text != "{" && ts[open].Text != ";" {
			open++
		}
		if open >= len(ts) || ts[open].Text != "{" {
			continue
		}
		end := pairs[open]
		type declaration struct {
			start, end int
			text       string
		}
		var declarations []declaration
		names := map[string]string{}
		for j := open + 1; j < end; j++ {
			if ts[j].Text != "static" || j+1 >= end {
				continue
			}
			stop := j + 1
			for stop < end && ts[stop].Text != ";" {
				if close, ok := pairs[stop]; ok && close > stop {
					stop = close
				}
				stop++
			}
			if stop >= end {
				continue
			}
			types := map[string]bool{ts[j+1].Text: true}
			raw := "{" + source[ts[j+1].Start:ts[stop].End] + "}"
			converted, err := lowerDeclarations(raw, types, true)
			if err != nil || converted == raw {
				continue
			}
			for _, m := range goVariableNames.FindAllStringSubmatch(converted, -1) {
				for _, name := range strings.Split(m[1], ",") {
					name = strings.TrimSpace(name)
					fresh := "spStatic_" + ts[i+1].Text + "_" + name
					for used[fresh] {
						fresh += "_"
					}
					used[fresh] = true
					names[name] = fresh
				}
			}
			declarations = append(declarations, declaration{ts[j].Start, ts[stop].End, converted[1 : len(converted)-1]})
			j = stop
		}
		if len(declarations) == 0 {
			continue
		}
		// C++ local statics can initialize lazily from calls or automatic variables.
		// Hoist only literal/zero/address initializers; retain dynamic initialization.
		constantInitializers := true
		for _, d := range declarations {
			tokens, _, err := significantTokens(d.text)
			if err != nil {
				return "", err
			}
			inValue := false
			numericEnd := -1
			for at, tok := range tokens {
				if tok.Text == ";" {
					inValue = false
				}
				if tok.Text == "=" {
					inValue = true
					continue
				}
				if !inValue {
					continue
				}
				if tok.Start < numericEnd {
					continue
				}
				if literal := castNumber.FindString(d.text[tok.Start:]); literal != "" {
					numericEnd = tok.Start + len(literal)
					continue
				}
				if tok.Kind != "word" {
					continue
				}
				if tok.Text == "NULL" || tok.Text == "nil" || tok.Text == "true" || tok.Text == "false" || scalarCastType(tok.Text) != "" {
					continue
				}
				if at > 0 && tokens[at-1].Text == "&" && (names[tok.Text] != "" || globals[tok.Text]) {
					continue
				}
				constantInitializers = false
			}
		}
		if !constantInitializers {
			continue
		}

		inside := func(at int) bool {
			for _, d := range declarations {
				if at >= d.start && at < d.end {
					return true
				}
			}
			return false
		}
		// Leave shadowed local-static names for review rather than rename ambiguously.
		shadow := false
		for _, parameter := range ts[i+3 : pairs[i+2]] {
			if names[parameter.Text] != "" {
				shadow = true
			}
		}

		for j := open + 1; j < end; j++ {
			if ts[j].Text == "var" && j+1 < end && names[ts[j+1].Text] != "" && !inside(ts[j].Start) {
				shadow = true
			}
		}
		if shadow {
			continue
		}
		var globals strings.Builder
		for _, d := range declarations {
			all, err := Lex(d.text)
			if err != nil {
				return "", err
			}
			var rename []Edit
			for _, t := range all {
				if replacement := names[t.Text]; t.Kind == "word" && replacement != "" {
					rename = append(rename, Edit{t.Start, t.End, replacement})
				}
			}
			text, err := ApplyEdits(d.text, rename)
			if err != nil {
				return "", err
			}
			globals.WriteString(text)
			globals.WriteByte('\n')
			edits = append(edits, Edit{d.start, d.end, ""})
		}
		edits = append(edits, Edit{t.Start, t.Start, globals.String()})
		for j := open + 1; j < end; j++ {
			if inside(ts[j].Start) || (j > 0 && (ts[j-1].Text == "." || ts[j-1].Text == "->")) {
				continue
			}
			if replacement := names[ts[j].Text]; ts[j].Kind == "word" && replacement != "" {
				edits = append(edits, Edit{ts[j].Start, ts[j].End, replacement})
			}
		}
	}
	return ApplyEdits(source, edits)
}

func lowerRemainingStructFields(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	types := declarationTypes("", nil)
	for i, t := range ts {
		if t.Text == "type" && i+1 < len(ts) {
			types[ts[i+1].Text] = true
		}
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "struct" || i+1 >= len(ts) || ts[i+1].Text != "{" {
			continue
		}
		end := pairs[i+1]
		start := i + 2
		for j := start; j < end; j++ {
			if close, ok := pairs[j]; ok && close > j {
				j = close
				continue
			}
			if ts[j].Text != ";" {
				continue
			}
			if start < j && types[ts[start].Text] {
				fragment := source[ts[start].Start:ts[j].End]
				if strings.ContainsAny(fragment, "[,") {
					wrapped := "{" + fragment + "}"
					converted, err := lowerDeclarations(wrapped, types, true)
					if err != nil {
						return "", err
					}
					if converted != wrapped {
						converted = strings.ReplaceAll(converted[1:len(converted)-1], "var ", "")
						edits = append(edits, Edit{ts[start].Start, ts[j].End, converted})
					}
				}
			}
			start = j + 1
		}
	}
	return ApplyEdits(source, edits)
}
