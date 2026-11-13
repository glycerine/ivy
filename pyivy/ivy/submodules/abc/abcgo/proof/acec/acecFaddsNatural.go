package acec

// Natural carry-chain timing boxes from src/proof/acec/acecFadds.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Gia_ManIllustrateBoxes(p *Gia_Man_t) {
	tm, _ := p.BoxTiming()
	if tm == nil {
		panic("acec: no box timing")
	}
	ci, co := tim.Tim_ManPiNum(tm), 0
	for i := 0; i < tim.Tim_ManBoxNum(tm); i++ {
		ni, no := tim.Tim_ManBoxInputNum(tm, i), tim.Tim_ManBoxOutputNum(tm, i)
		fmt.Printf("Box %4d  [%d x %d] :   Input obj IDs = ", i, ni, no)
		for k := 0; k < ni; k++ {
			fmt.Printf("%d ", gia.Gia_ObjFaninId0p(p, gia.Gia_ManCo(p, int32(co+k))))
		}
		fmt.Print("   Output obj IDs = ")
		for k := 0; k < no; k++ {
			fmt.Printf("%d ", gia.Gia_ObjId(p, gia.Gia_ManCi(p, int32(ci+k))))
		}
		fmt.Println()
		co += ni
		ci += no
	}
}
func Gia_ManDupFadd(dst, p *Gia_Man_t, chain, adds, m *Vec_Int_t, chains *Vec_Wec_t, mapChain, truths *Vec_Int_t) {
	for i, box := range vec.Vec_IntArray(chain) {
		for k := 0; k < 3; k++ {
			if i != 0 && k == 0 {
				continue
			}
			o := gia.Gia_ManObj(p, vec.Vec_IntEntry(adds, int(5*box)+k))
			Gia_ManDupWithFaddBoxes_rec(dst, p, o, adds, m, chains, mapChain, truths)
		}
	}
	carry := int32(0)
	last := int32(-1)
	for _, box := range vec.Vec_IntArray(chain) {
		last = box
		sumTruth, carryTruth := vec.Vec_IntEntry(truths, int(2*box)), vec.Vec_IntEntry(truths, int(2*box+1))
		var lits [3]int32
		for k := 0; k < 3; k++ {
			o := gia.Gia_ManObj(p, vec.Vec_IntEntry(adds, int(5*box)+k))
			lits[k] = int32(o.Value)
			if k == 0 && carry != 0 {
				lits[k] = carry
			}
			if lits[k] < 0 {
				panic("acec: missing adder input")
			}
		}
		switch carryTruth {
		case 0x4d:
			lits[0] ^= 1
			carryTruth = 0x8e
			sumTruth ^= 0xff
		case 0xd4:
			lits[0] ^= 1
			carryTruth = 0xe8
			sumTruth ^= 0xff
		case 0x2b:
			lits[1] ^= 1
			carryTruth = 0x8e
			sumTruth ^= 0xff
		case 0xb2:
			lits[1] ^= 1
			carryTruth = 0xe8
			sumTruth ^= 0xff
		}
		switch carryTruth {
		case 0x8e:
			lits[2] ^= 1
			carryTruth = 0xe8
			sumTruth ^= 0xff
		case 0x71:
			lits[2] ^= 1
			carryTruth = 0x17
			sumTruth ^= 0xff
		case 0xe8, 0x17:
		default:
			panic("acec: unexpected carry truth")
		}
		if lits[0]&1 != 0 {
			for k := range lits {
				lits[k] ^= 1
			}
			sumTruth ^= 0xff
			carryTruth ^= 0xff
		}
		for _, lit := range lits {
			gia.Gia_ManAppendCo(dst, lit)
		}
		if sumTruth != 0x96 && sumTruth != 0x69 {
			panic("acec: unexpected sum truth")
		}
		sum := gia.Gia_ManAppendCi(dst)
		if sumTruth == 0x69 {
			sum ^= 1
		}
		gia.Gia_ManObj(p, vec.Vec_IntEntry(adds, int(5*box+3))).Value = uint32(sum)
		if carryTruth != 0xe8 && carryTruth != 0x17 {
			panic("acec: unexpected carry phase")
		}
		carry = gia.Gia_ManAppendCi(dst)
		if carryTruth == 0x17 {
			carry ^= 1
		}
	}
	if last < 0 {
		panic("acec: empty adder chain")
	}
	gia.Gia_ManObj(p, vec.Vec_IntEntry(adds, int(5*last+4))).Value = uint32(carry)
}
func Gia_ManDupWithFaddBoxes_rec(dst, p *Gia_Man_t, o *Gia_Obj_t, adds, m *Vec_Int_t, chains *Vec_Wec_t, mapChain, truths *Vec_Int_t) {
	if o.Value != ^uint32(0) {
		return
	}
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("acec: expected AND while duplicating")
	}
	Gia_ManDupWithFaddBoxes_rec(dst, p, gia.Gia_ObjFanin0(o), adds, m, chains, mapChain, truths)
	Gia_ManDupWithFaddBoxes_rec(dst, p, gia.Gia_ObjFanin1(o), adds, m, chains, mapChain, truths)
	if gia.Gia_ObjIsMux(p, o) != 0 {
		Gia_ManDupWithFaddBoxes_rec(dst, p, gia.Gia_ObjFanin2(p, o), adds, m, chains, mapChain, truths)
	}
	o.Value = uint32(acecCopyLogic(dst, p, o))
}
func Gia_ManDupWithNaturalBoxes(p *Gia_Man_t, min, verbose int32) *Gia_Man_t {
	start := time.Now()
	if gia.Gia_ManBoxNum(p) > 0 {
		fmt.Println("Currently natural carry-chains cannot be detected when boxes are present.")
		return nil
	}
	adds := Gia_ManDetectFullAdders(p, verbose, nil)
	m := Gia_ManCreateMap(p, adds)
	Gia_ManFindChains(p, adds, m)
	chains := Gia_ManCollectTopmost(p, adds, m, min)
	if verbose != 0 {
		Gia_ManPrintChains(p, adds, m, chains)
	}
	if vec.Vec_WecSize(chains) == 0 {
		return gia.Gia_ManDup(p)
	}
	mapChain := Gia_ManFindMapping(p, adds, m, chains)
	truths := Gia_ManCollectTruthTables(p, adds)
	if verbose != 0 {
		fmt.Printf("Carry-chain detection time = %v\n", time.Since(start))
	}
	start = time.Now()
	gia.Gia_ManFillValue(p)
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	dst.CopyIdentityFrom(p)
	gia.Gia_ManConst0(p).Value = 0
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ManCi(p, i).Value = uint32(gia.Gia_ManAppendCi(dst))
	}
	for i := 0; i < vec.Vec_WecSize(chains); i++ {
		Gia_ManDupFadd(dst, p, vec.Vec_WecEntry(chains, i), adds, m, chains, mapChain, truths)
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		Gia_ManDupWithFaddBoxes_rec(dst, p, gia.Gia_ObjFanin0(gia.Gia_ManCo(p, i)), adds, m, chains, mapChain, truths)
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(gia.Gia_ManCo(p, i)))
	}
	if gia.Gia_ManRegNum(p) != 0 && verbose != 0 {
		fmt.Println("Warning: Sequential design is converted into combinational one by adding white boxes.")
	}
	if gia.Gia_ManHasDangling(dst) != 0 {
		panic("acec: dangling node after box duplication")
	}
	n := (gia.Gia_ManCiNum(dst) - gia.Gia_ManCiNum(p)) / 2
	if n != (gia.Gia_ManCoNum(dst)-gia.Gia_ManCoNum(p))/3 {
		panic("acec: mismatched timing-box count")
	}
	dst.SetBoxTiming(Gia_ManGenerateTim(gia.Gia_ManCiNum(p), gia.Gia_ManCoNum(p), n, 3, 2), Gia_ManGenerateExtraAig(n, 3, 2))
	if verbose != 0 {
		fmt.Printf("AIG with boxes construction time = %v\n", time.Since(start))
	}
	return dst
}
