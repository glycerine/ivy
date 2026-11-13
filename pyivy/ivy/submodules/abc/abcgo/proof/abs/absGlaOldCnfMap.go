package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

// glaOldCnfForGia expresses mapped AIG clauses and LUT boundaries in GIA
// object IDs. The Go AIG exporter may reorder CIs relative to logic nodes.
func glaOldCnfForGia(g *gia.Gia_Man_t) (*cnf.Cnf_Dat_t, error) {
	// The original old-GLA AIG conversion treats every internal node as an
	// AND. Real XOR/MUX nodes require prior lowering to ordinary ANDs.
	for id := int32(1); id < gia.Gia_ManObjNum(g); id++ {
		obj := gia.Gia_ManObj(g, id)
		if gia.Gia_ObjIsXor(obj) != 0 || gia.Gia_ObjIsMux(g, obj) != 0 {
			return nil, fmt.Errorf("old GLA: lower XOR/MUX object %d to ANDs first", id)
		}
	}
	aigGraph := gia.Gia_ManToAigSimple(g)
	data, err := cnf.Cnf_DeriveOtherMapped(aigGraph, 1)
	if err != nil {
		return nil, err
	}
	n := int(gia.Gia_ManObjNum(g))
	reverse := make([]int32, len(aigGraph.Nodes()))
	inverted := make([]int32, len(reverse))
	for i := range reverse {
		reverse[i] = -1
	}
	for id := 0; id < n; id++ {
		lit := gia.Gia_ManObj(g, int32(id)).Value
		if lit == ^uint32(0) || int(lit>>1) >= len(reverse) {
			return nil, fmt.Errorf("old GLA: object %d has no AIG mapping", id)
		}
		aid := int(lit >> 1)
		if reverse[aid] >= 0 {
			// ABC's original simple conversion asserts this cannot occur. Use
			// immediate-fanin cuts to preserve each GIA node independently.
			return glaOldLocalCnf(g)
		}
		reverse[aid] = int32(id)
		inverted[aid] = int32(lit & 1)
	}
	mapID := func(id int32) (int32, error) {
		if id < 0 || int(id) >= len(reverse) || reverse[id] < 0 {
			return 0, fmt.Errorf("old GLA: unmapped AIG object %d", id)
		}
		return reverse[id], nil
	}
	counts, offsets := make([]int32, n), make([]int32, n)
	for i := range counts {
		counts[i], offsets[i] = -1, -1
	}
	mapping := vec.Vec_IntStart(n)
	for aid, gid := range reverse {
		if gid < 0 {
			continue
		}
		counts[gid], offsets[gid] = data.Obj2Count[aid], data.Obj2Clause[aid]
		if aid >= len(data.Mapping) || data.Mapping[aid] == 0 {
			continue
		}
		off := int(data.Mapping[aid])
		if off+4 >= len(data.Mapping) {
			return nil, fmt.Errorf("old GLA: malformed LUT mapping")
		}
		leaves := make([]int32, 0, 4)
		leafInverted := make([]int32, 0, 4)
		for k := 1; k <= 4; k++ {
			leaf := data.Mapping[off+k]
			if leaf < 0 {
				continue
			}
			mapped, e := mapID(leaf)
			if e != nil {
				return nil, e
			}
			leaves = append(leaves, mapped)
			leafInverted = append(leafInverted, inverted[leaf])
		}
		oldTruth := uint32(data.Mapping[off])
		var truth uint32
		for assignment := 0; assignment < 16; assignment++ {
			oldAssignment := 0
			for i := range leaves {
				bit := int32((assignment>>i)&1) ^ leafInverted[i]
				oldAssignment |= int(bit) << i
			}
			value := int32(oldTruth>>oldAssignment)&1 ^ inverted[aid]
			truth |= uint32(value) << assignment
		}
		vec.Vec_IntWriteEntry(mapping, int(gid), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, int32(truth))
		for _, leaf := range leaves {
			vec.Vec_IntPush(mapping, leaf)
		}
		for len(leaves) < 4 {
			vec.Vec_IntPush(mapping, -1)
			leaves = append(leaves, -1)
		}
	}
	clauses := make([][]int32, len(data.Clauses))
	for i, clause := range data.Clauses {
		clauses[i] = make([]int32, len(clause))
		for j, lit := range clause {
			aid := lit >> 1
			gid, e := mapID(aid)
			if e != nil {
				return nil, e
			}
			clauses[i][j] = 2*gid + (lit & 1) ^ inverted[aid]
		}
	}
	result := cnf.NewMappedCNF(n, clauses, nil, offsets, counts)
	result.Mapping = append([]int32(nil), vec.Vec_IntArray(mapping)...)
	return result, nil
}

// glaOldLocalCnf preserves object identity when AIG hashing would merge or
// simplify distinct GIA ANDs. Each AND uses its valid immediate-fanin cut.
func glaOldLocalCnf(g *gia.Gia_Man_t) (*cnf.Cnf_Dat_t, error) {
	n := int(gia.Gia_ManObjNum(g))
	mapping := vec.Vec_IntStart(n)
	counts, offsets := make([]int32, n), make([]int32, n)
	for i := range counts {
		counts[i], offsets[i] = -1, -1
	}
	clauses := make([][]int32, 0, 3*n)
	for id := int32(0); id < int32(n); id++ {
		obj := gia.Gia_ManObj(g, id)
		switch {
		case gia.Gia_ObjIsConst0(obj) != 0:
			offsets[id] = int32(len(clauses))
			counts[id] = 1
			clauses = append(clauses, []int32{2*id + 1})
		case gia.Gia_ObjIsCi(obj) != 0:
			counts[id] = 0
		case gia.Gia_ObjIsAnd(obj) != 0:
			if gia.Gia_ObjIsXor(obj) != 0 || gia.Gia_ObjIsMux(g, obj) != 0 {
				return nil, fmt.Errorf("old GLA: local cut fallback does not support XOR or MUX object %d", id)
			}
			a, b := gia.Gia_ObjId(g, gia.Gia_ObjFanin0(obj)), gia.Gia_ObjId(g, gia.Gia_ObjFanin1(obj))
			ca, cb := gia.Gia_ObjFaninC0(obj), gia.Gia_ObjFaninC1(obj)
			vec.Vec_IntWriteEntry(mapping, int(id), int32(vec.Vec_IntSize(mapping)))
			var truth uint32
			for assignment := 0; assignment < 16; assignment++ {
				av := int32(assignment&1) ^ ca
				bv := int32((assignment>>1)&1) ^ cb
				truth |= uint32(av&bv) << assignment
			}
			vec.Vec_IntPush(mapping, int32(truth))
			vec.Vec_IntPush(mapping, a)
			vec.Vec_IntPush(mapping, b)
			vec.Vec_IntPush(mapping, -1)
			vec.Vec_IntPush(mapping, -1)
			offsets[id] = int32(len(clauses))
			counts[id] = 3
			clauses = append(clauses, []int32{2*id + 1, 2*a + ca}, []int32{2*id + 1, 2*b + cb}, []int32{2 * id, 2*a + (ca ^ 1), 2*b + (cb ^ 1)})
		case gia.Gia_ObjIsCo(obj) != 0:
			a := gia.Gia_ObjId(g, gia.Gia_ObjFanin0(obj))
			c := gia.Gia_ObjFaninC0(obj)
			offsets[id] = int32(len(clauses))
			counts[id] = 2
			clauses = append(clauses, []int32{2 * id, 2*a + (c ^ 1)}, []int32{2*id + 1, 2*a + c})
		default:
			return nil, fmt.Errorf("old GLA: unsupported GIA object %d", id)
		}
	}
	data := cnf.NewMappedCNF(n, clauses, nil, offsets, counts)
	data.Mapping = append([]int32(nil), vec.Vec_IntArray(mapping)...)
	return data, nil
}
