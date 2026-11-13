package bmc

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Bmc_ManBCoreReadPivots reads source object IDs from a whitespace-delimited file.
func Bmc_ManBCoreReadPivots(name string) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(100)
	data, err := os.ReadFile(name)
	if err != nil {
		return result
	}
	for _, id := range parsePivotIDs(data) {
		vec.Vec_IntPush(result, id)
	}
	return result
}

// Bmc_ManBCoreCollectPivots maps object IDs to all corresponding SAT variables.
func Bmc_ManBCoreCollectPivots(p *gia.Gia_Man_t, name string, varMap *vec.Vec_Int_t) *vec.Vec_Int_t {
	selected := make(map[int32]bool)
	for _, id := range vec.Vec_IntArray(Bmc_ManBCoreReadPivots(name)) {
		if id > 0 && id < gia.Gia_ManObjNum(p) {
			selected[id] = true
		} else {
			fmt.Printf("Cannot find object with Id %d in the given AIG.\n", id)
		}
	}
	result := vec.Vec_IntAlloc(100)
	values := vec.Vec_IntArray(varMap)
	for i := 0; i+1 < len(values); i += 2 {
		if selected[values[i]] {
			vec.Vec_IntPush(result, int32(i/2))
		}
	}
	return result
}

// Bmc_ManBCoreCollect constructs the source's reverse-time SAT unrolling.
// The returned pairs map SAT variable indices to (object ID, frame).
func Bmc_ManBCoreCollect(p *gia.Gia_Man_t, frame, output int32, solver *bsat.Solver) *vec.Vec_Int_t {
	if p == nil || solver == nil || frame < 0 || output < 0 || output >= gia.Gia_ManPoNum(p) {
		panic("bmc: invalid core request")
	}
	nodes := vec.Vec_IntAlloc(100)
	vec.Vec_IntPushTwo(nodes, -1, -1)
	allocate := func(id, f int32) int32 {
		v := int32(vec.Vec_IntSize(nodes) / 2)
		vec.Vec_IntPushTwo(nodes, id, f)
		for solver.NumVars() <= v {
			solver.NewVariable()
		}
		return v
	}
	po := gia.Gia_ObjId(p, gia.Gia_ManPo(p, output))
	outVar := allocate(po, frame)
	roots := []int32{po}
	links := map[int32]int32{}
	pis := gia.Gia_ManPiNum(p)
	inputIndex := map[int32]int32{}
	for i, lit := range p.InputLiterals() {
		inputIndex[lit>>1] = int32(i)
	}
	add := func(c ...bsat.Literal) { solver.AddClause(c...) }
	for f := frame; f >= 0; f-- {
		values := map[int32]int32{0: allocate(0, f)}
		add(bsat.Negative(values[0]))
		visited := map[int32]bool{0: true}
		var next, order []int32
		var visit func(int32)
		visit = func(id int32) {
			if visited[id] {
				return
			}
			visited[id] = true
			values[id] = allocate(id, f)
			order = append(order, id)
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 1:
				if inputIndex[id] >= pis {
					idx := inputIndex[id] - pis
					next = append(next, gia.Gia_ObjId(p, gia.Gia_ManRi(p, idx)))
				}
				return
			case 2, 3:
				visit(a >> 1)
				visit(b >> 1)
			case 4:
				visit(a >> 1)
				visit(b >> 1)
				visit(c >> 1)
			case 5:
				visit(a >> 1)
			default:
				panic("bmc: unsupported BCore object")
			}
		}
		for _, id := range roots {
			a, _, _, kind := p.Gate(id)
			if kind != 6 {
				panic("bmc: invalid core root")
			}
			visit(a >> 1)
		}
		lit := func(a int32) int32 { return bsat.Positive(values[a>>1]) ^ (a & 1) }
		for _, id := range roots {
			a, _, _, _ := p.Gate(id)
			if f == frame {
				forced := lit(a)
				add(forced, bsat.Positive(outVar))
				add(forced, bsat.Negative(outVar))
			} else {
				target := bsat.Positive(links[id])
				input := lit(a)
				add(target^1, input)
				add(target, input^1)
			}
		}
		newLinks := make(map[int32]int32)
		for _, ri := range next {
			idx := int32(-1)
			for j := int32(0); j < gia.Gia_ManRegNum(p); j++ {
				if gia.Gia_ObjId(p, gia.Gia_ManRi(p, j)) == ri {
					idx = j
					break
				}
			}
			if idx < 0 {
				panic("bmc: invalid register input")
			}
			ro := gia.Gia_ObjId(p, gia.Gia_ManRo(p, idx))
			newLinks[ri] = values[ro]
		}
		for _, id := range order {
			a, b, c, kind := p.Gate(id)
			z := bsat.Positive(values[id])
			var clauses [][]bsat.Literal
			switch kind {
			case 1:
				continue
			case 2:
				x, y := lit(a), lit(b)
				clauses = [][]bsat.Literal{{z ^ 1, x}, {z ^ 1, y}, {z, x ^ 1, y ^ 1}}
			case 3:
				x, y := lit(a), lit(b)
				clauses = [][]bsat.Literal{{x ^ 1, y ^ 1, z ^ 1}, {x, y, z ^ 1}, {x, y ^ 1, z}, {x ^ 1, y, z}}
			case 4:
				x, y, w := lit(a), lit(b), lit(c)
				clauses = [][]bsat.Literal{{x ^ 1, y ^ 1, z}, {x ^ 1, y, z ^ 1}, {x, w ^ 1, z}, {x, w, z ^ 1}}
			case 5:
				x := lit(a)
				clauses = [][]bsat.Literal{{z ^ 1, x}, {z, x ^ 1}}
			}
			for _, clause := range clauses {
				add(clause...)
			}
		}
		roots, links = next, newLinks
	}
	for _, id := range roots {
		v := links[id]
		idx := int32(-1)
		for j := int32(0); j < gia.Gia_ManRegNum(p); j++ {
			if gia.Gia_ObjId(p, gia.Gia_ManRi(p, j)) == id {
				idx = j
				break
			}
		}
		if idx < 0 {
			panic("bmc: invalid initial root")
		}
		ro := gia.Gia_ObjId(p, gia.Gia_ManRo(p, idx))
		helper := allocate(ro, -1)
		add(bsat.Negative(v), bsat.Positive(helper))
		add(bsat.Negative(v), bsat.Negative(helper))
	}
	return nodes
}

// Bmc_ManBCorePerform computes and writes a learned-clause BMC proof core.
func Bmc_ManBCorePerform(p *gia.Gia_Man_t, params *Bmc_BCorePar_t) {
	start := time.Now()
	solver := bsat.NewSolver()
	solver.EnableClauseStore()
	if params.nTimeOut > 0 {
		solver.SetDeadline(start.Add(time.Duration(params.nTimeOut) * time.Second))
	}
	mapping := Bmc_ManBCoreCollect(p, params.iFrame, params.iOutput, solver)
	solver.MarkClauseStoreRoots()
	if params.pFilePivots != "" {
		pivots := Bmc_ManBCoreCollectPivots(p, params.pFilePivots, mapping)
		solver.SetPivotVariables(vec.Vec_IntArray(pivots))
	}
	status := solver.Solve()
	if status == bsat.Unknown {
		fmt.Println("Timeout of conflict limit is reached.")
		return
	}
	if status == bsat.Satisfiable {
		fmt.Println("The BMC problem is SAT.")
		return
	}
	if params.fVerbose != 0 {
		fmt.Printf("SAT solver returned UNSAT after %7d conflicts. Time %s\n", solver.Statistics().Conflicts, time.Since(start))
	}
	store := solver.ReleaseClauseStore()
	proof := bsat.Intp_ManAlloc()
	core := bsat.Intp_ManUnsatCore(proof, store, 1, params.fVerbose)
	if err := bsat.Intp_ManError(proof); err != nil {
		fmt.Printf("Cannot derive BMC proof: %v\n", err)
		return
	}
	sort.Slice(core, func(i, j int) bool { return core[i] < core[j] })
	file := os.Stdout
	if params.pFileProof != "" {
		var err error
		file, err = os.Create(params.pFileProof)
		if err != nil {
			fmt.Printf("Cannot write BMC proof: %v\n", err)
			return
		}
		defer file.Close()
	}
	bsat.Intp_ManUnsatCorePrintForBmc(file, store, core, vec.Vec_IntArray(mapping))
}
