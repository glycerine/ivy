package gia

const GIA_LIMIT = 10

// Gia_ManSat_t retains packed partition records for the manager's lifetime.
type Gia_ManSat_t struct{ partitions [][]uint32 }

func Gia_ManSatStart() *Gia_ManSat_t { return &Gia_ManSat_t{} }
func Gia_ManSatStop(p *Gia_ManSat_t) { p.partitions = nil }

// Gia_ManSatPartCreate_rec stores an AND supergate at place and its entries
// beginning at start. Headers contain a fan count and relative byte offset;
// terminals encode their literal in the upper 31 bits and set the low bit.
func Gia_ManSatPartCreate_rec(p *Gia_Man_t, obj *Gia_Obj_t, store []uint32, place, start int) int32 {
	var n int32
	var super [2 * GIA_LIMIT]int32
	Gia_ManSatPartCollectSuper(p, obj, super[:], &n)
	offset := start - place
	if offset < 0 || offset > 255 || n > 127 || start+int(n) > len(store) {
		panic("gia: SAT partition storage overflow")
	}
	store[place] = uint32(n<<1) | uint32(offset)<<8
	used := int(n)
	for i, lit := range super[:n] {
		fan := Gia_ManObj(p, lit>>1)
		if fan.fMark0 != 0 {
			if lit < 0 || uint32(lit) > 0x7fffffff {
				panic("gia: SAT partition literal overflow")
			}
			store[start+i] = uint32(lit)<<1 | 1
		} else {
			if lit&1 == 0 {
				panic("gia: unmarked supergate fanin is not complemented")
			}
			used += int(Gia_ManSatPartCreate_rec(p, fan, store, start+i, start+used))
		}
	}
	return int32(used)
}
func Gia_ManSatPartCreate(p *Gia_Man_t, obj *Gia_Obj_t, store []uint32) int32 {
	if len(store) == 0 {
		panic("gia: empty SAT partition storage")
	}
	return 1 + Gia_ManSatPartCreate_rec(p, obj, store, 0, 1)
}
