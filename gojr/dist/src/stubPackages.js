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
export const syncIntrinsicSource = `package sync

type Locker interface {
  Lock()
  Unlock()
}

type Mutex struct{}

func (m *Mutex) Lock() {}
func (m *Mutex) TryLock() bool { return true }
func (m *Mutex) Unlock() {}

type RWMutex struct{}

func (rw *RWMutex) Lock() {}
func (rw *RWMutex) TryLock() bool { return true }
func (rw *RWMutex) Unlock() {}
func (rw *RWMutex) RLock() {}
func (rw *RWMutex) TryRLock() bool { return true }
func (rw *RWMutex) RUnlock() {}
func (rw *RWMutex) RLocker() Locker { return (*rlocker)(rw) }

type rlocker RWMutex

func (r *rlocker) Lock() { (*RWMutex)(r).RLock() }
func (r *rlocker) Unlock() { (*RWMutex)(r).RUnlock() }

type Once struct {
  done bool
}

func (o *Once) Do(f func()) {
  if o.done {
    return
  }
  o.done = true
  f()
}

func OnceFunc(f func()) func() {
  var once Once
  return func() {
    once.Do(f)
  }
}

func OnceValue[T any](f func() T) func() T {
  var once Once
  var value T
  return func() T {
    once.Do(func() {
      value = f()
    })
    return value
  }
}

func OnceValues[T1, T2 any](f func() (T1, T2)) func() (T1, T2) {
  var once Once
  var value1 T1
  var value2 T2
  return func() (T1, T2) {
    once.Do(func() {
      value1, value2 = f()
    })
    return value1, value2
  }
}

type Pool struct {
  New func() any
  items []any
}

func (p *Pool) Put(x any) {
  if x == nil {
    return
  }
  p.items = append(p.items, x)
}

func (p *Pool) Get() any {
  n := len(p.items)
  if n > 0 {
    x := p.items[n-1]
    p.items = p.items[:n-1]
    return x
  }
  newValue := p.New
  if newValue != nil {
    return newValue()
  }
  return nil
}

type WaitGroup struct {
  n int
}

func (wg *WaitGroup) Add(delta int) { wg.n += delta }
func (wg *WaitGroup) Done() { wg.Add(-1) }
func (wg *WaitGroup) Wait() {}

type Cond struct {
  L Locker
}

func NewCond(l Locker) *Cond { return &Cond{L: l} }
func (c *Cond) Broadcast() {}
func (c *Cond) Signal() {}
func (c *Cond) Wait() {}

type Map struct {
  m map[any]any
}

func (m *Map) ensure() {
  if m.m == nil {
    m.m = make(map[any]any)
  }
}

func (m *Map) Clear() {
  m.m = make(map[any]any)
}

func (m *Map) CompareAndDelete(key, old any) (deleted bool) {
  if current, ok := m.m[key]; ok && current == old {
    delete(m.m, key)
    return true
  }
  return false
}

func (m *Map) CompareAndSwap(key, old, new any) bool {
  m.ensure()
  if current, ok := m.m[key]; ok && current == old {
    m.m[key] = new
    return true
  }
  return false
}

func (m *Map) Delete(key any) {
  delete(m.m, key)
}

func (m *Map) Load(key any) (value any, ok bool) {
  value, ok = m.m[key]
  return value, ok
}

func (m *Map) LoadAndDelete(key any) (value any, loaded bool) {
  value, loaded = m.m[key]
  if loaded {
    delete(m.m, key)
  }
  return value, loaded
}

func (m *Map) LoadOrStore(key, value any) (actual any, loaded bool) {
  m.ensure()
  if actual, loaded = m.m[key]; loaded {
    return actual, true
  }
  m.m[key] = value
  return value, false
}

func (m *Map) Range(f func(key, value any) bool) {
  for key, value := range m.m {
    if !f(key, value) {
      return
    }
  }
}

func (m *Map) Store(key, value any) {
  m.ensure()
  m.m[key] = value
}

func (m *Map) Swap(key, value any) (previous any, loaded bool) {
  m.ensure()
  previous, loaded = m.m[key]
  m.m[key] = value
  return previous, loaded
}
`;
