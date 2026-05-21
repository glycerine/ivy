# Fix: Add matching xtracer.trace() calls to Python for RankingL2STactic compileInvar loop
Created: 2026-05-21 UTC

## Context

Golden test diverges at trace line 228596. Go emits `ranking.RankingL2STactic compileInvar[0] pre-compile HASH` but Python emits `proof.GoalVocab ENTER label=sys_live.live`. Go's traces are the conformance spec — Python must be brought into alignment.

**Root cause:** Python's `ivy_ranking.py:129` compiles tactic invariants with a silent list comprehension:
```python
invars = [ilg.label_temporal(ipr.compile_with_goal_vocab(inv,goal),proof_label) for inv in tactic_invars]
```
No trace calls around the loop. Go has three (pre-compile, compiled=nil, post-compile) that Python is missing.

Go's loop in `check_ranking.go:285-296` emits:
1. `ranking.RankingL2STactic compileInvar[%d] pre-compile HASH`
2. (GoalVocab ENTER/EXIT from inside GoalVocab call — already emitted by both sides)
3. `ranking.RankingL2STactic compileInvar[%d] compiled=nil` (if nil)
4. `ranking.RankingL2STactic compileInvar[%d] post-compile HASH`

## Sibling audit (per memory rule "Port at every mirror site")

`check_l2s.go:438-449` also has a `compiled=nil` trace (line 443). Python's `ivy_l2s.py:214-219` does NOT have this trace. Add it there too.

## Changes

### 1. `~/ivy/pyivy/ivy/ivy/ivy_ranking.py` — primary fix

Replace line 129 list comprehension with a for loop that matches Go's trace structure:

**Before (line 129):**
```python
    invars = [ilg.label_temporal(ipr.compile_with_goal_vocab(inv,goal),proof_label) for inv in tactic_invars]
```

**After:**
```python
    invars = []
    for idx, inv in enumerate(tactic_invars):
        if __debug__: xtracer.trace("ranking.RankingL2STactic compileInvar[%d] pre-compile HASH canon=%s" % (idx, inv.canon()))
        compiled = ipr.compile_with_goal_vocab(inv, goal)
        if compiled is None:
            if __debug__: xtracer.trace("ranking.RankingL2STactic compileInvar[%d] compiled=nil" % idx)
            continue
        labeled = ilg.label_temporal(compiled, proof_label)
        if __debug__: xtracer.trace("ranking.RankingL2STactic compileInvar[%d] post-compile HASH canon=%s" % (idx, labeled.canon()))
        invars.append(labeled)
```

### 2. `~/ivy/pyivy/ivy/ivy/ivy_l2s.py` — sibling nil-trace fix

Add the compiled=nil trace to match Go's `check_l2s.go:443`.

**Before (ivy_l2s.py:214-219):**
```python
    for idx, inv in enumerate(tactic_invars):
        if __debug__: xtracer.trace("l2s.l2sTacticInt compileInvar[%d] pre-compile HASH canon=%s" % (idx, inv.canon()))
        compiled = ipr.compile_with_goal_vocab(inv,goal)
        labeled = ilg.label_temporal(compiled,proof_label)
        if __debug__: xtracer.trace("l2s.l2sTacticInt compileInvar[%d] post-compile HASH canon=%s" % (idx, labeled.canon()))
        invars.append(labeled)
```

**After:**
```python
    for idx, inv in enumerate(tactic_invars):
        if __debug__: xtracer.trace("l2s.l2sTacticInt compileInvar[%d] pre-compile HASH canon=%s" % (idx, inv.canon()))
        compiled = ipr.compile_with_goal_vocab(inv,goal)
        if compiled is None:
            if __debug__: xtracer.trace("l2s.l2sTacticInt compileInvar[%d] compiled=nil" % idx)
            continue
        labeled = ilg.label_temporal(compiled,proof_label)
        if __debug__: xtracer.trace("l2s.l2sTacticInt compileInvar[%d] post-compile HASH canon=%s" % (idx, labeled.canon()))
        invars.append(labeled)
```

## Critical files

- `~/ivy/pyivy/ivy/ivy/ivy_ranking.py` line 129
- `~/ivy/pyivy/ivy/ivy/ivy_l2s.py` lines 214-219

## Verification

```
cd ~/ivy/goivy && make golden-all
```

Expect the divergence at trace 228596 to disappear. Test should either pass fully or diverge at a higher trace number.
