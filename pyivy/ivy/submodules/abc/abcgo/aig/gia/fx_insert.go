package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dau"
	"strings"
)

// Gia_ManFxInsert rebuilds a network from extracted cube covers. Small
// covers use DSD decomposition; larger covers are factored as SOPs.
func Gia_ManFxInsert(p *Gia_Man_t, cubes *Vec_Wec_t, phases *Vec_Str_t) *Gia_Man_t {
	var first, count *Vec_Int_t
	order := Gia_ManFxTopoOrder(cubes, Gia_ManCiNum(p), int32(Vec_StrSize(phases)), &first, &count)
	if order == nil {
		return Gia_ManDup(p)
	}
	result := Gia_ManStart(p.nObjs)
	result.pName, result.PSpec = p.pName, p.PSpec
	Gia_ManHashStart(result)
	copies := make([]int32, Vec_IntSize(order))
	for i := range copies {
		copies[i] = -1
	}
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		copies[i] = Gia_ManAppendCi(result)
	}
	for k := int(Gia_ManCiNum(p)); k < Vec_IntSize(order); k++ {
		owner := Vec_IntEntry(order, k)
		start, n := Vec_IntEntry(first, int(owner)), Vec_IntEntry(count, int(owner))
		index := make(map[int32]int)
		var fanins []int32
		for c := int32(0); c < n; c++ {
			cube := Vec_WecEntry(cubes, int(start+c))
			for j := 1; j < Vec_IntSize(cube); j++ {
				input := Vec_IntEntry(cube, j) >> 1
				if _, ok := index[input]; !ok {
					index[input] = len(fanins)
					fanins = append(fanins, input)
				}
			}
		}
		// As in the original, an absent cover represents the same constant
		// as an empty cube; the phase flag distinguishes zero from one.
		lit := int32(1)
		if len(fanins) != 0 {
			leaves := Vec_IntAlloc(len(fanins))
			for _, input := range fanins {
				if copies[input] < 0 {
					panic("gia: extraction cube has unresolved input")
				}
				Vec_IntPush(leaves, copies[input])
			}
			var sop strings.Builder
			for c := int32(0); c < n; c++ {
				row := []byte(strings.Repeat("-", len(fanins)))
				cube := Vec_WecEntry(cubes, int(start+c))
				for j := 1; j < Vec_IntSize(cube); j++ {
					input := Vec_IntEntry(cube, j)
					char := byte('1')
					if input&1 != 0 {
						char = '0'
					}
					row[index[input>>1]] = char
				}
				sop.Write(row)
				sop.WriteString(" 1\n")
			}
			if len(fanins) <= 6 {
				var truth uint64
				for c := int32(0); c < n; c++ {
					term := ^uint64(0)
					cube := Vec_WecEntry(cubes, int(start+c))
					for j := 1; j < Vec_IntSize(cube); j++ {
						input := Vec_IntEntry(cube, j)
						v := index[input>>1]
						var pattern uint64
						for bit := 0; bit < 64; bit++ {
							if bit&(1<<v) != 0 {
								pattern |= uint64(1) << bit
							}
						}
						if input&1 != 0 {
							pattern = ^pattern
						}
						term &= pattern
					}
					truth |= term
				}
				if truth == 0 {
					truth = ^truth
				}
				lit = giaFxDsdToGia(result, truth, Vec_IntArray(leaves))
			} else {
				lit = Gia_ManFactorNode(result, sop.String(), leaves)
			}
		}
		if int(owner) < Vec_StrSize(phases) && Vec_StrEntry(phases, int(owner)) != 0 {
			lit ^= 1
		}
		copies[owner] = lit
	}
	Gia_ManHashStop(result)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		driver := Gia_ObjFanin0(obj)
		lit := int32(0)
		if driver.id != 0 {
			lit = copies[driver.Value]
			if lit < 0 {
				panic("gia: extraction output has unresolved driver")
			}
		}
		Gia_ManAppendCo(result, lit^int32(obj.fCompl0))
	}
	Gia_ManSetRegNum(result, Gia_ManRegNum(p))
	return Gia_ManCleanup(result)
}

// giaFxDsdToGia imports the decomposer's AND graph into the destination's
// structural hash. The temporary graph uses the opposite constant polarity.
func giaFxDsdToGia(p *Gia_Man_t, truth uint64, leaves []int32) int32 {
	graph := aig.NewGraph()
	inputs := make([]aig.Edge, len(leaves))
	levels := make([]int32, p.nObjs)
	for id := int32(1); id < p.nObjs; id++ {
		obj := p.pObjs[id]
		if Gia_ObjIsAnd(obj) != 0 {
			levels[id] = 1 + max(levels[Gia_ObjFaninId0(obj, id)], levels[Gia_ObjFaninId1(obj, id)])
		}
	}
	for i, lit := range leaves {
		inputs[i] = graph.NewInput()
		aig.Aig_ObjSetLevel(graph, inputs[i], levels[lit>>1])
	}
	root := dau.Dsm_ManTruthToGia(graph, []uint64{truth}, inputs, nil)
	nodes := graph.Nodes()
	values := make([]int32, len(nodes))
	values[0] = 1
	for i, input := range inputs {
		values[input.NodeID()] = leaves[i]
	}
	edge := func(e aig.Edge) int32 { return values[e.NodeID()] ^ int32(uint32(e)&1) }
	for id, node := range nodes {
		if node.Kind == aig.And {
			values[id] = Gia_ManHashAnd(p, edge(node.Fanins[0]), edge(node.Fanins[1]))
		}
	}
	return edge(root)
}
