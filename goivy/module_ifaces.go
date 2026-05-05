// Package module — interfaces formerly defined locally in isolate/.
//
// These interfaces describe capabilities of AST definition types
// (MixinBeforeDef, ExportDef, DelegateDef, IsolateDef, etc.)
// that are stored in Module fields. Moving them here breaks the
// import cycle between isolate/ and module/.

package goivy

// IsolateDefInterface is the interface that isolate definitions must
// implement for IterIsolate to work. It provides access to verified
// and present component names.
//
// Concrete implementor: *ast.IsolateDef.
type IsolateDefInterface interface {
	// VerifiedNames returns the names of verified components.
	VerifiedNames() []string
	// PresentNames returns the names of present components.
	PresentNames() []string
	// IsExtract returns true if this is an extract (vs isolate) definition.
	IsExtract() bool
}

// MixinDef is the interface for mixin definitions stored in Module.Mixins.
// It abstracts over before/after/implement mixin kinds.
//
// Concrete implementors: *ast.MixinBeforeDef, *ast.MixinAfterDef,
// *ast.MixinImplementDef.
type MixinDef interface {
	// Mixer returns the name of the mixer action.
	Mixer() string
	// Mixee returns the name of the mixee (target) action.
	Mixee() string
	// IsAfter returns true if this is an after-mixin (appended after the action).
	IsAfter() bool
}

// Exporter is the interface for export definitions stored in Module.Exports.
//
// Concrete implementors: *ast.ExportDef, isolate.exportStub.
type Exporter interface {
	// Exported returns the exported action name.
	Exported() string
	// Scope returns the scope name. Empty string means global scope.
	Scope() string
}

// Delegator is the interface for delegate definitions stored in Module.Delegates.
//
// Concrete implementor: *ast.DelegateDef.
type Delegator interface {
	// Delegated returns the name of the delegated action.
	Delegated() string
	// Delegee returns the name of the delegee (target) action.
	Delegee() string
}
