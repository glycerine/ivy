package exor

import (
	"encoding/binary"
	"fmt"
	"slices"
	"sort"
)

func cubeKey(c *Cube) string {
	b := make([]byte, len(c.Inputs)*4)
	for i, w := range c.Inputs {
		binary.LittleEndian.PutUint32(b[i*4:], w)
	}
	return string(b)
}
func zeroOutputs(c *Cube) bool {
	for _, w := range c.Outputs {
		if w != 0 {
			return false
		}
	}
	return true
}

// normalize cancels identical product terms output by output. Zero products
// disappear; all remaining value sets are valid nonempty binary subsets.
func normalize(cubes []*Cube) []*Cube {
	table := make(map[string]*Cube, len(cubes))
	var order []string
	for _, c := range cubes {
		key := cubeKey(c)
		if prior := table[key]; prior != nil {
			for i, w := range c.Outputs {
				prior.Outputs[i] ^= w
			}
		} else {
			table[key] = c.Clone()
			order = append(order, key)
		}
	}
	out := make([]*Cube, 0, len(table))
	for _, key := range order {
		if c := table[key]; !zeroOutputs(c) {
			out = append(out, c)
		}
	}
	return out
}
func (m *Manager) reindex() {
	for _, c := range m.Cubes {
		if c.ID == 0 {
			c.ID = m.nextID
			m.nextID++
		}
		c.active = true
	}
	m.Stats.CubesAfter = len(m.Cubes)
	m.Stats.LiteralsAfter = CountLiterals(m)
	m.Stats.QCostAfter = CountQCost(m)
}
func (m *Manager) Normalize() {
	for _, c := range m.Cubes {
		c.active = false
	}
	m.Cubes = normalize(m.Cubes)
	for {
		found := false
		for i := 0; i < len(m.Cubes) && !found; i++ {
			for j := i + 1; j < len(m.Cubes); j++ {
				a, b := m.Cubes[i], m.Cubes[j]
				if !slices.Equal(a.Outputs, b.Outputs) {
					continue
				}
				d := FindDiffVars(a, b)
				if len(d) == 1 && d[0] >= 0 {
					a.setVar(d[0], GetVar(a, d[0])^GetVar(b, d[0]))
					m.Cubes = append(m.Cubes[:j], m.Cubes[j+1:]...)
					m.Cubes = normalize(m.Cubes)
					found = true
					break
				}
			}
		}
		if !found {
			break
		}
	}
	m.reindex()
}
func CubeInsert(m *Manager, c *Cube) {
	if c.active {
		panic("exor: cube already active")
	}
	if err := m.validate(c); err != nil {
		panic(err)
	}
	c.active = true
	m.Cubes = append(m.Cubes, c)
}
func CubeExtract(m *Manager, c *Cube) *Cube {
	for i, x := range m.Cubes {
		if x == c {
			m.Cubes = append(m.Cubes[:i], m.Cubes[i+1:]...)
			c.active = false
			return c
		}
	}
	return nil
}
func IterCubeSetStart(m *Manager) *Cube { m.iter = 0; return IterCubeSetNext(m) }
func IterCubeSetNext(m *Manager) *Cube {
	if m.iter >= len(m.Cubes) {
		return nil
	}
	c := m.Cubes[m.iter]
	m.iter++
	return c
}
func CheckForCloseCubes(m *Manager, c *Cube, add bool) int {
	before := len(m.Cubes)
	if add {
		if err := m.Add(c); err != nil {
			panic(err)
		}
	} else {
		for _, other := range m.Cubes {
			if GetDistance(c, other) <= 1 {
				return 1
			}
		}
		return 0
	}
	m.Normalize()
	return before + 1 - len(m.Cubes)
}
func CheckAndInsert(m *Manager, c *Cube) int { return CheckForCloseCubes(m, c, true) }
func UndoRecentChanges(m *Manager) {
	if m.undo == nil {
		panic("exor: no saved cover")
	}
	m.Cubes = m.undo
	m.undo = nil
	m.reindex()
}
func AllocateCubeSets(m *Manager) int { m.Cubes = make([]*Cube, 0); return (m.Inputs + m.Outputs) * 8 }
func DelocateCubeSets(m *Manager) {
	for _, c := range m.Cubes {
		c.active = false
	}
	m.Cubes = nil
}
func AllocateQueques(m *Manager, n int) int {
	if n < 0 {
		panic("exor: negative queue allocation")
	}
	m.pairs = make([]cubePair, 0, n)
	m.pairIndex = 0
	return n * 24
}
func DelocateQueques(m *Manager) { m.pairs = nil; m.pairIndex = 0 }
func NewRangeReset(m *Manager)   { m.rangeStart = len(m.pairs) }
func NewRangeInsertCubePair(m *Manager, d CubeDist, a, b *Cube) {
	distance := GetDistance(a, b)
	if distance != int(d)+2 {
		panic("exor: incorrect queue distance")
	}
	m.pairs = append(m.pairs, cubePair{a, b, distance})
}
func NewRangeAdd(m *Manager) {
	for i := range m.Cubes {
		for j := i + 1; j < len(m.Cubes); j++ {
			d := GetDistance(m.Cubes[i], m.Cubes[j])
			if d >= 2 && d <= 4 {
				m.pairs = append(m.pairs, cubePair{m.Cubes[i], m.Cubes[j], d})
			}
		}
	}
}
func MarkSet(m *Manager)            { m.mark = m.pairIndex }
func MarkRewind(m *Manager)         { m.pairIndex = m.mark }
func GetPosDiff(begin, end int) int { return end - begin }
func IteratorCubePairStart(m *Manager, d CubeDist) (*Cube, *Cube) {
	m.pairs = nil
	NewRangeAdd(m)
	selected := m.pairs[:0]
	for _, pair := range m.pairs {
		if pair.distance == int(d)+2 {
			selected = append(selected, pair)
		}
	}
	m.pairs = selected
	m.pairIndex = 0
	return IteratorCubePairNext(m)
}
func IteratorCubePairNext(m *Manager) (*Cube, *Cube) {
	for m.pairIndex < len(m.pairs) {
		pair := m.pairs[m.pairIndex]
		m.pairIndex++
		if pair.a.active && pair.b.active {
			m.current = pair
			return pair.a, pair.b
		}
	}
	return nil, nil
}
func GetQuequeStats(m *Manager, d CubeDist) int {
	n := 0
	for _, p := range m.pairs {
		if p.distance == int(d)+2 && p.a.active && p.b.active {
			n++
		}
	}
	return n
}
func PrintQuequeStats(m *Manager) {
	fmt.Printf("ExorLink queues: %d %d %d\n", GetQuequeStats(m, DIST2), GetQuequeStats(m, DIST3), GetQuequeStats(m, DIST4))
}

type coverCost struct{ cubes, literals, quantum int }

func (m *Manager) cost(cubes []*Cube) coverCost {
	cost := coverCost{cubes: len(cubes)}
	for _, c := range cubes {
		cost.literals += literalCount(c)
		cost.quantum += ComputeQCostBits(c)
	}
	return cost
}
func (m *Manager) better(a, b coverCost) bool {
	if m.Options.UseQCost && a.quantum != b.quantum {
		return a.quantum < b.quantum
	}
	if a.cubes != b.cubes {
		return a.cubes < b.cubes
	}
	if a.literals != b.literals {
		return a.literals < b.literals
	}
	return a.quantum < b.quantum
}
func (m *Manager) applyDistance(distance int) int {
	m.Normalize()
	baseline := m.cost(m.Cubes)
	var best []*Cube
	bestCost := baseline
	for i, a := range m.Cubes {
		for j := i + 1; j < len(m.Cubes); j++ {
			b := m.Cubes[j]
			if GetDistance(a, b) != distance {
				continue
			}
			for _, group := range ExorLink(a, b, distance) {
				candidate := make([]*Cube, 0, len(m.Cubes)+distance-2)
				for k, c := range m.Cubes {
					if k != i && k != j {
						candidate = append(candidate, c)
					}
				}
				candidate = append(candidate, group...)
				trial := &Manager{Inputs: m.Inputs, Outputs: m.Outputs, Options: m.Options, Cubes: normalize(candidate)}
				trial.Normalize()
				cost := m.cost(trial.Cubes)
				if m.better(cost, bestCost) {
					best, bestCost = trial.Cubes, cost
				}
			}
		}
	}
	if best == nil {
		return 0
	}
	m.undo = cloneCubes(m.Cubes)
	for _, c := range m.Cubes {
		c.active = false
	}
	m.Cubes = best
	m.reindex()
	m.Stats.Rewrites++
	return 1
}
func IterativelyApplyExorLink2(m *Manager, flags int) int { return m.applyDistance(2) }
func IterativelyApplyExorLink3(m *Manager, flags int) int { return m.applyDistance(3) }
func IterativelyApplyExorLink4(m *Manager, flags int) int { return m.applyDistance(4) }
func (m *Manager) Sort() {
	sort.SliceStable(m.Cubes, func(i, j int) bool { return cubeKey(m.Cubes[i]) < cubeKey(m.Cubes[j]) })
}

// coverKey ignores cube order and IDs, so neutral exploration cannot loop.
func coverKey(cubes []*Cube) string {
	keys := make([]string, len(cubes))
	for i, c := range cubes {
		b := make([]byte, len(c.Outputs)*4)
		for j, w := range c.Outputs {
			binary.LittleEndian.PutUint32(b[4*j:], w)
		}
		keys[i] = cubeKey(c) + string(b)
	}
	sort.Strings(keys)
	return fmt.Sprintf("%q", keys)
}

// Quality bounds neutral ExorLink exploration. Only a strictly improved cover
// is committed; failed exploratory states never leak to the caller.
func (m *Manager) escapePlateau(budget int) bool {
	baseline := m.cost(m.Cubes)
	trial := &Manager{Inputs: m.Inputs, Outputs: m.Outputs, Options: m.Options, Cubes: cloneCubes(m.Cubes)}
	seen := map[string]bool{coverKey(trial.Cubes): true}
	for step := 0; step < budget; step++ {
		var next []*Cube
	outer:
		for i, a := range trial.Cubes {
			for j := i + 1; j < len(trial.Cubes); j++ {
				b := trial.Cubes[j]
				distance := GetDistance(a, b)
				if distance < 2 || distance > 4 {
					continue
				}
				for _, group := range ExorLink(a, b, distance) {
					var candidate []*Cube
					for k, c := range trial.Cubes {
						if k != i && k != j {
							candidate = append(candidate, c)
						}
					}
					candidate = append(candidate, group...)
					probe := &Manager{Inputs: m.Inputs, Outputs: m.Outputs, Options: m.Options, Cubes: normalize(candidate)}
					probe.Normalize()
					cost := m.cost(probe.Cubes)
					if m.better(cost, baseline) {
						next = probe.Cubes
						break outer
					}
					if cost != baseline {
						continue
					}
					key := coverKey(probe.Cubes)
					if !seen[key] {
						seen[key] = true
						next = probe.Cubes
						break outer
					}
				}
			}
		}
		if next == nil {
			return false
		}
		trial.Cubes = next
		for {
			gain := 0
			for d := 2; d <= 4; d++ {
				gain += trial.applyDistance(d)
			}
			if gain == 0 {
				break
			}
		}
		if m.better(m.cost(trial.Cubes), baseline) {
			m.undo = cloneCubes(m.Cubes)
			m.Cubes = trial.Cubes
			m.Stats.Rewrites += trial.Stats.Rewrites + step + 1
			m.reindex()
			return true
		}
	}
	return false
}
