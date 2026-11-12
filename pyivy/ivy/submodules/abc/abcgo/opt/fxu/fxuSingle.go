package fxu

import "sort"

func Fxu_SingleCountCoincidence(p *Fxu_Matrix, a, b *Fxu_Var) int32 {
	present := make(map[*Fxu_Cube]bool, len(a.Literals))
	for _, lit := range a.Literals {
		if !lit.Cube.deleted {
			present[lit.Cube] = true
		}
	}
	count := int32(0)
	for _, lit := range b.Literals {
		if present[lit.Cube] {
			count++
		}
	}
	return count
}
func collectSingleCounts(p *Fxu_Matrix) map[[2]int32]int32 {
	counts := make(map[[2]int32]int32)
	for _, c := range p.Cubes {
		if c.deleted {
			continue
		}
		columns := c.columns()
		for i, a := range columns {
			for _, b := range columns[i+1:] {
				counts[[2]int32{a, b}]++
			}
		}
	}
	return counts
}
func rebuildSingles(p *Fxu_Matrix, initial bool, maximum int32) {
	old := p.singleCatalog
	p.singleHeap.clear()
	p.singleCatalog = make(map[[2]int32]*Fxu_Single)
	p.Singles = nil
	type entry struct {
		key    [2]int32
		weight int32
	}
	var entries []entry
	for key, count := range collectSingleCounts(p) {
		weight := count - 2
		threshold := p.weightLimit
		if !initial && old[key] != nil {
			threshold = 0
		}
		if weight >= threshold {
			entries = append(entries, entry{key, weight})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if initial && entries[i].weight != entries[j].weight {
			return entries[i].weight > entries[j].weight
		}
		if entries[i].key[1] != entries[j].key[1] {
			return entries[i].key[1] < entries[j].key[1]
		}
		return entries[i].key[0] > entries[j].key[0]
	})
	if initial && int32(len(entries)) > maximum {
		entries = entries[:maximum]
		if len(entries) > 0 {
			p.weightLimit = entries[len(entries)-1].weight
		}
	}
	for _, e := range entries {
		Fxu_MatrixAddSingle(p, p.Vars[e.key[0]], p.Vars[e.key[1]], e.weight)
	}
}
func Fxu_MatrixComputeSingles(p *Fxu_Matrix, use0, maximum int32) {
	p.weightLimit = 1 - use0
	rebuildSingles(p, true, maximum)
}
func Fxu_MatrixComputeSinglesOne(p *Fxu_Matrix, v *Fxu_Var) {
	for _, other := range p.Vars {
		if other.Index >= v.Index {
			continue
		}
		weight := Fxu_SingleCountCoincidence(p, other, v) - 2
		if weight >= p.weightLimit {
			Fxu_MatrixAddSingle(p, other, v, weight)
		}
	}
}
func Fxu_MatrixComputeSinglesOneCollect(p *Fxu_Matrix, v *Fxu_Var) []*Fxu_Single {
	var result []*Fxu_Single
	for _, other := range p.Vars {
		if other.Index >= v.Index {
			continue
		}
		weight := Fxu_SingleCountCoincidence(p, other, v) - 2
		if weight >= p.weightLimit {
			result = append(result, &Fxu_Single{Weight: weight, Var1: other, Var2: v, heapIndex: -1})
		}
	}
	return result
}
