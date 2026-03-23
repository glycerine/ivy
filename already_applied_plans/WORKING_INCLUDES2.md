# Fix: Module Registry Not Merged from Included Files

**Created: 2026-03-22**

## Context

After fixing include file loading, `goivy_check` fails with:
```
error: at line 383: unbounded_sequence undefined in instantiation
```
The file `order.ivy` is loaded via `include order` and its declarations are returned. But `unbounded_sequence` (defined as a `module` in `order.ivy`) is not available when `instance lclock : unbounded_sequence` is processed.

## Root Cause

In Python, module expansion happens **at parse time**. When the parser sees `instance X : module_name`, it looks up the module in `ivy.modules` and inlines it immediately. And when `include` loads a file, it explicitly merges the child's module registry: `p[0].modules.update(module.modules)`.

In Go, the parser has the same `p.modules` map and `expandInstantiation()` mechanism. But when `parseIncludeDeclMulti()` calls the Importer, the child parser parses the included file, registers modules in `childParser.modules`, returns `[]ast.Node`, and is discarded. The child's `modules` map is **never merged** into the parent parser.

### Proof: Python include handler (ivy_parser.py:370-385)
```python
module = importer(p[3])           # Load and parse child file
for decl in module.decls:
    p[0].declare(decl)             # Merge declarations
p[0].modules.update(module.modules)  # <-- MERGE MODULE REGISTRY
```

### Current Go include handler (parser/decl.go:2104-2127)
```go
decls, err := p.Importer(name)    // Load and parse child file
return decls                       // Return declarations only
// MISSING: no merge of child parser's modules map
```

## Fix

### Approach: Return modules alongside declarations from the Importer

The Importer callback signature needs to return the child parser's module registry in addition to declarations. The cleanest way: change the Importer callback type to return a `ParseResult` struct that carries both.

### Step 1: Define ParseResult type

**File: `parser/parser.go`**

```go
// ParseResult holds the output of parsing a file.
// Matches Python's Ivy object which has both .decls and .modules.
type ParseResult struct {
    Decls   []ast.Node
    Modules map[string]*ast.ModuleDecl
}
```

Change Parser fields:
```go
type Parser struct {
    // ... existing fields ...
    Importer func(name string) (*ParseResult, error)  // Changed return type
}
```

Change `Parse()` to return `*ParseResult`:
```go
func (p *Parser) Parse() (*ParseResult, error) {
    // ... existing parsing loop ...
    return &ParseResult{
        Decls:   decls,
        Modules: p.modules,
    }, err
}
```

### Step 2: Merge modules in parseIncludeDeclMulti

**File: `parser/decl.go`**

```go
func (p *Parser) parseIncludeDeclMulti(tok lexer.Token) []ast.Node {
    // ... existing name parsing, double-include check ...

    result, err := p.Importer(name)
    if err != nil {
        p.errorf("include %s: %v", name, err)
        return nil
    }
    // Merge child parser's module registry into parent
    // Python: p[0].modules.update(module.modules)
    for k, v := range result.Modules {
        p.modules[k] = v
    }
    return result.Decls
}
```

### Step 3: Update ImportModule and ReadModule

**File: `ivyinit/ivyinit.go`**

`ReadModule` currently returns `([]ast.Node, error)`. Change to return `(*parser.ParseResult, error)`:

```go
func ReadModule(filename string, nested bool) (*parser.ParseResult, error) {
    // ... existing file reading, version detection ...
    p := parser.New(s, version)
    p.Included = globalIncluded
    p.Importer = func(name string) (*parser.ParseResult, error) {
        return ImportModule(name)
    }
    result, parseErr := p.Parse()
    if parseErr != nil {
        return nil, fmt.Errorf("parse error in %s: %w", filename, parseErr)
    }
    return result, nil
}

func ImportModule(name string) (*parser.ParseResult, error) {
    fname := name + ".ivy"
    if _, err := os.Stat(fname); err != nil {
        stdDir := iu.GetStdIncludeDir()
        fname = filepath.Join(stdDir, fname)
        if _, err := os.Stat(fname); err != nil {
            return nil, fmt.Errorf("module %s not found", name)
        }
    }
    return ReadModule(fname, true)
}
```

### Step 4: Update SourceFile

**File: `ivyinit/ivyinit.go`**

`SourceFile` calls `ReadModule` and passes `decls` to the compiler:

```go
func SourceFile(filename string, mod *module.Module, sig *il.Sig, kwargs map[string]interface{}) error {
    result, err := ReadModule(filename, false)
    if err != nil {
        return err
    }
    // Compile the declarations
    comp := compiler.New(sig, mod)
    di := compiler.NewDomainSetup(comp)
    if err := di.ProcessDecls(result.Decls); err != nil {
        return err
    }
    // ... rest unchanged
}
```

### Step 5: Update any other callers of ReadModule/Parse

Search for all call sites of `ReadModule`, `ImportModule`, and `parser.Parse()` and update them to use `ParseResult`:

- `ivyinit/ivyinit.go:IvyInit()` — uses ReadModule
- Any test files

## Unit Tests

### Test 1: Module registry merging from includes
```go
func TestIncludeMergesModules(t *testing.T) {
    // Parent source: "include child\ninstance x : mymod"
    // Child source defines: "module mymod = { type this }"
    // After include, parent parser's p.modules should contain "mymod"
    parent := parser.New("\ninclude child\ninstance x : mymod", lexer.Version{1, 8})
    parent.Importer = func(name string) (*parser.ParseResult, error) {
        child := parser.New("\nmodule mymod = { type this }", lexer.Version{1, 8})
        return child.Parse()
    }
    result, err := parent.Parse()
    if err != nil {
        t.Fatalf("parse failed: %v", err)
    }
    // The instance should have been expanded (not left as InstantiateDecl)
    // because mymod should be in the parent's module registry after include merge
    foundModule := false
    for _, d := range result.Decls {
        if _, ok := d.(*ast.TypeDecl); ok {
            foundModule = true // expanded module body contains a TypeDecl
        }
    }
    if !foundModule {
        t.Error("module from include was not expanded into instance")
    }
}
```

### Test 2: End-to-end order.ivy include
```go
func TestIncludeOrder_UnboundedSequence(t *testing.T) {
    iu.SetStdIncludeDir("../ivy-lang-examples/ivy/include/1.8")
    defer iu.SetStdIncludeDir("")
    iu.SetStringVersion("1.8")

    src := "#lang ivy1.8\ninclude order\ninstance x : unbounded_sequence\n"
    tmpFile := writeTempFile(t, src)
    mod := module.New()
    mod.Cfg = module.NewConfig()
    err := ivyinit.SourceFile(tmpFile, mod, mod.Sig, map[string]interface{}{"create_isolate": false})
    if err != nil {
        t.Fatalf("SourceFile failed: %v", err)
    }
}
```

### Test 3: Nested includes with module propagation
```go
func TestNestedIncludeModulePropagation(t *testing.T) {
    // grandchild defines module M
    // child includes grandchild
    // parent includes child, then uses instance x : M
    // M should be available in parent
}
```

## Files to Modify

| File | Change |
|------|--------|
| `parser/parser.go` | Add `ParseResult` struct; change `Parse()` to return `*ParseResult`; update `Importer` field type |
| `parser/decl.go` | Update `parseIncludeDeclMulti` to merge `result.Modules` into `p.modules` |
| `ivyinit/ivyinit.go` | Update `ReadModule`, `ImportModule`, `SourceFile`, `IvyInit` to use `*ParseResult` |
| `parser/parser_test.go` | Add module merge tests |
| `ivyinit/ivyinit_test.go` | Add end-to-end include tests |

## Verification

1. `go build ./...` compiles
2. `go test ./...` — all tests pass
3. `./goivy_check ivy-lang-examples/test/action1.ivy` — still works
4. `./goivy_check isolate=cf_live ord_live.ivy` — gets past "unbounded_sequence undefined"
