package ivyrepl

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	goivy "github.com/glycerine/ivy/goivy"
	"github.com/glycerine/ivy/goivy/smt"
	"github.com/glycerine/ivy/goivy/zlisp"
)

type Z3ContextRef struct {
	Ctx *smt.Z3Context
}

type Z3SortRef struct {
	Sort smt.Z3Sort
}

type Z3ExprRef struct {
	Expr smt.Z3Expr
}

type Z3FuncDeclRef struct {
	Decl smt.FuncDecl
}

type Z3SolverRef struct {
	Solver *smt.Z3Solver
}

type Z3ModelRef struct {
	Model *smt.Model
}

type IvyModuleRef struct {
	Module *goivy.Module
}

type IvySolverRef struct {
	Solver *goivy.Solver
}

type IvyTranslatorRef struct {
	Translator *goivy.Translator
}

type IvyExprRef struct {
	Expr goivy.Expr
}

type IvySortRef struct {
	Sort goivy.Sort
}

type IvyZ3UtilsRef struct {
	Utils *goivy.Z3Utils
}

func (r *Z3ContextRef) SexpString(ps *zlisp.PrintState) string {
	return "#<z3-context>"
}

func (r *Z3ContextRef) Type() *zlisp.RegisteredType { return replType("z3-context") }

func (r *Z3SortRef) SexpString(ps *zlisp.PrintState) string {
	if r == nil {
		return "#<z3-sort nil>"
	}
	return fmt.Sprintf("#<z3-sort %s>", (&r.Sort).String())
}

func (r *Z3SortRef) Type() *zlisp.RegisteredType { return replType("z3-sort") }

func (r *Z3ExprRef) SexpString(ps *zlisp.PrintState) string {
	if r == nil {
		return "#<z3-expr nil>"
	}
	return fmt.Sprintf("#<z3-expr %s>", (&r.Expr).String())
}

func (r *Z3ExprRef) Type() *zlisp.RegisteredType { return replType("z3-expr") }

func (r *Z3FuncDeclRef) SexpString(ps *zlisp.PrintState) string {
	return "#<z3-func-decl>"
}

func (r *Z3FuncDeclRef) Type() *zlisp.RegisteredType { return replType("z3-func-decl") }

func (r *Z3SolverRef) SexpString(ps *zlisp.PrintState) string {
	return "#<z3-solver>"
}

func (r *Z3SolverRef) Type() *zlisp.RegisteredType { return replType("z3-solver") }

func (r *Z3ModelRef) SexpString(ps *zlisp.PrintState) string {
	return "#<z3-model>"
}

func (r *Z3ModelRef) Type() *zlisp.RegisteredType { return replType("z3-model") }

func (r *IvyModuleRef) SexpString(ps *zlisp.PrintState) string {
	return "#<ivy-module>"
}

func (r *IvyModuleRef) Type() *zlisp.RegisteredType { return replType("ivy-module") }

func (r *IvySolverRef) SexpString(ps *zlisp.PrintState) string {
	return "#<ivy-solver>"
}

func (r *IvySolverRef) Type() *zlisp.RegisteredType { return replType("ivy-solver") }

func (r *IvyTranslatorRef) SexpString(ps *zlisp.PrintState) string {
	return "#<ivy-translator>"
}

func (r *IvyTranslatorRef) Type() *zlisp.RegisteredType { return replType("ivy-translator") }

func (r *IvyExprRef) SexpString(ps *zlisp.PrintState) string {
	if r == nil || r.Expr == nil {
		return "#<ivy-expr nil>"
	}
	return fmt.Sprintf("#<ivy-expr %s>", goivy.NodeRep(r.Expr))
}

func (r *IvyExprRef) Type() *zlisp.RegisteredType { return replType("ivy-expr") }

func (r *IvySortRef) SexpString(ps *zlisp.PrintState) string {
	if r == nil || r.Sort == nil {
		return "#<ivy-sort nil>"
	}
	return fmt.Sprintf("#<ivy-sort %s>", goivy.IvySortName(r.Sort))
}

func (r *IvySortRef) Type() *zlisp.RegisteredType { return replType("ivy-sort") }

func (r *IvyZ3UtilsRef) SexpString(ps *zlisp.PrintState) string {
	return "#<ivy-z3-utils>"
}

func (r *IvyZ3UtilsRef) Type() *zlisp.RegisteredType { return replType("ivy-z3-utils") }

var registerOnce sync.Once

func registerTypes() {
	registerOnce.Do(func() {
		registerOpaque("z3-context", func() any { return &Z3ContextRef{} })
		registerOpaque("z3-sort", func() any { return &Z3SortRef{} })
		registerOpaque("z3-expr", func() any { return &Z3ExprRef{} })
		registerOpaque("z3-func-decl", func() any { return &Z3FuncDeclRef{} })
		registerOpaque("z3-solver", func() any { return &Z3SolverRef{} })
		registerOpaque("z3-model", func() any { return &Z3ModelRef{} })
		registerOpaque("ivy-module", func() any { return &IvyModuleRef{} })
		registerOpaque("ivy-solver", func() any { return &IvySolverRef{} })
		registerOpaque("ivy-translator", func() any { return &IvyTranslatorRef{} })
		registerOpaque("ivy-expr", func() any { return &IvyExprRef{} })
		registerOpaque("ivy-sort", func() any { return &IvySortRef{} })
		registerOpaque("ivy-z3-utils", func() any { return &IvyZ3UtilsRef{} })
	})
}

func registerOpaque(name string, factory func() any) {
	zlisp.GoStructRegistry.RegisterUserdef(&zlisp.RegisteredType{
		GenDefMap: false,
		Factory: func(env *zlisp.Zlisp, h *zlisp.SexpHash) (interface{}, error) {
			return factory(), nil
		},
	}, false, name)
}

func replType(name string) *zlisp.RegisteredType {
	return zlisp.GoStructRegistry.Lookup(name)
}

func addFunctions(env *zlisp.Zlisp, fns map[string]zlisp.ZlispUserFunction) {
	for name, fn := range fns {
		env.AddFunction(name, fn)
	}
}

func sxStr(s string) zlisp.Sexp {
	return &zlisp.SexpStr{S: s}
}

func sxBool(b bool) zlisp.Sexp {
	return &zlisp.SexpBool{Val: b}
}

func sxInt(i int64) zlisp.Sexp {
	return &zlisp.SexpInt{Val: i}
}

func sxArray(env *zlisp.Zlisp, xs []zlisp.Sexp) zlisp.Sexp {
	return env.NewSexpArray(xs)
}

func asString(arg zlisp.Sexp, name string) (string, error) {
	switch v := arg.(type) {
	case *zlisp.SexpStr:
		return v.S, nil
	case *zlisp.SexpSymbol:
		return v.Name(), nil
	default:
		return "", fmt.Errorf("%s: expected string or symbol, got %T", name, arg)
	}
}

func asInt(arg zlisp.Sexp, name string) (int, error) {
	switch v := arg.(type) {
	case *zlisp.SexpInt:
		return int(v.Val), nil
	case *zlisp.SexpUint64:
		return int(v.Val), nil
	default:
		return 0, fmt.Errorf("%s: expected integer, got %T", name, arg)
	}
}

func asInt64(arg zlisp.Sexp, name string) (int64, error) {
	switch v := arg.(type) {
	case *zlisp.SexpInt:
		return v.Val, nil
	case *zlisp.SexpUint64:
		return int64(v.Val), nil
	default:
		return 0, fmt.Errorf("%s: expected integer, got %T", name, arg)
	}
}

func asBool(arg zlisp.Sexp, name string) (bool, error) {
	v, ok := arg.(*zlisp.SexpBool)
	if !ok {
		return false, fmt.Errorf("%s: expected bool, got %T", name, arg)
	}
	return v.Val, nil
}

func asContext(arg zlisp.Sexp, name string) (*smt.Z3Context, error) {
	ref, ok := arg.(*Z3ContextRef)
	if !ok || ref.Ctx == nil {
		return nil, fmt.Errorf("%s: expected z3-context, got %T", name, arg)
	}
	return ref.Ctx, nil
}

func asSort(arg zlisp.Sexp, name string) (smt.Z3Sort, error) {
	ref, ok := arg.(*Z3SortRef)
	if !ok {
		return smt.Z3Sort{}, fmt.Errorf("%s: expected z3-sort, got %T", name, arg)
	}
	return ref.Sort, nil
}

func asExpr(arg zlisp.Sexp, name string) (smt.Z3Expr, error) {
	ref, ok := arg.(*Z3ExprRef)
	if !ok {
		return smt.Z3Expr{}, fmt.Errorf("%s: expected z3-expr, got %T", name, arg)
	}
	return ref.Expr, nil
}

func asFuncDecl(arg zlisp.Sexp, name string) (smt.FuncDecl, error) {
	ref, ok := arg.(*Z3FuncDeclRef)
	if !ok {
		return smt.FuncDecl{}, fmt.Errorf("%s: expected z3-func-decl, got %T", name, arg)
	}
	return ref.Decl, nil
}

func asZ3Solver(arg zlisp.Sexp, name string) (*smt.Z3Solver, error) {
	ref, ok := arg.(*Z3SolverRef)
	if !ok || ref.Solver == nil {
		return nil, fmt.Errorf("%s: expected z3-solver, got %T", name, arg)
	}
	return ref.Solver, nil
}

func asModel(arg zlisp.Sexp, name string) (*smt.Model, error) {
	ref, ok := arg.(*Z3ModelRef)
	if !ok || ref.Model == nil {
		return nil, fmt.Errorf("%s: expected z3-model, got %T", name, arg)
	}
	return ref.Model, nil
}

func asIvyModule(arg zlisp.Sexp, name string) (*goivy.Module, error) {
	ref, ok := arg.(*IvyModuleRef)
	if !ok || ref.Module == nil {
		return nil, fmt.Errorf("%s: expected ivy-module, got %T", name, arg)
	}
	return ref.Module, nil
}

func asIvySolver(arg zlisp.Sexp, name string) (*goivy.Solver, error) {
	ref, ok := arg.(*IvySolverRef)
	if !ok || ref.Solver == nil {
		return nil, fmt.Errorf("%s: expected ivy-solver, got %T", name, arg)
	}
	return ref.Solver, nil
}

func asIvyTranslator(arg zlisp.Sexp, name string) (*goivy.Translator, error) {
	ref, ok := arg.(*IvyTranslatorRef)
	if !ok || ref.Translator == nil {
		return nil, fmt.Errorf("%s: expected ivy-translator, got %T", name, arg)
	}
	return ref.Translator, nil
}

func asIvyExpr(arg zlisp.Sexp, name string) (goivy.Expr, error) {
	ref, ok := arg.(*IvyExprRef)
	if ok {
		if ref.Expr == nil {
			return nil, fmt.Errorf("%s: expected ivy-expr, got nil", name)
		}
		return ref.Expr, nil
	}
	sortRef, ok := arg.(*IvySortRef)
	if ok {
		if sortRef.Sort == nil {
			return nil, fmt.Errorf("%s: expected ivy-sort, got nil", name)
		}
		return sortRef.Sort, nil
	}
	return nil, fmt.Errorf("%s: expected ivy-expr, got %T", name, arg)
}

func asIvySort(arg zlisp.Sexp, name string) (goivy.Sort, error) {
	ref, ok := arg.(*IvySortRef)
	if !ok || ref.Sort == nil {
		return nil, fmt.Errorf("%s: expected ivy-sort, got %T", name, arg)
	}
	return ref.Sort, nil
}

func asZ3Utils(arg zlisp.Sexp, name string) (*goivy.Z3Utils, error) {
	ref, ok := arg.(*IvyZ3UtilsRef)
	if !ok || ref.Utils == nil {
		return nil, fmt.Errorf("%s: expected ivy-z3-utils, got %T", name, arg)
	}
	return ref.Utils, nil
}

func asExprSlice(arg zlisp.Sexp, name string) ([]smt.Z3Expr, error) {
	arr, ok := arg.(*zlisp.SexpArray)
	if !ok {
		return nil, fmt.Errorf("%s: expected array of z3-expr, got %T", name, arg)
	}
	out := make([]smt.Z3Expr, len(arr.Val))
	for i, item := range arr.Val {
		expr, err := asExpr(item, name)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", name, i, err)
		}
		out[i] = expr
	}
	return out, nil
}

func asSortSlice(arg zlisp.Sexp, name string) ([]smt.Z3Sort, error) {
	arr, ok := arg.(*zlisp.SexpArray)
	if !ok {
		return nil, fmt.Errorf("%s: expected array of z3-sort, got %T", name, arg)
	}
	out := make([]smt.Z3Sort, len(arr.Val))
	for i, item := range arr.Val {
		sort, err := asSort(item, name)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", name, i, err)
		}
		out[i] = sort
	}
	return out, nil
}

func asStringSlice(arg zlisp.Sexp, name string) ([]string, error) {
	arr, ok := arg.(*zlisp.SexpArray)
	if !ok {
		return nil, fmt.Errorf("%s: expected array of strings/symbols, got %T", name, arg)
	}
	out := make([]string, len(arr.Val))
	for i, item := range arr.Val {
		s, err := asString(item, name)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", name, i, err)
		}
		out[i] = s
	}
	return out, nil
}

func asIvyExprSlice(arg zlisp.Sexp, name string) ([]goivy.Expr, error) {
	arr, ok := arg.(*zlisp.SexpArray)
	if !ok {
		return nil, fmt.Errorf("%s: expected array of ivy-expr, got %T", name, arg)
	}
	out := make([]goivy.Expr, len(arr.Val))
	for i, item := range arr.Val {
		expr, err := asIvyExpr(item, name)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", name, i, err)
		}
		out[i] = expr
	}
	return out, nil
}

func asIvySortSlice(arg zlisp.Sexp, name string) ([]goivy.Sort, error) {
	arr, ok := arg.(*zlisp.SexpArray)
	if !ok {
		return nil, fmt.Errorf("%s: expected array of ivy-sort, got %T", name, arg)
	}
	out := make([]goivy.Sort, len(arr.Val))
	for i, item := range arr.Val {
		sort, err := asIvySort(item, name)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", name, i, err)
		}
		out[i] = sort
	}
	return out, nil
}

func wrapExprs(env *zlisp.Zlisp, exprs []smt.Z3Expr) zlisp.Sexp {
	out := make([]zlisp.Sexp, len(exprs))
	for i, expr := range exprs {
		out[i] = &Z3ExprRef{Expr: expr}
	}
	return sxArray(env, out)
}

func wrapSorts(env *zlisp.Zlisp, sorts []smt.Z3Sort) zlisp.Sexp {
	out := make([]zlisp.Sexp, len(sorts))
	for i, sort := range sorts {
		out[i] = &Z3SortRef{Sort: sort}
	}
	return sxArray(env, out)
}

func wrapBools(env *zlisp.Zlisp, bools []bool) zlisp.Sexp {
	out := make([]zlisp.Sexp, len(bools))
	for i, b := range bools {
		out[i] = sxBool(b)
	}
	return sxArray(env, out)
}

func optionalVersion(args []zlisp.Sexp, index int) (goivy.Version, error) {
	if len(args) <= index {
		return goivy.Version{1, 7}, nil
	}
	raw, err := asString(args[index], "version")
	if err != nil {
		return goivy.Version{}, err
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return goivy.Version{}, fmt.Errorf("version: expected major.minor, got %q", raw)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return goivy.Version{}, fmt.Errorf("version: bad major in %q: %w", raw, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return goivy.Version{}, fmt.Errorf("version: bad minor in %q: %w", raw, err)
	}
	return goivy.Version{major, minor}, nil
}
