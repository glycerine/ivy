package ast

import (
	"fmt"

	iu "github.com/glycerine/goivy/ivyutils"
)

// Canon() implementations for all declaration types defined in decl.go.

// --- LabeledFormula ---

func (lf *LabeledFormula) Canon() iu.Canonical {
	// python vs go different id, so show :0 for now. no lineno either
	//return iu.Canonical(fmt.Sprintf("(labeledFormula%v label:%v formula:%v id:%d lineno:%d temporal:%v explicit:%v isDefinition:%v assumed:%v unprovable:%v)",
	return iu.Canonical(fmt.Sprintf("(labeledFormula label:%v formula:%v id:%d temporal:%v explicit:%v isDefinition:%v assumed:%v unprovable:%v)",
		//lf.Base.canonFields(),
		nodeCanon(lf.Label), nodeCanon(lf.Formula),
		lf.ID,
		//lf.Lineno,
		//0, // faked out zero lineno
		boolPtrCanon(lf.Temporal), lf.Explicit, lf.IsDefinition, lf.Assumed, lf.Unprovable))
}

// --- DeclBase-only types (no extra fields beyond DeclBase) ---

func (d *MacroDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(macroDecl%v)", d.DeclBase.canonFields()))
}

func (d *ObjectDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(objectDecl%v)", d.DeclBase.canonFields()))
}

func (d *ActionDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(actionDecl%v)", d.DeclBase.canonFields()))
}

func (d *RelationDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(relationDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConstantDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(constantDecl%v)", d.DeclBase.canonFields()))
}

func (d *ParameterDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(parameterDecl%v)", d.DeclBase.canonFields()))
}

func (d *DestructorDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(destructorDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConstructorDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(constructorDecl%v)", d.DeclBase.canonFields()))
}

func (d *TypeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(typeDecl%v)", d.DeclBase.canonFields()))
}

func (d *VariantDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(variantDecl%v)", d.DeclBase.canonFields()))
}

func (d *AxiomDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(axiomDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConjectureDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(conjectureDecl%v)", d.DeclBase.canonFields()))
}

func (d *ProofDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(proofDecl%v)", d.DeclBase.canonFields()))
}

func (d *NamedDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(namedDecl%v)", d.DeclBase.canonFields()))
}

func (d *SchemaDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(schemaDecl%v)", d.DeclBase.canonFields()))
}

func (d *TheoremDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(theoremDecl%v)", d.DeclBase.canonFields()))
}

func (d *DerivedDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(derivedDecl%v)", d.DeclBase.canonFields()))
}

func (d *DefinitionDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(definitionDecl%v)", d.DeclBase.canonFields()))
}

func (d *ProgressDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(progressDecl%v)", d.DeclBase.canonFields()))
}

func (d *RelyDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(relyDecl%v)", d.DeclBase.canonFields()))
}

func (d *MixOrdDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixOrdDecl%v)", d.DeclBase.canonFields()))
}

func (d *ConceptDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(conceptDecl%v)", d.DeclBase.canonFields()))
}

func (d *InitDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(initDecl%v)", d.DeclBase.canonFields()))
}

func (d *StateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(stateDecl%v)", d.DeclBase.canonFields()))
}

func (d *UpdateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(updateDecl%v)", d.DeclBase.canonFields()))
}

func (d *AssertDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(assertDecl%v)", d.DeclBase.canonFields()))
}

func (d *InterpretDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(interpretDecl%v)", d.DeclBase.canonFields()))
}

func (d *MixinDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinDecl%v)", d.DeclBase.canonFields()))
}

func (d *IsolateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(isolateDecl%v)", d.DeclBase.canonFields()))
}

func (d *ExportDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(exportDecl%v)", d.DeclBase.canonFields()))
}

func (d *ImportDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(importDecl%v)", d.DeclBase.canonFields()))
}

func (d *PrivateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(privateDecl%v)", d.DeclBase.canonFields()))
}

func (d *AliasDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(aliasDecl%v)", d.DeclBase.canonFields()))
}

func (d *DelegateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(delegateDecl%v)", d.DeclBase.canonFields()))
}

func (d *ImplementTypeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(implementTypeDecl%v)", d.DeclBase.canonFields()))
}

func (d *NativeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeDecl%v)", d.DeclBase.canonFields()))
}

func (d *AttributeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(attributeDecl%v)", d.DeclBase.canonFields()))
}

func (d *InstantiateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(instantiateDecl%v)", d.DeclBase.canonFields()))
}

func (d *AutoInstanceDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(autoInstanceDecl%v)", d.DeclBase.canonFields()))
}

func (d *ScenarioDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioDecl%v)", d.DeclBase.canonFields()))
}

func (d *SubclassDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(subclassDecl%v)", d.DeclBase.canonFields()))
}

// --- Embedding wrapper types ---

func (d *PropertyDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(propertyDecl%v)", d.DeclBase.canonFields()))
}

func (d *FreshConstantDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(freshConstantDecl%v)", d.DeclBase.canonFields()))
}

func (d *GhostTypeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(ghostTypeDef%v name:%v value:%v finite:%v)", d.Base.canonFields(), nodeCanon(d.Name), nodeCanon(d.Value), d.Finite))
}

func (d *TrustedIsolateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(trustedIsolateDef%v elems:%v withArgs:%d trusted:%v isObject:%v)", d.Base.canonFields(), SliceCanon(d.Elems), d.WithArgs, d.Trusted, d.IsObject))
}

func (d *ExtractDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(extractDef%v elems:%v withArgs:%d trusted:%v isObject:%v)", d.Base.canonFields(), SliceCanon(d.Elems), d.WithArgs, d.Trusted, d.IsObject))
}

func (d *ProcessDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(processDef%v elems:%v withArgs:%d trusted:%v isObject:%v)", d.Base.canonFields(), SliceCanon(d.Elems), d.WithArgs, d.Trusted, d.IsObject))
}

func (d *IsolateObjectDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(isolateObjectDecl%v)", d.DeclBase.canonFields()))
}

// --- Types with custom fields ---

func (a *ActionDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(actionDef%v name:%v body:%v formalParams:%v formalReturns:%v)",
		a.Base.canonFields(), nodeCanon(a.Name), nodeCanon(a.Body), SliceCanon(a.FormalParams), SliceCanon(a.FormalReturns)))
}

func (t *TypeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(typeDef%v name:%v value:%v finite:%v)",
		t.Base.canonFields(), nodeCanon(t.Name), nodeCanon(t.Value), t.Finite))
}

func (v *VariantDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(variantDef%v name:%v vSort:%v)",
		v.Base.canonFields(), nodeCanon(v.Name), nodeCanon(v.VSort)))
}

func (s *SchemaBody) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(schemaBody%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}

func (s *Schema) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(schema%v defn:%v fresh:%v instances:%v)",
		s.Base.canonFields(), nodeCanon(s.Defn), SliceCanon(s.Fresh), SliceCanon(s.Instances)))
}

// Python MixinDef subclasses use the generic _ast_canon fallback:
//   (typeName lineno_fields)
// They do NOT emit mixer/mixee fields in canon output.

func (m *MixinBeforeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinBeforeDef%v)", m.Base.canonFields()))
}

func (m *MixinImplementDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinImplementDef%v)", m.Base.canonFields()))
}

func (m *MixinAfterDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinAfterDef%v)", m.Base.canonFields()))
}

func (i *IsolateDef) Canon() iu.Canonical {
	typeName := "isolateDef"
	if i.Trusted {
		typeName = "trustedIsolateDef"
	}
	return iu.Canonical(fmt.Sprintf("(%v%v elems:%v withArgs:%d trusted:%v isObject:%v)",
		typeName, i.Base.canonFields(), SliceCanon(i.Elems), i.WithArgs, i.Trusted, i.IsObject))
}

// Python uses generic _ast_canon for these types: (typeName lineno_fields)
// Only types with specific canon in Python get detailed field output.

func (e *ExportDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(exportDef%v)", e.Base.canonFields()))
}

func (i *ImportDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(importDef%v)", i.Base.canonFields()))
}

func (d *DelegateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(delegateDef%v)", d.Base.canonFields()))
}

func (n *NativeCode) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeCode%v)", n.Base.canonFields()))
}

func (n *NativeType) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeType%v)", n.Base.canonFields()))
}

func (n *NativeExpr) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeExpr%v)", n.Base.canonFields()))
}

func (n *NativeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeDef%v)", n.Base.canonFields()))
}

func (a *AttributeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(attributeDef%v)", a.Base.canonFields()))
}

func (i *Instantiation) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(instantiation%v)", i.Base.canonFields()))
}

func (s *StateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(stateDef%v)", s.Base.canonFields()))
}

// Renaming has specific canon in Python: (renaming lineno elems:[...])
func (r *Renaming) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(renaming%v elems:%v)", r.Base.canonFields(), SliceCanon(r.Elems)))
}

func (p *PlaceList) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(placeList%v)", p.Base.canonFields()))
}

func (s *ScenarioTransition) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioTransition%v)", s.Base.canonFields()))
}

func (s *ScenarioDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioDef%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}

func (s *ScenarioBeforeMixin) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioBeforeMixin%v mixer:%v def:%v)",
		s.Base.canonFields(), nodeCanon(s.Mixer), nodeCanon(s.Def)))
}

func (s *ScenarioAfterMixin) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioAfterMixin%v mixer:%v def:%v)",
		s.Base.canonFields(), nodeCanon(s.Mixer), nodeCanon(s.Def)))
}

func (p *PrivateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(privateDef%v elems:%v)", p.Base.canonFields(), SliceCanon(p.Elems)))
}

func (d *ImplementTypeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(implementTypeDef%v elems:%v)", d.Base.canonFields(), SliceCanon(d.Elems)))
}

func (p *PatternBasedUpdate) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(patternBasedUpdate%v dfns:%v deps:%v patterns:%v)",
		p.Base.canonFields(), nodeCanon(p.Dfns), nodeCanon(p.Deps), nodeCanon(p.Patterns)))
}

func (u *UpdatePattern) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(updatePattern%v params:%v action:%v requires:%v ensures:%v)",
		u.Base.canonFields(), nodeCanon(u.Params), nodeCanon(u.Action), nodeCanon(u.Requires), nodeCanon(u.Ensures)))
}

func (u *UpdatePatternList) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(updatePatternList%v elems:%v)", u.Base.canonFields(), SliceCanon(u.Elems)))
}

func (s *SymbolList) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(symbolList%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}

func (r *RME) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(rME%v requiresFmla:%v modifiesList:%v ensuresFmla:%v)",
		r.Base.canonFields(), nodeCanon(r.RequiresFmla), SliceCanon(r.ModifiesList), nodeCanon(r.EnsuresFmla)))
}

func (n *NamedSpace) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(namedSpace%v lit:%v)", n.Base.canonFields(), nodeCanon(n.Lit)))
}

func (p *ProductSpace) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(productSpace%v elems:%v)", p.Base.canonFields(), SliceCanon(p.Elems)))
}

func (s *SumSpace) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(sumSpace%v elems:%v)", s.Base.canonFields(), SliceCanon(s.Elems)))
}
