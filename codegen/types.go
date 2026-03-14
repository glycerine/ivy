// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_cpp.py.

package codegen

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// CppTyper — interface for C++ type representations.
// ---------------------------------------------------------------------------

// CppTyper is the interface every C++ type must implement.
type CppTyper interface {
	ShortName() string
	LongName() string
	Declare()
	Instantiate(name string, initializer *CodeText) string
}

// ---------------------------------------------------------------------------
// CppInt64
// ---------------------------------------------------------------------------

// CppInt64 represents the C++ "long" (64-bit signed integer) type.
type CppInt64 struct{}

func (t *CppInt64) ShortName() string                                 { return "long" }
func (t *CppInt64) LongName() string                                  { return "long" }
func (t *CppInt64) Declare()                                          {} // built-in
func (t *CppInt64) Instantiate(name string, init *CodeText) string {
	return defaultInstantiate(t, name, init)
}

// ---------------------------------------------------------------------------
// CppVoid
// ---------------------------------------------------------------------------

// CppVoid represents the C++ void type.
type CppVoid struct{}

func (t *CppVoid) ShortName() string                                 { return "void" }
func (t *CppVoid) LongName() string                                  { return "void" }
func (t *CppVoid) Declare()                                          {} // built-in
func (t *CppVoid) Instantiate(name string, init *CodeText) string {
	return defaultInstantiate(t, name, init)
}

// defaultInstantiate is the shared logic for simple type instantiation
// (mirrors CppType.instantiate in Python).
func defaultInstantiate(t CppTyper, name string, init *CodeText) string {
	if init != nil {
		return fmt.Sprintf("%s %s=%s;", t.ShortName(), name, init.GetFile())
	}
	return fmt.Sprintf("%s %s;", t.ShortName(), name)
}

// ---------------------------------------------------------------------------
// CppClass
// ---------------------------------------------------------------------------

// CppClass represents a C++ class with members.
type CppClass struct {
	MembersText *CodeText
	Name        string
	Parent      string // enclosing classname at construction time
	BaseClass   string // "" means no base class

	// saved state for Enter/Exit
	oldMembers   CodeWriter
	oldClassName string
}

// NewCppClass creates a CppClass and immediately declares it as a member
// of the current context.
func NewCppClass(classname string, baseclass string) *CppClass {
	c := &CppClass{
		MembersText: NewCodeText(),
		Name:        classname,
		Parent:      CurrentClassName(),
		BaseClass:   baseclass,
	}
	c.Declare()
	return c
}

func (c *CppClass) Declare()   { AddMember(c) }
func (c *CppClass) ShortName() string { return RelName(c.Parent, c.Name) }
func (c *CppClass) LongName() string  { return FullName(c.Parent, c.Name) }

func (c *CppClass) Instantiate(name string, init *CodeText) string {
	return defaultInstantiate(c, name, init)
}

// String renders the class declaration.
func (c *CppClass) String() string {
	base := ""
	if c.BaseClass != "" {
		base = " : public " + c.BaseClass + " "
	}
	// Temporarily enter the class context for proper name resolution
	// inside member rendering.
	c.Enter()
	body := c.MembersText.Get(1)
	c.Exit()
	return "class " + c.Name + base + "{\npublic:\n" + body + "\n};"
}

// Enter opens the class scope: members and classname are redirected.
func (c *CppClass) Enter() {
	ctx := currentContext
	c.oldMembers = ctx.Members
	c.oldClassName = ctx.ClassName
	ctx.Members = c.MembersText
	if ctx.ClassName != "" {
		ctx.ClassName = ctx.ClassName + "::" + c.Name
	} else {
		ctx.ClassName = c.Name
	}
}

// Exit restores the previous members/classname.
func (c *CppClass) Exit() {
	ctx := currentContext
	ctx.Members = c.oldMembers
	ctx.ClassName = c.oldClassName
}

// ---------------------------------------------------------------------------
// CppClassName — context that temporarily sets the current class name.
// ---------------------------------------------------------------------------

// CppClassName is a lightweight context that only changes the current
// class name (for backward compatibility).
type CppClassName struct {
	name         string
	oldClassName string
}

func NewCppClassName(classname string) *CppClassName {
	return &CppClassName{name: classname}
}

func (cn *CppClassName) Enter() {
	cn.oldClassName = currentContext.ClassName
	currentContext.ClassName = cn.name
}

func (cn *CppClassName) Exit() {
	currentContext.ClassName = cn.oldClassName
}

// ---------------------------------------------------------------------------
// CppArrFunType — base for array/function/reference types.
// ---------------------------------------------------------------------------

// CppArrFunType is the base for CppArray, CppFunction, and CppReference.
type CppArrFunType struct {
	CppType  CppTyper // element / return type
	TypeName string   // typedef name, "" if anonymous
	Parent   string   // enclosing classname at declare time

	// self points to the outer (embedding) struct so that Declare()
	// registers an object whose String() method is the subtype's.
	self interface{}

	// Hooks for subtypes to override behaviour.
	suffixFn   func() string
	prefixFn   func() string
	initFn     func(init *CodeText) string
	longNameFn func() string // override LongName for subtypes like CppReference
}

func (a *CppArrFunType) ShortName() string {
	if a.TypeName != "" {
		return RelName(a.Parent, a.TypeName)
	}
	return a.longName()
}

func (a *CppArrFunType) longName() string {
	if a.longNameFn != nil {
		return a.longNameFn()
	}
	if a.TypeName == "" {
		panic("codegen: anonymous array/function types have no long name")
	}
	return FullName(a.Parent, a.TypeName)
}

func (a *CppArrFunType) LongName() string {
	return a.longName()
}

func (a *CppArrFunType) Declare() {
	a.Parent = CurrentClassName()
	if a.TypeName != "" {
		obj := a.self
		if obj == nil {
			obj = a
		}
		AddMember(obj)
	}
}

func (a *CppArrFunType) suffix() string {
	if a.suffixFn != nil {
		return a.suffixFn()
	}
	return ""
}

func (a *CppArrFunType) prefix() string {
	if a.prefixFn != nil {
		return a.prefixFn()
	}
	return ""
}

func (a *CppArrFunType) initStr(init *CodeText) string {
	if a.initFn != nil {
		return a.initFn(init)
	}
	if init != nil {
		return fmt.Sprintf("=%s;", init.GetFile())
	}
	return ";"
}

func (a *CppArrFunType) InstantiateLong(name string, init *CodeText) string {
	return fmt.Sprintf("%s %s%s%s%s", a.CppType.ShortName(), a.prefix(), name, a.suffix(), a.initStr(init))
}

func (a *CppArrFunType) Instantiate(name string, init *CodeText) string {
	if a.TypeName != "" && init == nil {
		return defaultInstantiate(a, name, nil)
	}
	return a.InstantiateLong(name, init)
}

// ---------------------------------------------------------------------------
// CppArray
// ---------------------------------------------------------------------------

// CppArray represents an array of cpptype with given dimensions.
type CppArray struct {
	CppArrFunType
	Dims []int
}

func NewCppArray(cpptype CppTyper, dims []int, name string) *CppArray {
	a := &CppArray{
		CppArrFunType: CppArrFunType{
			CppType:  cpptype,
			TypeName: name,
		},
		Dims: dims,
	}
	a.self = a
	a.suffixFn = a.arraySuffix
	a.Declare()
	return a
}

func (a *CppArray) arraySuffix() string {
	var b strings.Builder
	for _, d := range a.Dims {
		fmt.Fprintf(&b, "[%d]", d)
	}
	return b.String()
}

func (a *CppArray) String() string {
	return "typedef " + a.InstantiateLong(a.TypeName, nil)
}

// ---------------------------------------------------------------------------
// CppFunction
// ---------------------------------------------------------------------------

// CppFunction represents a function type.
type CppFunction struct {
	CppArrFunType
	ArgTypes []CppTyper
}

func NewCppFunction(rettype CppTyper, argtypes []CppTyper, name string) *CppFunction {
	f := &CppFunction{
		CppArrFunType: CppArrFunType{
			CppType:  rettype,
			TypeName: name,
		},
		ArgTypes: argtypes,
	}
	f.self = f
	f.suffixFn = f.funcSuffix
	f.initFn = f.funcInitStr
	f.Declare()
	return f
}

func (f *CppFunction) funcSuffix() string {
	parts := make([]string, len(f.ArgTypes))
	for i, t := range f.ArgTypes {
		argname := fmt.Sprintf("arg%d", i)
		decl := t.Instantiate(argname, nil)
		// Remove trailing semicolon from the declaration.
		decl = strings.TrimSuffix(decl, ";")
		parts[i] = decl
	}
	return "(" + strings.Join(parts, ",") + ")"
}

func (f *CppFunction) funcInitStr(init *CodeText) string {
	if init != nil {
		return "{\n" + init.Get(1) + "\n}"
	}
	return ";"
}

func (f *CppFunction) String() string {
	return "typedef " + f.InstantiateLong(f.TypeName, nil)
}

// ---------------------------------------------------------------------------
// CppReference
// ---------------------------------------------------------------------------

// CppReference represents a C++ reference type (cpptype&).
type CppReference struct {
	CppArrFunType
	Const bool
}

func NewCppReference(cpptype CppTyper, name string, isConst bool) *CppReference {
	r := &CppReference{
		CppArrFunType: CppArrFunType{
			CppType:  cpptype,
			TypeName: name,
		},
		Const: isConst,
	}
	r.self = r
	r.prefixFn = r.refPrefix
	r.longNameFn = r.refLongName
	r.Declare()
	return r
}

func (r *CppReference) refLongName() string {
	return r.CppType.ShortName() + "&"
}

func (r *CppReference) refPrefix() string {
	if r.Const {
		return "const &"
	}
	return "&"
}

// LongName is handled by refLongName via the longNameFn hook.

// ---------------------------------------------------------------------------
// TypeDef
// ---------------------------------------------------------------------------

// TypeDef represents "typedef oldtype newname;".
type TypeDef struct {
	OldType CppTyper
	TDName  string
	parent  string
}

func NewTypeDef(oldtype CppTyper, name string) *TypeDef {
	td := &TypeDef{OldType: oldtype, TDName: name}
	td.Declare()
	return td
}

func (td *TypeDef) Declare() {
	td.parent = CurrentClassName()
	if td.TDName != "" {
		AddMember(td)
	}
}

func (td *TypeDef) ShortName() string {
	if td.TDName != "" {
		return RelName(td.parent, td.TDName)
	}
	return td.LongName()
}

func (td *TypeDef) LongName() string {
	return td.OldType.LongName()
}

func (td *TypeDef) String() string {
	return fmt.Sprintf("typedef %s %s;", td.LongName(), td.TDName)
}

func (td *TypeDef) Instantiate(name string, init *CodeText) string {
	return defaultInstantiate(td, name, init)
}

// ---------------------------------------------------------------------------
// CppVector
// ---------------------------------------------------------------------------

// CppVector represents std::vector<cpptype>.
type CppVector struct {
	ElemType CppTyper
	VecName  string
	parent   string
}

func NewCppVector(cpptype CppTyper, name string) *CppVector {
	v := &CppVector{ElemType: cpptype, VecName: name}
	v.Declare()
	return v
}

func (v *CppVector) Declare() {
	v.parent = CurrentClassName()
	AddHeader("<vector>")
	if v.VecName != "" {
		AddMember(v)
	}
}

func (v *CppVector) ShortName() string {
	if v.VecName != "" {
		return RelName(v.parent, v.VecName)
	}
	return v.LongName()
}

func (v *CppVector) LongName() string {
	return "std::vector<" + v.ElemType.ShortName() + "> "
}

func (v *CppVector) String() string {
	return fmt.Sprintf("typedef %s%s;", v.LongName(), v.VecName)
}

func (v *CppVector) Instantiate(name string, init *CodeText) string {
	return defaultInstantiate(v, name, init)
}

// ---------------------------------------------------------------------------
// CppMember
// ---------------------------------------------------------------------------

// CppMember represents a member of a C++ class.  When used as a context
// (Enter/Exit), it captures an initializer for the member.
type CppMember struct {
	MType       CppTyper
	MName       string
	MParent     string
	Initializer *CodeText

	scope    func() CodeWriter // returns the scope to write into
	oldExpr  *CodeText
	oldLocal CodeWriter
}

// NewCppMember creates a member and writes it to the current member scope.
func NewCppMember(cpptype CppTyper, name string) *CppMember {
	m := &CppMember{
		MType:   cpptype,
		MName:   name,
		MParent: CurrentClassName(),
		scope:   memberScope,
	}
	m.getScope().Write(m)
	return m
}

func memberScope() CodeWriter { return currentContext.Members }
func localScope() CodeWriter  { return currentContext.Locals }

func (m *CppMember) getScope() CodeWriter {
	return m.scope()
}

func (m *CppMember) String() string {
	return m.MType.Instantiate(m.MName, m.Initializer)
}

// Ref returns the relative name for this member in the current context.
func (m *CppMember) Ref() string {
	return RelName(m.MParent, m.MName)
}

// Enter begins the initializer capture.  For function types, it captures
// into locals; otherwise into expr.
func (m *CppMember) Enter() {
	scope := m.getScope()
	ct, ok := scope.(*CodeText)
	if !ok {
		panic("codegen: CppMember.Enter: scope is not a *CodeText")
	}
	last := ct.Last()
	if last != m {
		panic("codegen: CppMember.Enter: member is not the last element in scope")
	}
	if m.Initializer != nil {
		panic("codegen: CppMember.Enter: initializer already set")
	}
	m.Initializer = NewCodeText()
	ct.RemoveLast()

	if _, isFunc := m.MType.(*CppFunction); isFunc {
		m.oldLocal = currentContext.Locals
		currentContext.Locals = m.Initializer
	} else {
		m.oldExpr = currentContext.Expr
		currentContext.Expr = m.Initializer
	}
}

// Exit restores the previous scope and appends this member back.
func (m *CppMember) Exit() {
	if _, isFunc := m.MType.(*CppFunction); isFunc {
		currentContext.Locals = m.oldLocal
	} else {
		currentContext.Expr = m.oldExpr
	}
	scope := m.getScope()
	ct, ok := scope.(*CodeText)
	if !ok {
		panic("codegen: CppMember.Exit: scope is not a *CodeText")
	}
	ct.Write(m)
}

// ---------------------------------------------------------------------------
// CppLocal
// ---------------------------------------------------------------------------

// CppLocal is like CppMember but writes to the local scope.
func NewCppLocal(cpptype CppTyper, name string) *CppMember {
	m := &CppMember{
		MType:   cpptype,
		MName:   name,
		MParent: CurrentClassName(),
		scope:   localScope,
	}
	m.getScope().Write(m)
	return m
}

// ---------------------------------------------------------------------------
// CppScope
// ---------------------------------------------------------------------------

// CppScope represents a nested { } scope inside a function body.
type CppScope struct {
	ScopeCode *CodeText
	oldLocals CodeWriter
}

// NewCppScope creates a scope and adds it to the current local scope.
func NewCppScope() *CppScope {
	s := &CppScope{ScopeCode: NewCodeText()}
	AddLocal(s)
	return s
}

func (s *CppScope) String() string {
	return "{\n" + s.ScopeCode.Get(1) + "\n}"
}

// Enter redirects locals to this scope's code text.
func (s *CppScope) Enter() {
	s.oldLocals = currentContext.Locals
	currentContext.Locals = s.ScopeCode
}

// Exit restores the previous locals.
func (s *CppScope) Exit() {
	currentContext.Locals = s.oldLocals
}
