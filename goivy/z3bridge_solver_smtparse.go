package goivy

import "github.com/glycerine/ivy/goivy/smt"

// ParseSMTLIB2Assertion parses an SMT-LIB2 assertion using this solver's
// signature as the declaration environment. This mirrors the generated C++
// ivy_z3_gen::parse_expr path, where sorts and declarations are preloaded into
// Z3_parse_smtlib2_string.
func (s *Solver) ParseSMTLIB2Assertion(input string) (smt.Z3Expr, error) {
	sortDecls, funcDecls, err := s.smtlibParseDecls()
	if err != nil {
		return smt.Z3Expr{}, err
	}
	return s.tr.Ctx.ParseSMTLIB2String(input, sortDecls, funcDecls)
}

func (s *Solver) smtlibParseDecls() ([]smt.SMTLIBSortDecl, []smt.SMTLIBFuncDecl, error) {
	if s == nil || s.sig == nil {
		return nil, nil, nil
	}
	sortDecls, enumDecls, err := s.smtlibParseSortDecls()
	if err != nil {
		return nil, nil, err
	}
	funcDecls, err := s.smtlibParseFuncDecls()
	if err != nil {
		return nil, nil, err
	}
	funcDecls = append(enumDecls, funcDecls...)
	return sortDecls, funcDecls, nil
}

func (s *Solver) smtlibParseSortDecls() ([]smt.SMTLIBSortDecl, []smt.SMTLIBFuncDecl, error) {
	type namedSort struct {
		name string
		sort Sort
	}
	var sorts []namedSort
	seenInput := map[string]bool{}
	if s.mod != nil {
		for _, name := range s.mod.SortOrder {
			if name == "" || seenInput[name] {
				continue
			}
			st, ok := s.sig.Sorts.Get2(name)
			if !ok || st == nil {
				continue
			}
			sorts = append(sorts, namedSort{name: name, sort: st})
			seenInput[name] = true
		}
	}
	for name, st := range s.sig.Sorts.All() {
		if seenInput[name] {
			continue
		}
		if name == "" || name == "bool" || name == "int" || st == nil {
			continue
		}
		sorts = append(sorts, namedSort{name: name, sort: st})
		seenInput[name] = true
	}

	var sortDecls []smt.SMTLIBSortDecl
	var funcDecls []smt.SMTLIBFuncDecl
	seenSort := map[string]bool{"bool": true, "int": true}
	seenFunc := map[string]bool{}
	for _, ns := range sorts {
		zs, err := s.tr.TranslateSort(ns.sort)
		if err != nil {
			return nil, nil, err
		}
		if !seenSort[ns.name] {
			sortDecls = append(sortDecls, smt.SMTLIBSortDecl{Name: ns.name, Sort: zs})
			seenSort[ns.name] = true
		}
		if es, ok := ns.sort.(*LogicEnumeratedSort); ok {
			for _, value := range es.Extension {
				if value == "" || seenFunc[value] {
					continue
				}
				key := NodeKey(value + ":" + string(es.Sexp()))
				expr, ok := s.tr.cache.consts[key]
				if !ok {
					continue
				}
				funcDecls = append(funcDecls, smt.SMTLIBFuncDecl{Name: value, Decl: expr.Decl()})
				seenFunc[value] = true
			}
		}
	}
	return sortDecls, funcDecls, nil
}

func (s *Solver) smtlibParseFuncDecls() ([]smt.SMTLIBFuncDecl, error) {
	var syms []*Const
	for _, sym := range s.sig.AllSymbols() {
		if sym == nil || sym.Name == "" || s.sig.Constructors[sym.Name] {
			continue
		}
		syms = append(syms, sym)
	}

	seen := map[string]bool{}
	var out []smt.SMTLIBFuncDecl
	bfeCheck := func(c *Const) bool { return s.bfeToZ3(c) != nil }
	for _, sym := range syms {
		z3name, err := SolverName(sym, s.sig, bfeCheck)
		if err != nil {
			z3name = sym.Name
		}
		if z3name == "" || seen[z3name] {
			continue
		}
		fd, err := s.smtlibFuncDecl(z3name, sym.CSort)
		if err != nil {
			return nil, err
		}
		out = append(out, smt.SMTLIBFuncDecl{Name: z3name, Decl: fd})
		seen[z3name] = true
	}
	return out, nil
}

func (s *Solver) smtlibFuncDecl(name string, ivySort Sort) (smt.FuncDecl, error) {
	var domain []Sort
	rangeSort := ivySort
	if fs, ok := ivySort.(*LogicFunctionSort); ok {
		domain = fs.Domain()
		rangeSort = fs.Range()
	}
	zDomain := make([]smt.Z3Sort, len(domain))
	for i, d := range domain {
		zs, err := s.tr.TranslateSort(d)
		if err != nil {
			return smt.FuncDecl{}, err
		}
		zDomain[i] = zs
	}
	zRange, err := s.tr.TranslateSort(rangeSort)
	if err != nil {
		return smt.FuncDecl{}, err
	}
	return s.tr.Ctx.Function(name, zDomain, zRange), nil
}
