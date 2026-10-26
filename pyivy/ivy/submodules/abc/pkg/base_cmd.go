package abc

// Ported from src/base/cmd/cmd.c, cmd.h, cmdApi.c, cmdAuto.c, cmdInt.h,
// cmdLoad.c, cmdPlugin.c, cmdStarter.c, cmdUtils.c, and module.make.

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/glycerine/insdict"
)

// CmdCommandID preserves the stable public command identifiers.
type CmdCommandID uint8

const (
	CmdCommandUnknown CmdCommandID = iota
	CmdCommandAlias
	CmdCommandSource
	CmdCommandSet
	CmdCommandReadBlif
	CmdCommandReadAiger
	CmdCommandStrash
	CmdCommandPrintStats
	CmdCommandPDR
	CmdCommandWriteAigerCEX
)

// CmdCommandNameToID implements the complete stable-name mapping from cmdApi.c.
func CmdCommandNameToID(name string) CmdCommandID {
	switch name {
	case "alias":
		return CmdCommandAlias
	case "source":
		return CmdCommandSource
	case "set":
		return CmdCommandSet
	case "read_blif":
		return CmdCommandReadBlif
	case "read_aiger":
		return CmdCommandReadAiger
	case "strash":
		return CmdCommandStrash
	case "print_stats":
		return CmdCommandPrintStats
	case "pdr":
		return CmdCommandPDR
	case "write_aiger_cex":
		return CmdCommandWriteAigerCEX
	default:
		return CmdCommandUnknown
	}
}

// CmdLimits bounds retained state and cumulative work for one execution.
type CmdLimits struct {
	MaxCommands   int
	MaxArguments  int
	MaxAliases    int
	MaxVariables  int
	MaxHistory    int
	MaxSources    int
	MaxOperations int
	MaxBytes      int
	MaxText       int
	MaxDepth      int
}

func (l CmdLimits) normalized() CmdLimits {
	if l.MaxCommands <= 0 {
		l.MaxCommands = 1 << 14
	}
	if l.MaxArguments <= 0 {
		l.MaxArguments = 1 << 16
	}
	if l.MaxAliases <= 0 {
		l.MaxAliases = 1 << 12
	}
	if l.MaxVariables <= 0 {
		l.MaxVariables = 1 << 12
	}
	if l.MaxHistory <= 0 {
		l.MaxHistory = 1 << 15
	}
	if l.MaxSources <= 0 {
		l.MaxSources = 1 << 12
	}
	if l.MaxOperations <= 0 {
		l.MaxOperations = 1 << 22
	}
	if l.MaxBytes <= 0 {
		l.MaxBytes = 32 << 20
	}
	if l.MaxText <= 0 {
		l.MaxText = 8 << 20
	}
	if l.MaxDepth <= 0 {
		l.MaxDepth = 64
	}
	return l
}

func validateCmdLimits(l CmdLimits) error {
	if l.MaxCommands < 0 || l.MaxArguments < 0 || l.MaxAliases < 0 || l.MaxVariables < 0 || l.MaxHistory < 0 || l.MaxSources < 0 || l.MaxOperations < 0 || l.MaxBytes < 0 || l.MaxText < 0 || l.MaxDepth < 0 {
		return ErrDimension
	}
	return nil
}

type cmdBudget struct {
	ctx   context.Context
	limit CmdLimits
	ops   int
	bytes int
}

func (b *cmdBudget) take(operations, bytes int) error {
	if b == nil || operations < 0 || bytes < 0 || b.ops > b.limit.MaxOperations-operations || b.bytes > b.limit.MaxBytes-bytes {
		return ErrLimit
	}
	b.ops += operations
	b.bytes += bytes
	if b.ctx != nil {
		if err := b.ctx.Err(); err != nil {
			return ErrCancelled
		}
	}
	return nil
}

// CmdResult is an owned command result.
type CmdResult struct {
	Output  string
	Status  int
	Changed bool
	Quit    bool
}

// CmdExecution supplies bounded session access to a concrete command handler.
type CmdExecution struct {
	session *CmdSession
	budget  *cmdBudget
	depth   int
}

func (e *CmdExecution) Session() *CmdSession { return e.session }

func (e *CmdExecution) Charge(operations, bytes int) error {
	if e == nil {
		return ErrInvariant
	}
	return e.budget.take(operations, bytes)
}

// CmdHandler replaces the native function pointer with a concrete Go function.
type CmdHandler func(context.Context, *CmdExecution, []string) (CmdResult, error)

// CmdDefinition is an owned registry entry.
type CmdDefinition struct {
	Name    string
	Group   string
	ID      CmdCommandID
	Changes bool
	Handler CmdHandler
}

// CmdRegistry preserves insertion order for deterministic help output.
type CmdRegistry struct {
	commands *insdict.Dict[string, CmdDefinition]
	limits   CmdLimits
	owner    *CmdSession
	bytes    int
	closed   bool
}

func NewCmdRegistry(limits CmdLimits) (*CmdRegistry, error) {
	if err := validateCmdLimits(limits); err != nil {
		return nil, err
	}
	limits = limits.normalized()
	return &CmdRegistry{commands: insdict.NewDictSize[string, CmdDefinition](min(limits.MaxCommands, 256)), limits: limits}, nil
}

func cmdIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for index, r := range name {
		if !(r == '_' || unicode.IsLetter(r) || index > 0 && unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

// Add atomically adds or replaces a command without changing existing order.
func (r *CmdRegistry) Add(definition CmdDefinition) error {
	if r == nil || r.commands == nil || r.closed {
		return ErrInvariant
	}
	if !cmdIdentifier(definition.Name) || definition.Group == "" || definition.Handler == nil {
		return ErrDimension
	}
	definition.ID = CmdCommandNameToID(definition.Name)
	old, found := r.commands.Get2(definition.Name)
	oldBytes := 0
	if found {
		oldBytes = len(old.Name) + len(old.Group)
	} else if r.commands.Len() >= r.limits.MaxCommands {
		return ErrLimit
	}
	newBytes := len(definition.Name) + len(definition.Group)
	retained := r.bytes
	if r.owner != nil {
		if r.owner.closed || r.owner.bytes > r.limits.MaxBytes-retained {
			return ErrLimit
		}
		retained += r.owner.bytes
	}
	if newBytes > r.limits.MaxBytes-retained+oldBytes {
		return ErrLimit
	}
	r.commands.Set(definition.Name, definition)
	r.bytes += newBytes - oldBytes
	return nil
}

func (r *CmdRegistry) Definition(name string) (CmdDefinition, bool, error) {
	if r == nil || r.commands == nil || r.closed {
		return CmdDefinition{}, false, ErrInvariant
	}
	definition, found := r.commands.Get2(name)
	return definition, found, nil
}

func (r *CmdRegistry) Definitions() ([]CmdDefinition, error) {
	if r == nil || r.commands == nil || r.closed {
		return nil, ErrInvariant
	}
	result := make([]CmdDefinition, 0, r.commands.Len())
	for _, definition := range r.commands.All() {
		result = append(result, definition)
	}
	return result, nil
}

func (r *CmdRegistry) Close() {
	if r == nil {
		return
	}
	r.closed = true
	r.commands = nil
	r.owner = nil
	r.bytes = 0
}

// CmdSession owns aliases, flags, history, source text, and a command registry.
type CmdSession struct {
	registry *CmdRegistry
	aliases  *CommandAliases
	flags    *CommandFlags
	history  *CommandHistory
	sources  *insdict.Dict[string, string]
	limits   CmdLimits
	bytes    int
	closed   bool
	quit     bool
}

func NewCmdSession(limits CmdLimits) (*CmdSession, error) {
	if err := validateCmdLimits(limits); err != nil {
		return nil, err
	}
	limits = limits.normalized()
	registry, err := NewCmdRegistry(limits)
	if err != nil {
		return nil, err
	}
	aliases, err := NewCommandAliases(limits.MaxAliases)
	if err != nil {
		return nil, err
	}
	flags, err := NewCommandFlags(limits.MaxVariables)
	if err != nil {
		return nil, err
	}
	history, err := NewCommandHistory(CommandHistoryOptions{MaxEntries: limits.MaxHistory, MaxCommandBytes: limits.MaxText})
	if err != nil {
		return nil, err
	}
	session := &CmdSession{registry: registry, aliases: aliases, flags: flags, history: history, sources: insdict.NewDictSize[string, string](min(limits.MaxSources, 128)), limits: limits}
	registry.owner = session
	if err := session.installBuiltins(); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *CmdSession) retainedBytes() (int, error) {
	if s == nil || s.closed || s.registry == nil || s.registry.closed {
		return 0, ErrInvariant
	}
	if s.registry.bytes > s.limits.MaxBytes-s.bytes {
		return 0, ErrLimit
	}
	return s.bytes + s.registry.bytes, nil
}

func (s *CmdSession) Registry() *CmdRegistry {
	if s == nil || s.closed {
		return nil
	}
	return s.registry
}

func (s *CmdSession) AddSource(name, text string) error {
	if s == nil || s.closed || s.sources == nil {
		return ErrInvariant
	}
	if !cmdIdentifier(name) {
		return ErrDimension
	}
	if len(text) > s.limits.MaxText || len(text) > s.limits.MaxBytes {
		return ErrLimit
	}
	old, found := s.sources.Get2(name)
	if !found && s.sources.Len() >= s.limits.MaxSources {
		return ErrLimit
	}
	oldBytes := 0
	if found {
		oldBytes = len(name) + len(old)
	}
	newBytes := len(name) + len(text)
	retained, err := s.retainedBytes()
	if err != nil {
		return err
	}
	if newBytes > s.limits.MaxBytes-retained+oldBytes {
		return ErrLimit
	}
	s.sources.Set(name, text)
	s.bytes += newBytes - oldBytes
	return nil
}

func (s *CmdSession) QuitRequested() bool { return s != nil && s.quit }

func (s *CmdSession) Close() {
	if s == nil || s.closed {
		return
	}
	s.closed = true
	if s.registry != nil {
		s.registry.Close()
	}
	s.aliases = nil
	s.flags = nil
	s.history = nil
	s.sources = nil
	s.bytes = 0
}

// ParseCmdText tokenizes commands with quotes, escapes, comments, and separators.
func ParseCmdText(ctx context.Context, text string, limits CmdLimits) ([][]string, error) {
	if err := validateCmdLimits(limits); err != nil {
		return nil, err
	}
	limits = limits.normalized()
	if len(text) > limits.MaxText || len(text) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if ctx == nil {
		ctx = context.Background()
	}
	budget := &cmdBudget{ctx: ctx, limit: limits}
	return parseCmdText(text, budget)
}

func parseCmdText(text string, budget *cmdBudget) ([][]string, error) {
	commands := make([][]string, 0)
	args := make([]string, 0)
	var token strings.Builder
	var quote rune
	escaped := false
	started := false
	comment := false
	argumentCount := 0
	flushToken := func() error {
		if !started {
			return nil
		}
		value := token.String()
		if err := budget.take(1, len(value)); err != nil {
			return err
		}
		if argumentCount >= budget.limit.MaxArguments {
			return ErrLimit
		}
		args = append(args, value)
		argumentCount++
		token.Reset()
		started = false
		return nil
	}
	flushCommand := func() error {
		if err := flushToken(); err != nil {
			return err
		}
		if len(args) == 0 {
			return nil
		}
		if len(commands) >= budget.limit.MaxCommands {
			return ErrLimit
		}
		commands = append(commands, args)
		args = nil
		return nil
	}
	for _, r := range text {
		if err := budget.take(1, 0); err != nil {
			return nil, err
		}
		if comment {
			if r == '\n' {
				comment = false
				if err := flushCommand(); err != nil {
					return nil, err
				}
			}
			continue
		}
		if escaped {
			token.WriteRune(r)
			started = true
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			started = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				token.WriteRune(r)
			}
			started = true
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
			started = true
		case '#':
			if err := flushToken(); err != nil {
				return nil, err
			}
			comment = true
		case ';', '\n':
			if err := flushCommand(); err != nil {
				return nil, err
			}
		default:
			if unicode.IsSpace(r) {
				if err := flushToken(); err != nil {
					return nil, err
				}
			} else {
				token.WriteRune(r)
				started = true
			}
		}
	}
	if escaped || quote != 0 {
		return nil, ErrDimension
	}
	if err := flushCommand(); err != nil {
		return nil, err
	}
	return commands, nil
}

func (s *CmdSession) Execute(ctx context.Context, text string) (CmdResult, error) {
	if s == nil || s.closed || s.registry == nil {
		return CmdResult{}, ErrInvariant
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if len(text) > s.limits.MaxText || len(text) > s.limits.MaxBytes {
		return CmdResult{}, ErrLimit
	}
	budget := &cmdBudget{ctx: ctx, limit: s.limits}
	execution := &CmdExecution{session: s, budget: budget}
	result, err := execution.executeText(text, true)
	if err != nil {
		return CmdResult{}, err
	}
	return result, nil
}

func (e *CmdExecution) executeText(text string, record bool) (result CmdResult, resultErr error) {
	if e.depth > e.budget.limit.MaxDepth {
		return CmdResult{}, ErrLimit
	}
	commands, err := parseCmdText(text, e.budget)
	if err != nil {
		return CmdResult{}, err
	}
	retained := strings.TrimSuffix(text, "\n")
	historyAppend := false
	historyReserved := 0
	historyCommitted := false
	defer func() {
		if historyReserved != 0 && !historyCommitted {
			e.session.bytes -= historyReserved
		}
	}()
	if record && len(commands) != 0 && commandHistoryEligible(retained) {
		entries, historyErr := e.session.history.Entries(e.session.limits.MaxHistory)
		if historyErr != nil {
			return CmdResult{}, historyErr
		}
		historyAppend = true
		for index := max(0, len(entries)-10); index < len(entries); index++ {
			if entries[index] == retained {
				historyAppend = false
				break
			}
		}
		retainedBytes, retainedErr := e.session.retainedBytes()
		if retainedErr != nil {
			return CmdResult{}, retainedErr
		}
		if historyAppend && (len(entries) >= e.session.limits.MaxHistory || len(retained) > e.session.limits.MaxBytes-retainedBytes) {
			return CmdResult{}, ErrLimit
		}
		if historyAppend {
			historyReserved = len(retained)
			e.session.bytes += historyReserved
		}
	}
	var output strings.Builder
	combined := CmdResult{}
	for _, args := range commands {
		// CmdHistorySubstitution is deliberately a no-op in the tracked source;
		// bang-prefixed tokens therefore go through ordinary command dispatch.
		args, err = e.expandAliases(args)
		if err != nil {
			return CmdResult{}, err
		}
		if len(args) == 0 {
			continue
		}
		definition, found, lookupErr := e.session.registry.Definition(args[0])
		if lookupErr != nil {
			return CmdResult{}, lookupErr
		}
		if !found {
			return CmdResult{}, fmt.Errorf("%w: unknown command %q", ErrUnsupported, args[0])
		}
		if err = e.budget.take(1, 0); err != nil {
			return CmdResult{}, err
		}
		commandResult, commandErr := definition.Handler(e.budget.ctx, e, slices.Clone(args))
		if commandErr != nil {
			return CmdResult{}, commandErr
		}
		if len(commandResult.Output) > e.budget.limit.MaxText-output.Len() {
			return CmdResult{}, ErrLimit
		}
		if err = e.budget.take(0, len(commandResult.Output)); err != nil {
			return CmdResult{}, err
		}
		output.WriteString(commandResult.Output)
		combined.Status = commandResult.Status
		combined.Changed = combined.Changed || definition.Changes || commandResult.Changed
		combined.Quit = combined.Quit || commandResult.Quit
		if commandResult.Quit || commandResult.Status != 0 {
			break
		}
	}
	if record && len(commands) != 0 {
		added := false
		if added, err = e.session.history.Add(text); err != nil {
			return CmdResult{}, ErrInvariant
		}
		if added != historyAppend {
			return CmdResult{}, ErrInvariant
		}
		historyCommitted = true
	}
	combined.Output = output.String()
	return combined, nil
}

func (e *CmdExecution) expandAliases(args []string) ([]string, error) {
	seen := make(map[string]struct{})
	for depth := 0; len(args) != 0; depth++ {
		if depth >= e.budget.limit.MaxDepth {
			return nil, ErrLimit
		}
		if err := e.budget.take(1, 0); err != nil {
			return nil, err
		}
		alias, found, err := e.session.aliases.Alias(args[0])
		if err != nil || !found {
			return args, err
		}
		if _, duplicate := seen[args[0]]; duplicate {
			return nil, ErrInvariant
		}
		seen[args[0]] = struct{}{}
		if len(alias.Args)+len(args)-1 > e.budget.limit.MaxArguments {
			return nil, ErrLimit
		}
		next := slices.Clone(alias.Args)
		next = append(next, args[1:]...)
		args = next
	}
	return args, nil
}

func (s *CmdSession) installBuiltins() error {
	builtins := []CmdDefinition{
		{Name: "echo", Group: "Basic", Handler: cmdBuiltinEcho},
		{Name: "set", Group: "Basic", Handler: cmdBuiltinSet},
		{Name: "unset", Group: "Basic", Handler: cmdBuiltinUnset},
		{Name: "alias", Group: "Basic", Handler: cmdBuiltinAlias},
		{Name: "unalias", Group: "Basic", Handler: cmdBuiltinUnalias},
		{Name: "history", Group: "Basic", Handler: cmdBuiltinHistory},
		{Name: "source", Group: "Basic", Handler: cmdBuiltinSource},
		{Name: "help", Group: "Basic", Handler: cmdBuiltinHelp},
		{Name: "quit", Group: "Basic", Handler: cmdBuiltinQuit},
		{Name: "empty", Group: "Basic", Handler: cmdBuiltinEmpty, Changes: true},
	}
	for _, definition := range builtins {
		if err := s.registry.Add(definition); err != nil {
			return err
		}
	}
	return nil
}

func cmdBuiltinEcho(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	text := strings.Join(args[1:], " ") + "\n"
	if len(text) > e.budget.limit.MaxText {
		return CmdResult{}, ErrLimit
	}
	return CmdResult{Output: text}, nil
}

func cmdBuiltinSet(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) == 1 {
		var out strings.Builder
		for _, entry := range e.session.flags.Entries() {
			if err := e.budget.take(1, 0); err != nil {
				return CmdResult{}, err
			}
			line := entry.Name + "=" + entry.Value + "\n"
			if len(line) > e.budget.limit.MaxText-out.Len() {
				return CmdResult{}, ErrLimit
			}
			out.WriteString(line)
		}
		return CmdResult{Output: out.String()}, nil
	}
	if len(args) > 3 {
		return CmdResult{}, ErrDimension
	}
	value := ""
	if len(args) == 3 {
		value = args[2]
	}
	if err := e.budget.take(1, len(args[1])+len(value)); err != nil {
		return CmdResult{}, err
	}
	old, found, err := e.session.flags.Read(args[1])
	if err != nil {
		return CmdResult{}, err
	}
	oldBytes := 0
	if found {
		oldBytes = len(args[1]) + len(old)
	}
	newBytes := len(args[1]) + len(value)
	retained, err := e.session.retainedBytes()
	if err != nil {
		return CmdResult{}, err
	}
	if newBytes > e.session.limits.MaxBytes-retained+oldBytes {
		return CmdResult{}, ErrLimit
	}
	if err := e.session.flags.Update(args[1], &value); err != nil {
		return CmdResult{}, err
	}
	e.session.bytes += newBytes - oldBytes
	return CmdResult{}, nil
}

func cmdBuiltinUnset(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) < 2 {
		return CmdResult{}, ErrDimension
	}
	type removedFlag struct{ name, value string }
	removed := make([]removedFlag, 0, len(args)-1)
	seen := make(map[string]struct{}, len(args)-1)
	for _, name := range args[1:] {
		if err := e.budget.take(1, 0); err != nil {
			return CmdResult{}, err
		}
		old, found, readErr := e.session.flags.Read(name)
		if readErr != nil {
			return CmdResult{}, readErr
		}
		if found {
			if _, duplicate := seen[name]; duplicate {
				continue
			}
			seen[name] = struct{}{}
			removed = append(removed, removedFlag{name, old})
		}
	}
	for _, entry := range removed {
		deleted, err := e.session.flags.Delete(entry.name)
		if err != nil || !deleted {
			return CmdResult{}, ErrInvariant
		}
		e.session.bytes -= len(entry.name) + len(entry.value)
	}
	return CmdResult{}, nil
}

func cmdBuiltinAlias(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) == 1 {
		var out strings.Builder
		for _, alias := range e.session.aliases.Entries() {
			if err := e.budget.take(1, 0); err != nil {
				return CmdResult{}, err
			}
			line, err := FormatCommandAlias(alias)
			if err != nil || len(line) > e.budget.limit.MaxText-out.Len() {
				if err != nil {
					return CmdResult{}, err
				}
				return CmdResult{}, ErrLimit
			}
			out.WriteString(line)
		}
		return CmdResult{Output: out.String()}, nil
	}
	if len(args) == 2 {
		alias, found, err := e.session.aliases.Alias(args[1])
		if err != nil || !found {
			return CmdResult{}, err
		}
		line, err := FormatCommandAlias(alias)
		return CmdResult{Output: line}, err
	}
	if !cmdIdentifier(args[1]) {
		return CmdResult{}, ErrDimension
	}
	if err := e.budget.take(len(args), len(args[1])+cmdStringsBytes(args[2:])); err != nil {
		return CmdResult{}, err
	}
	old, found, err := e.session.aliases.Alias(args[1])
	if err != nil {
		return CmdResult{}, err
	}
	if found {
		_, _ = e.session.aliases.Delete(args[1])
	}
	oldBytes := 0
	if found {
		oldBytes = len(old.Name) + cmdStringsBytes(old.Args)
	}
	newBytes := len(args[1]) + cmdStringsBytes(args[2:])
	retained, err := e.session.retainedBytes()
	if err != nil {
		if found {
			_ = e.session.aliases.Add(old.Name, old.Args)
		}
		return CmdResult{}, err
	}
	if newBytes > e.session.limits.MaxBytes-retained+oldBytes {
		if found {
			_ = e.session.aliases.Add(old.Name, old.Args)
		}
		return CmdResult{}, ErrLimit
	}
	if err = e.session.aliases.Add(args[1], args[2:]); err != nil {
		if found {
			_ = e.session.aliases.Add(old.Name, old.Args)
		}
		return CmdResult{}, err
	}
	e.session.bytes += newBytes - oldBytes
	return CmdResult{}, nil
}

func cmdStringsBytes(values []string) int {
	total := 0
	for _, value := range values {
		if len(value) > int(^uint(0)>>1)-total {
			return int(^uint(0) >> 1)
		}
		total += len(value)
	}
	return total
}

func cmdBuiltinUnalias(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) < 2 {
		return CmdResult{}, ErrDimension
	}
	removed := make([]CommandAlias, 0, len(args)-1)
	seen := make(map[string]struct{}, len(args)-1)
	for _, name := range args[1:] {
		if err := e.budget.take(1, 0); err != nil {
			return CmdResult{}, err
		}
		old, found, aliasErr := e.session.aliases.Alias(name)
		if aliasErr != nil {
			return CmdResult{}, aliasErr
		}
		if found {
			if _, duplicate := seen[name]; duplicate {
				continue
			}
			seen[name] = struct{}{}
			removed = append(removed, old)
		}
	}
	for _, old := range removed {
		deleted, err := e.session.aliases.Delete(old.Name)
		if err != nil || !deleted {
			return CmdResult{}, ErrInvariant
		}
		e.session.bytes -= len(old.Name) + cmdStringsBytes(old.Args)
	}
	return CmdResult{}, nil
}

func cmdBuiltinHistory(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) > 2 {
		return CmdResult{}, ErrDimension
	}
	limit := e.session.limits.MaxHistory
	if len(args) == 2 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil || parsed < 0 {
			return CmdResult{}, ErrDimension
		}
		limit = parsed
	}
	entries, err := e.session.history.Entries(limit)
	if err != nil {
		return CmdResult{}, err
	}
	var out strings.Builder
	for _, entry := range entries {
		if err := e.budget.take(1, 0); err != nil {
			return CmdResult{}, err
		}
		if len(entry)+1 > e.budget.limit.MaxText-out.Len() {
			return CmdResult{}, ErrLimit
		}
		out.WriteString(entry)
		out.WriteByte('\n')
	}
	return CmdResult{Output: out.String()}, nil
}

func cmdBuiltinSource(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) != 2 {
		return CmdResult{}, ErrDimension
	}
	if e.depth >= e.budget.limit.MaxDepth {
		return CmdResult{}, ErrLimit
	}
	text, found := e.session.sources.Get2(args[1])
	if !found {
		return CmdResult{}, ErrInvariant
	}
	child := &CmdExecution{session: e.session, budget: e.budget, depth: e.depth + 1}
	return child.executeText(text, false)
}

func cmdBuiltinHelp(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) > 2 {
		return CmdResult{}, ErrDimension
	}
	definitions, err := e.session.registry.Definitions()
	if err != nil {
		return CmdResult{}, err
	}
	var out strings.Builder
	for _, definition := range definitions {
		if err := e.budget.take(1, 0); err != nil {
			return CmdResult{}, err
		}
		if len(args) == 2 && definition.Group != args[1] {
			continue
		}
		line := definition.Group + ": " + definition.Name + "\n"
		if len(line) > e.budget.limit.MaxText-out.Len() {
			return CmdResult{}, ErrLimit
		}
		out.WriteString(line)
	}
	return CmdResult{Output: out.String()}, nil
}

func cmdBuiltinQuit(_ context.Context, e *CmdExecution, args []string) (CmdResult, error) {
	if len(args) != 1 {
		return CmdResult{}, ErrDimension
	}
	e.session.quit = true
	return CmdResult{Quit: true}, nil
}

func cmdBuiltinEmpty(_ context.Context, _ *CmdExecution, args []string) (CmdResult, error) {
	if len(args) != 1 {
		return CmdResult{}, ErrDimension
	}
	return CmdResult{Changed: true}, nil
}

// CmdAutoCandidate is one compact cmdAuto.c option/value pair. Binary options
// have no numeric value and are toggled during option enumeration.
type CmdAutoCandidate struct {
	Parameter byte
	Value     float64
	Binary    bool
}

// CmdAutoChoice is one source configuration line. A line may hold one binary
// option or multiple numeric candidates.
type CmdAutoChoice struct {
	Candidates []CmdAutoCandidate
}

// ParseCmdAutoChoices parses the compact cmdAuto.c configuration syntax.
func ParseCmdAutoChoices(ctx context.Context, text string, limits CmdLimits) ([]CmdAutoChoice, error) {
	if err := validateCmdLimits(limits); err != nil {
		return nil, err
	}
	limits = limits.normalized()
	if len(text) > limits.MaxText || len(text) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if ctx == nil {
		ctx = context.Background()
	}
	budget := &cmdBudget{ctx: ctx, limit: limits}
	result := make([]CmdAutoChoice, 0)
	candidateCount := 0
	for _, raw := range strings.Split(text, "\n") {
		if err := budget.take(len(raw)+1, 0); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || len(raw) != 0 && unicode.IsSpace(rune(raw[0])) {
			continue
		}
		if len(result) >= limits.MaxCommands {
			return nil, ErrLimit
		}
		choice := CmdAutoChoice{}
		for index := 0; index < len(line); {
			for index < len(line) && unicode.IsSpace(rune(line[index])) {
				index++
			}
			if index == len(line) {
				break
			}
			parameter := line[index]
			index++
			if candidateCount >= limits.MaxArguments {
				return nil, ErrLimit
			}
			if parameter >= 'a' && parameter <= 'z' {
				if err := budget.take(1, 24); err != nil {
					return nil, err
				}
				choice.Candidates = append(choice.Candidates, CmdAutoCandidate{Parameter: parameter, Binary: true})
				candidateCount++
				continue
			}
			if parameter < 'A' || parameter > 'Z' || index == len(line) || !((line[index] >= '0' && line[index] <= '9') || line[index] == '.') {
				return nil, ErrDimension
			}
			start := index
			for index < len(line) && ((line[index] >= '0' && line[index] <= '9') || line[index] == '.') {
				index++
			}
			value, err := strconv.ParseFloat(line[start:index], 64)
			if err != nil {
				return nil, ErrDimension
			}
			if err := budget.take(1, 24); err != nil {
				return nil, err
			}
			choice.Candidates = append(choice.Candidates, CmdAutoCandidate{Parameter: parameter, Value: value})
			candidateCount++
		}
		if len(choice.Candidates) == 0 {
			return nil, ErrDimension
		}
		result = append(result, choice)
	}
	return result, nil
}

// CmdBuildInventory is the exact module.make translation-unit list.
func CmdBuildInventory() []string {
	return []string{"cmd.c", "cmdAlias.c", "cmdApi.c", "cmdAuto.c", "cmdFlag.c", "cmdHist.c", "cmdLoad.c", "cmdPlugin.c", "cmdStarter.c", "cmdUtils.c"}
}

// CmdSourceUnit records the audited definition surface of one tracked source.
type CmdSourceUnit struct {
	Source      string
	Definitions []string
	Complete    bool
	Gap         string
}

// CmdSourceInventory returns the complete audited definition inventory for the
// ten source units assigned to this port lane. Returned slices are caller-owned.
func CmdSourceInventory() []CmdSourceUnit {
	units := []CmdSourceUnit{
		{Source: "cmd.c", Definitions: []string{
			"Cmd_Init", "CmdCommandDispatchBasicById", "Cmd_End", "CmdCommandTime", "CmdCommandEcho", "CmdCommandQuit", "CmdCommandAbcrc", "CmdCommandHistory", "CmdCommandAlias", "CmdCommandUnalias", "CmdCommandHelp", "CmdCommandSource", "CmdCommandSetVariable", "CmdCommandUnsetVariable", "CmdCommandUndo", "CmdCommandRecall", "CmdCommandEmpty", "CmdCommandScanDir", "CmfFindNumber", "CnfDupFileUnzip", "CmdCommandRenameFiles", "CmdCommandLs", "CmdCommandScrGen", "CmdCommandSis", "CmdCommandMvsis", "Gia_ManGnuplotShow", "CmdCommandCapo", "CmdCommandStarter", "CmdCommandAutoTuner", "CmdCommandVersion",
		}, Gap: "process, filesystem, plotting, placement, external-program, frame-network, undo, and advanced source behavior remains outside the owned session"},
		{Source: "cmd.h", Definitions: []string{
			"Abc_Command", "Abc_Alias", "Cmd_CommandId_t", "Cmd_Init", "Cmd_End", "Cmd_CommandFuncType", "Cmd_CommandIsDefined", "Cmd_CommandAdd", "Cmd_CommandExecute", "Cmd_FlagReadByName", "Cmd_FlagDeleteByName", "Cmd_FlagUpdateValue", "Cmd_HistoryAddCommand", "Cmd_HistoryRead", "Cmd_HistoryWrite", "Cmd_HistoryPrint", "CmdCommandLoad",
		}, Gap: "native frame and pointer declarations are represented by concrete owned Go types"},
		{Source: "cmdApi.c", Definitions: []string{
			"Cmd_CommandNameToId", "Cmd_CommandIsDefined", "Cmd_CommandAdd", "Cmd_CommandHandleSpecial", "Cmd_CommandExecute",
		}, Gap: "mapped-network assertion directives and frame backup integration remain absent"},
		{Source: "cmdAuto.c", Definitions: []string{
			"Cmd_RunAutoTunerPrintOptions", "Cmd_RunAutoTunerEvalSimple", "Cmd_RunAutoTunerEval", "Cmd_RunAutoTunerEvalWorkerThread", "Cmd_DeriveConvertIntoString", "Cmd_DeriveOptionFromSettings", "Cmf_CreateOptions_rec", "Cmf_CreateOptions", "Cmf_IsSpace", "Cmf_IsLowerCaseChar", "Cmf_IsUpperCaseChar", "Cmf_IsDigit", "Cmd_ReadParamChoices", "Cmd_ReadFiles", "Cmd_RunAutoTuner",
		}, Gap: "configuration parsing is represented; SAT evaluation, AIG loading, threading, and solver-specific options remain absent"},
		{Source: "cmdInt.h", Definitions: []string{
			"MvCommand", "MvAlias", "CmdCommandAliasAdd", "CmdCommandAliasPrint", "CmdCommandAliasLookup", "CmdCommandAliasFree", "Cmd_CommandNameToId", "CmdCommandDispatch", "CmdCommandDispatchBasicById", "IoCommandDispatchById", "AbcCommandDispatchById", "CmdSplitLine", "CmdApplyAlias", "CmdHistorySubstitution", "CmdFileOpen", "CmdFreeArgv", "CmdAddToArgv", "CmdCommandFree", "CmdCommandPrint", "CmdPrintTable",
		}, Gap: "native frame, file, table, and dispatcher pointer surfaces are adapted to owned session operations"},
		{Source: "cmdLoad.c", Definitions: []string{
			"CmdCommandLoad", "CmdCollectFileNames", "Load_Init", "Load_End",
		}, Gap: "external executable discovery and process execution are intentionally unsupported"},
		{Source: "cmdPlugin.c", Definitions: []string{
			"Abc_GetBinaryName", "Abc_ManReadFile", "Abc_ManReadBinary", "Abc_ManReadInteger", "Abc_ManReadStatus", "Abc_ManExpandCex", "textToBin", "Abc_ManReadAig", "Cmd_CommandAbcPlugIn", "Cmd_CommandAbcLoadPlugIn",
		}, Gap: "dynamic loading, process exchange, AIG decoding, and counterexample expansion remain absent"},
		{Source: "cmdStarter.c", Definitions: []string{
			"Abc_RunThread", "Cmd_RunStarter",
		}, Gap: "external starter processes and threads remain absent"},
		{Source: "cmdUtils.c", Definitions: []string{
			"CmdCommandPrintCompare", "cmdCheckShellEscape", "CmdCommandDispatch", "CmdSplitLine", "CmdApplyAlias", "CmdHistorySubstitution", "CmdFileOpen", "CmdFreeArgv", "CmdDupArgv", "CmdAddToArgv", "CmdCommandFree", "CmdCommandPrint", "CmdNamePrintCompare", "CmdPrintTable",
		}, Gap: "shell escape and filesystem lookup remain unsupported; owned parser, dispatch, aliasing, and deterministic diagnostics are represented"},
		{Source: "module.make", Definitions: CmdBuildInventory(), Complete: true},
	}
	result := make([]CmdSourceUnit, len(units))
	for index, unit := range units {
		result[index] = unit
		result[index].Definitions = slices.Clone(unit.Definitions)
	}
	return result
}
