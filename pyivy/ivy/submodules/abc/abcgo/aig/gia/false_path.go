package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"sort"
	"time"
)

func Gia_ManFalseRebuildOne(q, p *Gia_Man_t, hook *Vec_Int_t, verbose, veryVerbose int32) {
	ids := vec.Vec_IntArray(hook)
	if len(ids) == 0 {
		panic("gia: empty false path")
	}
	first := Gia_ManObj(p, ids[0])
	if len(ids) == 1 {
		Gia_ObjSetValue(first, 0)
		return
	}
	second := Gia_ManObj(p, ids[1])
	var control int32
	if Gia_ObjFanin0(second) == first {
		control = Gia_ObjFaninC0(second)
	} else if Gia_ObjFanin1(second) == first {
		control = Gia_ObjFaninC1(second)
	} else {
		panic("gia: invalid false path")
	}
	var prev *Gia_Obj_t
	var previous int32
	for i, id := range ids {
		obj := Gia_ManObj(p, id)
		old := Gia_ObjValue(obj)
		value := control
		if i > 0 {
			value = Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		}
		Gia_ObjSetValue(obj, value)
		if prev != nil {
			Gia_ObjSetValue(prev, previous)
		}
		prev, previous = obj, old
	}
	if veryVerbose != 0 {
		fmt.Printf("Eliminated path: %v\n", ids)
	}
}
func Gia_ManFalseRebuild(p *Gia_Man_t, hooks *Vec_Wec_t, verbose, veryVerbose int32) *Gia_Man_t {
	q := Gia_ManStart(max(1, 4*p.nObjs/3))
	q.pName = p.pName
	q.PSpec = p.PSpec
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	Gia_ManHashAlloc(q)
	for _, obj := range p.pObjs[1:] {
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			hook := vec.Vec_WecEntry(hooks, int(obj.id))
			if vec.Vec_IntSize(hook) > 0 {
				Gia_ManFalseRebuildOne(q, p, hook, verbose, veryVerbose)
			} else {
				Gia_ObjSetValue(obj, Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
			}
		case Gia_ObjIsCi(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManAppendCi(q))
		case Gia_ObjIsCo(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManAppendCo(q, Gia_ObjFanin0Copy(obj)))
		}
	}
	Gia_ManHashStop(q)
	Gia_ManSetRegNum(q, p.nRegs)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
func Gia_ManFalseRebuildPath(p *Gia_Man_t, path *Vec_Int_t, verbose, veryVerbose int32) *Gia_Man_t {
	ids := vec.Vec_IntArray(path)
	if len(ids) == 0 {
		panic("gia: empty false path")
	}
	hooks := vec.Vec_WecStart(int(p.nObjs))
	for _, id := range ids {
		vec.Vec_WecPush(hooks, int(ids[len(ids)-1]), id)
	}
	Gia_ManFillValue(p)
	return Gia_ManFalseRebuild(p, hooks, verbose, veryVerbose)
}
func Gia_ManCollectPath_rec(p *Gia_Man_t, obj *Gia_Obj_t, path *Vec_Int_t) {
	if Gia_ObjIsAnd(obj) != 0 {
		a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
		if Gia_ObjLevel(p, a) > Gia_ObjLevel(p, b) {
			Gia_ManCollectPath_rec(p, a, path)
		} else {
			Gia_ManCollectPath_rec(p, b, path)
		}
	}
	vec.Vec_IntPush(path, obj.id)
}
func Gia_ManCollectPath(p *Gia_Man_t, obj *Gia_Obj_t) *Vec_Int_t {
	path := vec.Vec_IntAlloc(int(p.nLevels))
	if Gia_ObjIsCo(obj) != 0 {
		obj = Gia_ObjFanin0(obj)
	}
	Gia_ManCollectPath_rec(p, obj, path)
	return path
}
func giaFalsePathSolve(p *Gia_Man_t, root int32, timeout int32, includeInput bool) (*Vec_Int_t, int) {
	path := Gia_ManCollectPath(p, Gia_ManObj(p, root))
	ids := vec.Vec_IntArray(path)
	objects := vec.Vec_IntArray(Gia_ManCollectNodesCis(p, []int32{root}, 1))
	n := len(ids)
	shift := [2]int{3 * n, 3*n + len(objects)}
	vars := make(map[int32]int, len(objects))
	for i, id := range objects {
		vars[id] = len(objects) - 1 - i
	}
	solver := satoko.New()
	solver.SetVariables(3*n + 2*len(objects))
	if timeout > 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	addAnd := func(out, a, b int) {
		solver.AddClause([]int{out ^ 1, a})
		solver.AddClause([]int{out ^ 1, b})
		solver.AddClause([]int{out, a ^ 1, b ^ 1})
	}
	addXor := func(out, a, b int) {
		solver.AddClause([]int{a, b, out ^ 1})
		solver.AddClause([]int{a ^ 1, b ^ 1, out ^ 1})
		solver.AddClause([]int{a, b ^ 1, out})
		solver.AddClause([]int{a ^ 1, b, out})
	}
	for _, id := range objects {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
		for k := 0; k < 2; k++ {
			addAnd((vars[id]+shift[k])*2, (vars[a]+shift[k])*2+int(Gia_ObjFaninC0(obj)), (vars[b]+shift[k])*2+int(Gia_ObjFaninC1(obj)))
		}
	}
	assumptions := make([]int, 0, n)
	for i, id := range ids {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) != 0 {
			if i == 0 {
				panic("gia: malformed false path")
			}
			previous := ids[i-1]
			for k := 0; k < 2; k++ {
				var a, b int
				if Gia_ObjFaninId0(obj, id) == previous {
					a = (i-1+(k+1)*n)*2 + int(Gia_ObjFaninC0(obj))
					b = (vars[Gia_ObjFaninId1(obj, id)]+shift[k])*2 + int(Gia_ObjFaninC1(obj))
				} else if Gia_ObjFaninId1(obj, id) == previous {
					a = (vars[Gia_ObjFaninId0(obj, id)]+shift[k])*2 + int(Gia_ObjFaninC0(obj))
					b = (i-1+(k+1)*n)*2 + int(Gia_ObjFaninC1(obj))
				} else {
					panic("gia: disconnected false path")
				}
				addAnd((i+(k+1)*n)*2, a, b)
			}
			addXor(i*2, (i+n)*2, (i+2*n)*2)
			assumptions = append(assumptions, i*2)
		} else if includeInput && Gia_ObjIsCi(obj) != 0 {
			addXor(i*2, (vars[id]+shift[0])*2, (vars[id]+shift[1])*2)
			assumptions = append(assumptions, i*2)
		}
	}
	core, status := solver.MinimizeAssumptions(assumptions, int64(timeout))
	if status != satoko.SATOKO_UNSAT || len(core) == 0 {
		return nil, status
	}
	begin, end := n, 0
	for _, lit := range core {
		begin = min(begin, lit>>1)
		end = max(end, lit>>1)
	}
	out := vec.Vec_IntAlloc(end - begin + 2)
	for i := max(begin-1, 0); i <= end; i++ {
		vec.Vec_IntPush(out, ids[i])
	}
	return out, status
}

func Gia_ManCheckFalseOne(p *Gia_Man_t, out, timeout int32, hooks *Vec_Wec_t, verbose, veryVerbose int32) {
	start := time.Now()
	pivot := Gia_ManCo(p, out)
	path, status := giaFalsePathSolve(p, Gia_ObjFaninId0(pivot, pivot.id), timeout, true)
	if path != nil && vec.Vec_IntSize(path) <= 20 {
		ids := vec.Vec_IntArray(path)
		available := true
		for _, id := range ids {
			if vec.Vec_IntSize(vec.Vec_WecEntry(hooks, int(id))) > 0 {
				available = false
				break
			}
		}
		if available {
			for _, id := range ids {
				vec.Vec_WecPush(hooks, int(ids[len(ids)-1]), id)
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("PO %6d : Level = %3d  ", out, Gia_ObjLevel(p, pivot))
		switch status {
		case satoko.SATOKO_UNDEC:
			fmt.Printf("Timeout reached after %d seconds. ", timeout)
		case satoko.SATOKO_SAT:
			fmt.Print("There is no false path. ")
		default:
			if path != nil {
				fmt.Printf("False path contains %d nodes. ", vec.Vec_IntSize(path))
				if veryVerbose != 0 {
					fmt.Printf("%v ", vec.Vec_IntArray(path))
				}
			}
		}
		fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
	}
}
func Gia_ManCheckFalse2(p *Gia_Man_t, slack, timeout, verbose, veryVerbose int32) *Gia_Man_t {
	Gia_ManLevelNum(p)
	order := make([]int32, Gia_ManCoNum(p))
	for i := range order {
		order[i] = int32(i)
	}
	sort.SliceStable(order, func(i, j int) bool {
		return Gia_ObjLevel(p, Gia_ManCo(p, order[i])) > Gia_ObjLevel(p, Gia_ManCo(p, order[j]))
	})
	hooks := vec.Vec_WecStart(int(p.nObjs))
	for _, i := range order {
		if Gia_ObjLevel(p, Gia_ManCo(p, i)) < p.nLevels-slack {
			break
		}
		Gia_ManCheckFalseOne(p, i, timeout, hooks, verbose, veryVerbose)
	}
	if verbose != 0 {
		fmt.Printf("Collected %d non-overlapping false paths.\n", vec.Vec_WecSizeUsed(hooks))
	}
	return Gia_ManFalseRebuild(p, hooks, verbose, veryVerbose)
}
func Gia_ManCheckOne(p *Gia_Man_t, out, id, timeout, verbose, veryVerbose int32) *Gia_Man_t {
	start := time.Now()
	path, status := giaFalsePathSolve(p, id, timeout, false)
	if verbose != 0 {
		fmt.Printf("PO %6d : Level = %3d  ", out, Gia_ObjLevelId(p, id))
		switch status {
		case satoko.SATOKO_UNDEC:
			fmt.Printf("Timeout reached after %d seconds. ", timeout)
		case satoko.SATOKO_SAT:
			fmt.Print("There is no false path. ")
		default:
			if path != nil {
				fmt.Printf("False path contains %d nodes. ", vec.Vec_IntSize(path))
				if veryVerbose != 0 {
					fmt.Printf("%v ", vec.Vec_IntArray(path))
				}
			}
		}
		fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
	}
	if path == nil {
		return nil
	}
	return Gia_ManFalseRebuildPath(p, path, verbose, veryVerbose)
}
func Gia_ManCheckFalseAll(p *Gia_Man_t, slack, timeout, verbose, veryVerbose int32) *Gia_Man_t {
	p = Gia_ManDup(p)
	tried, changed := 0, 0
	for {
		before := changed
		Gia_ManLevelNum(p)
		for id := int32(0); id < p.nObjs; id++ {
			obj := Gia_ManObj(p, id)
			if Gia_ObjIsAnd(obj) == 0 || Gia_ObjLevel(p, obj) > slack {
				continue
			}
			tried++
			q := Gia_ManCheckOne(p, -1, id, timeout, verbose, veryVerbose)
			if q == nil {
				continue
			}
			changed++
			Gia_ManStop(p)
			p = q
			Gia_ManLevelNum(p)
		}
		if before == changed {
			break
		}
	}
	fmt.Printf("Performed %d attempts and %d changes.\n", tried, changed)
	return p
}
func Gia_ManCheckFalse(p *Gia_Man_t, slack, timeout, verbose, veryVerbose int32) *Gia_Man_t {
	p = Gia_ManDup(p)
	tried := make([]bool, Gia_ManCoNum(p))
	attempts, changed := 0, 0
	for {
		before := changed
		level := Gia_ManLevelNum(p)
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			obj := Gia_ManCo(p, i)
			driver := Gia_ObjFanin0(obj)
			if Gia_ObjIsAnd(driver) == 0 || Gia_ObjLevel(p, driver) < level-slack || tried[i] {
				continue
			}
			attempts++
			q := Gia_ManCheckOne(p, i, driver.id, timeout, verbose, veryVerbose)
			if q == nil {
				tried[i] = true
				continue
			}
			changed++
			Gia_ManStop(p)
			p = q
			level = Gia_ManLevelNum(p)
		}
		if before == changed {
			break
		}
	}
	fmt.Printf("Performed %d attempts and %d changes.\n", attempts, changed)
	return p
}
