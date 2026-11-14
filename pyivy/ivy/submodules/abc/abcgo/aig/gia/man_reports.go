package gia

import (
	"encoding/json"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
)

func Gia_ManPrintClasses_old(p *Gia_Man_t) {
	if p.VFlopClasses == nil {
		return
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		fmt.Printf("%d", vec.Vec_IntEntry(p.VFlopClasses, int(i)))
	}
	fmt.Println()
	for class := int32(1); class <= 2; class++ {
		graph := Gia_ManDupFlopClass(p, class)
		if err := Gia_AigerWrite(graph, fmt.Sprintf("dom%d.aig", class), 0, 0, 0); err != nil {
			fmt.Println(err)
		}
		Gia_ManStop(graph)
	}
}
func Gia_ManPrintPlacement(p *Gia_Man_t) {
	if p.pPlacement == nil {
		return
	}
	var fixed, undef uint32
	for _, place := range p.pPlacement[:p.nObjs] {
		fixed += uint32(place.fFixed)
		undef += uint32(place.fUndef)
	}
	fmt.Printf("Placement:  Objects = %8d.  Fixed = %8d.  Undef = %8d.\n", p.nObjs, fixed, undef)
}
func Gia_ManPrintTents_rec(p *Gia_Man_t, obj *Gia_Obj_t, objects *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	vec.Vec_IntPush(objects, obj.id)
	if Gia_ObjIsCi(obj) != 0 {
		return
	}
	Gia_ManPrintTents_rec(p, Gia_ObjFanin0(obj), objects)
	if Gia_ObjIsAnd(obj) != 0 {
		Gia_ManPrintTents_rec(p, Gia_ObjFanin1(obj), objects)
	}
}
func Gia_ManPrintTents(p *Gia_Man_t) {
	if Gia_ManPoNum(p) <= 0 {
		panic("gia: no primary outputs")
	}
	objects := vec.Vec_IntAlloc(100)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	vec.Vec_IntPush(objects, 0)
	previous := vec.Vec_IntSize(objects)
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		Gia_ManPrintTents_rec(p, Gia_ManPo(p, i), objects)
	}
	fmt.Print("Tents:  ")
	for t := 1; previous < vec.Vec_IntSize(objects); t++ {
		current := vec.Vec_IntSize(objects)
		pis := int32(0)
		for i := previous; i < current; i++ {
			obj := Gia_ManObj(p, vec.Vec_IntEntry(objects, i))
			pis += Gia_ObjIsPi(p, obj)
			if Gia_ObjIsRo(p, obj) != 0 {
				Gia_ManPrintTents_rec(p, Gia_ObjRoToRi(p, obj), objects)
			}
		}
		fmt.Printf("%d=%d(%d)  ", t, current-previous, pis)
		previous = current
	}
	fmt.Printf(" Unused=%d\n", int(p.nObjs)-vec.Vec_IntSize(objects))
}
func Gia_ManPrintInitClasses(initial *Vec_Int_t) {
	var counts [6]int
	for _, v := range vec.Vec_IntArray(initial) {
		if v < 0 || v >= 6 {
			panic("gia: invalid initialization class")
		}
		counts[v]++
	}
	for i, n := range counts {
		if n != 0 {
			fmt.Printf("%d = %d  ", i, n)
		}
	}
	fmt.Printf("  B = %d  X = %d  Q = %d\n", counts[0]+counts[1], counts[2]+counts[3], counts[4]+counts[5])
	for _, v := range vec.Vec_IntArray(initial) {
		fmt.Print(v)
	}
	fmt.Println()
}
func Gia_ManPrintChoiceStats(p *Gia_Man_t) {
	Gia_ManMarkFanoutDrivers(p)
	var equivs, choices int32
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjSibl(p, obj.id) == 0 {
			continue
		}
		equivs++
		if obj.fMark0 != 0 {
			choices++
		}
		sibling := Gia_ObjSiblObj(p, obj.id)
		if sibling.fMark0 != 0 || Gia_ObjIsAnd(sibling) == 0 {
			panic("gia: invalid choice")
		}
	}
	fmt.Printf("Choice stats: Equivs =%7d. Choices =%7d.\n", equivs, choices)
	Gia_ManCleanMark0(p)
}
func Gia_ManPrintEdges(p *Gia_Man_t) int32 {
	fmt.Printf("Edges (Q=2)    :                edge =%8d  lev =%5.1f\n", (vec.Vec_IntCountPositive(p.VEdge1)+vec.Vec_IntCountPositive(p.VEdge2))/2, 0.1*float64(Gia_ManEvalEdgeDelay(p)))
	return 0
}
func Gia_ManLogAigStats(p *Gia_Man_t, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	data := struct {
		Name   string `json:"name"`
		Input  int32  `json:"input"`
		Output int32  `json:"output"`
		And    int32  `json:"and"`
		Level  int32  `json:"level"`
	}{p.pName, Gia_ManCiNum(p), Gia_ManCoNum(p), Gia_ManAndNum(p), Gia_ManLevelNum(p)}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")
	if err := enc.Encode(data); err != nil {
		fmt.Println(err)
	}
}
func Gia_ManPrintStatsShort(p *Gia_Man_t) {
	fmt.Printf("i/o =%7d/%7d  ff =%7d  and =%8d  lev =%5d  \n", Gia_ManPiNum(p), Gia_ManPoNum(p), Gia_ManRegNum(p), Gia_ManAndNum(p), Gia_ManLevelNum(p))
}
func Gia_ManPrintMiterStatus(p *Gia_Man_t) {
	var sat, unsat, undec int32
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		obj := Gia_ManPo(p, i)
		child := Gia_ObjChild0(obj)
		if child == Gia_ManConst0(p) {
			unsat++
		} else if child == Gia_ManConst1(p) || Gia_ObjIsPi(p, Gia_Regular(child)) != 0 {
			sat++
		} else {
			undec++
		}
	}
	fmt.Printf("Outputs = %7d.  Unsat = %7d.  Sat = %7d.  Undec = %7d.\n", Gia_ManPoNum(p), unsat, sat, undec)
}
func Gia_ManReportImprovement(p, q *Gia_Man_t) {
	pct := func(a, b int32) float64 {
		if a == 0 {
			return 0
		}
		return 100 * float64(a-b) / float64(a)
	}
	fmt.Printf("REG: Beg = %5d. End = %5d. (R =%5.1f %%)  AND: Beg = %6d. End = %6d. (R =%5.1f %%)\n", Gia_ManRegNum(p), Gia_ManRegNum(q), pct(Gia_ManRegNum(p), Gia_ManRegNum(q)), Gia_ManAndNum(p), Gia_ManAndNum(q), pct(Gia_ManAndNum(p), Gia_ManAndNum(q)))
}

func Gia_ManPrintStatsMiter(p *Gia_Man_t, verbose int32) {
	Gia_ManLevelNum(p)
	Gia_ManCreateRefs(p)
	if LfSwitchingProvider == nil {
		panic("gia: switching simulator is not connected")
	}
	copy := Gia_ManDup(p)
	frames := int32(1)
	if Gia_ManRegNum(p) != 0 {
		frames = 16
	}
	probability := LfSwitchingProvider(copy, frames, 0, 1)
	Gia_ManStop(copy)
	if probability == nil || vec.Vec_FltSize(probability) < int(p.nObjs) {
		panic("gia: invalid output probabilities")
	}
	fmt.Println("Statistics for each outputs of the miter:")
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		obj := Gia_ManPo(p, i)
		id := obj.id
		roots := []int32{id}
		fmt.Printf("%4d : Level = %5d  Supp = %5d  Cone = %5d  Mffc = %5d  Prob = %8.4f  \n", i, Gia_ObjLevelId(p, id), Gia_ManSuppSize(p, roots, 1), Gia_ManConeSize(p, roots, 1), Gia_NodeMffcSize(p, Gia_ObjFanin0(obj)), vec.Vec_FltEntry(probability, int(id)))
	}
}
