package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"math"
)

// Dar_LibEvalBuild chooses and constructs a prepared four-input DAR cut.
func Dar_LibEvalBuild(p *Gia_Man_t, leaves *Vec_Int_t, truth uint32, keepLevel int32, bestLeaves *Vec_Int_t) int32 {
	if vec.Vec_IntSize(leaves) != 4 || truth>>16 != 0 {
		panic("gia: DAR needs a four-input truth table")
	}
	if truth == 0 {
		return 0
	}
	if truth == 0xffff {
		return 1
	}
	match, nodes, roots := dar.Dar_LibPrepare(5).CutImplementations(uint16(truth))
	if len(roots) == 0 {
		panic("gia: DAR class has no implementation")
	}
	values := make([]int32, len(nodes)+4)
	levels := make([]int32, len(values))
	for i := range values {
		values[i] = -1
	}
	for i := 0; i < 4; i++ {
		lit := vec.Vec_IntEntry(leaves, match.Permutation[i])
		if match.InputPhase&(1<<uint(i)) != 0 {
			lit ^= 1
		}
		values[i] = lit
		levels[i] = Gia_ObjLevelId(p, lit>>1)
	}
	for i, node := range nodes {
		a, b := node.Fanins[0], node.Fanins[1]
		levels[i+4] = 1 + max(levels[a>>1], levels[b>>1])
		x, y := values[a>>1], values[b>>1]
		if x < 0 || y < 0 {
			continue
		}
		x ^= int32(a & 1)
		y ^= int32(b & 1)
		switch {
		case x == 0 || y == 0 || x == y^1:
			values[i+4] = 0
		case x == 1 || x == y:
			values[i+4] = y
		case y == 1:
			values[i+4] = x
		default:
			found := Gia_ManHashLookupInt(p, x, y)
			if found != 0 {
				values[i+4] = found
			}
		}
		if values[i+4] >= 0 {
			levels[i+4] = Gia_ObjLevelId(p, values[i+4]>>1)
		}
	}
	best, bestArea, bestLevel := -1, math.MaxInt, int32(math.MaxInt32)
	for i, root := range roots {
		seen := make([]bool, len(values))
		var cost func(int) int
		cost = func(id int) int {
			if seen[id] || id < 4 || values[id] >= 0 {
				return 0
			}
			seen[id] = true
			n := nodes[id-4]
			return 1 + cost(int(n.Fanins[0]>>1)) + cost(int(n.Fanins[1]>>1))
		}
		area, level := cost(root.Index), levels[root.Index]
		if best < 0 || (keepLevel != 0 && (level < bestLevel || level == bestLevel && area < bestArea)) || (keepLevel == 0 && (area < bestArea || area == bestArea && level < bestLevel)) {
			best, bestArea, bestLevel = i, area, level
		}
	}
	if bestLeaves != nil {
		vec.Vec_IntClear(bestLeaves)
		for i := 0; i < 4; i++ {
			vec.Vec_IntPush(bestLeaves, values[i])
		}
	}
	var build func(int) int32
	build = func(id int) int32 {
		if values[id] >= 0 {
			return values[id]
		}
		n := nodes[id-4]
		a := build(int(n.Fanins[0]>>1)) ^ int32(n.Fanins[0]&1)
		b := build(int(n.Fanins[1]>>1)) ^ int32(n.Fanins[1]&1)
		lit := Gia_ManHashAnd(p, a, b)
		values[id] = lit
		obj := Gia_ManObj(p, lit>>1)
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ObjSetAndLevel(p, obj)
		}
		Gia_ObjSetPhase(p, obj)
		return lit
	}
	root := roots[best]
	result := build(root.Index)
	if root.Complement != match.OutputComplement {
		result ^= 1
	}
	return result
}
