# ivy2go ↔ ivy2cpp function-name divergence audit

Created: 2026-05-25T19:53:50.254290Z

## Method

Source files: ivy2cpp/*.go (canonical) and ivy2go/*.go (port).
Extraction: every top-level `func` declaration, including method receivers.
Scope: 28 common files between ivy2cpp and ivy2go. Excluded files:
  - ivy2cpp only: build_findvs.go, build_findvs_nonwindows.go, build_findvs_windows.go, clauses_helpers.go, cpp_context.go, cpp_types.go, ptype.go
  - ivy2go only: config.go, doc.go, go_context.go, go_types.go, state.go

Normalisation: cppX ↔ goX prefixes treated as equivalent; cppWriter ↔ goWriter receiver; Generator and Config considered role-equivalent.

## OMITTED — ivy2cpp functions missing from ivy2go

These are CLAUDE.md rule B.6 violations: functions that exist in the canonical ivy2cpp but are absent from ivy2go.

| ivy2cpp signature | source file:line | role |
| --- | --- | --- |
| `(g *Generator) emitSet(w *cppWriter, a *goivy.LogicSetAction) {` | action.go:148 | |
| `setTargetAndValue(lit goivy.Expr) (goivy.Expr, string) {` | action.go:168 | |
| `(g *Generator) closeAssignmentLoops(w *cppWriter, loops int) {` | action.go:207 | |
| `(g *Generator) openAssignmentLoops(w *cppWriter, lhs goivy.Expr) (int, bool) {` | action.go:213 | |
| `(g *Generator) someConditionLoopHeaders(some *goivy.SomeCondition) ([]string, error) {` | action.go:332 | |
| `firstParamIsIndex(some *goivy.SomeCondition) bool {` | action.go:474 | |
| `(g *Generator) emitIfSomeExtensional(w *cppWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) bool {` | action.go:488 | |
| `(g *Generator) emitIfSomeVariantDowncast(w *cppWriter, a *goivy.LogicIfAction, some *goivy.SomeCondition) bool {` | action.go:565 | |
| `(g *Generator) emitCallStackPush(w *cppWriter, a *goivy.LogicCallAction) bool {` | action.go:798 | |
| `(g *Generator) emitCallStackPop(w *cppWriter, stacked bool) {` | action.go:806 | |
| `(g *Generator) emitLet(w *cppWriter, a *goivy.LogicLetAction) {` | action.go:845 | |
| `(g *Generator) emitBindOlds(w *cppWriter, a *goivy.LogicBindOldsAction) {` | action.go:882 | |
| `(g *Generator) emitAssignField(w *cppWriter, a *goivy.LogicAssignFieldAction) {` | action.go:900 | |
| `(g *Generator) emitNullField(w *cppWriter, a *goivy.LogicNullFieldAction) {` | action.go:910 | |
| `(g *Generator) emitCopyField(w *cppWriter, a *goivy.LogicCopyFieldAction) {` | action.go:924 | |
| `(g *Generator) emitFieldRef(obj, field goivy.Expr) (string, error) {` | action.go:938 | |
| `fieldRangeSort(field goivy.Expr) (goivy.Sort, error) {` | action.go:953 | |
| `(g *Generator) emitDebug(w *cppWriter, a *goivy.LogicDebugAction) {` | action.go:983 | |
| `(g *Generator) emitPrintExpr(w *cppWriter, expr goivy.Expr) {` | action.go:1010 | |
| `debugEventName(e goivy.Expr) string {` | action.go:1043 | |
| `escapeString(s string) string {` | action.go:1055 | |
| `escapeComment(s string) string {` | action.go:1061 | |
| `(g *Generator) emitActionGenClassHeader(w *cppWriter, plan *actionGenPlan) {` | action_gen.go:193 | |
| `(g *Generator) emitActionGenMemberDecls(w *cppWriter, plan *actionGenPlan) {` | action_gen.go:207 | |
| `(g *Generator) emitPythonTestActionGen(w *cppWriter, plan *actionGenPlan) {` | action_gen.go:383 | |
| `(g *Generator) emitWeakActionGenerator(w *cppWriter, plan *actionGenPlan) {` | action_gen.go:514 | |
| `(g *Generator) emitPythonTestActionGenExecute(w *cppWriter, plan *actionGenPlan) {` | action_gen.go:639 | |
| `(g *Generator) formulaToSmtlib(fmla goivy.Expr) (string, bool) {` | action_gen.go:694 | |
| `preDefinedNames(clauses *goivy.Clauses) map[string]bool {` | action_gen.go:709 | |
| `preUsedContains(used *goivy.InsMap[goivy.NodeKey, goivy.Expr], name string) bool {` | action_gen.go:725 | |
| `(p *actionGenPlan) defedParamSet() map[goivy.NodeKey]bool {` | action_gen.go:738 | |
| `exprRoot(f goivy.Expr) goivy.Expr {` | action_gen.go:754 | |
| `exprAsConst(f goivy.Expr) (*goivy.Const, bool) {` | action_gen.go:767 | |
| `(g *Generator) emitTracedLHSInto(b *strings.Builder, lhs goivy.Expr) error {` | assign.go:93 | |
| `(g *Generator) emitTraceArgs(b *strings.Builder, args []goivy.Expr) error {` | assign.go:143 | |
| `(g *Generator) assignBoundsExpr(a *goivy.LogicAssignAction) goivy.Expr {` | assign.go:167 | |
| `(g *Generator) openAssignmentLoopsBounded(w *cppWriter, lhs goivy.Expr, body goivy.Expr) (int, bool) {` | assign.go:206 | |
| `(g *Generator) canOpenAssignmentLoopsBounded(lhs goivy.Expr, body goivy.Expr) bool {` | assign.go:247 | |
| `(g *Generator) actionsCfg() *goivy.ActionsConfig {` | assign.go:269 | |
| `(e *missingZ3ToolchainError) Error() string {` | build.go:20 | |
| `isMissingZ3ToolchainError(err error) bool {` | build.go:24 | |
| `buildPlanCompiler(compiler string) (string, error) {` | build.go:137 | |
| `msvcBuildPlan(out *Output, compiler, cppPath, outputPath string, compileOnly bool) (*BuildPlan, error) {` | build.go:156 | |
| `msvcIncludeDirArgs(dirs []string) []string {` | build.go:197 | |
| `msvcLibDirArgs(dirs []string) []string {` | build.go:208 | |
| `msvcIncludeArgs(args []string) []string {` | build.go:219 | |
| `msvcLinkArgs(args []string) []string {` | build.go:231 | |
| `supportIncludeArgs() ([]string, error) {` | build.go:247 | |
| `cxxCompiler() (string, error) {` | build.go:259 | |
| `cxxCompilerFor(compiler string) (string, error) {` | build.go:263 | |
| `outputUsesZ3(out *Output) bool {` | build.go:294 | |
| `outputUsesWideBV(out *Output) bool {` | build.go:301 | |
| `z3BuildArgs() ([]string, []string, error) {` | build.go:311 | |
| `existingIncludeDirs(candidates []string) []string {` | build.go:340 | |
| `existingZ3LibDirs(candidates []string) []string {` | build.go:359 | |
| `hasZ3Lib(dir string) bool {` | build.go:374 | |
| `packageGoivyRoot() (string, error) {` | build.go:383 | |
| `includeLibSpecArgs(out *Output) []string {` | build.go:391 | |
| `linkLibSpecArgs(out *Output) []string {` | build.go:401 | |
| `combinedLibSpecs(out *Output) []string {` | build.go:422 | |
| `readSpecsFileLibs() []string {` | build.go:443 | |
| `bvRawType(it cppInterpType) string {` | bv_expr.go:196 | |
| `(g *Generator) bvCastExpr(it cppInterpType, expr string) string {` | bv_expr.go:203 | |
| `bvZeroExpr(it cppInterpType) string {` | bv_expr.go:207 | |
| `(g *Generator) bvShiftAmountExpr(value string, sort goivy.Sort) string {` | bv_expr.go:214 | |
| `pruneStateStoresToSignature(mod *goivy.Module) {` | compile.go:121 | |
| `snapshotCPPInterface(mod *goivy.Module) cppInterfaceSnapshot {` | compile.go:401 | |
| `restoreCPPInterface(mod *goivy.Module, snap cppInterfaceSnapshot) {` | compile.go:447 | |
| `restoreActionSignature(mod *goivy.Module, name string, sig cppActionSignature) {` | compile.go:490 | |
| `restoreSortForCPPInterface(mod *goivy.Module, snap cppInterfaceSnapshot, sort goivy.Sort, seen map[string]bool) {` | compile.go:506 | |
| `ensureSortOrderName(mod *goivy.Module, snap cppInterfaceSnapshot, name string) {` | compile.go:557 | |
| `prepareModuleForCPP(mod *goivy.Module, cfg Config) {` | compile.go:589 | |
| `ensureSortOrderForCPP(mod *goivy.Module) {` | compile.go:618 | |
| `addConjsToActions(mod *goivy.Module) {` | compile.go:714 | |
| `descriptorJSON(mod *goivy.Module, cfg Config, outputs []*Output, isolates []string) (string, error) {` | compile.go:748 | |
| `describeParams(params []*goivy.Const) []descriptorParamDesc {` | compile.go:778 | |
| `(g *Generator) constructorActionFor(cons *goivy.Const, destrs []*goivy.Const) (string, goivy.Action, error) {` | constructors.go:24 | |
| `(g *Generator) emitConstructors(w *cppWriter) {` | constructors.go:82 | |
| `(g *Generator) derivedDefinitions() []derivedDefinition {` | definitions.go:17 | |
| `(g *Generator) nativeDefinitions() []derivedDefinition {` | definitions.go:35 | |
| `(g *Generator) allDefinitions() []derivedDefinition {` | definitions.go:53 | |
| `newDerivedDefinition(def *goivy.LogicDefinition) (derivedDefinition, bool) {` | definitions.go:58 | |
| `(g *Generator) definitionNames() map[string]bool {` | definitions.go:79 | |
| `(g *Generator) definitionByName(name string) (derivedDefinition, bool) {` | definitions.go:91 | |
| `(g *Generator) derivedActionFor(d derivedDefinition) (string, goivy.Action, error) {` | definitions.go:127 | |
| `destructorIndexVarName(i int) string {` | destructor.go:13 | |
| `(g *Generator) emitDomainLoops(w *cppWriter, dom []goivy.Sort) ([]string, func()) {` | destructor.go:21 | |
| `(g *Generator) destructorSolverName(d *goivy.Const) string {` | destructor.go:44 | |
| `(g *Generator) isReallyUninterpretedRange(s goivy.Sort) bool {` | destructor.go:55 | |
| `(g *Generator) cppSortCardStr(s goivy.Sort) string {` | destructor.go:79 | |
| `(g *Generator) emitDestructorStructHash(w *cppWriter, destructors []*goivy.Const) {` | destructor.go:91 | |
| `(g *Generator) emitDestructorStructComparators(w *cppWriter, name string, destructors []*goivy.Const) {` | destructor.go:122 | |
| `(g *Generator) emitDestructorStructWriter(w *cppWriter, name string, destructors []*goivy.Const) {` | destructor.go:182 | |
| `(g *Generator) emitDestructorSortArgSpecDecls(w *cppWriter) {` | destructor.go:258 | |
| `(g *Generator) emitDestructorImpls(w *cppWriter) {` | destructor.go:298 | |
| `(g *Generator) emitDestructorOutSerImpls(w *cppWriter) {` | destructor.go:310 | |
| `(g *Generator) emitDestructorArgDeserZ3Impls(w *cppWriter) {` | destructor.go:325 | |
| `(g *Generator) emitDestructorImpl(w *cppWriter, name string) {` | destructor.go:343 | |
| `(g *Generator) emitDestructorOutImpl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {` | destructor.go:367 | |
| `(g *Generator) emitDestructorSerImpl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {` | destructor.go:418 | |
| `(g *Generator) emitDestructorDeserImpl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {` | destructor.go:450 | |
| `(g *Generator) emitDestructorArgImpl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {` | destructor.go:491 | |
| `(g *Generator) emitDestructorZeroInit(w *cppWriter, lhsPrefix string, destrs []*goivy.Const) {` | destructor.go:582 | |
| `(g *Generator) emitZeroAssign(w *cppWriter, lhs string, s goivy.Sort) {` | destructor.go:608 | |
| `(g *Generator) emitDestructorZ3Impl(w *cppWriter, name, typeName string, destrs []*goivy.Const) {` | destructor.go:636 | |
| `variantPayloadField(s goivy.Sort) string {` | expr.go:402 | |
| `(g *Generator) emitQuantWithHeaders(vars []*goivy.LogicVariable, body goivy.Expr, forall bool, headers []string, iterFirstOnly bool) (string, error) {` | expr.go:554 | |
| `(g *Generator) quantIterableHeader(v *goivy.LogicVariable) (string, bool, error) {` | expr.go:604 | |
| `(g *Generator) emitExistsVariantRelation(vars []*goivy.LogicVariable, body goivy.Expr) (string, bool, error) {` | expr.go:620 | |
| `(g *Generator) emitExtensionalQuant(vars []*goivy.LogicVariable, body goivy.Expr, forall bool) (string, bool, error) {` | expr.go:654 | |
| `(g *Generator) matchExtensionalBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]*goivy.Apply) {` | expr.go:765 | |
| `containsVariableByName(terms []goivy.Expr, name string) bool {` | expr.go:851 | |
| `(g *Generator) matchBoundExprs(v0 *goivy.LogicVariable, body goivy.Expr, exists bool, res *[]boundExpr) {` | expr.go:879 | |
| `(g *Generator) sortHasNegativeValues(s goivy.Sort) bool {` | expr.go:965 | |
| `nameOfTerm(t goivy.Expr) string {` | expr.go:973 | |
| `nameIn(vars []*goivy.LogicVariable, n string) bool {` | expr.go:985 | |
| `(g *Generator) getBounds(v0 *goivy.LogicVariable, others []*goivy.LogicVariable, body goivy.Expr, exists bool) (string, string, error) {` | expr.go:999 | |
| `(g *Generator) sortCardinalityAttr(s goivy.Sort) (string, bool) {` | expr.go:1078 | |
| `(g *Generator) getAllBounds(vars []*goivy.LogicVariable, body goivy.Expr, exists bool) ([][2]string, error) {` | expr.go:1118 | |
| `(g *Generator) iterableSortFor(s goivy.Sort) (string, goivy.Sort, bool) {` | expr.go:1138 | |
| `(g *Generator) someLoopHeaders(vars []*goivy.LogicVariable, body goivy.Expr) ([]string, error) {` | expr.go:1213 | |
| `(g *Generator) emitSomeVariantRelation(s *goivy.LogicSome) (string, bool, error) {` | expr.go:1242 | |
| `(g *Generator) emitSomeWithElse(s *goivy.LogicSome) (string, error) {` | expr.go:1300 | |
| `finiteValues(s goivy.Sort) ([]string, bool) {` | expr.go:1347 | |
| `(g *Generator) loopHeaderForSortBounds(s goivy.Sort, name, lo, hi string) (string, error) {` | expr.go:1395 | |
| `loopIntCType(g *Generator, s goivy.Sort) string {` | expr.go:1413 | |
| `(g *Generator) rangeSortFor(s goivy.Sort) (*goivy.RangeSort, bool) {` | expr.go:1433 | |
| `numericRangeBounds(rs *goivy.RangeSort) (string, string, bool) {` | expr.go:1447 | |
| `(g *Generator) extensionalRels() map[string]bool {` | extensional.go:11 | |
| `(g *Generator) markBadExtensional(sub goivy.Action, bad map[string]bool) {` | extensional.go:85 | |
| `(g *Generator) collectInitedExtensional(action goivy.Action, inited map[string]bool) {` | extensional.go:123 | |
| `(g *Generator) isDestructorSort(name string) bool {` | extensional.go:148 | |
| `repName(e goivy.Expr) string {` | extensional.go:160 | |
| `argsOf(e goivy.Expr) []goivy.Expr {` | extensional.go:173 | |
| `allArgsNonVariable(args []goivy.Expr) bool {` | extensional.go:182 | |
| `allArgsVariable(args []goivy.Expr) bool {` | extensional.go:193 | |
| `(g *Generator) emitExtensionalRelationClear(w *cppWriter, a *goivy.LogicAssignAction) bool {` | extensional.go:213 | |
| `(g *Generator) nextTemp(prefix string) string {` | generator.go:226 | |
| `(g *Generator) emitHeader() error {` | generator.go:232 | |
| `(g *Generator) emitImpl() error {` | generator.go:291 | |
| `(g *Generator) constructorSignature(qualified bool) string {` | generator.go:414 | |
| `(g *Generator) emitConstructorParamAssignments(w *cppWriter) {` | generator.go:428 | |
| `(g *Generator) emitDestructorEqualityInlines(w *cppWriter) {` | generator.go:594 | |
| `(g *Generator) emitCTupleDecls(w *cppWriter) {` | generator.go:625 | |
| `(g *Generator) emitCTupleHashDecls(w *cppWriter) {` | generator.go:652 | |
| `(g *Generator) emitCTupleEqualities(w *cppWriter) {` | generator.go:673 | |
| `(g *Generator) emitVariantSuperComparators(w *cppWriter, typeName string, variants []goivy.Sort) {` | generator.go:714 | |
| `(g *Generator) emitVariantSuperWriter(w *cppWriter, typeName string, variants []goivy.Sort) {` | generator.go:746 | |
| `(g *Generator) emitStateDecls(w *cppWriter) {` | generator.go:805 | |
| `(g *Generator) cardinalitySortNames() []string {` | generator.go:816 | |
| `(g *Generator) emitCardinalityDecls(w *cppWriter) {` | generator.go:841 | |
| `(g *Generator) emitCardinalityInitializers(w *cppWriter) {` | generator.go:851 | |
| `(g *Generator) shouldInitializeCardinality(name string) bool {` | generator.go:867 | |
| `(g *Generator) sortNeededForRuntimeSpecs(name string) bool {` | generator.go:885 | |
| `(g *Generator) exprReferencesSortName(e goivy.Expr, name string, seen map[goivy.NodeKey]bool) bool {` | generator.go:923 | |
| `(g *Generator) sortDependencyReferencesName(s goivy.Sort, name string, seen map[string]bool) bool {` | generator.go:943 | |
| `sortReferencesName(s goivy.Sort, name string) bool {` | generator.go:973 | |
| `(g *Generator) allStateSymbols() []stateSymbol {` | generator.go:1006 | |
| `(g *Generator) stateSymbols() []stateSymbol {` | generator.go:1067 | |
| `(g *Generator) isSortConstructorName(name string) bool {` | generator.go:1098 | |
| `(g *Generator) emitMethodDecls(w *cppWriter) {` | generator.go:1119 | |
| `(g *Generator) emitMethodDeclLine(w *cppWriter, name string, act goivy.Action) {` | generator.go:1135 | |
| `(g *Generator) methodSignature(name string, act goivy.Action, qualified, inline bool) string {` | generator.go:1147 | |
| `(g *Generator) emitMethods(w *cppWriter) {` | generator.go:1252 | |
| `(g *Generator) emitSomeAction(w *cppWriter, name string, act goivy.Action) {` | generator.go:1272 | |
| `(g *Generator) emitTraceActionPrologue(w *cppWriter, name string, formals []*goivy.Const) {` | generator.go:1370 | |
| `formalListContains(formals []*goivy.Const, target *goivy.Const) bool {` | generator.go:1424 | |
| `(g *Generator) emitReplSupport(w *cppWriter) {` | generator.go:1438 | |
| `(g *Generator) hasNonInitPublicActions() bool {` | generator.go:1488 | |
| `(g *Generator) emitPythonZeroParamTestMain(w *cppWriter) {` | generator.go:1527 | |
| `(g *Generator) emitTestLoopBody(w *cppWriter) {` | generator.go:1852 | |
| `(g *Generator) emitTestLoopGenBranch(w *cppWriter) {` | generator.go:1937 | |
| `(g *Generator) emitTestLoopSelectBranch(w *cppWriter) {` | generator.go:1976 | |
| `isFinalizeName(name string) bool {` | generator.go:2081 | |
| `(g *Generator) hasFinalizeExport() bool {` | generator.go:2085 | |
| `(g *Generator) actionWeight(name string) float64 {` | generator.go:2099 | |
| `pythonFloatLiteral(f float64) string {` | generator.go:2128 | |
| `(g *Generator) emitGenMain(w *cppWriter) {` | generator.go:2137 | |
| `(g *Generator) emitTestDefaults(w *cppWriter) {` | generator.go:2157 | |
| `(g *Generator) emitGeneratorInvocations(w *cppWriter) {` | generator.go:2166 | |
| `(g *Generator) emitConstructDefaultObject(w *cppWriter) {` | generator.go:2183 | |
| `(g *Generator) emitConstructFromParams(w *cppWriter) {` | generator.go:2194 | |
| `(g *Generator) constructorDefaultArgs() []string {` | generator.go:2206 | |
| `(g *Generator) initialMixinActionNames() map[string]bool {` | generator.go:2214 | |
| `(g *Generator) hasInitialMixinActions() bool {` | generator.go:2225 | |
| `(g *Generator) initialConditionActions() ([]goivy.Action, error) {` | init.go:9 | |
| `(g *Generator) initialConditionActionsFor(f goivy.Expr) ([]goivy.Action, error) {` | init.go:24 | |
| `(g *Generator) initialConditionAction(f goivy.Expr) (goivy.Action, error) {` | init.go:47 | |
| `(g *Generator) isStateTarget(e goivy.Expr) bool {` | init.go:85 | |
| `stateTargetSymbol(e goivy.Expr) *goivy.Const {` | init.go:115 | |
| `(g *Generator) initialStateConstraints() (*initialStateConstraints, error) {` | initial_state.go:23 | |
| `(g *Generator) checkInitialStateParameters(kind string, lfs []*goivy.LabeledFormula) error {` | initial_state.go:56 | |
| `usedSymbolNames(formulas []goivy.Expr) map[string]bool {` | initial_state.go:85 | |
| `(g *Generator) emitOneInitialState(w *cppWriter) error {` | initial_state.go:101 | |
| `(g *Generator) emitDefaultInitialState(w *cppWriter, sym stateSymbol) {` | initial_state.go:134 | |
| `(g *Generator) emitSolvedInitialState(w *cppWriter, slv *goivy.Solver, model *smt.Model, sym stateSymbol) error {` | initial_state.go:142 | |
| `initialStateRange(s goivy.Sort) goivy.Sort {` | initial_state.go:178 | |
| `initialStateSymbolTerm(sym stateSymbol, args []goivy.Expr) goivy.Expr {` | initial_state.go:185 | |
| `(g *Generator) initialModelValue(slv *goivy.Solver, model *smt.Model, term goivy.Expr, s goivy.Sort) (string, error) {` | initial_state.go:205 | |
| `(g *Generator) modelValueToCpp(value smt.Z3Expr, s goivy.Sort) (string, error) {` | initial_state.go:217 | |
| `stripZ3Bars(s string) string {` | initial_state.go:250 | |
| `parseTrailingModelIndex(s string) (int, bool) {` | initial_state.go:257 | |
| `(g *Generator) initialDomainTuples(domain []goivy.Sort) ([][]initialDomainValue, bool) {` | initial_state.go:269 | |
| `(g *Generator) initialDomainValues(s goivy.Sort) ([]initialDomainValue, bool) {` | initial_state.go:289 | |
| `(g *Generator) emitProgressCounterResets(w *cppWriter, obj string) {` | initial_state.go:329 | |
| `(g *Generator) emitProgressCounterReset(w *cppWriter, p progressDecl, obj string) {` | initial_state.go:335 | |
| `(g *Generator) emitZ3InitialConstraints(w *cppWriter) error {` | initial_state.go:353 | |
| `(g *Generator) emitZ3AddInitialFormula(w *cppWriter, f goivy.Expr, env map[string]goivy.Sort) error {` | initial_state.go:366 | |
| `(g *Generator) emitZ3ForAllInitialFormula(w *cppWriter, f *goivy.ForAll, env map[string]goivy.Sort) error {` | initial_state.go:387 | |
| `cloneSortEnv(env map[string]goivy.Sort) map[string]goivy.Sort {` | initial_state.go:408 | |
| `(g *Generator) z3InitialExpr(e goivy.Expr, env map[string]goivy.Sort) (string, error) {` | initial_state.go:416 | |
| `(g *Generator) z3InitialNary(terms []goivy.Expr, op, ident string, env map[string]goivy.Sort) (string, error) {` | initial_state.go:478 | |
| `(g *Generator) z3InitialConst(c *goivy.Const, env map[string]goivy.Sort) (string, error) {` | initial_state.go:493 | |
| `(g *Generator) z3InitialVariable(name string, s goivy.Sort, env map[string]goivy.Sort) (string, error) {` | initial_state.go:528 | |
| `(g *Generator) z3InitialApply(a *goivy.Apply, env map[string]goivy.Sort) (string, error) {` | initial_state.go:540 | |
| `(g *Generator) z3InitialApplyArg(e goivy.Expr, env map[string]goivy.Sort) (string, error) {` | initial_state.go:567 | |
| `(g *Generator) z3HasDecl(name string) bool {` | initial_state.go:598 | |
| `(g *Generator) emitZ3InitialStateEvaluation(w *cppWriter, obj string) error {` | initial_state.go:610 | |
| `(g *Generator) emitZ3EvaluateStateSymbol(w *cppWriter, obj string, sym stateSymbol) error {` | initial_state.go:629 | |
| `(g *Generator) isParamName(name string) bool {` | initial_state.go:635 | |
| `classifyNativeTag(raw string) (nativeTagKind, string) {` | native.go:53 | |
| `(g *Generator) buildNativeBlocks() ([]nativeBlock, error) {` | native.go:86 | |
| `splitNativeCode(code string) (string, string) {` | native.go:130 | |
| `nativeExpr(node goivy.Node) (goivy.Expr, error) {` | native.go:142 | |
| `(g *Generator) onceMemo() map[string]bool {` | native.go:162 | |
| `(g *Generator) encodedSet() map[string]bool {` | native.go:175 | |
| `(g *Generator) emitHeaderNatives(w *cppWriter) error {` | native.go:185 | |
| `(g *Generator) emitClassMemberNatives(w *cppWriter) error {` | native.go:213 | |
| `(g *Generator) emitImplNatives(w *cppWriter) error {` | native.go:237 | |
| `(g *Generator) emitInitNatives(w *cppWriter) error {` | native.go:277 | |
| `(g *Generator) emitInlineNatives(w *cppWriter) error {` | native.go:298 | |
| `(g *Generator) renderNativeTemplate(code string, params []goivy.Expr) (string, error) {` | native.go:321 | |
| `(g *Generator) nativeTypeFull(nt *goivy.NativeType) (string, error) {` | native.go:354 | |
| `nativeCodeText(node goivy.Node) (string, error) {` | native.go:380 | |
| `(g *Generator) nativeReferenceInType(node goivy.Node) (string, error) {` | native.go:395 | |
| `(g *Generator) nativeTypeForSort(name string) (*goivy.NativeType, bool) {` | native.go:441 | |
| `(g *Generator) emitNativeClassTypeDecl(w *cppWriter, name, base string) {` | native.go:471 | |
| `(g *Generator) isCallbackAction(arg goivy.Node) (string, bool) {` | native.go:484 | |
| `callbackActionName(mod *goivy.Module, arg goivy.Node) (string, bool) {` | native.go:491 | |
| `nativeArgName(child goivy.Node) string {` | native.go:509 | |
| `(g *Generator) nativeReference(arg goivy.Expr) (string, error) {` | native.go:526 | |
| `(g *Generator) nativeTypeOf(arg goivy.Expr) (string, error) {` | native.go:580 | |
| `(g *Generator) nativeZ3Name(arg goivy.Expr) (string, error) {` | native.go:587 | |
| `(g *Generator) sortByName(name string) (goivy.Sort, bool) {` | native.go:600 | |
| `emitNativeLines(w *cppWriter, code string) {` | native.go:607 | |
| `nativeIndent(line string) int {` | native.go:636 | |
| `(g *Generator) collectCallbackActions() []string {` | native_thunk.go:27 | |
| `collectCallbackActionNames(mod *goivy.Module) []string {` | native_thunk.go:34 | |
| `unwrapCompiled(n goivy.Node) goivy.Node {` | native_thunk.go:85 | |
| `(g *Generator) emitCallbackThunks(w *cppWriter) {` | native_thunk.go:100 | |
| `(g *Generator) emitCallbackThunk(w *cppWriter, name string, act goivy.Action) {` | native_thunk.go:122 | |
| `(g *Generator) mkNondet(w *cppWriter, varExpr string, _rng int, label string, uniqueID int64, sort goivy.Sort) {` | nondet.go:32 | |
| `(g *Generator) mkNondetWithCType(w *cppWriter, varExpr string, label string, uniqueID int64, ctype string) {` | nondet.go:40 | |
| `(g *Generator) mkNondetSym(w *cppWriter, local goivy.Expr, name string, uniqueID int64) {` | nondet.go:54 | |
| `(g *Generator) mkNondetValue(w *cppWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64) {` | nondet.go:122 | |
| `(g *Generator) mkNondetValueScoped(w *cppWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64, className string) {` | nondet.go:126 | |
| `(g *Generator) mkNondetVariant(w *cppWriter, lhsExpr string, super goivy.Sort, name string, uniqueID int64) {` | nondet.go:145 | |
| `(g *Generator) mkNondetVariantScoped(w *cppWriter, lhsExpr string, super goivy.Sort, name string, uniqueID int64, className string) {` | nondet.go:149 | |
| `(g *Generator) mkNondetStructFields(w *cppWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64) {` | nondet.go:186 | |
| `(g *Generator) mkNondetStructFieldsScoped(w *cppWriter, lhsExpr string, sort goivy.Sort, name string, uniqueID int64, className string) {` | nondet.go:190 | |
| `(g *Generator) nondetSkipSort(s goivy.Sort) bool {` | nondet.go:216 | |
| `(g *Generator) makeNondetThunk(w *cppWriter, domSorts []goivy.Sort, rngSort goivy.Sort, name string, uniqueID int64) string {` | nondet.go:253 | |
| `TokenizeCPP(text string) []CPPToken {` | oracle_compare.go:25 | |
| `CompareCPPTokens(leftName, leftText, rightName, rightText string) CPPTokenDiff {` | oracle_compare.go:89 | |
| `cppTokenDiff(leftName, leftText, rightName, rightText string, left, right []CPPToken, index int) CPPTokenDiff {` | oracle_compare.go:119 | |
| `scanCPPQuoted(text string, start int) int {` | oracle_compare.go:144 | |
| `scanCPPRawString(text string, start int) (int, bool) {` | oracle_compare.go:160 | |
| `scanCPPNumber(text string, start int) int {` | oracle_compare.go:185 | |
| `scanCPPOperator(text string, start int) (string, bool) {` | oracle_compare.go:208 | |
| `cppContext(text string, offset, width int) string {` | oracle_compare.go:222 | |
| `isCPPWhitespace(c byte) bool {` | oracle_compare.go:248 | |
| `isCPPDigit(c byte) bool {` | oracle_compare.go:252 | |
| `isCPPAlpha(c byte) bool {` | oracle_compare.go:256 | |
| `isCPPIdentStart(c byte) bool {` | oracle_compare.go:260 | |
| `isCPPIdentPart(c byte) bool {` | oracle_compare.go:264 | |
| `(g *Generator) enumSortsForArgSpecs() []*goivy.LogicEnumeratedSort {` | repl.go:18 | |
| `(g *Generator) encodedSortSet() map[string]bool {` | repl.go:53 | |
| `(g *Generator) emitEnumSortArgSpecDecls(w *cppWriter) {` | repl.go:66 | |
| `(g *Generator) emitEnumSortArgSpecImpls(w *cppWriter) {` | repl.go:126 | |
| `(g *Generator) emitEnumSortOutSerImpls(w *cppWriter) {` | repl.go:142 | |
| `(g *Generator) emitEnumSortArgDeserImpls(w *cppWriter) {` | repl.go:150 | |
| `(g *Generator) emitEnumOperatorOut(w *cppWriter, st *goivy.LogicEnumeratedSort) {` | repl.go:159 | |
| `(g *Generator) emitEnumSer(w *cppWriter, st *goivy.LogicEnumeratedSort) {` | repl.go:182 | |
| `(g *Generator) emitEnumArg(w *cppWriter, st *goivy.LogicEnumeratedSort) {` | repl.go:200 | |
| `(g *Generator) emitEnumDeser(w *cppWriter, st *goivy.LogicEnumeratedSort) {` | repl.go:231 | |
| `(g *Generator) emitCmdReader(w *cppWriter) {` | repl.go:260 | |
| `(g *Generator) emitPythonTestCmdReaderRaw(w *cppWriter, readerClass, reprClass string) {` | repl.go:336 | |
| `(g *Generator) emitCmdReaderDispatchChain(w *cppWriter) {` | repl.go:416 | |
| `(g *Generator) emitPythonTestCmdReaderDispatchChain(w *cppWriter) {` | repl.go:479 | |
| `(g *Generator) emitDispatchArgExprs(act goivy.Action) []string {` | repl.go:547 | |
| `(g *Generator) argExprForSort(argsExpr, idxExpr string, s goivy.Sort) string {` | repl.go:556 | |
| `(g *Generator) argExprForSortBound(argsExpr, idxExpr string, s goivy.Sort, bound string) string {` | repl.go:560 | |
| `(g *Generator) emitTracePrelude(w *cppWriter, username string, argExprs []string) {` | repl.go:568 | |
| `(g *Generator) emitValueParser(w *cppWriter, p *goivy.Const, srcExpr string, lineno goivy.Location) {` | repl.go:596 | |
| `(g *Generator) emitMainParamSetup(w *cppWriter) {` | repl.go:625 | |
| `(g *Generator) emitParamKeyValueDispatch(w *cppWriter) {` | repl.go:693 | |
| `(g *Generator) emitPositionalParamParse(w *cppWriter) {` | repl.go:738 | |
| `(g *Generator) emitOnePositionalParam(w *cppWriter, p *goivy.Const, idx int) {` | repl.go:774 | |
| `(g *Generator) functionAppLHS(p *goivy.Const, dom []goivy.Sort, domArgs []string) string {` | repl.go:818 | |
| `isLargeFunctionDomain(g *Generator, dom []goivy.Sort) bool {` | repl.go:832 | |
| `(g *Generator) positionalParams() []*goivy.Const {` | repl.go:853 | |
| `paramDefaultText(n goivy.Node) string {` | repl.go:867 | |
| `(g *Generator) emitWinsockInit(w *cppWriter) {` | repl.go:883 | |
| `(g *Generator) replNeedsNumericParser(s goivy.Sort) bool {` | repl.go:909 | |
| `(g *Generator) publicActionNamesSorted() []string {` | repl.go:927 | |
| `(g *Generator) runtimeUsesGenerator() bool {` | runtime.go:11 | |
| `(g *Generator) runtimeUsesReplSubclass() bool {` | runtime.go:15 | |
| `(g *Generator) hostOS() string {` | runtime.go:19 | |
| `(g *Generator) emitRuntimeHeaderPreamble(w *cppWriter) {` | runtime.go:26 | |
| `(g *Generator) emitRuntimeHeaderForwardDecls(w *cppWriter) {` | runtime.go:63 | |
| `emitHashThunkSupport(w *cppWriter) {` | runtime.go:74 | |
| `(g *Generator) emitRuntimeClassMembers(w *cppWriter) {` | runtime.go:101 | |
| `(g *Generator) emitRuntimeImplPreamble(w *cppWriter) {` | runtime.go:154 | |
| `(g *Generator) emitRuntimeValueIncludes(w *cppWriter) {` | runtime.go:192 | |
| `(g *Generator) emitZ3Boilerplate1(w *cppWriter) {` | runtime.go:217 | |
| `(g *Generator) emitRuntimeConstructorPrelude(w *cppWriter) {` | runtime.go:233 | |
| `(g *Generator) emitRuntimeMethods(w *cppWriter) {` | runtime.go:257 | |
| `(g *Generator) emitRuntimeLockMethods(w *cppWriter) {` | runtime.go:264 | |
| `(g *Generator) emitRuntimeInstallMethods(w *cppWriter) {` | runtime.go:295 | |
| `(g *Generator) emitRuntimeThreadInstallMethod(w *cppWriter, method, typ, winFn, pthreadFn string) {` | runtime.go:325 | |
| `(g *Generator) emitRuntimeDestructor(w *cppWriter) {` | runtime.go:379 | |
| `(g *Generator) emitRuntimeChoose(w *cppWriter) {` | runtime.go:414 | |
| `(g *Generator) emitRuntimeReplSubclass(w *cppWriter) {` | runtime.go:456 | |
| `(g *Generator) emitReplImportCallbacks(w *cppWriter) {` | runtime.go:485 | |
| `(g *Generator) emitReplImportCallback(w *cppWriter, name string, act goivy.Action) {` | runtime.go:516 | |
| `(g *Generator) emitRuntimeReplAssertOverride(w *cppWriter, method, event, text string) {` | runtime.go:551 | |
| `(g *Generator) replSubclassConstructorSignature() string {` | runtime.go:576 | |
| `(g *Generator) baseConstructorCall() string {` | runtime.go:584 | |
| `(g *Generator) runtimeMainClassName() string {` | runtime.go:592 | |
| `(g *Generator) emitRuntimeOutputSetup(w *cppWriter) {` | runtime.go:599 | |
| `(g *Generator) emitRuntimeArgCapture(w *cppWriter, obj string) {` | runtime.go:605 | |
| `(g *Generator) emitRuntimeBindReaders(w *cppWriter) {` | runtime.go:617 | |
| `cleanSmtlib(s string) string {` | solver_emit.go:30 | |
| `(g *Generator) emitDeclSolver(w *cppWriter, sym stateSymbol) {` | solver_emit.go:48 | |
| `(g *Generator) emitDeclSolverWithName(w *cppWriter, sym stateSymbol, symNameExpr, prefix string) {` | solver_emit.go:52 | |
| `(o emitSetSolverOptions) rhsBase() string {` | solver_emit.go:123 | |
| `(o emitSetSolverOptions) addConstraint(w *cppWriter, text string) {` | solver_emit.go:130 | |
| `(g *Generator) emitSetSolverCustom(w *cppWriter, sym stateSymbol, opts emitSetSolverOptions) {` | solver_emit.go:138 | |
| `(g *Generator) isLargeType(s goivy.Sort) bool {` | solver_emit.go:255 | |
| `(g *Generator) emitSetField(w *cppWriter, destr *goivy.Const, lhs, rhs string, nvars int) bool {` | solver_emit.go:293 | |
| `(g *Generator) emitSetFieldCustom(w *cppWriter, destr *goivy.Const, lhs, rhs string, nvars int, opts emitSetSolverOptions) bool {` | solver_emit.go:309 | |
| `(g *Generator) emitRandomizeSolver(w *cppWriter, sym stateSymbol) error {` | solver_emit.go:378 | |
| `(g *Generator) emitEvalSolver(w *cppWriter, sym stateSymbol, lhsObj string) error {` | solver_emit.go:439 | |
| `(g *Generator) emitEvalSolverTo(w *cppWriter, sym stateSymbol, lhsObj, lhsOverride string) error {` | solver_emit.go:443 | |
| `(g *Generator) emitEvalSig(w *cppWriter, obj string, used map[string]bool) error {` | solver_emit.go:450 | |
| `(g *Generator) emitFromSolverLoop(w *cppWriter, obj string, sym stateSymbol, lhsOverride string) error {` | solver_emit.go:473 | |
| `(g *Generator) recordRangeType(s goivy.Sort) string {` | solver_emit.go:549 | |
| `(g *Generator) isRecordRange(s goivy.Sort) bool {` | solver_emit.go:570 | |
| `(g *Generator) uninterpretedRandomizeRangeError(s goivy.Sort) error {` | solver_emit.go:592 | |
| `(g *Generator) emitPythonTestRandomizeSolver(w *cppWriter, sym stateSymbol) error {` | solver_emit.go:605 | |
| `(g *Generator) emitPythonTestFromSolverLoop(w *cppWriter, obj string, sym stateSymbol, lhsOverride string) error {` | solver_emit.go:653 | |
| `(g *Generator) pythonTestRandExpr(s goivy.Sort) (string, bool) {` | solver_emit.go:707 | |
| `(g *Generator) pythonTestLoopHeaderForSort(s goivy.Sort, name string) (string, bool) {` | solver_emit.go:736 | |
| `(g *Generator) pythonTestSortBounds(s goivy.Sort) (string, string, bool) {` | solver_emit.go:744 | |
| `(g *Generator) pythonTestIntToZ3(s goivy.Sort, val string) string {` | solver_emit.go:767 | |
| `(g *Generator) mkRand(s goivy.Sort) (string, bool) {` | solver_emit.go:774 | |
| `(g *Generator) isDestructorRecordRange(s goivy.Sort) bool {` | solver_emit.go:784 | |
| `(g *Generator) destructorsOfRange(s goivy.Sort) []*goivy.Const {` | solver_emit.go:798 | |
| `joinArgs(args []string) string {` | solver_emit.go:806 | |
| `z3ValueForSortWithPrefix(s goivy.Sort, name, prefix string) string {` | solver_emit.go:813 | |
| `z3ApplyCall(prefix, nameExpr string, args []string) string {` | solver_emit.go:817 | |
| `(g *Generator) unsupportedEmitSet(w *cppWriter, sym stateSymbol, reason string) {` | solver_emit.go:824 | |
| `(g *Generator) emitDefinedInputs(` | solver_emit.go:843 | |
| `(g *Generator) emitDefinedInputExpr(` | solver_emit.go:880 | |
| `bracketize(args []string) string {` | solver_emit.go:961 | |
| `(g *Generator) emitInitGenPerSymbolDispatch(w *cppWriter, obj string) error {` | solver_emit.go:987 | |
| `(g *Generator) isPrimitiveRange(s goivy.Sort) bool {` | solver_emit.go:1021 | |
| `(g *Generator) emitAssignArrayFromModel(w *cppWriter, obj string, sym stateSymbol) {` | solver_emit.go:1034 | |
| `(g *Generator) makePythonTestNondetThunk(w *cppWriter, domSorts []goivy.Sort, rngSort goivy.Sort, name string, uniqueID int64) string {` | solver_emit.go:1090 | |
| `(g *Generator) pythonTestNondetZ3ValueExpr(s goivy.Sort, val string) string {` | solver_emit.go:1117 | |
| `(g *Generator) emitHashThunkToSolver(w *cppWriter, domSorts []goivy.Sort, ctName string) {` | solver_emit.go:1129 | |
| `(g *Generator) emitAllCtuplesToSolver(w *cppWriter) {` | solver_emit.go:1184 | |
| `(g *Generator) thunkSubstituteArgs(vs []*goivy.LogicVariable, expr goivy.Expr) (goivy.Expr, error) {` | thunk.go:164 | |
| `appliedFunctionConstKeys(expr goivy.Expr) map[goivy.NodeKey]bool {` | thunk.go:233 | |
| `(g *Generator) expandDerivedForThunk(expr goivy.Expr) goivy.Expr {` | thunk.go:255 | |
| `(g *Generator) expandDerivedForThunkOnce(expr goivy.Expr, defs []derivedDefinition) goivy.Expr {` | thunk.go:271 | |
| `allDefinitionParamsVariables(params []goivy.Expr) bool {` | thunk.go:311 | |
| `(g *Generator) emitThunkToZ3(w *cppWriter, name string, vs []*goivy.LogicVariable, origExpr goivy.Expr, envSyms []*goivy.Const) {` | thunk.go:320 | |
| `(g *Generator) emitThunkLocalZ3Symbol(w *cppWriter, sym *goivy.Const) string {` | thunk.go:440 | |
| `(g *Generator) allNumericOrEnumeratedConstants(used *goivy.InsMap[goivy.NodeKey, goivy.Expr]) bool {` | thunk.go:449 | |
| `(g *Generator) isNumericOrEnumeratedConstant(sym goivy.Expr) bool {` | thunk.go:461 | |
| `(g *Generator) thunkFastPathUsesToSolver(s goivy.Sort) bool {` | thunk.go:482 | |
| `(g *Generator) isPrimitiveSort(s goivy.Sort) bool {` | thunk.go:489 | |
| `(g *Generator) progressDecls() []progressDecl {` | tick.go:24 | |
| `(g *Generator) progressDeclFrom(item any) (progressDecl, bool, error) {` | tick.go:52 | |
| `progressTermVariable(term goivy.Expr) (*goivy.LogicVariable, error) {` | tick.go:78 | |
| `progressKey(expr goivy.Expr) string {` | tick.go:89 | |
| `(g *Generator) relyDecls() []relyDecl {` | tick.go:93 | |
| `relyDeclFrom(expr goivy.Expr) (relyDecl, bool, error) {` | tick.go:111 | |
| `(g *Generator) needsTickMax() bool {` | tick.go:129 | |
| `(g *Generator) emitProgressCounterDecls(w *cppWriter) {` | tick.go:138 | |
| `(g *Generator) progressCounterDecl(p progressDecl) string {` | tick.go:148 | |
| `progressCounterStorage(g *Generator, domain []goivy.Sort) cppFunctionStorage {` | tick.go:163 | |
| `(g *Generator) emitTick(w *cppWriter) {` | tick.go:198 | |
| `(g *Generator) emitProgressTickUpdates(w *cppWriter, progress []progressDecl) {` | tick.go:220 | |
| `(g *Generator) emitProgressRelyChecks(w *cppWriter, progress []progressDecl) {` | tick.go:235 | |
| `hasBareRely(relies []relyDecl) bool {` | tick.go:265 | |
| `(g *Generator) emitRelyMax(w *cppWriter, maxt string, p progressDecl, r relyDecl) {` | tick.go:274 | |
| `progressRelyAliases(p progressDecl, r relyDecl) map[string]goivy.Expr {` | tick.go:334 | |
| `progressExprArgs(expr goivy.Expr) []goivy.Expr {` | tick.go:348 | |
| `extraRelyVars(rhs goivy.Expr, aliases map[string]goivy.Expr) []*goivy.LogicVariable {` | tick.go:355 | |
| `(g *Generator) openProgressLoops(w *cppWriter, p progressDecl) int {` | tick.go:366 | |
| `(g *Generator) progressCounterLValue(p progressDecl) string {` | tick.go:383 | |
| `(g *Generator) progressCounterLValueWithObj(p progressDecl, obj string) string {` | tick.go:387 | |
| `cppFunctionType(s *goivy.LogicFunctionSort) string {` | types.go:110 | |
| `cppQualifiedType(s goivy.Sort, className string) string {` | types.go:114 | |
| `cppQualifiedFunctionType(s *goivy.LogicFunctionSort, className string) string {` | types.go:121 | |
| `(g *Generator) cppQualifiedType(s goivy.Sort, className string) string {` | types.go:129 | |
| `cppArraySuffix(dims []int) string {` | types.go:190 | |
| `cppCTupleName(domain []goivy.Sort, className string) string {` | types.go:200 | |
| `cppCTupleNameWith(g *Generator, domain []goivy.Sort, className string) string {` | types.go:204 | |
| `cppCTupleLocalName(domain []goivy.Sort) string {` | types.go:224 | |
| `cppCTupleLocalNameWith(g *Generator, domain []goivy.Sort) string {` | types.go:228 | |
| `cppHashType(g *Generator, s goivy.Sort) string {` | types.go:241 | |
| `numericRangeBoundsInt(rs *goivy.RangeSort) (int, int, bool) {` | types.go:309 | |
| `cppZeroValueInScope(s goivy.Sort, className string) string {` | types.go:345 | |
| `(g *Generator) cppZeroValueInScope(s goivy.Sort) string {` | types.go:395 | |
| `(g *Generator) nativeTypeName(s goivy.Sort, className string) (string, bool) {` | types.go:441 | |
| `(g *Generator) sortInterpString(s goivy.Sort) (string, bool) {` | types.go:459 | |
| `(g *Generator) hasStringInterp(s goivy.Sort) bool {` | types.go:471 | |
| `(g *Generator) hasNatInterp(s goivy.Sort) bool {` | types.go:476 | |
| `(g *Generator) cppStorageDecl(name string, s goivy.Sort, className string) string {` | types.go:539 | |
| `(g *Generator) cppFunctionStorageDecl(name string, domain []goivy.Sort, rng goivy.Sort, className string) string {` | types.go:546 | |
| `(g *Generator) cppStorageParamDecl(c *goivy.Const, className string) string {` | types.go:556 | |
| `(g *Generator) cppDestructorFieldAccess(field string, dom []goivy.Sort, rng goivy.Sort, args []string, obj string) string {` | types.go:567 | |
| `(g *Generator) cppStorageAccessBase(base string, sort goivy.Sort, args []string) string {` | types.go:597 | |
| `cppIndexSuffix(args []string) string {` | types.go:616 | |
| `(g *Generator) allHashThunkDomains() []string {` | types.go:630 | |
| `(g *Generator) cppCTuples() [][]goivy.Sort {` | types.go:671 | |
| `(g *Generator) variantIndex(super, sub goivy.Sort) int {` | variant.go:11 | |
| `(g *Generator) variantIsaExpr(superExpr string, super, sub goivy.Sort) string {` | variant.go:18 | |
| `(g *Generator) variantClassName(className string) string {` | variant.go:22 | |
| `(g *Generator) variantDowncastExpr(superExpr string, super, sub goivy.Sort, className string) string {` | variant.go:29 | |
| `(g *Generator) variantUpcastExpr(super, sub goivy.Sort, expr string, className string) string {` | variant.go:36 | |
| `(g *Generator) maybeVariantUpcast(target, value goivy.Sort, expr string, className string) string {` | variant.go:44 | |
| `variantSolverRelationName(super, sub goivy.Sort) string {` | variant.go:51 | |
| `(g *Generator) emitPythonTestVariantConstraintAdd(w *cppWriter, smt string) bool {` | variant.go:55 | |
| `(g *Generator) emitVariantWrapperDecl(w *cppWriter, name string) {` | variant.go:106 | |
| `(g *Generator) emitVariantImpls(w *cppWriter) {` | variant.go:171 | |
| `(g *Generator) emitVariantEqualityForwardDecls(w *cppWriter) {` | variant.go:183 | |
| `(g *Generator) emitVariantEqualityInlines(w *cppWriter) {` | variant.go:196 | |
| `(g *Generator) emitVariantImpl(w *cppWriter, name string) {` | variant.go:210 | |
| `(g *Generator) emitVariantSubArgImpl(w *cppWriter, sub goivy.Sort) {` | variant.go:222 | |
| `(g *Generator) emitVariantEqualityImpl(w *cppWriter, super goivy.Sort, typeName string) {` | variant.go:232 | |
| `(g *Generator) emitVariantStreamImpl(w *cppWriter, super goivy.Sort, typeName string) {` | variant.go:247 | |
| `(g *Generator) emitVariantArgImpl(w *cppWriter, super goivy.Sort, typeName string) {` | variant.go:262 | |
| `(g *Generator) emitVariantSerImpl(w *cppWriter, super goivy.Sort, typeName string) {` | variant.go:286 | |
| `(g *Generator) emitVariantDeserImpl(w *cppWriter, super goivy.Sort, typeName string) {` | variant.go:296 | |
| `(g *Generator) emitVariantZ3Impl(w *cppWriter, super goivy.Sort, typeName string) {` | variant.go:315 | |
| `init() {` | vprint.go:31 | |
| `nice(tm time.Time) string {` | vprint.go:62 | |
| `nice9(tm time.Time) string {` | vprint.go:65 | |
| `pp(format string, a ...interface{}) {` | vprint.go:69 | |
| `zz(format string, a ...interface{}) {}` | vprint.go:75 | |
| `vv(format string, a ...interface{}) {` | vprint.go:80 | |
| `alwaysPrintf(format string, a ...interface{}) {` | vprint.go:86 | |
| `tsPrintf(format string, a ...interface{}) {` | vprint.go:93 | |
| `ts() string {` | vprint.go:109 | |
| `printf(format string, a ...interface{}) (n int, err error) {` | vprint.go:121 | |
| `fileLine(depth int) string {` | vprint.go:125 | |
| `p(format string, a ...interface{}) {` | vprint.go:136 | |
| `caller(upStack int) string {` | vprint.go:142 | |
| `panicOn(err error) {` | vprint.go:165 | |
| `stack() string {` | vprint.go:172 | |
| `allstacks() string {` | vprint.go:177 | |
| `isNil(face interface{}) bool {` | vprint.go:190 | |
| `dirExists(name string) bool {` | vprint.go:201 | |
| `thisStack() []byte {` | vprint.go:209 | |
| `goroNumber() int {` | vprint.go:217 | |
| `stop(msg interface{}) {` | vprint.go:250 | |
| `stopOn(err error) {` | vprint.go:261 | |
| `panicf(format string, a ...interface{}) {` | vprint.go:269 | |
| `assert(b bool) {` | vprint.go:273 | |
| `getBinaryVersion() string {` | vprint.go:279 | |
| `newCPPWriter(text *CppText) cppWriter {` | writer.go:14 | |
| `(g *Generator) emitZ3SolverTemplates(w *cppWriter) {` | z3.go:32 | |
| `(g *Generator) emitZ3SolverConversions(w *cppWriter) {` | z3.go:84 | |
| `(g *Generator) emitZ3EnumSolverConversion(w *cppWriter, s *goivy.LogicEnumeratedSort) {` | z3.go:107 | |
| `(g *Generator) emitZ3RandomValueHelpers(w *cppWriter) {` | z3.go:146 | |
| `(g *Generator) emitZ3EnumRandomHelper(w *cppWriter, s *goivy.LogicEnumeratedSort) {` | z3.go:183 | |
| `(g *Generator) emitZ3VariantRandomHelper(w *cppWriter, s goivy.Sort) {` | z3.go:199 | |
| `(g *Generator) emitZ3CPPInterpRandomHelper(w *cppWriter, s goivy.Sort, it cppInterpType) {` | z3.go:234 | |
| `(g *Generator) emitZ3NumericRandomHelper(w *cppWriter, s goivy.Sort) {` | z3.go:260 | |
| `(g *Generator) emitZ3RangeRandomHelper(w *cppWriter, s goivy.Sort, rs *goivy.RangeSort) {` | z3.go:271 | |
| `(g *Generator) emitZ3Setup(w *cppWriter) {` | z3.go:289 | |
| `(g *Generator) emitZ3SortRegistrations(w *cppWriter) {` | z3.go:311 | |
| `(g *Generator) emitZ3DeclRegistrations(w *cppWriter) {` | z3.go:371 | |
| `(g *Generator) z3DeclSymbols() []stateSymbol {` | z3.go:382 | |
| `(g *Generator) emitZ3Randomize(w *cppWriter) {` | z3.go:422 | |
| `(g *Generator) emitZ3RandomizeSymbol(w *cppWriter, sym stateSymbol) {` | z3.go:436 | |
| `(g *Generator) z3RandomValueExpr(s goivy.Sort) (string, bool) {` | z3.go:475 | |
| `(g *Generator) z3RandomValueExprFrom(s goivy.Sort, genExpr string) (string, bool) {` | z3.go:479 | |
| `(g *Generator) z3LoopHeaderForSort(s goivy.Sort, name string) (string, bool) {` | z3.go:523 | |
| `finiteValuesInClassScope(s goivy.Sort, className string) ([]string, bool) {` | z3.go:543 | |
| `z3DeclSignature(s goivy.Sort) ([]string, string) {` | z3.go:561 | |
| `z3RandomHelperName(s goivy.Sort) string {` | z3.go:573 | |
| `(g *Generator) emitZ3GeneratorClasses(w *cppWriter) error {` | z3.go:581 | |
| `(g *Generator) actionGeneratorClassName(name string) string {` | z3.go:666 | |
| `(g *Generator) emitPythonTestZ3GeneratorClasses(w *cppWriter) error {` | z3.go:674 | |
| `(g *Generator) emitPythonTestInitGen(w *cppWriter) error {` | z3.go:701 | |
| `(g *Generator) emitVariantPrepares(w *cppWriter) {` | z3.go:742 | |
| `(g *Generator) emitVariantCleanups(w *cppWriter) {` | z3.go:755 | |
| `(g *Generator) emitPythonTestActionGenClassHeader(w *cppWriter, plan *actionGenPlan) {` | z3.go:768 | |
| `(g *Generator) emitPythonTestZ3Sig(w *cppWriter, extra []*goivy.Const) {` | z3.go:781 | |
| `(g *Generator) emitPythonTestZ3SortRegistrations(w *cppWriter, extra []*goivy.Const) {` | z3.go:788 | |
| `(g *Generator) emitPythonTestZ3SortRegistration(w *cppWriter, name string, s goivy.Sort) {` | z3.go:858 | |
| `(g *Generator) pythonTestZ3SigSymbols() []stateSymbol {` | z3.go:890 | |
| `(g *Generator) emitPythonTestDeclSolver(w *cppWriter, sym stateSymbol, symNameExpr string) {` | z3.go:926 | |
| `(g *Generator) emitPythonTestInitialConstraint(w *cppWriter) error {` | z3.go:947 | |
| `z3SortName(s goivy.Sort) string {` | z3.go:972 | |

## INVENTED — ivy2go functions absent from ivy2cpp

These functions exist in ivy2go but not in ivy2cpp. Each should be classified as:

- `delete` — Go-specific helper that should be removed; call sites should use ivy2cpp-equivalent logic
- `rename` — same logic as an ivy2cpp function but under a different name; unify the name
- `legitimate` — genuinely Go-specific scaffolding with no C++ analogue; justify why

| ivy2go signature | source file:line | verdict | notes |
| --- | --- | --- | --- |
| `(g *Generator) actionParamName(c *goivy.Const) string {` | action.go:558 | legitimate | Enforces Go keyword + receiver-collision avoidance; no C++ analogue. |
| `(g *Generator) isImportCaller(name string) bool {` | action.go:585 | delete | Thin wrapper over importCallers() map; cpp uses the map directly at call sites. |
| `(g *Generator) emitActionTraceLine(w *goWriter, dir, name string, params []*goivy.Const) {` | action.go:651 | rename | Maps to ivy2cpp emitTraceActionPrologue; rename for clarity. |
| `(g *Generator) emitActionMethods(w *goWriter) {` | action.go:694 | rename | Mirrors ivy2cpp emitMethods; rename to emitMethods. |
| `(g *Generator) emitActionMethod(w *goWriter, name string, act goivy.Action) {` | action.go:715 | rename | Mirrors ivy2cpp emitSomeAction; rename to emitSomeAction. |
| `(g *Generator) emitActionGenStructs(w *goWriter) {` | action_gen.go:143 | legitimate | Per-plan dispatch loop; cpp emits class+constructor inline in emitActionGen. |
| `(g *Generator) emitActionGenStructDecl(w *goWriter, plan *actionGenPlan) {` | action_gen.go:166 | rename | Phase-A split of cpp emitActionGenClassHeader + emitActionGenMemberDecls. |
| `(g *Generator) emitActionGenConstructor(w *goWriter, plan *actionGenPlan) {` | action_gen.go:195 | rename | Phase-A split of cpp emitActionGen's inline constructor block. |
| `(g *Generator) emitActionGenGenerate(w *goWriter, plan *actionGenPlan) {` | action_gen.go:208 | rename | Phase-A split of cpp emitActionGen's generate() body. |
| `(g *Generator) emitActionGenClose(w *goWriter, plan *actionGenPlan) {` | action_gen.go:356 | rename | Phase-A split of cpp emitActionGen's destructor (RAII); Go is explicit Close(). |
| `(g *Generator) emitCloseSolver(w *goWriter) { _ = w }` | action_gen.go:371 | delete | No-op stub; per-plan Close() now lives in emitActionGenClose. |
| `(g *Generator) emitInputExtraction(w *goWriter, i int, p *goivy.Const) {` | action_gen.go:375 | legitimate | Emits Go pickBool/pickUint glue for goivy.Solver model extraction. |
| `(g *Generator) emitFallbackInputAssignments(w *goWriter, plan *actionGenPlan) {` | action_gen.go:404 | rename | Mirrors cpp emitWeakActionGenerator's fallback path. |
| `goActionGenFieldName(paramName string) string {` | action_gen.go:428 | legitimate | Builds `In_<Exported>` Go field name avoiding method/field clash. |
| `(g *Generator) actionGenNames() []string {` | action_gen.go:438 | legitimate | Public-action selector for Go test main; cpp inlines similar selection. |
| `(g *Generator) reifyExprAsGoCode(e goivy.Expr, params []*goivy.Const) (string, bool) {` | action_gen.go:477 | legitimate | Serialises Expr tree to Go construction code; cpp uses SMT-LIB strings. |
| `(g *Generator) reifyVariableSlice(vars []*goivy.LogicVariable) (string, bool) {` | action_gen.go:633 | legitimate | Companion of reifyExprAsGoCode; no cpp analogue (SMT-LIB path). |
| `(g *Generator) reifySortAsGoCode(s goivy.Sort) (string, bool) {` | action_gen.go:653 | legitimate | Companion of reifyExprAsGoCode; cpp uses goivy.Translator at runtime. |
| `(g *Generator) reifyBoundAsGoCode(b goivy.NumeralOrCompiledBound) (string, bool) {` | action_gen.go:692 | legitimate | Companion of reifyExprAsGoCode for RangeSort bounds. |
| `(g *Generator) requireMustHelpers() {` | action_gen.go:714 | legitimate | Marks runtime to emit must* facades for goivy ctor error-plumbing. |
| `(g *Generator) emitStructInputAssembly(w *goWriter, i int, typeName, recName string) {` | action_gen.go:723 | legitimate | Per-field struct literal assembly for destructor-record params (Go syntax). |
| `(g *Generator) emitVariantInputAssembly(w *goWriter, i int, typeName, superName string) {` | action_gen.go:748 | legitimate | Tag-pick + per-leaf ctor invocation for variants (Go switch syntax). |
| `(g *Generator) canOpenAssignmentLoops(vs []*goivy.LogicVariable) bool {` | assign.go:111 | rename | Mirrors cpp canOpenAssignmentLoopsBounded; align name. |
| `(g *Generator) emitAssignmentLoopsBody(w *goWriter, vs []*goivy.LogicVariable, rhsOrLHS goivy.Expr, tmpName string, intoTemp bool) {` | assign.go:167 | legitimate | Inner-loop body emitter for two-phase assign; cpp inlines into openAssignmentLoopsBounded. |
| `(g *Generator) tempKeyExpression(vs []*goivy.LogicVariable, tmpName string) string {` | assign.go:217 | legitimate | Builds Go indexed-access expression (array vs tuple map key); cpp inlines. |
| `(g *Generator) emitTraceWrite(w *goWriter, lhs goivy.Expr, rhsCode string) {` | assign.go:252 | delete | Trivial shim around emitTracedLHS; collapse callers. |
| `checkBuildableBinary(out *Output) error {` | build.go:118 | legitimate | Go-specific preflight: `func main()` requires `package main`. |
| `checkBuildContext(pkgDir string, out *Output) error {` | build.go:135 | legitimate | Go-specific preflight: output dir must live under a go.mod. |
| `findEnclosingGoMod(dir string) string {` | build.go:149 | legitimate | Walk for go.mod; no C++ counterpart (no module system). |
| `isWindows() bool {` | build.go:170 | legitimate | Adds .exe suffix on Windows; cpp uses build_findvs_windows.go. |
| `bvMask(bits int) string {` | bv_expr.go:305 | rename | Mirrors cpp bvMask in bv_expr.go; same name, listed because cpp also has it. |
| `(g *Generator) requireUint128() {` | bv_expr.go:368 | legitimate | Marks runtime to emit Uint128 helpers; cpp uses `unsigned __int128`. |
| `(g *Generator) requireBigInt() {` | bv_expr.go:378 | legitimate | Marks runtime to emit math/big helpers; cpp uses ivy_uint<N> templates. |
| `(g *Generator) emitWideBVApply(name string, a *goivy.Apply, result goInterpType) (string, bool, error) {` | bv_expr.go:393 | legitimate | Routes wide BV through math/big; cpp routes through ivy_uint<N> template. |
| `(g *Generator) emitWideBVShift(name string, a *goivy.Apply, result goInterpType, method string) (string, bool, error) {` | bv_expr.go:467 | legitimate | Companion of emitWideBVApply; math/big shift dispatch. |
| `wideBVMethodFor(name string) string {` | bv_expr.go:484 | legitimate | Maps op name to Go method name on *big.Int; cpp uses operator overloads. |
| `wideBVAsCall(expr string) string {` | bv_expr.go:508 | legitimate | Wraps expr in toBigInt(); cpp uses implicit conversion. |
| `prepareModuleForGo(mod *goivy.Module, cfg Config) {` | compile.go:292 | rename | Mirrors cpp prepareModuleForCPP; rename target follows package convention. |
| `ensureSortOrderForGo(mod *goivy.Module) {` | compile.go:319 | rename | Mirrors cpp ensureSortOrderForCPP; rename target follows package convention. |
| `(g *Generator) emitDestructorHash(w *goWriter, typeName string, destrs []*goivy.Const) {` | destructor.go:108 | rename | Mirrors cpp emitDestructorStructHash; rename to align. |
| `(g *Generator) emitDestructorLess(w *goWriter, typeName string, destrs []*goivy.Const) {` | destructor.go:167 | legitimate | Go Less() method for stable sorts; cpp uses operator< via emitDestructorStructComparators. |
| `arrayIndexLoops(dims []int) (open, close []string) {` | destructor.go:209 | legitimate | Emits Go for-loop headers + braces; cpp inlines similar shape inline. |
| `arrayIndexSuffix(dims []int) string {` | destructor.go:219 | legitimate | Builds `[__i0][__i1]…` indexing suffix; Go-array-specific helper. |
| `(g *Generator) destructorFieldType(domain []goivy.Sort, rng goivy.Sort) string {` | destructor.go:231 | legitimate | Resolves Go field type (array/map vs scalar); cpp uses cppStorageDecl. |
| `(g *Generator) destructorScalarFields(sortName string) []destructorField {` | destructor.go:254 | legitimate | Scalar-field enumeration for action-gen struct synthesis; cpp uses inline SMT path. |
| `(g *Generator) emitDestructorEqual(w *goWriter, typeName string, destrs []*goivy.Const) {` | destructor.go:291 | rename | Mirrors cpp emitDestructorStructComparators (==); rename to align. |
| `(g *Generator) isStateSymbolName(name string) bool {` | expr.go:388 | legitimate | Tells emitExpr to route through `s.<exported>`; Go requires receiver prefix. |
| `(g *Generator) isEnumConstantName(name string) bool {` | expr.go:401 | legitimate | Tells emitExpr to use PascalCase Go constant; cpp uses scoped enum. |
| `(g *Generator) requestIteHelper(s goivy.Sort) string {` | expr.go:422 | legitimate | Requests per-type ite_<T> helper; cpp uses C++ ternary directly. |
| `iteHelperSuffix(typ string) string {` | expr.go:435 | legitimate | Sanitises Go type expression for use in helper name. |
| `(g *Generator) emitExtensionalDecls(w *goWriter) {` | extensional.go:17 | legitimate | Stream-hook stub (M4 placeholder); aligns with stream-based Go file split. |
| `(g *Generator) emitTypes() {` | generator.go:226 | rename | Phase-A stream emitter; per-file split of cpp emitHeader's type block. |
| `(g *Generator) emitState() {` | generator.go:231 | rename | Phase-A stream emitter; per-file split of cpp emitHeader's state class. |
| `(g *Generator) emitActions() {` | generator.go:236 | rename | Phase-A stream emitter; per-file split of cpp emitImpl's action methods. |
| `(g *Generator) emitRuntime() {` | generator.go:246 | rename | Phase-A stream emitter; per-file split of cpp emitImpl's runtime helpers. |
| `(g *Generator) emitNondet() {` | generator.go:252 | rename | Phase-A stream emitter for nondet helpers; cpp inlines into emitImpl. |
| `(g *Generator) emitExtensional() {` | generator.go:256 | rename | Phase-A stream emitter for extensional rels; cpp inlines into emitImpl. |
| `(g *Generator) emitThunks() {` | generator.go:264 | rename | Phase-A stream emitter for thunk decls; cpp inlines into emitImpl. |
| `(g *Generator) emitNative() {` | generator.go:268 | rename | Phase-A stream emitter for native blocks; cpp uses emitHeaderNatives/emitImplNatives. |
| `(g *Generator) emitRepl() {` | generator.go:272 | rename | Phase-A stream emitter for REPL; cpp uses emitReplSupport. |
| `(g *Generator) emitMain() {` | generator.go:279 | rename | Phase-A stream emitter for main(); cpp uses emitGenMain/emitReplMain. |
| `(g *Generator) finalize() (map[string]string, error) {` | generator.go:407 | legitimate | Composes per-stream Go files; cpp writes header+impl only via emitHeader/emitImpl. |
| `(g *Generator) emitInitMethod(w *goWriter) {` | init.go:15 | rename | Mirrors cpp emitInit; rename to emitInit. |
| `(g *Generator) emitAfterInitActions(w *goWriter) {` | init.go:30 | rename | Mirrors cpp emitInit's InitialActions loop; extract for parity. |
| `(g *Generator) emitInitialState(w *goWriter) {` | initial_state.go:22 | rename | Mirrors cpp emitOneInitialState / emitDefaultInitialState. |
| `(g *Generator) emitScalarChoice(w *goWriter, sym stateSymbol) {` | initial_state.go:48 | rename | Mirrors cpp emitDefaultInitialState's per-sym mkNondet path. |
| `isFunctionSort(s goivy.Sort) bool {` | initial_state.go:68 | legitimate | Local type-assertion helper; trivial, no cpp counterpart needed. |
| `goIdent(name string) string {` | names.go:112 | legitimate | Enforces Go keyword/predeclared-name avoidance; cpp has no equivalent restriction. |
| `goExportedName(name string) string {` | names.go:125 | legitimate | Enforces Go PascalCase export rule; cpp has no exported/unexported split. |
| `goPackageName(base string) string {` | names.go:154 | legitimate | Enforces Go package-name conventions; cpp has no package concept. |
| `upperFirst(s string) string {` | names.go:168 | legitimate | Local string helper used by goExportedName; trivial. |
| `(g *Generator) emitNativeBlocks() {` | native.go:27 | rename | Mirrors cpp emit*Natives family; rename for clarity. |
| `(g *Generator) collectNativeGoBlocks() []nativeGoBlock {` | native.go:38 | rename | Mirrors cpp buildNativeBlocks; filters for "go*" tags instead of cpp tags. |
| `(g *Generator) renderNativeGoTemplate(body string, params []goivy.Expr) string {` | native.go:89 | rename | Mirrors cpp renderNativeTemplate; antiquote substitution identical. |
| `strconvAtoiSafe(s string) (int, error) {` | native.go:138 | delete | Local digit parser; replace with strconv.Atoi (no leading-empty quirk needed). |
| `splitNativeGoCode(code string) (string, string) {` | native.go:157 | rename | Mirrors cpp splitNativeCode; rename to align. |
| `(g *Generator) dispatchNativeGoBlock(blk nativeGoBlock) {` | native.go:185 | legitimate | Routes go/go_header/go_init/go_inline tags to Go streams; cpp tag set differs. |
| `(g *Generator) emitNativeThunkDecls(w *goWriter) {` | native_thunk.go:9 | legitimate | Stream-hook stub for M8/M10; matches Go file-split architecture. |
| `(g *Generator) emitNondetDecls(w *goWriter) {` | nondet.go:8 | legitimate | Stream-hook stub for M5; matches Go file-split architecture. |
| `CompareGoVsCpp(goBin, cppBin, stdin string) error {` | oracle_compare.go:29 | legitimate | Tier-4 cross-language oracle harness; cpp's oracle_compare.go is for tokens. |
| `runForOracle(bin, stdin string) (string, error) {` | oracle_compare.go:46 | legitimate | Helper for CompareGoVsCpp; runs a binary with stdin. |
| `diffNormalised(goOut, cppOut string) error {` | oracle_compare.go:74 | legitimate | Helper for CompareGoVsCpp; line-diff with normalisation. |
| `splitNormalisedLines(s string) []string {` | oracle_compare.go:96 | legitimate | Helper for diffNormalised. |
| `sideOrEmpty(lines []string, i int) string {` | oracle_compare.go:108 | legitimate | Helper for diffNormalised's tail-mismatch report. |
| `(g *Generator) emitReplLoop(w *goWriter) {` | repl.go:20 | rename | Mirrors cpp emitCmdReader / emitReplSupport family. |
| `(g *Generator) emitReplDispatch(w *goWriter) {` | repl.go:65 | rename | Mirrors cpp emitCmdReaderDispatchChain. |
| `(g *Generator) emitReplArgParsers(w *goWriter) {` | repl.go:122 | legitimate | Emits per-sort Go arg parsers; cpp uses C++ stream operators for parsing. |
| `(g *Generator) replParserName(s goivy.Sort) string {` | repl.go:145 | legitimate | Stable Go parser fn name per sort; cpp dispatches through stream operators. |
| `(g *Generator) emitOneReplArgParser(w *goWriter, parser string, s goivy.Sort) {` | repl.go:152 | legitimate | Emits one Go parseArg_<T> function; cpp uses operator>>. |
| `(g *Generator) replActionNames() []string {` | repl.go:196 | rename | Mirrors cpp publicActionNamesSorted; rename to align. |
| `hasPrefixAny(s string, prefixes ...string) bool {` | repl.go:214 | legitimate | Trivial string helper; cpp uses inline checks. |
| `joinComma(parts []string) string {` | repl.go:225 | delete | Reimplements strings.Join(parts, ", "); replace with stdlib call. |
| `(g *Generator) emitRuntimeHelpers(w *goWriter) {` | runtime.go:23 | rename | Mirrors cpp emitRuntimeMethods / emitRuntimeImplPreamble role. |
| `(g *Generator) emitRuntimeHelpersLate(w *goWriter) {` | runtime.go:32 | legitimate | Second-pass conditional helper emission; needed because Go has no .hpp/.cpp split. |
| `(g *Generator) emitRuntimePreamble(w *goWriter) {` | runtime.go:67 | rename | Mirrors cpp emitRuntimeImplPreamble. |
| `(g *Generator) emitUint128Helpers(w *goWriter) {` | runtime.go:127 | legitimate | Emits Uint128 struct + methods; cpp uses `unsigned __int128` intrinsic. |
| `(g *Generator) emitBigIntHelpers(w *goWriter) {` | runtime.go:218 | legitimate | Emits math/big BV helpers; cpp uses ivy_uint<N> templated header. |
| `(g *Generator) collectIteHelpers() []iteHelperReq {` | runtime.go:344 | legitimate | Collects ite_<T> requests from OnceGlobals; cpp uses ternary inline. |
| `iteHelperTypeFromSuffix(suffix string) string {` | runtime.go:365 | legitimate | Inverse of iteHelperSuffix; reconstructs Go type from helper name. |
| `(g *Generator) emitMixHashHelper(w *goWriter) {` | runtime.go:376 | legitimate | Emits runtime mixHash; cpp uses std::hash combine in headers. |
| `(g *Generator) emitLessOrdHelper(w *goWriter) {` | runtime.go:410 | legitimate | Emits runtime lessOrd; cpp uses operator< directly. |
| `(g *Generator) emitPickInputHelpers(w *goWriter) {` | runtime.go:444 | legitimate | Emits goivy.Solver model-extraction glue; cpp uses Z3 C++ API directly. |
| `(g *Generator) emitTestFlagsHelper(w *goWriter) {` | runtime.go:485 | rename | Mirrors cpp test_iters argv plumbing; rename for parity. |
| `(g *Generator) emitMustHelpers(w *goWriter) {` | runtime.go:510 | legitimate | Emits must* facades over goivy ctors that return (T, error); Go-specific. |
| `(g *Generator) emitIteHelper(w *goWriter, req iteHelperReq) {` | runtime.go:547 | legitimate | Per-type ite_<T> helper emission; cpp uses ternary. |
| `(g *Generator) emitStateSymbolFacts(w *goWriter, sym stateSymbol) {` | solver_emit.go:89 | legitimate | Emits stateFactsAsClauses body for goivy.Solver; cpp asserts via SMT-LIB add(). |
| `cellNameExpr(sym string, arity int) string {` | solver_emit.go:155 | legitimate | Builds Go expression for per-cell symbol name; cpp builds the C++ string inline. |
| `(g *Generator) emitPreconditionHelpers(w *goWriter) {` | solver_emit.go:172 | legitimate | Emits mkAnd/conjClauses Go helpers; cpp uses goivy.Solver via Z3 C++ directly. |
| `(g *Generator) emitTickMethod(w *goWriter) {` | tick.go:10 | rename | Mirrors cpp emitTick (currently stub); rename to emitTick. |
| `(g *Generator) goScalarType(s goivy.Sort) string {` | types.go:72 | rename | Mirrors cpp cppScalarTypeWith; rename to scalarType for parity. |
| `goArrayPrefix(dims []int) string {` | types.go:145 | legitimate | Go puts dims before type name; cpp's cppArraySuffix puts them after. |
| `(g *Generator) emitEnumDecl(w *goWriter, st *goivy.LogicEnumeratedSort) {` | types.go:478 | legitimate | Emits Go `type T int` + `const (... iota)`; cpp uses `enum class`. |
| `(g *Generator) emitRangeDecl(w *goWriter, st *goivy.RangeSort) {` | types.go:502 | legitimate | Emits Go named-int alias for a RangeSort; cpp uses typedef. |
| `(g *Generator) emitRangeDeclNamed(w *goWriter, name string, rs *goivy.RangeSort) {` | types.go:510 | legitimate | Shared helper for RangeSort + range-interpreted UninterpretedSort. |
| `(g *Generator) emitGoTypeDecl(w *goWriter, s goivy.Sort, it goInterpType) {` | types.go:526 | legitimate | Emits Go struct/alias for interpreted sort; cpp uses cppInterpType lowering. |
| `(g *Generator) variantLeaves(superName string) []variantLeafInfo {` | variant.go:172 | legitimate | Variant leaf metadata for input-synthesis; cpp inlines in solver_emit path. |
| `(g *Generator) traceFormatFor(lhs goivy.Expr, rhsCode string) (string, []string) {` | vprint.go:69 | legitimate | Builds Go fmt.Fprintf format+args; cpp uses ostream chaining. |
| `escapeFmt(s string) string {` | vprint.go:125 | legitimate | Escapes `%` for Go fmt strings; cpp has no analogue (ostream is type-driven). |
| `newGoWriter(text *GoText) goWriter {` | writer.go:19 | legitimate | Go-side counterpart of cpp newCPPWriter; different writer state. |
| `(w *goWriter) closeParen() {` | writer.go:69 | legitimate | Emits closing `)` for Go's parenthesised blocks (import/const/var/type). |

## Summary

Counts: OMITTED=510, INVENTED=120.

## Verdict counts

- delete=5  ← **all drained 2026-05-25** (isImportCaller, emitCloseSolver, emitTraceWrite, strconvAtoiSafe, joinComma)
- rename=14 done / 26 reclassified to legitimate (see Status 2026-05-26 below) — total 40 originally
- legitimate=75 + 26 reclassified = 101 (no further action needed)

## Status (2026-05-25)

**Phase A (action_gen.go literal port)** — completed.
- Deleted `wouldFail_<Name>` / `buildPrecondition_<Name>` static-gate inventions.
- Rebuilt `actionGenPlan` / `buildActionGenPlan` / per-method emitters to mirror ivy2cpp.
- Solver is now the single source of truth for fire/skip decisions.
- Fixed `stateSymbols()` to exclude enum constants / constructors / definitions (mirrors ivy2cpp filter).
- Fixed `stateFactsAsClauses` to pin enum-sorted state symbols (was silently leaving them as free solver vars, forcing every precondition UNSAT after the first state mutation).
- Verified: pingpong fires hit/ping/pong correctly over 1000 iterations with zero double-pongs and zero ivyAssume panics.

**Phase B (this audit)** — completed.
- 28 common files compared; 510 OMITTED + 120 INVENTED rows produced.
- 120 INVENTED rows classified (5 delete / 40 rename / 75 legitimate).

**Phase C (drain queue)** — partially complete.
- All 5 `delete` rows drained.
- The 40 `rename` items are pending. Most are trivial mechanical renames (emitActionMethods→emitMethods, emitActionMethod→emitSomeAction, emitActionTraceLine→emitTraceActionPrologue, etc.) but draining requires per-row work plus a test update where callers reference the old name.
- The 510 OMITTED rows are NOT a single-session task. Many represent significant ivy2cpp subsystems (assign.go's openAssignmentLoops family, destructor.go's 35+ functions, generator.go's 60+ functions). Drain these incrementally, prioritising by user-visible regression risk.

## Status (2026-05-26): rename batch 1 + reclassification

**14 renames landed** (function moved to its ivy2cpp name; callers + tests updated):

| ivy2go (old)              | ivy2cpp (target, now used in ivy2go) |
| ---                       | ---                                  |
| emitActionTraceLine       | emitTraceActionPrologue              |
| emitActionMethods         | emitMethods                          |
| emitActionMethod          | emitSomeAction                       |
| emitDestructorHash        | emitDestructorStructHash             |
| emitDestructorEqual       | emitDestructorStructEqual            |
| emitInitMethod            | emitInit                             |
| emitInitialState          | emitOneInitialState                  |
| emitRuntimePreamble       | emitRuntimeImplPreamble              |
| emitTickMethod            | emitTick                             |
| collectNativeGoBlocks     | buildNativeBlocks                    |
| renderNativeGoTemplate    | renderNativeTemplate                 |
| splitNativeGoCode         | splitNativeCode                      |
| emitReplDispatch          | emitCmdReaderDispatchChain           |
| replActionNames           | publicActionNamesSorted              |

Plus a collision-avoidance rename in generator.go: the Phase-A stream-orchestration wrapper `emitInit()` (no cpp 1:1) became `emitInitStream()` so the init.go method can carry the proper `emitInit` name.

**26 reclassified `rename` → `legitimate`** — review showed no clean cpp 1:1 mapping; each is either a Phase-A intentional decomposition or a Go-specific abstraction:

| ivy2go function                            | why legitimate, not rename                                                |
| ---                                        | ---                                                                       |
| (generator.go) emitTypes, emitState, emitActions, emitRuntime, emitNondet, emitExtensional, emitThunks, emitNative, emitRepl, emitMain | Phase-A per-file stream wrappers; cpp uses two-stream `emitHeader`/`emitImpl` so there is no 1:1 |
| (action_gen.go) emitActionGenStructDecl, emitActionGenConstructor, emitActionGenGenerate, emitActionGenClose, emitFallbackInputAssignments | Phase-A split of cpp's monolithic `emitActionGen` into per-method emitters; Go-idiomatic decomposition |
| (assign.go) canOpenAssignmentLoops | Different signature than cpp's `canOpenAssignmentLoopsBounded(lhs, body)`; takes a `[]*LogicVariable` directly. Distinct abstraction. |
| (bv_expr.go) bvMask | False positive: cpp has the same name. Spurious INVENTED row from the extraction pass. |
| (compile.go) prepareModuleForGo, ensureSortOrderForGo | Package-convention naming (`*ForGo` mirrors cpp's `*ForCPP`). Same role, intentional naming difference. |
| (init.go) emitAfterInitActions | Extracted from cpp `emitInit`'s `InitialActions` loop as a separate method for clarity. Go-idiomatic decomposition. |
| (initial_state.go) emitScalarChoice | Extracted from cpp's inline `mkNondetSym` invocation in `emitDefaultInitialState`. Go-idiomatic decomposition. |
| (native.go) emitNativeBlocks | Top-level distributor with no cpp 1:1 (cpp uses per-stream `emit*Natives` family). |
| (repl.go) emitReplLoop | Vague mapping; cpp uses `emitCmdReader` + `emitReplSupport`. Distinct shape. |
| (runtime.go) emitRuntimeHelpers, emitTestFlagsHelper | Vague mapping into cpp's runtime-emission family; no clean rename target. |
| (types.go) goScalarType | Returns Go type string; cpp `cppScalarTypeWith` takes more args and produces qualified C++ type. Different abstraction. |

**Verification after rename pass:** all ivy2go tests pass under `SLOW_GO_TEST=1`; pingpong 500-iter run produces clean trace with zero double-pongs.

**Remaining Phase C work:** 510 OMITTED rows — same scope as before this session, no progress yet.

## Suggested next-step ordering for Phase C continuation

1. **Rename pass** (40 items) — small, mechanical, mostly atomic. Each rename can be done with `sed` + a test re-run.
2. **OMITTED triage** — split the 510 into tiers:
   - Tier 1: functions whose absence demonstrably blocks a fixture (e.g. anything action.go missing for IfSome / Let / BindOlds / AssignField / CopyField — these surface as `unsupported` panics).
   - Tier 2: functions whose absence silently degrades the emitted code (e.g. destructor hash/equality helpers).
   - Tier 3: legitimately-deferred subsystems (REPL, build, vprint completeness).
3. Land each tier as its own focused change; update this document's status as items close.

This audit document remains the source of truth for Phase C work. Delete it only when the rename queue is empty AND OMITTED has been triaged into a separate plan.
