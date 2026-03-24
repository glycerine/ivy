package ivyutils

import (
	"regexp"
	"strings"

	cristalbase64 "github.com/cristalhq/base64"
	"github.com/glycerine/blake3"
)

// Canonical strings are 1:1 with the state
// they represent, so immune to internal collection
// reordering differences: the are s-expressions
// strings in compact (not pretty printed) form.
type Canonical string

// Canonizer supports Canon calls.
type Canonizer interface {
	// Canon returns a canonical S-expression string
	// for the given implementor.
	Canon() Canonical
}

// MerkleState tracks a rolling Merkle root for incremental state verification.
// Used by the parser accumulator to hash each declared AST node and maintain
// a running root that can be compared across Go and Python.
type MerkleState struct {
	PrevRoot string // blake3 hash string
}

// AddLeaf hashes a canonical string and combines it with the running root.
// Returns the leaf hash and the new root hash.
// Line numbers and filenames are stripped before hashing so that
// Go and Python can have different (but both correct) line numbers
// without causing hash mismatches. The full canon string with line
// numbers is still available for verbose display.
func (ms *MerkleState) AddLeaf(c Canonical) (leafB3, rootB3 string) {
	stripped := StripLocations(c)
	leafB3 = stripped.Blake3()
	combined := Canonical(ms.PrevRoot + leafB3)
	ms.PrevRoot = combined.Blake3()
	rootB3 = ms.PrevRoot
	return
}

// StripLocations removes lineno:N and filename:"..." fields from a
// canonical s-expression so that hashes are location-independent.
func StripLocations(c Canonical) Canonical {
	s := string(c)
	s = reLineno.ReplaceAllString(s, "")
	s = reFilename.ReplaceAllString(s, "")
	// Clean up double spaces left by removal
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	// Clean up "( " left at start of type names
	s = strings.ReplaceAll(s, "( ", "(")
	return Canonical(s)
}

var reLineno = regexp.MustCompile(`\s*lineno:\d+`)
var reFilename = regexp.MustCompile(`\s*filename:"[^"]*"`)

// PrettySexp formats a canonical s-expression with indentation for readability.
// Opening parens and brackets increase indent; closing ones decrease it.
// Each field (key:value) gets its own line at the current indent level.
func PrettySexp(s string) string {
	var sb strings.Builder
	indent := 0
	i := 0
	n := len(s)

	writeIndent := func() {
		for j := 0; j < indent; j++ {
			sb.WriteString("  ")
		}
	}

	for i < n {
		ch := s[i]
		switch ch {
		case '(':
			// Start of a struct: find the type name (until first space or closing paren)
			sb.WriteByte('(')
			i++
			// Read type name
			start := i
			for i < n && s[i] != ' ' && s[i] != ')' {
				i++
			}
			sb.WriteString(s[start:i])
			indent++
			// If next char is space, emit newline + indent for first field
			if i < n && s[i] == ' ' {
				i++ // skip space
				sb.WriteByte('\n')
				writeIndent()
			}
		case '[':
			sb.WriteByte('[')
			i++
			indent++
			// Check if it's empty []
			if i < n && s[i] == ']' {
				sb.WriteByte(']')
				i++
				indent--
			} else if i < n && s[i] == '(' {
				// Array of structs: newline before first element
				sb.WriteByte('\n')
				writeIndent()
			}
		case ')':
			indent--
			sb.WriteByte(')')
			i++
		case ']':
			indent--
			sb.WriteByte(']')
			i++
		case ' ':
			// Space between fields: emit newline + indent
			i++
			sb.WriteByte('\n')
			writeIndent()
		case '"':
			// Quoted string: copy verbatim until closing quote
			sb.WriteByte('"')
			i++
			for i < n {
				if s[i] == '\\' && i+1 < n {
					sb.WriteByte(s[i])
					sb.WriteByte(s[i+1])
					i += 2
				} else if s[i] == '"' {
					sb.WriteByte('"')
					i++
					break
				} else {
					sb.WriteByte(s[i])
					i++
				}
			}
		default:
			sb.WriteByte(ch)
			i++
		}
	}
	return sb.String()
}

// DiffSexp pretty-prints two canonical s-expressions and returns a
// diff string highlighting the differences. Returns "" if they are
// structurally identical (ignoring location fields like lineno/filename).
//
// The output shows full lines WITH line numbers for debugging context.
// Lines that differ only in locations are shown with "~go:"/"~py:" prefix.
// Lines with real structural differences use "go:"/"py:".
// Matching lines use "   " (3 spaces).
func DiffSexp(a, b string) string {
	// Pretty-print with locations stripped — for structural comparison
	strippedA := PrettySexp(string(StripLocations(Canonical(a))))
	strippedB := PrettySexp(string(StripLocations(Canonical(b))))
	if strippedA == strippedB {
		return ""
	}
	// Pretty-print with locations intact — for display
	fullA := PrettySexp(a)
	fullB := PrettySexp(b)
	linesA := strings.Split(fullA, "\n")
	linesB := strings.Split(fullB, "\n")
	sLinesA := strings.Split(strippedA, "\n")
	sLinesB := strings.Split(strippedB, "\n")

	var sb strings.Builder
	maxLen := len(sLinesA)
	if len(sLinesB) > maxLen {
		maxLen = len(sLinesB)
	}
	for i := 0; i < maxLen; i++ {
		dispA, dispB := "", ""
		stripA, stripB := "", ""
		if i < len(linesA) {
			dispA = linesA[i]
		}
		if i < len(linesB) {
			dispB = linesB[i]
		}
		if i < len(sLinesA) {
			stripA = sLinesA[i]
		}
		if i < len(sLinesB) {
			stripB = sLinesB[i]
		}
		if stripA == stripB {
			if dispA == dispB {
				sb.WriteString("      " + dispA + "\n")
			} else {
				// Location-only difference
				sb.WriteString("~go:  " + dispA + "\n")
				sb.WriteString("~py:  " + dispB + "\n")
			}
		} else {
			// Real structural difference
			sb.WriteString(" go:  " + dispA + "\n")
			sb.WriteString(" py:  " + dispB + "\n")
		}
	}
	return sb.String()
}

// Blake3 returns the first 33 bytes of
// 64 byte (512 bit) blake3 hash of a Canonical string.
// The hash is then base-64 URL encoded,
// and prefixed with "blake3.33B-".
// It is goroutine safe and lock free, since
// it creates a new hasher every time.
func (c Canonical) Blake3() string {
	h := blake3.New(64, nil)
	h.Write([]byte(c))
	sum := h.Sum(nil)
	return "blake3.33B-" + cristalbase64.URLEncoding.EncodeToString(sum[:33])
}
