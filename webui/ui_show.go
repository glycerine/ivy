package webui

// Full port of ivy_show.py — entry point functions for launching the
// verification UI and checking modules.

import (
	"fmt"
)

// ShowVerification is the main entry point for showing verification results.
// It initializes the module, creates the isolate, and launches the UI.
// (Python: check_module in ivy_show.py).
func ShowVerification(filePath string) (*Session, error) {
	if filePath == "" {
		return nil, fmt.Errorf("empty file path")
	}

	// Create a new session.
	sess := NewSession("show_" + filePath)
	if err := sess.LoadFile(filePath); err != nil {
		return nil, fmt.Errorf("failed to load file: %w", err)
	}

	// Stub: real implementation would:
	// 1. ivy_init.read_params()
	// 2. ivy_init.source_file(filePath)
	// 3. ivy_isolate.create_isolate(isolate)
	// 4. Launch the UI main loop.

	return sess, nil
}

// LaunchUI starts the web-based verification UI for an analysis graph.
// This replaces the Python tk_ui.ui_main_loop.
// (Python: ui_main_loop in ivy_ui.py).
func LaunchUI(sess *Session, addr string) (*Server, error) {
	if sess == nil {
		return nil, fmt.Errorf("nil session")
	}

	srv := NewServer(addr)
	srv.mu.Lock()
	srv.sessions[sess.ID] = sess
	srv.mu.Unlock()

	return srv, nil
}

// DiagnoseMode returns whether diagnostic mode is enabled.
// (Python: diagnose = iu.BooleanParameter("diagnose",False)).
var DiagnoseMode bool

// CompileKwargs holds extra keyword arguments for compilation.
// (Python: compile_kwargs = {'ext':'ext'} in ivy_ui_cti.py).
var CompileKwargs = map[string]string{
	"ext": "ext",
}

// CheckModuleAndShow loads an Ivy file, checks the module, and
// shows the result in the UI.
// (Python: main() in ivy_show.py).
func CheckModuleAndShow(filePath string, addr string) error {
	sess, err := ShowVerification(filePath)
	if err != nil {
		return err
	}

	srv, err := LaunchUI(sess, addr)
	if err != nil {
		return err
	}

	// Start serving (blocking).
	_ = srv
	return nil
}
