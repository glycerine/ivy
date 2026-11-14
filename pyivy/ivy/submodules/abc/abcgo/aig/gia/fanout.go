package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_FanoutCreate(id, num int32) int32 {
	if id < 0 || num < 0 || num > 1 {
		panic("gia: invalid fanout edge")
	}
	return id<<1 | num
}
func Gia_FanoutObj(data []int32, id int32) *int32      { return &data[5*id] }
func Gia_FanoutPrev(data []int32, edge int32) *int32   { return &data[5*(edge>>1)+1+(edge&1)] }
func Gia_FanoutNext(data []int32, edge int32) *int32   { return &data[5*(edge>>1)+3+(edge&1)] }
func Gia_ObjFanout0Int(p *Gia_Man_t, id int32) int32   { return *Gia_FanoutObj(p.pFanData, id) }
func Gia_ObjFanoutNext(p *Gia_Man_t, edge int32) int32 { return *Gia_FanoutNext(p.pFanData, edge) }
func Gia_ManFanoutStart(p *Gia_Man_t) {
	if p.pFanData != nil {
		panic("gia: fanouts already initialized")
	}
	p.nFansAlloc = 2 * p.nObjs
	if p.nFansAlloc < 1<<12 {
		p.nFansAlloc = 1 << 12
	}
	p.pFanData = make([]int32, 5*int(p.nFansAlloc))
	for _, obj := range p.pObjs {
		if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsAnd(obj) != 0 {
			Gia_ObjAddFanout(p, Gia_ObjFanin0(obj), obj)
		}
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsBuf(obj) == 0 {
			Gia_ObjAddFanout(p, Gia_ObjFanin1(obj), obj)
		}
	}
}
func Gia_ManFanoutStop(p *Gia_Man_t) { p.pFanData = nil; p.nFansAlloc = 0 }
func Gia_ObjRemoveFanout(p *Gia_Man_t, obj, fanout *Gia_Obj_t) {
	source, target := Gia_ObjId(p, obj), Gia_ObjId(p, fanout)
	edge := target << 1
	if Gia_ObjFanin0(fanout) != obj {
		if Gia_ObjIsAnd(fanout) == 0 || Gia_ObjFanin1(fanout) != obj {
			panic("gia: object is not a fanin")
		}
		edge |= 1
	}
	prev, next := *Gia_FanoutPrev(p.pFanData, edge), *Gia_FanoutNext(p.pFanData, edge)
	if *Gia_FanoutNext(p.pFanData, prev) != edge || *Gia_FanoutPrev(p.pFanData, next) != edge {
		panic("gia: corrupt fanout ring")
	}
	first := Gia_FanoutObj(p.pFanData, source)
	if *first == edge {
		if next == edge {
			*first = 0
			*Gia_FanoutPrev(p.pFanData, edge) = 0
			*Gia_FanoutNext(p.pFanData, edge) = 0
			return
		}
		*first = next
	}
	*Gia_FanoutNext(p.pFanData, prev) = next
	*Gia_FanoutPrev(p.pFanData, next) = prev
	*Gia_FanoutPrev(p.pFanData, edge) = 0
	*Gia_FanoutNext(p.pFanData, edge) = 0
}
func giaObjectFanins(p *Gia_Man_t, obj *Gia_Obj_t) []int32 {
	if Gia_ObjIsCo(obj) != 0 {
		return []int32{Gia_ObjFaninId0(obj, obj.id)}
	}
	if Gia_ObjIsAnd(obj) == 0 {
		return nil
	}
	result := []int32{Gia_ObjFaninId0(obj, obj.id)}
	if Gia_ObjIsBuf(obj) == 0 {
		result = append(result, Gia_ObjFaninId1(obj, obj.id))
	}
	if int(obj.id) < len(p.PMuxes) && p.PMuxes[obj.id] != 0 {
		result = append(result, int32(p.PMuxes[obj.id]>>1))
	}
	return result
}
func Gia_ManStartFanoutMap(p *Gia_Man_t, counts *Vec_Int_t) *Vec_Int_t {
	total := int(p.nObjs)
	for i := int32(0); i < p.nObjs; i++ {
		count := vec.Vec_IntEntry(counts, int(i))
		if count < 0 {
			panic("gia: negative fanout count")
		}
		total += int(count)
	}
	out := vec.Vec_IntStart(total)
	offset := p.nObjs
	for i := int32(0); i < p.nObjs; i++ {
		vec.Vec_IntWriteEntry(out, int(i), offset)
		offset += vec.Vec_IntEntry(counts, int(i))
	}
	return out
}
func Gia_ManStaticFanoutStart(p *Gia_Man_t) {
	if p.vFanoutNums != nil || p.vFanout != nil {
		panic("gia: static fanouts already initialized")
	}
	counts := vec.Vec_IntStart(int(p.nObjs))
	for _, obj := range p.pObjs {
		for _, fanin := range giaObjectFanins(p, obj) {
			vec.Vec_IntAddToEntry(counts, int(fanin), 1)
		}
	}
	p.vFanoutNums = counts
	p.vFanout = Gia_ManStartFanoutMap(p, counts)
	used := make([]int32, p.nObjs)
	for _, obj := range p.pObjs {
		for _, fanin := range giaObjectFanins(p, obj) {
			Gia_ObjSetFanoutInt(p, p.pObjs[fanin], used[fanin], obj.id)
			used[fanin]++
		}
	}
}
func Gia_ManStaticFanoutStop(p *Gia_Man_t) { p.vFanout = nil; p.vFanoutNums = nil }
func Gia_ObjFanoutNumId(p *Gia_Man_t, id int32) int32 {
	return vec.Vec_IntEntry(p.vFanoutNums, int(id))
}
func Gia_ObjFanoutNum(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjFanoutNumId(p, obj.id) }
func Gia_ObjFanoutId(p *Gia_Man_t, id, index int32) int32 {
	offset := vec.Vec_IntEntry(p.vFanout, int(id))
	return vec.Vec_IntEntry(p.vFanout, int(offset+index))
}
func Gia_ObjFanout(p *Gia_Man_t, obj *Gia_Obj_t, index int32) *Gia_Obj_t {
	return Gia_ManObj(p, Gia_ObjFanoutId(p, obj.id, index))
}
func Gia_ObjFanout0(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t { return Gia_ObjFanout(p, obj, 0) }
func Gia_ObjSetFanoutInt(p *Gia_Man_t, obj *Gia_Obj_t, index, value int32) {
	offset := vec.Vec_IntEntry(p.vFanout, int(obj.id))
	vec.Vec_IntWriteEntry(p.vFanout, int(offset+index), value)
}
func Gia_ObjSetFanout(p *Gia_Man_t, obj *Gia_Obj_t, index int32, fanout *Gia_Obj_t) {
	Gia_ObjSetFanoutInt(p, obj, index, fanout.id)
}
