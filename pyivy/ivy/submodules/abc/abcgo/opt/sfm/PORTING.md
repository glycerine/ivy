# SFM port status

The package builds and its Go tests run as a whole. The mapped-network entry
point is `Sfm_OptimizeMapped`; ABC network conversion and the ABC-facing entry
point are in `base/abc/sfm_mapped_adapter.go`. The original C is retained in
`src/opt/sfm/sfmDec.c` for comparison.

The mapped driver has working area and block-delay paths, but package build
success is not a claim of complete ABC behavioral parity. Before treating the
port as complete, compare these remaining areas with the C implementation:

- The C driver's statistics, verbose reports and the externally exposed
  decomposition helper APIs have not all been translated.
- The C simulation patterns persist between nodes; the native driver currently
  initializes window-local simulation patterns. SAT validation still guards
  decomposition, but search choices and results may differ.
- The upstream alternate timing manager is itself a placeholder. Integration
  with ABC's external timing constraints is not yet implemented here.
- Multi-candidate search, fallback ordering and timing-driven optimization need
  broader differential testing against ABC on real mapped designs.
- The ABC adapter currently rejects complemented mapped edges and maps the
  network through a temporary representation; metadata and external timing
  interactions require end-to-end review.

The overall project remains incomplete. GIA currently fails to parse its
unported `giaMini.go`; full-project builds and both requested final review
passes are still outstanding.
