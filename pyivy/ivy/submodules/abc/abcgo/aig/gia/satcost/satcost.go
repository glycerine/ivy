// Package satcost runs source-ordered MF CNF through the native Satoko solver.
// CNF generation is explicit so a different encoding cannot silently change a
// command tuner's conflict-cost objective.
package satcost

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

// Backend identifies the exact CNF generator used for a tuning experiment.
// Implementations must be safe for concurrent calls when used by CmdAuto.
type Backend interface {
	Name() string
	Generate(*ioa.Model) (*mf.CNF, error)
}
type Result struct {
	Status    int   // native SATOKO_SAT/UNSAT/UNDEC constants
	Conflicts int64 // this solve's n_conflicts, not the cumulative counter
	Elapsed   time.Duration
	Backend   string
}
type Evaluator struct{ Backend Backend }

func (e Evaluator) Evaluate(model *ioa.Model, options satoko.Options) (Result, error) {
	if e.Backend == nil {
		return Result{}, fmt.Errorf("satcost: an explicit mapped-CNF backend is required")
	}
	cnf, err := e.Backend.Generate(model)
	if err != nil {
		return Result{}, err
	}
	result, err := RunCNF(cnf, options)
	result.Backend = e.Backend.Name()
	return result, err
}

// Cost matches the command tuner's injected AutoEvaluator signature.
func (e Evaluator) Cost(model *ioa.Model, options satoko.Options) (int64, error) {
	result, err := e.Evaluate(model, options)
	return result.Conflicts, err
}

// RunCNF follows Gia_ManSatokoInit/Create/CallOne: insert clauses in source
// order, configure, simplify, solve, and read the per-solve conflict counter.
// A contradiction during insertion/simplification has zero tuning cost.
func RunCNF(cnf *mf.CNF, options satoko.Options) (Result, error) {
	start := time.Now()
	result := Result{Status: satoko.SATOKO_UNSAT}
	if cnf == nil || cnf.Variables < 0 {
		return result, fmt.Errorf("satcost: invalid CNF")
	}
	for _, clause := range cnf.Clauses {
		for _, lit := range clause {
			if lit < 0 || int(lit>>1) >= cnf.Variables {
				return result, fmt.Errorf("satcost: invalid CNF literal %d", lit)
			}
		}
	}
	solver := satoko.New()
	for _, clause := range cnf.Clauses {
		literals := make([]int, len(clause))
		for i, lit := range clause {
			literals[i] = int(lit)
		}
		if !solver.AddClause(literals) {
			result.Elapsed = time.Since(start)
			return result, nil
		}
	}
	*solver.Options() = options
	if !solver.Simplify() {
		result.Elapsed = time.Since(start)
		return result, nil
	}
	result.Status = solver.Solve()
	result.Conflicts = solver.Stats().Conflicts
	result.Elapsed = time.Since(start)
	return result, nil
}

// MFBackend uses the actual eight-cut MF coarsening/mapping/CNF pipeline.
// The native solver's conflict trace is audited independently; CNF fidelity
// does not by itself establish identical costs to the original C solver.
type MFBackend struct{}

func (MFBackend) Name() string                               { return "mf8-source-order/native-satoko" }
func (MFBackend) Generate(model *ioa.Model) (*mf.CNF, error) { return mf.GenerateModelCNF(model) }

// EvaluateModel is the native pure-Go MF8/Satoko conflict-cost adapter. It is
// deliberately registered by callers, so choosing this audited native backend
// is explicit; audited profiles and remaining trace-coverage limits are documented in REVIEW.md.
func EvaluateModel(model *ioa.Model, options satoko.Options) (int64, error) {
	return (Evaluator{Backend: MFBackend{}}).Cost(model, options)
}
