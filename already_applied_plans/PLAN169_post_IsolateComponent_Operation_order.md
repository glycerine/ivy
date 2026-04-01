# Fix Post-IsolateComponent Operation Order to Match Python

**Created**: 2026-04-01

## Context

After `isolate_component` / `IsolateComponent` returns, `create_isolate` / `CreateIsolate` performs a series of post-processing steps. Go has these in a different order than Python, which causes trace divergences because some operations (like `fix_initializers` → `loop_action`) trigger LF.clone and CallAction/LocalAction traces at different relative positions.

## Current Orders

**Python** (ivy_isolate.py lines 1899–1954):
1. Label public actions (1899–1900)
2. Create ext action (1901–1907)
3. `slv.check_compat()` (1911)
4. `mod.update_conjs()` (1915)
5. Cone of influence / pedantic warnings (1919–1936)
6. `fix_initializers(mod, after_inits)` (1939)
7. `mod.canonize_types()` (1941)
8. Bracket actions (1944–1946)
9. Set isolate_proof (1948)
10. EXIT (1954)

**Go** (isolate/create.go, current):
1. CreateImports (Go-only, lines 200–266)
2. `FixInitializers` (268–271)
3. `CanonizeTypes` (273–277)
4. Bracket actions (279–289)
5. `CheckCompatStatic` (291–300)
6. `UpdateConjs` (302–305)
7. Label public actions (307–322)
8. Create ext action (324–360)
9. Cone of influence (362–373)
10. Set isolate_proof (375–380)
11. EXIT (382)

## Desired Go Order (matching Python)

1. CreateImports (Go-only, stays first — no Python equivalent at this point)
2. **Label public actions** (moved from position 7)
3. **Create ext action** (moved from position 8)
4. **check_compat** (moved from position 5)
5. **update_conjs** (moved from position 6)
6. **Cone of influence** (moved from position 9)
7. **fix_initializers** (moved from position 2)
8. **canonize_types** (moved from position 3)
9. **bracket actions** (moved from position 4)
10. Set isolate_proof (stays same)
11. EXIT (stays same)

## Change

**File**: `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/isolate/create.go`

Reorder the code blocks between line 197 (`after_isolate_component` trace) and line 382 (EXIT trace). Each block is self-contained with its own xtracer calls and can be moved as a unit. No logic changes — pure reordering.

The blocks to rearrange (identified by their xtracer landmarks):

| Block | Current lines | Move to position |
|-------|--------------|-----------------|
| CreateImports | 199–266 | 1 (stays) |
| Label public | 307–322 | 2 |
| Create ext action | 324–360 | 3 |
| check_compat | 291–300 | 4 |
| update_conjs | 302–305 | 5 |
| Cone of influence | 362–373 | 6 |
| fix_initializers | 268–271 | 7 |
| canonize_types | 273–277 | 8 |
| bracket actions | 279–289 | 9 |
| isolate_proof | 375–380 | 10 (stays) |
| EXIT | 382 | 11 (stays) |

## Verification

```bash
cd ~/ivy/goivy && go build ./... && make golden
```

Check that the trace divergence point advances past line 156456.
