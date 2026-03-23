# PY_STRUCTEQ.md — Python Structural Equality Audit

Generated: 2026-03-17

This document catalogs every use of structural equality on logic nodes
(recstruct objects: Var/Variable, Const/Symbol, Apply, ForAll, Exists,
Sort types, Clauses, etc.) in the Python Ivy codebase (`~/pyivy/ivy/ivy/`).

Python's `recstruct` gives all logic nodes `__eq__` and `__hash__` based
on their field values, so they can be used as dict keys, set members, and
compared with `==`. Go pointer structs do NOT have this property by default.

---

## 1. `logic.py`

### 1.1 ForAll/Exists `_preprocess_` — frozenset of variables
- **Lines 380, 398**: `frozenset(variables), body`
- Variables are stored in a frozenset (structural hash/equality)

---

## 2. `logic_util.py`

### 2.1 frozenset operations on Var objects
- **Line 22**: `frozenset.union(*sets)` — union of Var frozensets
- **Line 37**: `return frozenset((t,))` — Var in frozenset literal
- **Line 44**: `union(used_variables(t.body), t.variables)` — frozenset union
- **Line 68**: `return frozenset((t.name if by_name else t,))` — Var in frozenset
- **Line 75**: `_free_variables(t.body) - _free_variables(*t.variables)` — frozenset difference
- **Line 99**: `union(bound_variables(t.body), t.variables)` — frozenset union
- **Line 116**: `return frozenset((t,))` — Const in frozenset literal
- **Line 267**: `free_variables(t.body) & frozenset(t.variables)` — frozenset intersection

### 2.2 Dict key lookups with logic nodes
- **Line 150**: `if t in subs: return subs[t]` — Var/Const as dict key (substitute)
- **Lines 219, 232**: `if t.func in subs`, `if k not in t.variables` — Const/Var dict/frozenset membership

### 2.3 frozenset membership and set operations
- **Line 166**: `if k not in t.variables` — frozenset membership test on Var
- **Line 168**: `forbidden_variables.isdisjoint(t.variables)` — frozenset.isdisjoint on Var
- **Line 178**: `forbidden_variables.intersection(t.variables)` — frozenset.intersection on Var

### 2.4 Equality comparisons
- **Line 276**: `return type(t) is Eq and t.t1 == t.t2` — structural == on logic nodes (is_tautology_equality)
- **Line 312**: `return m1.get(t,t) == m2.get(u,u)` — Var == comparison (equal_mod_alpha)
- **Line 324**: `t.func == u.func` — Const == comparison (equal_mod_alpha Apply case)
- **Line 327**: `return t == u` — Const == comparison (equal_mod_alpha Const case)

---

## 3. `ivy_logic.py`

### 3.1 Clauses.defidx — dict with Symbol keys
- **Line 44**: `self.defidx = dict((d.defines(),d) for d in defs)` — Symbol as dict key

### 3.2 Set operations on Var objects
- **Line 498**: `univs.update(fmla.variables)` — set.update with Var objects
- **Line 501**: `univs.remove(v)` — set.remove with Var
- **Line 530**: `univs.update(fmla.variables)` — set.update with Var objects
- **Line 653**: `frozenset.union(uvars, term.variables)` — frozenset union on Var
- **Line 655**: `frozenset.intersection(lu.free_variables(term), uvars)` — frozenset intersection on Var

### 3.3 Sort membership tests
- **Line 930**: `assert symbol.sort in sort.sorts` — Sort in list membership
- **Line 931**: `sort.sorts.remove(symbol.sort)` — list.remove with Sort
- **Line 940**: `return symbol.sort in sort.sorts` — Sort in list membership

### 3.4 BindSymbols/BindSymbolValues — set/dict with Symbol keys
- **Line 1006**: `if sym in self.env`, `self.env.remove(sym)`, `self.env.add(sym)` — Symbol set ops
- **Lines 1024-1033**: `if sym in self.env`, `env[sym]`, `del self.env[sym]`, `self.env[sym] = val` — Symbol dict ops

### 3.5 ast_match — dict with logic node keys
- **Line 1215**: `if y in placeholders`, `if y in subst`, `subst[y] = x` — logic nodes as dict keys
- **Line 1220**: `return x == y` — structural == on logic nodes

### 3.6 VariableUniqifier — dict with Var keys
- **Lines 1647-1678**: `if v in vmap`, `vmap[v] = newv`, `self.invmap[newv] = fmla`, `del vmap[v]`, `vmap.update(...)` — Var as dict key throughout
- **Line 1674**: `freevars.update(lu.free_variables(fmla))` — set.update with Var objects
- **Line 1697**: `if not forbidden.isdisjoint(newvars)` — frozenset.isdisjoint on Var
- **Line 1698**: `forbidden.intersection(newvars)` — frozenset.intersection on Var

---

## 4. `ivy_logic_utils.py`

### 4.1 Clauses.defidx — dict with Symbol keys
- **Line 44**: `self.defidx = dict((d.defines(),d) for d in defs)` — Symbol as dict key

### 4.2 Sort as dict key
- **Line 407**: `if sort in subs: return subs[sort]` — Sort objects as dict keys (resort_sort)

### 4.3 Symbol set/dict membership
- **Line 1033**: `sym not in used_syms and sym in cls.defidx` — Symbol in set and dict membership
- **Lines 1035-1036**: `cls.defidx[sym]`, `d.args[0].rep in used_syms` — Symbol dict lookup and set membership
- **Lines 1305-1309**: `if sym in clauses.defidx`, `d.defines() not in deadset` — Symbol dict/set membership
- **Line 1320**: `defd = set(d.defines() for ...)` — set of Symbol objects
- **Lines 1341-1344**: `if s not in defidx`, `defidx[s] = d` — Symbol as dict key
- **Lines 1373-1380**: `if s not in defidx`, `defidx[s] = d` — Symbol as dict key
- **Line 1449**: `symset.remove(lhs)` — set.remove with logic node
- **Line 1622**: `sksubs = dict((c, skolemizer(...)) for c in cs)` — Const as dict key

---

## 5. `ivy_solver.py`

### 5.1 z3_sorts/z3_constants/z3_functions — dict with .rep keys
- **Line 246**: `z3_sorts[us.rep] = s` — Sort.rep as dict key
- **Line 260**: `z3_sorts[es.rep] = res` — EnumeratedSort.rep as dict key
- **Line 449**: `z3_constants[term.rep] = res` — term.rep as dict key
- **Line 464**: `z3_functions[term.rep] = fun` — term.rep as dict key

### 5.2 Variable.rep equality in env dict
- **Line 759**: `if x.rep in env:` — Variable.rep as dict key
- **Line 760**: `if y.rep != env[x.rep]:` — Variable.rep equality comparison
- **Line 762**: `env[x.rep] = y.rep` — Variable.rep as dict key
- **Line 764**: `if x.rep != y.rep:` — Variable.rep != comparison

### 5.3 Constant.rep as dict key in model operations
- **Line 1369**: `foom[c.rep] = num` — Constant.rep as dict key
- **Line 1390**: `m = dict((c.rep, ...) for ...)` — Constant.rep as dict key
- **Line 1405**: `reps[c.rep]` — Constant.rep dict access
- **Lines 1471-1476**: `mc.rep not in reps`, `reps[mc.rep] = ...`, `e.rep not in reps` — Constant.rep membership/assignment
- **Line 1506**: `repset = set(c.rep for e,c in reps.items())` — set of Constant.rep
- **Line 1593**: `if t.rep in ivy_logic.sig.constructors:` — Symbol.rep membership

---

## 6. `ivy_transrel.py`

### 6.1 Symbol set operations
- **Line 170**: `moded = set(updated)` — set of Symbol objects
- **Line 181**: `defnd = set(df.defines() for df in axioms.defs)` — set of Symbol objects
- **Line 312**: `mid = us1.intersection(us2)` — set intersection on Symbols
- **Line 331**: `new_updated = list(us1.union(us2))` — set union on Symbols
- **Line 641**: `img = set(renaming[s] for s in renaming if ...)` — set of Symbols

---

## 7. `ivy_actions.py`

### 7.1 Dict with .rep keys for substitution
- **Line 43**: `subst = dict((x.rep,y.rep) for x,y in zip(defn.args[0].args,params))` — .rep as dict key
- **Line 585**: `rn = dict((a.rep,v) for v,a in zip(vs,args) if isinstance(a,Variable))` — Variable.rep as dict key
- **Line 600**: `rn = dict((a.rep,v) for v,a in zip(vs,args) if isinstance(a,Variable))` — Variable.rep as dict key
- **Line 734**: `subst = dict((x.rep,y) for x,y in zip(fparams,aparams))` — .rep as dict key
- **Line 736**: `psubst = dict((x.rep,y.rep) for x,y in zip(...))` — .rep as dict key
- **Line 1087**: `subst = dict((a.args[0].rep,a.args[1].rep) for a in self.args[0:-1])` — .rep as dict key

### 7.2 Membership tests with .rep
- **Line 422**: `ast.rep not in domain.relations and ast.rep != "="` — .rep membership/equality
- **Line 424**: `ast.rep in domain.relations` — .rep membership

---

## 8. `ivy_compiler.py`

### 8.1 Dict with Symbol keys
- **Line 1028**: `self.domain.aliases[dfn.defines()] = resolve_alias(...)` — Symbol as dict key
- **Line 1746**: `mp = dict((lf.formula.defines(),lf.formula.rhs()) for lf in mod.definitions)` — Symbol as dict key
- **Line 1764**: `dmap = dict((d.formula.defines(),d) for d in mod.definitions)` — Symbol as dict key
- **Line 1772**: `d = dmap[scc[0]]` — Symbol dict lookup

### 8.2 Set with Symbol members
- **Line 1623**: `schemata.add(inst.defines())` — Symbol added to set

### 8.3 Tuples with Symbol for dependency analysis
- **Line 1761**: `arcs = [(d.formula.defines(),x) for d in mod.definitions for x in lu.symbols_ast(...)]` — Symbol tuples

---

## 9. `ivy_module.py`

### 9.1 Dict with Symbol keys
- **Line 199**: `non_epr[ldf.formula.defines()] = (ldf,cnst)` — Symbol as dict key
- **Line 389**: `dfn_map = dict((ldf.formula.defines(),ldf.formula.args[1]) for ...)` — Symbol dict key
- **Line 390**: `sym in dfn_map` — Symbol dict membership

### 9.2 Sort equality
- **Line 206**: `rsort in self.variants[lsort.name]` — Sort in list membership
- **Line 211**: `if sort == rsort:` — Sort == comparison

### 9.3 Set with logic node members
- **Line 334**: `term not in matched` — logic node in set membership
- **Line 342**: `matched.add(term)` — logic node added to set
- **Line 384**: `if dfns in subst:` — Symbol in dict membership
- **Line 391**: `ldf.formula.defines() in rch` — Symbol in set membership

---

## 10. `ivy_isolate.py`

### 10.1 Set with Symbol members
- **Line 1202**: `all_syms.add(x.formula.defines())` — Symbol added to set

---

## 11. `ivy_fragment.py`

### 11.1 universally_quantified_variables — dict with Variable keys
- **Line 348**: `universally_quantified_variables = dict()` — Variable dict
- **Line 353**: `universally_quantified_variables[v] = lf` — Variable as dict key
- **Lines 101, 252, 260, 437**: `if v/fmla in universally_quantified_variables:` — Variable membership

### 11.2 strat_map — defaultdict with Variable/Symbol/(Symbol,int) keys
- **Line 359**: `strat_map = defaultdict(UFNode)` — defaultdict
- **Line 102**: `if fmla not in strat_map:` — Variable as key
- **Line 105**: `strat_map[fmla] = res` — Variable as key
- **Line 115**: `strat_map[il.Symbol('=',fmla.args[0])]` — new Symbol as key (structural equality!)
- **Line 144**: `strat_map[(func,idx)]` — (Symbol, int) tuple as key

### 11.3 macro maps — dict with Variable/Symbol keys
- **Lines 234-239**: `macro_map`, `macro_dep_map`, `macro_var_map`, `macro_value_map` dicts
- **Lines 107, 241, 252-256**: Variable/Symbol as dict keys in membership tests

### 11.4 Set of Variables
- **Line 306**: `fvs = set(il.free_variables(fmla))` — set of Variable objects

---

## 12. `ivy_proof.py`

### 12.1 Sort as dict key / equality check
- **Lines 814-821**: `if x in dmatch and dmatch[x] != y:` / `if x != y:` — Sort as dict key + Sort != comparison
- **Line 903**: `{v.sort:rhsvs[v.name].sort}` — Sort as dict key
- **Line 1243**: `if pat.sort == inst.sort:` — Sort == comparison

### 12.2 Variable as dict key
- **Line 826**: `vvmap = dict((x,y.resort(x.sort)) for x,y in zip(concargs,declargs))` — Variable as dict key
- **Line 825**: `constants = set(x for x in prob.constants if x not in declargs)` — Variable set and membership

### 12.3 Logic node equality in pattern matching
- **Line 1320**: `return dict() if pat == inst else None` — logic node == comparison
- **Line 1341**: `if x == y:` — Lambda == comparison
- **Line 1344**: `return x.body == il.substitute(y.body,...)` — formula body == comparison
- **Line 1330**: `if sym in res:` — Symbol dict membership

### 12.4 Symbol/Variable membership
- **Lines 260, 302**: `sym in self.stale`, `sym in vocab.sorts or sym in vocab.symbols` — Symbol membership
- **Lines 1238-1244**: `if pat in freesyms`, `if pat.sort in freesyms` — logic node and Sort in set

---

## 13. `ivy_mc.py`

### 13.1 Symbol/Formula sets
- **Lines 697, 735-736**: `insts = set()`, `if fmla not in insts:`, `insts.add(fmla)` — Formula in set
- **Line 787**: `defnd = set(dfn.defines() for dfn in trans.defs)` — Symbol set
- **Lines 1169, 1197-1204**: `defsyms = set(...)`, `funs = set()`, `funs.update(...)` — Symbol sets
- **Lines 1278, 1284**: `stvarset = set(stvars)`, `finite_syms_set = set()` — Variable/Symbol sets
- **Lines 1315, 1330, 1395**: `expr not in ...`, `sym in stvarset`, `sym not in def_set` — membership tests
- **Line 1426**: `cnsts = set(sym for ...)` — Symbol set

### 13.2 Symbol as dict key
- **Lines 125, 362**: `dmap = dict((df.defines(),df.args[1]) for df in defs)` — Symbol dict key
- **Lines 808, 1170**: `rn = dict((sym,tr.new(sym)) for sym in ...)` — Symbol dict key
- **Lines 1348, 1369, 1389, 1487, 1538, 1543**: various `dict(...)` with Symbol/Variable keys

---

## 14. `concept.py`

### 14.1 Const frozenset operations
- **Lines 497, 499**: `sig_symbols = frozenset(list(sig.symbols.values()))` — Const frozenset
- **Lines 583-584**: `symbols | frozenset(...)`, `symbols - frozenset(elements)` — Const frozenset union/difference
- **Line 772**: `elements = list(set([uc for ...]))` — Const set

### 14.2 Const as dict key / equality
- **Lines 693, 696**: `if lit.t1 not in nodes:`, `assert lit.t2 in nodes` — Const in dict membership
- **Lines 700, 714**: `for uc, node_name in nodes.items():` — Const dict iteration
- **Lines 703, 717**: `polarity = uc == lit.t2` — Const == comparison
- **Lines 701, 715**: `if uc.sort != lit.t2.sort:` — Sort != comparison

### 14.3 Variable set comparison
- **Lines 37, 89**: `if set(variables) != free_variables(formula):` — Variable set != set

---

## 15. `ivy_art.py`

### 15.1 State objects in sets
- **Lines 414-424**: `covered = set(...)`, `joined = set()`, `covered.add(state)`, `joined.add(s)`, `s not in covered and s not in joined` — State objects in sets

---

## 16. `ivy_interp.py`

### 16.1 Sort as dict key
- **Line 302**: `dict((s,[c.skolem() for c in m.sort_universe(s)]) for s in m.sorts())` — Sort dict key

---

## Summary by Pattern Type

| Pattern | Count | Primary Files |
|---------|-------|---------------|
| Dict with Symbol/Const keys | ~35 | ivy_logic.py, ivy_logic_utils.py, ivy_compiler.py, ivy_module.py, ivy_mc.py, ivy_proof.py |
| Set/frozenset of Var objects | ~20 | logic_util.py, ivy_logic.py, ivy_fragment.py |
| Set of Symbol/Const objects | ~15 | ivy_transrel.py, ivy_mc.py, ivy_compiler.py, ivy_module.py, concept.py |
| Sort as dict key or == comparison | ~10 | ivy_logic.py, ivy_module.py, ivy_proof.py, ivy_solver.py, concept.py |
| Logic node == comparison | ~8 | logic_util.py, ivy_logic.py, ivy_proof.py, concept.py |
| Formula in set | ~5 | ivy_mc.py, ivy_fragment.py |
| .rep (string) as dict key | ~15 | ivy_solver.py, ivy_actions.py (these use string keys, OK) |

### Key Observation

Many Python patterns use `.rep` (which is a string/name for simple types) as dict keys.
The dangerous patterns are where
the **object itself** (not its `.rep` or `.name`) is used as a dict key or set member.

Many errors we found in the Go port were due to not conforming to the Python
structual equivalence semantics. Audit carefully where any string map key
is derived from, and check to see if it could collide and needs to use 
the Sexp() method for the NodeKey field.
