
Great! Now I understand everything. Let me create a comprehensive summary for the user. Let me
put together the final report now:

Comprehensive Analysis: Python Ivy L2S Pipeline and LabeledFormula Cloning

Based on my thorough examination of the Python Ivy codebase, here's what I found:

1. mod_pass Function (Lines 759-770 in ivy_l2s.py)

The Python equivalent of the Go modPass function is defined at line 759 in ivy_l2s.py:

def mod_pass(transform):
    if __debug__:
        nPropPrems = sum(1 for p in prems if ipr.goal_is_property(p))
        xtracer.trace("l2s.modPass ENTER nInvars=%d nAsms=%d nBindings=%d nPrems=%d
nPropPrems=%d" %
            (len(model.invars), len(model.asms), len(model.bindings), len(prems), nPropPrems))

    model.invars = [transform(x) for x in model.invars]
    model.asms = [transform(x) for x in model.asms]
    # TODO: what about axioms and properties?
    model.bindings = [b.clone([transform(b.action)]) for b in model.bindings]
    model.init = transform(model.init)
    list_transform(prems,transform)

What it iterates over and clones:
- invars: All invariants (LabeledFormulas) - transformed in place
- asms: All assumptions (LabeledFormulas) - transformed in place
- bindings: All action bindings - CLONED with b.clone([transform(b.action)])
- init: The initialization action - transformed in place
- prems: Premises from the proof goal - transformed in place

Key observation: Each binding is CLONED once per mod_pass call because of line 768.

2. The L2S Pipeline Steps and Cloning Order (Lines 744-1310 in ivy_l2s.py)

The step-by-step flow that clones LabeledFormulas:

Step 1 (Line 744): Add tactic invariants to model
model.invars = model.invars + invars

Step 2 (Line 795): FIRST mod_pass - Replace temporals with named binders
mod_pass(replace_temporals_by_l2s_g)
This transforms:
- All invars (clones each node via ast.clone())
- All asms
- All bindings (CLONES binding objects)
- init
- All prems

Step 3 (Line 808): SECOND mod_pass - Normalize named binders
mod_pass(ilu.normalize_named_binders)
This transforms AGAIN:
- All invars
- All asms
- All bindings (CLONES again)
- init
- All prems

Step 4 (Lines 1153-1210): Instrument all actions with events
model.bindings = [b.clone([b.action.clone([instr_stmt(b.action.stmt,b.action.labels)])])
                  for b in model.bindings]
This CLONES each binding a third time with instrumented statements. Additionally:
- CallActions may be split via split_returns() if they modify monitored symbols (line 1157)
- Events are added before and after statements for temporal properties

Step 5 (Line 1260-1261): Add idle action
model.bindings.append(itm.ActionTermBinding('idle',itm.ActionTerm([],[],[],idle_action)))
model.calls.append('idle')

Step 6 (Line 1310): THIRD mod_pass - Replace named binders with fresh constants (BULK
SUBSTITUTION)
mod_pass(lambda ast: ilu.replace_named_binders_ast(ast, subs))
This is the massive cloning operation using the subs dictionary built at lines 1300-1304:
subs = dict(
    (b, lg.Const('{}_{}'.format(k, i), b.sort))
    for k, v in named_binders.items()
    for i, b in enumerate(v)
)

This CLONES:
- All invars (with all named binders substituted)
- All asms (with all named binders substituted)
- All bindings (CLONES again with all named binders substituted in actions)
- init (with all named binders substituted)
- All prems (with all named binders substituted)

3. NormalProgram in Python (Lines 171-209 in ivy_temporal.py)

class NormalProgram(ia.AST):
    def __init__(self,bindings,init,invars,asms,calls):
        self.bindings,self.init,self.invars,self.asms,self.calls =
bindings,init,invars,asms,calls

    @property
    def fmlas(self):
        res = self.bindings + [self.init] + self.invars + self.asms
        if hasattr(self,'postconds'):
            for x in self.postconds.values():
                res.extend(x)
        return res

Fields:
- bindings: List of ActionTermBinding objects (each contains an action)
- init: Initialization action
- invars: List of invariant LabeledFormulas
- asms: List of assumption LabeledFormulas
- calls: List of action name strings (public actions)
- postconds (optional): Dictionary mapping action names to lists of postcondition
LabeledFormulas

Important: The fmlas property returns bindings + [init] + invars + asms + postconds.values().
However, postconds are NOT processed by mod_pass (see line 766 TODO comment), so they may not
get all the transformations that invars and asms get.

4. split_returns for CallActions (Lines 1435-1443 in ivy_actions.py)

def split_returns(self):
    actual_returns = self.args[1:]
    rn = iu.UniqueRenamer('',used_symbols_ast(self))
    new_returns = [x.rename(rn) for x in actual_returns]
    asgn = Sequence(*([self.clone([self.args[0]] + new_returns)]
                      +[AssignAction(x,y).sln(self.lineno)
                        for x,y in zip(actual_returns,new_returns)]))
    res =
LocalAction(*(new_returns+[asgn]),caller="actions.CallAction.action_update").sln(self.lineno)
    return res

When it's called (Line 1157 in ivy_l2s.py):
if (isinstance(stmt,CallAction)):
    actual_returns = stmt.args[1:]
    if any(sym in symprops or sym in symwhens or sym in symwaits for sym in actual_returns):
        return instr_stmt(stmt.split_returns(),labels)

This splits CallActions that modify monitored symbols (those in temporal properties),
converting the call and returns into separate statements.

5. Key Cloning Operations in the L2S Flow

Python clones formulas at multiple points:

┌──────┬────────────────────────────────────┬───────┬─────────────────────────────────────┐
│ Step │             Operation              │ Lines │          What Gets Cloned           │
├──────┼────────────────────────────────────┼───────┼─────────────────────────────────────┤
│ 1    │ Add tactic invars                  │ 744   │ Appends invars to invars list       │
├──────┼────────────────────────────────────┼───────┼─────────────────────────────────────┤
│ 2    │ First mod_pass (temporals→binders) │ 795   │ invars, asms, bindings, init, prems │
├──────┼────────────────────────────────────┼───────┼─────────────────────────────────────┤
│ 3    │ Second mod_pass (normalize)        │ 808   │ invars, asms, bindings, init, prems │
├──────┼────────────────────────────────────┼───────┼─────────────────────────────────────┤
│ 4    │ Instrument actions                 │ 1210  │ bindings (with instrumented stmts)  │
├──────┼────────────────────────────────────┼───────┼─────────────────────────────────────┤
│ 5    │ Third mod_pass (replace binders)   │ 1310  │ invars, asms, bindings, init, prems │
└──────┴────────────────────────────────────┴───────┴─────────────────────────────────────┘

Why Python clones MORE than Go:

The critical insight is that Python applies mod_pass THREE times (lines 795, 808, 1310), and
each time it CLONES bindings via b.clone([transform(b.action)]). Additionally:

1. substitute_ast and ast_rewrite equivalents: The Python code uses:
  - ilu.replace_temporals_by_named_binder_g_ast() (line 794) - recursively clones AST nodes
  - ilu.normalize_named_binders() (line 808) - recursively clones AST nodes via ast.clone(args)
 (line 277)
  - ilu.replace_named_binders_ast() (line 1310) - recursively clones ALL AST nodes via
ast.clone(args) (line 357)
2. The third mod_pass substitution (line 1310) is particularly expensive because it:
  - Builds a massive substitution dict subs with named binder→constant mappings (lines
1300-1304)
  - Applies this to EVERY formula: all invars, all asms, all bindings' actions, init, and all prems
  - Each application clones the entire AST tree with replace_named_binders_ast, which
recursively calls ast.clone(args) at line 357
3. Bindings are cloned multiple times:
  - After first mod_pass (line 795): Each binding cloned once
  - After second mod_pass (line 808): Each binding cloned again
  - After instrumentation (line 1210): Each binding cloned again with instrumented statements
  - After third mod_pass (line 1310): Each binding cloned again with all named binders replaced

6. The PRESERVE Clone Clue

The massive cloning at line 1310 (mod_pass(lambda ast: ilu.replace_named_binders_ast(ast,
subs))) happens after all the temporal conversion and normalization is complete. This final
pass replaces ALL named binders with fresh constant symbols at a FIXED point in the code, which
 would explain why Go saw a divergence at counter=2306 (all the PRESERVE clones from this
single mod_pass operation).

Summary: Python's l2s_tactic_int function clones LabeledFormulas much more aggressively than
Go's version because:
1. It calls mod_pass THREE times instead of possibly fewer in Go
2. Each mod_pass call clones bindings with b.clone()
3. The transformations themselves (especially replace_named_binders_ast) recursively clone every node in the AST
4. Postconds exist but are NOT included in mod_pass transformations, suggesting incomplete alignment between the two implementations
