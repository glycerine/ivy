package gia

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Gia_ManDupWithMuxPos adds outputs at shared MUX boundaries.
func Gia_ManDupWithMuxPos(p *Gia_Man_t) *Gia_Man_t {
	q := Gia_ManDup(p)
	if Gia_ManRegNum(q) != 0 {
		panic("gia: MUX outputs require a combinational graph")
	}
	type quad struct{ root, control, otherwise, then int32 }
	var quads []quad
	heads, datas := make([]bool, q.nObjs), make([]bool, q.nObjs)
	for _, obj := range q.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var then, otherwise *Gia_Obj_t
		control := Gia_ObjRecognizeMux(obj, &then, &otherwise)
		item := quad{obj.id, Gia_Regular(control).id, Gia_Regular(otherwise).id, Gia_Regular(then).id}
		quads = append(quads, item)
		heads[item.root] = true
		datas[item.otherwise] = true
		datas[item.then] = true
	}
	for i := int32(0); i < Gia_ManCoNum(q); i++ {
		Gia_ObjFanin0(Gia_ManCo(q, i)).fMark0 = 1
	}
	points := make(map[int32]bool)
	for _, item := range quads {
		if !(heads[item.root] && datas[item.root] || heads[item.otherwise] && datas[item.otherwise] || heads[item.then] && datas[item.then]) {
			continue
		}
		for _, id := range []int32{item.root, item.control, item.otherwise, item.then} {
			obj := Gia_ManObj(q, id)
			if Gia_ObjIsAnd(obj) != 0 && obj.fMark0 == 0 {
				points[id] = true
			}
		}
	}
	Gia_ManCleanMark0(q)
	ids := make([]int32, 0, len(points))
	for id := range points {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		Gia_ManAppendCo(q, id<<1)
	}
	return q
}
func Gia_ComputeTest() {
	f, err := os.Open("data.txt")
	if err != nil {
		fmt.Println("Input file \"data.txt\" cannot be opened.")
		return
	}
	defer f.Close()
	var total float64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "xxx") || !strings.Contains(line, "yyy") {
			continue
		}
		i := strings.Index(line, "zzz")
		if i < 0 || i+4 > len(line) {
			continue
		}
		var value float64
		if _, err := fmt.Sscan(line[i+4:], &value); err == nil {
			total -= value
		}
	}
	fmt.Printf("Total = %.2f\n", total)
}
