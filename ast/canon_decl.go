package ast

import (
	"fmt"

	iu "github.com/glycerine/goivy/ivyutils"
)

// Canon() implementations for all declaration types defined in decl.go.

// --- LabeledFormula ---

func (lf *LabeledFormula) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(labeledFormula base:%v label:%v formula:%v id:%d lineno:%d temporal:%v explicit:%v isDefinition:%v assumed:%v unprovable:%v)",
		lf.Base.Canon(), nodeCanon(lf.Label), nodeCanon(lf.Formula), lf.ID, lf.Lineno,
		boolPtrCanon(lf.Temporal), lf.Explicit, lf.IsDefinition, lf.Assumed, lf.Unprovable))
}

// --- DeclBase-only types (no extra fields beyond DeclBase) ---

func (d *MacroDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(macroDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ObjectDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(objectDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ActionDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(actionDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *RelationDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(relationDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ConstantDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(constantDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ParameterDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(parameterDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *DestructorDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(destructorDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ConstructorDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(constructorDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *TypeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(typeDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *VariantDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(variantDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *AxiomDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(axiomDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ConjectureDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(conjectureDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ProofDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(proofDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *NamedDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(namedDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *SchemaDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(schemaDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *TheoremDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(theoremDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *DerivedDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(derivedDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *DefinitionDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(definitionDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ProgressDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(progressDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *RelyDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(relyDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *MixOrdDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixOrdDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ConceptDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(conceptDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *InitDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(initDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *StateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(stateDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *UpdateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(updateDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *AssertDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(assertDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *InterpretDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(interpretDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *MixinDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *IsolateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(isolateDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ExportDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(exportDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ImportDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(importDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *PrivateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(privateDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *AliasDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(aliasDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *DelegateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(delegateDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ImplementTypeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(implementTypeDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *NativeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *AttributeDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(attributeDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *InstantiateDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(instantiateDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *AutoInstanceDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(autoInstanceDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *ScenarioDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioDecl declBase:%v)", d.DeclBase.Canon()))
}

func (d *SubclassDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(subclassDecl declBase:%v)", d.DeclBase.Canon()))
}

// --- Embedding wrapper types ---

func (d *PropertyDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(propertyDecl axiomDecl:%v)", d.AxiomDecl.Canon()))
}

func (d *FreshConstantDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(freshConstantDecl constantDecl:%v)", d.ConstantDecl.Canon()))
}

func (d *GhostTypeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(ghostTypeDef typeDef:%v)", d.TypeDef.Canon()))
}

func (d *TrustedIsolateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(trustedIsolateDef isolateDef:%v)", d.IsolateDef.Canon()))
}

func (d *ExtractDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(extractDef isolateDef:%v)", d.IsolateDef.Canon()))
}

func (d *ProcessDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(processDef extractDef:%v)", d.ExtractDef.Canon()))
}

func (d *IsolateObjectDecl) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(isolateObjectDecl isolateDecl:%v)", d.IsolateDecl.Canon()))
}

// --- Types with custom fields ---

func (a *ActionDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(actionDef base:%v name:%v body:%v formalParams:%v formalReturns:%v)",
		a.Base.Canon(), nodeCanon(a.Name), nodeCanon(a.Body), sliceCanon(a.FormalParams), sliceCanon(a.FormalReturns)))
}

func (t *TypeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(typeDef base:%v name:%v value:%v finite:%v)",
		t.Base.Canon(), nodeCanon(t.Name), nodeCanon(t.Value), t.Finite))
}

func (v *VariantDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(variantDef base:%v name:%v vSort:%v)",
		v.Base.Canon(), nodeCanon(v.Name), nodeCanon(v.VSort)))
}

func (s *SchemaBody) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(schemaBody base:%v elems:%v)", s.Base.Canon(), sliceCanon(s.Elems)))
}

func (s *Schema) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(schema base:%v defn:%v fresh:%v instances:%v)",
		s.Base.Canon(), nodeCanon(s.Defn), sliceCanon(s.Fresh), sliceCanon(s.Instances)))
}

func (m *MixinBeforeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinBeforeDef base:%v mixerNode:%v mixeeNode:%v)",
		m.Base.Canon(), nodeCanon(m.MixerNode), nodeCanon(m.MixeeNode)))
}

func (m *MixinImplementDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinImplementDef base:%v mixerNode:%v mixeeNode:%v)",
		m.Base.Canon(), nodeCanon(m.MixerNode), nodeCanon(m.MixeeNode)))
}

func (m *MixinAfterDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(mixinAfterDef base:%v mixerNode:%v mixeeNode:%v)",
		m.Base.Canon(), nodeCanon(m.MixerNode), nodeCanon(m.MixeeNode)))
}

func (i *IsolateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(isolateDef base:%v elems:%v withArgs:%d trusted:%v isObject:%v)",
		i.Base.Canon(), sliceCanon(i.Elems), i.WithArgs, i.Trusted, i.IsObject))
}

func (e *ExportDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(exportDef base:%v exportedNode:%v scopeNode:%v)",
		e.Base.Canon(), nodeCanon(e.ExportedNode), nodeCanon(e.ScopeNode)))
}

func (i *ImportDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(importDef base:%v imported:%v scope:%v)",
		i.Base.Canon(), nodeCanon(i.Imported), nodeCanon(i.Scope)))
}

func (d *DelegateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(delegateDef base:%v elems:%v)", d.Base.Canon(), sliceCanon(d.Elems)))
}

func (n *NativeCode) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeCode base:%v code:%q)", n.Base.Canon(), n.Code))
}

func (n *NativeType) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeType base:%v elems:%v)", n.Base.Canon(), sliceCanon(n.Elems)))
}

func (n *NativeExpr) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeExpr base:%v elems:%v aSort:%v)",
		n.Base.Canon(), sliceCanon(n.Elems), nodeCanon(n.ASort)))
}

func (n *NativeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(nativeDef base:%v elems:%v)", n.Base.Canon(), sliceCanon(n.Elems)))
}

func (a *AttributeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(attributeDef base:%v name:%v value:%v)",
		a.Base.Canon(), nodeCanon(a.Name), nodeCanon(a.Value)))
}

func (i *Instantiation) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(instantiation base:%v name:%v sort:%v)",
		i.Base.Canon(), nodeCanon(i.Name), nodeCanon(i.Sort)))
}

func (s *StateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(stateDef base:%v name:%q state:%v)",
		s.Base.Canon(), s.Name, nodeCanon(s.State)))
}

func (r *Renaming) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(renaming base:%v elems:%v)", r.Base.Canon(), sliceCanon(r.Elems)))
}

func (p *PlaceList) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(placeList base:%v elems:%v)", p.Base.Canon(), sliceCanon(p.Elems)))
}

func (s *ScenarioTransition) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioTransition base:%v from:%v to:%v action:%v)",
		s.Base.Canon(), nodeCanon(s.From), nodeCanon(s.To), nodeCanon(s.Action)))
}

func (s *ScenarioDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioDef base:%v elems:%v)", s.Base.Canon(), sliceCanon(s.Elems)))
}

func (s *ScenarioBeforeMixin) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioBeforeMixin base:%v mixer:%v def:%v)",
		s.Base.Canon(), nodeCanon(s.Mixer), nodeCanon(s.Def)))
}

func (s *ScenarioAfterMixin) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(scenarioAfterMixin base:%v mixer:%v def:%v)",
		s.Base.Canon(), nodeCanon(s.Mixer), nodeCanon(s.Def)))
}

func (p *PrivateDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(privateDef base:%v elems:%v)", p.Base.Canon(), sliceCanon(p.Elems)))
}

func (d *ImplementTypeDef) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(implementTypeDef base:%v elems:%v)", d.Base.Canon(), sliceCanon(d.Elems)))
}

func (p *PatternBasedUpdate) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(patternBasedUpdate base:%v dfns:%v deps:%v patterns:%v)",
		p.Base.Canon(), nodeCanon(p.Dfns), nodeCanon(p.Deps), nodeCanon(p.Patterns)))
}

func (u *UpdatePattern) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(updatePattern base:%v params:%v action:%v requires:%v ensures:%v)",
		u.Base.Canon(), nodeCanon(u.Params), nodeCanon(u.Action), nodeCanon(u.Requires), nodeCanon(u.Ensures)))
}

func (u *UpdatePatternList) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(updatePatternList base:%v elems:%v)", u.Base.Canon(), sliceCanon(u.Elems)))
}

func (s *SymbolList) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(symbolList base:%v elems:%v)", s.Base.Canon(), sliceCanon(s.Elems)))
}

func (r *RME) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(rME base:%v requiresFmla:%v modifiesList:%v ensuresFmla:%v)",
		r.Base.Canon(), nodeCanon(r.RequiresFmla), sliceCanon(r.ModifiesList), nodeCanon(r.EnsuresFmla)))
}

func (n *NamedSpace) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(namedSpace base:%v lit:%v)", n.Base.Canon(), nodeCanon(n.Lit)))
}

func (p *ProductSpace) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(productSpace base:%v elems:%v)", p.Base.Canon(), sliceCanon(p.Elems)))
}

func (s *SumSpace) Canon() iu.Canonical {
	return iu.Canonical(fmt.Sprintf("(sumSpace base:%v elems:%v)", s.Base.Canon(), sliceCanon(s.Elems)))
}
