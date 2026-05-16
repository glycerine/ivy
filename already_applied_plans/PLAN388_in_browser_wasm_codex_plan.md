Yes. Here’s the plan without a service worker.

**Core Shape**
The browser architecture should be:

```text
Main UI thread
  owns UIDataModel, source text, selected isolate, personality, sheet state
  owns persistence
  owns job control UI
  launches/cancels engine jobs

Dedicated engine worker
  runs Go webui wasm + Z3 wasm
  treats all state as derived/cache
  receives complete job input snapshots
  returns typed model snapshots/deltas
  can be hard-killed on cancel

IvyApiAdapter
  hides whether backend is hosted Go or browser wasm
```

The rule: **the worker may cache, but it may not be the source of truth.**

**Phase 1: Define The Contract**
Audit every `IvyApiAdapter` method and classify it:

- Pure query: returns data, should not mutate model.
- Model mutation: must return a typed `UIDataModel` snapshot or patch.
- Long-running job: must have a job id, progress events, cancellation, and atomic final result.

For each call, define the full input snapshot it needs:

```ts
{
  requestId,
  modelRevision,
  sourceText,
  fileName,
  selectedIsolate,
  personality,        // CTI or reachability
  analysisMode,       // induction/pdr/etc when relevant
  uiDataSnapshot,     // only the model-owned state needed by the command
  command
}
```

No replay log. No “backend remembers what happened.” If the worker dies, the next job starts from a fresh snapshot.

**Phase 2: Make Job Results Atomic**
Long-running jobs should not partially mutate visible UI state while running.

They may emit:

- progress text
- phase changes
- logs
- elapsed time
- solver/status updates

But graph/model changes should commit only when the job succeeds and returns a final typed result. If canceled, the UI model remains as it was before the job, except for the job record saying `cancelled`.

This matters because hard worker cancellation is clean only if there is no half-applied backend state.

**Phase 3: Browser Wasm Adapter**
Add a real browser implementation beside the hosted one:

- `HostedGoIvyApiAdapter`: current remote/server path.
- `BrowserWasmIvyApiAdapter`: new worker-backed path.

The existing “run in browser / run remote” setting should switch the adapter, not just the label.

The browser adapter owns:

- worker creation
- wasm boot
- request/response ids
- progress event routing
- worker termination on cancel
- worker recreation after cancel/crash

The worker should live under `webui/`, not `svk/`. We can copy useful bridge ideas from `webvue/`, but `svk/` should not become production dependency.

**Phase 4: Worker Protocol**
Use a small message protocol:

```ts
init
ready
request
progress
result
error
cancelled
```

For browser cancellation, the main thread does not politely ask Z3 to stop. It terminates the whole engine worker, marks the job canceled, and starts a fresh idle worker.

For hosted Go cancellation, use the same job-control UI, but route cancellation through the existing server context/Z3 interrupt path.

**Phase 5: Model Sync Discipline**
Before launching a job:

1. Flush pending UI changes into `UIDataModel`.
2. Increment or capture `modelRevision`.
3. Build the job snapshot from the model.
4. Persist the current source/model state locally.
5. Submit the snapshot to the selected adapter.

When a result comes back:

1. Check `baseRevision`.
2. Reject stale results.
3. Apply only through model/store methods.
4. Let render services derive Cytoscape/DOM from the model.

Initial implementation should prefer returning full typed model snapshots for graph-changing operations. Patches can come later after the contract is proven.

**Phase 6: Job Control UI**
The job-control page should be model-driven too:

- job id
- backend: `remote` or `browser`
- command: induction, PDR, bounded check, CTI action, etc.
- selected isolate
- started time
- phase
- latest progress message
- cancel button
- final status

Cancel behavior:

- Browser backend: terminate worker.
- Remote backend: abort request and signal backend cancellation.
- UI: mark job `cancelling`, then `cancelled`.

**Phase 7: First Supported Browser Jobs**
Do this incrementally.

1. Boot browser worker and run a harmless capability request.
2. Load/compile source snapshot in worker.
3. Run a simple check in browser mode.
4. Add hard cancel for a simulated long job.
5. Add real hard cancel for induction/PDR/BMC jobs.
6. Add CTI-specific actions only once their state contract is explicit.

Do not start with the most stateful CTI mutation. Start with checks where final result can be atomic.

**Phase 8: Tests**
Add tests at the boundary where bugs would hurt us:

- browser adapter sends complete snapshots, not hidden incremental state
- stale worker result is ignored
- canceled worker result cannot mutate `UIDataModel`
- after hard cancel, a fresh worker can run the next job from frontend state
- job-control UI shows browser/remote jobs consistently
- graph updates still flow model-first
- selected isolate/personality/mode are included in job inputs
- no direct Cytoscape/DOM mutation from worker results

Then add one end-to-end browser-mode test:

1. Load spec.
2. Select isolate.
3. Start long check.
4. Cancel it.
5. Confirm graph/model unchanged.
6. Start a second check.
7. Confirm result applies normally.

**Acceptance Criteria**
This is done when:

- “Run in browser” actually uses the wasm worker.
- Cancel kills browser jobs reliably.
- Cancel does not corrupt UI/model state.
- The frontend can recover without replaying logs.
- Worker state is treated as disposable cache.
- All visual changes still derive from `UIDataModel`.

That gives us local execution without pretending the browser backend can safely own hidden durable state.

