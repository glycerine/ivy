package webui

// Full port of ivy_ui_util.py — utility functions and types for the
// verification UI, adapted from Tk to web-based interaction.

import (
	"fmt"
	"strings"
)

// CenterWindow computes centered position for a window of given size
// on a screen of given dimensions. Returns (x, y) offset.
// (Python: center_window).
func CenterWindow(screenW, screenH, winW, winH int) (int, int) {
	x := screenW/2 - winW/2
	y := screenH/2 - winH/2
	return x, y
}

// CenterWindowOnWindow computes centered position of a window relative
// to another window. Returns (x, y) offset.
// (Python: center_window_on_window).
func CenterWindowOnWindow(parentX, parentY, parentW, parentH, winW, winH int) (int, int) {
	xc := parentX + parentW/2
	yc := parentY + parentH/2
	x := xc - winW/2
	y := yc - winH/2
	return x, y
}

// DialogResult represents the result of a UI dialog interaction.
type DialogResult struct {
	Confirmed bool        `json:"confirmed"`
	Value     interface{} `json:"value,omitempty"`
	Index     int         `json:"index,omitempty"`
	Text      string      `json:"text,omitempty"`
}

// OkDialogRequest represents a request to show an OK dialog.
type OkDialogRequest struct {
	Message string `json:"message"`
}

// EntryDialogRequest represents a request to show a text entry dialog.
type EntryDialogRequest struct {
	Message      string `json:"message"`
	CommandLabel string `json:"command_label,omitempty"`
	InitialValue string `json:"initial_value,omitempty"`
}

// TextDialogRequest represents a request to show a text display dialog.
type TextDialogRequest struct {
	Message      string `json:"message"`
	Text         string `json:"text"`
	CommandLabel string `json:"command_label,omitempty"`
	HasCancel    bool   `json:"has_cancel"`
}

// ListboxDialogRequest represents a request to show a listbox selection dialog.
type ListboxDialogRequest struct {
	Message  string   `json:"message"`
	Items    []string `json:"items"`
	Multiple bool     `json:"multiple"`
}

// OkCancelDialogRequest represents a request to show an OK/Cancel dialog.
type OkCancelDialogRequest struct {
	Message string `json:"message"`
}

// ButtonsDialogRequest represents a request to show a dialog with custom buttons.
type ButtonsDialogRequest struct {
	Message string           `json:"message"`
	Buttons []ButtonOption   `json:"buttons"`
}

// ButtonOption is a labeled button in a dialog.
type ButtonOption struct {
	Label  string `json:"label"`
	Action string `json:"action"`
}

// IntDialogRequest represents a request for integer input.
type IntDialogRequest struct {
	Message  string `json:"message"`
	MinVal   *int   `json:"min_val,omitempty"`
	MaxVal   *int   `json:"max_val,omitempty"`
	InitVal  *int   `json:"init_val,omitempty"`
}

// ConvertToInt parses and validates an integer string with optional bounds.
// (Python: _convert_to_int).
func ConvertToInt(s string, minVal, maxVal *int) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	if err != nil {
		return 0, fmt.Errorf("entered value %q is not an integer", s)
	}
	if minVal != nil && v < *minVal {
		return 0, fmt.Errorf("entered value %d is below minimum %d", v, *minVal)
	}
	if maxVal != nil && v > *maxVal {
		return 0, fmt.Errorf("entered value %d is above maximum %d", v, *maxVal)
	}
	return v, nil
}

// FileBrowserState represents the state of a source file browser
// (Python: class FileBrowser, adapted for web).
type FileBrowserState struct {
	Filename    string   `json:"filename"`
	Content     string   `json:"content"`
	HighlightLine int    `json:"highlight_line"`
	Lines       []string `json:"-"`
}

// NewFileBrowserState creates a new file browser state.
func NewFileBrowserState() *FileBrowserState {
	return &FileBrowserState{}
}

// SetFile loads a file and optionally highlights a line.
func (fb *FileBrowserState) SetFile(filename string, content string, lineno int) {
	fb.Filename = filename
	fb.Content = content
	fb.Lines = strings.Split(content, "\n")
	fb.HighlightLine = lineno
}

// GetLine returns a specific line (1-indexed) from the current file.
func (fb *FileBrowserState) GetLine(lineno int) string {
	if lineno < 1 || lineno > len(fb.Lines) {
		return ""
	}
	return fb.Lines[lineno-1]
}

// LineCount returns the number of lines in the current file.
func (fb *FileBrowserState) LineCount() int {
	return len(fb.Lines)
}

// MenuBarDef represents a complete menu bar definition for the web UI.
type MenuBarDef struct {
	Menus []MenuDef `json:"menus"`
}

// BuildMenuBar constructs a MenuBarDef from a list of menu definitions.
func BuildMenuBar(menus []MenuDef) *MenuBarDef {
	return &MenuBarDef{Menus: menus}
}

// RunContext is a context manager analog for handling errors during
// verification operations. In Go, this is implemented as a helper
// that wraps operations and catches panics.
// (Python: class RunContext).
type RunContext struct {
	OnError func(err error)
}

// NewRunContext creates a new RunContext.
func NewRunContext(onError func(err error)) *RunContext {
	return &RunContext{OnError: onError}
}

// Run executes a function within the run context, recovering from panics.
func (rc *RunContext) Run(fn func() error) error {
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case error:
					err = v
				case string:
					err = fmt.Errorf("%s", v)
				default:
					err = fmt.Errorf("panic: %v", v)
				}
			}
		}()
		err = fn()
	}()
	if err != nil && rc.OnError != nil {
		rc.OnError(err)
	}
	return err
}
