package abc

// Command history from src/base/cmd/cmdHist.c.

import (
	"bytes"
	"os"
	"slices"
	"strings"
)

type CommandHistoryOptions struct {
	Batch, MaxEntries, MaxCommandBytes int
}

type CommandHistory struct {
	entries         []string
	batch           bool
	maxEntries      int
	maxCommandBytes int
}

func NewCommandHistory(options CommandHistoryOptions) (*CommandHistory, error) {
	if options.Batch < 0 || options.Batch > 1 || options.MaxEntries < 0 || options.MaxCommandBytes < 0 {
		return nil, ErrDimension
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	if options.MaxCommandBytes == 0 {
		options.MaxCommandBytes = 1 << 20
	}
	return &CommandHistory{batch: options.Batch != 0, maxEntries: options.MaxEntries, maxCommandBytes: options.MaxCommandBytes}, nil
}

func commandHistoryEligible(command string) bool {
	if len(command) <= 3 || command[len(command)-1] == '?' {
		return false
	}
	for _, prefix := range []string{"set", "unset", "time", "quit", "alias", "source abc.rc", "source ..\\abc.rc", "history", "hi "} {
		if strings.HasPrefix(command, prefix) {
			return false
		}
	}
	return command != "hi"
}

// Add applies the source's admission rules. A duplicate among the last ten
// entries is moved to the end; older duplicates are retained and a new entry is
// appended.
func (h *CommandHistory) Add(command string) (bool, error) {
	if h == nil {
		return false, ErrInvariant
	}
	if len(command) > h.maxCommandBytes {
		return false, ErrLimit
	}
	if h.batch {
		return false, nil
	}
	if strings.HasSuffix(command, "\n") {
		command = command[:len(command)-1]
	}
	if !commandHistoryEligible(command) {
		return false, nil
	}
	start := max(0, len(h.entries)-10)
	for index := start; index < len(h.entries); index++ {
		if h.entries[index] != command {
			continue
		}
		entry := h.entries[index]
		copy(h.entries[index:], h.entries[index+1:])
		h.entries[len(h.entries)-1] = entry
		return false, nil
	}
	if len(h.entries) >= h.maxEntries {
		return false, ErrLimit
	}
	h.entries = append(h.entries, command)
	return true, nil
}

func (h *CommandHistory) Entries(limit int) ([]string, error) {
	if h == nil {
		return nil, ErrInvariant
	}
	if limit < 0 {
		return nil, ErrDimension
	}
	start := max(0, len(h.entries)-limit)
	return slices.Clone(h.entries[start:]), nil
}

func (h *CommandHistory) Marshal(limit int) ([]byte, error) {
	entries, err := h.Entries(limit)
	if err != nil {
		return nil, err
	}
	var result bytes.Buffer
	for _, entry := range entries {
		result.WriteString(entry)
		result.WriteByte('\n')
	}
	return result.Bytes(), nil
}

// Load replaces an empty history with newline-delimited entries, matching the
// source read precondition. It is atomic and does not apply Add's command
// filters or last-ten deduplication.
func (h *CommandHistory) Load(data []byte) error {
	if h == nil {
		return ErrInvariant
	}
	if len(h.entries) != 0 {
		return ErrInvariant
	}
	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) != 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > h.maxEntries {
		return ErrLimit
	}
	entries := make([]string, len(lines))
	for index, line := range lines {
		if len(line) > h.maxCommandBytes {
			return ErrLimit
		}
		entries[index] = string(line)
	}
	h.entries = entries
	return nil
}

func (h *CommandHistory) Format(limit int) (string, error) {
	data, err := h.Marshal(limit)
	if err != nil {
		return "", err
	}
	return "================== Command history ==================\n" + string(data) + "=====================================================\n", nil
}

func (h *CommandHistory) ReadFile(path string) error {
	if h == nil {
		return ErrInvariant
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return h.Load(data)
}

func (h *CommandHistory) WriteFile(path string, limit int) error {
	data, err := h.Marshal(limit)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o666)
}
