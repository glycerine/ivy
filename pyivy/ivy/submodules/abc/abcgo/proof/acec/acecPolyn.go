package acec

// Priority-ordered polynomial expansion from acecPolyn.c.
import (
	"fmt"
	"sort"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

type Pln_Man_t struct {
	graph                *gia.Gia_Man_t
	constants, monomials *vec.Hsh_VecMan_t
	queue                *vec.Vec_Que_t
	counts               *vec.Vec_Flt_t
	coefs                *vec.Vec_Int_t
	tempC                [2]*vec.Vec_Int_t
	tempM                [4]*vec.Vec_Int_t
	order                *vec.Vec_Int_t
	builds, used         int32
}

func Pln_ManAlloc(graph *gia.Gia_Man_t, order *vec.Vec_Int_t) *Pln_Man_t {
	p := &Pln_Man_t{graph: graph, constants: vec.Hsh_VecManStart(1000), monomials: vec.Hsh_VecManStart(1000), queue: vec.Vec_QueAlloc(1000), counts: vec.Vec_FltAlloc(1000), coefs: vec.Vec_IntAlloc(1000)}
	for i := range p.tempC {
		p.tempC[i] = vec.Vec_IntAlloc(100)
	}
	for i := range p.tempM {
		p.tempM[i] = vec.Vec_IntAlloc(100)
	}
	if order == nil {
		p.order = vec.Vec_IntStartNatural(int(gia.Gia_ManObjNum(graph)))
	} else {
		p.order = vec.Vec_IntDup(order)
	}
	if vec.Vec_IntSize(p.order) != int(gia.Gia_ManObjNum(graph)) {
		panic("invalid polynomial collapse order")
	}
	vec.Vec_QueSetPriority(p.queue, vec.Vec_FltArrayP(p.counts))
	vec.Hsh_VecManAdd(p.constants, p.tempC[0])
	vec.Hsh_VecManAdd(p.monomials, p.tempM[0])
	vec.Vec_FltPush(p.counts, 0)
	vec.Vec_IntPush(p.coefs, 0)
	return p
}
func Pln_ManStop(p *Pln_Man_t) {
	if p != nil {
		*p = Pln_Man_t{}
	}
}
func Pln_ManCompare3(a, b []int32) int32 {
	if a[0] < b[0] || a[0] == b[0] && a[1] < b[1] {
		return -1
	}
	if a[0] > b[0] || a[0] == b[0] && a[1] > b[1] {
		return 1
	}
	return 0
}
func Pln_ManPrintFinal(p *Pln_Man_t, verbose, veryVerbose int32) {
	type row struct{ key, first, constant, mono int32 }
	rows := []row{}
	for mono, c := range vec.Vec_IntArray(p.coefs) {
		if c == 0 {
			continue
		}
		cv := vec.Hsh_VecReadArray(p.constants, c)
		mv := vec.Hsh_VecReadArray(p.monomials, int32(mono))
		first := int32(0)
		if len(mv) > 0 {
			first = mv[0]
		}
		rows = append(rows, row{cv[0], first, c, int32(mono)})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].key < rows[j].key || rows[i].key == rows[j].key && rows[i].first < rows[j].first
	})
	if verbose != 0 {
		for i, r := range rows {
			fmt.Printf("%-6d : ", i)
			for _, c := range vec.Hsh_VecReadArray(p.constants, r.constant) {
				sign := "+"
				if c < 0 {
					sign = "-"
				}
				fmt.Printf("%s%d", sign, int64(1)<<uint(abs32(c)-1))
			}
			for _, v := range vec.Hsh_VecReadArray(p.monomials, r.mono) {
				fmt.Printf(" * %d", v)
			}
			fmt.Println()
		}
	}
	fmt.Printf("HashC = %d. HashM = %d.  Total = %d. Used = %d.  ", vec.Hsh_VecSize(p.constants), vec.Hsh_VecSize(p.monomials), p.builds, len(rows))
}
func plnMinus2x(dst, src *vec.Vec_Int_t) {
	vec.Vec_IntClear(dst)
	for _, c := range vec.Vec_IntArray(src) {
		if c > 0 {
			vec.Vec_IntPush(dst, -c-1)
		} else {
			vec.Vec_IntPush(dst, -c+1)
		}
	}
}
func Gia_PolynMergeConstOne(coef *vec.Vec_Int_t, n int32) {
	if n == 0 {
		panic("zero polynomial coefficient")
	}
	for i, old := range vec.Vec_IntArray(coef) {
		if old == n {
			vec.Vec_IntDrop(coef, i)
			if n > 0 {
				Gia_PolynMergeConstOne(coef, n+1)
			} else {
				Gia_PolynMergeConstOne(coef, n-1)
			}
			return
		}
		if abs32(old) == abs32(n) {
			vec.Vec_IntDrop(coef, i)
			return
		}
		if old+n == 1 || old+n == -1 {
			vec.Vec_IntDrop(coef, i)
			v := abs32(old)
			if abs32(n) < v {
				v = abs32(n)
			}
			if old+n < 0 {
				v = -v
			}
			Gia_PolynMergeConstOne(coef, v)
			return
		}
	}
	vec.Vec_IntPushUniqueOrder(coef, n)
}
func Gia_PolynMergeConst(coef *vec.Vec_Int_t, p *Pln_Man_t, index int32) {
	for _, v := range append([]int32(nil), vec.Hsh_VecReadArray(p.constants, index)...) {
		Gia_PolynMergeConstOne(coef, v)
	}
}
func Gia_PolynBuildAdd(p *Pln_Man_t, coef, mono *vec.Vec_Int_t) {
	id := int32(0)
	if mono != nil {
		id = vec.Hsh_VecManAdd(p.monomials, mono)
	}
	p.builds++
	if int(id) == vec.Vec_IntSize(p.coefs) {
		c := vec.Hsh_VecManAdd(p.constants, coef)
		vec.Vec_IntPush(p.coefs, c)
		driver := vec.Vec_IntEntryLast(mono)
		vec.Vec_FltPush(p.counts, float32(vec.Vec_IntEntry(p.order, int(driver))))
		vec.Vec_QuePush(p.queue, int(id))
		if c != 0 {
			p.used++
		}
		return
	}
	old := vec.Vec_IntEntry(p.coefs, int(id))
	if old != 0 {
		Gia_PolynMergeConst(coef, p, old)
	}
	current := vec.Hsh_VecManAdd(p.constants, coef)
	vec.Vec_IntWriteEntry(p.coefs, int(id), current)
	if old != 0 && current == 0 {
		p.used--
	} else if old == 0 && current != 0 {
		p.used++
	}
}
func Gia_PolynBuildOne(p *Pln_Man_t, mono int32) {
	values := append([]int32(nil), vec.Hsh_VecReadArray(p.monomials, mono)...)
	if len(values) == 0 {
		return
	}
	driver := values[len(values)-1]
	obj := gia.Gia_ManObj(p.graph, driver)
	if gia.Gia_ObjIsAnd(obj) == 0 {
		return
	}
	if gia.Gia_ObjIsMux(p.graph, obj) != 0 {
		panic("polynomial expansion does not support mux")
	}
	c := vec.Vec_IntEntry(p.coefs, int(mono))
	if c == 0 {
		return
	}
	vec.Vec_IntWriteEntry(p.coefs, int(mono), 0)
	p.used--
	a, b := gia.Gia_ObjFaninId0(obj, driver), gia.Gia_ObjFaninId1(obj, driver)
	for k := 0; k < 4; k++ {
		v := p.tempM[k]
		vec.Vec_IntClear(v)
		vec.Vec_IntPushArray(v, values, len(values)-1)
		if k == 1 || k == 3 {
			vec.Vec_IntPushUniqueOrderCost(v, a, p.order)
		}
		if k == 2 || k == 3 {
			vec.Vec_IntPushUniqueOrderCost(v, b, p.order)
		}
	}
	coefficients := append([]int32(nil), vec.Hsh_VecReadArray(p.constants, c)...)
	add := func(idx int, negate, twice bool) {
		v := p.tempC[0]
		vec.Vec_IntClear(v)
		for _, x := range coefficients {
			if twice {
				if x > 0 {
					x = -x - 1
				} else {
					x = -x + 1
				}
			} else if negate {
				x = -x
			}
			vec.Vec_IntPush(v, x)
		}
		Gia_PolynBuildAdd(p, v, p.tempM[idx])
	}
	if gia.Gia_ObjIsXor(obj) != 0 {
		add(1, false, false)
		add(2, false, false)
		add(3, false, true)
	} else if gia.Gia_ObjFaninC0(obj) != 0 && gia.Gia_ObjFaninC1(obj) != 0 {
		add(0, false, false)
		add(1, true, false)
		add(2, true, false)
		add(3, false, false)
	} else if gia.Gia_ObjFaninC0(obj) != 0 {
		add(2, false, false)
		add(3, true, false)
	} else if gia.Gia_ObjFaninC1(obj) != 0 {
		add(1, false, false)
		add(3, true, false)
	} else {
		add(3, false, false)
	}
}
func Gia_PolynBuild(graph *gia.Gia_Man_t, order *vec.Vec_Int_t, signed, verbose, veryVerbose int32) {
	started := time.Now()
	p := Pln_ManAlloc(graph, order)
	defer Pln_ManStop(p)
	for i := gia.Gia_ManCoNum(graph) - 1; i >= 0; i-- {
		obj := gia.Gia_ManCo(graph, i)
		id := gia.Gia_ObjId(graph, obj)
		lit := gia.Gia_ObjFaninLit0(obj, id)
		coefficient := i + 1
		if signed != 0 && i == gia.Gia_ManCoNum(graph)-1 {
			coefficient = -coefficient
		}
		vec.Vec_IntFill(p.tempC[0], 1, coefficient)
		vec.Vec_IntFill(p.tempC[1], 1, -coefficient)
		vec.Vec_IntFill(p.tempM[0], 1, lit>>1)
		if lit&1 != 0 {
			Gia_PolynBuildAdd(p, p.tempC[0], nil)
			Gia_PolynBuildAdd(p, p.tempC[1], p.tempM[0])
		} else {
			Gia_PolynBuildAdd(p, p.tempC[0], p.tempM[0])
		}
	}
	for vec.Vec_QueSize(p.queue) > 0 {
		mono := vec.Vec_QuePop(p.queue)
		Gia_PolynBuildOne(p, int32(mono))
	}
	Pln_ManPrintFinal(p, verbose, veryVerbose)
	fmt.Printf("Time = %.2f sec\n", time.Since(started).Seconds())
}

// Gia_PolynBuild2 is empty in the original source apart from allocating and
// releasing two hash tables.
func Gia_PolynBuild2(_ *gia.Gia_Man_t, _, _, _ int32) {}
