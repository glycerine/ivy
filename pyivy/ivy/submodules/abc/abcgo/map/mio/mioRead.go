// Genlib reader from mioRead.c. Parsing is local and errors include line numbers;
// no strtok globals, destructive buffers or process exits are used.
package mio

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type genlibToken struct {
	text string
	line int
}

func genlibTokens(text string) []genlibToken {
	text = strings.ReplaceAll(text, "\\\r\n", "  \n")
	text = strings.ReplaceAll(text, "\\\n", " \n")
	var tokens []genlibToken
	line := 1
	for i := 0; i < len(text); {
		switch text[i] {
		case '#':
			for i < len(text) && text[i] != '\n' {
				i++
			}
			continue
		case '\n':
			line++
			i++
			continue
		case ' ', '\t', '\r':
			i++
			continue
		}
		start := i
		if text[i] == '=' || text[i] == ';' {
			i++
		} else {
			for i < len(text) && !strings.ContainsRune(" \t\r\n#=;", rune(text[i])) {
				i++
			}
		}
		tokens = append(tokens, genlibToken{text[start:i], line})
	}
	return tokens
}

type genlibParser struct {
	tokens []genlibToken
	next   int
}

func (p *genlibParser) peek() string {
	if p.next == len(p.tokens) {
		return ""
	}
	return p.tokens[p.next].text
}
func (p *genlibParser) take() (string, error) {
	if p.next == len(p.tokens) {
		return "", fmt.Errorf("unexpected end of genlib")
	}
	s := p.tokens[p.next].text
	p.next++
	return s, nil
}
func (p *genlibParser) number() (float64, error) {
	s, err := p.take()
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, fmt.Errorf("invalid finite number %q", s)
	}
	return v, nil
}
func (p *genlibParser) expect(s string) error {
	actual, err := p.take()
	if err != nil {
		return err
	}
	if actual != s {
		return fmt.Errorf("expected %q, found %q", s, actual)
	}
	return nil
}
func (p *genlibParser) pin(extended bool) (*Mio_Pin_t, error) {
	name, err := p.take()
	if err != nil {
		return nil, err
	}
	phase, err := p.take()
	if err != nil {
		return nil, err
	}
	pin := &Mio_Pin_t{pName: name}
	switch phase {
	case MIO_STRING_UNKNOWN:
		pin.Phase = MIO_PHASE_UNKNOWN
	case MIO_STRING_INV:
		pin.Phase = MIO_PHASE_INV
	case MIO_STRING_NONINV:
		pin.Phase = MIO_PHASE_NONINV
	default:
		return nil, fmt.Errorf("invalid pin phase %q", phase)
	}
	count := 6
	if extended {
		count = 8
	}
	values := make([]float64, count)
	for i := range values {
		values[i], err = p.number()
		if err != nil {
			return nil, err
		}
	}
	pin.dLoadInput, pin.dLoadMax = values[0], values[1]
	pin.dDelayBlockRise, pin.dDelayFanoutRise = values[2], values[3]
	if extended {
		pin.dDelayBlockFall, pin.dDelayFanoutFall = values[5], values[6]
	} else {
		pin.dDelayBlockFall, pin.dDelayFanoutFall = values[4], values[5]
	}
	pin.dDelayBlockMax = max(pin.dDelayBlockRise, pin.dDelayBlockFall)
	return pin, nil
}
func (p *genlibParser) gate(extended bool) (*Mio_Gate_t, error) {
	name, err := p.take()
	if err != nil {
		return nil, err
	}
	area, err := p.number()
	if err != nil || area < 0 {
		return nil, fmt.Errorf("gate %s has invalid area", name)
	}
	output, err := p.take()
	if err != nil {
		return nil, err
	}
	if err = p.expect("="); err != nil {
		return nil, err
	}
	var formula []string
	for p.peek() != ";" {
		token, err := p.take()
		if err != nil {
			return nil, err
		}
		if token == "GATE" || token == "PIN" {
			return nil, fmt.Errorf("missing formula semicolon")
		}
		formula = append(formula, token)
	}
	p.next++
	g := &Mio_Gate_t{pName: name, dArea: area, pOutName: output, pForm: strings.Join(formula, " "), Cell: -1}
	var tail *Mio_Pin_t
	for p.peek() == MIO_STRING_PIN {
		p.next++
		pin, err := p.pin(extended)
		if err != nil {
			return nil, err
		}
		if tail == nil {
			g.pPins = pin
		} else {
			tail.pNext = pin
		}
		tail = pin
	}
	return g, nil
}
func parseGenlib(name, text string, exclude map[string]bool, extended bool) (*Mio_Library_t, error) {
	p := genlibParser{tokens: genlibTokens(text)}
	lib := &Mio_Library_t{pName: name, tName2Gate: make(map[string]*Mio_Gate_t)}
	var tail *Mio_Gate_t
	for p.peek() != "" && p.peek() != ".end" {
		start := p.next
		token, _ := p.take()
		if token == MIO_STRING_LATCH {
			for p.peek() != "" && p.peek() != ".end" && p.peek() != MIO_STRING_GATE {
				p.next++
			}
			continue
		}
		if token != MIO_STRING_GATE {
			return nil, fmt.Errorf("%s:%d: expected GATE, found %q", name, p.tokens[start].line, token)
		}
		g, err := p.gate(extended)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", name, p.tokens[start].line, err)
		}
		if exclude[g.pName] {
			continue
		}
		g.pLib = lib
		if err = deriveGate(g); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", name, p.tokens[start].line, err)
		}
		if base := lib.tName2Gate[g.pName]; base != nil {
			if base.pTwin != nil || base.pOutName == g.pOutName {
				return nil, fmt.Errorf("%s:%d: duplicate gate output %s/%s", name, p.tokens[start].line, g.pName, g.pOutName)
			}
			base.pTwin = g
			g.pTwin = base
		} else {
			lib.tName2Gate[g.pName] = g
		}
		if tail == nil {
			lib.pGates = g
		} else {
			tail.pNext = g
		}
		tail = g
		lib.nGates++
	}
	if p.peek() == ".end" {
		p.next++
		if p.peek() != "" {
			return nil, fmt.Errorf("%s: unexpected data after .end", name)
		}
	}
	if lib.nGates == 0 {
		return nil, fmt.Errorf("%s: library contains no gates", name)
	}
	Mio_LibraryDetectSpecialGates(lib)
	return lib, nil
}

// ReadLibrary accepts ordinary or extended genlib timing records.
func ReadLibrary(name, text string, exclude map[string]bool) (*Mio_Library_t, error) {
	lib, err := parseGenlib(name, text, exclude, false)
	if err == nil {
		return lib, nil
	}
	if extended, extendedErr := parseGenlib(name, text, exclude, true); extendedErr == nil {
		return extended, nil
	}
	return nil, err
}
func ReadLibraryFile(filename string, exclude map[string]bool) (*Mio_Library_t, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ReadLibrary(filename, string(data), exclude)
}
func Mio_ReadFile(filename string, addEnd int32) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	if addEnd != 0 {
		return string(data) + "\n.end\n"
	}
	return string(data)
}
func Mio_LibraryRead(filename, buffer, excludeFile string, verbose int32) *Mio_Library_t {
	excludes := make(map[string]bool)
	if excludeFile != "" {
		if Mio_LibraryReadExclude(excludeFile, excludes) < 0 {
			return nil
		}
	}
	var lib *Mio_Library_t
	var err error
	if buffer == "" {
		lib, err = ReadLibraryFile(filename, excludes)
	} else {
		lib, err = ReadLibrary(filename, buffer, excludes)
	}
	if err != nil {
		return nil
	}
	return lib
}
func Mio_LibraryReadBuffer(buffer string, extended int32, exclude map[string]bool, verbose int32) *Mio_Library_t {
	lib, err := parseGenlib("<buffer>", buffer, exclude, extended != 0)
	if err != nil {
		return nil
	}
	return lib
}
func Mio_LibraryReadExclude(filename string, exclude map[string]bool) int32 {
	data, err := os.ReadFile(filename)
	if err != nil {
		return -1
	}
	var count int32
	for _, line := range strings.Split(string(data), "\n") {
		line, _, _ = strings.Cut(line, "#")
		for _, name := range strings.Fields(line) {
			if !exclude[name] {
				count++
				exclude[name] = true
			}
		}
	}
	return count
}
func Mio_LibraryCompareGatesByName(a, b *Mio_Gate_t) int32 {
	return int32(strings.Compare(a.pName, b.pName))
}
func Mio_LibrarySortGates(lib *Mio_Library_t) {
	var gates []*Mio_Gate_t
	for g := lib.pGates; g != nil; g = g.pNext {
		g.Cell = int32(len(gates))
		gates = append(gates, g)
	}
	lib.ppGates0 = append([]*Mio_Gate_t(nil), gates...)
	sort.SliceStable(gates, func(i, j int) bool { return gates[i].pName < gates[j].pName })
	lib.ppGatesName = gates
	for i, g := range gates {
		g.pNext = nil
		if i+1 < len(gates) {
			g.pNext = gates[i+1]
		}
	}
	if len(gates) > 0 {
		lib.pGates = gates[0]
	}
}
func Mio_GateCompare(current, next *Mio_Gate_t, truth uint64) *Mio_Gate_t {
	if next.nInputs > 6 || next.uTruth != truth {
		return current
	}
	if current == nil || next.dArea < current.dArea || next.dArea == current.dArea && next.pName < current.pName {
		return next
	}
	return current
}
func Mio_LibraryDetectSpecialGates(lib *Mio_Library_t) {
	Mio_LibrarySortGates(lib)
	targets := []struct {
		truth uint64
		slot  **Mio_Gate_t
	}{{0, &lib.pGate0}, {^uint64(0), &lib.pGate1}, {expTruth6[0], &lib.pGateBuf}, {^expTruth6[0], &lib.pGateInv}, {expTruth6[0] & expTruth6[1], &lib.pGateAnd2}, {^(expTruth6[0] & expTruth6[1]), &lib.pGateNand2}, {expTruth6[0] | expTruth6[1], &lib.pGateOr2}, {^(expTruth6[0] | expTruth6[1]), &lib.pGateNor2}}
	for _, target := range targets {
		*target.slot = nil
		for g := lib.pGates; g != nil; g = g.pNext {
			*target.slot = Mio_GateCompare(*target.slot, g, target.truth)
		}
	}
}
