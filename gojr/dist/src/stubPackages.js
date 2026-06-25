export function stubSourcePackageFiles(importPath) {
    switch (importPath) {
        case "github.com/klauspost/cpuid/v2":
            return [{
                    filename: "gojr:stub/github.com/klauspost/cpuid/v2/cpuid.go",
                    source: klauspostCpuidV2StubSource
                }];
        case "github.com/gopherjs/gopherjs/js":
            return [{
                    filename: "gojr:stub/github.com/gopherjs/gopherjs/js/js.go",
                    source: gopherjsJSStubSource
                }];
        case "github.com/jtolds/gls":
            return [{
                    filename: "gojr:stub/github.com/jtolds/gls/gls.go",
                    source: jtoldsGlsStubSource
                }];
        case "runtime/pprof":
            return [{
                    filename: "gojr:stub/runtime/pprof/pprof.go",
                    source: runtimePprofStubSource
                }];
        default:
            return undefined;
    }
}
export function isStubSourcePackageImport(importPath) {
    return stubSourcePackageFiles(importPath) !== undefined;
}
export function isStubSourcePackageStandardLibrary(importPath) {
    return importPath === "runtime/pprof";
}
const klauspostCpuidV2StubSource = `package cpuid

// FeatureID is the ID of a specific CPU feature. Go-junior deliberately
// exposes a deterministic JavaScript target profile rather than probing native
// hardware.
type FeatureID int

const (
  UNKNOWN FeatureID = -1
  AVX2 FeatureID = 8
  AVX512F FeatureID = 16
)

type CPUInfo struct{}

var CPU CPUInfo

func (c CPUInfo) Supports(features ...FeatureID) bool {
  return false
}
`;
const gopherjsJSStubSource = `package js

// Object is a deterministic Go-junior host placeholder for GopherJS JavaScript
// values. The real package relies on GopherJS compiler rewrites, so Go-junior
// provides ordinary Go bodies with conservative no-op behavior.
type Object struct{}

func (o *Object) Get(key string) *Object { return Undefined }
func (o *Object) Set(key string, value any) {}
func (o *Object) Delete(key string) {}
func (o *Object) Length() int { return 0 }
func (o *Object) Index(i int) *Object { return Undefined }
func (o *Object) SetIndex(i int, value any) {}
func (o *Object) Call(name string, args ...any) *Object { return Undefined }
func (o *Object) Invoke(args ...any) *Object { return Undefined }
func (o *Object) New(args ...any) *Object { return &Object{} }
func (o *Object) Bool() bool { return false }
func (o *Object) String() string { return "" }
func (o *Object) Int() int { return 0 }
func (o *Object) Int64() int64 { return 0 }
func (o *Object) Uint64() uint64 { return 0 }
func (o *Object) Float() float64 { return 0 }
func (o *Object) Interface() any { return nil }
func (o *Object) Unsafe() uintptr { return 0 }

type Error struct {
  *Object
}

func (err *Error) Error() string { return "JavaScript error" }
func (err *Error) Stack() string { return "" }

var Global = &Object{}
var Module = &Object{}
var Undefined = &Object{}

func Debugger() {}
func InternalObject(i any) *Object { return nil }
func MakeFunc(fn func(this *Object, arguments []*Object) any) *Object { return &Object{} }
func Keys(o *Object) []string { return nil }
func MakeWrapper(i any) *Object { return &Object{} }
func MakeFullWrapper(i any) *Object { return &Object{} }
func NewArrayBuffer(b []byte) *Object { return &Object{} }

type M map[string]any
type S []any
`;
const jtoldsGlsStubSource = `package gls

type Values map[interface{}]interface{}

type ContextKey struct{ id uint64 }

type ContextManager struct {
  stack []Values
}

var (
  keyCounter uint64
  mgrRegistry = make(map[*ContextManager]bool)
)

func GenSym() ContextKey {
  keyCounter++
  return ContextKey{id: keyCounter}
}

func NewContextManager() *ContextManager {
  mgr := &ContextManager{}
  mgrRegistry[mgr] = true
  return mgr
}

func (m *ContextManager) Unregister() {
  delete(mgrRegistry, m)
}

func (m *ContextManager) SetValues(newValues Values, contextCall func()) {
  if len(newValues) == 0 {
    contextCall()
    return
  }
  state := make(Values)
  if current := m.getValues(); current != nil {
    for key, value := range current {
      state[key] = value
    }
  }
  for key, value := range newValues {
    state[key] = value
  }
  m.stack = append(m.stack, state)
  defer func() {
    m.stack = m.stack[:len(m.stack)-1]
  }()
  contextCall()
}

func (m *ContextManager) GetValue(key interface{}) (value interface{}, ok bool) {
  state := m.getValues()
  if state == nil {
    return nil, false
  }
  value, ok = state[key]
  return value, ok
}

func (m *ContextManager) getValues() Values {
  if len(m.stack) == 0 {
    return nil
  }
  return m.stack[len(m.stack)-1]
}

func Go(cb func()) {
  var managers []*ContextManager
  var values []Values
  for mgr := range mgrRegistry {
    current := mgr.getValues()
    if len(current) == 0 {
      continue
    }
    copied := make(Values, len(current))
    for key, value := range current {
      copied[key] = value
    }
    managers = append(managers, mgr)
    values = append(values, copied)
  }
  go func() {
    for i, mgr := range managers {
      mgr.stack = append(mgr.stack, values[i])
    }
    defer func() {
      for i := len(managers)-1; i >= 0; i-- {
        mgr := managers[i]
        mgr.stack = mgr.stack[:len(mgr.stack)-1]
      }
    }()
    cb()
  }()
}

func GetGoroutineId() (gid uint, ok bool) {
  return 1, true
}

func EnsureGoroutineId(cb func(gid uint)) {
  cb(1)
}
`;
const runtimePprofStubSource = `package pprof

import (
  "context"
  "io"
)

type LabelSet struct{}

type Profile struct{}

func NewProfile(name string) *Profile {
  panic("gojr error: runtime/pprof.NewProfile not implemented")
}

func Lookup(name string) *Profile {
  panic("gojr error: runtime/pprof.Lookup not implemented")
}

func Profiles() []*Profile {
  panic("gojr error: runtime/pprof.Profiles not implemented")
}

func (p *Profile) Name() string {
  panic("gojr error: runtime/pprof.(*Profile).Name not implemented")
}

func (p *Profile) Count() int {
  panic("gojr error: runtime/pprof.(*Profile).Count not implemented")
}

func (p *Profile) Add(value any, skip int) {
  panic("gojr error: runtime/pprof.(*Profile).Add not implemented")
}

func (p *Profile) Remove(value any) {
  panic("gojr error: runtime/pprof.(*Profile).Remove not implemented")
}

func (p *Profile) WriteTo(w io.Writer, debug int) error {
  panic("gojr error: runtime/pprof.(*Profile).WriteTo not implemented")
}

func WriteHeapProfile(w io.Writer) error {
  panic("gojr error: runtime/pprof.WriteHeapProfile not implemented")
}

func StartCPUProfile(w io.Writer) error {
  panic("gojr error: runtime/pprof.StartCPUProfile not implemented")
}

func StopCPUProfile() {
  panic("gojr error: runtime/pprof.StopCPUProfile not implemented")
}

func WithLabels(ctx context.Context, labels LabelSet) context.Context {
  panic("gojr error: runtime/pprof.WithLabels not implemented")
}

func Labels(args ...string) LabelSet {
  panic("gojr error: runtime/pprof.Labels not implemented")
}

func Label(ctx context.Context, key string) (string, bool) {
  panic("gojr error: runtime/pprof.Label not implemented")
}

func ForLabels(ctx context.Context, f func(key, value string) bool) {
  panic("gojr error: runtime/pprof.ForLabels not implemented")
}

func SetGoroutineLabels(ctx context.Context) {
  panic("gojr error: runtime/pprof.SetGoroutineLabels not implemented")
}

func Do(ctx context.Context, labels LabelSet, f func(context.Context)) {
  panic("gojr error: runtime/pprof.Do not implemented")
}
`;
