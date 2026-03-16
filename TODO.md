# TODO: Python Backend Oracle for ivyweb

## Status

Steps 1–7 of the plan are implemented. The Go Backend interface, GoBackend,
PyBackend, ConformBackend, refactored server/handlers, CLI flags, and the
Python sidecar are all in place and tested. What remains is hardening,
conformance tuning, and the Z3 fork alignment.

---

## 1. Python sidecar: wire remaining endpoints

The sidecar (`~/pyivy/ivy/ivy/sidecar.py`) handles session/new, load,
concept, arg, check, toggles, and health. The following endpoints are
stubs that return `{"status":"ok"}` without doing real work:

- **POST /session/{id}/action** — `_execute_action` ignores the action
  name and args. Must dispatch to the Ivy AnalysisGraph methods:
  `undo`, `redo`, `recalculate`, `gather`, `conjecture`, `remember`,
  `splatter`, `get_conjectures`, `add_relation`, etc.
  Reference: `webui/session.go:ExecuteAction` (Go) and
  `ivy_ui_cti.py` / `ivy_graph.py` (Python).

- **POST /session/{id}/concept/split** — must call
  `ConceptInteractiveSession.split()` or equivalent.
  Same for `concept/empty`, `concept/remove`, `concept/undo`,
  `concept/materialize`, `concept/reset`, `concept/diagram`,
  `concept/projection`. The Python `concept_interactive_session.py`
  has the implementations; the sidecar just needs to call them
  and return `{"status":"ok"}` on success.

- **POST /session/{id}/arg/action** — must dispatch ARG node actions
  (view_state, check_safety, extend, decompose, etc.) to the
  AnalysisGraph. Reference: `webui/session.go:ArgNodeAction`.

- **POST /session/{id}/proof/action** — must dispatch proof goal
  actions. Reference: `webui/session.go:ProofGoalAction`.

- **GET /session/{id}/arg** — currently returns `{"elements":[]}`.
  After `_load_content`, `sess.ag` has states and transitions but
  `_get_arg` uses `id(s)` for node identity which doesn't match
  Go's integer-ID scheme. Must produce Cytoscape JSON matching
  Go's `RenderARG` output exactly (same node IDs, labels, classes).

- **GET /session/{id}/events** — SSE endpoint exists but no
  operations emit events yet. Each operation that modifies state
  should push events to `sess.events` matching the Go event types
  (`file_loaded`, `check_started`, `check_completed`,
  `action_started`, `action_completed`, `concept_updated`, etc.).

---

## 2. Canonical JSON conformance

For `-conform` mode to work, Go and Python must produce **byte-identical**
JSON for the same logical data. Current gaps:

- **Go uses `json.Marshal`** (compact, sorted map keys). The sidecar
  uses `json.dumps(sort_keys=True, separators=(",",":"))`. These
  should match for simple cases, but edge cases need testing:
  - `null` vs absent keys (Go omits empty strings with `omitempty`;
    Python includes them as `""`)
  - Empty arrays: Go may emit `null` for nil slices; Python emits `[]`
  - Number formatting: Go `json.Marshal` on float64 may differ from
    Python `json.dumps` (e.g., `1.0` vs `1`)
  - Unicode escaping differences

- **Array ordering**: Both backends must sort arrays of relations,
  edges, nodes, node_labels identically. Go's `GoBackend.GetConcept`
  already calls `sort.Strings`; Python's `_get_concept` also sorts.
  But other endpoints (action results, check results) may have
  unsorted arrays.

- **Error messages**: Go error messages must be updated to match
  Python's exact strings. Key messages to align:
  - Check result messages: `"The following conjecture is not relatively inductive:"`
  - `"Inductive invariant found:\n..."` — the conjecture text
    formatting (Go uses `fmt.Sprint(lc.Formula)`, Python uses
    `str(conj)`) must produce identical strings.
  - `"No conjectures to check"`, `"No module loaded"`, etc.

- **Concept graph elements**: The CyElements structure differs:
  - Go includes `shape` in node data; Python's `CyElements.add_node`
    also includes `shape`. But Go also adds `actions` as a JSON
    array of `{label, action}` objects; Python has callbacks (stripped).
  - Go omits `cluster`, `events`, `locked` from some elements;
    Python always includes them.
  - These structural differences must be harmonized field by field.

- **Check result structure**: Go returns
  `{"status":"ok","result":"pass","mode":"induction","message":"...",
  "failed_conjecture":"","failed_label":"","used_relations":null}`.
  Python must match exactly — including empty strings vs null,
  and the `used_relations` being `null` (not `[]`) when not set.

---

## 3. Go port must use Ivy's bundled Z3 4.7.1 fork

### Background

Python Ivy ships its own Z3 fork (v4.7.1) from `ruijiefang/z3-ivy`
(Ken McMillan's fork with Ivy-specific patches). The submodule is at
`~/pyivy/ivy/submodules/z3/`. The built `libz3.dylib` lives at
`~/pyivy/ivy/build/lib.macosx-14.0-x86_64-cpython-310/ivy/z3/libz3.dylib`
and is symlinked to `~/pyivy/ivy/ivy/z3/libz3.dylib`.

The Go port (`goivy/solver/`) currently links against whatever system
Z3 is available. For conformance, the Go solver must link against the
**same Z3 4.7.1 fork** so that:

- SAT/UNSAT results are identical for the same input
- Counterexample models (when SAT) have the same structure
- Sort inference and quantifier instantiation behave identically

### Tasks

1. **Build the Z3 fork as a C library for Go CGO linking.**
   The submodule at `~/pyivy/ivy/submodules/z3/` can be built with:
   ```
   cd ~/pyivy/ivy/submodules/z3
   python scripts/mk_make.py --staticlib
   cd build && make -j4
   ```
   This produces `libz3.a` (static) or `libz3.dylib` (shared).

2. **Update `goivy/solver/` CGO directives** to point to the fork's
   headers and library:
   ```go
   // #cgo CFLAGS: -I${PYIVY_ROOT}/submodules/z3/src/api
   // #cgo LDFLAGS: -L${PYIVY_ROOT}/submodules/z3/build -lz3
   ```
   Or use the pre-built dylib from the ivy build output.

3. **Verify Go Z3 bindings compatibility.** The Go solver's
   `z3convert.go` uses Z3 C API calls. Ensure all referenced
   symbols exist in Z3 4.7.1. The Python bindings at
   `z3core.py` line 740 reference `Z3_get_parser_error` which
   was missing from a newer Z3 — the fork's 4.7.1 API may
   lack symbols the Go code expects from a newer Z3.
   Audit `goivy/solver/*.go` for all `C.Z3_*` calls and
   cross-reference with the fork's `z3_api.h`.

4. **Pin the Z3 version in the Go build.** Add a `Makefile` or
   build script that:
   - Checks out the correct Z3 submodule commit
   - Builds `libz3.a`
   - Sets CGO flags for the Go build

5. **Test conformance of Z3 results.** Load the same `.ivy` file
   in both backends, run `check` with mode `induction`, and verify
   the check results (pass/fail, counterexample formulas) are
   byte-identical.

---

## 4. Pre-existing Python Ivy bugs to track

Two bugs were found and worked around in the sidecar:

1. **`ivy_compiler.py` line 2239 variable shadowing:** The `for iso in
   list(im.module.isolates.values()):` loop reassigns the local `iso`
   to an `IsolateDef`, shadowing `import ivy_isolate as iso` on line 23.
   Line 2252 then calls `iso.create_isolate()` on the wrong object.
   **Workaround:** The sidecar calls `ivy_load_file(sio, create_isolate=False)`
   and then calls `ivy_isolate.create_isolate()` directly.
   **Proper fix:** Rename the loop variable in `ivy_compiler.py` from
   `iso` to `isol` or similar.

2. **`concept.py` line 76 `ConceptDict.add`:** References `self.category`
   instead of `self[category]`.
   **Workaround:** The sidecar builds concept data directly from
   `im.module.sig` instead of calling `get_initial_concept_domain()`.
   **Proper fix:** Change line 76 from `self.category.append(name)`
   to `self[category].append(name)`.

---

## 5. Multi-session support in the sidecar

Python Ivy uses extensive module-level globals (`ivy_logic.sig`,
`ivy_module.module`, etc.). The sidecar currently uses a threading lock
(`_ivy_lock`) so only one session's compilation runs at a time. This means:

- Only one session can be loading/compiling at a time.
- After compilation, `im.module.sig` reflects the **last loaded** session.
  A second session's `_get_concept` would return the wrong data.

### Options

1. **Single-session mode** (current): Adequate for A/B testing where
   only one `.ivy` file is loaded at a time. Document the limitation.

2. **Per-session subprocess**: Fork a new Python process per session.
   Each process has its own global state. The sidecar becomes a
   dispatcher that routes requests to the correct subprocess.
   Use `multiprocessing` with JSON IPC.

3. **Save/restore globals**: Before each operation, restore the
   session's saved `il.sig`, `im.module`, etc. After the operation,
   save them back. Fragile but avoids the subprocess overhead.

For now, single-session mode is sufficient. Document it in the
`-conform` help text.

---

## 6. Integration tests

- **`webui/backend_test.go`**: Add tests that exercise `GoBackend`
  directly (not through HTTP). Cover `NewSession`, `Load`, `GetConcept`,
  `Check` with a real `.ivy` file.

- **`webui/backend_py_test.go`**: Integration test that starts the
  sidecar, creates a session, loads a file, gets concept data, and
  runs a check. Skip if Python/Z3 not available.

- **`webui/backend_conform_test.go`**: Load the same `.ivy` file
  through `ConformBackend`. Verify no conformance errors for the
  basic flow (new session → load → concept → check). This is the
  key test — it will fail until all JSON differences are resolved.

- **End-to-end browser test**: Extend `browser_test.go` to test
  with `-py` backend (requires sidecar to be running).

---

## 7. ConformBackend: SSE event conformance

The current `ConformBackend.Events()` returns only the Go event
channel and does not compare events from both backends. To fully
verify conformance:

- Subscribe to both Go and Python event streams.
- For each operation (load, check, action), collect the events
  emitted by both backends.
- After the operation completes, compare the event sequences.
- Log mismatches; optionally surface them to the user.

This is lower priority than response conformance since events are
informational and don't affect correctness of displayed data.

---

## 8. Documentation and deployment

- **README update**: Document the `-py` and `-conform` flags,
  prerequisites (Python 3.10+, Ivy installed, Z3 built), and
  the `PYIVY_ROOT` environment variable.

- **Build script**: Add a `Makefile` target or script that:
  1. Builds the Z3 fork from the submodule
  2. Symlinks `ivy/z3` if needed
  3. Builds `ivyweb` with CGO flags for the Z3 fork

- **CI**: Add a CI job that runs conformance tests with both
  backends. Requires Python + Ivy + Z3 in the CI environment.
