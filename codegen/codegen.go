// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_cpp.py.

// Package codegen provides a code generation framework for accumulating
// code strings with indentation, managing scoped contexts (globals,
// impls, members, locals, expr), and representing C++ types. It is a
// generic infrastructure layer used by higher-level code generators
// (e.g. cppgen, gogen).
package codegen

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// CodeWriter — anything that can accept code fragments.
// ---------------------------------------------------------------------------

// CodeWriter is the interface satisfied by CodeText and DeadCode.
type CodeWriter interface {
	Write(code interface{})
}

// ---------------------------------------------------------------------------
// CodeText — accumulates code fragments with indentation support.
// ---------------------------------------------------------------------------

// CodeText accumulates code fragments (strings or fmt.Stringer values).
// It replaces the Python CppText class.
type CodeText struct {
	Code []interface{}
}

// NewCodeText creates an empty CodeText.
func NewCodeText() *CodeText {
	return &CodeText{}
}

// Write appends a code fragment.
func (ct *CodeText) Write(code interface{}) {
	ct.Code = append(ct.Code, code)
}

// Get returns all accumulated code, with every line indented by indent
// levels (4 spaces each).
func (ct *CodeText) Get(indent int) string {
	prefix := strings.Repeat("    ", indent)
	var lines []string
	for _, c := range ct.Code {
		s := fmt.Sprint(c)
		for _, l := range strings.Split(s, "\n") {
			lines = append(lines, prefix+l)
		}
	}
	return strings.Join(lines, "\n")
}

// GetFile concatenates all code elements as strings with no extra
// formatting.  This mirrors CppText.get_file() for backward compat.
func (ct *CodeText) GetFile() string {
	var b strings.Builder
	for _, c := range ct.Code {
		fmt.Fprint(&b, c)
	}
	return b.String()
}

// Last returns the last element of Code, or nil.
func (ct *CodeText) Last() interface{} {
	if len(ct.Code) == 0 {
		return nil
	}
	return ct.Code[len(ct.Code)-1]
}

// RemoveLast removes and returns the last element of Code.
func (ct *CodeText) RemoveLast() interface{} {
	n := len(ct.Code)
	if n == 0 {
		return nil
	}
	v := ct.Code[n-1]
	ct.Code = ct.Code[:n-1]
	return v
}

// ---------------------------------------------------------------------------
// DeadCode — sentinel that panics on Write.
// ---------------------------------------------------------------------------

// DeadCode is a CodeWriter that panics on any write attempt.
type DeadCode struct{}

// Write panics — you should never emit code into a DeadCode scope.
func (dc *DeadCode) Write(code interface{}) {
	panic("codegen: cannot write code here (DeadCode)")
}

// ---------------------------------------------------------------------------
// Temp variable name generator (global, atomic counter).
// ---------------------------------------------------------------------------

var tempCounter int64

// GetTemp returns a unique temporary variable name like "tmp0", "tmp1", ...
func GetTemp() string {
	n := atomic.AddInt64(&tempCounter, 1) - 1
	return fmt.Sprintf("tmp%d", n)
}

// ResetTemp resets the temp counter (useful in tests or when entering a
// new context scope, mirroring the Python behavior).
func ResetTemp(val int64) {
	atomic.StoreInt64(&tempCounter, val)
}

// currentTemp returns the current temp counter value (for save/restore).
func currentTemp() int64 {
	return atomic.LoadInt64(&tempCounter)
}

// ---------------------------------------------------------------------------
// CodeContext — manages scoped code emission.
// ---------------------------------------------------------------------------

// CodeContext replaces the Python CppContext.  It holds code accumulators
// for several scopes: globals (header), impls (implementation), members
// (class body), locals (local variables), and expr (current expression).
type CodeContext struct {
	Globals        *CodeText
	Impls          *CodeText
	Members        CodeWriter // may be CodeText or DeadCode
	Locals         CodeWriter // may be CodeText or DeadCode
	Expr           *CodeText
	ClassName      string
	GlobalIncludes []string
	ImplIncludes   []string
	OnceGlobals    map[string]bool

	// saved state for Enter/Exit
	old            *CodeContext // previous context
	oldTempCounter int64
}

// NewCodeContext creates a fresh CodeContext (mirrors CppContext.__init__).
func NewCodeContext() *CodeContext {
	g := NewCodeText()
	return &CodeContext{
		Globals:     g,
		Impls:       NewCodeText(),
		Members:     g, // members default to globals
		Locals:      &DeadCode{},
		Expr:        NewCodeText(),
		OnceGlobals: make(map[string]bool),
	}
}

// currentContext is the package-level "active" context, mirroring the
// Python module-level `context` variable.
var currentContext *CodeContext

// CurrentContext returns the active CodeContext (may be nil).
func CurrentContext() *CodeContext {
	return currentContext
}

// Enter makes this context the current context.  It saves and restores
// the temp counter so that nested contexts get independent temp numbering.
func (cc *CodeContext) Enter() {
	cc.oldTempCounter = currentTemp()
	cc.old = currentContext
	currentContext = cc
}

// Exit restores the previous context and temp counter.
func (cc *CodeContext) Exit() {
	currentContext = cc.old
	ResetTemp(cc.oldTempCounter)
}

// ---------------------------------------------------------------------------
// Convenience functions — operate on the current context.
// ---------------------------------------------------------------------------

// AddGlobal appends code to the current global (header) scope.
func AddGlobal(code interface{}) {
	currentContext.Globals.Write(code)
}

// AddOnceGlobal appends code to the global scope only once (deduped by
// the string representation).
func AddOnceGlobal(code interface{}) {
	s := fmt.Sprint(code)
	if !currentContext.OnceGlobals[s] {
		currentContext.Globals.Write(code)
		currentContext.OnceGlobals[s] = true
	}
}

// AddImpl appends code to the implementation scope.
func AddImpl(code interface{}) {
	currentContext.Impls.Write(code)
}

// AddMember appends code to the current member scope.
func AddMember(code interface{}) {
	currentContext.Members.Write(code)
}

// AddLocal appends code to the current local scope.
func AddLocal(code interface{}) {
	currentContext.Locals.Write(code)
}

// AddExpr appends code to the current expression scope.
func AddExpr(code interface{}) {
	currentContext.Expr.Write(code)
}

// CurrentClassName returns the name of the current class (scope for
// member declarations), or "" if at global scope.
func CurrentClassName() string {
	return currentContext.ClassName
}

// AddHeader records a header include.
func AddHeader(name string) {
	currentContext.GlobalIncludes = append(currentContext.GlobalIncludes, name)
}

// ---------------------------------------------------------------------------
// Name resolution helpers.
// ---------------------------------------------------------------------------

// FullName returns "classname::membername" or just membername if
// classname is empty (global scope).
func FullName(classname, membername string) string {
	if classname == "" {
		return membername
	}
	return classname + "::" + membername
}

// RelName returns the relative name: if classname matches the current
// context classname then just membername, otherwise the full name.
func RelName(classname, membername string) string {
	if CurrentClassName() == classname {
		return membername
	}
	return FullName(classname, membername)
}
