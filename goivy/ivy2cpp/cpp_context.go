package ivy2cpp

import (
	"fmt"
	"os"
	"strings"
)

type CppFile struct {
	Filename string
	Mode     string
	Indent   int

	file *os.File
}

func NewCppFile(filename, mode string) (*CppFile, error) {
	if mode == "" {
		mode = "w"
	}
	flag := os.O_CREATE | os.O_WRONLY
	if strings.Contains(mode, "a") {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	f, err := os.OpenFile(filename, flag, 0o644)
	if err != nil {
		return nil, err
	}
	return &CppFile{Filename: filename, Mode: mode, file: f}, nil
}

func (f *CppFile) Close() error {
	if f == nil || f.file == nil {
		return nil
	}
	return f.file.Close()
}

func (f *CppFile) Write(code any) {
	if f == nil || f.file == nil {
		panic("ivy2cpp: write to closed CppFile")
	}
	lines := strings.Split(fmt.Sprint(code), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		_, _ = f.file.WriteString(strings.Repeat("    ", f.Indent))
		_, _ = f.file.WriteString(line)
		_, _ = f.file.WriteString("\n")
	}
}

type CppText struct {
	Code []any
}

func NewCppText() *CppText {
	return &CppText{}
}

func (t *CppText) Close() {}

func (t *CppText) Write(code any) {
	if t == nil {
		panic("ivy2cpp: write to nil CppText")
	}
	t.Code = append(t.Code, code)
}

func (t *CppText) Get(indent int) string {
	if t == nil {
		return ""
	}
	prefix := strings.Repeat("    ", indent)
	var out []string
	for _, code := range t.Code {
		for _, line := range strings.Split(fmt.Sprint(code), "\n") {
			out = append(out, prefix+line)
		}
	}
	return strings.Join(out, "\n")
}

func (t *CppText) GetFile() string {
	if t == nil {
		return ""
	}
	var b strings.Builder
	for _, code := range t.Code {
		b.WriteString(fmt.Sprint(code))
	}
	return b.String()
}

func (t *CppText) removeLast(code any) {
	if t == nil || len(t.Code) == 0 || t.Code[len(t.Code)-1] != code {
		panic("ivy2cpp: context stack mismatch")
	}
	t.Code = t.Code[:len(t.Code)-1]
}

type DeadCode struct{}

func (d *DeadCode) Write(code any) {
	panic("Cannot write code here")
}

type Context interface {
	Enter(*CppContext) func()
}

type CppContext struct {
	Globals        *CppText
	Impls          *CppText
	Members        *CppText
	Locals         *CppText
	Expr           *CppText
	Classname      string
	GlobalIncludes []string
	ImplIncludes   []string
	OnceGlobals    map[string]bool

	tempCounter int
}

func NewCppContext() *CppContext {
	globals := NewCppText()
	return &CppContext{
		Globals:     globals,
		Impls:       NewCppText(),
		Members:     globals,
		Expr:        NewCppText(),
		OnceGlobals: map[string]bool{},
	}
}

func (c *CppContext) Enter() func() {
	return func() {}
}

func (c *CppContext) AddGlobal(code any) {
	c.Globals.Write(code)
}

func (c *CppContext) AddOnceGlobal(code any) {
	key := fmt.Sprint(code)
	if c.OnceGlobals[key] {
		return
	}
	c.OnceGlobals[key] = true
	c.Globals.Write(code)
}

func (c *CppContext) AddImpl(code any) {
	c.Impls.Write(code)
}

func (c *CppContext) AddMember(code any) {
	c.Members.Write(code)
}

func (c *CppContext) AddLocal(code any) {
	if c.Locals == nil {
		(&DeadCode{}).Write(code)
		return
	}
	c.Locals.Write(code)
}

func (c *CppContext) AddExpr(code any) {
	c.Expr.Write(code)
}

func (c *CppContext) CurrentClassname() string {
	return c.Classname
}

func (c *CppContext) AddHeader(name string) {
	c.GlobalIncludes = append(c.GlobalIncludes, name)
}

func (c *CppContext) AddImplHeader(name string) {
	c.ImplIncludes = append(c.ImplIncludes, name)
}

func (c *CppContext) GetTemp() string {
	res := fmt.Sprintf("__temp__%d", c.tempCounter)
	c.tempCounter++
	return res
}

func (c *CppContext) Relname(classname, membername string) string {
	if c.CurrentClassname() == classname {
		return membername
	}
	return Fullname(classname, membername)
}

func Fullname(classname, membername string) string {
	if classname == "" {
		return membername
	}
	return classname + "::" + membername
}

type CppType interface {
	Instantiate(name string, initializer *CppText) string
	ShortName(ctx *CppContext) string
	LongName(ctx *CppContext) string
	Declare(ctx *CppContext)
}

type CppInt64 struct{}

func (CppInt64) Instantiate(name string, initializer *CppText) string {
	return instantiateCppType("long", name, initializer)
}
func (CppInt64) ShortName(ctx *CppContext) string { return "long" }
func (CppInt64) LongName(ctx *CppContext) string  { return "long" }
func (CppInt64) Declare(ctx *CppContext)          {}

type CppVoid struct{}

func (CppVoid) Instantiate(name string, initializer *CppText) string {
	return instantiateCppType("void", name, initializer)
}
func (CppVoid) ShortName(ctx *CppContext) string { return "void" }
func (CppVoid) LongName(ctx *CppContext) string  { return "void" }
func (CppVoid) Declare(ctx *CppContext)          {}

func instantiateCppType(typeName, name string, initializer *CppText) string {
	if initializer != nil {
		return fmt.Sprintf("%s %s=%s;", typeName, name, initializer.GetFile())
	}
	return fmt.Sprintf("%s %s;", typeName, name)
}

type CppClass struct {
	Classname string
	Baseclass string
	Members   *CppText

	parent string
}

func NewCppClass(ctx *CppContext, classname, baseclass string) *CppClass {
	c := &CppClass{
		Classname: classname,
		Baseclass: baseclass,
		Members:   NewCppText(),
		parent:    ctx.CurrentClassname(),
	}
	c.Declare(ctx)
	return c
}

func (c *CppClass) Declare(ctx *CppContext) {
	ctx.AddMember(c)
}

func (c *CppClass) Instantiate(name string, initializer *CppText) string {
	return instantiateCppType(c.ShortName(nil), name, initializer)
}

func (c *CppClass) ShortName(ctx *CppContext) string {
	if ctx != nil {
		return ctx.Relname(c.parent, c.Classname)
	}
	if c.parent == "" {
		return c.Classname
	}
	return Fullname(c.parent, c.Classname)
}

func (c *CppClass) LongName(ctx *CppContext) string {
	return Fullname(c.parent, c.Classname)
}

func (c *CppClass) Enter(ctx *CppContext) func() {
	oldMembers := ctx.Members
	oldClassname := ctx.Classname
	ctx.Members = c.Members
	ctx.Classname = Fullname(ctx.Classname, c.Classname)
	return func() {
		ctx.Members = oldMembers
		ctx.Classname = oldClassname
	}
}

func (c *CppClass) String() string {
	base := ""
	if c.Baseclass != "" {
		base = " : public " + c.Baseclass + " "
	}
	return "class " + c.Classname + base + "{\npublic:\n" + c.Members.Get(1) + "\n};"
}

type CppClassName struct {
	Classname string
}

func (c *CppClassName) Enter(ctx *CppContext) func() {
	oldClassname := ctx.Classname
	ctx.Classname = c.Classname
	return func() {
		ctx.Classname = oldClassname
	}
}

type cppArrFunType struct {
	CppType CppType
	Name    string
	parent  string
}

func (t *cppArrFunType) declare(ctx *CppContext, obj any) {
	t.parent = ctx.CurrentClassname()
	if t.Name != "" {
		ctx.AddMember(obj)
	}
}

func (t *cppArrFunType) shortName(ctx *CppContext, longName string) string {
	if t.Name == "" {
		return longName
	}
	if ctx == nil {
		return Fullname(t.parent, t.Name)
	}
	return ctx.Relname(t.parent, t.Name)
}

func (t *cppArrFunType) initstr(initializer *CppText) string {
	if initializer != nil {
		return "=" + initializer.GetFile() + ";"
	}
	return ";"
}

type CppArray struct {
	CppType CppType
	Dims    []int
	Name    string

	parent string
}

func NewCppArray(ctx *CppContext, cpptype CppType, dims []int, name string) *CppArray {
	a := &CppArray{CppType: cpptype, Dims: append([]int(nil), dims...), Name: name}
	a.Declare(ctx)
	return a
}

func (a *CppArray) Declare(ctx *CppContext) {
	a.parent = ctx.CurrentClassname()
	if a.Name != "" {
		ctx.AddMember(a)
	}
}

func (a *CppArray) suffix() string {
	var b strings.Builder
	for _, d := range a.Dims {
		b.WriteString(fmt.Sprintf("[%d]", d))
	}
	return b.String()
}

func (a *CppArray) InstantiateLong(name string, initializer *CppText) string {
	return fmt.Sprintf("%s %s%s%s", a.CppType.ShortName(nil), name, a.suffix(), a.initstr(initializer))
}

func (a *CppArray) initstr(initializer *CppText) string {
	if initializer != nil {
		return "=" + initializer.GetFile() + ";"
	}
	return ";"
}

func (a *CppArray) Instantiate(name string, initializer *CppText) string {
	if a.Name != "" && initializer == nil {
		return instantiateCppType(a.ShortName(nil), name, nil)
	}
	return a.InstantiateLong(name, initializer)
}

func (a *CppArray) ShortName(ctx *CppContext) string {
	if a.Name == "" {
		return a.LongName(ctx)
	}
	if ctx == nil {
		return Fullname(a.parent, a.Name)
	}
	return ctx.Relname(a.parent, a.Name)
}

func (a *CppArray) LongName(ctx *CppContext) string {
	if a.Name == "" {
		panic("C++ array and function types cannot be named!")
	}
	return Fullname(a.parent, a.Name)
}

func (a *CppArray) String() string {
	return "typedef " + a.InstantiateLong(a.Name, nil)
}

type CppFunction struct {
	CppType  CppType
	ArgTypes []CppType
	Name     string

	parent string
}

func NewCppFunction(ctx *CppContext, cpptype CppType, argtypes []CppType, name string) *CppFunction {
	f := &CppFunction{CppType: cpptype, ArgTypes: append([]CppType(nil), argtypes...), Name: name}
	f.Declare(ctx)
	return f
}

func (f *CppFunction) Declare(ctx *CppContext) {
	f.parent = ctx.CurrentClassname()
	if f.Name != "" {
		ctx.AddMember(f)
	}
}

func (f *CppFunction) suffix(names []string) string {
	if names == nil {
		names = make([]string, len(f.ArgTypes))
		for i := range f.ArgTypes {
			names[i] = fmt.Sprintf("arg%d", i)
		}
	}
	parts := make([]string, len(f.ArgTypes))
	for i, typ := range f.ArgTypes {
		parts[i] = strings.TrimSuffix(typ.Instantiate(names[i], nil), ";")
	}
	return "(" + strings.Join(parts, ",") + ")"
}

func (f *CppFunction) initstr(initializer *CppText) string {
	if initializer != nil {
		return "{\n" + initializer.Get(1) + "\n}"
	}
	return ";"
}

func (f *CppFunction) InstantiateLong(name string, initializer *CppText) string {
	return fmt.Sprintf("%s %s%s%s", f.CppType.ShortName(nil), name, f.suffix(nil), f.initstr(initializer))
}

func (f *CppFunction) Instantiate(name string, initializer *CppText) string {
	if f.Name != "" && initializer == nil {
		return instantiateCppType(f.ShortName(nil), name, nil)
	}
	return f.InstantiateLong(name, initializer)
}

func (f *CppFunction) ShortName(ctx *CppContext) string {
	if f.Name == "" {
		return f.LongName(ctx)
	}
	if ctx == nil {
		return Fullname(f.parent, f.Name)
	}
	return ctx.Relname(f.parent, f.Name)
}

func (f *CppFunction) LongName(ctx *CppContext) string {
	if f.Name == "" {
		panic("C++ array and function types cannot be named!")
	}
	return Fullname(f.parent, f.Name)
}

func (f *CppFunction) String() string {
	return "typedef " + f.InstantiateLong(f.Name, nil)
}

type CppReference struct {
	CppType CppType
	Name    string
	Const   bool

	parent string
}

func NewCppReference(ctx *CppContext, cpptype CppType, name string, constRef bool) *CppReference {
	r := &CppReference{CppType: cpptype, Name: name, Const: constRef}
	r.Declare(ctx)
	return r
}

func (r *CppReference) Declare(ctx *CppContext) {
	r.parent = ctx.CurrentClassname()
	if r.Name != "" {
		ctx.AddMember(r)
	}
}

func (r *CppReference) prefix() string {
	if r.Const {
		return "const &"
	}
	return "&"
}

func (r *CppReference) InstantiateLong(name string, initializer *CppText) string {
	init := ";"
	if initializer != nil {
		init = "=" + initializer.GetFile() + ";"
	}
	return fmt.Sprintf("%s %s%s%s", r.CppType.ShortName(nil), r.prefix(), name, init)
}

func (r *CppReference) Instantiate(name string, initializer *CppText) string {
	if r.Name != "" && initializer == nil {
		return instantiateCppType(r.ShortName(nil), name, nil)
	}
	return r.InstantiateLong(name, initializer)
}

func (r *CppReference) ShortName(ctx *CppContext) string {
	if r.Name == "" {
		return r.LongName(ctx)
	}
	if ctx == nil {
		return Fullname(r.parent, r.Name)
	}
	return ctx.Relname(r.parent, r.Name)
}

func (r *CppReference) LongName(ctx *CppContext) string {
	return r.CppType.ShortName(ctx) + "&"
}

func (r *CppReference) String() string {
	return "typedef " + r.InstantiateLong(r.Name, nil)
}

type CppVector struct {
	CppType CppType
	Name    string

	parent string
}

func NewCppVector(ctx *CppContext, cpptype CppType, name string) *CppVector {
	v := &CppVector{CppType: cpptype, Name: name}
	v.Declare(ctx)
	return v
}

func (v *CppVector) Declare(ctx *CppContext) {
	v.parent = ctx.CurrentClassname()
	ctx.AddHeader("<vector>")
	if v.Name != "" {
		ctx.AddMember(v)
	}
}

func (v *CppVector) Instantiate(name string, initializer *CppText) string {
	return instantiateCppType(v.ShortName(nil), name, initializer)
}

func (v *CppVector) ShortName(ctx *CppContext) string {
	if v.Name == "" {
		return v.LongName(ctx)
	}
	if ctx == nil {
		return Fullname(v.parent, v.Name)
	}
	return ctx.Relname(v.parent, v.Name)
}

func (v *CppVector) LongName(ctx *CppContext) string {
	return "std::vector<" + v.CppType.ShortName(ctx) + "> "
}

func (v *CppVector) String() string {
	return "typedef " + v.LongName(nil) + v.Name + ";"
}

type TypeDef struct {
	OldType CppType
	NewName string

	parent string
}

func NewTypeDef(ctx *CppContext, oldtype CppType, newname string) *TypeDef {
	td := &TypeDef{OldType: oldtype, NewName: newname}
	td.Declare(ctx)
	return td
}

func (t *TypeDef) Declare(ctx *CppContext) {
	t.parent = ctx.CurrentClassname()
	if t.NewName != "" {
		ctx.AddMember(t)
	}
}

func (t *TypeDef) Instantiate(name string, initializer *CppText) string {
	return instantiateCppType(t.ShortName(nil), name, initializer)
}

func (t *TypeDef) ShortName(ctx *CppContext) string {
	if t.NewName == "" {
		return t.LongName(ctx)
	}
	if ctx == nil {
		return Fullname(t.parent, t.NewName)
	}
	return ctx.Relname(t.parent, t.NewName)
}

func (t *TypeDef) LongName(ctx *CppContext) string {
	return t.OldType.LongName(ctx)
}

func (t *TypeDef) String() string {
	return fmt.Sprintf("typedef %s %s;", t.LongName(nil), t.NewName)
}

type CppMember struct {
	CppType CppType
	Name    string
	Static  bool
	Inline  bool

	Parent      string
	Initializer *CppText
	scope       *CppText
}

func NewCppMember(ctx *CppContext, cpptype CppType, name string, static, inline bool) *CppMember {
	m := &CppMember{CppType: cpptype, Name: name, Static: static, Inline: inline}
	m.Parent = ctx.CurrentClassname()
	m.scope = ctx.Members
	m.scope.Write(m)
	return m
}

func (m *CppMember) String() string {
	prefix := ""
	if m.Static {
		prefix += "static "
	}
	if m.Inline {
		prefix += "inline "
	}
	return prefix + m.CppType.Instantiate(m.Name, m.Initializer)
}

func (m *CppMember) Ref(ctx *CppContext) string {
	return ctx.Relname(m.Parent, m.Name)
}

func (m *CppMember) Enter(ctx *CppContext) func() {
	m.scope.removeLast(m)
	if m.Initializer != nil {
		panic("ivy2cpp: member initializer already active")
	}
	m.Initializer = NewCppText()
	if _, ok := m.CppType.(*CppFunction); ok {
		oldLocals := ctx.Locals
		ctx.Locals = m.Initializer
		return func() {
			ctx.Locals = oldLocals
			m.scope.Write(m)
		}
	}
	oldExpr := ctx.Expr
	ctx.Expr = m.Initializer
	return func() {
		ctx.Expr = oldExpr
		m.scope.Write(m)
	}
}

type CppLocal struct {
	CppMember
}

func NewCppLocal(ctx *CppContext, cpptype CppType, name string) *CppLocal {
	if ctx.Locals == nil {
		(&DeadCode{}).Write("")
	}
	l := &CppLocal{}
	l.CppType = cpptype
	l.Name = name
	l.Parent = ctx.CurrentClassname()
	l.scope = ctx.Locals
	l.scope.Write(l)
	return l
}

func (l *CppLocal) String() string {
	return l.CppMember.String()
}

type CppScope struct {
	Code *CppText
}

func NewCppScope(ctx *CppContext) *CppScope {
	s := &CppScope{Code: NewCppText()}
	ctx.AddLocal(s)
	return s
}

func (s *CppScope) Enter(ctx *CppContext) func() {
	oldLocals := ctx.Locals
	ctx.Locals = s.Code
	return func() {
		ctx.Locals = oldLocals
	}
}

func (s *CppScope) String() string {
	return "{\n" + s.Code.Get(1) + "\n}"
}

func (s *CppScope) Ref(ctx *CppContext) string {
	panic("cannot reference a C++ scope")
}
