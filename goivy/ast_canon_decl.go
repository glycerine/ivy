package goivy

import (
	"fmt"
)

// Canon() implementations for all declaration types defined in decl.go.

// --- LabeledFormula ---

func (lf *LabeledFormula) Canon() Canonical {
	// python vs go different id, so show :0 for now. no lineno either
	//return iu.Canonical(fmt.Sprintf("(labeledFormula%v label:%v formula:%v id:%d lineno:%d temporal:%v explicit:%v isDefinition:%v assumed:%v unprovable:%v)",
	return Canonical(fmt.Sprintf("(labeledFormula label:%v formula:%v id:%d temporal:%v explicit:%v isDefinition:%v assumed:%v unprovable:%v)",
		//lf.Base.canonFields(),
		nodeCanon(lf.Label), nodeCanon(lf.Formula),
		lf.ID,
		//lf.Lineno,
		//0, // faked out zero lineno
		boolPtrCanon(lf.Temporal), lf.Explicit, lf.IsDefinition, lf.Assumed, lf.Unprovable))
}

// --- DeclBase-only types (no extra fields beyond DeclBase) ---

func (d *MacroDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(macroDecl%v)", d.DeclBase.canonFields()))
}

func (d *ObjectDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(objectDecl%v)", d.DeclBase.canonFields()))
}

func (d *ActionDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(actionDecl%v)", d.DeclBase.canonFields()))
}

func (d *RelationDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(relationDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConstantDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(constantDecl%v)", d.DeclBase.canonFields()))
}

func (d *ParameterDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(parameterDecl%v)", d.DeclBase.canonFields()))
}

func (d *DestructorDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(destructorDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConstructorDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(constructorDecl%v)", d.DeclBase.canonFields()))
}

func (d *TypeDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(typeDecl%v)", d.DeclBase.canonFields()))
}

func (d *VariantDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(variantDecl%v)", d.DeclBase.canonFields()))
}

func (d *AxiomDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(axiomDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConjectureDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(conjectureDecl%v)", d.DeclBase.canonFields()))
}

func (d *ProofDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(proofDecl%v)", d.DeclBase.canonFields()))
}

func (d *NamedDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(namedDecl%v)", d.DeclBase.canonFields()))
}

func (d *SchemaDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(schemaDecl%v)", d.DeclBase.canonFields()))
}

func (d *TheoremDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(theoremDecl%v)", d.DeclBase.canonFields()))
}

func (d *DerivedDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(derivedDecl%v)", d.DeclBase.canonFields()))
}

func (d *DefinitionDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(definitionDecl%v)", d.DeclBase.canonFields()))
}

func (d *ProgressDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(progressDecl%v)", d.DeclBase.canonFields()))
}

func (d *RelyDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(relyDecl%v)", d.DeclBase.canonFields()))
}

func (d *MixOrdDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(mixOrdDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConceptDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(conceptDecl%v)", d.DeclBase.canonFields()))
}

func (d *InitDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(initDecl%v)", d.DeclBase.canonFields()))
}

func (d *StateDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(stateDecl%v)", d.DeclBase.canonFields()))
}

func (d *UpdateDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(updateDecl%v)", d.DeclBase.canonFields()))
}

func (d *AssertDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(assertDecl%v)", d.DeclBase.canonFields()))
}

func (d *InterpretDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(interpretDecl%v)", d.DeclBase.canonFields()))
}

func (d *MixinDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(mixinDecl%v)", d.DeclBase.canonFields()))
}

func (d *IsolateDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(isolateDecl%v)", d.DeclBase.canonFields()))
}

func (d *ExportDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(exportDecl%v)", d.DeclBase.canonFields()))
}

func (d *ImportDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(importDecl%v)", d.DeclBase.canonFields()))
}

func (d *PrivateDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(privateDecl%v)", d.DeclBase.canonFields()))
}

func (d *AliasDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(aliasDecl%v)", d.DeclBase.canonFields()))
}

func (d *DelegateDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(delegateDecl%v)", d.DeclBase.canonFields()))
}

func (d *ImplementTypeDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(implementTypeDecl%v)", d.DeclBase.canonFields()))
}

func (d *NativeDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(nativeDecl%v)", d.DeclBase.canonFields()))
}

func (d *AttributeDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(attributeDecl%v)", d.DeclBase.canonFields()))
}

func (d *InstantiateDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(instantiateDecl%v)", d.DeclBase.canonFields()))
}

func (d *AutoInstanceDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(autoInstanceDecl%v)", d.DeclBase.canonFields()))
}

func (d *ScenarioDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(scenarioDecl%v)", d.DeclBase.canonFields()))
}

func (d *SubclassDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(subclassDecl%v)", d.DeclBase.canonFields()))
}

// --- Embedding wrapper types ---

func (d *PropertyDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(propertyDecl%v)", d.DeclBase.canonFields()))
}

func (d *FreshConstantDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(freshConstantDecl%v)", d.DeclBase.canonFields()))
}

func (d *GhostTypeDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(ghostTypeDef%v name:%v value:%v finite:%v)", d.Base.canonFields(), nodeCanon(d.Name), nodeCanon(d.Value), d.Finite))
}

func (d *TrustedIsolateDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(trustedIsolateDef%v elems:%v withArgs:%d trusted:%v isObject:%v)", d.Base.canonFields(), SliceCanon(d.Elems), d.WithArgs, d.Trusted, d.IsObject))
}

func (d *ExtractDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(extractDef%v elems:%v withArgs:%d trusted:%v isObject:%v)", d.Base.canonFields(), SliceCanon(d.Elems), d.WithArgs, d.Trusted, d.IsObject))
}

func (d *ProcessDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(processDef%v elems:%v withArgs:%d trusted:%v isObject:%v)", d.Base.canonFields(), SliceCanon(d.Elems), d.WithArgs, d.Trusted, d.IsObject))
}

func (d *IsolateObjectDecl) Canon() Canonical {
	return Canonical(fmt.Sprintf("(isolateObjectDecl%v)", d.DeclBase.canonFields()))
}

// --- Types with custom fields ---

func (a *ActionDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(actionDef%v name:%v body:%v formalParams:%v formalReturns:%v)",
		a.Base.canonFields(), nodeCanon(a.Name), nodeCanon(a.Body), SliceCanon(a.FormalParams), SliceCanon(a.FormalReturns)))
}

func (t *TypeDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(typeDef%v name:%v value:%v finite:%v)",
		t.Base.canonFields(), nodeCanon(t.Name), nodeCanon(t.Value), t.Finite))
}

func (v *VariantDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(variantDef%v name:%v vSort:%v)",
		v.Base.canonFields(), nodeCanon(v.Name), nodeCanon(v.VSort)))
}

func (s *SchemaBody) Canon() Canonical {
	return Canonical(fmt.Sprintf("(schemaBody%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}

func (s *Schema) Canon() Canonical {
	return Canonical(fmt.Sprintf("(schema%v defn:%v fresh:%v instances:%v)",
		s.Base.canonFields(), nodeCanon(s.Defn), SliceCanon(s.Fresh), SliceCanon(s.Instances)))
}

// Python MixinDef subclasses use the generic _ast_canon fallback:
//   (typeName lineno_fields)
// They do NOT emit mixer/mixee fields in canon output.

func (m *MixinBeforeDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(mixinBeforeDef%v)", m.Base.canonFields()))
}

func (m *MixinImplementDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(mixinImplementDef%v)", m.Base.canonFields()))
}

func (m *MixinAfterDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(mixinAfterDef%v)", m.Base.canonFields()))
}

func (i *IsolateDef) Canon() Canonical {
	typeName := "isolateDef"
	if i.Trusted {
		typeName = "trustedIsolateDef"
	} else if i.Kind == "extract" {
		typeName = "extractDef"
	} else if i.Kind == "process" {
		typeName = "processDef"
	}
	return Canonical(fmt.Sprintf("(%v%v elems:%v withArgs:%d trusted:%v isObject:%v)",
		typeName, i.Base.canonFields(), SliceCanon(i.Elems), i.WithArgs, i.Trusted, i.IsObject))
}

// Python uses generic _ast_canon for these types: (typeName lineno_fields)
// Only types with specific canon in Python get detailed field output.

func (e *ExportDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(exportDef%v)", e.Base.canonFields()))
}

func (i *ImportDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(importDef%v)", i.Base.canonFields()))
}

func (d *DelegateDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(delegateDef%v)", d.Base.canonFields()))
}

func (n *NativeCode) Canon() Canonical {
	return Canonical(fmt.Sprintf("(nativeCode%v)", n.Base.canonFields()))
}

func (n *NativeType) Canon() Canonical {
	return Canonical(fmt.Sprintf("(nativeType%v)", n.Base.canonFields()))
}

func (n *NativeExpr) Canon() Canonical {
	return Canonical(fmt.Sprintf("(nativeExpr%v)", n.Base.canonFields()))
}

func (n *NativeDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(nativeDef%v)", n.Base.canonFields()))
}

func (a *AttributeDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(attributeDef%v)", a.Base.canonFields()))
}

func (i *Instantiation) Canon() Canonical {
	return Canonical(fmt.Sprintf("(instantiation%v)", i.Base.canonFields()))
}

func (s *StateDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(stateDef%v)", s.Base.canonFields()))
}

// Renaming has specific canon in Python: (renaming lineno elems:[...])
func (r *Renaming) Canon() Canonical {
	return Canonical(fmt.Sprintf("(renaming%v elems:%v)", r.Base.canonFields(), SliceCanon(r.Elems)))
}

func (p *PlaceList) Canon() Canonical {
	return Canonical(fmt.Sprintf("(placeList%v)", p.Base.canonFields()))
}

func (s *ScenarioTransition) Canon() Canonical {
	return Canonical(fmt.Sprintf("(scenarioTransition%v)", s.Base.canonFields()))
}

func (s *ScenarioDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(scenarioDef%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}

func (s *ScenarioBeforeMixin) Canon() Canonical {
	return Canonical(fmt.Sprintf("(scenarioBeforeMixin%v mixer:%v def:%v)",
		s.Base.canonFields(), nodeCanon(s.Mixer), nodeCanon(s.Def)))
}

func (s *ScenarioAfterMixin) Canon() Canonical {
	return Canonical(fmt.Sprintf("(scenarioAfterMixin%v mixer:%v def:%v)",
		s.Base.canonFields(), nodeCanon(s.Mixer), nodeCanon(s.Def)))
}

func (p *PrivateDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(privateDef%v elems:%v)", p.Base.canonFields(), SliceCanon(p.Elems)))
}

func (d *ImplementTypeDef) Canon() Canonical {
	return Canonical(fmt.Sprintf("(implementTypeDef%v elems:%v)", d.Base.canonFields(), SliceCanon(d.Elems)))
}

func (p *PatternBasedUpdate) Canon() Canonical {
	return Canonical(fmt.Sprintf("(patternBasedUpdate%v dfns:%v deps:%v patterns:%v)",
		p.Base.canonFields(), nodeCanon(p.Dfns), nodeCanon(p.Deps), nodeCanon(p.Patterns)))
}

func (u *UpdatePattern) Canon() Canonical {
	return Canonical(fmt.Sprintf("(updatePattern%v params:%v action:%v requires:%v ensures:%v)",
		u.Base.canonFields(), nodeCanon(u.Params), nodeCanon(u.Action), nodeCanon(u.Requires), nodeCanon(u.Ensures)))
}

func (u *UpdatePatternList) Canon() Canonical {
	return Canonical(fmt.Sprintf("(updatePatternList%v elems:%v)", u.Base.canonFields(), SliceCanon(u.Elems)))
}

func (s *SymbolList) Canon() Canonical {
	return Canonical(fmt.Sprintf("(symbolList%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}

func (r *RME) Canon() Canonical {
	return Canonical(fmt.Sprintf("(rME%v requiresFmla:%v modifiesList:%v ensuresFmla:%v)",
		r.Base.canonFields(), nodeCanon(r.RequiresFmla), SliceCanon(r.ModifiesList), nodeCanon(r.EnsuresFmla)))
}

func (n *NamedSpace) Canon() Canonical {
	return Canonical(fmt.Sprintf("(namedSpace%v lit:%v)", n.Base.canonFields(), nodeCanon(n.Lit)))
}

func (p *ProductSpace) Canon() Canonical {
	return Canonical(fmt.Sprintf("(productSpace%v elems:%v)", p.Base.canonFields(), SliceCanon(p.Elems)))
}

func (s *SumSpace) Canon() Canonical {
	return Canonical(fmt.Sprintf("(sumSpace%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}
