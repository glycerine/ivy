Here’s the test-driven plan I’d use to close the remaining gaps and move from “working subset” toward Python `ivy_to_cpp.py` parity.

**Guiding Rules**
- Add failing tests first, then implement the smallest generator behavior that makes them pass.
- After each small step: `go test -v -tags xtracer_off ./ivy2cpp`.
- At phase boundaries: `make test`.
- Use Python `ivy_to_cpp` as an oracle for small fixtures, but compare structure/snippets unless exact output is genuinely stable.
- Keep `ivy2cpp` as its own package; add accessors to `goivy` only when needed.

**Phase 1: Oracle Harness Expansion**
Add a fixture runner that generates Python and Go C++ for the same Ivy source across targets.

Tests:
- `TestPythonOracleFixturesImplReplTestGen`
- `TestPythonAndGoGeneratedCppCompileMatrix`
- `TestPythonOracleRecordsDiffArtifactsOnFailure`

Work:
- Build a table-driven oracle helper with fixture name, source, target, expected shared snippets, and compile mode.
- Add compile support for Python `target=gen` by supplying required stubs such as `__ivy_modelfile` when needed.
- Store normalized output in temp dirs on failure for inspection.

**Phase 2: Type System Parity**
Fill in missing `ivy_cpp_types.py` style behavior.

Tests:
- `TestCTypeNativePrimitiveAndClass`
- `TestCTypeBitVectorLikeNativeTypes`
- `TestVariantTypeDeclarationMatchesPythonShape`
- `TestTupleHashAndMapKeyDeclarations`
- `TestEnumSerializationShapeAgainstPython`

Work:
- Port richer type helpers for native interpreted sorts, variants, tuple/hash support, string/bitvector-ish native conventions where Python emits them.
- Add explicit variant declarations and constructors.
- Add hash/ordering support only where generated C++ actually needs it.

**Phase 3: Module/Class Shell Parity**
Close missing class-level runtime pieces.

Tests:
- `TestClassRuntimeMembersMatchPythonShape`
- `TestTickProgressHooksGenerated`
- `TestAssertAssumeProgressVirtuals`
- `TestThreadTimerReaderDeclarationsWhenRepl`

Work:
- Add Python-like runtime members where needed: progress stack, tick hook, assert/assume/check progress signatures.
- Keep non-REPL output lean unless Python parity or C++ compilation requires runtime support.
- Add `EmitMain` behavior or explicitly remove it from `Config` if unused.

**Phase 4: Serialization And REPL Protocol**
Current REPL dispatch is minimal. Python has real command parsing/serialization.

Tests:
- `TestReplSerializesEnumArgs`
- `TestReplDeserializesBoolIntEnum`
- `TestReplRejectsBadActionName`
- `TestReplParameterizedActionUsesParsedArgs`
- `TestPythonOracleReplCommandShape`

Work:
- Port enough `ivy_value`, serializer/deserializer, and reader code for exported action arguments.
- Replace default placeholder arguments with parsed command args.
- Add generated C++ smoke tests that run the executable with stdin commands for small fixtures.

**Phase 5: Expression Gaps**
Broaden expression support beyond the current common cases.

Tests:
- `TestEmitExprInterpretedArithmeticOps`
- `TestEmitExprIteNested`
- `TestEmitExprApplyNullaryFunction`
- `TestEmitExprVariantConstructorAndDestructor`
- `TestEmitExprNativeAntiquotesPythonParity`

Work:
- Audit Python expression emitters and add missing interpreted operators.
- Add constructor/destructor/variant expression support.
- Keep unsupported errors actionable with Go type and Ivy expression string.

**Phase 6: Action Gaps**
Current action coverage is good but not complete.

Tests:
- `TestEmitAssignVariantField`
- `TestEmitCallExternalAndInternalNames`
- `TestEmitChoiceWithNestedReturns`
- `TestEmitOldBindSemantics`
- `TestEmitNativeActionWithZ3Antiquotes`

Work:
- Tighten call name parity: `ext:` and implementation/internal naming.
- Improve `bindolds`/old-state behavior instead of pass-through where semantics require snapshots.
- Add missing variant/constructor field assignments.
- Expand native antiquote handling for solver/Z3 contexts.

**Phase 7: Initial State Parity**
We currently support `after init` and a small equality/iff slice. Python can derive more from constraints/models.

Tests:
- `TestInitForAllRelationConstraint`
- `TestInitExistsReturnsUnsupportedClearError`
- `TestInitModelFiniteEnumConstant`
- `TestInitModelFiniteRelation`
- `TestPythonOracleInitialStateFixtures`

Work:
- Port safe finite initial-condition lowering: equality, iff, universal finite assignments.
- For model-based initialization, use existing Go solver/model APIs only after tests pin expected behavior.
- Keep unsupported constraints failing clearly rather than silently emitting wrong C++.

**Phase 8: Full `target=test` / `target=gen`**
We have the first Z3 slice; this is the biggest remaining parity gap.

Tests:
- `TestTargetTestGeneratesInitGeneratorClass`
- `TestTargetTestGeneratesActionGeneratorClass`
- `TestTargetGenExternalHooksMatchPythonShape`
- `TestZ3FromSolverEnumRangeBool`
- `TestZ3RandomizeRelationArityTwo`
- `TestGeneratedGenFixtureCompilesWithStubs`
- `TestGeneratedTestExecutableRunsSmallFixture`

Work:
- Port Python’s generator class structure: `init_gen`, per-action generators, `generate`, `execute`, push/pop/progress.
- Implement solver conversion templates for bool, int/range, enum, strings/native where feasible.
- Expand randomizers for functions/relations of multiple arities.
- Add generator hooks for action parameters and return values.

**Phase 9: Build And CLI Completeness**
`build=true` works for basic cases. Harden it.

Tests:
- `TestCLIBuildTrueImpl`
- `TestCLIBuildTrueRepl`
- `TestCLIBuildTrueTestWithZ3`
- `TestCLIRejectsUnsupportedParam`
- `TestCLIUsesOutdirAndClassnameTogether`
- `TestBuildUsesZ3DIRWhenSet`

Work:
- Make build behavior consistent across targets.
- Add clearer build diagnostics for missing compiler/Z3.
- Support `Z3DIR`, vendored Z3, and system Z3 predictably.
- Consider returning build artifact metadata from a higher-level library API.

**Phase 10: Regression Fixture Growth**
Add small fixtures one feature at a time.

Fixture order:
- empty
- enum
- range
- bool relation
- unary/binary relation
- function
- initializer
- local/if/while/choice
- action call with returns
- native blocks
- destructor struct
- variant
- REPL parsing
- test/gen random action
- model-based init

Each fixture should have:
- Python oracle structural comparison where Python supports it.
- Go generated C++ compile test.
- Runtime smoke test only when stable and cheap.

**Done Criteria**
- `go test -v -tags xtracer_off ./ivy2cpp` passes with oracle tests enabled.
- `make test` passes.
- For each supported target, at least one Python/Go oracle fixture compiles.
- Unsupported features fail with explicit errors, not emitted placeholder comments.
- The README or CLI help documents current supported targets, params, and known gaps.
