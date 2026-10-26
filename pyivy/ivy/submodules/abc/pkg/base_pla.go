package abc

// Stand-alone port of src/base/pla.

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type PlaFileType uint8

const (
	PlaFileFD PlaFileType = iota
	PlaFileF
	PlaFileFR
	PlaFileFDR
	PlaFileNone
)

type PlaLiteral uint8

const (
	PlaLitDash PlaLiteral = iota
	PlaLitZero
	PlaLitOne
	PlaLitFull
)

type PlaLimits struct {
	MaxInputs, MaxOutputs, MaxCubes, MaxBytes, MaxOperations int
	Stop                                                     func() bool
}

func normalizePlaLimits(l PlaLimits) (PlaLimits, error) {
	if l.MaxInputs < 0 || l.MaxOutputs < 0 || l.MaxCubes < 0 || l.MaxBytes < 0 || l.MaxOperations < 0 {
		return l, ErrDimension
	}
	if l.MaxInputs == 0 {
		l.MaxInputs = 1 << 20
	}
	if l.MaxOutputs == 0 {
		l.MaxOutputs = 1 << 20
	}
	if l.MaxCubes == 0 {
		l.MaxCubes = 1 << 24
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = math.MaxInt32
	}
	if l.MaxOperations == 0 {
		l.MaxOperations = math.MaxInt32
	}
	return l, nil
}

type PlaCube struct{ Inputs, Outputs []PlaLiteral }

func clonePlaCube(c PlaCube) PlaCube {
	return PlaCube{Inputs: append([]PlaLiteral(nil), c.Inputs...), Outputs: append([]PlaLiteral(nil), c.Outputs...)}
}

type PlaManager struct {
	name, spec      string
	typ             PlaFileType
	inputs, outputs int
	cubes           []PlaCube
	limits          PlaLimits
	closed          bool
}

func NewPlaManager(name, spec string, inputs, outputs int, typ PlaFileType, limits PlaLimits) (*PlaManager, error) {
	limits, err := normalizePlaLimits(limits)
	if err != nil {
		return nil, err
	}
	if inputs <= 0 || outputs <= 0 || inputs > limits.MaxInputs || outputs > limits.MaxOutputs || typ > PlaFileNone || strings.ContainsAny(name, "\r\n") {
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &PlaManager{name: name, spec: spec, typ: typ, inputs: inputs, outputs: outputs, limits: limits}, nil
}

func (p *PlaManager) check() error {
	if p == nil || p.closed || p.inputs <= 0 || p.outputs <= 0 || p.typ > PlaFileNone {
		return ErrInvariant
	}
	if len(p.cubes) > p.limits.MaxCubes {
		return ErrLimit
	}
	return nil
}

func (p *PlaManager) Name() string {
	if p == nil || p.closed {
		return ""
	}
	return p.name
}
func (p *PlaManager) Spec() string {
	if p == nil || p.closed {
		return ""
	}
	return p.spec
}
func (p *PlaManager) InputCount() int {
	if p == nil || p.closed {
		return 0
	}
	return p.inputs
}
func (p *PlaManager) OutputCount() int {
	if p == nil || p.closed {
		return 0
	}
	return p.outputs
}
func (p *PlaManager) CubeCount() int {
	if p == nil || p.closed {
		return 0
	}
	return len(p.cubes)
}
func (p *PlaManager) FileType() PlaFileType {
	if p == nil || p.closed {
		return PlaFileNone
	}
	return p.typ
}

func validPlaLiteral(l PlaLiteral) bool { return l <= PlaLitFull }

func (p *PlaManager) AddCube(inputs, outputs []PlaLiteral) error {
	if err := p.check(); err != nil {
		return err
	}
	if p.limits.Stop != nil && p.limits.Stop() {
		return ErrCancelled
	}
	if len(inputs) != p.inputs || len(outputs) != p.outputs {
		return ErrDimension
	}
	if len(p.cubes) == p.limits.MaxCubes {
		return ErrLimit
	}
	for _, l := range inputs {
		if !validPlaLiteral(l) {
			return ErrDimension
		}
	}
	for _, l := range outputs {
		if !validPlaLiteral(l) {
			return ErrDimension
		}
	}
	p.cubes = append(p.cubes, clonePlaCube(PlaCube{Inputs: inputs, Outputs: outputs}))
	return nil
}

func (p *PlaManager) Cubes() ([]PlaCube, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	r := make([]PlaCube, len(p.cubes))
	for i := range p.cubes {
		r[i] = clonePlaCube(p.cubes[i])
	}
	return r, nil
}

func (p *PlaManager) Clone() (*PlaManager, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	r := *p
	r.cubes = make([]PlaCube, len(p.cubes))
	for i := range p.cubes {
		r.cubes[i] = clonePlaCube(p.cubes[i])
	}
	return &r, nil
}

func (p *PlaManager) Close() {
	if p != nil && !p.closed {
		p.cubes = nil
		p.name, p.spec = "", ""
		p.closed = true
	}
}

func (p *PlaManager) InputLiteralCount() (int, error) {
	if err := p.check(); err != nil {
		return 0, err
	}
	n := 0
	for _, c := range p.cubes {
		for _, l := range c.Inputs {
			if l != PlaLitDash {
				n++
			}
		}
	}
	return n, nil
}
func (p *PlaManager) OutputLiteralCount() (int, error) {
	if err := p.check(); err != nil {
		return 0, err
	}
	n := 0
	for _, c := range p.cubes {
		for _, l := range c.Outputs {
			if l == PlaLitOne {
				n++
			}
		}
	}
	return n, nil
}

func plaCubeMatches(c PlaCube, assignment uint64) bool {
	for i, l := range c.Inputs {
		bit := (assignment >> i) & 1
		if l == PlaLitFull || l == PlaLitZero && bit != 0 || l == PlaLitOne && bit == 0 {
			return false
		}
	}
	return true
}

func (p *PlaManager) Evaluate(assignment uint64) ([]bool, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	if p.inputs < 64 && assignment >= uint64(1)<<p.inputs {
		return nil, ErrDimension
	}
	r := make([]bool, p.outputs)
	for _, c := range p.cubes {
		if plaCubeMatches(c, assignment) {
			for i, l := range c.Outputs {
				r[i] = r[i] || l == PlaLitOne || l == PlaLitFull
			}
		}
	}
	return r, nil
}

func PlaCubesDistanceOne(a, b []PlaLiteral) bool {
	if len(a) != len(b) {
		return false
	}
	differences := 0
	for i := range a {
		if a[i] != b[i] {
			differences++
			if differences > 1 {
				return false
			}
		}
	}
	return differences == 1
}

func PlaCubesConsensus(a, b []PlaLiteral) (int, bool) {
	if len(a) != len(b) {
		return 0, false
	}
	opposite := -1
	for i := range a {
		difference := a[i] ^ b[i]
		if difference&1 != 0 && difference&2 != 0 {
			if opposite >= 0 {
				return 0, false
			}
			opposite = i
		}
	}
	return opposite, opposite >= 0
}

type PlaDistancePair struct{ First, Second, Variable int }

func (p *PlaManager) DistanceOnePairs() ([]PlaDistancePair, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	pairs := make([]PlaDistancePair, 0)
	operations := 0
	for i := range p.cubes {
		for j := i + 1; j < len(p.cubes); j++ {
			operations++
			if operations > p.limits.MaxOperations {
				return nil, ErrLimit
			}
			if p.limits.Stop != nil && p.limits.Stop() {
				return nil, ErrCancelled
			}
			if !PlaCubesDistanceOne(p.cubes[i].Inputs, p.cubes[j].Inputs) {
				continue
			}
			v := 0
			for p.cubes[i].Inputs[v] == p.cubes[j].Inputs[v] {
				v++
			}
			pairs = append(pairs, PlaDistancePair{i, j, v})
		}
	}
	return pairs, nil
}

func plaHashCube(c []PlaLiteral) uint64 {
	h := uint64(1469598103934665603)
	for i, l := range c {
		h ^= uint64(i<<2) | uint64(l)
		h *= 1099511628211
	}
	return h
}

func (p *PlaManager) Hashes() ([]uint64, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	r := make([]uint64, len(p.cubes))
	for i := range p.cubes {
		r[i] = plaHashCube(p.cubes[i].Inputs)
	}
	return r, nil
}

func (p *PlaManager) HashDistanceOneCount() (int, error) {
	pairs, err := p.DistanceOnePairs()
	return len(pairs), err
}

func equalPlaLiterals(a, b []PlaLiteral) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func plaCubeKey(c PlaCube) string {
	b := make([]byte, 0, len(c.Inputs)+len(c.Outputs)+1)
	for _, l := range c.Inputs {
		b = append(b, byte(l))
	}
	b = append(b, 255)
	for _, l := range c.Outputs {
		b = append(b, byte(l))
	}
	return string(b)
}

func (p *PlaManager) CombineDistanceOne() (int, error) {
	if err := p.check(); err != nil {
		return 0, err
	}
	candidate, _ := p.Clone()
	total, operations := 0, 0
	for {
		used := make([]bool, len(candidate.cubes))
		next := make([]PlaCube, 0, len(candidate.cubes))
		keys := make([]string, 0)
		for i := range candidate.cubes {
			for j := i + 1; j < len(candidate.cubes); j++ {
				operations++
				if operations > p.limits.MaxOperations {
					return 0, ErrLimit
				}
				if p.limits.Stop != nil && p.limits.Stop() {
					return 0, ErrCancelled
				}
				if !equalPlaLiterals(candidate.cubes[i].Outputs, candidate.cubes[j].Outputs) || !PlaCubesDistanceOne(candidate.cubes[i].Inputs, candidate.cubes[j].Inputs) {
					continue
				}
				c := clonePlaCube(candidate.cubes[i])
				for v := range c.Inputs {
					if c.Inputs[v] != candidate.cubes[j].Inputs[v] {
						c.Inputs[v] = PlaLitDash
					}
				}
				k := plaCubeKey(c)
				pos := sort.SearchStrings(keys, k)
				if pos == len(keys) || keys[pos] != k {
					keys = append(keys, "")
					copy(keys[pos+1:], keys[pos:])
					keys[pos] = k
					next = append(next, PlaCube{})
					copy(next[pos+1:], next[pos:])
					next[pos] = c
				}
				used[i], used[j] = true, true
			}
		}
		merged := 0
		for _, u := range used {
			if u {
				merged++
			}
		}
		if merged == 0 {
			break
		}
		total += merged - len(next)
		for i, c := range candidate.cubes {
			if !used[i] {
				next = append(next, clonePlaCube(c))
			}
		}
		if len(next) > p.limits.MaxCubes {
			return 0, ErrLimit
		}
		candidate.cubes = next
	}
	p.cubes = candidate.cubes
	return total, nil
}

func plaTypeName(t PlaFileType) string { return [...]string{"fd", "f", "fr", "fdr", "???"}[t] }
func plaType(text string) (PlaFileType, bool) {
	for i, s := range []string{"fd", "f", "fr", "fdr", "???"} {
		if text == s {
			return PlaFileType(i), true
		}
	}
	return PlaFileNone, false
}
func plaLiteralByte(l PlaLiteral) byte { return "-01?"[l] }

func (p *PlaManager) MarshalPLA(timestamp string) ([]byte, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	if p.limits.Stop != nil && p.limits.Stop() {
		return nil, ErrCancelled
	}
	if strings.ContainsAny(timestamp, "\r\n") {
		return nil, ErrDimension
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "# SOP %q written via PLA package in ABC on %s\n\n", p.name, timestamp)
	if p.typ != PlaFileFD {
		fmt.Fprintf(&b, ".type %s\n", plaTypeName(p.typ))
	}
	fmt.Fprintf(&b, ".i %d\n.o %d\n.p %d\n", p.inputs, p.outputs, len(p.cubes))
	for _, c := range p.cubes {
		if p.limits.Stop != nil && p.limits.Stop() {
			return nil, ErrCancelled
		}
		for _, l := range c.Inputs {
			b.WriteByte(plaLiteralByte(l))
		}
		b.WriteByte(' ')
		for _, l := range c.Outputs {
			b.WriteByte(plaLiteralByte(l))
		}
		b.WriteByte('\n')
		if b.Len() > p.limits.MaxBytes {
			return nil, ErrLimit
		}
	}
	b.WriteString(".e\n\n")
	if b.Len() > p.limits.MaxBytes {
		return nil, ErrLimit
	}
	return append([]byte(nil), b.Bytes()...), nil
}

func ParsePLA(data []byte, name string, limits PlaLimits) (*PlaManager, error) {
	limits, err := normalizePlaLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, ErrDimension
	}
	typ := PlaFileFD
	ins, outs, declared := -1, -1, -1
	rows := make([][2]string, 0)
	ended := false
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, raw := range lines {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		line := strings.TrimSpace(strings.SplitN(raw, "#", 2)[0])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if strings.HasPrefix(line, ".") {
			if ended {
				return nil, ErrDimension
			}
			switch fields[0] {
			case ".i", ".o", ".p":
				if len(fields) != 2 {
					return nil, ErrDimension
				}
				n, e := strconv.Atoi(fields[1])
				if e != nil || n < 0 {
					return nil, ErrDimension
				}
				if fields[0] == ".i" {
					if ins != -1 {
						return nil, ErrDimension
					}
					ins = n
				} else if fields[0] == ".o" {
					if outs != -1 {
						return nil, ErrDimension
					}
					outs = n
				} else {
					if declared != -1 {
						return nil, ErrDimension
					}
					declared = n
				}
			case ".type":
				if len(fields) != 2 {
					return nil, ErrDimension
				}
				var ok bool
				typ, ok = plaType(fields[1])
				if !ok {
					return nil, ErrDimension
				}
			case ".e", ".end":
				if len(fields) != 1 {
					return nil, ErrDimension
				}
				ended = true
			default:
				return nil, ErrDimension
			}
			continue
		}
		if ended || len(fields) != 2 {
			return nil, ErrDimension
		}
		rows = append(rows, [2]string{fields[0], fields[1]})
		if len(rows) > limits.MaxCubes {
			return nil, ErrLimit
		}
	}
	if ins <= 0 || outs <= 0 || ins > limits.MaxInputs || outs > limits.MaxOutputs || declared >= 0 && declared != len(rows) {
		return nil, ErrDimension
	}
	p, err := NewPlaManager(name, name, ins, outs, typ, limits)
	if err != nil {
		return nil, err
	}
	parseLit := func(ch byte, out bool) (PlaLiteral, bool) {
		switch ch {
		case '-', '2':
			return PlaLitDash, true
		case '0':
			return PlaLitZero, true
		case '1':
			return PlaLitOne, true
		case '?':
			return PlaLitFull, true
		case '~':
			if !out {
				return 0, false
			}
			switch typ {
			case PlaFileF, PlaFileFD:
				return PlaLitZero, true
			case PlaFileFR:
				return PlaLitDash, true
			case PlaFileFDR:
				return PlaLitFull, true
			}
		}
		return 0, false
	}
	for _, row := range rows {
		if len(row[0]) != ins || len(row[1]) != outs {
			return nil, ErrDimension
		}
		ci := make([]PlaLiteral, ins)
		co := make([]PlaLiteral, outs)
		for i := range ci {
			var ok bool
			ci[i], ok = parseLit(row[0][i], false)
			if !ok {
				return nil, ErrDimension
			}
		}
		for i := range co {
			var ok bool
			co[i], ok = parseLit(row[1][i], true)
			if !ok {
				return nil, ErrDimension
			}
		}
		if err = p.AddCube(ci, co); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func ReadPLAFile(path string, limits PlaLimits) (*PlaManager, error) {
	data, err := plaReadFileLimited(path, limits.MaxBytes)
	if err != nil {
		return nil, err
	}
	return ParsePLA(data, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), limits)
}

// ReadFileLimited is intentionally supplied by callers through bytes in most
// uses; this filesystem helper uses the package's bounded utility.
func plaReadFileLimited(path string, maximum int) ([]byte, error) {
	if maximum < 0 {
		return nil, ErrDimension
	}
	if maximum == 0 {
		maximum = math.MaxInt32
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > int64(maximum) {
		return nil, ErrLimit
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) > maximum {
		return nil, ErrLimit
	}
	return data, nil
}

func (p *PlaManager) WritePLAFile(path, timestamp string) error {
	data, err := p.MarshalPLA(timestamp)
	if err != nil {
		return err
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".pla-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	ok := false
	defer func() {
		temporary.Close()
		if !ok {
			_ = os.Remove(temporaryName)
		}
	}()
	if _, err = temporary.Write(data); err != nil {
		return err
	}
	if err = temporary.Sync(); err != nil {
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	if err = os.Rename(temporaryName, path); err != nil {
		return err
	}
	ok = true
	return nil
}

type PlaFXDivisor struct{ First, Second int }
type PlaFXNetwork struct {
	Inputs      int
	OutputCount int
	Terms       [][]int
	Outputs     [][]PlaLiteral
	Divisors    []PlaFXDivisor
}

func plaInputTerms(p *PlaManager) ([][]int, [][]PlaLiteral) {
	terms := make([][]int, len(p.cubes))
	outs := make([][]PlaLiteral, len(p.cubes))
	for i, c := range p.cubes {
		for v, l := range c.Inputs {
			if l == PlaLitZero {
				terms[i] = append(terms[i], 2*v+1)
			} else if l == PlaLitOne {
				terms[i] = append(terms[i], 2*v)
			}
		}
		outs[i] = append([]PlaLiteral(nil), c.Outputs...)
	}
	return terms, outs
}

func (p *PlaManager) FastExtract() (*PlaFXNetwork, error) {
	if err := p.check(); err != nil {
		return nil, err
	}
	terms, outs := plaInputTerms(p)
	divs := make([]PlaFXDivisor, 0)
	ops := 0
	for {
		bestA, bestB, bestCount := -1, -1, 1
		for _, t := range terms {
			for i := 0; i < len(t); i++ {
				for j := i + 1; j < len(t); j++ {
					ops++
					if ops > p.limits.MaxOperations {
						return nil, ErrLimit
					}
					a, b := t[i], t[j]
					if a > b {
						a, b = b, a
					}
					count := 0
					for _, u := range terms {
						ops++
						if ops > p.limits.MaxOperations {
							return nil, ErrLimit
						}
						if plaHas(u, a) && plaHas(u, b) {
							count++
						}
					}
					if count > bestCount || count == bestCount && (bestA < 0 || a < bestA || a == bestA && b < bestB) {
						bestA, bestB, bestCount = a, b, count
					}
				}
			}
		}
		if bestA < 0 {
			break
		}
		newLit := 2 * (p.inputs + len(divs))
		divs = append(divs, PlaFXDivisor{bestA, bestB})
		for i, t := range terms {
			if plaHas(t, bestA) && plaHas(t, bestB) {
				u := make([]int, 0, len(t)-1)
				for _, l := range t {
					if l != bestA && l != bestB {
						u = append(u, l)
					}
				}
				u = append(u, newLit)
				sort.Ints(u)
				terms[i] = u
			}
		}
		if p.limits.Stop != nil && p.limits.Stop() {
			return nil, ErrCancelled
		}
	}
	return &PlaFXNetwork{Inputs: p.inputs, OutputCount: p.outputs, Terms: terms, Outputs: outs, Divisors: divs}, nil
}
func plaHas(a []int, x int) bool { i := sort.SearchInts(a, x); return i < len(a) && a[i] == x }
func (n *PlaFXNetwork) Evaluate(assignment uint64) ([]bool, error) {
	outputCount, err := n.validate()
	if err != nil {
		return nil, err
	}
	if n.Inputs < 64 && assignment >= uint64(1)<<n.Inputs {
		return nil, ErrDimension
	}
	values := make([]bool, n.Inputs+len(n.Divisors))
	for i := 0; i < n.Inputs; i++ {
		values[i] = (assignment>>i)&1 != 0
	}
	for i, d := range n.Divisors {
		values[n.Inputs+i] = plaEvalLit(values, d.First) && plaEvalLit(values, d.Second)
	}
	r := make([]bool, outputCount)
	for i, t := range n.Terms {
		match := true
		for _, l := range t {
			if !plaEvalLit(values, l) {
				match = false
				break
			}
		}
		if match {
			for o, l := range n.Outputs[i] {
				r[o] = r[o] || l == PlaLitOne || l == PlaLitFull
			}
		}
	}
	return r, nil
}

func (n *PlaFXNetwork) validate() (int, error) {
	if n == nil || n.Inputs <= 0 || n.Inputs >= 64 || len(n.Terms) != len(n.Outputs) {
		return 0, ErrDimension
	}
	outputCount := n.OutputCount
	if outputCount == 0 && len(n.Outputs) > 0 {
		outputCount = len(n.Outputs[0])
	}
	if outputCount <= 0 {
		return 0, ErrInvariant
	}
	variableCount := n.Inputs + len(n.Divisors)
	for i, d := range n.Divisors {
		for _, literal := range []int{d.First, d.Second} {
			if literal < 0 || literal>>1 >= n.Inputs+i {
				return 0, ErrInvariant
			}
		}
	}
	for i, term := range n.Terms {
		if len(n.Outputs[i]) != outputCount {
			return 0, ErrInvariant
		}
		for _, literal := range term {
			if literal < 0 || literal>>1 >= variableCount {
				return 0, ErrInvariant
			}
		}
		for _, literal := range n.Outputs[i] {
			if !validPlaLiteral(literal) {
				return 0, ErrInvariant
			}
		}
	}
	return outputCount, nil
}

func (n *PlaFXNetwork) MarshalBLIF(model, timestamp string, maxBytes int) ([]byte, error) {
	outputCount, validationErr := n.validate()
	if validationErr != nil {
		return nil, validationErr
	}
	if strings.ContainsAny(model+timestamp, "\r\n") || maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = math.MaxInt32
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "# BLIF network written by ABC on %s\n\n.model %s\n.inputs", timestamp, model)
	for variable := 0; variable < n.Inputs; variable++ {
		fmt.Fprintf(&b, " i%d", variable)
	}
	b.WriteString("\n.outputs")
	for output := 0; output < outputCount; output++ {
		fmt.Fprintf(&b, " o%d", output)
	}
	b.WriteByte('\n')
	variableCount := n.Inputs + len(n.Divisors)
	for output := 0; output < outputCount; output++ {
		b.WriteString(".names")
		for variable := 0; variable < variableCount; variable++ {
			fmt.Fprintf(&b, " i%d", variable)
		}
		fmt.Fprintf(&b, " o%d\n", output)
		for cube, term := range n.Terms {
			literal := n.Outputs[cube][output]
			if literal != PlaLitOne && literal != PlaLitFull {
				continue
			}
			row := bytes.Repeat([]byte{'-'}, variableCount)
			for _, inputLiteral := range term {
				variable := inputLiteral >> 1
				if variable < 0 || variable >= variableCount {
					return nil, ErrInvariant
				}
				row[variable] = '1'
				if inputLiteral&1 != 0 {
					row[variable] = '0'
				}
			}
			b.Write(row)
			b.WriteString(" 1\n")
		}
	}
	for index, divisor := range n.Divisors {
		first, second := divisor.First>>1, divisor.Second>>1
		if first < 0 || first >= n.Inputs+index || second < 0 || second >= n.Inputs+index {
			return nil, ErrInvariant
		}
		fmt.Fprintf(&b, ".names i%d i%d i%d\n", first, second, n.Inputs+index)
		fmt.Fprintf(&b, "%d%d 1\n", 1-(divisor.First&1), 1-(divisor.Second&1))
	}
	b.WriteString(".end\n\n")
	if b.Len() > maxBytes {
		return nil, ErrLimit
	}
	return append([]byte(nil), b.Bytes()...), nil
}

func plaEvalLit(values []bool, l int) bool {
	v := values[l>>1]
	if l&1 != 0 {
		return !v
	}
	return v
}

func PlaPrimesTable(variables int, limits PlaLimits) ([]bool, error) {
	limits, err := normalizePlaLimits(limits)
	if err != nil {
		return nil, err
	}
	if variables < 1 || variables >= 31 || variables > limits.MaxInputs {
		return nil, ErrDimension
	}
	size := 1 << variables
	if size > limits.MaxOperations {
		return nil, ErrLimit
	}
	r := make([]bool, size)
	for i := 2; i < size; i++ {
		r[i] = true
	}
	for n := 2; n*n < size; n++ {
		if r[n] {
			for i := n * n; i < size; i += n {
				r[i] = false
			}
		}
	}
	return r, nil
}

func PlaSimpleMinimize(on []bool, variables int, name string, limits PlaLimits) (*PlaManager, error) {
	limits, err := normalizePlaLimits(limits)
	if err != nil {
		return nil, err
	}
	if variables <= 0 || variables >= 31 || len(on) != 1<<variables {
		return nil, ErrDimension
	}
	count := 0
	for _, set := range on {
		if set {
			count++
		}
	}
	if count > limits.MaxCubes {
		return nil, ErrLimit
	}
	p, err := NewPlaManager(name, name, variables, 1, PlaFileFD, limits)
	if err != nil {
		return nil, err
	}
	for assignment, set := range on {
		if !set {
			continue
		}
		inputs := make([]PlaLiteral, variables)
		for variable := range inputs {
			if assignment&(1<<variable) != 0 {
				inputs[variable] = PlaLitOne
			} else {
				inputs[variable] = PlaLitZero
			}
		}
		if err = p.AddCube(inputs, []PlaLiteral{PlaLitOne}); err != nil {
			return nil, err
		}
	}
	if _, err = p.CombineDistanceOne(); err != nil {
		return nil, err
	}
	return p, nil
}

func PlaPrimesDetector(variables int, limits PlaLimits) (*PlaManager, error) {
	table, err := PlaPrimesTable(variables, limits)
	if err != nil {
		return nil, err
	}
	return PlaSimpleMinimize(table, variables, fmt.Sprintf("primes%02d", variables), limits)
}

func PlaFromMinterms(name string, variables int, minterms []int, limits PlaLimits) (*PlaManager, error) {
	limits, err := normalizePlaLimits(limits)
	if err != nil {
		return nil, err
	}
	if variables <= 0 || variables >= 31 {
		return nil, ErrDimension
	}
	if len(minterms) > limits.MaxCubes || len(minterms) > limits.MaxOperations {
		return nil, ErrLimit
	}
	maximum := 1 << variables
	seen := make([]bool, maximum)
	manager, err := NewPlaManager(name, name, variables, 1, PlaFileFD, limits)
	if err != nil {
		return nil, err
	}
	for _, minterm := range minterms {
		if minterm < 0 || minterm >= maximum || seen[minterm] {
			return nil, ErrDimension
		}
		seen[minterm] = true
		inputs := make([]PlaLiteral, variables)
		for variable := range inputs {
			if minterm>>variable&1 != 0 {
				inputs[variable] = PlaLitOne
			} else {
				inputs[variable] = PlaLitZero
			}
		}
		if err = manager.AddCube(inputs, []PlaLiteral{PlaLitOne}); err != nil {
			return nil, err
		}
	}
	return manager, nil
}

// GenerateSorterPLA returns the source generator's population-count sorter:
// each output row contains all ones before all zeroes.
func GenerateSorterPLA(variables int, limits PlaLimits) (*PlaManager, error) {
	limits, err := normalizePlaLimits(limits)
	if err != nil {
		return nil, err
	}
	if variables <= 0 || variables >= 31 || variables > limits.MaxInputs || variables > limits.MaxOutputs {
		return nil, ErrDimension
	}
	count := (1 << variables) - 1
	if count > limits.MaxCubes || count > limits.MaxOperations {
		return nil, ErrLimit
	}
	manager, err := NewPlaManager(fmt.Sprintf("sorter%02d", variables), "", variables, variables, PlaFileFD, limits)
	if err != nil {
		return nil, err
	}
	for assignment := 1; assignment < 1<<variables; assignment++ {
		inputs := make([]PlaLiteral, variables)
		ones := 0
		for position := 0; position < variables; position++ {
			bit := (assignment >> (variables - 1 - position)) & 1
			if bit != 0 {
				inputs[position] = PlaLitOne
				ones++
			} else {
				inputs[position] = PlaLitZero
			}
		}
		outputs := make([]PlaLiteral, variables)
		for output := 0; output < ones; output++ {
			outputs[output] = PlaLitOne
		}
		if err = manager.AddCube(inputs, outputs); err != nil {
			return nil, err
		}
	}
	return manager, nil
}

func GeneratePLA(inputs, outputs, cubes int, seed int64, limits PlaLimits) (*PlaManager, error) {
	limits, err := normalizePlaLimits(limits)
	if err != nil {
		return nil, err
	}
	if inputs <= 0 || inputs >= 31 || outputs <= 0 || outputs >= 31 || cubes <= 0 || cubes > 1<<inputs || cubes > limits.MaxCubes {
		return nil, ErrDimension
	}
	if 1<<inputs > limits.MaxOperations {
		return nil, ErrLimit
	}
	random := rand.New(rand.NewSource(seed))
	selected := random.Perm(1 << inputs)[:cubes]
	sort.Ints(selected)
	p, err := NewPlaManager(fmt.Sprintf("rand_%d_%d_%d", inputs, outputs, cubes), "", inputs, outputs, PlaFileFD, limits)
	if err != nil {
		return nil, err
	}
	for _, assignment := range selected {
		inputLiterals := make([]PlaLiteral, inputs)
		for variable := range inputLiterals {
			if assignment&(1<<variable) != 0 {
				inputLiterals[variable] = PlaLitOne
			} else {
				inputLiterals[variable] = PlaLitZero
			}
		}
		outputLiterals := make([]PlaLiteral, outputs)
		pattern := 1
		if outputs > 1 {
			pattern = 1 + random.Intn((1<<outputs)-1)
		}
		for output := range outputLiterals {
			if pattern&(1<<output) != 0 {
				outputLiterals[output] = PlaLitOne
			} else {
				outputLiterals[output] = PlaLitZero
			}
		}
		if err = p.AddCube(inputLiterals, outputLiterals); err != nil {
			return nil, err
		}
	}
	return p, nil
}
