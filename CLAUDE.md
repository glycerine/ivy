This project is goivy. goivy is meant to be a MECHANICAL PORT of 
the Ivy python project into Go. We are not porting the ivygo/cppgen/ any longer; that old stub was deleted.

A. Source of truth: the python version 

The python version of Ivy is the source of truth, and
all new features and bug fixes should conform to its example. It is:
~/goivy/pyivy/ivy

The Go port (in ~/goivy/goivy ) must follow the Python for all execution flow.

The goivy Go port, which should conform to the original Python Ivy logic:
~/goivy/goivy

B. MECHANICAL PORT RULES:

1. Python class names -> Go struct names: SAME NAME. App stays App. Atom stays Atom. Symbol
stays Symbol. Variable stays Variable. Do not rename.

2. Python function names -> Go function names: SAME NAME with capital first letter.
substitute_ast -> SubstituteAst. Do not invent new names.

3. Python field names -> Go field names: SAME NAME with capital first letter. self.rep ->
Rep. self.args -> Args. Do not rename fields.

4. One Python file -> one Go file with the same base name. ivy_logic_utils.py ->
ivy_logic_utils.go. Do not reorganize into different packages.

5. Do not merge Python classes. If Python has App and Atom as separate classes, Go has App
and Atom as separate structs.

6. Do not omit functions. If a Python function exists, a Go function must exist 
with the same name (different capitalization and substituting PascalCase for snake_case is allowed).

7. Do not add abstractions, interfaces, or helper types that don't exist in Python.

8. When in doubt, translate literally. A wrong but literal translation is easier to fix than a creative one.

9. "Correct in practice", and "good enough for now", and "simplest correct things for now" are lazy shortcuts we do not tolerate. Never slack off a task with these lazy excuses. These excuses for not doing a faithful port just waste time since then we need to do the item again. Deeply pursue the goal, no matter how large the change appears.

C. NO GLOBAL VARIABLES — Config System for Python Globals

Python Ivy uses module-level globals freely (counters, flags, mutable state).
Go goivy does NOT. All mutable state that was a Python global must live on a
per-package Config struct, never as a Go package-level `var`.

WHY: We run thread pools of ivy models on multi-core machines. Package-level
vars are shared across goroutines and break multi-tenancy. Each concurrent
Ivy session gets its own Config instances.

HOW TO PORT A PYTHON GLOBAL:

1. Find or create the package's Config struct. Most packages already have one.
   Examples:
     ast/config.go          → AstConfig
     lalr_full/config.go    → ParserConfig
     ivyutils/config.go     → IvyUtilsConfig
     interp/interp.go       → InterpConfig
     transrel/phase4.go     → TransrelConfig
     codegen/codegen.go     → CodegenConfig
     isolate/isolate.go     → IsolateConfig
     actions/action.go      → ActionsConfig
     module/config.go       → module.Config (top-level hub)

2. Add the former-global as a FIELD on the Config struct:
     Python:  label_counter = 0          (module-level global)
     Go:      type ParserConfig struct {
                  LabelCounter int       // was Python label_counter
              }

3. Make the function that used the global into a METHOD on *Config:
     Python:  def newlabel(pref): ...    (reads global label_counter)
     Go:      func (cfg *ParserConfig) NewLabel(pref string) *ast.Atom { ... }

   This is PREFERRED over adding cfg as a function parameter because it
   guides future use — callers must have a config to call the method.

4. Thread the Config from entry points:
   - Parse entry: ParseV17 creates ParserConfig with AstCfg field
   - Compile entry: IvyCompile receives *module.Module → mod.Cfg.AstCfg
   - Tests: create a fresh config locally, e.g. cfg := ast.NewAstConfig()

5. For AST node constructors specifically:
   - All ast.New* constructors are methods on *ast.AstConfig:
       cfg.NewAtom("x")    NOT  ast.NewAtom("x")
   - Every AST node carries cfg via Base.Cfg (set by the constructor).
   - Clone methods access cfg from the receiver: c.Cfg.SomeCounter++
   - In grammar actions: acfg(v17lex).NewAtom(...)
   - In compiler: mod.Cfg.AstCfg.NewAtom(...)

6. NEVER create a package-level `var` for mutable state. NEVER create a
   `var DefaultFooConfig = NewFooConfig()` transitional global. Port it
   correctly the first time by threading the Config through callers.

7. module.Config is the top-level hub. Sub-package configs are fields on it:
     type Config struct {
         AstCfg  *ast.AstConfig
         IuCfg   *ivyutils.IvyUtilsConfig
         // ... other sub-configs
     }

8. Packages that cannot import module (like ast/) have their own standalone
   Config struct. The caller provides it.

9. Read-only state is exempt: compiled regexps, singleton constants,
   goyacc parser tables, and init-once maps are safe as package-level vars.
   Only MUTABLE state (counters, flags, accumulated lists) must be on Config.

10. Debug-only globals in vprint.go files are exempt (not production state).

D. never use git. I commit in the background, so git is off limits to you.

D2. NEVER search, read, grep, or glob files under these directories:
   - already_applied_plans/
   - ivy-lang-examples/
   These are in .claudeignore. Skip any search results from them.

E. All plans produced should have the creation date and creation time just after their title.

F. Canonical S-expression System for Cross-Language AST Comparison

Both Go and Python parsers can produce canonical s-expression strings via Canon()/canon()
methods on every AST node. These are used to verify that both parsers produce identical
internal state — not just that the same productions fire, but that the resulting AST
trees are structurally identical.

1. Architecture overview:

   Go side:
   - ivyutils/canon.go        — defines Canonical type and Canonizer interface
   - ast/ast.go               — Base.canonFields(), nodeCanon(), sliceCanon() helpers
   - ast/decl.go              — DeclBase.canonFields()
   - ast/canon_decl.go        — Canon() for all declaration types
   - ast/formula.go           — Canon() inline on formula types
   - ast/tactic.go            — Canon() inline on tactic types
   - ast/sort.go              — Canon() inline on sort types
   - logic/canon.go           — Canon() wrapping Sexp() for logic IR types

   Python side:
   - ~/pyivy/ivy/ivy/canon.py      — helpers: node_canon, slice_canon, lineno_fields, decl_fields
   - ~/pyivy/ivy/ivy/canon_ast.py  — monkey-patches canon() onto all AST/action classes
                                      Call canon_ast.install() once at startup.

2. S-expression format rules (both sides MUST agree):

   - Struct:  (typeName field:value field2:value)
   - Slices:  [elem1 elem2 elem3]   (space separated, NO commas)
   - Maps:    (hash key:value key2:value2)  (sorted keys)
   - Strings: "double quoted"
   - Bool:    true / false
   - Int:     42
   - Nil:     nil
   - Single space between fields, no commas anywhere
   - Struct names: lowercase first letter (atom, symbol, axiomDecl)
   - Field names: lowercase first letter, matching Go field names

3. Flattening rule — Go struct embedding is flattened:

   Go has Base embedded in every type, and DeclBase embedded in Decl types.
   These are promoted (flattened) into the parent s-expression. Python has
   flat class inheritance so this is natural there.

   Example:  (axiomDecl lineno:42 declArgs:[...] attributes:[] common:nil)
   NOT:      (axiomDecl declBase:(declBase base:(base ...) ...))

   Go uses Base.canonFields() and DeclBase.canonFields() to produce the inline
   field strings. Python uses lineno_fields() and decl_fields() helpers.

4. Field name mapping (Go field name is canonical; Python maps to it):

   Go Atom.Terms      → Python Atom.args      → canon key: terms:
   Go Atom.ASort      → Python Atom.sort       → canon key: aSort:
   Go App.Rep         → Python App.rep         → canon key: rep:
   Go Variable.VSort  → Python Variable.sort   → canon key: vSort:
   Go DeclBase.DeclArgs → Python Decl.args     → canon key: declArgs:
   Go Forall.Bounds   → Python Quantifier.bounds → canon key: bounds:
   Go Forall.Body     → Python Quantifier.args[0] → canon key: body:
   Go Sequence.Stmts  → Python Sequence.args   → canon key: stmts:
   Go CallAction.Elems → Python CallAction.args → canon key: elems:
   Go CallAction.UniqueID → Python CallAction.unique_id → canon key: uniqueID:

5. How to debug a canon mismatch:

   a. Get the Go canon output for a declaration:
      In Go, after parsing: fmt.Println(decl.Canon())

   b. Get the Python canon output:
      from ivy.ivy.canon_ast import install; install()
      print(decl.canon())

   c. Diff the two strings. The first difference tells you exactly which
      field diverges. Common causes:

      - Wrong field value: the grammar rule set a field incorrectly.
        Fix the grammar rule's semantic action.

      - Missing field: Go has a field that Python doesn't emit (or vice versa).
        Check if a field was forgotten in the canon() method, or if the Go struct
        has a field that Python doesn't populate.

      - Wrong type name: e.g., Go emits (assumeGlobalTactic ...) but Python
        emits (assumeTactic ...). This means the wrong AST type is being
        constructed in the grammar rule.

      - Slice ordering: elements in a different order. This usually means
        the grammar rule appends children in a different sequence.

      - lineno difference: one side set lineno and the other didn't. Check
        getLineno/get_lineno calls in the grammar rule.

6. How to adjust canon() when adding new AST types:

   Go:
   - Add Canon() method to the new type. If it embeds Base, start with:
     func (x *NewType) Canon() iu.Canonical {
         return iu.Canonical(fmt.Sprintf("(newType %v field:%v)", x.Base.canonFields(), ...))
     }
   - If it embeds DeclBase, use x.DeclBase.canonFields() instead.
   - Always emit ALL fields, even if empty (use sliceCanon for empty slices → "[]").

   Python:
   - Add to canon_ast.py install() function:
     def _newtype_canon(self):
         return '(newType {} field:{})'.format(lineno_fields(self), ...)
     module.NewType.canon = _newtype_canon

7. How to adjust canon() when fixing a field mismatch:

   If Python emits (atom lineno:0 rep:"foo" terms:[] aSort:nil) but Go emits
   (atom lineno:0 rep:"foo" terms:[] aSort:(constantSort lineno:0 elems:[])),
   the problem is that Go's Atom.ASort is set to a ConstantSort while Python's
   Atom.sort is None. Fix the grammar rule that creates the Atom — don't change
   the canon() method unless the method itself is wrong.

   Rule of thumb: canon() methods should be a faithful mirror of the struct/class
   fields. If the canon output differs, the fix is almost always in the parser
   or grammar rule, not in the canon method.

8. Slice fields must ALWAYS be emitted (even when empty):

   Use sliceCanon() in Go and slice_canon() in Python. These produce "[]" for
   empty slices. Do NOT conditionally omit slice fields — this breaks cross-language
   matching since the other side may emit "[]".
