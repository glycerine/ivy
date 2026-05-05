// This file provides a focused Z3 wrapper for translating Ivy logic nodes to
// Z3 expressions and checking satisfiability.
//
// This package wraps the Z3 C API directly via CGo rather than depending on
// go-z3, because go-z3 lacks quantifier support (ForAll/Exists) which is
// essential for Ivy's first-order logic.
package goivy

/*
#cgo CFLAGS: -I${SRCDIR}/z3vendor/z3/src/api
#cgo LDFLAGS: ${SRCDIR}/z3vendor/z3/build/libz3.a

// Use libstdc++ on Linux
#cgo linux LDFLAGS: -lstdc++ -lm -lgomp

// Use libc++ on macOS (Darwin)
#cgo darwin LDFLAGS: -lc++

#include <z3.h>
#include <stdlib.h>

extern void goZ3BridgeErrorHandler(Z3_context c, Z3_error_code e);
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
	//"github.com/glycerine/ivy/goivy/xtracer"
)

// --- Z3Context ---

// Z3Context wraps a Z3 context.
type Z3Context struct {
	c      C.Z3_context
	mu     sync.Mutex
	syms   map[string]C.Z3_symbol
	closed bool

	// z3CheckCounter provides a per Z3Context
	// sequence number for Z3 check calls.
	z3CheckCounter atomic.Int64

	// z3Merkle is a rolling Merkle hash for Z3 check
	// conformance auditing.
	z3Merkle MerkleState
}

//export goZ3BridgeErrorHandler
func goZ3BridgeErrorHandler(ctx C.Z3_context, e C.Z3_error_code) {

	// Z3 error codes (See #Z3_get_error_code):
	// ---------------------  -----------------
	//                 Z3_OK: No error.
	//         Z3_SORT_ERROR: User tried to build an invalid (type incorrect) AST.
	//                Z3_IOB: Index out of bounds.
	//        Z3_INVALID_ARG: Invalid argument was provided.
	//       Z3_PARSER_ERROR: An error occurred when parsing a string or file.
	//          Z3_NO_PARSER: Parser output is not available, that is, user
	//                        didn't invoke #Z3_parse_smtlib2_string or #Z3_parse_smtlib2_file.
	//    Z3_INVALID_PATTERN: Invalid pattern was used to build a quantifier.
	//        Z3_MEMOUT_FAIL: A memory allocation failure was encountered.
	// Z3_FILE_ACCESS_ERRROR: A file could not be accessed.
	//      Z3_INVALID_USAGE: API call is invalid in the current state.
	//     Z3_INTERNAL_FATAL: An error internal to Z3 occurred.
	//      Z3_DEC_REF_ERROR: Trying to decrement the reference counter
	//                        of an AST that was deleted or the reference
	//                        counter was not initialized with #Z3_inc_ref.
	//          Z3_EXCEPTION: Internal Z3 exception. Additional details can
	//                        be retrieved using #Z3_get_error_msg.

	if e == C.Z3_OK {
		// what are we even doing here then...
		return
	}
	msg := C.Z3_get_error_msg(ctx, e)
	panic("z3 bridge panic on error: " + C.GoString(msg))
}

// NewZ3Context creates a new Z3 context.
// Z3 contexts are not thread-safe: each context (and all objects created
// within it) must be used from a single OS thread. In Go, use
// runtime.LockOSThread() to pin the goroutine to its thread before
// creating a context and performing Z3 operations.
//
/*
# thread-safety

Q: Is the Z3 c++ API thread-safe at version 4.7.1 ?

A: The short answer is no. The Z3 C++ API is not thread-safe
if you are attempting to share a single z3::context (and
the objects created within it) across multiple threads.
This was true in version 4.7.1 and remains true in the
latest releases.

Here is how Z3 handles concurrency and what you need
to know to safely multi-thread your application:

The Golden Rule: One Context per Thread

In Z3, almost everything—symbols, sorts, expressions
(z3::expr), and solvers (z3::solver) -- is tied to
the z3::context that created it. The context internally
manages state, memory allocation, and reference
counting for all of its AST (Abstract Syntax Tree) nodes.

Because Z3 does not use internal locking for these
operations (to maximize single-threaded performance),
accessing or modifying a context or its child objects
from multiple threads simultaneously will lead to
race conditions, memory corruption, and crashes.

To run Z3 safely in a multi-threaded C++ environment, you must:

1. Instantiate a completely separate z3::context
inside each thread.

2. Ensure that any z3::solver or z3::expr is only
manipulated by the thread that owns its parent context.

Q: What about a single goroutine holding a Z3 Context,
is that safe? goroutines get executed on different
threads at different times.

A: You have hit on a brilliant and incredibly
subtle edge case. Your understanding of the Go runtime
is spot-on: goroutines are multiplexed onto OS
threads (an M:N scheduling model), meaning a goroutine
can execute a CGO call on OS Thread A, yield, and
later execute the next CGO call on OS Thread B.

Because of this, no, it is not entirely safe to
just let a standard goroutine hold a Z3 context
without taking precautions.

Here is exactly why, and how to fix it.

The Problem: The Thread-Local Storage (TLS) Trap

From a strict concurrency standpoint, confining a
Z3 context to a single goroutine guarantees
sequential access. You won't get data races because
two operations won't happen at the exact same time.

However, C/C++ libraries like Z3 aren't just worried
about simultaneous access; they often care about OS
thread identity. Under the hood, Z3 has historically
used Thread-Local Storage (TLS) for things like:

1. Memory Management: Tracking memory allocations
and reference counts per thread.

2. Internal Caching: Certain thread-specific
execution states or error handlers.

If your goroutine hops threads, you might allocate
an AST node while the goroutine is on OS Thread A,
but later trigger garbage collection, reference
counting updates, or context destruction while the
goroutine is on OS Thread B. This causes Z3's
internal TLS states to get confused (e.g., Thread
B attempts to decrement or free memory that Thread
A "owns"), which can lead to silent memory leaks,
corrupted states, or outright segfaults inside
the CGO boundary.

The Solution: runtime.LockOSThread()

Go provides a specific tool for dealing with C
libraries that care about OS thread identity:

runtime.LockOSThread().

When you call this inside a goroutine, it permanently
wires that specific goroutine to its current
underlying OS thread. The Go scheduler guarantees
that this goroutine will never be moved to another
OS thread, and no other goroutines will be scheduled
on that specific OS thread until it is explicitly
unlocked.

To safely use Z3 in a Go concurrent environment,
you should implement a dedicated "Worker Goroutine" pattern:

~~~
package z3bridge

import (
	"runtime"
	// ... other imports
)

func Z3Worker(jobChannel <-chan Job, resultChannel chan<- Result) {
	// 1. Lock this goroutine to the current OS thread
	runtime.LockOSThread()

	// Optional: Unlock when the goroutine exits (if it ever does)
	defer runtime.UnlockOSThread()

	// 2. NOW it is perfectly safe to initialize your Z3 context
	// ctx := C.Z3_mk_context(...)
	// defer C.Z3_del_context(ctx)

	// 3. Listen for work on a channel and interact with Z3 sequentially
	for job := range jobChannel {
		// Process the job using the Z3 context...
		// Because of LockOSThread, Z3's TLS remains perfectly intact.

		// resultChannel <- result
	}
}
~~~

By confining your Z3 context not just to a single goroutine,
but to a single locked OS thread, you satisfy both Go's
concurrency model and Z3's internal C++ assumptions.

# open Context count limits

Q: Can Z3 have 100 open Context at the same time, or do
I need to close one before I make a new one, with
Z3_mk_context_rc()/Z3_del_context() ?

A: Yes, you can absolutely have 100 (or even 1,000)
open Z3 Contexts at the same time. You do not need to
close one before making a new one.

In Z3, a Z3_context is essentially a completely independent sandbox. Each context has its own memory manager, its own symbol table, and its own configuration. Because they do not share global state, they can coexist perfectly within the same process.

In fact, creating multiple contexts is the standard way to achieve parallel solving in Z3.

Here is how this applies to your Go/CGO setup:

The Worker Pool Pattern
Since we just discussed that a Z3 context must be tied to a single locked OS thread, having multiple open contexts is exactly how you scale your application. You can create a pool of 100 Go worker routines, where:

1. Each worker calls runtime.LockOSThread().

2. Each worker calls C.Z3_mk_context_rc() to initialize its own private context.

3. All 100 workers pull jobs from a Go channel and solve formulas simultaneously in parallel.

4. When a worker is permanently done, it calls C.Z3_del_context().

Things to Keep in Mind

Memory Consumption: While Z3 doesn't care how many contexts
you open, your RAM will. An empty Z3 context is relatively
lightweight, but once you start populating them with complex
ASTs (Abstract Syntax Trees) and solvers, they can consume
a significant amount of memory. 100 heavily loaded contexts
could easily eat up gigabytes of RAM.

No Cross-Context Sharing: Remember that AST nodes
(expressions, sorts, symbols) belong strictly to the context
that created them. You cannot take an expression created in
Context A and assert it in Context B without explicitly
translating it first using Z3_ast_translate.

Reference Counting (Z3_mk_context_rc): You mentioned
Z3_mk_context_rc. This is the recommended initialization
function when building bindings for garbage-collected
languages like Go. It turns on reference counting for AST
nodes, allowing you to tie the lifecycle of Z3 objects to
Go's finalizers. Just ensure you actually call Z3_inc_ref
and Z3_dec_ref on the objects you create, or you will
leak memory inside the C heap.

# memory management, reset the solver not the context,
unless fuzz testing (see below).

Q: can I re-use a Z3 Context by clearing it or reseting
it to empty somehow? what is the API for that?

A: "You cannot reset or clear a Z3_context. There is no
API function like Z3_reset_context.

If you want a completely empty context, you must destroy
the old one with Z3_del_context(ctx) and create a brand
new one with Z3_mk_context_rc(config).

However, it is highly likely that you don't actually need
to reset the context. You probably just want to reset the solver.

Here is a breakdown of why this is, and how to get the behavior you are looking for.

Why You Can't Reset a Context

In Z3, the Z3_context is the global memory manager and
factory for your mathematical universe. Every time you
create a symbol, a sort, or an expression (AST node),
Z3 "interns" it inside the context.

If Z3 provided a way to "wipe" the context, it would
immediately invalidate every single Z3_ast pointer your
Go program is currently holding, leading to catastrophic
use-after-free segfaults. Because of this design, the
context itself only grows. Memory is only fully reclaimed
when you delete the entire context (or when reference counts
drop to zero, if using _rc functions, but the underlying
memory pools don't strictly shrink on command).

The Solution: Resetting the Solver

If your goal is to clear out all your old formulas and
start solving a brand new problem, you should reuse the
context but reset the solver.

The solver is the stateful object that actually holds
your assertions and checks for satisfiability. You can
wipe a solver completely clean using the C API:

C.Z3_solver_reset(ctx, solver)

What Z3_solver_reset does:

It removes all assertions (formulas) you have added to the solver.

It clears any internal caches or learned clauses
from previous Z3_solver_check runs.

It leaves the Z3_context completely intact, meaning all
your variables, sorts, and functions are still valid
and can be reused immediately.

This is extremely fast and is the standard way to
run thousands of sequential queries in Z3 without the
overhead of tearing down and spinning up a new context every time.

When You Must Recreate the Context

The only time you should genuinely destroy and recreate
a Z3_context in a loop is if you are suffering from memory bloat.

If you run a fuzzer (like in your Makefile) that generates
millions of unique, complex AST nodes (massive equations), the
context's internal memory pools will swell. Even if you reset the
solver and decrement the reference counts of the ASTs, the context
might hold onto that allocated RAM pool. In long-running server
processes or heavy fuzzing loops, periodically nuking the context
and making a new one is a valid strategy to return that RAM to
the operating system.

docs:
https://z3prover.github.io/api/html/classz3_1_1solver.html

# fuzzing means you must delete the Context after each test.

Q: okay, but won't that re-using a context across tests
result in cross-talk between tests? that is a big non-starter
for Go fuzz testing.

A: You are 100% right to flag this.

For Go fuzz testing specifically, **reusing a `Z3_context`
is a bad idea**, and your instinct to demand strict
isolation is spot on.

Here is the breakdown of why reusing a context in a fuzzer
is dangerous, both in terms of logical cross-talk and
system stability.

### 1. The Risk of Logical Cross-Talk (Symbol Collisions)
While wiping the `Z3_solver` clears out the *equations*,
the `Z3_context` still remembers the *vocabulary*.

If Fuzz Iteration A creates a boolean variable named `"x"`
and asserts something about it, and Fuzz Iteration B also
asks the context for a variable named `"x"`, Z3 will
return the exact same underlying AST node from its
internal cache.

If your Go code isn't meticulously careful about
clearing out its own Go-side maps or ensuring globally
unique symbol names per iteration, you can easily end
up with cross-talk where a test accidentally references
a leftover constraint or type definition from a
previous run.

### 2. The Real Killer: Fuzzer OOM (Out of Memory)
Even if you perfectly isolate your logic and reset
your solvers, reusing a context during fuzzing will
almost certainly crash your machine.

Fuzzers run thousands of iterations per second, throwing
chaotic, garbage data at your logic. Every time your
bridge parses that data and creates a new Z3 expression,
sort, or symbol, the `Z3_context` allocates memory for it.
Because contexts are interning pools designed to never
invalidate active pointers, **a context never shrinks**.

If you reuse one context for a fuzz run, its memory
pool will swell continuously until the OS kills the
Go process for eating all your RAM.

### The Verdict for Go Fuzzing: Tear It Down

In the context of `go test -fuzz`, you should embrace
the overhead and completely create and destroy the
context for every single fuzz execution. The few
milliseconds of overhead to spin up a context is the
only way to guarantee memory safety and mathematical
isolation.

Because Go 1.18+ fuzzing runs your fuzz target
concurrently across multiple goroutines automatically,
you still need to respect the OS thread rules we talked
about.

Your ideal fuzz target should look something like this:

```go
func FuzzMyEq(f *testing.F) {
    // Add corpus data...
    f.Add(...)

    f.Fuzz(func(t *testing.T, data []byte) {
        // 1. Lock this specific fuzz execution to an OS thread
        runtime.LockOSThread()
        defer runtime.UnlockOSThread()

        // 2. Create a pristine, isolated universe for this iteration
        // Assuming you have a wrapper or call C directly:
        // ctx := NewZ3Context()
        // defer ctx.Close()

        // 3. Run your solver logic...
        // Even if this iteration panics or generates massive ASTs,
        // the deferred Z3_del_context call will completely
        // nuke the memory and state.
    })
}
```

This guarantees zero cross-talk, zero thread-local storage
corruption, and stable memory usage, no matter how long the fuzzer runs.

# Actual API for ref-counting or Z3 doing GC itself:

(We tried the non-rc version: got lots of CGO signal
problems and segfaults, and have retreated to _rc land).

Context and AST Reference Counting
Z3_context Z3_API 	Z3_mk_context (Z3_config c)
 	Create a context using the given configuration.

Z3_context Z3_API 	Z3_mk_context_rc (Z3_config c)
 	Create a context using the given configuration.
This function is similar to Z3_mk_context. However, in
the context returned by this function, the user is
responsible for managing Z3_ast reference counters.
Managing reference counters is a burden and error-prone,
but allows the user to use the memory more efficiently.
The user must invoke Z3_inc_ref for any Z3_ast returned
by Z3, and Z3_dec_ref whenever the Z3_ast is not needed
anymore. This idiom is similar to the one used in BDD
(binary decision diagrams) packages such as CUDD.

*/
func NewZ3Context() *Z3Context {
	cfg := C.Z3_mk_config()
	defer C.Z3_del_config(cfg)

	// if doing interpolation, you need to also:
	// C.Z3_set_param_value(cfg, "PROOF", "true")
	// C.Z3_set_param_value(cfg, "MODEL", "true")
	// which is precisely what
	// C.Z3_mk_interpolation_context(cfg)
	// does for you automatically.
	// See https://z3prover.github.io/api/html/group__capi.html#ga893d6f1df01056553b7ca1ba3a4e848c

	// _rc means with reference counting turned on...
	// "Just ensure you actually call Z3_inc_ref and Z3_dec_ref
	// on the objects you create, or you will leak memory inside the C heap."
	c := C.Z3_mk_context_rc(cfg)

	// note there is also a Z3-does-GC version, but I'd rather leak
	// memory for now that have Z3 do a rug pull because it couldn't
	// see we were using something and then get a mysterious crash.
	// Maybe later, if memory becomes an issue:
	//
	// --- begin docs from Claude ---
	// Z3_mk_context(cfg) (without _rc) is the simpler API — Z3 manages
	// all memory internally via garbage collection, with no reference counting at all.
	// When you call Z3_del_context(), everything is freed.
	//
	// The _rc variant (Z3_mk_context_rc) exists specifically for languages that want to
	// tie Z3 object lifetimes to their own GC via inc_ref/dec_ref — exactly the Python
	// __del__ pattern. But we've abandoned that pattern (commented-out finalizers, no
	// dec_ref calls). So right now we're paying the overhead of _rc mode (every
	// newExpr/newSort/newFuncDecl calls Z3_inc_ref via CGO) without getting any benefit
	// from it, since Z3_del_context nukes everything anyway.
	//
	// Switching to Z3_mk_context:
	// - Remove all Z3_inc_ref calls in newExpr, newSort, newFuncDecl
	// - Remove all commented-out Z3_dec_ref finalizer code
	// - Change Z3_mk_context_rc(cfg) → Z3_mk_context(cfg) in NewZ3Context
	// - Keep Z3_del_context in Close() — works the same either way
	//
	// The only thing to watch: Z3_mk_context uses its own internal GC that can collect
	// AST nodes when Z3 decides they're unreachable from its perspective. If Go holds a
	// C.Z3_ast pointer but Z3 doesn't know about it (no solver references it, no other
	// AST references it), Z3 might collect it. In practice this shouldn't happen because
	//  our patterns always feed ASTs into solvers or larger expressions before Z3's GC
	// runs, but it's worth knowing. If we ever hit a use-after-free, the fix would be to
	//  call Z3_persist_ast(ctx, ast) on long-lived nodes — but I'd cross that bridge
	// only if we see it.
	// --- end docs ---
	//
	// What does Python do?
	// Python Ivy uses one global Z3 context for the entire process
	// lifetime. It never closes it.
	//
	// Key details:
	//
	// 1. Single global context: z3.main_ctx() — a process-wide singleton
	// created by the Python Z3 binding on first use. Never destroyed
	// until process exit.
	//
	// 2. Module-level caches accumulate forever (ivy_solver.py:228-235):
	//   - z3_sorts — cached DeclareSort() results
	//   - z3_constants — cached Const() and enum constants
	//   - z3_functions — cached Function() declarations
	//   - z3_predicates — cached relation declarations
	//
	// 3. Solvers are transient: z3.Solver() is created per-query (21
	// instances across the codebase), used, then dropped for Python GC
	// to collect. No explicit reset() or del.
	//
	// 4. clear() function (line 228): Resets the four Ivy-level cache dicts to empty.
	// Called between compilation units (e.g., from sidecar.py). This drops Python
	// references to Z3 objects, but does NOT touch the Z3 context — the context's
	// internal memory pools keep growing.
	//
	// 5. No reference counting (directly) by Ivy: but Python's Z3 binding does and
	// handles ref counting.

	// quoting github.com/aclements/go-z3/z3/context.go:114,
	// "[This can be used to install] an error handler
	// that turns errors into Go panics.
	// This error handler is equivalent to a longjmp on the C++
	// side, but Z3 is actually designed to handle that, which is
	// nice because it saves us the trouble of checking the
	// context's error code all over the place."
	C.Z3_set_error_handler(c, (*C.Z3_error_handler)(C.goZ3BridgeErrorHandler))

	ctx := &Z3Context{c: c, syms: make(map[string]C.Z3_symbol)}

	// caller should prefer to arrange to "defer ctx.Close()" instead of:
	//runtime.SetFinalizer(ctx, func(ctx *Z3Context) {
	//	ctx.Close()
	//})
	return ctx
}

func (ctx *Z3Context) Close() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.closed {
		return nil
	}
	ctx.closed = true
	C.Z3_del_context(ctx.c)
	return nil
}

// do runs f with the context lock held.
func (ctx *Z3Context) do(f func()) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	f()
}

func (ctx *Z3Context) symbol(name string) C.Z3_symbol {
	if sym, ok := ctx.syms[name]; ok {
		return sym
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	sym := C.Z3_mk_string_symbol(ctx.c, cname)
	ctx.syms[name] = sym
	return sym
}

// --- Sort ---

// Sort wraps a Z3 sort (type).
type Z3Sort struct {
	ctx *Z3Context
	c   C.Z3_sort
}

// String returns the name of the Z3 sort.
func (s Z3Sort) String() string {
	var res string
	s.ctx.do(func() {
		sym := C.Z3_get_sort_name(s.ctx.c, s.c)
		res = C.GoString(C.Z3_get_symbol_string(s.ctx.c, sym))
	})
	runtime.KeepAlive(s)
	return res
}

// GetId returns the unique numeric AST ID for this sort.
// Corresponds to Python's get_id() which calls Z3_get_ast_id.
func (s Z3Sort) GetId() uint {
	//xtracer.Trace("ivy_solver.py:781 get_id() ENTER")
	var id uint
	s.ctx.do(func() {
		ast := C.Z3_sort_to_ast(s.ctx.c, s.c) // python's x.as_ast() does this internally.
		id = uint(C.Z3_get_ast_id(s.ctx.c, ast))
	})
	runtime.KeepAlive(s)
	return id
}

// incRefSort must be called with ctx lock held.
func (ctx *Z3Context) incRefSort(c C.Z3_sort) {
	C.Z3_inc_ref(ctx.c, C.Z3_sort_to_ast(ctx.c, c))
}

func (ctx *Z3Context) newSort(c C.Z3_sort) Z3Sort {
	// Called with lock held — do raw ref counting
	ctx.incRefSort(c)
	s := Z3Sort{ctx: ctx, c: c}
	/*
		runtime.SetFinalizer(&s, func(s *Sort) {
			s.ctx.do(func() {
				// caused panic: maybe b/c ctx was already closed?

				   // panic: z3 bridge panic on error: invalid dec_ref command

				   // goroutine 18 [running]:
				   // github.com/glycerine/ivy/goivy/z3bridge.goZ3BridgeErrorHandler(0x7f951902a208, 0xb)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:71 +0x10c
				   // github.com/glycerine/ivy/goivy/z3bridge._Cfunc_Z3_dec_ref(0x7f951902a208, 0x7f951903ced8)
				   // 	_cgo_gotypes.go:280 +0x5b
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).newSort.func1.1.1(...)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:528
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).newSort.func1.1()
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:528 +0xa5
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).do(0x3992eb5b4008?, 0x3992eb866000?)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:484 +0xdd
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).newSort.func1(0x0?)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:527 +0x49
				   // runtime.runFinalizers()
				   // 	/usr/local/go/src/runtime/mfinal.go:272 +0x3f7

				//C.Z3_dec_ref(s.ctx.c, C.Z3_sort_to_ast(s.ctx.c, s.c))
			})
		})
	*/
	return s
}

// BoolSort returns the Boolean sort.
func (ctx *Z3Context) BoolSort() Z3Sort {
	var s Z3Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_bool_sort(ctx.c))
	})
	return s
}

// UninterpretedSort returns an uninterpreted sort with the given name.
func (ctx *Z3Context) UninterpretedSort(name string) Z3Sort {
	sym := ctx.symbol(name)
	var s Z3Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_uninterpreted_sort(ctx.c, sym))
	})
	return s
}

// IntSort returns the integer sort.
func (ctx *Z3Context) IntSort() Z3Sort {
	var s Z3Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_int_sort(ctx.c))
	})
	return s
}

// RealSort returns the real number sort.
func (ctx *Z3Context) RealSort() Z3Sort {
	var s Z3Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_real_sort(ctx.c))
	})
	return s
}

// --- Expr (symbolic value) ---

// Expr wraps a Z3 expression (symbolic value).
type Z3Expr struct {
	ctx *Z3Context
	c   C.Z3_ast
}

// newExpr creates an Expr from a C Z3_ast. Must be called with ctx lock held.
func (ctx *Z3Context) newExpr(c C.Z3_ast) Z3Expr {
	C.Z3_inc_ref(ctx.c, c)
	e := Z3Expr{ctx: ctx, c: c}
	/*
		runtime.SetFinalizer(&e, func(e *Expr) {
			e.ctx.do(func() {
				// caused panic in fuzz test: === RUN   FuzzQuantConstraintsForAll

				   // translate2_fuzz_test.go:139 [goID 26] 2026-03-19 08:22:50.463886000 +0000 UTC ran fine
				   // panic: z3 bridge panic on error: invalid dec_ref command

				   // goroutine 5 [running]:
				   // github.com/glycerine/ivy/goivy/z3bridge.goZ3BridgeErrorHandler(0x7fa1e2008808, 0xb)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:71 +0x10c
				   // github.com/glycerine/ivy/goivy/z3bridge._Cfunc_Z3_dec_ref(0x7fa1e2008808, 0x7fa1e2019d90)
				   // 	_cgo_gotypes.go:280 +0x5b
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).newExpr.func2.1.1(...)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:606
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).newExpr.func2.1()
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:606 +0x9d
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).do(0x2e4d48034110?, 0x2e4d4807c610?)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:484 +0xdd
				   // github.com/glycerine/ivy/goivy/z3bridge.(*Z3Context).newExpr.func2(0x0?)
				   // 	/Users/jaten/goivy/z3bridge/quantifier.go:605 +0x49
				   // runtime.runFinalizers()
				   // 	/usr/local/go/src/runtime/mfinal.go:272 +0x3f7


				//C.Z3_dec_ref(e.ctx.c, e.c)
			})
		})
	*/
	return e
}

// String returns the S-expression representation.
func (e Z3Expr) String() string {
	var res string
	e.ctx.do(func() {
		res = C.GoString(C.Z3_ast_to_string(e.ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return res
}

// Const creates a named constant of the given sort.
func (ctx *Z3Context) Const(name string, sort Z3Sort) Z3Expr {
	sym := ctx.symbol(name)
	var e Z3Expr
	ctx.do(func() {
		e = ctx.newExpr(C.Z3_mk_const(ctx.c, sym, sort.c))
	})
	runtime.KeepAlive(sort)
	return e
}

// BoolVal returns a boolean literal.
func (ctx *Z3Context) BoolVal(val bool) Z3Expr {
	var e Z3Expr
	ctx.do(func() {
		if val {
			e = ctx.newExpr(C.Z3_mk_true(ctx.c))
		} else {
			e = ctx.newExpr(C.Z3_mk_false(ctx.c))
		}
	})
	return e
}

// IntVal returns an integer literal.
func (ctx *Z3Context) IntVal(val int64) Z3Expr {
	var e Z3Expr
	ctx.do(func() {
		sort := C.Z3_mk_int_sort(ctx.c)
		e = ctx.newExpr(C.Z3_mk_int64(ctx.c, C.int64_t(val), sort))
	})
	return e
}

// --- FuncDecl ---

// FuncDecl wraps a Z3 function declaration.
type FuncDecl struct {
	ctx *Z3Context
	c   C.Z3_func_decl
}

// newFuncDecl creates a FuncDecl. Must be called with ctx lock held.
func (ctx *Z3Context) newFuncDecl(c C.Z3_func_decl) FuncDecl {
	C.Z3_inc_ref(ctx.c, C.Z3_func_decl_to_ast(ctx.c, c))
	fd := FuncDecl{ctx: ctx, c: c}
	//runtime.SetFinalizer(&fd, func(fd *FuncDecl) {
	//	fd.ctx.do(func() {
	//		C.Z3_dec_ref(fd.ctx.c, C.Z3_func_decl_to_ast(fd.ctx.c, fd.c))
	//	})
	//})
	return fd
}

// Function creates an uninterpreted function declaration.
func (ctx *Z3Context) Function(name string, domain []Z3Sort, range_ Z3Sort) FuncDecl {
	sym := ctx.symbol(name)
	cdomain := make([]C.Z3_sort, len(domain))
	for i, s := range domain {
		cdomain[i] = s.c
	}
	var fd FuncDecl
	ctx.do(func() {
		var cdp *C.Z3_sort
		if len(cdomain) > 0 {
			cdp = &cdomain[0]
		}
		fd = ctx.newFuncDecl(C.Z3_mk_func_decl(ctx.c, sym, C.uint(len(cdomain)), cdp, range_.c))
	})
	runtime.KeepAlive(domain)
	runtime.KeepAlive(range_)
	return fd
}

// AsExpr converts a FuncDecl to an Expr via Z3_func_decl_to_ast.
// This is needed for Z3_substitute which operates on AST nodes.
// Matches Python's FuncDeclRef.as_ast() used in z3.substitute().
func (fd FuncDecl) AsExpr() Z3Expr {
	var e Z3Expr
	fd.ctx.do(func() {
		e = fd.ctx.newExpr(C.Z3_func_decl_to_ast(fd.ctx.c, fd.c))
	})
	runtime.KeepAlive(fd)
	return e
}

// Apply applies the function declaration to arguments, returning an expression.
func (fd FuncDecl) Apply(args ...Z3Expr) Z3Expr {
	cargs := make([]C.Z3_ast, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var e Z3Expr
	fd.ctx.do(func() {
		var cap *C.Z3_ast
		if len(cargs) > 0 {
			cap = &cargs[0]
		}
		e = fd.ctx.newExpr(C.Z3_mk_app(fd.ctx.c, fd.c, C.uint(len(cargs)), cap))
	})
	runtime.KeepAlive(fd)
	runtime.KeepAlive(args)
	return e
}

// --- Boolean operations ---

// Not returns the negation of e.
func (ctx *Z3Context) Not(e Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_not(ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// And returns the conjunction of expressions.
func (ctx *Z3Context) And(args ...Z3Expr) Z3Expr {
	if len(args) == 0 {
		return ctx.BoolVal(true)
	}
	cargs := make([]C.Z3_ast, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_and(ctx.c, C.uint(len(cargs)), &cargs[0]))
	})
	runtime.KeepAlive(args)
	return r
}

// Or returns the disjunction of expressions.
func (ctx *Z3Context) Or(args ...Z3Expr) Z3Expr {
	if len(args) == 0 {
		return ctx.BoolVal(false)
	}
	cargs := make([]C.Z3_ast, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_or(ctx.c, C.uint(len(cargs)), &cargs[0]))
	})
	runtime.KeepAlive(args)
	return r
}

// Implies returns e1 => e2.
func (ctx *Z3Context) Implies(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_implies(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Iff returns e1 <=> e2.
func (ctx *Z3Context) Iff(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_iff(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Eq returns e1 == e2.
func (ctx *Z3Context) Eq(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_eq(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Ite returns if cond then then_ else else_.
func (ctx *Z3Context) Ite(cond, then_, else_ Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_ite(ctx.c, cond.c, then_.c, else_.c))
	})
	runtime.KeepAlive(cond)
	runtime.KeepAlive(then_)
	runtime.KeepAlive(else_)
	return r
}

// --- Arithmetic ---

// Add returns e1 + e2 (integer or real arithmetic).
func (ctx *Z3Context) Add(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		args := [2]C.Z3_ast{e1.c, e2.c}
		r = ctx.newExpr(C.Z3_mk_add(ctx.c, 2, &args[0]))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Sub returns e1 - e2 (integer or real arithmetic).
func (ctx *Z3Context) Sub(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		args := [2]C.Z3_ast{e1.c, e2.c}
		r = ctx.newExpr(C.Z3_mk_sub(ctx.c, 2, &args[0]))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Mul returns e1 * e2 (integer or real arithmetic).
func (ctx *Z3Context) Mul(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		args := [2]C.Z3_ast{e1.c, e2.c}
		r = ctx.newExpr(C.Z3_mk_mul(ctx.c, 2, &args[0]))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Div returns e1 / e2 (integer division).
func (ctx *Z3Context) Div(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_div(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Gt returns e1 > e2 (arithmetic comparison).
func (ctx *Z3Context) Gt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_gt(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Lt returns e1 < e2 (arithmetic comparison).
func (ctx *Z3Context) Lt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_lt(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Ge returns e1 >= e2 (arithmetic comparison).
func (ctx *Z3Context) Ge(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_ge(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Le returns e1 <= e2 (arithmetic comparison).
func (ctx *Z3Context) Le(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_le(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// EnumSort creates a Z3 enumeration sort with the given name and element names.
// Returns the sort and the constructor constants for each element.
// Matches Python's z3.EnumSort(name, extension).
func (ctx *Z3Context) EnumSort(name string, elements []string) (Z3Sort, []Z3Expr) {
	var s Z3Sort
	consts := make([]Z3Expr, len(elements))
	ctx.do(func() {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		sym := C.Z3_mk_string_symbol(ctx.c, cName)

		n := C.unsigned(len(elements))
		cElems := make([]C.Z3_symbol, len(elements))
		for i, e := range elements {
			ce := C.CString(e)
			cElems[i] = C.Z3_mk_string_symbol(ctx.c, ce)
			C.free(unsafe.Pointer(ce))
		}

		cConsts := make([]C.Z3_func_decl, len(elements))
		cTesters := make([]C.Z3_func_decl, len(elements))

		var elemsPtr *C.Z3_symbol
		if len(cElems) > 0 {
			elemsPtr = &cElems[0]
		}
		var constsPtr *C.Z3_func_decl
		if len(cConsts) > 0 {
			constsPtr = &cConsts[0]
		}
		var testersPtr *C.Z3_func_decl
		if len(cTesters) > 0 {
			testersPtr = &cTesters[0]
		}

		zs := C.Z3_mk_enumeration_sort(ctx.c, sym, n, elemsPtr, constsPtr, testersPtr)
		s = ctx.newSort(zs)

		// Extract constructor constants
		for i := range elements {
			app := C.Z3_mk_app(ctx.c, cConsts[i], 0, nil)
			consts[i] = ctx.newExpr(app)
		}
	})
	return s, consts
}

// StringSort returns the string sort.
// Corresponds to Python's z3.StringSort().
func (ctx *Z3Context) StringSort() Z3Sort {
	var s Z3Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_string_sort(ctx.c))
	})
	return s
}

// StringVal creates a Z3 string constant.
// Corresponds to Python's z3.StringVal(s).
func (ctx *Z3Context) StringVal(s string) Z3Expr {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	var e Z3Expr
	ctx.do(func() {
		e = ctx.newExpr(C.Z3_mk_string(ctx.c, cs))
	})
	return e
}

// --- Bit-Vector Operations ---

// BvSort creates a bit-vector sort of the given width.
func (ctx *Z3Context) BvSort(width int) Z3Sort {
	var s Z3Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_bv_sort(ctx.c, C.unsigned(width)))
	})
	return s
}

// BvVal creates a bit-vector constant from an integer value.
func (ctx *Z3Context) BvVal(val int64, width int) Z3Expr {
	var e Z3Expr
	ctx.do(func() {
		sort := C.Z3_mk_bv_sort(ctx.c, C.unsigned(width))
		e = ctx.newExpr(C.Z3_mk_int64(ctx.c, C.int64_t(val), sort))
	})
	return e
}

// BvAnd returns bitwise AND of two bit-vectors.
func (ctx *Z3Context) BvAnd(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvand(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvOr returns bitwise OR of two bit-vectors.
func (ctx *Z3Context) BvOr(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvor(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvNot returns bitwise NOT of a bit-vector.
func (ctx *Z3Context) BvNot(e Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvnot(ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// BvAdd returns bit-vector addition.
func (ctx *Z3Context) BvAdd(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvadd(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvSub returns bit-vector subtraction.
func (ctx *Z3Context) BvSub(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvsub(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvMul returns bit-vector multiplication.
func (ctx *Z3Context) BvMul(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvmul(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUdiv returns unsigned bit-vector division.
func (ctx *Z3Context) BvUdiv(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvudiv(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvShl returns bit-vector shift left.
func (ctx *Z3Context) BvShl(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvshl(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvLshr returns bit-vector logical shift right.
func (ctx *Z3Context) BvLshr(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvlshr(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvAshr returns bit-vector arithmetic shift right.
func (ctx *Z3Context) BvAshr(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvashr(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvXor returns bitwise XOR of two bit-vectors.
func (ctx *Z3Context) BvXor(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvxor(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Concat returns the concatenation of two bit-vectors.
func (ctx *Z3Context) Concat(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_concat(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Extract returns bits [hi:lo] from a bit-vector (hi and lo are inclusive).
func (ctx *Z3Context) Extract(hi, lo int, e Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_extract(ctx.c, C.unsigned(hi), C.unsigned(lo), e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// Bv2Int converts a bit-vector to an integer (unsigned).
func (ctx *Z3Context) Bv2Int(e Z3Expr, isSigned bool) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		var s C.bool
		if isSigned {
			s = C.bool(true)
		}
		r = ctx.newExpr(C.Z3_mk_bv2int(ctx.c, e.c, s))
	})
	runtime.KeepAlive(e)
	return r
}

// Int2Bv converts an integer to a bit-vector of given width.
func (ctx *Z3Context) Int2Bv(width int, e Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_int2bv(ctx.c, C.unsigned(width), e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// BvUlt returns unsigned less-than comparison of bit-vectors.
func (ctx *Z3Context) BvUlt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvult(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUle returns unsigned less-than-or-equal comparison of bit-vectors.
func (ctx *Z3Context) BvUle(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvule(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUgt returns unsigned greater-than comparison of bit-vectors.
func (ctx *Z3Context) BvUgt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvugt(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUge returns unsigned greater-than-or-equal comparison of bit-vectors.
func (ctx *Z3Context) BvUge(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvuge(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// IsBvExpr returns true if the expression has a bit-vector sort.
func (ctx *Z3Context) IsBvExpr(e Z3Expr) bool {
	return ctx.IsBvSort(e.ExprSort())
}

// IsBvSort returns true if the sort is a bit-vector sort.
func (ctx *Z3Context) IsBvSort(s Z3Sort) bool {
	var r bool
	ctx.do(func() {
		r = C.Z3_get_sort_kind(ctx.c, s.c) == C.Z3_BV_SORT
	})
	return r
}

// BvSortSize returns the width of a bit-vector sort.
func (ctx *Z3Context) BvSortSize(s Z3Sort) int {
	var r int
	ctx.do(func() {
		r = int(C.Z3_get_bv_sort_size(ctx.c, s.c))
	})
	return r
}

// --- Quantifiers ---

// ForAll creates a universally quantified formula.
// bound are the bound variables (must be constants created with Const).
func (ctx *Z3Context) ForAll(bound []Z3Expr, body Z3Expr) Z3Expr {
	if len(bound) == 0 {
		return body
	}
	cbound := make([]C.Z3_app, len(bound))
	for i, b := range bound {
		cbound[i] = C.Z3_to_app(ctx.c, b.c)
	}
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_forall_const(
			ctx.c,
			0, // weight
			C.uint(len(cbound)),
			&cbound[0],
			0,   // num_patterns
			nil, // patterns
			body.c,
		))
	})
	runtime.KeepAlive(bound)
	runtime.KeepAlive(body)
	return r
}

// Exists creates an existentially quantified formula.
func (ctx *Z3Context) Exists(bound []Z3Expr, body Z3Expr) Z3Expr {
	if len(bound) == 0 {
		return body
	}
	cbound := make([]C.Z3_app, len(bound))
	for i, b := range bound {
		cbound[i] = C.Z3_to_app(ctx.c, b.c)
	}
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_exists_const(
			ctx.c,
			0, // weight
			C.uint(len(cbound)),
			&cbound[0],
			0,   // num_patterns
			nil, // patterns
			body.c,
		))
	})
	runtime.KeepAlive(bound)
	runtime.KeepAlive(body)
	return r
}

// --- Solver ---

// CheckResult represents the result of a satisfiability check.
type Z3CheckResult int

const (
	Sat     Z3CheckResult = 1
	Unsat   Z3CheckResult = -1
	Unknown Z3CheckResult = 0
)

func (r Z3CheckResult) String() string {
	switch r {
	case Sat:
		return "sat"
	case Unsat:
		return "unsat"
	default:
		return "unknown"
	}
}

// Z3Solver wraps a Z3 solver.
type Z3Solver struct {
	ctx *Z3Context
	c   C.Z3_solver
}

// NewZ3Solver creates a new solver.
func (ctx *Z3Context) NewZ3Solver() *Z3Solver {
	var s *Z3Solver
	ctx.do(func() {
		cs := C.Z3_mk_solver(ctx.c)
		C.Z3_solver_inc_ref(ctx.c, cs)
		s = &Z3Solver{ctx: ctx, c: cs}
	})
	//runtime.SetFinalizer(s, func(s *Z3Solver) {
	//	s.ctx.do(func() {
	//		C.Z3_solver_dec_ref(s.ctx.c, s.c)
	//	})
	//})
	return s
}

// Assert adds a constraint to the solver.
func (s *Z3Solver) Assert(e Z3Expr) {
	s.ctx.do(func() {
		C.Z3_solver_assert(s.ctx.c, s.c, e.c)
	})
	runtime.KeepAlive(e)
}

// Check checks satisfiability.
func (s *Z3Solver) Check() Z3CheckResult {
	var r Z3CheckResult
	s.ctx.do(func() {
		res := C.Z3_solver_check(s.ctx.c, s.c)
		r = Z3CheckResult(res)
	})
	runtime.KeepAlive(s)
	s.TraceCheck(r)
	return r
}

// Push creates a backtracking point.
func (s *Z3Solver) Push() {
	s.ctx.do(func() {
		C.Z3_solver_push(s.ctx.c, s.c)
	})
}

// Pop removes constraints added since the last Push.
func (s *Z3Solver) Pop() {
	s.ctx.do(func() {
		C.Z3_solver_pop(s.ctx.c, s.c, 1)
	})
}

// Reset removes all assertions from the solver.
func (s *Z3Solver) Reset() {
	s.ctx.do(func() {
		C.Z3_solver_reset(s.ctx.c, s.c)
	})
}

// String returns a string representation of the solver's assertions.
func (s *Z3Solver) String() string {
	var res string
	s.ctx.do(func() {
		res = C.GoString(C.Z3_solver_to_string(s.ctx.c, s.c))
	})
	runtime.KeepAlive(s)
	return res
}

// --- Model ---

// Model wraps a Z3 model (satisfying assignment).
type Model struct {
	ctx *Z3Context
	c   C.Z3_model
}

// Model returns the model from the last successful Check.
func (s *Z3Solver) Model() *Model {
	var m *Model
	s.ctx.do(func() {
		cm := C.Z3_solver_get_model(s.ctx.c, s.c)
		if cm != nil {
			C.Z3_model_inc_ref(s.ctx.c, cm)
			m = &Model{ctx: s.ctx, c: cm}
		}
	})
	if m != nil {
		//runtime.SetFinalizer(m, func(m *Model) {
		//	m.ctx.do(func() {
		//		C.Z3_model_dec_ref(m.ctx.c, m.c)
		//	})
		//})
	}
	runtime.KeepAlive(s)
	return m
}

// Eval evaluates an expression in the model. If completion is true,
// assigns default values for unconstrained variables.
func (m *Model) Eval(e Z3Expr, completion bool) (Z3Expr, bool) {
	var result Z3Expr
	var ok bool
	m.ctx.do(func() {
		var cresult C.Z3_ast
		if bool(C.Z3_model_eval(m.ctx.c, m.c, e.c, C.bool(completion), &cresult)) {
			result = m.ctx.newExpr(cresult)
			ok = true
		}
	})
	runtime.KeepAlive(m)
	runtime.KeepAlive(e)
	return result, ok
}

// Sorts returns all uninterpreted sorts in the model.
func (m *Model) Sorts() []Z3Sort {
	var result []Z3Sort
	m.ctx.do(func() {
		n := int(C.Z3_model_get_num_sorts(m.ctx.c, m.c))
		for i := 0; i < n; i++ {
			cs := C.Z3_model_get_sort(m.ctx.c, m.c, C.uint(i))
			result = append(result, m.ctx.newSort(cs))
		}
	})
	runtime.KeepAlive(m)
	return result
}

// SortUniverse returns the universe (all elements) for a sort in the model.
func (m *Model) SortUniverse(s Z3Sort) []Z3Expr {
	var result []Z3Expr
	m.ctx.do(func() {
		av := C.Z3_model_get_sort_universe(m.ctx.c, m.c, s.c)
		if av != nil {
			C.Z3_ast_vector_inc_ref(m.ctx.c, av)
			n := int(C.Z3_ast_vector_size(m.ctx.c, av))
			for i := 0; i < n; i++ {
				ce := C.Z3_ast_vector_get(m.ctx.c, av, C.uint(i))
				result = append(result, m.ctx.newExpr(ce))
			}
			C.Z3_ast_vector_dec_ref(m.ctx.c, av)
		}
	})
	runtime.KeepAlive(m)
	runtime.KeepAlive(s)
	return result
}

// String returns the model as a string.
func (m *Model) String() string {
	var res string
	m.ctx.do(func() {
		res = C.GoString(C.Z3_model_to_string(m.ctx.c, m.c))
	})
	runtime.KeepAlive(m)
	return res
}

// --- IC3/PDR extensions ---

// CheckAssumptions checks satisfiability under a set of assumptions.
// The assumptions are temporary — they are not added to the solver's assertion stack.
func (s *Z3Solver) CheckAssumptions(assumptions []Z3Expr) Z3CheckResult {
	cassumptions := make([]C.Z3_ast, len(assumptions))
	for i, a := range assumptions {
		cassumptions[i] = a.c
	}
	var r Z3CheckResult
	s.ctx.do(func() {
		var cap *C.Z3_ast
		if len(cassumptions) > 0 {
			cap = &cassumptions[0]
		}
		res := C.Z3_solver_check_assumptions(s.ctx.c, s.c, C.uint(len(cassumptions)), cap)
		r = Z3CheckResult(res)
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(assumptions)
	return r
}

// UnsatCore returns the unsat core from the last CheckAssumptions call that returned Unsat.
// The core is a subset of the assumptions that are sufficient to prove unsatisfiability.
func (s *Z3Solver) UnsatCore() []Z3Expr {
	var result []Z3Expr
	s.ctx.do(func() {
		vec := C.Z3_solver_get_unsat_core(s.ctx.c, s.c)
		C.Z3_ast_vector_inc_ref(s.ctx.c, vec)
		n := int(C.Z3_ast_vector_size(s.ctx.c, vec))
		result = make([]Z3Expr, n)
		for i := 0; i < n; i++ {
			ast := C.Z3_ast_vector_get(s.ctx.c, vec, C.uint(i))
			result[i] = s.ctx.newExpr(ast)
		}
		C.Z3_ast_vector_dec_ref(s.ctx.c, vec)
	})
	runtime.KeepAlive(s)
	return result
}

// NewZ3SolverForLogic creates a solver for a specific SMT logic (e.g., "QF_LIA" for quantifier-free linear integer arithmetic).
func NewZ3SolverForLogic(ctx *Z3Context, logic string) *Z3Solver {
	var s *Z3Solver
	ctx.do(func() {
		clogic := C.CString(logic)
		defer C.free(unsafe.Pointer(clogic))
		sym := C.Z3_mk_string_symbol(ctx.c, clogic)
		cs := C.Z3_mk_solver_for_logic(ctx.c, sym)
		C.Z3_solver_inc_ref(ctx.c, cs)
		s = &Z3Solver{ctx: ctx, c: cs}
	})
	//runtime.SetFinalizer(s, func(s *Z3Solver) {
	//	s.ctx.do(func() {
	//		C.Z3_solver_dec_ref(s.ctx.c, s.c)
	//	})
	//})
	return s
}

// Equal returns true if two expressions are structurally equal.
func (e Z3Expr) Equal(other Z3Expr) bool {
	var result bool
	e.ctx.do(func() {
		result = bool(C.Z3_is_eq_ast(e.ctx.c, e.c, other.c))
	})
	runtime.KeepAlive(e)
	runtime.KeepAlive(other)
	return result
}

// IsTrue returns true if the expression is the boolean constant true.
func (e Z3Expr) IsTrue() bool {
	var result bool
	e.ctx.do(func() {
		result = C.Z3_get_bool_value(e.ctx.c, e.c) == C.Z3_L_TRUE
	})
	runtime.KeepAlive(e)
	return result
}

// IsFalse returns true if the expression is the boolean constant false.
func (e Z3Expr) IsFalse() bool {
	var result bool
	e.ctx.do(func() {
		result = C.Z3_get_bool_value(e.ctx.c, e.c) == C.Z3_L_FALSE
	})
	runtime.KeepAlive(e)
	return result
}

// Substitute replaces expressions in e according to the from/to pairs.
func (ctx *Z3Context) Substitute(e Z3Expr, from, to []Z3Expr) Z3Expr {
	if len(from) != len(to) {
		panic("z3bridge: Substitute: from and to must have the same length")
	}
	cfrom := make([]C.Z3_ast, len(from))
	cto := make([]C.Z3_ast, len(to))
	for i := range from {
		cfrom[i] = from[i].c
		cto[i] = to[i].c
	}
	var r Z3Expr
	ctx.do(func() {
		var cfp, ctp *C.Z3_ast
		if len(cfrom) > 0 {
			cfp = &cfrom[0]
			ctp = &cto[0]
		}
		r = ctx.newExpr(C.Z3_substitute(ctx.c, e.c, C.uint(len(cfrom)), cfp, ctp))
	})
	runtime.KeepAlive(e)
	runtime.KeepAlive(from)
	runtime.KeepAlive(to)
	return r
}

// SetParam sets a solver parameter. The value is interpreted as a boolean
// ("true"/"false") if possible, then as an unsigned integer, otherwise as a symbol.
func (s *Z3Solver) SetParam(key, value string) {
	s.ctx.do(func() {
		ckey := C.CString(key)
		defer C.free(unsafe.Pointer(ckey))

		params := C.Z3_mk_params(s.ctx.c)
		C.Z3_params_inc_ref(s.ctx.c, params)

		keySym := C.Z3_mk_string_symbol(s.ctx.c, ckey)

		switch value {
		case "true":
			C.Z3_params_set_bool(s.ctx.c, params, keySym, C.bool(true))
		case "false":
			C.Z3_params_set_bool(s.ctx.c, params, keySym, C.bool(false))
		default:
			// Try to parse as uint, otherwise set as symbol.
			var uval uint64
			n, _ := fmt.Sscanf(value, "%d", &uval)
			if n == 1 {
				C.Z3_params_set_uint(s.ctx.c, params, keySym, C.uint(uval))
			} else {
				cval := C.CString(value)
				defer C.free(unsafe.Pointer(cval))
				valSym := C.Z3_mk_string_symbol(s.ctx.c, cval)
				C.Z3_params_set_symbol(s.ctx.c, params, keySym, valSym)
			}
		}

		C.Z3_solver_set_params(s.ctx.c, s.c, params)
		C.Z3_params_dec_ref(s.ctx.c, params)
	})
}

// ErrMsg is returned for Z3 errors that are caught.
type ErrMsg struct {
	Msg string
}

func (e *ErrMsg) Error() string { return fmt.Sprintf("z3: %s", e.Msg) }
