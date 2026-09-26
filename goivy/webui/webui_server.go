// Package webui provides an HTTP server with SSE push for Ivy's
// interactive verification UI.  It replaces the Python Tcl/Tk frontend
// with a Go HTTP + JSON backend, keeping the Cytoscape.js graph library
// on the browser side.
package webui

import (
	"encoding/json"
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"html"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// Server is the HTTP backend for the Ivy verification UI.
type Server struct {
	cfg     *goivy.Config
	addr    string
	backend Backend
	mux     *http.ServeMux
}

type tutorialSearchDocument struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Body  string `json:"body"`
}

var (
	tutorialTitleRE   = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	tutorialScriptRE  = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	tutorialStyleRE   = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	tutorialCommentRE = regexp.MustCompile(`(?is)<!--.*?-->`)
	tutorialTagRE     = regexp.MustCompile(`(?is)<[^>]+>`)
)

// NewServer creates a Server that will listen on addr (e.g. ":8080").
// If backend is nil, a default GoBackend is used.
func NewServer(cfg *goivy.Config, addr string, backend ...Backend) *Server {
	var be Backend
	if len(backend) > 0 && backend[0] != nil {
		be = backend[0]
	} else {
		be = NewGoBackend(cfg)
	}
	s := &Server{
		cfg:     cfg,
		addr:    addr,
		backend: be,
		mux:     http.NewServeMux(),
	}
	s.mux.HandleFunc("/", s.handleIndex)

	// live directory:
	s.mux.HandleFunc("/static/", s.handleStatic)

	// or embedded static version, makes ivyweb
	// runnable from anywhere: (but currently breaks tutorial on python backend!)
	//s.mux.Handle("/static/", http.FileServer(http.FS(staticContent)))

	s.mux.HandleFunc("/api/", s.handleAPI)
	// Note: no proxy endpoint — the BiB iframe loads external URLs directly.
	return s
}

// Start begins serving HTTP.  It blocks until the server is shut down.
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}
	log.Printf("webui: listening on %s", s.addr)
	return srv.ListenAndServe()
}

// ServeHTTP implements http.Handler so the server can be used in tests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// staticDir returns the absolute path to the static/ directory that lives
// next to this source file.  It first tries runtime.Caller (works when
// running from the source tree), then falls back to the current working
// directory.
func staticDir() string {
	// Try the directory containing this Go source file.
	_, srcFile, _, ok := runtime.Caller(0)
	if ok {
		dir := filepath.Join(filepath.Dir(srcFile), "static")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	// Fallback: cwd-relative.
	if wd, err := os.Getwd(); err == nil {
		dir := filepath.Join(wd, "webui", "static")
		if info, err2 := os.Stat(dir); err2 == nil && info.IsDir() {
			return dir
		}
		// Maybe we are already inside the webui directory.
		dir = filepath.Join(wd, "static")
		if info, err2 := os.Stat(dir); err2 == nil && info.IsDir() {
			return dir
		}
	}
	return "static" // last resort relative path
}

// handleIndex serves the main SPA page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	indexPath := filepath.Join(staticDir(), "index.html")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		// Fall back to inline HTML if index.html cannot be read.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write(data)
}

// handleStatic serves static assets (JS, CSS, tutorial) from the static/ directory.
// Tutorial files use no-store during development so link/content fixes are
// immediately visible in the browser tutorial pane.
// JS/CSS get no-cache during development so the browser always fetches the latest.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, "/static/tutorial/") {
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	} else {
		// Dev assets (JS, CSS): always revalidate
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
	dir := staticDir()
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(dir)))
	fs.ServeHTTP(w, r)
}

// handleAPI routes /api/* requests to the correct handler.
func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")

	w.Header().Set("Content-Type", "application/json")

	// POST /api/session/new
	if len(parts) == 2 && parts[0] == "session" && parts[1] == "new" {
		s.apiNewSession(w, r)
		return
	}

	if len(parts) == 2 && parts[0] == "tutorial" && parts[1] == "search-docs" {
		s.apiTutorialSearchDocs(w, r)
		return
	}

	// All remaining routes require a session id at parts[1].
	if len(parts) < 3 || parts[0] != "session" {
		writeErr(w, http.StatusNotFound, "unknown api endpoint")
		return
	}

	sid := parts[1]
	rest := strings.Join(parts[2:], "/")
	switch rest {
	case "load":
		s.apiLoad(w, r, sid)
	case "action":
		s.apiAction(w, r, sid)
	case "arg":
		s.apiARG(w, r, sid)
	case "menus":
		s.apiMenus(w, r, sid)
	case "concept":
		s.apiConcept(w, r, sid)
	case "concept/split":
		s.apiConceptSplit(w, r, sid)
	case "concept/empty":
		s.apiConceptEmpty(w, r, sid)
	case "concept/remove":
		s.apiConceptRemove(w, r, sid)
	case "concept/undo":
		s.apiConceptUndo(w, r, sid)
	case "concept/materialize":
		s.apiConceptMaterialize(w, r, sid)
	case "concept/reset":
		s.apiConceptReset(w, r, sid)
	case "concept/diagram":
		s.apiConceptDiagram(w, r, sid)
	case "toggles":
		s.apiToggles(w, r, sid)
	case "check":
		s.apiCheck(w, r, sid)
	case "proof":
		s.apiProof(w, r, sid)
	case "concept/projection":
		s.apiConceptProjection(w, r, sid)
	case "arg/action":
		s.apiArgAction(w, r, sid)
	case "arg/cti":
		s.apiCTIARG(w, r, sid)
	case "proof/action":
		s.apiProofAction(w, r, sid)
	case "save":
		s.apiSave(w, r, sid)
	case "events":
		s.apiEvents(w, r, sid)
	default:
		writeErr(w, http.StatusNotFound, "unknown api endpoint")
	}
}

func (s *Server) apiTutorialSearchDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	docs, err := tutorialSearchDocs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string][]tutorialSearchDocument{"documents": docs})
}

func tutorialSearchDocs() ([]tutorialSearchDocument, error) {
	root := filepath.Join(staticDir(), "tutorial")
	var docs []tutorialSearchDocument
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || strings.ToLower(filepath.Ext(path)) != ".html" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		url := "/static/tutorial/" + rel
		title, body := tutorialSearchText(string(data), rel)
		if body == "" {
			return nil
		}
		docs = append(docs, tutorialSearchDocument{
			ID:    url,
			Title: title,
			URL:   url,
			Body:  body,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].URL < docs[j].URL
	})
	return docs, nil
}

func tutorialSearchText(src, fallbackTitle string) (string, string) {
	title := fallbackTitle
	if match := tutorialTitleRE.FindStringSubmatch(src); len(match) > 1 {
		if text := normalizeTutorialSearchText(match[1]); text != "" {
			title = text
		}
	}
	text := tutorialScriptRE.ReplaceAllString(src, " ")
	text = tutorialStyleRE.ReplaceAllString(text, " ")
	text = tutorialCommentRE.ReplaceAllString(text, " ")
	text = tutorialTagRE.ReplaceAllString(text, " ")
	body := normalizeTutorialSearchText(text)
	return title, body
}

func normalizeTutorialSearchText(src string) string {
	text := html.UnescapeString(src)
	return strings.Join(strings.Fields(text), " ")
}

// writeJSON marshals v as JSON to w.
func writeJSON(w http.ResponseWriter, v interface{}) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Printf("webui: json encode: %v", err)
	}
}

// writeErr writes a JSON error response.
func writeErr(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	writeJSON(w, map[string]string{"error": msg})
}

// writeBackend writes pre-serialized backend JSON to the response.
func writeBackend(w http.ResponseWriter, data []byte) {
	w.Write(data)
}

// indexHTML is a minimal SPA shell served at "/".
const indexHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Ivy Verification UI</title>
<style>
body { font-family: sans-serif; margin: 0; padding: 1em; }
#cy { width: 100%; height: 600px; border: 1px solid #ccc; }
</style>
</head>
<body>
<h1>Ivy Verification UI</h1>
<div id="cy"></div>
<script src="https://unpkg.com/cytoscape@3/dist/cytoscape.min.js"></script>
</body>
</html>
`
