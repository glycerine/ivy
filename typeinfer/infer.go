package typeinfer

import (
	"fmt"

	"github.com/glycerine/goivy/logic"
)

// InferResult holds the inferred sort and a closure to concretize the term.
type InferResult struct {
	Sort       SortOrVar
	Concretize func() (logic.Node, error)
}

// InferSorts performs type inference on a term, returning the inferred sort
// and a closure that concretizes the term.
func InferSorts(t logic.Node, env map[string]SortOrVar) (*InferResult, error) {
	if env == nil {
		env = make(map[string]SortOrVar)
		collectNames(t, env)
	}

	switch n := t.(type) {
	case *logic.Var:
		if logic.IsPolymorphic(n) {
			s := InsertSortVars(n.VSort, map[string]SortOrVar{})
			return &InferResult{
				Sort: s,
				Concretize: func() (logic.Node, error) {
					cs := ConvertFromSortVars(s)
					return logic.NewVar(n.Name, cs)
				},
			}, nil
		}
		s, ok := env[n.Name]
		if !ok {
			s = NewSortVar()
			env[n.Name] = s
		}
		ts := ConvertToSortVars(n.VSort)
		if err := Unify(s, ts); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: s,
			Concretize: func() (logic.Node, error) {
				cs := ConvertFromSortVars(s)
				return logic.NewVar(n.Name, cs)
			},
		}, nil

	case *logic.Const:
		if logic.IsPolymorphic(n) {
			s := InsertSortVars(n.CSort, map[string]SortOrVar{})
			return &InferResult{
				Sort: s,
				Concretize: func() (logic.Node, error) {
					return logic.NewConst(n.Name, ConvertFromSortVars(s)), nil
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
			Concretize: func() (logic.Node, error) {
				return logic.NewConst(n.Name, ConvertFromSortVars(s)), nil
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
		allSorts := append(termSorts, SortOrVar(resultSort))
		// Build a FunctionSort from the argument sorts + result sort
		fsSorts := make([]logic.Sort, len(allSorts))
		for i, sv := range allSorts {
			if c := Unwrap(sv); c != nil {
				fsSorts[i] = c
			} else {
				fsSorts[i] = logic.NewTopSort()
			}
		}
		fsSort, _ := logic.NewFunctionSort(fsSorts...)
		if err := Unify(funcRes.Sort, Wrap(fsSort)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: resultSort,
			Concretize: func() (logic.Node, error) {
				fn, err := funcRes.Concretize()
				if err != nil {
					return nil, err
				}
				terms := make([]logic.Node, len(termResults))
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
				terms := make([]logic.Node, len(results))
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
			Concretize: func() (logic.Node, error) {
				terms := make([]logic.Node, len(results))
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
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
			Concretize: func() (logic.Node, error) {
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
		for _, v := range n.Variables {
			envCopy[v.Name] = NewSortVar()
		}
		varResults := make([]*InferResult, len(n.Variables))
		for i, v := range n.Variables {
			r, err := InferSorts(v, envCopy)
			if err != nil {
				return nil, err
			}
			varResults[i] = r
		}
		bodyRes, err := InferSorts(n.Body, envCopy)
		if err != nil {
			return nil, err
		}
		if err := Unify(bodyRes.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Node, error) {
				vars := make([]*logic.Var, len(varResults))
				for i, vr := range varResults {
					v, err := vr.Concretize()
					if err != nil {
						return nil, err
					}
					vars[i] = v.(*logic.Var)
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
		for _, v := range n.Variables {
			envCopy[v.Name] = NewSortVar()
		}
		varResults := make([]*InferResult, len(n.Variables))
		for i, v := range n.Variables {
			r, err := InferSorts(v, envCopy)
			if err != nil {
				return nil, err
			}
			varResults[i] = r
		}
		bodyRes, err := InferSorts(n.Body, envCopy)
		if err != nil {
			return nil, err
		}
		if err := Unify(bodyRes.Sort, Wrap(logic.Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(logic.Boolean),
			Concretize: func() (logic.Node, error) {
				vars := make([]*logic.Var, len(varResults))
				for i, vr := range varResults {
					v, err := vr.Concretize()
					if err != nil {
						return nil, err
					}
					vars[i] = v.(*logic.Var)
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
			Concretize: func() (logic.Node, error) {
				vars := make([]*logic.Var, len(varResults))
				for i, vr := range varResults {
					v, err := vr.Concretize()
					if err != nil {
						return nil, err
					}
					vars[i] = v.(*logic.Var)
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
func ConcretizeSorts(t logic.Node, s logic.Sort) (logic.Node, error) {
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

func collectNames(n logic.Node, env map[string]SortOrVar) {
	switch t := n.(type) {
	case *logic.Var:
		if _, ok := env[t.Name]; !ok {
			env[t.Name] = NewSortVar()
		}
	case *logic.Const:
		if _, ok := env[t.Name]; !ok {
			env[t.Name] = NewSortVar()
		}
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
