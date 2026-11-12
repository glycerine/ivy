# Paused BBR work checkpoint

Paused at the user's request on 2026-11-12. This is not a completed or validated
package. Original C sources remain unchanged. No GIA files were edited.

Written but not compiled or tested: bbr.go, bbrImage.go, bbrNtbdd.go, bbrCex.go,
and bbrReach.go. They contain AIG global-BDD construction, shared limited image
forwarding, breadth-first onion-ring reachability, backwards witness extraction,
and unused-input trim/remapping. bbr_.go remains the original empty template.

Known prerequisite integration still required:

- Graph.ObjectData(edge) and Graph.SetObjectData(edge, value) do not exist yet in
  aig/aig. The implementation needs real public access to existing objectData;
  no process-global side map or graph.Data substitution was introduced.
- saig.Saig_ParBbr_t fields are private. The draft currently uses exported names
  NBddMax, NIterMax, FPartition, FReorder, FReorderImage, FVerbose, FSilent,
  FSkipOutCheck, and IFrame, anticipating an approved shared API migration.
- No BBR ordinary package, race or vet gate has run. Tests, original-C fixtures
  and separate readability/bug reviews still need to be written. Do not treat
  these source edits as integration ready.

The shared extrab image scheduler gained Extra_bddImageStartLimited and optional
manager-node checks after scheduled merges and internal image recomputations,
matching the BBR source's check placement. Negative limit disables checks;
existing Extra_bddImageStart remains unlimited. Failed construction cleans the
partial forest. Ordinary extrab tests passed after this change; new limit-specific
regressions and full dependency race/vet still need to run when work resumes.

Inspection-only next assignment: giaCex.go and giaClp.go, neither edited. Clp is
BDD collapse/DSD rebuilding (not a clustering algorithm); its actual DSD/SOP
integration dependencies need to be checked before implementation. Cex includes
simulation/value tables, state/all-object extension, frame expansion and SAT care
minimization. Preserve existing counterexample.go ownership interfaces.
