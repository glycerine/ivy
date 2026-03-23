# Fix: Include Resolution — `include order` fails with "unbounded_sequence undefined"

**Created: 2026-03-22**

## Context

`goivy_check` parses ord_live.ivy successfully (3286 lines), but compilation fails at line 32:
```
error: at line 32: line 32: error: unbounded_sequence undefined in instantiation
```
Line 30 is `include order` and line 32 is `instance lclock : unbounded_sequence`. The `order.ivy` standard library (which defines `unbounded_sequence`) is never loaded.

## Root Cause (Two Bugs)

### Bug 1: Parser records includes as NamedDecl but never loads them

**Python behavior** (ivy_parser.py:370-385):
```python
def p_top_include_symbol(p):
    'top : top INCLUDE SYMBOL'
    module = importer(p[3])          # <-- LOADS FILE IMMEDIATELY
    for decl in module.decls:
        p[0].declare(decl)           # <-- MERGES declarations into current module
```
Python uses a callback (`importer`) set to `import_module()` before parsing begins. When the parser encounters `include order`, it immediately loads `order.ivy`, parses it, and injects its declarations into the current parse tree.

**Go behavior** (parser/decl.go:2095-2099):
```go
func (p *Parser) parseIncludeDecl(tok lexer.Token) ast.Node {
    p.advance()
    name := p.expect(lexer.SYMBOL)
    return ast.NewNamedDecl(ast.NewSymbol(name.Value, nil))  // Just records name, never loads
}
```
Go creates a `NamedDecl` AST node with the include name but never loads the file. The compiler later sees `NamedDecl` and tries to process it as a named property declaration — not an include.

### Bug 2: Include base directory resolution is wrong

**Python** (ivy_utils.py:594-595):
```python
inc_base_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'include')
# Resolves to: /path/to/ivy/ivy/include/
```
Python uses the **source package directory** where `ivy_utils.py` lives.

**Go** (ivyutils/names.go:278-291):
```go
func getIncludeBaseDir() string {
    // Try relative to executable
    dir := filepath.Join(filepath.Dir(exe), "include")  // goivy_check binary dir
    // Try CWD
    os.Stat("include")  // current working directory
    return "include"     // fallback
}
```
Go tries executable dir and CWD. Neither has an `include/` subdirectory. The actual include files are at `ivy-lang-examples/ivy/include/1.8/`.

## Fix

### Fix 1: Add importer callback to Parser and load includes eagerly

**File: `parser/parser.go`**

Add `Importer` callback field to Parser struct:
```go
type Parser struct {
    lex          *lexer.Lexer
    current      lexer.Token
    version      lexer.Version
    errors       []ParseError
    labelCounter int
    modules      map[string]*ast.ModuleDecl
    Importer     func(name string) ([]ast.Node, error)  // NEW: callback to load includes
    included     map[string]bool                         // NEW: track already-included modules
}
```

**File: `parser/decl.go`**

Rewrite `parseIncludeDecl` to match Python:
```go
func (p *Parser) parseIncludeDecl(tok lexer.Token) ast.Node {
    p.advance()
    name := p.expect(lexer.SYMBOL).Value

    // Prevent double-include (Python: if not any(p[3] in m.included for m in stack))
    if p.included[name] {
        return nil // already included
    }
    p.included[name] = true

    // Call importer callback to load and parse the included file
    if p.Importer != nil {
        decls, err := p.Importer(name)
        if err != nil {
            p.errorf("include %s: %v", name, err)
            return nil
        }
        // Return all included declarations (they get merged into the parent)
        // Use an IncludeDecl wrapper that carries the child declarations
        return &ast.IncludeDecl{Name: name, Decls: decls}
    }

    p.errorf("no importer configured for include %s", name)
    return nil
}
```

The `parseTopLevel` dispatch for INCLUDE should return the included declarations as multiple nodes (like Python does with `p[0].declare(decl)` for each):
```go
case lexer.INCLUDE:
    node := p.parseIncludeDecl(tok)
    if inc, ok := node.(*ast.IncludeDecl); ok && inc != nil {
        return inc.Decls  // return included declarations directly
    }
    return nil
```

**File: `ast/decl.go`**

Add `IncludeDecl` AST node type:
```go
type IncludeDecl struct {
    Name  string
    Decls []Node
    // ... standard AST fields
}
```

**File: `ivyinit/ivyinit.go`**

Wire the importer callback when creating the parser. Modify `ReadModule`:
```go
func ReadModule(filename string, nested bool) ([]ast.Node, error) {
    // ... existing header/version parsing ...
    p := parser.New(s, version)
    p.Importer = ImportModule    // Wire the callback
    p.included = make(map[string]bool)  // Or set via constructor
    decls, parseErr := p.Parse()
    // ...
}
```

### Fix 2: Fix include base directory resolution

**File: `ivyutils/names.go`**

`getIncludeBaseDir()` needs to find the include directory relative to a known location. Options:

**Option A (recommended)**: Allow setting the include base dir explicitly, and have `goivy_check` set it based on the source file's directory or a known path. Add an environment variable `GOIVY_INCLUDE` or a CLI parameter.

**Option B**: Embed the include files or use `go:embed` (heavyweight).

**Option C (simplest for now)**: Search relative to the source file being loaded. When loading `ord_live.ivy`, look for `include/` relative to where the Ivy standard library is installed. Add a search path that includes `ivy-lang-examples/ivy/include/`.

The most faithful port of Python would be:
```go
func getIncludeBaseDir() string {
    // Try GOIVY_INCLUDE environment variable first
    if dir := os.Getenv("GOIVY_INCLUDE"); dir != "" {
        return dir
    }
    // Try relative to executable (like Python's __file__ relative path)
    if exe, err := os.Executable(); err == nil {
        dir := filepath.Join(filepath.Dir(exe), "include")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
        // Also try ivy-lang-examples/ivy/include relative to executable
        dir = filepath.Join(filepath.Dir(exe), "ivy-lang-examples", "ivy", "include")
        if info, err := os.Stat(dir); err == nil && info.IsDir() {
            return dir
        }
    }
    // Try CWD
    if info, err := os.Stat("include"); err == nil && info.IsDir() {
        return "include"
    }
    return "include"
}
```

Also, reset `stdIncludeDir` cache when version changes (since different versions use different include dirs):
```go
func SetStringVersion(version string) {
    ivyLanguageVersion = strings.TrimSpace(version)
    stdIncludeDir = ""  // Reset cache so GetStdIncludeDir re-resolves
    // ... rest of existing code
}
```

### Fix 3: Compiler must handle included declarations

If using the IncludeDecl approach where included decls are returned from the parser as part of the declaration list, the compiler already processes them because they become regular declaration nodes (PropertyDecl, TypeDecl, etc.) in the parent's declaration list.

If instead the parser returns an IncludeDecl wrapper, add handling to the compiler:

**File: `compiler/decl.go`**

```go
case *ast.IncludeDecl:
    // Process included declarations as if they were in the current file
    for _, decl := range n.Decls {
        if err := d.processOneDecl(decl); err != nil {
            return err
        }
    }
```

## Unit Tests

### Test 1: Include resolution finds standard library
```go
func TestImportModule_Order(t *testing.T) {
    // Set include dir to point to test fixtures
    iu.SetStdIncludeDir("../ivy-lang-examples/ivy/include/1.8")
    defer iu.SetStdIncludeDir("")

    decls, err := ivyinit.ImportModule("order")
    if err != nil {
        t.Fatalf("ImportModule(order) failed: %v", err)
    }
    if len(decls) == 0 {
        t.Fatal("ImportModule(order) returned no declarations")
    }
    // Verify unbounded_sequence is defined
    found := false
    for _, d := range decls {
        if md, ok := d.(*ast.ModuleDecl); ok {
            if name := md.Name(); name == "unbounded_sequence" {
                found = true
            }
        }
    }
    if !found {
        t.Error("unbounded_sequence not found in order.ivy declarations")
    }
}
```

### Test 2: Include callback wired in parser
```go
func TestParserIncludeCallback(t *testing.T) {
    input := `include testmod`
    p := parser.New("\n"+input, lexer.Version{1, 8})
    called := false
    p.Importer = func(name string) ([]ast.Node, error) {
        called = true
        if name != "testmod" {
            t.Errorf("expected include name 'testmod', got %q", name)
        }
        // Return a simple type declaration
        return []ast.Node{ast.NewTypeDecl(ast.NewTypeDef("testtype", nil))}, nil
    }
    decls, err := p.Parse()
    if err != nil {
        t.Fatalf("parse failed: %v", err)
    }
    if !called {
        t.Error("importer callback was never called")
    }
    if len(decls) == 0 {
        t.Error("expected included declarations to be present")
    }
}
```

### Test 3: Double-include prevention
```go
func TestParserIncludeNoDuplicate(t *testing.T) {
    input := "include foo\ninclude foo"
    p := parser.New("\n"+input, lexer.Version{1, 8})
    count := 0
    p.Importer = func(name string) ([]ast.Node, error) {
        count++
        return nil, nil
    }
    p.Parse()
    if count != 1 {
        t.Errorf("importer called %d times, expected 1 (double-include prevention)", count)
    }
}
```

### Test 4: End-to-end include resolution
```go
func TestSourceFileWithIncludes(t *testing.T) {
    // Write a temp file that includes order
    iu.SetStdIncludeDir("ivy-lang-examples/ivy/include/1.8")
    defer iu.SetStdIncludeDir("")

    tmpFile := writeTemp(t, "#lang ivy1.8\ninclude order\ninstance x : unbounded_sequence\n")
    mod := module.New()
    mod.Cfg = module.NewConfig()
    err := ivyinit.SourceFile(tmpFile, mod, mod.Sig, map[string]interface{}{"create_isolate": false})
    if err != nil {
        t.Fatalf("SourceFile failed: %v", err)
    }
}
```

### Test 5: GetStdIncludeDir version selection
```go
func TestGetStdIncludeDir_Version18(t *testing.T) {
    iu.SetStringVersion("1.8")
    iu.SetStdIncludeDir("") // Reset cache
    // Point to test include structure
    // Verify it selects 1.8 directory
}
```

## Files to Modify

| File | Change |
|------|--------|
| `parser/parser.go` | Add `Importer` callback and `included` map to Parser struct; initialize in `New()` |
| `parser/decl.go` | Rewrite `parseIncludeDecl()` to call Importer; change INCLUDE dispatch to return multiple nodes |
| `ast/decl.go` | Add `IncludeDecl` AST node type (optional — can inline decls instead) |
| `ivyinit/ivyinit.go` | Wire `p.Importer = ImportModule` in `ReadModule()`; ensure `ImportModule` passes `included` set |
| `ivyutils/names.go` | Fix `getIncludeBaseDir()` to search additional paths; reset `stdIncludeDir` in `SetStringVersion()` |
| `compiler/decl.go` | Handle `IncludeDecl` in ProcessDecls if using wrapper approach |
| `ivyinit/ivyinit_test.go` | Add unit tests for include resolution |
| `parser/parser_test.go` | Add unit tests for include callback |

## Verification

1. `go build ./...` compiles
2. `go test ./...` — all existing + new tests pass
3. `./goivy_check isolate=cf_live ord_live.ivy` — gets past line 32 (include order loads successfully)
4. `./goivy_check ivy-lang-examples/test/action1.ivy` — still works (no includes needed)
