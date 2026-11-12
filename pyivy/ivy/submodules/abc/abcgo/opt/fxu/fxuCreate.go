// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

import (
	"fmt"
	"sort"
	"strings"
)

type sopCover struct {
	patterns []string
	phase    bool
}

func parseSOP(text string, width int) (sopCover, error) {
	cover := sopCover{phase: true}
	lines := strings.Split(strings.TrimRight(text, "\x00"), "\n")
	phaseSet := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		pattern, out := "", ""
		if width == 0 && len(fields) == 1 {
			out = fields[0]
		} else if len(fields) == 2 {
			pattern, out = fields[0], fields[1]
		} else {
			return cover, fmt.Errorf("invalid SOP row %q", line)
		}
		if len(pattern) != width || out != "0" && out != "1" {
			return cover, fmt.Errorf("SOP row width/polarity mismatch")
		}
		for _, c := range []byte(pattern) {
			if c != '0' && c != '1' && c != '-' {
				return cover, fmt.Errorf("invalid SOP literal %q", c)
			}
		}
		phase := out == "1"
		if phaseSet && phase != cover.phase {
			return cover, fmt.Errorf("mixed SOP output phases")
		}
		cover.phase = phase
		phaseSet = true
		cover.patterns = append(cover.patterns, pattern)
	}
	return cover, nil
}
func patternSubsumes(a, b string) bool {
	for i, c := range []byte(a) {
		if c != '-' && c != b[i] {
			return false
		}
	}
	return true
}
func Fxu_CreateMatrix(data *Fxu_Data_t) *Fxu_Matrix {
	p, err := createMatrix(data)
	data.LastError = err
	return p
}
func createMatrix(data *Fxu_Data_t) (*Fxu_Matrix, error) {
	if len(data.SOPs) != len(data.Fanins) {
		return nil, fmt.Errorf("FXU SOP/fanin vector lengths differ")
	}
	if data.NodesExt < 0 || data.SingleMax < 0 || data.PairsMax < 0 {
		return nil, fmt.Errorf("negative FXU resource limit")
	}
	if len(data.SOPs) > 500000 {
		return nil, fmt.Errorf("FXU exceeds source variable limit")
	}
	if err := validateConnections(data.Fanins); err != nil {
		return nil, err
	}
	p := Fxu_MatrixAllocate()
	p.OldNodes = int32(len(data.SOPs))
	p.params = data
	for i := 0; i < 2*len(data.SOPs); i++ {
		Fxu_MatrixAddVar(p)
	}
	eligible := 0
	var pairStore, pairCount int64
	for owner, sop := range data.SOPs {
		if sop == "" {
			continue
		}
		fans := data.Fanins[owner]
		seen := make(map[int32]bool)
		for _, v := range fans {
			if v < 0 || int(v) >= len(data.SOPs) || v == int32(owner) || seen[v] {
				return nil, fmt.Errorf("node%d has invalid or duplicate fanins", owner)
			}
			seen[v] = true
		}
		cover, err := parseSOP(sop, len(fans))
		if err != nil {
			return nil, fmt.Errorf("node%d: %w", owner, err)
		}
		p.phases[int32(owner)] = cover.phase
		if len(fans) < 2 || len(cover.patterns) == 0 {
			continue
		}
		eligible++
		n := int64(len(cover.patterns))
		pairStore += n * n
		pairCount += n * (n - 1) / 2
		if pairStore > 50000000 || pairCount > 10000000 {
			return nil, fmt.Errorf("FXU exceeds source cube-pair limit")
		}
		for i, a := range cover.patterns {
			for _, b := range cover.patterns[i+1:] {
				if patternSubsumes(a, b) || patternSubsumes(b, a) {
					return nil, fmt.Errorf("node%d cover is not SCC-free (duplicate or subsumed cubes)", owner)
				}
			}
		}
		for i, pattern := range cover.patterns {
			cube := Fxu_MatrixAddCube(p, p.Vars[2*owner+1], int32(i))
			columns := make([]int32, 0, len(fans))
			for k, c := range []byte(pattern) {
				if c == '1' {
					columns = append(columns, fans[k]*2)
				} else if c == '0' {
					columns = append(columns, fans[k]*2+1)
				}
			}
			for _, v := range sortedColumns(columns) {
				Fxu_MatrixAddLiteral(p, cube, p.Vars[v])
			}
		}
	}
	if eligible == 0 {
		return nil, fmt.Errorf("network has no SOPs eligible for extraction")
	}
	p.weightLimit = 1
	if data.Use0 {
		p.weightLimit = 0
	}
	rebuildDoubles(p, true)
	Fxu_MatrixComputeSingles(p, boolInt(data.Use0), data.SingleMax)
	return p, nil
}
func Fxu_CreateMatrixAddCube(p *Fxu_Matrix, cube *Fxu_Cube, pattern string, fans, order []int32) {
	for _, i := range order {
		if pattern[i] == '-' {
			continue
		}
		column := fans[i] * 2
		if pattern[i] == '0' {
			column++
		}
		Fxu_MatrixAddLiteral(p, cube, p.Vars[column])
	}
}
func Fxu_CreateCovers(p *Fxu_Matrix, data *Fxu_Data_t) {
	nodes := len(p.Vars) / 2
	data.NewSOPs = make([]string, nodes)
	copy(data.NewSOPs, data.SOPs)
	data.NewFanins = make([][]int32, nodes)
	for i, fans := range data.Fanins {
		data.NewFanins[i] = append([]int32(nil), fans...)
	}
	owners := make([]int, 0, len(p.changed))
	for owner := range p.changed {
		owners = append(owners, int(owner))
	}
	sort.Ints(owners)
	for _, owner := range owners {
		Fxu_CreateCoversNode(p, data, int32(owner))
	}
}
func Fxu_CreateCoversNode(p *Fxu_Matrix, data *Fxu_Data_t, owner int32) {
	present := make(map[int32]bool)
	for _, c := range p.covers[owner] {
		if c.deleted {
			continue
		}
		for _, lit := range c.Literals {
			present[lit.Var.Index/2] = true
		}
	}
	fans := make([]int32, 0, len(present))
	for v := range present {
		fans = append(fans, v)
	}
	sort.Slice(fans, func(i, j int) bool { return fans[i] < fans[j] })
	positions := make(map[int32]int, len(fans))
	for i, v := range fans {
		positions[v] = i
	}
	phase := '1'
	if normal, ok := p.phases[owner]; ok && !normal {
		phase = '0'
	}
	var sop strings.Builder
	for _, c := range p.covers[owner] {
		if c.deleted {
			continue
		}
		row := []byte(strings.Repeat("-", len(fans)))
		for _, lit := range c.Literals {
			value := byte('0')
			if columnPositive(p, lit.Var.Index) {
				value = '1'
			}
			row[positions[lit.Var.Index/2]] = value
		}
		fmt.Fprintf(&sop, "%s %c\n", row, phase)
	}
	data.NewSOPs[owner] = sop.String()
	data.NewFanins[owner] = fans
}
func Fxu_CreateCoversFirstCube(p *Fxu_Matrix, data *Fxu_Data_t, owner int32) *Fxu_Cube {
	for id := owner; int(id) < len(p.Vars)/2; id++ {
		for _, c := range p.covers[id] {
			if !c.deleted {
				return c
			}
		}
	}
	return nil
}
