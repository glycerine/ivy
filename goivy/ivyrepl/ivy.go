package ivyrepl

import (
	"fmt"

	goivy "github.com/glycerine/ivy/goivy"
	"github.com/glycerine/ivy/goivy/smt"
	"github.com/glycerine/ivy/goivy/zlisp"
)

func installIvy(env *zlisp.Zlisp) {
	addFunctions(env, map[string]zlisp.ZlispUserFunction{
		"ivy_new_module":             ivyNewModule,
		"ivy_new_solver":             ivyNewSolver,
		"ivy_solver_context":         ivySolverContext,
		"ivy_solver_clear":           ivySolverClear,
		"ivy_solver_close":           ivySolverClose,
		"ivy_new_translator":         ivyNewTranslator,
		"ivy_new_interp_translator":  ivyNewInterpolationTranslator,
		"ivy_translator_clear":       ivyTranslatorClear,
		"ivy_translator_close":       ivyTranslatorClose,
		"ivy_parse_formula":          ivyParseFormula,
		"ivy_parse_term":             ivyParseTerm,
		"ivy_node_rep":               ivyNodeRep,
		"ivy_expr_string":            ivyExprString,
		"ivy_expr_canon":             ivyExprCanon,
		"ivy_expr_sexp":              ivyExprSexp,
		"ivy_expr_to_z3":             ivyExprToZ3,
		"ivy_formula_to_z3":          ivyFormulaToZ3,
		"ivy_is_sat":                 ivyIsSat,
		"ivy_implies":                ivyImplies,
		"ivy_solver_z3_implies":      ivySolverZ3Implies,
		"ivy_solver_implies_batch":   ivySolverImpliesBatch,
		"ivy_z3_utils":               ivyNewZ3Utils,
		"ivy_z3_utils_clear":         ivyZ3UtilsClear,
		"ivy_z3_utils_close":         ivyZ3UtilsClose,
		"ivy_z3_utils_to_z3":         ivyZ3UtilsToZ3,
		"ivy_z3_utils_to_z3_expr":    ivyZ3UtilsToZ3Expr,
		"ivy_z3_utils_implies":       ivyZ3UtilsImplies,
		"ivy_z3_utils_implies_batch": ivyZ3UtilsImpliesBatch,
		"ivy_z3_ceil_log2":           ivyZ3CeilLog2,
		"ivy_bin_enc":                ivyBinEnc,
		"ivy_bin_enc_z3":             ivyBinEncZ3,
		"ivy_gebin":                  ivyGebin,
		"ivy_my_eq":                  ivyMyEq,
		"ivy_bool_sort":              ivyBoolSort,
		"ivy_uninterpreted_sort":     ivyUninterpretedSort,
		"ivy_enum_sort":              ivyEnumSort,
		"ivy_function_sort":          ivyFunctionSort,
		"ivy_const":                  ivyConst,
		"ivy_var":                    ivyVar,
		"ivy_apply":                  ivyApply,
		"ivy_true":                   ivyTrue,
		"ivy_false":                  ivyFalse,
		"ivy_eq":                     ivyEq,
		"ivy_not":                    ivyNot,
		"ivy_and":                    ivyAnd,
		"ivy_or":                     ivyOr,
		"ivy_implies_expr":           ivyImpliesExpr,
		"ivy_iff":                    ivyIff,
	})
}

func ivyNewModule(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 0 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	return &IvyModuleRef{Module: goivy.New()}, nil
}

func ivyNewSolver(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) > 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	var mod *goivy.Module
	if len(args) == 1 {
		var err error
		mod, err = asIvyModule(args[0], name)
		if err != nil {
			return zlisp.SexpNull, err
		}
	}
	return &IvySolverRef{Solver: goivy.NewSolver(mod, nil)}, nil
}

func ivySolverContext(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ContextRef{Ctx: solver.Context()}, nil
}

func ivySolverClear(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	solver.Clear()
	return zlisp.SexpNull, nil
}

func ivySolverClose(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return zlisp.SexpNull, solver.Close()
}

func ivyNewTranslator(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := optionalSolver(args, name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyTranslatorRef{Translator: solver.NewTranslator()}, nil
}

func ivyNewInterpolationTranslator(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := optionalSolver(args, name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyTranslatorRef{Translator: solver.NewTranslatorWithInterpolation()}, nil
}

func optionalSolver(args []zlisp.Sexp, name string) (*goivy.Solver, error) {
	if len(args) > 1 {
		return nil, zlisp.WrongNargs
	}
	if len(args) == 0 {
		return goivy.NewSolver(nil, nil), nil
	}
	return asIvySolver(args[0], name)
}

func ivyTranslatorClear(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	tr, err := asIvyTranslator(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	tr.Clear()
	return zlisp.SexpNull, nil
}

func ivyTranslatorClose(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	tr, err := asIvyTranslator(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return zlisp.SexpNull, tr.Close()
}

func ivyParseFormula(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	src, err := asString(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	compiler, version, err := compilerAndVersion(args, 1)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := compiler.ToFormulaV(src, version)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: expr}, nil
}

func ivyParseTerm(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 1 || len(args) > 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	src, err := asString(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	compiler, version, err := compilerAndVersion(args, 1)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := compiler.ToTermV(src, version)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: expr}, nil
}

func compilerAndVersion(args []zlisp.Sexp, index int) (*goivy.Compiler, goivy.Version, error) {
	mod := goivy.New()
	if len(args) > index {
		if ref, ok := args[index].(*IvyModuleRef); ok {
			if ref.Module == nil {
				return nil, goivy.Version{}, fmt.Errorf("module: expected ivy-module, got nil")
			}
			mod = ref.Module
			index++
		}
	}
	if len(args) > index+1 {
		return nil, goivy.Version{}, zlisp.WrongNargs
	}
	version, err := optionalVersion(args, index)
	if err != nil {
		return nil, goivy.Version{}, err
	}
	return goivy.NewFromModule(mod), version, nil
}

func ivyNodeRep(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	expr, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(goivy.NodeRep(expr)), nil
}

func ivyExprString(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	expr, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(expr.String()), nil
}

func ivyExprCanon(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	expr, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(string(expr.Canon())), nil
}

func ivyExprSexp(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	expr, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(string(expr.Sexp())), nil
}

func ivyExprToZ3(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	tr, err := asIvyTranslator(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	zexpr, err := tr.Translate(expr)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: zexpr}, nil
}

func ivyFormulaToZ3(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	zexpr, err := solver.FormulaToZ3(expr)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: zexpr}, nil
}

func ivyIsSat(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	ok, err := solver.IsSat(expr)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool(ok), nil
}

func ivyImplies(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	f1, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	f2, err := asIvyExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	ok, err := solver.Implies(f1, f2)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool(ok), nil
}

func ivySolverZ3Implies(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 3 || len(args) > 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	f1, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	f2, err := asIvyExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	timeout := false
	if len(args) == 4 {
		timeout, err = asBool(args[3], name)
		if err != nil {
			return zlisp.SexpNull, err
		}
	}
	ok, err := solver.Z3Implies(f1, f2, timeout)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool(ok), nil
}

func ivySolverImpliesBatch(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 3 || len(args) > 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asIvySolver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	premise, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	formulas, err := asIvyExprSlice(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	timeout := false
	if len(args) == 4 {
		timeout, err = asBool(args[3], name)
		if err != nil {
			return zlisp.SexpNull, err
		}
	}
	result, err := solver.ImpliesBatch(premise, formulas, timeout)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapBools(env, result), nil
}

func ivyNewZ3Utils(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 0 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	return &IvyZ3UtilsRef{Utils: goivy.NewZ3Utils()}, nil
}

func ivyZ3UtilsClear(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	utils, err := asZ3Utils(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	utils.Clear()
	return zlisp.SexpNull, nil
}

func ivyZ3UtilsClose(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	utils, err := asZ3Utils(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return zlisp.SexpNull, utils.Close()
}

func ivyZ3UtilsToZ3(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	utils, err := asZ3Utils(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	obj, err := utils.ToZ3(expr)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapZ3Object(name, obj)
}

func ivyZ3UtilsToZ3Expr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	utils, err := asZ3Utils(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	zexpr, err := utils.ToZ3Expr(expr)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: zexpr}, nil
}

func wrapZ3Object(name string, obj any) (zlisp.Sexp, error) {
	switch v := obj.(type) {
	case smt.Z3Expr:
		return &Z3ExprRef{Expr: v}, nil
	case smt.Z3Sort:
		return &Z3SortRef{Sort: v}, nil
	case smt.FuncDecl:
		return &Z3FuncDeclRef{Decl: v}, nil
	default:
		return zlisp.SexpNull, fmt.Errorf("%s: unsupported z3_utils result type %T", name, obj)
	}
}

func ivyZ3UtilsImplies(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 3 || len(args) > 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	utils, err := asZ3Utils(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	f1, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	f2, err := asIvyExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	timeout := false
	if len(args) == 4 {
		timeout, err = asBool(args[3], name)
		if err != nil {
			return zlisp.SexpNull, err
		}
	}
	ok, err := utils.Z3Implies(f1, f2, timeout)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool(ok), nil
}

func ivyZ3UtilsImpliesBatch(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 3 || len(args) > 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	utils, err := asZ3Utils(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	premise, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	formulas, err := asIvyExprSlice(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	timeout := false
	if len(args) == 4 {
		timeout, err = asBool(args[3], name)
		if err != nil {
			return zlisp.SexpNull, err
		}
	}
	result, err := utils.Z3ImpliesBatch(premise, formulas, timeout)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapBools(env, result), nil
}

func ivyZ3CeilLog2(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	n, err := asInt(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxInt(int64(goivy.Z3CeilLog2(n))), nil
}

func ivyBinEnc(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	m, err := asInt(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	n, err := asInt(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapBools(env, goivy.BinEnc(m, n)), nil
}

func ivyBinEncZ3(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	m, err := asInt(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	n, err := asInt(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapExprs(env, goivy.BinEncZ3(ctx, m, n)), nil
}

func ivyGebin(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	bits, err := asExprSlice(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	n, err := asInt(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: goivy.Gebin(ctx, bits, n)}, nil
}

func ivyMyEq(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	x, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	y, err := asExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: goivy.MyEq(ctx, x, y)}, nil
}

func ivyBoolSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 0 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	return &IvySortRef{Sort: goivy.Boolean}, nil
}

func ivyUninterpretedSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	sortName, err := asString(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvySortRef{Sort: &goivy.UninterpretedSort{Name: sortName}}, nil
}

func ivyEnumSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	sortName, err := asString(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	extension, err := asStringSlice(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvySortRef{Sort: &goivy.LogicEnumeratedSort{Name: sortName, Extension: extension}}, nil
}

func ivyFunctionSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	sorts, err := ivySortArgs(args, name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvySortRef{Sort: sort}, nil
}

func ivyConst(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	symName, err := asString(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, err := asIvySort(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: goivy.NewConst(symName, sort)}, nil
}

func ivyVar(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	varName, err := asString(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, err := asIvySort(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	v, err := goivy.NewVariable(varName, sort)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: v}, nil
}

func ivyApply(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	fn, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	terms, err := ivyExprArgs(args[1:], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	app, err := goivy.NewApply(fn, terms...)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: app}, nil
}

func ivyTrue(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 0 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	return &IvyExprRef{Expr: goivy.True}, nil
}

func ivyFalse(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 0 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	return &IvyExprRef{Expr: goivy.False}, nil
}

func ivyEq(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	t1, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	t2, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	eq, err := goivy.NewEq(t1, t2)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: eq}, nil
}

func ivyNot(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	body, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	not, err := goivy.NewNot(body)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: not}, nil
}

func ivyAnd(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	terms, err := ivyExprArgs(args, name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	and, err := goivy.NewAnd(terms...)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: and}, nil
}

func ivyOr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	terms, err := ivyExprArgs(args, name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	or, err := goivy.NewOr(terms...)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: or}, nil
}

func ivyImpliesExpr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	t1, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	t2, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	imp, err := goivy.NewImplies(t1, t2)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: imp}, nil
}

func ivyIff(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	t1, err := asIvyExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	t2, err := asIvyExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	iff, err := goivy.NewIff(t1, t2)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &IvyExprRef{Expr: iff}, nil
}

func ivyExprArgs(args []zlisp.Sexp, name string) ([]goivy.Expr, error) {
	if len(args) == 1 {
		if _, ok := args[0].(*zlisp.SexpArray); ok {
			return asIvyExprSlice(args[0], name)
		}
	}
	out := make([]goivy.Expr, len(args))
	for i, arg := range args {
		expr, err := asIvyExpr(arg, name)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", name, i, err)
		}
		out[i] = expr
	}
	return out, nil
}

func ivySortArgs(args []zlisp.Sexp, name string) ([]goivy.Sort, error) {
	if len(args) == 1 {
		if _, ok := args[0].(*zlisp.SexpArray); ok {
			return asIvySortSlice(args[0], name)
		}
	}
	out := make([]goivy.Sort, len(args))
	for i, arg := range args {
		sort, err := asIvySort(arg, name)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", name, i, err)
		}
		out[i] = sort
	}
	return out, nil
}
