# Plan: Port ivy_utils.py Section 12 (20 items) to Go

## Context

The AUDIT18MARCH.md identifies 20 items in `ivy_utils.py` that are missing or behaviorally different in the Go port. These span error infrastructure (items 1-7), utility functions (8-14), version/parameter globals (15-18), and two behavioral fixes (19-20). All must be faithfully ported from the Python source at `/Users/jaten/pyivy/ivy/ivy/ivy_utils.py`.

---

## Files to modify/create

| File | Action |
|------|--------|
| `ivyutils/error.go` | **CREATE** — IvyError, IvyUndefined, ErrorList, ErrorPrinter, Warn |
| `ivyutils/location.go` | **CREATE** — LocationTuple, Location(), Nowhere(), IsNowhere(), LinenoStr() |
| `ivyutils/source_file.go` | **CREATE** — SourceFile, WorkingDir context managers + global Filename |
| `ivyutils/listutils.go` | **CREATE** — Concat, UnzipAppend, UnzipPairs, Flatten, UnionOfList, UnionToList, ListUnion, ListDiff, DistinctUnorderedPairs, InverseMap, ComposeMaps, Partition, SplitList |
| `ivyutils/combinators.go` | **CREATE** — Unique, GenList, and other combinator-style helpers |
| `ivyutils/pretty.go` | **CREATE** — Pretty(s, maxLines) with proper formatting |
| `ivyutils/names.go` | **MODIFY** — add ParseIntSubscripts, GetNumericVersion, StringVersionToNumericVersion; fix SetStringVersion |
| `ivyutils/renamer.go` | **MODIFY** — add DistinctObjRenaming |
| `ivyutils/globals.go` | **CREATE** — UseNumerals, UseNewUI, Catch, DefaultUI, EnableDebug parameters; Dbg(); GetDefaultUIModule/Class |
| `trace/trace.go` | **MODIFY** — fix Pretty() to match Python formatting |
| `logic/error.go` | **MODIFY** — upgrade IvyError with Reference chain, integrate with Catch parameter |

---

## Detailed item-by-item plan

### Item 1: `IvyError` exception → upgrade `logic/error.go`

**Python** (lines 288-305): IvyError has `lineno` (a LocationTuple with optional reference chain), `msg`. Constructor checks global `catch` — if false, prints and asserts. `__str__` recurses through `lineno.reference` chain printing "instantiated here" at each level.

**Current Go** (`logic/error.go:9-33`): Simple struct with `Msg`, `Loc ast.Location`, `HasLoc`. No reference chain, no catch check.

**Plan**: Move the canonical IvyError to `ivyutils/error.go` to avoid circular imports (ivyutils has the `Catch` parameter). The existing `logic.IvyError` stays as-is for backward compat but the new `ivyutils.IvyError` is the full-featured one used going forward.

```go
// ivyutils/error.go
type IvyError struct {
    Lineno *LocationTuple
    Msg    string
}

func NewIvyError(lineno interface{}, msg string) *IvyError {
    // lineno can be: *LocationTuple, ast.Location, ast.Node (extract .GetLineno()), or nil
    loc := extractLocation(lineno)
    e := &IvyError{Lineno: loc, Msg: msg}
    if !Catch.GetBool() {
        fmt.Println(e.Error())
        panic("IvyError: " + msg)  // Python does assert False
    }
    return e
}

func (e *IvyError) Error() string {
    return e.recur(e.Lineno)
}

func (e *IvyError) recur(lineno *LocationTuple) string {
    if lineno != nil && lineno.Reference != nil {
        res := e.recur(lineno.Reference)
        ref := &LocationTuple{Filename: lineno.Filename, Line: lineno.Line}
        return res + "\n" + ref.String() + "error: instantiated here"
    }
    prefix := ""
    if lineno != nil {
        prefix = lineno.String()
    }
    return prefix + "error: " + e.Msg
}
```

### Item 2: `IvyUndefined` exception

**Python** (lines 307-310): Subclass of IvyError, prepends "undefined: " to name.

```go
// ivyutils/error.go
type IvyUndefined struct {
    IvyError
}

func NewIvyUndefined(lineno interface{}, name string) *IvyUndefined {
    e := NewIvyError(lineno, "undefined: "+name)
    return &IvyUndefined{IvyError: *e}
}
```

### Item 3: `ErrorList` class

**Python** (lines 535-541): Holds list of errors. `__repr__` joins them, conditionally prefixes with `self.filename`.

```go
// ivyutils/error.go
type ErrorList struct {
    Errors   []error
    Filename string
}

func NewErrorList(errors []error) *ErrorList {
    return &ErrorList{Errors: errors}
}

func (e *ErrorList) Error() string {
    var parts []string
    for _, err := range e.Errors {
        if ivyErr, ok := err.(*IvyError); ok && ivyErr.Lineno != nil && ivyErr.Lineno.Filename != "" {
            parts = append(parts, err.Error())
        } else if e.Filename != "" {
            parts = append(parts, e.Filename+": "+err.Error())
        } else {
            parts = append(parts, err.Error())
        }
    }
    return strings.Join(parts, "\n")
}
```

### Item 4: `ErrorPrinter` context manager

**Python** (lines 519-530): Catches IvyError, prints it, and exits with code 1.

Go equivalent: a function that takes a closure, recovers panics of type `*IvyError`.

```go
// ivyutils/error.go
func WithErrorPrinter(fn func()) {
    defer func() {
        if r := recover(); r != nil {
            if e, ok := r.(*IvyError); ok {
                fmt.Println(e.Error())
                os.Exit(1)
            } else if e, ok := r.(*IvyUndefined); ok {
                fmt.Println(e.Error())
                os.Exit(1)
            }
            panic(r) // re-panic non-IvyError
        }
    }()
    fn()
}
```

### Item 5: `SourceFile` context manager

**Python** (lines 207-226): Saves/restores global `filename`.

```go
// ivyutils/source_file.go
var Filename string // global, corresponds to Python's ivy_utils.filename

func WithSourceFile(fname string, fn func()) {
    oldFilename := Filename
    Filename = fname
    defer func() { Filename = oldFilename }()
    fn()
}
```

### Item 6: `WorkingDir` context manager

**Python** (lines 228-242): Saves/restores working directory.

```go
// ivyutils/source_file.go
func WithWorkingDir(dir string, fn func()) error {
    oldDir, err := os.Getwd()
    if err != nil {
        return err
    }
    if err := os.Chdir(dir); err != nil {
        return err
    }
    defer os.Chdir(oldDir)
    fn()
    return nil
}
```

### Item 7: `LocationTuple` / `Location()` / `nowhere()` / `is_nowhere()` / `lineno_str()`

**Python** (lines 245-286): LocationTuple extends tuple with `filename`, `line`, `reference` (3rd element). Platform-dependent string formatting.

```go
// ivyutils/location.go
type LocationTuple struct {
    Filename  string
    Line      int
    Reference *LocationTuple
}

func Location(filename string, line int) *LocationTuple {
    return &LocationTuple{Filename: filename, Line: line}
}

func Nowhere() *LocationTuple {
    return Location("nowhere", 0)
}

func IsNowhere(loc *LocationTuple) bool {
    return loc != nil && loc.Filename == "nowhere"
}

func (lt *LocationTuple) String() string {
    if lt.Reference != nil {
        return lt.Reference.String()
    }
    res := ""
    if lt.Filename != "" {
        res += lt.Filename + ": "
    }
    if lt.Line > 0 {
        res += "line " + strconv.Itoa(lt.Line) + ": "
    }
    return res
}

// LinenoStr extracts location string from an AST node.
func LinenoStr(node interface{}) string {
    // Accept ast.Node or anything with GetLineno()
    type hasLineno interface {
        GetLineno() ast.Location
    }
    if n, ok := node.(hasLineno); ok {
        loc := n.GetLineno()
        r := Location(loc.Filename, loc.Line).String()
        r = strings.TrimSuffix(r, ": ")
        return r
    }
    return ""
}
```

Note: We skip Windows formatting (Python has `platform.system() == 'Windows'` branch) since goivy targets Unix. The Python non-Windows path is what we port.

### Item 8: `warn(ast, msg)`

**Python** (lines 312-313): Creates IvyError string, replaces "error:" with "warning:".

```go
// ivyutils/error.go
func Warn(lineno interface{}, msg string) {
    // Temporarily set catch=true to prevent panic in NewIvyError
    oldCatch := Catch.GetBool()
    Catch.Value = true
    e := NewIvyError(lineno, msg)
    Catch.Value = oldCatch
    fmt.Println(strings.Replace(e.Error(), "error: ", "warning: ", -1))
}
```

Actually, simpler approach — just format the string directly without constructing the error:

```go
func Warn(lineno interface{}, msg string) {
    loc := extractLocation(lineno)
    prefix := ""
    if loc != nil {
        prefix = loc.String()
    }
    fmt.Println(prefix + "warning: " + msg)
}
```

### Item 9: `parse_with()` / `p_error()`

**Python** (lines 544-556): `parse_with` collects parse errors into a global list, runs parser, raises ParseErrorList if any errors. `p_error` is the PLY parser error callback.

These are tightly coupled to the Python PLY parser. In Go, the parser is hand-written. We port the error-collection pattern:

```go
// ivyutils/error.go  (or parser/errors.go if more appropriate)
var ParseErrorList []error

func ParseWith(parseFn func(string) (interface{}, error), s string) (interface{}, error) {
    ParseErrorList = nil
    result, err := parseFn(s)
    if err != nil {
        return nil, err
    }
    if len(ParseErrorList) > 0 {
        return nil, &ErrorList{Errors: ParseErrorList}
    }
    return result, nil
}

func PError(lineno int, value string, msg string) {
    if value != "" {
        ParseErrorList = append(ParseErrorList, NewIvyError(
            Location("", lineno), msg))
    } else {
        ParseErrorList = append(ParseErrorList, NewIvyError(
            nil, "unexpected end of input"))
    }
}
```

### Item 10: Combinator functions (10 functions)

**Python** (lines 15-78): Higher-order generator combinators. Many rely on Python generators and closures.

Go doesn't have generators. Port as concrete slice-based functions using generics:

```go
// ivyutils/combinators.go

// Unique returns a new slice with duplicates removed, preserving order.
func Unique[T comparable](gen []T) []T {
    memo := make(map[T]struct{})
    var result []T
    for _, x := range gen {
        if _, seen := memo[x]; !seen {
            memo[x] = struct{}{}
            result = append(result, x)
        }
    }
    return result
}

// GenList applies fn to each element of l and collects all results.
// Corresponds to Python's gen_list (flattening generator).
func GenList[T any, U any](fn func(T) []U, l []T) []U {
    var result []U
    for _, x := range l {
        result = append(result, fn(x)...)
    }
    return result
}

// GenListList applies fn to each element of each inner list.
func GenListList[T any, U any](fn func(T) []U, l [][]T) []U {
    var result []U
    for _, inner := range l {
        for _, x := range inner {
            result = append(result, fn(x)...)
        }
    }
    return result
}

// Unique2 given pairs (x,y), yields one x for each unique y.
func Unique2[X any, Y comparable](pairs []struct{ First X; Second Y }) []X {
    memo := make(map[Y]struct{})
    var result []X
    for _, p := range pairs {
        if _, seen := memo[p.Second]; !seen {
            memo[p.Second] = struct{}{}
            result = append(result, p.First)
        }
    }
    return result
}

// AnyIn returns true if fn(args) produces any element in the given set.
func AnyIn[T comparable](aSet map[T]struct{}, items []T) bool {
    for _, x := range items {
        if _, ok := aSet[x]; ok {
            return true
        }
    }
    return false
}
```

Note: `apply_func_to_list`, `apply_gen_to_list`, `apply_gen_to_list_list`, `gen_to_set`, `gen_to_dict`, `gen_unique`, `filter2` are higher-order function factories returning lambdas. In Go these are best ported as direct calls (GenList, etc.) since Go doesn't have the same closure/generator idiom. We port the *behavior* they provide, which callers use directly. The factory wrappers (`apply_func_to_list` etc.) are unnecessary in Go — callers simply call GenList/GenListList directly.

### Item 11: List utilities (13 functions)

```go
// ivyutils/listutils.go

// Concat flattens a slice of slices into one slice.
func Concat[T any](args [][]T) []T {
    var result []T
    for _, a := range args {
        result = append(result, a...)
    }
    return result
}

// UnzipAppend unzips a list of N-tuples (represented as [][]T) and concatenates each position.
func UnzipAppend[T any](tups [][]T) [][]T {
    if len(tups) == 0 {
        return nil
    }
    width := len(tups[0])
    result := make([][]T, width)
    for _, tup := range tups {
        for i := 0; i < width && i < len(tup); i++ {
            result[i] = append(result[i], tup[i])
        }
    }
    return result
}

// UnzipPairs unzips a list of pairs into two lists.
func UnzipPairs[A any, B any](pairs []struct{ First A; Second B }) ([]A, []B) {
    if len(pairs) == 0 {
        return nil, nil
    }
    as := make([]A, len(pairs))
    bs := make([]B, len(pairs))
    for i, p := range pairs {
        as[i] = p.First
        bs[i] = p.Second
    }
    return as, bs
}

// Flatten recursively flattens nested slices. In Go, since we don't have
// runtime type nesting like Python, this just flattens [][]T → []T.
// (Same as Concat for typed slices.)
func Flatten[T any](l [][]T) []T {
    return Concat(l)
}

// UnionOfList unions all sets (maps) in a list into one set.
func UnionOfList[T comparable](sets []map[T]struct{}) map[T]struct{} {
    res := make(map[T]struct{})
    for _, s := range sets {
        for k := range s {
            res[k] = struct{}{}
        }
    }
    return res
}

// UnionToList appends elements from fromList to toList, skipping duplicates.
func UnionToList[T comparable](toList *[]T, fromList []T) {
    used := make(map[T]struct{}, len(*toList))
    for _, x := range *toList {
        used[x] = struct{}{}
    }
    for _, x := range fromList {
        if _, ok := used[x]; !ok {
            *toList = append(*toList, x)
            used[x] = struct{}{}
        }
    }
}

// ListUnion returns the union of two lists preserving order.
func ListUnion[T comparable](l1, l2 []T) []T {
    res := make([]T, len(l1))
    copy(res, l1)
    UnionToList(&res, l2)
    return res
}

// ListDiff returns elements in l1 that are not in l2.
func ListDiff[T comparable](l1, l2 []T) []T {
    sl2 := make(map[T]struct{}, len(l2))
    for _, x := range l2 {
        sl2[x] = struct{}{}
    }
    var result []T
    for _, x := range l1 {
        if _, ok := sl2[x]; !ok {
            result = append(result, x)
        }
    }
    return result
}

// DistinctUnorderedPairs yields all (l[i], l[j]) where i < j.
func DistinctUnorderedPairs[T any](l []T) [][2]T {
    var result [][2]T
    for i := 0; i < len(l)-1; i++ {
        for j := i + 1; j < len(l); j++ {
            result = append(result, [2]T{l[i], l[j]})
        }
    }
    return result
}

// InverseMap inverts a map (swaps keys and values).
func InverseMap[K comparable, V comparable](m map[K]V) map[V]K {
    res := make(map[V]K, len(m))
    for k, v := range m {
        res[v] = k
    }
    return res
}

// ComposeMaps composes two maps as functions. Missing keys act as identity.
func ComposeMaps[K comparable](m1, m2 map[K]K) map[K]K {
    res := make(map[K]K, len(m2))
    for k, v := range m2 {
        res[k] = v
    }
    for k, v := range m1 {
        if v2, ok := m2[v]; ok {
            res[k] = v2
        } else {
            res[k] = v
        }
    }
    return res
}

// Partition groups elements by a key function.
func Partition[T any, K comparable](things []T, key func(T) K) map[K][]T {
    res := make(map[K][]T)
    for _, t := range things {
        k := key(t)
        res[k] = append(res[k], t)
    }
    return res
}

// SplitList splits l into two lists based on predicate map p.
// Returns (matching, non-matching).
func SplitList[T comparable](l []T, p map[T]bool) ([]T, []T) {
    var yes, no []T
    for _, x := range l {
        if p[x] {
            yes = append(yes, x)
        } else {
            no = append(no, x)
        }
    }
    return yes, no
}
```

Note: `transrel/transrel.go` already has `ListDiff`, `ComposeMaps`, `InverseMap` but they are string-specific (use `Renaming` type = `map[string]string`). The generic versions in `ivyutils` are the canonical ones matching Python. We keep both to avoid breaking transrel imports.

### Item 12: `pretty(s, max_lines)` — fix formatting

**Python** (lines 680-694): Splits on `;{}`→newlines, strips, truncates, indents by `{`/`}` nesting, appends closing `}` for unclosed braces.

**Current Go** (`trace/trace.go:159-166`): Only truncates lines, no formatting.

**Plan**: Replace `trace.Pretty` with a full port. Also add `ivyutils.Pretty` as the canonical location.

```go
// ivyutils/pretty.go
func Pretty(s string, maxLines int) string {
    s = strings.ReplaceAll(s, ";", ";\n")
    s = strings.ReplaceAll(s, "{", "{\n")
    s = strings.ReplaceAll(s, "}", "\n}")
    lines := strings.Split(s, "\n")
    for i, line := range lines {
        lines[i] = strings.TrimSpace(line)
    }
    if maxLines > 0 && len(lines) > maxLines {
        lines = lines[:maxLines-1]
        lines = append(lines, "...")
    }
    indent := 0
    var res []string
    for _, line := range lines {
        if strings.Contains(line, "}") {
            indent--
        }
        if indent < 0 {
            indent = 0
        }
        res = append(res, strings.Repeat("    ", indent)+line)
        if strings.Contains(line, "{") {
            indent++
        }
    }
    return strings.Join(res, "\n") + strings.Repeat("}", indent)
}
```

Update `trace/trace.go` to call `ivyutils.Pretty` or inline the same logic.

### Item 13: `parse_int_subscripts(name)`

**Python** (lines 758-765): Parses `"f[1][2]"` → `("f", [1, 2])`.

```go
// ivyutils/names.go
func ParseIntSubscripts(name string) (string, []int, error) {
    things := strings.Split(name, "[")
    thy := things[0]
    things = things[1:]
    for _, t := range things {
        if !strings.HasSuffix(t, "]") {
            return "", nil, fmt.Errorf("bad subscript syntax: %s", name)
        }
    }
    prms := make([]int, len(things))
    for i, t := range things {
        val, err := strconv.Atoi(t[:len(t)-1])
        if err != nil {
            return "", nil, fmt.Errorf("bad subscript syntax: %s", name)
        }
        prms[i] = val
    }
    return thy, prms, nil
}
```

### Item 14: `distinct_obj_renaming(names1, names2)`

**Python** (lines 186-188): Renames objects using UniqueRenamer. Objects must have `.rename()` method.

In Go, we need an interface. The Python code calls `s.rename(rn)` where `rn` is a callable UniqueRenamer. Port:

```go
// ivyutils/renamer.go

// Renameable is anything that can be renamed via a UniqueRenamer.
type Renameable interface {
    RenameWith(rn *UniqueRenamer) Renameable
    Key() string // for map key (string representation)
}

// DistinctObjRenaming maps each name in names1 to a renamed version avoiding names2.
func DistinctObjRenaming(names1 []Renameable, names2 []string) map[string]Renameable {
    rn := NewUniqueRenamer("", names2)
    result := make(map[string]Renameable, len(names1))
    for _, s := range names1 {
        result[s.Key()] = s.RenameWith(rn)
    }
    return result
}
```

### Item 15: Version functions

**Python** (lines 583-590): `string_version_to_numeric_version` splits on `.` and converts to `[]int`. `get_numeric_version` applies it to global version. `version_le` compares two version strings.

`VersionLE` already exists in `names.go:223-246`. Add the missing two:

```go
// ivyutils/names.go
func StringVersionToNumericVersion(v string) []int {
    parts := strings.Split(v, ".")
    result := make([]int, len(parts))
    for i, p := range parts {
        result[i], _ = strconv.Atoi(p)
    }
    return result
}

func GetNumericVersion() []int {
    return StringVersionToNumericVersion(ivyLanguageVersion)
}
```

### Item 16: Global parameters

**Python** (lines 716-720, 742): Five parameters defined at module level.

`UseNumerals` already exists in `transrel/phase4.go:322-340` as a standalone bool. It should also be a proper Parameter, or at minimum, the canonical location should be `ivyutils/globals.go` to match Python's module placement.

```go
// ivyutils/globals.go
var (
    // GlobalRegistry is the default parameter registry for ivy_utils parameters.
    GlobalRegistry = NewParameterRegistry()

    UseNumerals = NewBooleanParameterOn(GlobalRegistry, "use_numerals", true)
    UseNewUI    = NewBooleanParameterOn(GlobalRegistry, "new_ui", false)
    Catch       = NewBooleanParameterOn(GlobalRegistry, "catch", true)
    DefaultUI   = NewParameterOn(GlobalRegistry, "ui", "cti")
    EnableDebug = NewBooleanParameterOn(GlobalRegistry, "debug", false)
)
```

### Item 17: `dbg()` function

**Python** (lines 744-755): Evaluates string expressions in caller's frame. This uses Python's `eval()` + `inspect.currentframe()` which has no Go equivalent.

Port as a simplified debug printer that takes key-value pairs:

```go
// ivyutils/globals.go
func Dbg(kvs ...interface{}) {
    if !EnableDebug.GetBool() {
        panic("must use debug=true to enable debug output")
    }
    for i := 0; i+1 < len(kvs); i += 2 {
        fmt.Printf("%v:%v\n", kvs[i], kvs[i+1])
    }
}
```

Callers in Python do `dbg("x")` which evals "x" — in Go callers must do `Dbg("x", x)`. This is the closest mechanical equivalent since Go has no runtime eval.

### Item 18: `get_default_ui_module()` / `get_default_ui_class()`

**Python** (lines 722-739): Dynamic import based on `default_ui` parameter. Returns module/class.

Go can't do dynamic imports. Port as a registry pattern:

```go
// ivyutils/globals.go

// UIModule represents a UI module with an IvyUI class and compile_kwargs.
type UIModule struct {
    Name         string
    NewIvyUI     func() interface{}
    CompileKwargs map[string]interface{}
}

var uiModules = make(map[string]*UIModule)

func RegisterUIModule(name string, mod *UIModule) {
    uiModules[name] = mod
}

func GetDefaultUIModule() *UIModule {
    defui := DefaultUI.GetString()
    name := defui
    if defui == "art" {
        name = "ivy_ui"
    } else {
        name = "ivy_ui_" + defui
    }
    mod, ok := uiModules[name]
    if !ok {
        return nil
    }
    return mod
}

func GetDefaultUIClass() func() interface{} {
    mod := GetDefaultUIModule()
    if mod == nil {
        return nil
    }
    return mod.NewIvyUI
}

func GetDefaultUICompileKwargs() map[string]interface{} {
    mod := GetDefaultUIModule()
    if mod == nil {
        return nil
    }
    return mod.CompileKwargs
}
```

### Item 19 (Behavioral): Fix `SetStringVersion` compose character

**Python** (line 574): `ivy_compose_character = ':' if get_numeric_version() <= [1,1] else '.'`

**Current Go** (`names.go:155-156`): `if major < 1 || (major == 1 && minor < 3) { ComposeCharacter = "__" }`

**Two bugs**: wrong threshold (1.3 vs 1.1) AND wrong character (`__` vs `:`).

Also missing: `ivy_use_polymorphic_macros` (version > 1.5), `ivy_forbid_ghost_init` (version > 1.6), `symbol_chars_parser` regex update.

Fix `SetStringVersion` in `names.go`:

```go
func SetStringVersion(version string) {
    ivyLanguageVersion = strings.TrimSpace(version)
    nv := GetNumericVersion()
    // Python: ':' if version <= [1,1] else '.'
    if versionLESlice(nv, []int{1, 1}) {
        ComposeCharacter = ":"
    } else {
        ComposeCharacter = "."
    }
    SymbolCharsParser = regexp.MustCompile(`[^\[\]` + regexp.QuoteMeta(ComposeCharacter) + `]*`)
    // Python: ivy_have_polymorphism = not get_numeric_version() <= [1,2]
    IvyHavePolymorphism = !versionLESlice(nv, []int{1, 2})
    // Python: ivy_use_polymorphic_macros = not get_numeric_version() <= [1,5]
    IvyUsePolymorphicMacros = !versionLESlice(nv, []int{1, 5})
    // Python: ivy_forbid_ghost_init = not get_numeric_version() <= [1,6]
    IvyForbidGhostInit = !versionLESlice(nv, []int{1, 6})
}

// versionLESlice compares two numeric version slices (Python list <= comparison).
func versionLESlice(v1, v2 []int) bool {
    for i := 0; i < len(v1) || i < len(v2); i++ {
        a, b := 0, 0
        if i < len(v1) { a = v1[i] }
        if i < len(v2) { b = v2[i] }
        if a < b { return true }
        if a > b { return false }
    }
    return true // equal
}
```

Also add the new global variables:

```go
var IvyUsePolymorphicMacros = false
var IvyForbidGhostInit = false
var IvyLatestLanguageVersion = "1.7"
var SymbolCharsParser = regexp.MustCompile(`[^\[\]\.]*`)
```

Then update `ivylogic/classify_ext.go:344` to use `ivyutils.IvyUsePolymorphicMacros` instead of its local `UsePolymorphicMacros` variable.

### Item 20 (Behavioral): Fix `GetStdIncludeDir`

**Python** (lines 592-604): Scans `<package>/include/` for version-matching subdirectories, picks smallest version >= current.

**Current Go** (`names.go:190-201`): Just checks if `include/` exists in CWD.

Replace with full Python-matching logic:

```go
var incDirPat = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)

func GetStdIncludeDir() string {
    if stdIncludeDir != "" {
        return stdIncludeDir
    }
    // Find the directory containing this package (like Python's os.path.dirname(__file__))
    // In Go, use the executable path or a configured base path.
    incBaseDir := getIncludeBaseDir()

    entries, err := os.ReadDir(incBaseDir)
    if err != nil {
        panic(NewIvyError(nil, fmt.Sprintf(
            "cannot find standard library for language version %s", ivyLanguageVersion)))
    }

    var bestDir string
    for _, entry := range entries {
        d := entry.Name()
        if !entry.IsDir() {
            continue
        }
        if !incDirPat.MatchString(d) {
            continue
        }
        // version_le(ivy_language_version, d): current version <= directory version
        if !VersionLE(ivyLanguageVersion, d) {
            continue
        }
        // Pick smallest qualifying version
        if bestDir == "" || VersionLE(d, bestDir) {
            bestDir = d
        }
    }
    if bestDir == "" {
        panic(NewIvyError(nil, fmt.Sprintf(
            "cannot find standard library for language version %s", ivyLanguageVersion)))
    }
    stdIncludeDir = filepath.Join(incBaseDir, bestDir)
    return stdIncludeDir
}

// getIncludeBaseDir returns the base directory containing version subdirectories.
// Tries: executable dir + "/include", then CWD + "/include".
func getIncludeBaseDir() string {
    // Try relative to executable
    if exe, err := os.Executable(); err == nil {
        dir := filepath.Join(filepath.Dir(exe), "include")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
    }
    // Try CWD
    if info, err := os.Stat("include"); err == nil && info.IsDir() {
        return "include"
    }
    return "include" // fallback
}
```

---

## Dependency order for implementation

1. `ivyutils/location.go` — LocationTuple (no deps)
2. `ivyutils/globals.go` — Global parameters including Catch (depends on parameter.go)
3. `ivyutils/error.go` — IvyError, IvyUndefined, ErrorList, ErrorPrinter, Warn (depends on location.go, globals.go)
4. `ivyutils/source_file.go` — SourceFile, WorkingDir (no deps)
5. `ivyutils/combinators.go` — Unique, GenList, etc. (no deps)
6. `ivyutils/listutils.go` — Concat, Flatten, etc. (no deps)
7. `ivyutils/pretty.go` — Pretty (no deps)
8. `ivyutils/names.go` — add ParseIntSubscripts, GetNumericVersion, StringVersionToNumericVersion; fix SetStringVersion (depends on globals.go for new vars)
9. `ivyutils/renamer.go` — add DistinctObjRenaming (depends on renamer.go existing code)
10. `trace/trace.go` — update Pretty to call ivyutils.Pretty or inline full logic

---

## Verification

1. `cd /Users/jaten/go/src/github.com/glycerine/goivy && make test` — all existing tests must pass
2. Write unit tests for each new file:
   - `ivyutils/location_test.go` — LocationTuple.String(), Nowhere(), IsNowhere(), LinenoStr()
   - `ivyutils/error_test.go` — IvyError with reference chain, IvyUndefined, ErrorList, Warn
   - `ivyutils/listutils_test.go` — Concat, ListDiff, ListUnion, Partition, ComposeMaps, etc.
   - `ivyutils/pretty_test.go` — Pretty formatting with braces/semicolons
   - `ivyutils/names_test.go` — ParseIntSubscripts, SetStringVersion compose character fix
   - `ivyutils/globals_test.go` — parameter registry, Dbg
3. Verify SetStringVersion("1.0") sets ComposeCharacter to ":" (not "__")
4. Verify SetStringVersion("1.2") sets ComposeCharacter to "." and IvyHavePolymorphism to false
5. Verify GetStdIncludeDir scans version subdirectories correctly
