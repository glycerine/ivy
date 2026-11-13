# INT2 helper test inventory

This inventory maps the Go helpers to tests; it is not a claim that the
incomplete C interpolation driver works. See [SOURCE_FIDELITY.md](SOURCE_FIDELITY.md)
for undeclared C variables, inferred helpers, and retained source defects.

| Helper | Focused coverage |
| --- | --- |
| `Int2_ManSetDefaultParams` | `TestSetDefaultParamsResetsAllFields`, `TestSetDefaultParamsRejectsNil` |
| `Int2_ManCreate`, `Int2_ManStop` | Manager setup and teardown in the frame-construction tests; the C constructor contains undeclared `pFrames`. |
| `Int2_ManDupInit` | `TestDupInitMatchesMuxedStateTransitions`, `TestDupInitWithNoOriginalPrimaryInputs` |
| `Int2_ManCheckInit` | `TestCheckInitConstantTransitions`, `TestCheckInitAgainstExhaustiveTransitions`, `TestCheckInitIgnoresPropertyOutput` |
| `Int2_ManFrameInit` | `TestFrameUnrollingTracksLatestInput`, `TestFrameInitMatchesSequentialSimulation`, `TestSourceCopiesNativeXorAsAnd`, `TestZeroFrameBuildersReturnInitialState` |
| `Int2_ManSetupBmcSolver` | `TestSetupBmcSolverConstrainsFinalState` |
| `Int2_ManCheckFrames`, `Int2_ManWriteFrames`, `Int2_GoCreateFrames` | `TestCreateFramesPreservesComplementedAndFanins`, `TestCreateFramesMapsConstantAtEveryFrame`, `TestCreateFramesAllocatesDirectInputSATVariable`, `TestCreateFramesCanStartAtLaterFrame`, `TestSourceCopiesNativeXorAsAnd`. The C builder is incomplete; these tests establish Go-only behavior. `TestSourceFrameBuilderReportsUnsupportedBeforeMutation` checks the source-named entry point. |
| `Int2_ManCheckBmc` | `TestCreateFramesAllocatesDirectInputSATVariable`, `TestCheckBmcHandlesUnconstructedOutputs`, `TestCheckBmcFindsFailureAfterUnconstructedOutput`, `TestCheckBmcCubeAssumptionsMatchSourceStatus`, `TestCheckBmcReturnsUndecidedOnExpiredDeadline` |
| `Int2_ManUnroll`, `Int2_ManPreparePrefix` | `TestSourceUnrollAndPrefixReturnUnsupported` checks the source-named entry points. `TestFrameUnrollingTracksLatestInput`, `TestZeroFrameBuildersReturnInitialState`, and `TestPreparePrefixAtFrameZeroConstrainsInitialState` exercise separately named Go-only reconstructions. |
| `Int2_ManPrepareSuffix` | `TestPrepareSuffixReportsUnsupported`, `TestPrepareSuffixUnsupportedWithOptionalOutputs` check only the unsupported result. |
| `Int2_ManComputePreimage`, `Int2_ManPerformInterpolation` | `TestPreimageReportsUnsupportedForPrimaryInput`, `TestPerformInterpolationReportsUnsupported` check only the unsupported result. |
| `Int2_ManJustify_rec` | `TestSourceJustificationUsesPriorityAsBoolean`, `TestSourceJustificationChoosesByMarks`, `TestSourceJustificationSingleControllingFanin`, `TestSourceJustificationOutputMarkPreventsDuplicateSelection` |
| `Int2_ManRefineCube` | `TestRefineCubeSourceRejectsRegisterCI`, `TestSourceRefineCubeVisitsControllingPrimaryInput`, `TestSourceRefineCubeTrueAndWithUnitPriority`; these check source behavior, not correct cube minimization. |
| `Int2_ManComputeCoPres` | `TestInt2ManComputeCoPresRejectsMalformedSOP`, `TestSourceSOPIndexSkipsSecondContiguousLiteral`, `TestSourceCoPresDeduplicatesAndOrdersSingleLiteralCubes` |
| `Int2_ManCollectInternal`, `_rec` | `TestSourceCollectInternalVisitsSharedConesOnce` |
| `Int2_ManProbToGia` | `TestSourceNilSOPPanics`, `TestSourceEmptySOPProducesFalse`, `TestProbToGiaSingleLiteralCubesMatchSourceTransitions` |
| `int2HashCopyGate` | Used by frame tests; `TestSourceCopiesNativeXorAsAnd` records the C AND-copy behavior. |
| SAT bridge allocation, literals, clauses and solver | `TestSATBridgeLoadsClausesAndDetectsContradiction`, `TestSATBridgeClauseExpandsVariableRange`, `TestSATBridgeCumulativeGlobalLimits`, `TestSATBridgeExpiredGlobalBudgetReturnsUnknown` |
| SAT bridge AND, buffer, model and final core | `TestSATBridgeAndGateAllInputPolarities`, `TestSATBridgeBufferComplement`, `TestSATBridgeRejectsInconsistentGateClauses`, `TestSATBridgeModelValueUsesBooleanConvention`, `TestSATBridgeFinalReturnsNegatedAssumptions` |

The SAT bridge tests exercise an exhausted global propagation budget and
contradictory gate clauses, including the C AND helper's assertions; BMC
tests cover UNKNOWN from an expired deadline for both property and cube checks,
followed by a successful retry. Failure paths dependent on JF CNF derivation
and all SAT timeout modes are not exhaustively exercised. Reaching the bodies of the
unsupported suffix/preimage stubs does not validate those algorithms. The
zero-frame behavior and incremental builder have no executable C oracle.
