package typeinfer

import (
	"fmt"

	"github.com/glycerine/goivy/logic"
)

// InferResult holds the inferred sort and a closure to concretize the term.
type InferResult struct {
	Sort       SortOrVar
	Concretize func() (logic.Expr, error)
}

// InferSorts performs type inference on a term, returning the inferred sort
// and a closure that concretizes the term.
func InferSorts(t logic.Expr, env map[string]SortOrVar) (*InferResult, error) {
	if env == nil {
		env = make(map[string]SortOrVar)
		collectNames(t, env)
	}

	switch n := t.(type) {
	case *logic.Variable:
		// Check env first - if this variable was bound by a quantifier,
		// the env has a sort var that will be unified with the concrete
		// sort from the body's function applications.
		s, ok := env[n.Name]
		if !ok {
			if logic.IsPolymorphic(n) {
				s = InsertSortVars(n.VSort, map[string]SortOrVar{})
			} else {
				s = NewSortVar()
			}
			env[n.Name] = s
		}
		// Unify env sort var with the variable's declared sort (if concrete).
		if !logic.IsTopSort(n.VSort) {
			ts := ConvertToSortVars(n.VSort)
			if err := Unify(s, ts); err != nil {
				return nil, err
			}
		}
		return &InferResult{
			Sort: s,
			Concretize: func() (logic.Expr, error) {
				cs := ConvertFromSortVars(s)
				return logic.NewVariable(n.Name, cs)
			},
		}, nil

	case *logic.Symbol:
		if logic.IsPolymorphic(n) {
			s := InsertSortVars(n.CSort, map[string]SortOrVar{})
			return &InferResult{
				Sort: s,
				Concretize: func() (logic.Expr, error) {
					return logic.NewSymbol(n.Name, ConvertFromSortVars(s)), nil
				},
			}, nil
		}
		s, ok := env[n.Name]
		if !ok {
			s = NewSortVar()
			env[n.Name] = s
		}
		ts := ConvertToSortVars(n.CSort)
		if err := Unify(s, ts); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: s,
			Concretize: func() (logic.Expr, error) {
				return logic.NewSymbol(n.Name, ConvertFromSortVars(s)), nil
			},
		}, nil

	case *logic.Apply:
		funcRes, err := InferSorts(n.Func, env)
		if err != nil {
			return nil, err
		}
		termResults := make([]*InferResult, len(n.Terms))
		termSorts := make([]SortOrVar, len(n.Terms))
		for i, term := range n.Terms {
			res, err := InferSorts(term, env)
			if err != nil {
				return nil, err
			}
			termResults[i] = res
			termSorts[i] = res.Sort
		}
		resultSort := NewSortVar()
		// Unify the function's sort with FunctionSort(arg sorts -> result sort).
		// Mirrors Python type_inference.py lines 180-181:
		//   sorts = terms_s + [SortVar()]
		//   unify(func_s, FunctionSort(*sorts))
		// We use FunctionSortVar to preserve SortVar linkage.
		allSorts := make([]SortOrVar, len(termSorts)+1)
		copy(allSorts, termSorts)
		allSorts[len(termSorts)] = resultSort
		fsv := NewFunctionSortVar(allSorts...)
		if err := Unify(funcRes.Sort, fsv); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: resultSort,
			Concretize: func() (logic.Expr, error) {
				fn, err := funcRes.Concretize()
				if err != nil {
					return nil, err
				}
				terms := make([]logic.Expr, len(termResults))
				for i, tr := range termResults {
					t, err := tr.Concretize()
					if err != nil {
						return nil, err
					}
					terms[i] = t
				}
				return logic.NewApply(fn, terms...)
			},
		}, nil

	case *logic.Eq:
		r1, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		r2, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(r1.Sort, r2.Sort); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				t1, err := r1.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := r2.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewEq(t1, t2)
			},
		}, nil

	case *logic.Ite:
		rCond, err := InferSorts(n.Cond, env)
		if err != nil {
			return nil, err
		}
		rThen, err := InferSorts(n.Then, env)
		if err != nil {
			return nil, err
		}
		rElse, err := InferSorts(n.Else, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(rCond.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		if err := Unify(rThen.Sort, rElse.Sort); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: rThen.Sort,
			Concretize: func() (logic.Expr, error) {
				c, err := rCond.Concretize()
				if err != nil {
					return nil, err
				}
				t, err := rThen.Concretize()
				if err != nil {
					return nil, err
				}
				e, err := rElse.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewIte(c, t, e)
			},
		}, nil

	case *logic.Not:
		r, err := InferSorts(n.Body, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(r.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				b, err := r.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewNot(b)
			},
		}, nil

	case *logic.And:
		results := make([]*InferResult, len(n.Terms))
		for i, term := range n.Terms {
			res, err := InferSorts(term, env)
			if err != nil {
				return nil, err
			}
			if err := Unify(res.Sort, Wrap(logic.Boolean)); err != nil {
				return nil, err
			}
			results[i] = res
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				terms := make([]logic.Expr, len(results))
				for i, r := range results {
					t, err := r.Concretize()
					if err != nil {
						return nil, err
					}
					terms[i] = t
				}
				return logic.NewAnd(terms...)
			},
		}, nil

	case *logic.Or:
		results := make([]*InferResult, len(n.Terms))
		for i, term := range n.Terms {
			res, err := InferSorts(term, env)
			if err != nil {
				return nil, err
			}
			if err := Unify(res.Sort, Wrap(logic.Boolean)); err != nil {
				return nil, err
			}
			results[i] = res
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				terms := make([]logic.Expr, len(results))
				for i, r := range results {
					t, err := r.Concretize()
					if err != nil {
						return nil, err
					}
					terms[i] = t
				}
				return logic.NewOr(terms...)
			},
		}, nil

	case *logic.Implies:
		r1, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		r2, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(r1.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		if err := Unify(r2.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				t1, err := r1.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := r2.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewImplies(t1, t2)
			},
		}, nil

	case *logic.Iff:
		r1, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		r2, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(r1.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		if err := Unify(r2.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				t1, err := r1.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := r2.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewIff(t1, t2)
			},
		}, nil

	case *logic.Globally:
		r, err := InferSorts(n.Body, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(r.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				b, err := r.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewGlobally(n.Environ, b)
			},
		}, nil

	case *logic.Eventually:
		r, err := InferSorts(n.Body, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(r.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				b, err := r.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewEventually(n.Environ, b)
			},
		}, nil

	case *logic.WhenOperator:
		rThen, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		rCond, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(rCond.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: rThen.Sort,
			Concretize: func() (logic.Expr, error) {
				t1, err := rThen.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := rCond.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewWhenOperator(n.Name, t1, t2)
			},
		}, nil

	case *logic.Cond:
		rCond, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		rThen, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := Unify(rCond.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: rThen.Sort,
			Concretize: func() (logic.Expr, error) {
				t1, err := rCond.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := rThen.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewCond(t1, t2)
			},
		}, nil

	case *logic.ForAll:
		envCopy := copyEnv(env)
		// Create fresh sort variables for bound variables. These will be
		// unified with concrete sorts when the body is inferred (e.g.,
		// X:TopSort used in client_concept(X) will unify X's sort var
		// with the client sort).
		boundSortVars := make([]SortOrVar, len(n.Variables))
		for i, v := range n.Variables {
			sv := NewSortVar()
			envCopy[v.Name] = sv
			boundSortVars[i] = sv
		}
		// Also unify with the variable's declared sort if it's not TopSort.
		for i, v := range n.Variables {
			if !logic.IsTopSort(v.VSort) {
				if err := Unify(boundSortVars[i], Wrap(v.VSort)); err != nil {
					return nil, err
				}
			}
		}
		bodyRes, err := InferSorts(n.Body, envCopy)
		if err != nil {
			return nil, err
		}
		if err := Unify(bodyRes.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		// Capture the original variables and their inferred sort vars.
		origVars := n.Variables
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				vars := make([]*logic.Variable, len(origVars))
				for i, v := range origVars {
					// Use the resolved sort from the bound sort variable
					// (which was unified during body inference).
					cs := ConvertFromSortVars(boundSortVars[i])
					vars[i], err = logic.NewVariable(v.Name, cs)
					if err != nil {
						return nil, err
					}
				}
				body, err := bodyRes.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewForAll(vars, body)
			},
		}, nil

	case *logic.Exists:
		envCopy := copyEnv(env)
		boundSortVars := make([]SortOrVar, len(n.Variables))
		for i, v := range n.Variables {
			sv := NewSortVar()
			envCopy[v.Name] = sv
			boundSortVars[i] = sv
		}
		for i, v := range n.Variables {
			if !logic.IsTopSort(v.VSort) {
				if err := Unify(boundSortVars[i], Wrap(v.VSort)); err != nil {
					return nil, err
				}
			}
		}
		bodyRes, err := InferSorts(n.Body, envCopy)
		if err != nil {
			return nil, err
		}
		if err := Unify(bodyRes.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		origVars := n.Variables
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Expr, error) {
				vars := make([]*logic.Variable, len(origVars))
				for i, v := range origVars {
					cs := ConvertFromSortVars(boundSortVars[i])
					vars[i], err = logic.NewVariable(v.Name, cs)
					if err != nil {
						return nil, err
					}
				}
				body, err := bodyRes.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewExists(vars, body)
			},
		}, nil

	case *logic.NamedBinder:
		envCopy := copyEnv(env)
		for _, v := range n.Variables {
			envCopy[v.Name] = NewSortVar()
		}
		varResults := make([]*InferResult, len(n.Variables))
		varSorts := make([]SortOrVar, len(n.Variables))
		for i, v := range n.Variables {
			r, err := InferSorts(v, envCopy)
			if err != nil {
				return nil, err
			}
			varResults[i] = r
			varSorts[i] = r.Sort
		}
		bodyRes, err := InferSorts(n.Body, envCopy)
		if err != nil {
			return nil, err
		}
		var resultSort SortOrVar
		if len(n.Variables) > 0 {
			allSorts := append(varSorts, bodyRes.Sort)
			fsSorts := make([]logic.Sort, len(allSorts))
			for i, sv := range allSorts {
				if c := Unwrap(sv); c != nil {
					fsSorts[i] = c
				} else {
					fsSorts[i] = logic.NewTopSort()
				}
			}
			fs, _ := logic.NewFunctionSort(fsSorts...)
			resultSort = Wrap(fs)
		} else {
			resultSort = bodyRes.Sort
		}
		return &InferResult{
			Sort: resultSort,
			Concretize: func() (logic.Expr, error) {
				vars := make([]*logic.Variable, len(varResults))
				for i, vr := range varResults {
					v, err := vr.Concretize()
					if err != nil {
						return nil, err
					}
					vars[i] = v.(*logic.Variable)
				}
				body, err := bodyRes.Concretize()
				if err != nil {
					return nil, err
				}
				return logic.NewNamedBinder(n.Name, vars, n.Environ, body)
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %T", t)
	}
}

// ConcretizeSorts returns a term obtained from t by replacing TopSorts with
// concrete sorts. If sort is non-nil, the sort of t is unified with sort.
func ConcretizeSorts(t logic.Expr, s logic.Sort) (logic.Expr, error) {
	res, err := InferSorts(t, nil)
	if err != nil {
		return nil, err
	}
	if s != nil {
		if err := Unify(res.Sort, Wrap(s)); err != nil {
			return nil, err
		}
	}
	return res.Concretize()
}

// ConcretizeTerms concretizes sorts across multiple terms using a shared
// unification environment. Variables/constants with the same name across
// terms are unified to the same sort. If sorts is non-nil, each term's
// sort is unified with the corresponding constraint.
// Matches Python type_inference.py concretize_terms.
func ConcretizeTerms(terms []logic.Expr, sorts []logic.Sort) ([]logic.Expr, error) {
	// Build shared env across all terms
	env := make(map[string]SortOrVar)
	for _, t := range terms {
		collectNames(t, env)
	}

	// Infer sorts for each term using the shared env
	results := make([]*InferResult, len(terms))
	for i, t := range terms {
		res, err := InferSorts(t, env)
		if err != nil {
			return nil, err
		}
		results[i] = res
	}

	// Apply sort constraints if provided
	if sorts != nil {
		for i, sort := range sorts {
			if i < len(results) && sort != nil {
				if err := Unify(results[i].Sort, Wrap(sort)); err != nil {
					return nil, err
				}
			}
		}
	}

	// Concretize all terms
	out := make([]logic.Expr, len(terms))
	for i, res := range results {
		expr, err := res.Concretize()
		if err != nil {
			return nil, err
		}
		out[i] = expr
	}
	return out, nil
}

func collectNames(n logic.Expr, env map[string]SortOrVar) {
	switch t := n.(type) {
	case *logic.Variable:
		if _, ok := env[t.Name]; !ok {
			env[t.Name] = NewSortVar()
		}
	case *logic.Symbol:
		if _, ok := env[t.Name]; !ok {
			env[t.Name] = NewSortVar()
		}
	case *logic.Apply:
		// Explicitly walk Func since Children() now returns only Terms.
		collectNames(t.Func, env)
	}
	for _, c := range n.Children() {
		collectNames(c, env)
	}
	// Also collect from Variables in quantifiers/binders
	switch t := n.(type) {
	case *logic.ForAll:
		for _, v := range t.Variables {
			collectNames(v, env)
		}
	case *logic.Exists:
		for _, v := range t.Variables {
			collectNames(v, env)
		}
	case *logic.NamedBinder:
		for _, v := range t.Variables {
			collectNames(v, env)
		}
	case *logic.Lambda:
		for _, v := range t.Variables {
			collectNames(v, env)
		}
	}
}

func copyEnv(env map[string]SortOrVar) map[string]SortOrVar {
	cp := make(map[string]SortOrVar, len(env))
	for k, v := range env {
		cp[k] = v
	}
	return cp
}
