package ivyutils

import (
	"regexp"
	"runtime"
	"strings"
)

// IvyUtilsConfig holds per-session ivyutils state that was previously stored
// in package-level globals. Each concurrent Ivy model gets its own config.
type IvyUtilsConfig struct {
	// Filename is the current source file being processed.
	Filename string

	// ComposeCharacter is the separator used in qualified names ("." by default).
	ComposeCharacter string

	// ParseErrorList accumulates parse errors during a parse session.
	ParseErrorList []error

	// StdIncludeDir caches the standard include directory path.
	StdIncludeDir string

	// RuntimeCaller is a hook for runtime.Caller (allows testing).
	RuntimeCaller func(skip int) (uintptr, string, int, bool)

	// Language version state (set by SetStringVersion)
	LanguageVersion       string
	HavePolymorphism      bool
	UsePolymorphicMacros  bool
	ForbidGhostInit       bool
	LatestLanguageVersion string
	SymbolCharsParser     *regexp.Regexp

	// Parameter registry
	Registry    *ParameterRegistry
	UseNumerals *Parameter
	UseNewUI    *Parameter
	Catch       *Parameter
	DefaultUI   *Parameter
	EnableDebug *Parameter

	// UI modules
	UIModules map[string]*UIModule
}

// NewIvyUtilsConfig creates a fresh IvyUtilsConfig with defaults.
func NewIvyUtilsConfig() *IvyUtilsConfig {
	cfg := &IvyUtilsConfig{
		ComposeCharacter:      ".",
		LanguageVersion:       "1.8",
		HavePolymorphism:      true,
		LatestLanguageVersion: "1.8",
		SymbolCharsParser:     regexp.MustCompile(`[^\[\]\.]*`),
		RuntimeCaller:         runtime.Caller,
		UIModules:             make(map[string]*UIModule),
	}
	cfg.Registry = NewParameterRegistry()
	cfg.UseNumerals = NewParameterOn(cfg.Registry, "use_numerals", true)
	cfg.UseNewUI = NewParameterOn(cfg.Registry, "new_ui", false)
	cfg.Catch = NewParameterOn(cfg.Registry, "catch", true)
	cfg.DefaultUI = NewParameterOn(cfg.Registry, "ui", "cti")
	cfg.EnableDebug = NewParameterOn(cfg.Registry, "debug", false)
	return cfg
}

// DefaultIvyUtilsConfig is a transitional default for unmigrated callers.
var DefaultIvyUtilsConfig = NewIvyUtilsConfig()

// --- Method versions of key functions ---

// ComposeNames joins names with the config's ComposeCharacter, skipping "this".
func (cfg *IvyUtilsConfig) ComposeNames(names ...string) string {
	if len(names) > 0 && names[0] == "this" {
		names = names[1:]
	}
	return strings.Join(names, cfg.ComposeCharacter)
}

// GetStringVersion returns the config's language version string.
func (cfg *IvyUtilsConfig) GetStringVersion() string {
	return cfg.LanguageVersion
}

// SetStringVersion sets the language version and updates derived state.
func (cfg *IvyUtilsConfig) SetStringVersion(version string) {
	cfg.LanguageVersion = strings.TrimSpace(version)
	cfg.StdIncludeDir = "" // reset cache
	nv := GetNumericVersionFrom(cfg.LanguageVersion)
	if versionLESlice(nv, []int{1, 1}) {
		cfg.ComposeCharacter = ":"
	} else {
		cfg.ComposeCharacter = "."
	}
	cfg.SymbolCharsParser = regexp.MustCompile(`[^\[\]` + regexp.QuoteMeta(cfg.ComposeCharacter) + `]*`)
	cfg.HavePolymorphism = !versionLESlice(nv, []int{1, 2})
	cfg.UsePolymorphicMacros = !versionLESlice(nv, []int{1, 5})
	cfg.ForbidGhostInit = !versionLESlice(nv, []int{1, 6})
}

// WithSourceFile temporarily sets cfg.Filename, restoring the old value on return.
func (cfg *IvyUtilsConfig) WithSourceFile(fname string, fn func()) {
	old := cfg.Filename
	cfg.Filename = fname
	defer func() { cfg.Filename = old }()
	fn()
}

// RegisterUIModule registers a UI module by name on this config.
func (cfg *IvyUtilsConfig) RegisterUIModule(name string, mod *UIModule) {
	cfg.UIModules[name] = mod
}
