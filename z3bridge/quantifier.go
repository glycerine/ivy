package z3bridge

// This file provides quantifier support (ForAll/Exists) on top of go-z3,
// which doesn't expose the Z3 quantifier API.

/*
#cgo LDFLAGS: -lz3
#include <z3.h>
#include <stdlib.h>

// Helper to create a ForAll quantifier.
// bound_vars are the bound variables (Z3 constants used as patterns).
// body is the quantified body.
static Z3_ast mk_forall_const(Z3_context ctx, unsigned num_bound, Z3_app bound[], Z3_ast body) {
    return Z3_mk_forall_const(ctx, 0, num_bound, bound, 0, NULL, body);
}

// Helper to create an Exists quantifier.
static Z3_ast mk_exists_const(Z3_context ctx, unsigned num_bound, Z3_app bound[], Z3_ast body) {
    return Z3_mk_exists_const(ctx, 0, num_bound, bound, 0, NULL, body);
}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/aclements/go-z3/z3"
)

// ForAll creates a universally quantified formula: forall bound. body
// bound must be constants (created with ctx.Const or similar).
func ForAll(ctx *z3.Context, bound []z3.Value, body z3.Bool) z3.Bool {
	return mkQuantifier(ctx, true, bound, body)
}

// Exists creates an existentially quantified formula: exists bound. body
func Exists(ctx *z3.Context, bound []z3.Value, body z3.Bool) z3.Bool {
	return mkQuantifier(ctx, false, bound, body)
}

func mkQuantifier(ctx *z3.Context, isForall bool, bound []z3.Value, body z3.Bool) z3.Bool {
	if len(bound) == 0 {
		return body
	}

	// We need to access the underlying C pointers. The go-z3 library
	// exposes AST via AsAST(). We'll use unsafe to extract the C pointers.
	cBound := make([]C.Z3_app, len(bound))
	for i, v := range bound {
		ast := v.AsAST()
		// Z3_app is just a Z3_ast in the C API
		cBound[i] = (C.Z3_app)(astToPtr(ast))
	}

	bodyAST := body.AsAST()
	cBody := (C.Z3_ast)(astToPtr(bodyAST))
	cCtx := ctxToPtr(ctx)

	var result C.Z3_ast
	if isForall {
		result = C.mk_forall_const(cCtx, C.uint(len(bound)), &cBound[0], cBody)
	} else {
		result = C.mk_exists_const(cCtx, C.uint(len(bound)), &cBound[0], cBody)
	}

	C.Z3_inc_ref(cCtx, result)

	runtime.KeepAlive(bound)
	runtime.KeepAlive(body)
	runtime.KeepAlive(ctx)

	// Wrap back into a go-z3 Bool by constructing via AST
	return ptrToBool(ctx, result)
}

// These functions use unsafe to bridge between go-z3's opaque types and C pointers.
// This is necessary because go-z3 doesn't expose its internal C pointers.

func astToPtr(ast z3.AST) unsafe.Pointer {
	// z3.AST contains *astImpl which contains ctx and c C.Z3_ast
	// The C.Z3_ast is at offset sizeof(pointer) in the impl struct
	type astImpl struct {
		ctx unsafe.Pointer // *Context
		c   C.Z3_ast
	}
	type astWrapper struct {
		impl *astImpl
	}
	w := (*astWrapper)(unsafe.Pointer(&ast))
	return unsafe.Pointer(w.impl.c)
}

func ctxToPtr(ctx *z3.Context) C.Z3_context {
	// z3.Context contains *contextImpl which contains c C.Z3_context
	type contextImpl struct {
		c C.Z3_context
	}
	type contextWrapper struct {
		impl *contextImpl
	}
	w := (*contextWrapper)(unsafe.Pointer(ctx))
	return w.impl.c
}

func ptrToBool(ctx *z3.Context, ast C.Z3_ast) z3.Bool {
	// We create a temporary value by going through the AST interface.
	// Use ctx.FromBool to get a template Bool, then overwrite its AST.
	// Alternative: use the Value wrapping pattern.

	// The cleanest approach: create a FuncDecl that returns the quantifier.
	// But actually, the simplest is to use AST manipulation.

	// We'll set up a runtime finalizer for the ast ref.
	// The go-z3 library will manage its own refs via wrapAST.

	// Since we already called Z3_inc_ref above, we need to build a
	// proper wrapper. Let's use the fact that go-z3's Bool type is
	// just a value which is just { *valueImpl, noEq{} }
	// and valueImpl is just astImpl = { ctx, c C.Z3_ast }

	type valueImpl struct {
		ctx unsafe.Pointer
		c   C.Z3_ast
	}
	type noEq struct {
		_ [0]func()
	}
	type boolValue struct {
		impl *valueImpl
		_    noEq
	}

	// Extract the context pointer from ctx
	cCtx := ctxToPtr(ctx)
	_ = cCtx

	// Create a valueImpl pointing to our AST
	impl := &valueImpl{
		ctx: unsafe.Pointer(ctx),
		c:   ast,
	}

	// Set up finalizer to decrement ref
	runtime.SetFinalizer(impl, func(impl *valueImpl) {
		cCtx := ctxToPtr((*z3.Context)(impl.ctx))
		C.Z3_dec_ref(cCtx, impl.c)
	})

	var result z3.Bool
	bv := (*boolValue)(unsafe.Pointer(&result))
	bv.impl = impl

	runtime.KeepAlive(ctx)
	return result
}
