package ivyrepl

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/smt"
	"github.com/glycerine/ivy/goivy/zlisp"
)

func installZ3(env *zlisp.Zlisp) {
	addFunctions(env, map[string]zlisp.ZlispUserFunction{
		"z3_new_context":          z3NewContext,
		"z3_new_interp_context":   z3NewInterpolationContext,
		"z3_close":                z3Close,
		"z3_interrupt":            z3Interrupt,
		"z3_bool_sort":            z3BoolSort,
		"z3_int_sort":             z3IntSort,
		"z3_real_sort":            z3RealSort,
		"z3_string_sort":          z3StringSort,
		"z3_bv_sort":              z3BvSort,
		"z3_uninterpreted_sort":   z3UninterpretedSort,
		"z3_array_sort":           z3ArraySort,
		"z3_array_domain":         z3ArrayDomain,
		"z3_array_range":          z3ArrayRange,
		"z3_is_bv_sort":           z3IsBvSort,
		"z3_bv_sort_size":         z3BvSortSize,
		"z3_const":                z3Const,
		"z3_bool":                 z3BoolVal,
		"z3_int":                  z3IntVal,
		"z3_string":               z3StringVal,
		"z3_bv":                   z3BvVal,
		"z3_enum_sort":            z3EnumSort,
		"z3_function":             z3Function,
		"z3_apply":                z3Apply,
		"z3_func_as_expr":         z3FuncAsExpr,
		"z3_not":                  z3Not,
		"z3_and":                  z3And,
		"z3_and_app":              z3AndApp,
		"z3_or":                   z3Or,
		"z3_implies":              z3Implies,
		"z3_iff":                  z3Iff,
		"z3_eq":                   z3Eq,
		"z3_ite":                  z3Ite,
		"z3_add":                  z3Add,
		"z3_sub":                  z3Sub,
		"z3_mul":                  z3Mul,
		"z3_div":                  z3Div,
		"z3_gt":                   z3Gt,
		"z3_lt":                   z3Lt,
		"z3_ge":                   z3Ge,
		"z3_le":                   z3Le,
		"z3_bv_and":               z3BvAnd,
		"z3_bv_or":                z3BvOr,
		"z3_bv_not":               z3BvNot,
		"z3_bv_neg":               z3BvNeg,
		"z3_bv_add":               z3BvAdd,
		"z3_bv_sub":               z3BvSub,
		"z3_bv_mul":               z3BvMul,
		"z3_bv_udiv":              z3BvUdiv,
		"z3_bv_sdiv":              z3BvSdiv,
		"z3_bv_shl":               z3BvShl,
		"z3_bv_lshr":              z3BvLshr,
		"z3_bv_ashr":              z3BvAshr,
		"z3_bv_xor":               z3BvXor,
		"z3_concat":               z3Concat,
		"z3_extract":              z3Extract,
		"z3_bv2int":               z3Bv2Int,
		"z3_int2bv":               z3Int2Bv,
		"z3_bv_ult":               z3BvUlt,
		"z3_bv_ule":               z3BvUle,
		"z3_bv_ugt":               z3BvUgt,
		"z3_bv_uge":               z3BvUge,
		"z3_is_bv_expr":           z3IsBvExpr,
		"z3_forall":               z3ForAll,
		"z3_exists":               z3Exists,
		"z3_select":               z3Select,
		"z3_store":                z3Store,
		"z3_const_array":          z3ConstArray,
		"z3_new_solver":           z3NewSolver,
		"z3_new_solver_for_logic": z3NewSolverForLogic,
		"z3_assert":               z3Assert,
		"z3_check":                z3Check,
		"z3_push":                 z3Push,
		"z3_pop":                  z3Pop,
		"z3_reset":                z3Reset,
		"z3_solver_string":        z3SolverString,
		"z3_solver_smt2":          z3SolverSMT2,
		"z3_reason_unknown":       z3ReasonUnknown,
		"z3_set_param":            z3SetParam,
		"z3_check_assumptions":    z3CheckAssumptions,
		"z3_unsat_core":           z3UnsatCore,
		"z3_canon_assertions":     z3CanonAssertions,
		"z3_model":                z3Model,
		"z3_model_eval":           z3ModelEval,
		"z3_model_sorts":          z3ModelSorts,
		"z3_model_universe":       z3ModelUniverse,
		"z3_model_string":         z3ModelString,
		"z3_expr_string":          z3ExprString,
		"z3_expr_equal":           z3ExprEqual,
		"z3_expr_true?":           z3ExprIsTrue,
		"z3_expr_false?":          z3ExprIsFalse,
		"z3_substitute":           z3Substitute,
		"z3_mk_interpolant":       z3MkInterpolant,
		"z3_compute_interpolant":  z3ComputeInterpolant,
	})
}

func z3NewContext(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 0 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	return &Z3ContextRef{Ctx: smt.NewZ3Context()}, nil
}

func z3NewInterpolationContext(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 0 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	return &Z3ContextRef{Ctx: smt.NewInterpolationZ3Context()}, nil
}

func z3Close(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return zlisp.SexpNull, ctx.Close()
}

func z3Interrupt(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	ctx.Interrupt()
	return zlisp.SexpNull, nil
}

func z3BoolSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxSort(name, args, func(ctx *smt.Z3Context) smt.Z3Sort { return ctx.BoolSort() })
}

func z3IntSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxSort(name, args, func(ctx *smt.Z3Context) smt.Z3Sort { return ctx.IntSort() })
}

func z3RealSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxSort(name, args, func(ctx *smt.Z3Context) smt.Z3Sort { return ctx.RealSort() })
}

func z3StringSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxSort(name, args, func(ctx *smt.Z3Context) smt.Z3Sort { return ctx.StringSort() })
}

func z3BvSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	width, err := asInt(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SortRef{Sort: ctx.BvSort(width)}, nil
}

func z3UninterpretedSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sortName, err := asString(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SortRef{Sort: ctx.UninterpretedSort(sortName)}, nil
}

func z3ArraySort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	dom, err := asSort(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	rng, err := asSort(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SortRef{Sort: ctx.ArraySort(dom, rng)}, nil
}

func z3ArrayDomain(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	sort, err := singleSortArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SortRef{Sort: (&sort).ArrayDomain()}, nil
}

func z3ArrayRange(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	sort, err := singleSortArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SortRef{Sort: (&sort).ArrayRange()}, nil
}

func z3IsBvSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, err := asSort(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool(ctx.IsBvSort(sort)), nil
}

func z3BvSortSize(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, err := asSort(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxInt(int64(ctx.BvSortSize(sort))), nil
}

func z3Const(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	constName, err := asString(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, err := asSort(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.Const(constName, sort)}, nil
}

func z3BoolVal(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	val, err := asBool(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.BoolVal(val)}, nil
}

func z3IntVal(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	val, err := asInt64(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.IntVal(val)}, nil
}

func z3StringVal(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	val, err := asString(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.StringVal(val)}, nil
}

func z3BvVal(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	val, err := asInt64(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	width, err := asInt(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.BvVal(val, width)}, nil
}

func z3EnumSort(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sortName, err := asString(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	elements, err := asStringSlice(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, consts := ctx.EnumSort(sortName, elements)
	return sxArray(env, []zlisp.Sexp{&Z3SortRef{Sort: sort}, wrapExprs(env, consts)}), nil
}

func z3Function(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	funcName, err := asString(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	domain, err := asSortSlice(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	rng, err := asSort(args[3], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3FuncDeclRef{Decl: ctx.Function(funcName, domain, rng)}, nil
}

func z3Apply(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) < 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	decl, err := asFuncDecl(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	exprs, err := exprArgs(name, args[1:])
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: (&decl).Apply(exprs...)}, nil
}

func z3FuncAsExpr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	decl, err := asFuncDecl(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: (&decl).AsExpr()}, nil
}

func z3Not(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxExpr(name, args, func(ctx *smt.Z3Context, a smt.Z3Expr) smt.Z3Expr { return ctx.Not(a) })
}

func z3And(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return variadicCtxExpr(name, args, func(ctx *smt.Z3Context, xs []smt.Z3Expr) smt.Z3Expr { return ctx.And(xs...) })
}

func z3AndApp(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return variadicCtxExpr(name, args, func(ctx *smt.Z3Context, xs []smt.Z3Expr) smt.Z3Expr { return ctx.AndApp(xs...) })
}

func z3Or(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return variadicCtxExpr(name, args, func(ctx *smt.Z3Context, xs []smt.Z3Expr) smt.Z3Expr { return ctx.Or(xs...) })
}

func z3Implies(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Implies(a, b) })
}

func z3Iff(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Iff(a, b) })
}

func z3Eq(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Eq(a, b) })
}

func z3Ite(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	cond, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	thenExpr, err := asExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	elseExpr, err := asExpr(args[3], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.Ite(cond, thenExpr, elseExpr)}, nil
}

func z3Add(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Add(a, b) })
}

func z3Sub(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Sub(a, b) })
}

func z3Mul(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Mul(a, b) })
}

func z3Div(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Div(a, b) })
}

func z3Gt(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Gt(a, b) })
}

func z3Lt(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Lt(a, b) })
}

func z3Ge(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Ge(a, b) })
}

func z3Le(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Le(a, b) })
}

func z3BvAnd(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvAnd(a, b) })
}

func z3BvOr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvOr(a, b) })
}

func z3BvNot(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxExpr(name, args, func(ctx *smt.Z3Context, a smt.Z3Expr) smt.Z3Expr { return ctx.BvNot(a) })
}

func z3BvNeg(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxExpr(name, args, func(ctx *smt.Z3Context, a smt.Z3Expr) smt.Z3Expr { return ctx.BvNeg(a) })
}

func z3BvAdd(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvAdd(a, b) })
}

func z3BvSub(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvSub(a, b) })
}

func z3BvMul(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvMul(a, b) })
}

func z3BvUdiv(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvUdiv(a, b) })
}

func z3BvSdiv(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvSdiv(a, b) })
}

func z3BvShl(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvShl(a, b) })
}

func z3BvLshr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvLshr(a, b) })
}

func z3BvAshr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvAshr(a, b) })
}

func z3BvXor(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvXor(a, b) })
}

func z3Concat(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Concat(a, b) })
}

func z3Extract(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	hi, err := asInt(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	lo, err := asInt(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[3], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.Extract(hi, lo, expr)}, nil
}

func z3Bv2Int(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	signed, err := asBool(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.Bv2Int(expr, signed)}, nil
}

func z3Int2Bv(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	width, err := asInt(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.Int2Bv(width, expr)}, nil
}

func z3BvUlt(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvUlt(a, b) })
}

func z3BvUle(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvUle(a, b) })
}

func z3BvUgt(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvUgt(a, b) })
}

func z3BvUge(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.BvUge(a, b) })
}

func z3IsBvExpr(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool(ctx.IsBvExpr(expr)), nil
}

func z3ForAll(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return quantifier(name, args, func(ctx *smt.Z3Context, bound []smt.Z3Expr, body smt.Z3Expr) smt.Z3Expr {
		return ctx.ForAll(bound, body)
	})
}

func z3Exists(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return quantifier(name, args, func(ctx *smt.Z3Context, bound []smt.Z3Expr, body smt.Z3Expr) smt.Z3Expr {
		return ctx.Exists(bound, body)
	})
}

func z3Select(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return binaryCtxExpr(name, args, func(ctx *smt.Z3Context, a, b smt.Z3Expr) smt.Z3Expr { return ctx.Select(a, b) })
}

func z3Store(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	array, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	index, err := asExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	value, err := asExpr(args[3], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.Store(array, index, value)}, nil
}

func z3ConstArray(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	domain, err := asSort(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	value, err := asExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: ctx.ConstArray(domain, value)}, nil
}

func z3NewSolver(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SolverRef{Solver: ctx.NewZ3Solver()}, nil
}

func z3NewSolverForLogic(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	logic, err := asString(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SolverRef{Solver: smt.NewZ3SolverForLogic(ctx, logic)}, nil
}

func z3Assert(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asZ3Solver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	solver.Assert(expr)
	return zlisp.SexpNull, nil
}

func z3Check(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asZ3Solver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(solver.Check().String()), nil
}

func z3Push(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	solver.Push()
	return zlisp.SexpNull, nil
}

func z3Pop(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	solver.Pop()
	return zlisp.SexpNull, nil
}

func z3Reset(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	solver.Reset()
	return zlisp.SexpNull, nil
}

func z3SolverString(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(solver.String()), nil
}

func z3SolverSMT2(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(solver.ToSMT2()), nil
}

func z3ReasonUnknown(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(solver.ReasonUnknown()), nil
}

func z3SetParam(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asZ3Solver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	key, err := asString(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	value, err := asString(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	solver.SetParam(key, value)
	return zlisp.SexpNull, nil
}

func z3CheckAssumptions(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	solver, err := asZ3Solver(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	assumptions, err := asExprSlice(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(solver.CheckAssumptions(assumptions).String()), nil
}

func z3UnsatCore(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapExprs(env, solver.UnsatCore()), nil
}

func z3CanonAssertions(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(solver.CanonZ3Assertions()), nil
}

func z3Model(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	solver, err := singleSolverArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	model := solver.Model()
	if model == nil {
		return zlisp.SexpNull, nil
	}
	return &Z3ModelRef{Model: model}, nil
}

func z3ModelEval(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	model, err := asModel(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	completion, err := asBool(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	value, ok := model.Eval(expr, completion)
	return sxArray(env, []zlisp.Sexp{&Z3ExprRef{Expr: value}, sxBool(ok)}), nil
}

func z3ModelSorts(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	model, err := asModel(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapSorts(env, model.Sorts()), nil
}

func z3ModelUniverse(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	model, err := asModel(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	sort, err := asSort(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapExprs(env, model.SortUniverse(sort)), nil
}

func z3ModelString(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	model, err := asModel(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr(model.String()), nil
}

func z3ExprString(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	expr, err := singleExprArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxStr((&expr).String()), nil
}

func z3ExprEqual(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	a, err := asExpr(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	b, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool((&a).Equal(b)), nil
}

func z3ExprIsTrue(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	expr, err := singleExprArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool((&expr).IsTrue()), nil
}

func z3ExprIsFalse(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	expr, err := singleExprArg(name, args)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return sxBool((&expr).IsFalse()), nil
}

func z3Substitute(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 4 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	from, err := asExprSlice(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	to, err := asExprSlice(args[3], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	if len(from) != len(to) {
		return zlisp.SexpNull, fmt.Errorf("%s: from and to arrays must have the same length", name)
	}
	return &Z3ExprRef{Expr: ctx.Substitute(expr, from, to)}, nil
}

func z3MkInterpolant(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	return unaryCtxExpr(name, args, func(ctx *smt.Z3Context, a smt.Z3Expr) smt.Z3Expr { return ctx.MkInterpolant(a) })
}

func z3ComputeInterpolant(env *zlisp.Zlisp, name string, args []zlisp.Sexp) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	pattern, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	exprs, err := ctx.ComputeInterpolant(pattern)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return wrapExprs(env, exprs), nil
}

func unaryCtxSort(name string, args []zlisp.Sexp, fn func(*smt.Z3Context) smt.Z3Sort) (zlisp.Sexp, error) {
	if len(args) != 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3SortRef{Sort: fn(ctx)}, nil
}

func unaryCtxExpr(name string, args []zlisp.Sexp, fn func(*smt.Z3Context, smt.Z3Expr) smt.Z3Expr) (zlisp.Sexp, error) {
	if len(args) != 2 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	expr, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: fn(ctx, expr)}, nil
}

func binaryCtxExpr(name string, args []zlisp.Sexp, fn func(*smt.Z3Context, smt.Z3Expr, smt.Z3Expr) smt.Z3Expr) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	a, err := asExpr(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	b, err := asExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: fn(ctx, a, b)}, nil
}

func variadicCtxExpr(name string, args []zlisp.Sexp, fn func(*smt.Z3Context, []smt.Z3Expr) smt.Z3Expr) (zlisp.Sexp, error) {
	if len(args) < 1 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	exprs, err := exprArgs(name, args[1:])
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: fn(ctx, exprs)}, nil
}

func exprArgs(name string, args []zlisp.Sexp) ([]smt.Z3Expr, error) {
	exprs := make([]smt.Z3Expr, len(args))
	for i, arg := range args {
		expr, err := asExpr(arg, name)
		if err != nil {
			return nil, fmt.Errorf("%s arg %d: %w", name, i, err)
		}
		exprs[i] = expr
	}
	return exprs, nil
}

func singleSortArg(name string, args []zlisp.Sexp) (smt.Z3Sort, error) {
	if len(args) != 1 {
		return smt.Z3Sort{}, zlisp.WrongNargs
	}
	return asSort(args[0], name)
}

func singleExprArg(name string, args []zlisp.Sexp) (smt.Z3Expr, error) {
	if len(args) != 1 {
		return smt.Z3Expr{}, zlisp.WrongNargs
	}
	return asExpr(args[0], name)
}

func singleSolverArg(name string, args []zlisp.Sexp) (*smt.Z3Solver, error) {
	if len(args) != 1 {
		return nil, zlisp.WrongNargs
	}
	return asZ3Solver(args[0], name)
}

func quantifier(name string, args []zlisp.Sexp, fn func(*smt.Z3Context, []smt.Z3Expr, smt.Z3Expr) smt.Z3Expr) (zlisp.Sexp, error) {
	if len(args) != 3 {
		return zlisp.SexpNull, zlisp.WrongNargs
	}
	ctx, err := asContext(args[0], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	bound, err := asExprSlice(args[1], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	body, err := asExpr(args[2], name)
	if err != nil {
		return zlisp.SexpNull, err
	}
	return &Z3ExprRef{Expr: fn(ctx, bound, body)}, nil
}
