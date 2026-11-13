# LLB source fidelity gaps

The package builds and its tests pass, but that does not establish that the
original ABC algorithms have been ported. The following source-named entry
points still need source-faithful implementations and regression coverage:

- `Llb_ManComputeImage` in `llb1Reach.go` builds a monolithic transition
  relation and ignores its backward flag. The C implementation in
  `src/bdd/llb/llb1Reach.c` walks the partition matrix, selecting reverse
  partition order and backward quantification when requested. The surrounding
  reachability driver must also be reviewed against that source.
- `Llb_ManFlow`, `Llb_ManFlowCompute`, and `Llb_ManComputeCuts` in
  `llb2Flow.go` use placeholder cuts or a zero flow instead of the flow and
  cut algorithms in `src/bdd/llb/llb2Flow.c`. Their downstream callers and
  cut-map side effects require further review.
- `Llb_CoreExperiment` in `llb2Core.go` invokes the native reachability path
  instead of the flow-partitioned driver in `src/bdd/llb/llb2Core.c`; its
  working graph, cuts, and deadline are not used.
- The sweep routines in `llb4Sweep.go` include shortcut implementations that
  need comparison with `src/bdd/llb/llb4Sweep.c`. In particular, sweep ordering
  ignores `saveAll`, cutpoint selection ignores its limit, and sweep
  partitioning bypasses the source sweep traversal.
- `Llb_NonlinTrySubsetting` in `llb3Nonlin.go` is empty; review it and the
  remaining nonlinear drivers against `src/bdd/llb/llb3Nonlin.c` and
  `src/bdd/llb/llb4Nonlin.c`.

This is an initial inventory, not an exhaustive audit. In particular, passing
unit tests for the native reachability approach do not validate its behavior
against the C partitioned algorithms and options.
