package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
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
		fmt.Print("Eliminated path: ")
		vec.Vec_IntPrint(hook)
		for _, id := range ids {
			obj := p.pObjs[id]
			fmt.Printf("Level %3d : ", Gia_ObjLevel(p, obj))
			Gia_ObjPrint(p, obj)
		}
	}
}
func Gia_ManFalseRebuild(p *Gia_Man_t, hooks *Vec_Wec_t, verbose, veryVerbose int32) *Gia_Man_t {
	return giaFalseRebuild(p, hooks, verbose, veryVerbose, true)
}
func giaFalseRebuild(p *Gia_Man_t, hooks *Vec_Wec_t, verbose, veryVerbose int32, numberPaths bool) *Gia_Man_t {
	q := Gia_ManStart(max(1, 4*p.nObjs/3))
	q.pName = p.pName
	q.PSpec = p.PSpec
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	Gia_ManHashAlloc(q)
	read := func(lit int32) int32 { return int32(p.pObjs[lit>>1].Value) ^ (lit & 1) }
	counter := 0
	for _, obj := range p.pObjs[1:] {
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			hook := vec.Vec_WecEntry(hooks, int(obj.id))
			if vec.Vec_IntSize(hook) > 0 {
				if veryVerbose != 0 && numberPaths {
					fmt.Printf("Path %d : ", counter)
					counter++
				}
				Gia_ManFalseRebuildOne(q, p, hook, verbose, veryVerbose)
			} else {
				Gia_ObjSetValue(obj, giaHashMappedGate(q, p, obj, read))
			}
		case Gia_ObjIsCi(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManAppendCi(q))
		case Gia_ObjIsCo(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManAppendCo(q, read(Gia_ObjFaninLit0(obj, obj.id))))
		}
	}
	Gia_ManHashStop(q)
	Gia_ManDupRemapCis(q, p)
	Gia_ManDupRemapCos(q, p)
	Gia_ManSetRegNum(q, p.nRegs)
	giaCopyInterfaceMetadata(q, p)
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
	return giaFalseRebuild(p, hooks, verbose, veryVerbose, false)
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

// giaFalseQuery owns the original path and the final assumption conflict.
// Keep core size and rewritten interval separate, as the two C drivers do.
type giaFalseQuery struct {
	ids    []int32
	core   []bsat.Literal
	status bsat.Status
}

func (q giaFalseQuery) span() (int, int) {
	begin, end := int(q.core[len(q.core)-1]>>1), int(q.core[0]>>1)
	if begin > end {
		panic("gia: unordered false-path core")
	}
	return begin, end
}
func (q giaFalseQuery) hook() *Vec_Int_t {
	if q.status != bsat.Unsatisfiable || len(q.core) == 0 {
		return nil
	}
	begin, end := q.span()
	path := vec.Vec_IntAlloc(end - begin + 2)
	for i := max(begin-1, 0); i <= end; i++ {
		vec.Vec_IntPush(path, q.ids[i])
	}
	return path
}
func (q giaFalseQuery) addBatchHook(hooks *Vec_Wec_t) bool {
	if q.status != bsat.Unsatisfiable || len(q.core) == 0 {
		return false
	}
	begin, end := q.span()
	if end-begin >= 20 {
		return false
	}
	ids := q.ids[max(begin-1, 0) : end+1]
	for _, id := range ids {
		if vec.Vec_IntSize(vec.Vec_WecEntry(hooks, int(id))) > 0 {
			return false
		}
	}
	for _, id := range ids {
		vec.Vec_WecPush(hooks, int(ids[len(ids)-1]), id)
	}
	return true
}
func giaFalsePathSolve(p *Gia_Man_t, root, timeout int32, includeInput bool) (*Vec_Int_t, bsat.Status) {
	q := giaFalseSolveQuery(p, root, timeout, includeInput)
	return q.hook(), q.status
}
func giaFalseSolveQuery(p *Gia_Man_t, root int32, timeout int32, includeInput bool) giaFalseQuery {
	path := Gia_ManCollectPath(p, Gia_ManObj(p, root))
	ids := vec.Vec_IntArray(path)
	objects := vec.Vec_IntArray(Gia_ManCollectNodesCis(p, []int32{root}, 1))
	n := len(ids)
	shift := [2]int{3 * n, 3*n + len(objects)}
	for i, id := range objects {
		p.pObjs[id].Value = uint32(len(objects) - 1 - i)
	}
	satVar := func(id int32) int { return int(p.pObjs[id].Value) }
	solver := bsat.NewSolver()
	defer solver.Close()
	solver.SetNumVars(int32(3*n + 2*len(objects)))
	solver.SetLimits(int64(timeout), 0)
	if timeout > 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	addAnd := func(out, a, b int) {
		solver.AddClause(int32(out^1), int32(a))
		solver.AddClause(int32(out^1), int32(b))
		solver.AddClause(int32(out), int32(a^1), int32(b^1))
	}
	addXor := func(out, a, b int) { solver.AddXor(int32(out>>1), int32(a>>1), int32(b>>1), false) }
	for _, id := range objects {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
		for k := 0; k < 2; k++ {
			addAnd((satVar(id)+shift[k])*2, (satVar(a)+shift[k])*2+int(Gia_ObjFaninC0(obj)), (satVar(b)+shift[k])*2+int(Gia_ObjFaninC1(obj)))
		}
	}
	assumptions := make([]bsat.Literal, 0, n)
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
					b = (satVar(Gia_ObjFaninId1(obj, id))+shift[k])*2 + int(Gia_ObjFaninC1(obj))
				} else if Gia_ObjFaninId1(obj, id) == previous {
					a = (satVar(Gia_ObjFaninId0(obj, id))+shift[k])*2 + int(Gia_ObjFaninC0(obj))
					b = (i-1+(k+1)*n)*2 + int(Gia_ObjFaninC1(obj))
				} else {
					panic("gia: disconnected false path")
				}
				addAnd((i+(k+1)*n)*2, a, b)
			}
			addXor(i*2, (i+n)*2, (i+2*n)*2)
			assumptions = append(assumptions, int32(i*2))
		} else if includeInput && Gia_ObjIsCi(obj) != 0 {
			addXor(i*2, (satVar(id)+shift[0])*2, (satVar(id)+shift[1])*2)
			assumptions = append(assumptions, int32(i*2))
		}
	}
	// C uses one BSAT solve and its final assumption conflict, without a
	// separate minimization pass. Core choice determines the rewritten interval.
	status := solver.Solve(assumptions...)
	core := solver.UnsatCore()
	return giaFalseQuery{ids: ids, core: core, status: status}
}

func Gia_ManCheckFalseOne(p *Gia_Man_t, out, timeout int32, hooks *Vec_Wec_t, verbose, veryVerbose int32) {
	started := time.Now()
	pivot := Gia_ManCo(p, out)
	query := giaFalseSolveQuery(p, Gia_ObjFaninId0(pivot, pivot.id), timeout, true)
	query.addBatchHook(hooks)
	if verbose != 0 {
		fmt.Printf("PO %6d : Level = %3d  ", out, Gia_ObjLevel(p, pivot))
		switch query.status {
		case bsat.Unknown:
			fmt.Printf("Timeout reached after %d seconds. ", timeout)
		case bsat.Satisfiable:
			fmt.Print("There is no false path. ")
		case bsat.Unsatisfiable:
			if len(query.core) > 0 {
				fmt.Printf("False path contains %d nodes (out of %d):  top = %d  ", len(query.core), len(query.ids), query.ids[query.core[0]>>1])
				if veryVerbose != 0 {
					for _, lit := range query.core {
						fmt.Printf("%d ", lit>>1)
					}
				}
				fmt.Print("  ")
			}
		}
		fmt.Printf("Time =%9.2f sec\n", time.Since(started).Seconds())
	}
}
func Gia_ManCheckFalse2(p *Gia_Man_t, slack, timeout, verbose, veryVerbose int32) *Gia_Man_t {
	Gia_ManLevelNum(p)
	weights := make([]float32, Gia_ManCoNum(p))
	priority := vec.Vec_QueAlloc(len(weights))
	defer vec.Vec_QueFree(priority)
	vec.Vec_QueSetPriority(priority, &weights)
	for i := range weights {
		weights[i] = float32(Gia_ObjLevel(p, Gia_ManCo(p, int32(i))))
		vec.Vec_QuePush(priority, i)
	}
	hooks := vec.Vec_WecStart(int(p.nObjs))
	for vec.Vec_QueSize(priority) > 0 && vec.Vec_QueTopPriority(priority) >= float32(p.nLevels-slack) {
		Gia_ManCheckFalseOne(p, int32(vec.Vec_QuePop(priority)), timeout, hooks, verbose, veryVerbose)
	}
	if verbose != 0 {
		fmt.Printf("Collected %d non-overlapping false paths.\n", vec.Vec_WecSizeUsed(hooks))
	}
	return Gia_ManFalseRebuild(p, hooks, verbose, veryVerbose)
}
func Gia_ManCheckOne(p *Gia_Man_t, out, id, timeout, verbose, veryVerbose int32) *Gia_Man_t {
	started := time.Now()
	query := giaFalseSolveQuery(p, id, timeout, false)
	path := query.hook()
	if verbose != 0 {
		fmt.Printf("PO %6d : Level = %3d  ", out, Gia_ObjLevelId(p, id))
		switch query.status {
		case bsat.Unknown:
			fmt.Printf("Timeout reached after %d seconds. ", timeout)
		case bsat.Satisfiable:
			fmt.Print("There is no false path. ")
		case bsat.Unsatisfiable:
			if path != nil {
				fmt.Printf("False path contains %d nodes (out of %d):  ", vec.Vec_IntSize(path), len(query.ids))
				if veryVerbose != 0 {
					for i := len(query.core) - 1; i >= 0; i-- {
						fmt.Printf("%d ", query.ids[query.core[i]>>1])
					}
				}
				fmt.Print("  ")
			}
		}
		fmt.Printf("Time =%9.2f sec\n", time.Since(started).Seconds())
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
