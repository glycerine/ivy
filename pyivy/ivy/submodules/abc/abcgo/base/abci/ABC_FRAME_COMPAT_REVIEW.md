# ABC frame compatibility review

Compared the fields used by `src/base/main/mainInt.h` with `base/main.Frame`, its `ABCFrameState`, and the frame selectors in `abc.go`.

- `pGia` and its saved variants live in the Frame resource map. The update/get bridge preserves the source transfer order: compatible CI/CO name vectors move to the replacement manager, the previous current manager moves to `pGia2`, and taking the current manager clears the owned slot.
- `pNtkCur` uses `Frame.Network` and the existing current-network replacement API, preserving the main package's backup and close behavior. Other retained networks and C-only libraries use typed resource entries.
- `pCex`, multi-output CEX vectors, output equivalences, statuses, and verification status/frame count use the existing opaque `ABCFrameState` replacement/get APIs. Replacing CEX pointers keeps pointer-to-pointer transfer behavior.
- `vIndFlops`, object IDs, and best-mapping counters use resource-backed values/slots so their ownership and mutable address semantics are retained without zero-value shadows.
- `fBatchMode`/`fBridgeMode` map to the existing frame flags. C-only command/lifecycle names delegate to their owning Go packages.
- The liveness command adapters convert the current ABC network to the graph type used by `proof/live`, preserve CI/CO names, then convert a changed graph back through `Abc_NtkFromAigPhase` and replace the frame network.

The frame tests cover GIA name transfer, backup/take ownership, CEX pointer storage, and verification metadata. Package execution is pending unrelated `abcMiter.go` parser errors.
