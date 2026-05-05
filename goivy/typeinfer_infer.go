package goivy

import (
	"fmt"
)

// InferResult holds the inferred sort and a closure to concretize the term.
type InferResult struct {
	Sort       SortOrVar
	Concretize func() (Expr, error)
}

// InferSorts performs type inference on a term, returning the inferred sort
// and a closure that concretizes the term.
func InferSorts(t Expr, env map[string]SortOrVar) (*InferResult, error) {
	if env == nil {
		env = make(map[string]SortOrVar)
		collectNames(t, env)
	}

	switch n := t.(type) {
	case *LogicVariable:
		// Check env first - if this variable was bound by a quantifier,
		// the env has a sort var that will be unified with the concrete
		// sort from the body's function applications.
		s, ok := env[n.Name]
		if !ok {
			if IsPolymorphic(n) {
				s = InsertSortVars(n.VSort, map[string]SortOrVar{})
			} else {
				s = NewSortVar()
			}
			env[n.Name] = s
		}
		// Unify env sort var with the variable's declared sort (if concrete).
		if !IsTopSort(n.VSort) {
			ts := ConvertToSortVars(n.VSort)
			if err := TypeInferUnify(s, ts); err != nil {
				return nil, err
			}
		}
		return &InferResult{
			Sort: s,
			Concretize: func() (Expr, error) {
				cs := ConvertFromSortVars(s)
				return NewVariable(n.Name, cs)
			},
		}, nil

	case *Const:
		if IsPolymorphic(n) {
			s := InsertSortVars(n.CSort, map[string]SortOrVar{})
			return &InferResult{
				Sort: s,
				Concretize: func() (Expr, error) {
					return NewConst(n.Name, ConvertFromSortVars(s)), nil
				},
			}, nil
		}
		s, ok := env[n.Name]
		if !ok {
			s = NewSortVar()
			env[n.Name] = s
		}
		ts := ConvertToSortVars(n.CSort)
		if err := TypeInferUnify(s, ts); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: s,
			Concretize: func() (Expr, error) {
				return NewConst(n.Name, ConvertFromSortVars(s)), nil
			},
		}, nil

	case *Apply:
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
		if err := TypeInferUnify(funcRes.Sort, fsv); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: resultSort,
			Concretize: func() (Expr, error) {
				fn, err := funcRes.Concretize()
				if err != nil {
					return nil, err
				}
				terms := make([]Expr, len(termResults))
				for i, tr := range termResults {
					t, err := tr.Concretize()
					if err != nil {
						return nil, err
					}
					terms[i] = t
				}
				return NewApply(fn, terms...)
			},
		}, nil

	case *Eq:
		r1, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		r2, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r1.Sort, r2.Sort); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				t1, err := r1.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := r2.Concretize()
				if err != nil {
					return nil, err
				}
				return NewEq(t1, t2)
			},
		}, nil

	case *LogicIte:
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
		if err := TypeInferUnify(rCond.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		if err := TypeInferUnify(rThen.Sort, rElse.Sort); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: rThen.Sort,
			Concretize: func() (Expr, error) {
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
				return NewIte(c, t, e)
			},
		}, nil

	case *LogicNot:
		r, err := InferSorts(n.Body, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				b, err := r.Concretize()
				if err != nil {
					return nil, err
				}
				return NewNot(b)
			},
		}, nil

	case *LogicAnd:
		results := make([]*InferResult, len(n.Terms))
		for i, term := range n.Terms {
			res, err := InferSorts(term, env)
			if err != nil {
				return nil, err
			}
			if err := TypeInferUnify(res.Sort, Wrap(Boolean)); err != nil {
				return nil, err
			}
			results[i] = res
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				terms := make([]Expr, len(results))
				for i, r := range results {
					t, err := r.Concretize()
					if err != nil {
						return nil, err
					}
					terms[i] = t
				}
				return NewAnd(terms...)
			},
		}, nil

	case *LogicOr:
		results := make([]*InferResult, len(n.Terms))
		for i, term := range n.Terms {
			res, err := InferSorts(term, env)
			if err != nil {
				return nil, err
			}
			if err := TypeInferUnify(res.Sort, Wrap(Boolean)); err != nil {
				return nil, err
			}
			results[i] = res
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				terms := make([]Expr, len(results))
				for i, r := range results {
					t, err := r.Concretize()
					if err != nil {
						return nil, err
					}
					terms[i] = t
				}
				return NewOr(terms...)
			},
		}, nil

	case *LogicImplies:
		r1, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		r2, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r1.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r2.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				t1, err := r1.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := r2.Concretize()
				if err != nil {
					return nil, err
				}
				return NewImplies(t1, t2)
			},
		}, nil

	case *LogicIff:
		r1, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		r2, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r1.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r2.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				t1, err := r1.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := r2.Concretize()
				if err != nil {
					return nil, err
				}
				return NewIff(t1, t2)
			},
		}, nil

	case *LogicGlobally:
		r, err := InferSorts(n.Body, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				b, err := r.Concretize()
				if err != nil {
					return nil, err
				}
				return NewGlobally(n.Environ, b)
			},
		}, nil

	case *LogicEventually:
		r, err := InferSorts(n.Body, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(r.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				b, err := r.Concretize()
				if err != nil {
					return nil, err
				}
				return NewEventually(n.Environ, b)
			},
		}, nil

	case *LogicWhenOperator:
		rThen, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		rCond, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(rCond.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: rThen.Sort,
			Concretize: func() (Expr, error) {
				t1, err := rThen.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := rCond.Concretize()
				if err != nil {
					return nil, err
				}
				return NewWhenOperator(n.Name, t1, t2)
			},
		}, nil

	case *Cond:
		rCond, err := InferSorts(n.T1, env)
		if err != nil {
			return nil, err
		}
		rThen, err := InferSorts(n.T2, env)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(rCond.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		return &InferResult{
			Sort: rThen.Sort,
			Concretize: func() (Expr, error) {
				t1, err := rCond.Concretize()
				if err != nil {
					return nil, err
				}
				t2, err := rThen.Concretize()
				if err != nil {
					return nil, err
				}
				return NewCond(t1, t2)
			},
		}, nil

	case *ForAll:
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
			if !IsTopSort(v.VSort) {
				if err := TypeInferUnify(boundSortVars[i], Wrap(v.VSort)); err != nil {
					return nil, err
				}
			}
		}
		bodyRes, err := InferSorts(n.Body, envCopy)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(bodyRes.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		// Capture the original variables and their inferred sort vars.
		origVars := n.Variables
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				vars := make([]*LogicVariable, len(origVars))
				for i, v := range origVars {
					// Use the resolved sort from the bound sort variable
					// (which was unified during body inference).
					cs := ConvertFromSortVars(boundSortVars[i])
					vars[i], err = NewVariable(v.Name, cs)
					if err != nil {
						return nil, err
					}
				}
				body, err := bodyRes.Concretize()
				if err != nil {
					return nil, err
				}
				return NewForAll(vars, body)
			},
		}, nil

	case *LogicExists:
		envCopy := copyEnv(env)
		boundSortVars := make([]SortOrVar, len(n.Variables))
		for i, v := range n.Variables {
			sv := NewSortVar()
			envCopy[v.Name] = sv
			boundSortVars[i] = sv
		}
		for i, v := range n.Variables {
			if !IsTopSort(v.VSort) {
				if err := TypeInferUnify(boundSortVars[i], Wrap(v.VSort)); err != nil {
					return nil, err
				}
			}
		}
		bodyRes, err := InferSorts(n.Body, envCopy)
		if err != nil {
			return nil, err
		}
		if err := TypeInferUnify(bodyRes.Sort, Wrap(Boolean)); err != nil {
			return nil, err
		}
		origVars := n.Variables
		return &InferResult{
			Sort: Wrap(Boolean),
			Concretize: func() (Expr, error) {
				vars := make([]*LogicVariable, len(origVars))
				for i, v := range origVars {
					cs := ConvertFromSortVars(boundSortVars[i])
					vars[i], err = NewVariable(v.Name, cs)
					if err != nil {
						return nil, err
					}
				}
				body, err := bodyRes.Concretize()
				if err != nil {
					return nil, err
				}
				return NewExists(vars, body)
			},
		}, nil

	case *LogicNamedBinder:
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
			// Python type_inference.py:254-255 returns
			//     FunctionSort(*(vars_s + [body_s]))
			// where vars_s are still SortVars. Python's FunctionSort can hold
			// SortVars directly; Go's logic.FunctionSort cannot, so we use the
			// typeinfer wrapper FunctionSortVar that preserves SortVar linkage.
			// This matches the *logic.Apply arm above (see line 102).
			//
			// Eagerly unwrapping to logic.Sort here breaks invariants like
			//   ($l2s_s P0. active(P0))(P)
			// where P's sort must be inferred from the binder's domain — bound
			// variables' SortVars may still be un-concretized at this point and
			// would silently become TopSort, propagating to the free argument
			// and ultimately panicking in z3bridge.TranslateSort.
			allSorts := append(varSorts, bodyRes.Sort)
			resultSort = NewFunctionSortVar(allSorts...)
		} else {
			resultSort = bodyRes.Sort
		}
		return &InferResult{
			Sort: resultSort,
			Concretize: func() (Expr, error) {
				vars := make([]*LogicVariable, len(varResults))
				for i, vr := range varResults {
					v, err := vr.Concretize()
					if err != nil {
						return nil, err
					}
					vars[i] = v.(*LogicVariable)
				}
				body, err := bodyRes.Concretize()
				if err != nil {
					return nil, err
				}
				return NewNamedBinder(n.Name, vars, n.Environ, body)
			},
		}, nil

	default:
		// Mirror Python type_inference.py:264-269 generic fallback:
		//   elif hasattr(t,'clone'):
		//       xys = [infer_sorts(tt, env) for tt in t.args]
		//       terms_t = [y for x, y in xys]
		//       return TopSort(), lambda: t.clone([x() for x in terms_t])
		//
		// Needed for *logic.Definition (Def lhs rhs) and any other
		// logic node that lacks a specialized case above. Children's
		// sorts are inferred (so free vars get unified through the
		// shared env); the whole node's sort is TopSort.
		children := t.Children()
		childResults := make([]*InferResult, len(children))
		for i, c := range children {
			r, err := InferSorts(c, env)
			if err != nil {
				return nil, err
			}
			childResults[i] = r
		}
		return &InferResult{
			Sort: Wrap(TopS),
			Concretize: func() (Expr, error) {
				newChildren := make([]Node, len(childResults))
				for i, r := range childResults {
					c, err := r.Concretize()
					if err != nil {
						return nil, err
					}
					newChildren[i] = c
				}
				cloned := t.Clone(newChildren)
				e, ok := cloned.(Expr)
				if !ok {
					return nil, fmt.Errorf("clone result is not logic.Expr: %T", cloned)
				}
				return e, nil
			},
		}, nil
	}
}

// ConcretizeSorts returns a term obtained from t by replacing TopSorts with
// concrete sorts. If sort is non-nil, the sort of t is unified with sort.
func ConcretizeSorts(t Expr, s Sort) (Expr, error) {
	res, err := InferSorts(t, nil)
	if err != nil {
		return nil, err
	}
	if s != nil {
		if err := TypeInferUnify(res.Sort, Wrap(s)); err != nil {
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
func ConcretizeTerms(terms []Expr, sorts []Sort) ([]Expr, error) {
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
				if err := TypeInferUnify(results[i].Sort, Wrap(sort)); err != nil {
					return nil, err
				}
			}
		}
	}

	// Concretize all terms
	out := make([]Expr, len(terms))
	for i, res := range results {
		expr, err := res.Concretize()
		if err != nil {
			return nil, err
		}
		out[i] = expr
	}
	return out, nil
}

func collectNames(n Expr, env map[string]SortOrVar) {
	switch t := n.(type) {
	case *LogicVariable:
		if _, ok := env[t.Name]; !ok {
			env[t.Name] = NewSortVar()
		}
	case *Const:
		if _, ok := env[t.Name]; !ok {
			env[t.Name] = NewSortVar()
		}
	case *Apply:
		// Explicitly walk Func since Children() now returns only Terms.
		collectNames(t.Func, env)
	}
	for _, c := range n.Children() {
		collectNames(c, env)
	}
	// Also collect from Variables in quantifiers/binders
	switch t := n.(type) {
	case *ForAll:
		for _, v := range t.Variables {
			collectNames(v, env)
		}
	case *LogicExists:
		for _, v := range t.Variables {
			collectNames(v, env)
		}
	case *LogicNamedBinder:
		for _, v := range t.Variables {
			collectNames(v, env)
		}
	case *Lambda:
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
