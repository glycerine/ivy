(goivy-venv) jaten@jbook ~/ivy/goivy (master) $ make test-web
XTRACE_OFF=1 go test -v ./webui -count=1 -tags xtrace_off,web
=== RUN   TestConformNewSession
2026/04/25 20:08:36 pybackend: starting sidecar on port 57001 (python=python3, root=/Users/jaten/ivy/pyivy/ivy)
ivy sidecar: listening on 127.0.0.1:57001
2026/04/25 20:08:36 CONFORM: session mapping s1 (Go) → s1 (Py)
    backend_conform_test.go:79: session_id: s1
--- PASS: TestConformNewSession (0.83s)
=== RUN   TestConformLoad
2026/04/25 20:08:36 pybackend: starting sidecar on port 57008 (python=python3, root=/Users/jaten/ivy/pyivy/ivy)
ivy sidecar: listening on 127.0.0.1:57008
    backend_conform_test.go:119: Go  Load: {"filename":"test.ivy","status":"ok"}
    backend_conform_test.go:120: Py  Load: {"filename":"test.ivy","status":"ok"}
--- PASS: TestConformLoad (0.45s)
=== RUN   TestConformConcept
2026/04/25 20:08:37 pybackend: starting sidecar on port 57014 (python=python3, root=/Users/jaten/ivy/pyivy/ivy)
ivy sidecar: listening on 127.0.0.1:57014
    backend_conform_test.go:208:   field "relations": sidecar returns bare names, Go matches Tcl/Tk GUI:
            Go: ["link(X,Y)","semaphore(X)"]
            Py sidecar: ["link","semaphore"]
--- PASS: TestConformConcept (0.44s)
=== RUN   TestConformCheck
2026/04/25 20:08:37 pybackend: starting sidecar on port 57021 (python=python3, root=/Users/jaten/ivy/pyivy/ivy)
ivy sidecar: listening on 127.0.0.1:57021
    backend_conform_test.go:255: Go  Check: {"failed_conjecture":"","failed_label":"","message":"Inductive invariant found:\n((link(X,Y) -> ~semaphore(Y)))","mode":"induction","result":"pass","status":"ok","used_relations":null}
    backend_conform_test.go:256: Py  Check: {"failed_conjecture":"","failed_label":"","message":"Inductive invariant found:\n((link(X,Y) -> ~semaphore(Y)))","mode":"induction","result":"pass","status":"ok","used_relations":null}
--- PASS: TestConformCheck (1.20s)
=== RUN   TestConformARG
2026/04/25 20:08:38 pybackend: starting sidecar on port 57028 (python=python3, root=/Users/jaten/ivy/pyivy/ivy)
ivy sidecar: listening on 127.0.0.1:57028
    backend_conform_test.go:321: Go  ARG: {"elements":[]}
    backend_conform_test.go:322: Py  ARG: {"elements":[]}
--- PASS: TestConformARG (0.44s)
=== RUN   TestBrowserPageLoads
=== PAUSE TestBrowserPageLoads
=== RUN   TestBrowserHasTitle
=== PAUSE TestBrowserHasTitle
=== RUN   TestBrowserHasARGPanel
=== PAUSE TestBrowserHasARGPanel
=== RUN   TestBrowserHasConceptPanel
=== PAUSE TestBrowserHasConceptPanel
=== RUN   TestBrowserHasMenuBar
=== PAUSE TestBrowserHasMenuBar
=== RUN   TestBrowserHasStatusBar
=== PAUSE TestBrowserHasStatusBar
=== RUN   TestBrowserHasInfoPanel
=== PAUSE TestBrowserHasInfoPanel
=== RUN   TestBrowserCreateSession
=== PAUSE TestBrowserCreateSession
=== RUN   TestBrowserGetARG
=== PAUSE TestBrowserGetARG
=== RUN   TestBrowserGetConcept
=== PAUSE TestBrowserGetConcept
=== RUN   TestBrowserModeSelect
=== PAUSE TestBrowserModeSelect
=== RUN   TestBrowserCheckButton
=== PAUSE TestBrowserCheckButton
=== RUN   TestBrowserUndoButton
=== PAUSE TestBrowserUndoButton
=== RUN   TestBrowserCytoscapeLoads
=== PAUSE TestBrowserCytoscapeLoads
=== RUN   TestBrowserARGGraphInitializes
=== PAUSE TestBrowserARGGraphInitializes
=== RUN   TestBrowserContextMenuHidden
=== PAUSE TestBrowserContextMenuHidden
=== RUN   TestBrowserRightClickShowsMenu
=== PAUSE TestBrowserRightClickShowsMenu
=== RUN   TestBrowserDividerExists
=== PAUSE TestBrowserDividerExists
=== RUN   TestBrowserPanelResize
=== PAUSE TestBrowserPanelResize
=== RUN   TestBrowserSSEEvents
=== PAUSE TestBrowserSSEEvents
=== RUN   TestBrowserInvalidSession
=== PAUSE TestBrowserInvalidSession
=== RUN   TestBrowserStaticCSS
=== PAUSE TestBrowserStaticCSS
=== RUN   TestBrowserGraphHealthCheck
=== PAUSE TestBrowserGraphHealthCheck
=== RUN   TestBrowserConceptGraphRightClick
=== PAUSE TestBrowserConceptGraphRightClick
=== RUN   TestBrowserStaticJS
=== PAUSE TestBrowserStaticJS
=== RUN   TestCDConceptCreation
--- PASS: TestCDConceptCreation (0.00s)
=== RUN   TestCDConceptCreationEmpty
--- PASS: TestCDConceptCreationEmpty (0.00s)
=== RUN   TestCDConceptCreationHigherOrder
--- PASS: TestCDConceptCreationHigherOrder (0.00s)
=== RUN   TestCDConceptArity
--- PASS: TestCDConceptArity (0.00s)
=== RUN   TestCDConceptSorts
--- PASS: TestCDConceptSorts (0.00s)
=== RUN   TestCDConceptSort
--- PASS: TestCDConceptSort (0.00s)
=== RUN   TestCDConceptCall
--- PASS: TestCDConceptCall (0.00s)
=== RUN   TestCDConceptCallWrongArity
--- PASS: TestCDConceptCallWrongArity (0.00s)
=== RUN   TestCDConceptCallNoArgs
--- PASS: TestCDConceptCallNoArgs (0.00s)
=== RUN   TestCDConceptString
--- PASS: TestCDConceptString (0.00s)
=== RUN   TestCDConceptFormulaStr
--- PASS: TestCDConceptFormulaStr (0.00s)
=== RUN   TestCDConceptCombinerCall
--- PASS: TestCDConceptCombinerCall (0.00s)
=== RUN   TestCDConceptCombinerArity
--- PASS: TestCDConceptCombinerArity (0.00s)
=== RUN   TestCDConceptCombinerArities
--- PASS: TestCDConceptCombinerArities (0.00s)
=== RUN   TestCDConceptCombinerCallWrongArity
--- PASS: TestCDConceptCombinerCallWrongArity (0.00s)
=== RUN   TestCDConceptCombinerString
--- PASS: TestCDConceptCombinerString (0.00s)
=== RUN   TestCDConceptCombinerCallConceptArityMismatch
--- PASS: TestCDConceptCombinerCallConceptArityMismatch (0.00s)
=== RUN   TestCDConceptDictBasic
--- PASS: TestCDConceptDictBasic (0.00s)
=== RUN   TestCDConceptDictList
--- PASS: TestCDConceptDictList (0.00s)
=== RUN   TestCDConceptDictAppendToList
--- PASS: TestCDConceptDictAppendToList (0.00s)
=== RUN   TestCDConceptDictDelete
--- PASS: TestCDConceptDictDelete (0.00s)
=== RUN   TestCDConceptDictCopy
--- PASS: TestCDConceptDictCopy (0.00s)
=== RUN   TestCDConceptDictOrder
--- PASS: TestCDConceptDictOrder (0.00s)
=== RUN   TestCDConceptDictReorder
--- PASS: TestCDConceptDictReorder (0.00s)
=== RUN   TestCDConceptDictForEachConcept
--- PASS: TestCDConceptDictForEachConcept (0.00s)
=== RUN   TestCDConceptDomainCopy
--- PASS: TestCDConceptDomainCopy (0.00s)
=== RUN   TestCDConceptDomainConceptsByArity
--- PASS: TestCDConceptDomainConceptsByArity (0.00s)
=== RUN   TestCDConceptDomainPossibleNodeLabels
--- PASS: TestCDConceptDomainPossibleNodeLabels (0.00s)
=== RUN   TestCDConceptDomainSplit
--- PASS: TestCDConceptDomainSplit (0.00s)
=== RUN   TestCDConceptDomainReplaceConcept
--- PASS: TestCDConceptDomainReplaceConcept (0.00s)
=== RUN   TestCDConceptDomainGetFacts
--- PASS: TestCDConceptDomainGetFacts (0.00s)
=== RUN   TestCDConceptDomainGetFactsEdgeInfo
--- PASS: TestCDConceptDomainGetFactsEdgeInfo (0.00s)
=== RUN   TestCDConceptDomainGetFactsNoProjection
--- PASS: TestCDConceptDomainGetFactsNoProjection (0.00s)
=== RUN   TestCDConceptDomainOutput
Concepts:
    both : Concept X:S. (p(X) & q(X))
    none : Concept X:S. (~p(X) & ~q(X))
    onlyp : Concept X:S. (p(X) & ~q(X))
    onlyq : Concept X:S. (~p(X) & q(X))
    r : Concept X:S, Y:S. r(X,Y)
    nodes : [both none onlyp onlyq]
    edges : [r]
    node_labels : []
Concept Combiners:
    none : ConceptCombiner U:TopSort -> Boolean. ~(exists X. U(X))
    at_least_one : ConceptCombiner U:TopSort -> Boolean. exists X. U(X)
    at_most_one : ConceptCombiner U:TopSort -> Boolean. forall X,Y. (U(X) & U(Y)) -> X = Y
    node_necessarily : ConceptCombiner U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X. U1(X) -> U2(X)
    node_necessarily_not : ConceptCombiner U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X. U1(X) -> ~U2(X)
    mutually_exclusive : ConceptCombiner U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X,Y. ~(U1(X) & U2(Y))
    all_to_all : ConceptCombiner B:TopSort * TopSort -> Boolean, U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X,Y. (U1(X) & U2(Y)) -> B(X,Y)
    none_to_none : ConceptCombiner B:TopSort * TopSort -> Boolean, U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X,Y. (U1(X) & U2(Y)) -> ~B(X,Y)
    total : ConceptCombiner B:TopSort * TopSort -> Boolean, U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X. U1(X) -> (exists Y. (U2(Y) & B(X,Y)))
    functional : ConceptCombiner B:TopSort * TopSort -> Boolean, U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X,Y,Z. (U1(X) & U2(Y) & U2(Z) & B(X,Y) & B(X,Z)) -> Y = Z
    surjective : ConceptCombiner B:TopSort * TopSort -> Boolean, U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall Y. U2(Y) -> (exists X. (U1(X) & B(X,Y)))
    injective : ConceptCombiner B:TopSort * TopSort -> Boolean, U1:TopSort -> Boolean, U2:TopSort -> Boolean. forall X,Y,Z. (U1(X) & U1(Y) & U2(Z) & B(X,Z) & B(Y,Z)) -> X = Y
    node_info : [none at_least_one at_most_one]
    edge_info : [all_to_all none_to_none]
    node_label : [node_necessarily node_necessarily_not]
Combinations:
    [node_info node_info nodes]
    [edge_info edge_info edges nodes nodes]
    [node_label node_label nodes node_labels]
--- PASS: TestCDConceptDomainOutput (0.00s)
=== RUN   TestGetStandardCombiners
--- PASS: TestGetStandardCombiners (0.00s)
=== RUN   TestGetStandardCombinations
--- PASS: TestGetStandardCombinations (0.00s)
=== RUN   TestGetInitialConceptDomain
--- PASS: TestGetInitialConceptDomain (0.00s)
=== RUN   TestGetInitialConceptDomainEmpty
--- PASS: TestGetInitialConceptDomainEmpty (0.00s)
=== RUN   TestCISCreation
--- PASS: TestCISCreation (0.00s)
=== RUN   TestCISPushPop
--- PASS: TestCISPushPop (0.00s)
=== RUN   TestCISPopEmpty
--- PASS: TestCISPopEmpty (0.00s)
=== RUN   TestCISUndo
--- PASS: TestCISUndo (0.43s)
=== RUN   TestCISSplit
--- PASS: TestCISSplit (0.93s)
=== RUN   TestCISSupposeEmpty
--- PASS: TestCISSupposeEmpty (0.42s)
=== RUN   TestCISRemoveConcepts
--- PASS: TestCISRemoveConcepts (0.13s)
=== RUN   TestCISToFormula
--- PASS: TestCISToFormula (0.00s)
=== RUN   TestCISFreshConstName
--- PASS: TestCISFreshConstName (0.00s)
=== RUN   TestCISClone
--- PASS: TestCISClone (0.00s)
=== RUN   TestCISSaveDomain
--- PASS: TestCISSaveDomain (0.00s)
=== RUN   TestCISLoadDomain
--- PASS: TestCISLoadDomain (0.39s)
=== RUN   TestCISLoadDomainNotFound
--- PASS: TestCISLoadDomainNotFound (0.00s)
=== RUN   TestCISReplaceDomain
--- PASS: TestCISReplaceDomain (0.25s)
=== RUN   TestCISMaterializeNode
--- PASS: TestCISMaterializeNode (0.57s)
=== RUN   TestCISMaterializeEdge
--- PASS: TestCISMaterializeEdge (0.79s)
=== RUN   TestCISSuppose
--- PASS: TestCISSuppose (0.00s)
=== RUN   TestCISSupposeTautology
--- PASS: TestCISSupposeTautology (0.00s)
=== RUN   TestCISAddEdge
--- PASS: TestCISAddEdge (0.66s)
=== RUN   TestCISAddCustomEdge
--- PASS: TestCISAddCustomEdge (0.39s)
=== RUN   TestCISAddCustomNodeLabel
--- PASS: TestCISAddCustomNodeLabel (0.39s)
=== RUN   TestAlphaNoSolver
--- PASS: TestAlphaNoSolver (0.00s)
=== RUN   TestAlphaNoSolverEmptyCache
--- PASS: TestAlphaNoSolverEmptyCache (0.00s)
=== RUN   TestTagString
--- PASS: TestTagString (0.00s)
=== RUN   TestTagFromString
--- PASS: TestTagFromString (0.00s)
=== RUN   TestCombination
--- PASS: TestCombination (0.00s)
=== RUN   TestCartesianProduct
--- PASS: TestCartesianProduct (0.00s)
=== RUN   TestCartesianProductEmpty
--- PASS: TestCartesianProductEmpty (0.00s)
=== RUN   TestCartesianProductSingle
--- PASS: TestCartesianProductSingle (0.00s)
=== RUN   TestUnionLists
--- PASS: TestUnionLists (0.00s)
=== RUN   TestToConceptSpaceAtom
--- PASS: TestToConceptSpaceAtom (0.00s)
=== RUN   TestToConceptSpaceNegated
--- PASS: TestToConceptSpaceNegated (0.00s)
=== RUN   TestToConceptSpaceProduct
--- PASS: TestToConceptSpaceProduct (0.00s)
=== RUN   TestToConceptSpaceSum
--- PASS: TestToConceptSpaceSum (0.00s)
=== RUN   TestToConceptSpaceNested
--- PASS: TestToConceptSpaceNested (0.00s)
=== RUN   TestToConceptSpaceSimpleSymbol
--- PASS: TestToConceptSpaceSimpleSymbol (0.00s)
=== RUN   TestNamedSpaceEnumerate
--- PASS: TestNamedSpaceEnumerate (0.00s)
=== RUN   TestProductSpaceEnumerate
--- PASS: TestProductSpaceEnumerate (0.00s)
=== RUN   TestSumSpaceEnumerate
--- PASS: TestSumSpaceEnumerate (0.00s)
=== RUN   TestProductSpaceEnumerateEmpty
--- PASS: TestProductSpaceEnumerateEmpty (0.00s)
=== RUN   TestCSLiteralNegate
--- PASS: TestCSLiteralNegate (0.00s)
=== RUN   TestGetStructureConceptDomain
--- PASS: TestGetStructureConceptDomain (0.00s)
=== RUN   TestGetStructureConceptAbstractValue
--- PASS: TestGetStructureConceptAbstractValue (0.00s)
=== RUN   TestUniverseElementToConceptName
--- PASS: TestUniverseElementToConceptName (0.00s)
=== RUN   TestUniverseElementToConceptNameAlreadyContainsSort
--- PASS: TestUniverseElementToConceptNameAlreadyContainsSort (0.00s)
=== RUN   TestGetDiagramConceptDomain
--- PASS: TestGetDiagramConceptDomain (0.00s)
=== RUN   TestGetStructureRenaming
--- PASS: TestGetStructureRenaming (0.00s)
=== RUN   TestCDCombinerDictCopy
--- PASS: TestCDCombinerDictCopy (0.00s)
=== RUN   TestCDCombinerDictHas
--- PASS: TestCDCombinerDictHas (0.00s)
=== RUN   TestCDConceptCallBinary
--- PASS: TestCDConceptCallBinary (0.00s)
=== RUN   TestCDConceptDomainGetFactsProjectionFilter
--- PASS: TestCDConceptDomainGetFactsProjectionFilter (0.00s)
=== RUN   TestCDConceptDomainGetCombFacts
--- PASS: TestCDConceptDomainGetCombFacts (0.00s)
=== RUN   TestNewCyElements
--- PASS: TestNewCyElements (0.00s)
=== RUN   TestAddNode
--- PASS: TestAddNode (0.00s)
=== RUN   TestAddEdge
--- PASS: TestAddEdge (0.00s)
=== RUN   TestAddMultipleClasses
--- PASS: TestAddMultipleClasses (0.00s)
=== RUN   TestCyElementsJSON
--- PASS: TestCyElementsJSON (0.00s)
=== RUN   TestCyElementsJSONNodeFields
--- PASS: TestCyElementsJSONNodeFields (0.00s)
=== RUN   TestCyElementsJSONEdgeFields
--- PASS: TestCyElementsJSONEdgeFields (0.00s)
=== RUN   TestNodeWidthHeuristic
--- PASS: TestNodeWidthHeuristic (0.00s)
=== RUN   TestRenderARGEmpty
--- PASS: TestRenderARGEmpty (0.00s)
=== RUN   TestRenderARGNodes
--- PASS: TestRenderARGNodes (0.00s)
=== RUN   TestRenderARGTransitions
--- PASS: TestRenderARGTransitions (0.00s)
=== RUN   TestRenderARGJoinEdge
--- PASS: TestRenderARGJoinEdge (0.00s)
=== RUN   TestRenderARGCovering
--- PASS: TestRenderARGCovering (0.00s)
=== RUN   TestRenderARGFullJSON
--- PASS: TestRenderARGFullJSON (0.00s)
=== RUN   TestRenderProofStackNil
--- PASS: TestRenderProofStackNil (0.00s)
=== RUN   TestRenderProofStackGoals
--- PASS: TestRenderProofStackGoals (0.00s)
=== RUN   TestRenderConceptGraphNil
--- PASS: TestRenderConceptGraphNil (0.00s)
=== RUN   TestRenderConceptGraphNodes
--- PASS: TestRenderConceptGraphNodes (0.00s)
=== RUN   TestRenderConceptGraphEdgeVisibility
--- PASS: TestRenderConceptGraphEdgeVisibility (0.00s)
=== RUN   TestConceptShapeOctagon
--- PASS: TestConceptShapeOctagon (0.00s)
=== RUN   TestConceptStyleJSON
--- PASS: TestConceptStyleJSON (0.00s)
=== RUN   TestARGStyleJSON
--- PASS: TestARGStyleJSON (0.00s)
=== RUN   TestProofStyleJSON
--- PASS: TestProofStyleJSON (0.00s)
=== RUN   TestDisplayCheckboxesDefault
--- PASS: TestDisplayCheckboxesDefault (0.00s)
=== RUN   TestDisplayCheckboxesSetEdge
--- PASS: TestDisplayCheckboxesSetEdge (0.00s)
=== RUN   TestDisplayCheckboxesSetNodeLabel
--- PASS: TestDisplayCheckboxesSetNodeLabel (0.00s)
=== RUN   TestDisplayCheckboxesNilReceiver
--- PASS: TestDisplayCheckboxesNilReceiver (0.00s)
=== RUN   TestConceptSessionSplit
--- PASS: TestConceptSessionSplit (0.00s)
=== RUN   TestConceptSessionUndoSplit
--- PASS: TestConceptSessionUndoSplit (0.00s)
=== RUN   TestConceptSessionRemove
--- PASS: TestConceptSessionRemove (0.00s)
=== RUN   TestConceptSessionSupposeEmpty
--- PASS: TestConceptSessionSupposeEmpty (0.00s)
=== RUN   TestNewGraph
--- PASS: TestNewGraph (0.00s)
=== RUN   TestGraphSortNames
--- PASS: TestGraphSortNames (0.00s)
=== RUN   TestGraphReset
--- PASS: TestGraphReset (0.00s)
=== RUN   TestGraphNewRelation
--- PASS: TestGraphNewRelation (0.00s)
=== RUN   TestGraphConceptLabel
--- PASS: TestGraphConceptLabel (0.00s)
=== RUN   TestGraphNodeLabel
--- PASS: TestGraphNodeLabel (0.00s)
=== RUN   TestGraphConceptLabelNil
--- PASS: TestGraphConceptLabelNil (0.00s)
=== RUN   TestGraphNodeLabelNil
--- PASS: TestGraphNodeLabelNil (0.00s)
=== RUN   TestDisplayCheckboxesEdge
--- PASS: TestDisplayCheckboxesEdge (0.00s)
=== RUN   TestDisplayCheckboxesNodeLabel
--- PASS: TestDisplayCheckboxesNodeLabel (0.00s)
=== RUN   TestDisplayCheckboxesEnsure
--- PASS: TestDisplayCheckboxesEnsure (0.00s)
=== RUN   TestDisplayCheckboxesEnsureNodeLabel
--- PASS: TestDisplayCheckboxesEnsureNodeLabel (0.00s)
=== RUN   TestDisplayCheckboxesUnknownEdge
--- PASS: TestDisplayCheckboxesUnknownEdge (0.00s)
=== RUN   TestGraphSetCheckbox
--- PASS: TestGraphSetCheckbox (0.00s)
=== RUN   TestGraphSetCheckboxTransitive
--- PASS: TestGraphSetCheckboxTransitive (0.00s)
=== RUN   TestGraphShowCheckboxes
--- PASS: TestGraphShowCheckboxes (0.00s)
=== RUN   TestGraphStackUndoRedo
--- PASS: TestGraphStackUndoRedo (0.00s)
=== RUN   TestGraphStackCheckpointBacktrack
--- PASS: TestGraphStackCheckpointBacktrack (0.00s)
=== RUN   TestGraphCopy
--- PASS: TestGraphCopy (0.00s)
=== RUN   TestGraphAttributes
--- PASS: TestGraphAttributes (0.00s)
=== RUN   TestGetTransitiveReduction
--- PASS: TestGetTransitiveReduction (0.00s)
=== RUN   TestGraphProjection
--- PASS: TestGraphProjection (0.00s)
=== RUN   TestStandardGraph
--- PASS: TestStandardGraph (0.00s)
=== RUN   TestGetShape
--- PASS: TestGetShape (0.00s)
=== RUN   TestNodeGT
--- PASS: TestNodeGT (0.00s)
=== RUN   TestOptionValue
--- PASS: TestOptionValue (0.00s)
=== RUN   TestRegression_ConceptViewShowsBothSorts
--- PASS: TestRegression_ConceptViewShowsBothSorts (0.00s)
=== RUN   TestRegression_RelationListShowsParams
--- PASS: TestRegression_RelationListShowsParams (0.00s)
=== RUN   TestRegression_StateGraphHasInitialNode
--- PASS: TestRegression_StateGraphHasInitialNode (0.00s)
=== RUN   TestRodTimingBreakdown
--- PASS: TestRodTimingBreakdown (0.00s)
=== RUN   TestNewServer
--- PASS: TestNewServer (0.00s)
=== RUN   TestHandleIndex
--- PASS: TestHandleIndex (0.00s)
=== RUN   TestHandleIndex404
--- PASS: TestHandleIndex404 (0.00s)
=== RUN   TestStaticNotFound
--- PASS: TestStaticNotFound (0.00s)
=== RUN   TestAPINewSession
--- PASS: TestAPINewSession (0.00s)
=== RUN   TestAPINewSessionGETFails
--- PASS: TestAPINewSessionGETFails (0.00s)
=== RUN   TestAPILoadFile
--- PASS: TestAPILoadFile (0.00s)
=== RUN   TestAPILoadFileEmpty
--- PASS: TestAPILoadFileEmpty (0.00s)
=== RUN   TestAPILoadFileBadJSON
--- PASS: TestAPILoadFileBadJSON (0.00s)
=== RUN   TestAPIAction
--- PASS: TestAPIAction (0.00s)
=== RUN   TestAPIActionEmpty
--- PASS: TestAPIActionEmpty (0.00s)
=== RUN   TestAPIARG
--- PASS: TestAPIARG (0.00s)
=== RUN   TestAPIConcept
--- PASS: TestAPIConcept (0.00s)
=== RUN   TestAPIConceptSplitNotFound
--- PASS: TestAPIConceptSplitNotFound (0.00s)
=== RUN   TestAPIConceptEmptyNotFound
--- PASS: TestAPIConceptEmptyNotFound (0.00s)
=== RUN   TestAPIConceptRemoveNotFound
--- PASS: TestAPIConceptRemoveNotFound (0.00s)
=== RUN   TestAPIConceptUndoEmpty
--- PASS: TestAPIConceptUndoEmpty (0.00s)
=== RUN   TestAPIConceptMaterializeNotFound
--- PASS: TestAPIConceptMaterializeNotFound (0.00s)
=== RUN   TestAPICheck
--- PASS: TestAPICheck (0.00s)
=== RUN   TestAPISessionNotFound
--- PASS: TestAPISessionNotFound (0.00s)
=== RUN   TestAPIUnknownEndpoint
--- PASS: TestAPIUnknownEndpoint (0.00s)
=== RUN   TestAPIEventsSSE
--- PASS: TestAPIEventsSSE (0.00s)
=== RUN   TestMultipleSessions
--- PASS: TestMultipleSessions (0.00s)
=== RUN   TestAPIMethodNotAllowed
--- PASS: TestAPIMethodNotAllowed (0.00s)
=== RUN   TestNewAnalysisGraphUI
--- PASS: TestNewAnalysisGraphUI (0.00s)
=== RUN   TestAnalysisGraphUISetMode
--- PASS: TestAnalysisGraphUISetMode (0.00s)
=== RUN   TestAnalysisGraphUIMenus
--- PASS: TestAnalysisGraphUIMenus (0.00s)
=== RUN   TestAnalysisGraphUIStart
--- PASS: TestAnalysisGraphUIStart (0.00s)
=== RUN   TestAnalysisGraphUINodeColor
--- PASS: TestAnalysisGraphUINodeColor (0.00s)
=== RUN   TestAnalysisGraphUIGetNodeActions
--- PASS: TestAnalysisGraphUIGetNodeActions (0.00s)
=== RUN   TestAnalysisGraphUIGetEdgeActions
--- PASS: TestAnalysisGraphUIGetEdgeActions (0.00s)
=== RUN   TestAnalysisGraphUINodeCommands
--- PASS: TestAnalysisGraphUINodeCommands (0.00s)
=== RUN   TestAnalysisGraphUIMarkNode
--- PASS: TestAnalysisGraphUIMarkNode (0.00s)
=== RUN   TestAnalysisGraphUIRememberGraph
--- PASS: TestAnalysisGraphUIRememberGraph (0.00s)
=== RUN   TestAnalysisGraphUIDeleteNode
--- PASS: TestAnalysisGraphUIDeleteNode (0.00s)
=== RUN   TestAnalysisGraphUICheckSafety
--- PASS: TestAnalysisGraphUICheckSafety (0.00s)
=== RUN   TestStateEquationLabel
--- PASS: TestStateEquationLabel (0.00s)
=== RUN   TestGraphWidgetMenus
--- PASS: TestGraphWidgetMenus (0.00s)
=== RUN   TestGraphWidgetUndoRedo
--- PASS: TestGraphWidgetUndoRedo (0.00s)
=== RUN   TestGraphWidgetBacktrack
--- PASS: TestGraphWidgetBacktrack (0.00s)
=== RUN   TestGraphWidgetGetNodeActions
--- PASS: TestGraphWidgetGetNodeActions (0.00s)
=== RUN   TestGraphWidgetGetEdgeActions
--- PASS: TestGraphWidgetGetEdgeActions (0.00s)
=== RUN   TestGraphWidgetSelectNode
--- PASS: TestGraphWidgetSelectNode (0.00s)
=== RUN   TestGraphWidgetClearSelection
--- PASS: TestGraphWidgetClearSelection (0.00s)
=== RUN   TestGraphWidgetShowRelation
--- PASS: TestGraphWidgetShowRelation (0.00s)
=== RUN   TestGraphWidgetClearEdges
--- PASS: TestGraphWidgetClearEdges (0.00s)
=== RUN   TestGraphWidgetApplyStructureRenaming
--- PASS: TestGraphWidgetApplyStructureRenaming (0.00s)
=== RUN   TestExtensionPointRegisterAndInvoke
--- PASS: TestExtensionPointRegisterAndInvoke (0.00s)
=== RUN   TestExtensionPointAction
--- PASS: TestExtensionPointAction (0.00s)
=== RUN   TestExtensionPointUnregister
--- PASS: TestExtensionPointUnregister (0.00s)
=== RUN   TestExtensionPointUnregisterOutOfRange
--- PASS: TestExtensionPointUnregisterOutOfRange (0.00s)
=== RUN   TestArgNodeActionsInitRegistered
--- PASS: TestArgNodeActionsInitRegistered (0.00s)
=== RUN   TestGoalNodeActionsEmpty
--- PASS: TestGoalNodeActionsEmpty (0.00s)
=== RUN   TestNewCTIAnalysisGraphUI
--- PASS: TestNewCTIAnalysisGraphUI (0.00s)
=== RUN   TestCTIMenus
--- PASS: TestCTIMenus (0.00s)
=== RUN   TestCTIWeaken
--- PASS: TestCTIWeaken (0.00s)
=== RUN   TestCTIWeakenEmpty
--- PASS: TestCTIWeakenEmpty (0.00s)
=== RUN   TestCTISaveConjectures
--- PASS: TestCTISaveConjectures (0.00s)
=== RUN   TestWriteConjecture
--- PASS: TestWriteConjecture (0.00s)
=== RUN   TestConceptGraphUIMenus
--- PASS: TestConceptGraphUIMenus (0.00s)
=== RUN   TestEventTraceViewerNewSheet
--- PASS: TestEventTraceViewerNewSheet (0.00s)
=== RUN   TestEventTraceViewerLookup
--- PASS: TestEventTraceViewerLookup (0.00s)
=== RUN   TestFilterEvents
--- PASS: TestFilterEvents (0.00s)
=== RUN   TestFindEvent
--- PASS: TestFindEvent (0.00s)
=== RUN   TestFormatEvent
--- PASS: TestFormatEvent (0.00s)
=== RUN   TestFormatTrace
--- PASS: TestFormatTrace (0.00s)
=== RUN   TestEventTraceViewerPatterns
--- PASS: TestEventTraceViewerPatterns (0.00s)
=== RUN   TestCenterWindow
--- PASS: TestCenterWindow (0.00s)
=== RUN   TestCenterWindowOnWindow
--- PASS: TestCenterWindowOnWindow (0.00s)
=== RUN   TestConvertToInt
--- PASS: TestConvertToInt (0.00s)
=== RUN   TestFileBrowserState
--- PASS: TestFileBrowserState (0.00s)
=== RUN   TestRunContext
--- PASS: TestRunContext (0.00s)
=== RUN   TestBuildMenuBar
--- PASS: TestBuildMenuBar (0.00s)
=== RUN   TestShowVerification
--- PASS: TestShowVerification (0.00s)
=== RUN   TestLaunchUI
--- PASS: TestLaunchUI (0.00s)
=== RUN   TestNewAnalysisSessionWidget
--- PASS: TestNewAnalysisSessionWidget (0.00s)
=== RUN   TestConceptSessionControlsButtons
--- PASS: TestConceptSessionControlsButtons (0.00s)
=== RUN   TestEdgeClassClickToggles
--- PASS: TestEdgeClassClickToggles (0.00s)
=== RUN   TestWeakenInteraction
--- PASS: TestWeakenInteraction (0.00s)
=== RUN   TestWeakenCancel
--- PASS: TestWeakenCancel (0.00s)
=== RUN   TestStrengthenAddsConjecture
--- PASS: TestStrengthenAddsConjecture (0.00s)
=== CONT  TestBrowserPageLoads
=== CONT  TestBrowserCytoscapeLoads
=== CONT  TestBrowserCreateSession
=== CONT  TestBrowserSSEEvents
=== CONT  TestBrowserRightClickShowsMenu
=== CONT  TestBrowserHasMenuBar
=== CONT  TestBrowserContextMenuHidden
=== CONT  TestBrowserModeSelect
--- PASS: TestBrowserContextMenuHidden (12.03s)
=== CONT  TestBrowserARGGraphInitializes
=== NAME  TestBrowserCreateSession
    browser_test.go:281: session id from browser: s2
--- PASS: TestBrowserCreateSession (12.61s)
=== CONT  TestBrowserGetConcept
=== NAME  TestBrowserSSEEvents
    browser_test.go:596: SSE events received: 2
--- PASS: TestBrowserSSEEvents (13.14s)
=== CONT  TestBrowserGetARG
--- PASS: TestBrowserHasMenuBar (13.28s)
=== CONT  TestBrowserHasInfoPanel
--- PASS: TestBrowserModeSelect (13.58s)
=== CONT  TestBrowserHasStatusBar
=== NAME  TestBrowserPageLoads
    browser_test.go:155: page title: IVy Interactive Verification
--- PASS: TestBrowserPageLoads (14.27s)
=== CONT  TestBrowserPanelResize
--- PASS: TestBrowserCytoscapeLoads (14.37s)
=== CONT  TestBrowserDividerExists
--- PASS: TestBrowserRightClickShowsMenu (15.69s)
=== CONT  TestBrowserCheckButton
--- PASS: TestBrowserARGGraphInitializes (6.28s)
=== CONT  TestBrowserUndoButton
--- PASS: TestBrowserGetConcept (12.60s)
=== CONT  TestBrowserStaticJS
--- PASS: TestBrowserGetARG (12.77s)
=== CONT  TestBrowserConceptGraphRightClick
--- PASS: TestBrowserDividerExists (15.92s)
=== CONT  TestBrowserGraphHealthCheck
--- PASS: TestBrowserHasStatusBar (17.35s)
=== CONT  TestBrowserStaticCSS
--- PASS: TestBrowserHasInfoPanel (18.89s)
=== CONT  TestBrowserInvalidSession
=== NAME  TestBrowserPanelResize
    browser_test.go:553: panel width: before=200 after=214
--- PASS: TestBrowserPanelResize (18.46s)
=== CONT  TestBrowserHasARGPanel
--- PASS: TestBrowserUndoButton (20.11s)
=== CONT  TestBrowserHasConceptPanel
=== NAME  TestBrowserCheckButton
    browser_test.go:373: status after check: Check PASSED
--- PASS: TestBrowserCheckButton (29.67s)
=== CONT  TestBrowserHasTitle
--- PASS: TestBrowserGraphHealthCheck (21.74s)
--- PASS: TestBrowserHasConceptPanel (15.26s)
--- PASS: TestBrowserStaticJS (28.56s)
--- PASS: TestBrowserHasARGPanel (21.10s)
--- PASS: TestBrowserInvalidSession (21.69s)
--- PASS: TestBrowserHasTitle (9.33s)
=== NAME  TestBrowserConceptGraphRightClick
    browser_test.go:102: newPage cleanup: page.Close: context deadline exceeded
--- FAIL: TestBrowserConceptGraphRightClick (30.31s)
panic: context deadline exceeded [recovered, repanicked]

goroutine 153 [running]:
testing.tRunner.func1.2({0x103b12640, 0x103c9bc60})
	/usr/local/go/src/testing/testing.go:1974 +0x232
testing.tRunner.func1()
	/usr/local/go/src/testing/testing.go:1977 +0x349
panic({0x103b12640?, 0x103c9bc60?})
	/usr/local/go/src/runtime/panic.go:860 +0x13a
github.com/go-rod/rod/lib/utils.init.func2({0x103b12640?, 0x103c9bc60?})
	/Users/jaten/go/pkg/mod/github.com/go-rod/rod@v0.116.2/lib/utils/utils.go:68 +0x1d
github.com/glycerine/ivy/goivy/webui.initSharedBrowser.New.(*Browser).WithPanic.genE.func1({0x3078b35b8340?, 0x1030394ab?, 0x3078b3177e68?})
	/Users/jaten/go/pkg/mod/github.com/go-rod/rod@v0.116.2/must.go:36 +0x5c
github.com/go-rod/rod.(*Page).MustEval(0x3078b33986e0, {0x1030394ab?, 0x3078b3158f50?}, {0x0?, 0x1?, 0x755?})
	/Users/jaten/go/pkg/mod/github.com/go-rod/rod@v0.116.2/must.go:489 +0x92
github.com/glycerine/ivy/goivy/webui.TestBrowserConceptGraphRightClick(0x3078b32adb08)
	/Users/jaten/ivy/goivy/webui/browser_test.go:747 +0x225
testing.tRunner(0x3078b32adb08, 0x103bd1b40)
	/usr/local/go/src/testing/testing.go:2036 +0xea
created by testing.(*T).Run in goroutine 1
	/usr/local/go/src/testing/testing.go:2101 +0x4c5
FAIL	github.com/glycerine/ivy/goivy/webui	65.028s
FAIL
make: *** [test-web] Error 1
(goivy-venv) jaten@jbook ~/ivy/goivy (master) $ 