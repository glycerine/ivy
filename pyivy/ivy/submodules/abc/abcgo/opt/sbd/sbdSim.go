package sbd

// Port of src/opt/sbd/sbdSim.c: experimental simulation (not used by &mfsd).
// C keeps the simulation words in the manager's vSims/vSimsPi/iPatsPi
// fields; the port passes them in an sbdSim value.

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type sbdSim struct {
	p       *Gia_Man_t
	iPatsPi int32
	vSims   []word
	vSimsPi []word
}

func Sbd_ObjSims(s *sbdSim, i int32) []word { return s.vSims[s.iPatsPi*i : s.iPatsPi*(i+1)] }
func Sbd_ObjCtrl(s *sbdSim, i int32) []word { return s.vSimsPi[s.iPatsPi*i : s.iPatsPi*(i+1)] }

// Sbd_GiaSimRoundBack2: "This does not work." (per the original).
func Sbd_GiaSimRoundBack2(s *sbdSim) {
	p := s.p
	nWords := s.iPatsPi
	// init primary outputs
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Id := Gia_ObjId(p, Gia_ManCo(p, i))
		for w := int32(0); w < nWords; w++ {
			Sbd_ObjSims(s, Id)[w] = Gia_ManRandomW(0)
		}
	}
	// transfer to nodes
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		pObj := Gia_ManCo(p, i)
		pSims := Sbd_ObjSims(s, Gia_ObjId(p, pObj))
		util.Abc_TtCopy(Sbd_ObjSims(s, Gia_ObjFaninId0p(p, pObj)), pSims, nWords, Gia_ObjFaninC0(pObj))
	}
	// simulate nodes
	for i := Gia_ManObjNum(p) - 1; i > 0; i-- {
		pObj := Gia_ManObj(p, i)
		if !Gia_ObjIsAnd(pObj) {
			continue
		}
		pSims := Sbd_ObjSims(s, i)
		pSims0 := Sbd_ObjSims(s, Gia_ObjFaninId0(pObj, i))
		pSims1 := Sbd_ObjSims(s, Gia_ObjFaninId1(pObj, i))
		Rand := Gia_ManRandomW(0)
		for w := int32(0); w < nWords; w++ {
			pSims0[w] = pSims[w] | Rand
			pSims1[w] = pSims[w] | ^Rand
		}
		if Gia_ObjFaninC0(pObj) != 0 {
			util.Abc_TtNot(pSims0, nWords)
		}
		if Gia_ObjFaninC1(pObj) != 0 {
			util.Abc_TtNot(pSims1, nWords)
		}
	}
	// primary inputs are initialized
}

// Sbd_GiaSatOne_rec tries to falsify a sequence of two-literal SAT problems.
func Sbd_GiaSatOne_rec(s *sbdSim, pObj *Gia_Obj_t, Value, fFirst, iPat int32) int32 {
	p := s.p
	id := Gia_ObjId(p, pObj)
	if Gia_ObjIsTravIdCurrentId(p, id) {
		return b2i(b2i(Gia_ObjMark0(pObj)) == Value)
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	Gia_ObjSetMark0(pObj, Value)
	if Gia_ObjIsCi(pObj) {
		pSims := Sbd_ObjSims(s, id)
		if util.Abc_TtGetBit(pSims, iPat) != Value {
			util.Abc_TtXorBit(pSims, iPat)
		}
		return 1
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	assert(!Gia_ObjIsXor(pObj), "!Gia_ObjIsXor(pObj)")
	if Value != 0 {
		return b2i(Sbd_GiaSatOne_rec(s, Gia_ObjFanin0(pObj), 1-Gia_ObjFaninC0(pObj), fFirst, iPat) != 0 &&
			Sbd_GiaSatOne_rec(s, Gia_ObjFanin1(pObj), 1-Gia_ObjFaninC1(pObj), fFirst, iPat) != 0)
	}
	if fFirst != 0 {
		return Sbd_GiaSatOne_rec(s, Gia_ObjFanin0(pObj), Gia_ObjFaninC0(pObj), fFirst, iPat)
	}
	return Sbd_GiaSatOne_rec(s, Gia_ObjFanin1(pObj), Gia_ObjFaninC1(pObj), fFirst, iPat)
}
func Sbd_GiaSatOne(s *sbdSim, vPairs *Vec_Int_t) int32 {
	p := s.p
	iPat := int32(0)
	pairs := vec.Vec_IntArray(vPairs)
	for k := 0; k+1 < len(pairs); k += 2 {
		Var1, Var2 := pairs[k], pairs[k+1]
		pObj1 := Gia_ManObj(p, Var1)
		pObj2 := Gia_ManObj(p, Var2)
		ph1, ph2 := gia.Gia_ObjPhase(pObj1), gia.Gia_ObjPhase(pObj2)
		assert(Var2 > 0, "Var2 > 0")
		var n int32
		if Var1 == 0 {
			for n = 0; n < 2; n++ {
				Gia_ManIncrementTravId(p)
				if Sbd_GiaSatOne_rec(s, pObj2, 1-ph2, n, iPat) != 0 {
					iPat++
					break
				}
			}
			if n == 2 {
				fmt.Printf("%c", '.')
			} else {
				fmt.Printf("%c", 'c')
			}
		} else {
			for n = 0; n < 2; n++ {
				Gia_ManIncrementTravId(p)
				if Sbd_GiaSatOne_rec(s, pObj1, 1-ph1, n, iPat) != 0 && Sbd_GiaSatOne_rec(s, pObj2, ph2, n, iPat) != 0 {
					iPat++
					break
				}
				Gia_ManIncrementTravId(p)
				if Sbd_GiaSatOne_rec(s, pObj1, ph1, n, iPat) != 0 && Sbd_GiaSatOne_rec(s, pObj2, 1-ph2, n, iPat) != 0 {
					iPat++
					break
				}
			}
			if n == 2 {
				fmt.Printf("%c", '.')
			} else {
				fmt.Printf("%c", 'e')
			}
		}
		if iPat == 64*s.iPatsPi-1 {
			break
		}
	}
	fmt.Printf("\n")
	return iPat
}

func Sbd_GiaSimRoundBack(s *sbdSim) {
	p := s.p
	vReprs := vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	vPairs := vec.Vec_IntAlloc(1000)
	nConsts, nClasses := 0, 0
	var vMap *Vec_Int_t // maps each node into its class
	Sbd_GiaSimRound(s, 0, &vMap)
	for i := int32(0); i < Gia_ManObjNum(p); i++ {
		if !Gia_ObjIsAnd(Gia_ManObj(p, i)) {
			continue
		}
		m := int(vec.Vec_IntEntry(vMap, int(i)))
		if m == 0 {
			vec.Vec_IntPush(vPairs, 0)
			vec.Vec_IntPush(vPairs, i)
			nConsts++
		} else if vec.Vec_IntEntry(vReprs, m) == 0 {
			vec.Vec_IntWriteEntry(vReprs, m, i)
		} else if vec.Vec_IntEntry(vReprs, m) != -1 {
			vec.Vec_IntPush(vPairs, vec.Vec_IntEntry(vReprs, m))
			vec.Vec_IntPush(vPairs, i)
			vec.Vec_IntWriteEntry(vReprs, m, -1)
			nClasses++
		}
	}
	fmt.Printf("Constants = %d.   Classes = %d.\n", nConsts, nClasses)
	nPats := Sbd_GiaSatOne(s, vPairs)
	fmt.Printf("Generated %d patterns.\n", nPats)
}

func Sbd_GiaSimRound(s *sbdSim, fTry int32, pvMap **Vec_Int_t) {
	p := s.p
	nWords := s.iPatsPi
	vMap := vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	// init primary inputs
	if fTry != 0 {
		Sbd_GiaSimRoundBack(s)
		for i := int32(0); i < Gia_ManCiNum(p); i++ {
			Sbd_ObjSims(s, Gia_ManCiIdToId(p, i))[0] <<= 1
		}
	} else {
		for i := int32(0); i < Gia_ManCiNum(p); i++ {
			Id := Gia_ManCiIdToId(p, i)
			for w := int32(0); w < nWords; w++ {
				Sbd_ObjSims(s, Id)[w] = Gia_ManRandomW(0) << uint(b2i(w == 0))
			}
		}
	}
	// simulate internal nodes
	vStore := vec.Vec_MemAlloc(int(nWords), 16) // 2^12 N-word entries per page
	vec.Vec_MemHashAlloc(vStore, 1<<16)
	RetValue := vec.Vec_MemHashInsert(vStore, Sbd_ObjSims(s, 0)) // const zero
	assert(RetValue == 0, "RetValue == 0")
	for i := int32(0); i < Gia_ManObjNum(p); i++ {
		pObj := Gia_ManObj(p, i)
		if !Gia_ObjIsAnd(pObj) {
			continue
		}
		pSims := Sbd_ObjSims(s, i)
		if Gia_ObjIsXor(pObj) {
			util.Abc_TtXor(pSims,
				Sbd_ObjSims(s, Gia_ObjFaninId0(pObj, i)),
				Sbd_ObjSims(s, Gia_ObjFaninId1(pObj, i)),
				nWords,
				Gia_ObjFaninC0(pObj)^Gia_ObjFaninC1(pObj))
		} else {
			util.Abc_TtAndCompl(pSims,
				Sbd_ObjSims(s, Gia_ObjFaninId0(pObj, i)), Gia_ObjFaninC0(pObj),
				Sbd_ObjSims(s, Gia_ObjFaninId1(pObj, i)), Gia_ObjFaninC1(pObj),
				nWords)
		}
		// hash sim info
		fCompl := pSims[0]&1 != 0
		if fCompl {
			util.Abc_TtNot(pSims, nWords)
		}
		vec.Vec_IntWriteEntry(vMap, int(i), int32(vec.Vec_MemHashInsert(vStore, pSims)))
		if fCompl {
			util.Abc_TtNot(pSims, nWords)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		pObj := Gia_ManCo(p, i)
		pSims := Sbd_ObjSims(s, Gia_ObjId(p, pObj))
		util.Abc_TtCopy(pSims, Sbd_ObjSims(s, Gia_ObjFaninId0p(p, pObj)), nWords, Gia_ObjFaninC0(pObj))
		assert(gia.Gia_ObjPhase(pObj) == int32(pSims[0]&1), "Gia_ObjPhase(pObj) == (int)(pSims[0] & 1)")
	}
	fmt.Printf("Objects = %6d.  Unique = %6d.\n", gia_ManAndNum(p), vec.Vec_IntCountUnique(vMap))
	if pvMap != nil {
		*pvMap = vMap
	}
}
func Sbd_GiaSimTest(pGia *Gia_Man_t) {
	gia.Gia_ManSetPhase(pGia)
	// allocate simulation info
	s := &sbdSim{p: pGia, iPatsPi: 32}
	s.vSims = make([]word, Gia_ManObjNum(pGia)*s.iPatsPi)
	s.vSimsPi = make([]word, Gia_ManObjNum(pGia)*s.iPatsPi)
	gia.Gia_ManRandom(1)
	Sbd_GiaSimRound(s, 0, nil)
	Sbd_GiaSimRound(s, 0, nil)
	Sbd_GiaSimRound(s, 0, nil)
	fmt.Printf("\n")
	Sbd_GiaSimRound(s, 1, nil)
	fmt.Printf("\n")
	Sbd_GiaSimRound(s, 1, nil)
	fmt.Printf("\n")
	Sbd_GiaSimRound(s, 1, nil)
}
