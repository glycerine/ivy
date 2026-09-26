# Upstream Fix Audit

Generated: 2026-09-26

Scope: audit the 54 upstream commits listed in `/home/jaten/ivy/upstream.fork.recent.diff.txt`, starting at `28cce61b66695ea5c3ac561d24d6c881b4929bcf`, looking for bug fixes worth considering for the local Python fork and the Go port. This document filters out new upstream features such as VMT, hardware/register/wire syntax, `ivy_to_vmt`, trace replay, derived invariants, and `using`/`patdef`, unless a commit contains a separable bug fix.

Priority legend:

- P0: correctness bug that can silently prove/check/generate the wrong thing, or produce incorrect executable semantics.
- P1: real crash, rejected-valid-program bug, or current-toolchain compatibility failure.
- P2: cleanup, diagnostics, runtime nicety, or bug in filtered/example code.
- Conditional: real upstream bug fix, but it applies to a feature this fork/Go port does not appear to have yet.
- Skip: feature/example/docs/merge-only change outside the requested scope.
- Done: already applied locally.

## Local Observations

- Go `actions_phase3.go:PCA` still appends `.ivy` unconditionally. Upstream `054188a` is relevant.
- FIXED: Go `actions_phase3.go:DestrAsgnVal` and `actions_update.go:destrAsgnVal` now use the upstream `c94e328` shape: `mkAssignClauses(mut, nondet(mut.args))`.
- Go `compiler.go:CompileApp` applies explicit result-sort ascription only to numerals. It does not have upstream `ascribe_result_sort` for general applications/constants.
- Go `compiler_decl.go:DefinitionDecl` still compiles definitions immediately and appends `DerivedUpdate` immediately. There is no `fix_definitions(mod)` pass like upstream `01d1ace`.
- Go `compiler_ivy_compile.go:AttachProofs` accepts labels found in props/conjs or `mod.Isolates`; it does not explicitly accept `lab == "this"` like upstream `28cce61`.
- Go `ivy2cpp/thunk.go` still emits direct `z3::expr(... Z3_parse_smtlib2_string ...)`; upstream `c849303` changed to parse an assertion vector and wrap with `z3::mk_and`.
- Go `proof_skolem.go` and `proof_skolem_test.go` already contain the `ConstantDecl` early return corresponding to upstream `8f7763f`.
- I did not find a local `invardeps` implementation in either the local Python fork or Go source, so zero-delay invariant-dependency fixes are documented as conditional on importing that feature.

## Recommended First Wave

1. P0 FIXED `c94e3286afb8cb320cae8c1e854c942e6fdac2d0`: assignment VC generator bug in destructor-chain assignments.
2. P1 `054188a5789d78f95eb3c93697890e017f5e387f`: `assert=file.ivy:line` should not become `file.ivy.ivy`.
3. P1 `d5e30fb67a6470caddeae597b7d1d87deb2ccec0`: explicit result-sort ascriptions are dropped except for numerals.
4. P1 `01d1ace4f0793d044ce5df92f2d5b893b2c1c517`: definitions are compiled too early, rejecting valid forward references.
5. P1 `18ccb0008f9198a2f659c106869a0ecb1b940e8f`: unnamed assertions must not read `.name` from a nil label.
6. P1 `c84930320bca1e7923324f5d670a10454217e80a`: generated C++ Z3 code should use the current `Z3_parse_smtlib2_string` vector API.
7. P1 `28cce61b66695ea5c3ac561d24d6c881b4929bcf`: `proof [this]` should attach to the current isolate.
8. Done `8f7763fdf643d8fd75ed9bc061f6e977742f45d6`: `var_subst_goal` returns `ConstantDecl` unchanged; already applied locally.

Conditional but severe if their feature is adopted:

- P0 conditional `2129b5cccf689bd0127e962aec23a806784ecfa1`: invisible zero-delay invariant dependency must be an error, not silently dropped.
- P1 conditional `dfb65f3704e2033858a7c4984be37717d3f38e44` and `a5ecac6e49e074e726533ac4f9892478540eae56`: unlabeled invariants/conjectures must not crash invariant-dependency lookup.
- P1/P2 conditional parts of `bba0acf0c0547e05b9f3b5a985c9fe57d865bb4c`: several fixes are hidden inside the register-feature commit.

## Commit Ledger: All 54 Commits

| # | Commit | Rank | Decision |
|---:|---|---|---|
| 1 | `a5ecac6e49e074e726533ac4f9892478540eae56` | P1 conditional | Fixes unlabeled conjectures in zero-delay dependency visibility checks. Applies only if invariant dependencies are imported. |
| 2 | `901af252c5d1684cf553e1ba96e5a93e20a0ee39` | Skip | Hardware example tuning for derived invariants. |
| 3 | `0b1561ecf5f2583595bc1769324f8bdb808bdd55` | Skip/conditional | New `derived invariant` feature. If imported, also import commits 1, 5, and 21. |
| 4 | `8c68700393f8e406e94c25f16b43d2554cae8956` | Skip | New `using`/`patdef` invariant-selection feature. |
| 5 | `2129b5cccf689bd0127e962aec23a806784ecfa1` | P0 conditional | Invisible zero-delay dependency was silently dropped. Serious if invariant dependencies exist. |
| 6 | `cefc289f54ea8fc026120b2df28afafc9b194e2b` | Skip | Documentation. |
| 7 | `8b3abf741835982fe190ca27386740487c337de2` | Skip | Hardware docs/examples. |
| 8 | `117dd30cf15bda9883265923aee9ab51df6cc171` | Skip | Example adjustment. |
| 9 | `c1795273aa8dc210cdef24e84cfcb4a77c5b01e3` | P2 | Removes stray debug printing from `ivy_l2s.py`. |
| 10 | `01d1ace4f0793d044ce5df92f2d5b893b2c1c517` | P1 | Allows forward references in definition statements. |
| 11 | `bba0acf0c0547e05b9f3b5a985c9fe57d865bb4c` | P1/P2 conditional | Register feature commit with several separable bug fixes. Do not port wholesale. |
| 12 | `45b2f13d133de22fc11ec942f3af703f2eef4666` | Skip | Hardware/bfe/concat syntax work. |
| 13 | `9cb5cb9e7a74f73bda46b381b88abe847bdc06b3` | Skip/conditional | Variadic concat feature; review only if bit-vector concat is being aligned. |
| 14 | `18ccb0008f9198a2f659c106869a0ecb1b940e8f` | P1 | Fixes regression on unnamed assertions in Ivy 1.5. |
| 15 | `44770493355173408f54aeec7aef6d0cd40dad32` | Skip | `trace_dir` feature. |
| 16 | `054188a5789d78f95eb3c93697890e017f5e387f` | P1 | `assert=` extension handling bug; `check=` feature optional. |
| 17 | `44eb960100f55c539643f0da32218b2425593a99` | Skip/P2 | Trace replay feature; some diagnostic polish. |
| 18 | `d5e30fb67a6470caddeae597b7d1d87deb2ccec0` | P1 | Explicit result-sort ascription bug. Wire half is conditional on wire support. |
| 19 | `0b959984716b2852072721b6fb4f575c8ac056cc` | Skip | Tests/examples for filtered features. |
| 20 | `532fc86aca6744dfe1f224965f3853f188c73a7b` | P2 conditional | Wire-feature isolate export regression. |
| 21 | `dfb65f3704e2033858a7c4984be37717d3f38e44` | P1 conditional | Crash fix for unlabeled invariants in invariant-dependency lookup. |
| 22 | `c84930320bca1e7923324f5d670a10454217e80a` | P1 | Generated C++ Z3 compatibility with current Z3/macOS. |
| 23 | `46cd023ebd6c2d57253fa5e5e252e5a424f732b4` | P2 | Bind local TCP/UDP examples to loopback rather than any interface. |
| 24 | `ea9a5149ab593611f73ea2714ce172523054263c` | Skip | Hardware bfe/concat operator/example feature. |
| 25 | `001da592c1738ff18b62eee07672249eb2c10a96` | Skip/conditional | New invariant-dependency feature. |
| 26 | `3893ae6e97fd375a7ce0b56c84d2164873e8d302` | Skip | Wire feature. |
| 27 | `737e6b0ab620e8585c522ef854c2f54db7c4b009` | Skip | GUI/Tix cleanup. |
| 28 | `f5dffa7a8bbb36c75527765d878f7961e08cb80d` | Skip/P2 | Trace replay feature and diagnostic polish. |
| 29 | `912e29603ee6b92502d4d1d82197f0683eacdc4e` | Skip | `ticket_ranking` example. |
| 30 | `f5ac0875253aaafaef96493305af54f550db598d` | Skip | CAV 2020 examples. |
| 31 | `992566a0de84a2163aa8b0005ce512dd409c85da` | Skip | Documentation and local variable rename. |
| 32 | `346fe2d7a0c55baf337c21bcb589fa8daf6cd810` | P1 Python | Python setup/Z3 context reset. Mostly Python-side. |
| 33 | `06cbcf362d8ae261d0629e7605e7ef7f9b3251b7` | P1 Python | External `z3`, enum-sort cache, reset enum cache. Mostly Python-side. |
| 34 | `c94e3286afb8cb320cae8c1e854c942e6fdac2d0` | P0 FIXED | Assignment VC generator bug. |
| 35 | `7431454b44e4e60511e6eea3d825c345e6e9c5ff` | Skip | Shrink option feature. |
| 36 | `2dfa55f94c112e57bb9d43b613478f7d1db70a5c` | P2 | Strip/remove `ext:` tags in generated naming. |
| 37 | `f141f5fdf45c4400e50bbe9eac28ac50534a35db` | P2 | Remove debug solver-state file emission. |
| 38 | `dd87dde9e60de72021710b954056fb60fd0ce1f1` | P2 | Remove debug print. |
| 39 | `e34f3cd8fe5f6d581805d53ae6b007aed5113213` | P1 | Ranking proof bug when helpful condition has temporal operators. |
| 40 | `df843284e0c535a7b3e30c3416af413c38f77e4b` | P2 conditional | VMT/debug cleanup; nil-update guard appears already present in Go check path. |
| 41 | `d8e4ea35effd3a419a5ed5674db7f06ba3c59bee` | Skip | Merge. |
| 42 | `33fbdea22d742230c35d847e3bdc3d7b4552af92` | Skip | Large VMT/datatypes/structs feature. |
| 43 | `4e4747745a98ffba5ec6254f4fb0034f3c408c89` | Skip | Merge. |
| 44 | `c7f10b9127ef0b86dc40ec3df0dae545b9447b21` | P2 | TCP model bug: missing `id` argument in `connect`. |
| 45 | `2189d68a484f9c5028f135a8bc7a5c2d123d5982` | Skip | Ranking-infer file addition. |
| 46 | `494e37692d005aefce651a3ed7b4cba0e2654643` | Skip | Ranking-infer tests. |
| 47 | `435635b23220193a1e6cc353b875824ea3fd3cc3` | Skip | Ranking-infer tactic feature. |
| 48 | `18c08b3c42179781bf99291764539e43a7e46128` | Skip/conditional | VMT/lambda work. Local Python already has simple exists/forall branch correct. |
| 49 | `3a2feca0abeddc308ba7696e1cfb139f0240600c` | Skip | VMT lambda feature start. |
| 50 | `4773d57ba991351ade8d495a5f24ea986c062d3c` | Skip | VMT axioms/uninterpreted symbols feature. |
| 51 | `1f8039cc26ae7aeb123b14e08692e53f3a30559c` | Skip | Alternate VMT engine. |
| 52 | `28cce61b66695ea5c3ac561d24d6c881b4929bcf` | P1 | `proof [this]` fix. |
| 53 | `8f7763fdf643d8fd75ed9bc061f6e977742f45d6` | Done | `var_subst_goal` returns `ConstantDecl` unchanged. Already applied locally. |
| 54 | `925cedadc9cb67ec211df8e7f137e9a227c03230` | Skip | CAV 2024 artifact docs/baseline before fork point. |

## Detailed Fix Candidates

### P0 FIXED `c94e3286afb8cb320cae8c1e854c942e6fdac2d0`: assignment VC generator

Upstream subject: `fixed bug in VC generator for assignments`

Why it matters: the old VC construction used placeholder arguments for the mutated relation/function instead of the actual receiver/index arguments being assigned through. That can make the generated transition relation too weak or about the wrong update point.

Python patch:

```diff
 def destr_asgn_val(lhs,fmlas):
     mut = lhs.args[0]
     rest = list(lhs.args[1:])
     mut_n = mut.rep
     if mut_n.name in ivy_module.module.destructor_sorts:
         lval,new_clauses,mutated = destr_asgn_val(mut,fmlas)
     else:
         nondet = mut_n.suffix("_nd").skolem()
-        new_clauses = (mk_assign_clauses(mut_n,nondet(*sym_placeholders(mut_n))))
+#        new_clauses = (mk_assign_clauses(mut_n,nondet(*sym_placeholders(mut_n))))
+        new_clauses = (mk_assign_clauses(mut,nondet(*mut.args)))
         lval = nondet(*mut.args)
         mutated = mut_n
```

Go translation:

- Target: `actions_phase3.go:DestrAsgnVal`.
- Current Go builds `phs := SymPlaceholders(mutSym)` and calls `mkAssignClauses(mutSym, nondetApp)`.
- Build `nondetApp` from `nodeArgs(mut)` and call `mkAssignClauses(mut, nondetApp)`.
- `mkAssignClauses(lhs, rhs Expr)` already accepts `Expr`, so no signature change should be needed.
- Preserve existing `lval = nondet(*mut.args)` behavior.

TDD test guidance:

- Add a fast unit test near `actions_phase3_test.go`.
- Construct a destructor-chain assignment where the mutated base is an application with concrete args, e.g. the Go AST equivalent of assigning through `field(obj(k), i)`.
- Call `DestrAsgnVal` directly.
- Red assertion: generated clauses mention only placeholder variables for the mutated relation/function.
- Fixed assertion: generated clauses use the concrete `mut` application and its actual args.

### P1 `054188a5789d78f95eb3c93697890e017f5e387f`: `assert=` extension handling

Upstream subject: `Add check= option to select checks by name; fix assert= extension handling`

Bug patch:

```diff
 def p_c_a(s):
     a = s.split(':')
-    return iu.Location(a[0]+'.ivy',int(a[1]))
+    fn = a[0] if a[0].endswith('.ivy') else a[0]+'.ivy'
+    return iu.Location(fn,int(a[1]))
```

The `check=` option in the same commit is a feature. The small filename normalization is the relevant bug fix.

Go translation:

- Target: `actions_phase3.go:PCA`.
- Replace unconditional `parts[0] + ".ivy"` with a suffix check using `strings.HasSuffix`.

TDD test guidance:

- Direct unit test for `PCA`.
- `PCA("foo:12")` should produce `foo.ivy`, line 12.
- `PCA("foo.ivy:12")` should also produce `foo.ivy`, line 12. This should be red now.

### P1 `d5e30fb67a6470caddeae597b7d1d87deb2ccec0`: result-sort ascriptions

Upstream subject: `Fix two type/wire compiler bugs affecting the hardware models`

Relevant Python patch:

```diff
+def ascribe_result_sort(term,sort):
+    if ivy_logic.is_topsort(sort):
+        return term
+    if isinstance(term,ivy_logic.App):
+        fsort = term.func.sort
+        if isinstance(fsort,ivy_logic.FunctionSort):
+            newfunc = ivy_logic.Symbol(term.func.name,
+                                       ivy_logic.FunctionSort(*(list(fsort.dom) + [sort])))
+            return newfunc(*term.args)
+    elif ivy_logic.is_constant(term):
+        return ivy_logic.Symbol(term.name,sort)
+    return term
+
 def compile_app(self,old=False):
@@
-    sym = rep.cmpl() if isinstance(rep,ivy_ast.NamedBinder) else ivy_logic.Equals if rep == '=' else ivy_logic.find_polymorphic_symbol(rep,throw=False) 
+    sym = rep.cmpl() if isinstance(rep,ivy_ast.NamedBinder) else ivy_logic.Equals if rep == '=' else ivy_logic.find_polymorphic_symbol(rep,throw=False)
+    ascribed = hasattr(self,'sort') and self.sort != 'S'
     if sym is not ivy_logic.Equals:
         if ivy_logic.is_numeral(sym):
-            if hasattr(self,'sort') and self.sort != 'S':
+            if ascribed:
                 sym = ivy_logic.Symbol(sym.name,variable_sort(self))
+                ascribed = False
     if sym is not None:
         sym = old_sym(sym,old)
-        return (sym)(*args)
-    res = compile_field_reference(rep,args,self.lineno,old=old)
+        res = (sym)(*args)
+    else:
+        res = compile_field_reference(rep,args,self.lineno,old=old)
+    if ascribed:
+        res = ascribe_result_sort(res,variable_sort(self))
     return res
```

Wire-specific half of this commit: skip `DerivedUpdate` for separately declared-and-defined wires. Keep that conditional unless wire support is being touched.

Go translation:

- Target: `compiler.go:CompileApp`.
- Add helper `ascribeResultSort(term Expr, sort Sort) Expr`.
- Compute `ascribed := n.ASort != nil && compilerExtractSortRep(n.ASort) != "S"` before building the result.
- Keep numeral special handling and clear `ascribed` after applying it to a numeral symbol.
- For non-numeral symbol applications and field references, build the result first, then apply the helper.

TDD test guidance:

- Fast compiler unit test.
- Construct a polymorphic function/operator application with explicit result sort where argument sorts do not determine the range.
- Red behavior: compiled result does not have the explicit range sort or inference fails.
- Fixed behavior: compiled `Apply` has a function sort whose range is the ascribed sort.

### P1 `01d1ace4f0793d044ce5df92f2d5b893b2c1c517`: forward references in definitions

Upstream subject: `allow forward references in definition statements`

Patch shape:

```diff
 class IvyDomainSetup(IvyDeclInterp):
@@
     def definition(self,ldf):
-        label = ldf.label
-        df = ldf.formula
-        df = compile_defn(df)
-        self.add_definition(ldf.clone([label,df]))
-        if df.args[0].rep not in self.domain.wires:
-            self.domain.updates.append(DerivedUpdate(df))
-        self.domain.symbol_order.append(df.args[0].rep)
-        if not self.domain.sig.contains_symbol(df.args[0].rep):
-            add_symbol(df.args[0].rep.name,df.args[0].rep.sort)
+        # postpone compiling definitions so they can have forward references
+        defs = self.domain.native_definitions if isinstance(ldf.formula.args[1],ivy_ast.NativeExpr) else self.domain.labeled_props
+        defs.append(ldf)
+        self.last_fact = ldf
@@
+def fix_definitions(mod):
+    def fix(ldf):
+        label = ldf.label
+        df = ldf.formula
+        if isinstance(df,ivy_ast.Definition):
+            df = compile_defn(df)
+            ldf = ldf.clone([label,df])
+            mod.symbol_order.append(df.args[0].rep)
+            if not mod.sig.contains_symbol(df.args[0].rep):
+                add_symbol(df.args[0].rep.name,df.args[0].rep.sort)
+            if df.args[0].rep not in mod.wires:
+                mod.updates.append(DerivedUpdate(df))
+        return ldf
+    mod.labeled_props = [fix(x) for x in mod.labeled_props]
+    mod.native_definitions = [fix(x) for x in mod.native_definitions]
@@
         with TopContext(collect_actions(decls.decls)):
             IvyDomainSetup(mod)(decls)
+            fix_definitions(mod)
             fix_constructors(mod)
```

Also stringifies dependency-cycle SCC elements:

```diff
- raise iu.IvyError(None,'these definitions form a dependency cycle: {}'.format(','.join(scc)))
+ raise iu.IvyError(None,'these definitions form a dependency cycle: {}'.format(','.join(str(x) for x in scc)))
```

Go translation:

- Target: `compiler_decl.go:DefinitionDecl`, `compiler_decl.go:AddDefinition`, and pass ordering in `compiler_ivy_compile.go`.
- First pass should append raw labeled definitions to `Module.LabeledProps` or `Module.NativeDefinitions` and set `LastFact`.
- Add `FixDefinitions(mod)` after `DomainSetup` and before constructor/proof/property setup.
- `FixDefinitions` should compile definitions, replace stored labeled formulas with compiled clones, add symbols/order, and append `DerivedUpdate` only after successful compilation.

TDD test guidance:

- In-process compiler test with an earlier definition referring to a later definition/symbol.
- Red behavior: compile fails during immediate definition compilation.
- Fixed behavior: compile succeeds after deferred `FixDefinitions`.

### P1 `18ccb0008f9198a2f659c106869a0ecb1b940e8f`: unnamed assertions

Upstream subject: `fixed regression on unnamed assertions in ivy1.5`

Patch:

```diff
 class AssertAction(Action):
@@
         unprovable = False
         aname = None
         if isinstance(fmla,ivy_ast.LabeledFormula):
             unprovable = fmla.unprovable
-            aname = fmla.name
+            if fmla.label is not None:
+                aname = fmla.name
             fmla = fmla.formula
```

Related name helper pattern:

```diff
- if args and isinstance(args[0], LabeledFormula):
+ if args and isinstance(args[0], LabeledFormula) and args[0].args[0] is not None:
     return args[0].name
```

Go translation:

- Target: assert-action update code and any labeled-formula name helper.
- Ensure every extraction of a check/assertion name verifies `lf.Label != nil` first.

TDD test guidance:

- Construct an assert action wrapping a `LabeledFormula` with `Label == nil`.
- Call action update directly.
- Red behavior: panic/error extracting name. Fixed behavior: normal clauses/update.

### P1 `c84930320bca1e7923324f5d670a10454217e80a`: generated C++ Z3 compatibility

Upstream subject: `Fix Z3 usage in generated test code for current Z3 / macOS`

Key patches:

```diff
-    return 'z3::expr(g.ctx,Z3_parse_smtlib2_string({}ctx, "{}", ...))'.format(...)
+    return 'z3::mk_and(z3::expr_vector(g.ctx,Z3_parse_smtlib2_string({}ctx, "{}", ...)))'.format(...)
```

```diff
     void add(const std::string &z3inp) {
-        z3::expr fmla(ctx,Z3_parse_smtlib2_string(ctx, z3inp.c_str(), ...));
+        Z3_ast_vector x = Z3_parse_smtlib2_string(ctx, z3inp.c_str(), ...);
         ctx.check_error();
+        z3::expr_vector fmlas(ctx,x);
+        z3::expr fmla = z3::mk_and(fmlas);
         slvr.add(fmla);
     }
```

```diff
- if (Z3_get_numeral_int(ctx,foo,&v) != Z3_TRUE) {
+ if (Z3_get_numeral_int(ctx,foo,&v) != Z3_L_TRUE) {
```

Other patch pieces: prefer pip `z3.__file__` for headers/libs, add `find_openssl()` for macOS, and avoid registering enum constructors in one declaration path.

Go translation:

- Target: `ivy2cpp/thunk.go`, `ivy2cpp/thunk_z3_test.go`, and any generated Z3 helper analogous to `gen.add`.
- Emit `Z3_ast_vector`, `z3::expr_vector`, and `z3::mk_and` for parsed SMTLIB assertions.
- Update tests that currently expect direct `z3::expr(g.ctx,Z3_parse_smtlib2_string(...))`.
- Search generated C++ paths for `Z3_TRUE` around numeral getters and use `Z3_L_TRUE`.

TDD test guidance:

- Pure source-generation test, no C++ compile needed.
- Red assertion: generated thunk code lacks `z3::mk_and` and contains old direct parse shape.
- Fixed assertion: generated thunk code contains the vector/mk_and shape.

### P1 `28cce61b66695ea5c3ac561d24d6c881b4929bcf`: `proof [this]`

Patch:

```diff
-            elif lab in mod.isolates:
+            elif lab in mod.isolates or lab == 'this':
                 mod.isolate_proofs[lab] = pf[1]
```

Go translation:

- Target: `compiler_ivy_compile.go:AttachProofs` and parser label grammar if needed.
- Add `|| lab == "this"` to the isolate-proof branch.

TDD test guidance:

- Direct `AttachProofs` unit test with label-only proof `[this]` and no explicit `mod.Isolates["this"]`.
- Red behavior: no matching property/conjecture/isolate error.
- Fixed behavior: `mod.IsolateProofs["this"]` is populated.

### P1 `e34f3cd8fe5f6d581805d53ae6b007aed5113213`: ranking helpful condition with temporal operators

Patch:

```diff
         D = 0
-        E = 1
+        # E = 1
@@
+            E = 0 if ilg.has_temporal(work_helpful[1]) else 1
+
             # work_created, work_needed and work_done must have same sort
```

Go translation:

- Target likely: `check_ranking_tactic.go` and helper logic mirroring Python `l2s_tactic_int`.
- Replace any fixed assumption about helpful-body index with a temporal-aware choice.

TDD test guidance:

- Unit test a `work_helpful` definition whose body contains a temporal operator.
- Assert the tactic/helper chooses the correct body index and does not build obligations from the wrong term.

### P1 Python-side `06cbcf362d8ae261d0629e7605e7ef7f9b3251b7` and `346fe2d7a0c55baf337c21bcb589fa8daf6cd810`: Z3 import/cache/context

Relevant patches:

```diff
-import ivy.z3 as z3
+import z3
```

```diff
 def clear():
-    global z3_sorts, z3_predicates, z3_constants, z3_functions
+    global z3_sorts, z3_predicates, z3_constants, z3_functions, z3_enums
@@
+    z3_enums = dict()
```

```diff
 def enumeratedsort(es):
+    if es.name in z3_enums:
+        return z3_enums[es.name]
     res,consts = z3.EnumSort(es.name,es.extension)
@@
+    z3_enums[es.name] = res
     return res
```

```diff
 def clear():
+    z3.z3._main_ctx = z3.Context()
```

```diff
-        cmd = 'python scripts/mk_make.py --python --prefix {} --pypkgdir {}/'.format(cwd,ivydir)
+        cmd = 'python3 scripts/mk_make.py --python --prefix {} --pypkgdir {}/'.format(cwd,ivydir)
```

Go translation:

- No direct one-for-one port unless Go Z3 bridge has stale context/cache bugs.
- Check `z3bridge_session_cache.go`, `z3bridge_translate.go`, and enum-native tests only if failures appear.

### Done `8f7763fdf643d8fd75ed9bc061f6e977742f45d6`: `var_subst_goal` and `ConstantDecl`

Patch:

```diff
 def var_subst_goal(goal,subst):
     """ Apply a variable substitution to a goal. """
+    if isinstance(goal, ia.ConstantDecl):
+        return goal
     prems = [var_subst_goal(prem,subst) for prem in goal_prems(goal)]
```

Local status: already applied in Python and Go, with Go regression coverage in `proof_skolem_test.go`.

## Conditional Fix Bundles

### Invariant dependency / zero-delay `with` fixes

Affected commits: `001da592`, `dfb65f37`, `2129b5cc`, `a5ecac6e`, `0b1561ec`, `8c687003`.

Current local status: no local `invardeps` found, so these are not immediate bug ports. If invariant dependencies are imported later, apply these fixes immediately.

Crash fix for unlabeled invariants:

```diff
-        depnames = set(mod.invardeps.get(c.name,[]))
+        depnames = set(mod.invardeps.get(c.name,[])) if c.label is not None else set([])
```

Invisible dependency error:

```diff
-        deps = [x.formula for x in mod.assumed_invariants if x.name in depnames]
+        available = set(x.name for x in mod.assumed_invariants)
+        missing = sorted(n for n in depnames if n not in available)
+        if missing:
+            raise iu.IvyError(c,
+                "zero-delay dependency {} of this invariant is not visible in this isolate "
+                "(this can be caused by a missing entry in the `with` clause of the isolate)".format(
+                    ', '.join(missing)))
+        deps = [x.formula for x in mod.assumed_invariants if x.name in depnames]
```

Unlabeled conjectures/postconditions in visibility check:

```diff
-        available = set(x.name for x in mod.assumed_invariants)
-        available.update(x.name for x in conjs)
+        available = set(x.name for x in mod.assumed_invariants if x.label is not None)
+        available.update(x.name for x in conjs if x.label is not None)
@@
-        deps = [x.formula for x in (mod.assumed_invariants+conjs) if x.name in depnames]
+        deps = [x.formula for x in (mod.assumed_invariants+conjs) if x.label is not None and x.name in depnames]
```

Go guidance if feature is adopted:

- Add `Invardeps map[string][]string` or equivalent module field.
- In check-conjecture paths, only look up dependencies for labeled conjectures.
- Build `available` only from labeled assumed invariants and labeled current conjectures/postconditions.
- Raise an `IvyError` naming missing dependencies instead of silently omitting them.

### Register/wire/solver fixes hidden in `bba0acf0c0547e05b9f3b5a985c9fe57d865bb4c`

Do not port the register feature wholesale. Separable fixes:

Non-wire background theory for sequential composition:

```diff
-        axioms = domain.background_theory(pvars)
+        axioms = domain.non_wire_background_theory(pvars)
```

Assumed invariants included in interference symbol collection:

```diff
-    for x in [mod.labeled_axioms,mod.labeled_props,mod.labeled_inits,mod.labeled_conjs,non_wires]:
+    for x in [mod.labeled_axioms,mod.labeled_props,mod.labeled_inits,mod.labeled_conjs,non_wires,
+              mod.assumed_invariants]:
         asts.extend(y.formula for y in x if not isinstance(y.formula,ivy_ast.SchemaBody))
```

Safer `z3_and` helper:

```diff
+def z3_and(*args):
+    if len(args) == 1 and isinstance(args[0],list):
+        args = args[0]
+    if len(args) == 0:
+        return z3.BoolVal(True)
+    if len(args) == 1:
+        return args[0]
+    return z3.And(*args)
```

Port each only if the corresponding Go area lacks the behavior or tests show the bug. The Go Z3 bridge already appears to have empty/singleton conjunction tests.

### VMT/lambda commits

Filtered commits: `18c08b3`, `3a2feca`, `4773d57`, `1f8039c`, `33fbdea`, `df843284`.

The tempting general bug in `18c08b3` is:

```diff
-        q = forall if ivy_logic.is_forall(fmla) else exists if ivy_logic.is_forall(fmla) else mylambda
+        q = forall if ivy_logic.is_forall(fmla) else exists if ivy_logic.is_exists(fmla) else mylambda
```

But the local Python fork currently has the simpler non-lambda branch:

```python
q = forall if ivy_logic.is_forall(fmla) else exists
```

So the local fork does not appear to have that exact existential-as-lambda bug. Skip the VMT/lambda bundle unless lambda support is imported.

## P2 Cleanups and Runtime Nits

- `c1795273aa8dc210cdef24e84cfcb4a77c5b01e3`: remove `print('reset_w:')` and debug dump loop from `ivy_l2s.py`.
- `46cd023ebd6c2d57253fa5e5e252e5a424f732b4`: bind local TCP/UDP examples to loopback rather than any interface.
- `2dfa55f94c112e57bb9d43b613478f7d1db70a5c`: remove/strip internal `ext:` tags from generated naming.
- `f141f5fdf45c4400e50bbe9eac28ac50534a35db`: remove debug solver-state file emission.
- `dd87dde9e60de72021710b954056fb60fd0ce1f1`: remove debug print.
- `c7f10b9127ef0b86dc40ec3df0dae545b9447b21`: TCP model fix, `intf.connect(dst)` -> `intf.connect(id,dst)`.

## Suggested Future Fix Order

1. `c94e328` P0 FIXED assignment VC generator.
2. `054188` P1 `PCA`/`assert=` filename normalization.
3. `d5e30` P1 result-sort ascription, keeping the wire half separate.
4. `01d1ace` P1 deferred definition compilation.
5. `c849303` P1 `ivy2cpp` current-Z3 source generation.
6. `18ccb` and `28cce` P1 nil-label/proof-label robustness.
7. Feature bundles only when the feature is imported or touched: invariant dependencies, wires/registers, VMT/lambda, trace replay.
