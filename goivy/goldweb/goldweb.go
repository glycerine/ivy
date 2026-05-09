package main

// goldweb is the first browser-conformance harness for Go Ivy.
//
// The intentional narrow waist is WASI stdout/stderr:
//   - the browser worker runs Go Ivy compiled to wasip1 and Z3 compiled to wasm;
//   - every WASI fd_write byte for stdout and stderr is sent to this server;
//   - the server compares XTRACE lines against Python ivy_check, while still
//     retaining and exposing non-XTRACE output for diagnosis.
//
// The Go Ivy wasm command boundary is deliberately small and not the
// cmd/goivy_check CLI. The CLI is only documentation for parameter semantics.
// The browser wasm module is expected to export:
//   - goivy_check_alloc(n uint32) uint32
//   - goivy_check_free(ptr uint32, n uint32)
//   - goivy_check_run(specPtr, specLen, metaPtr, metaLen uint32) int32
//
// meta is JSON: {"filename":"browser_input.ivy","params":{"isolate":"x"}}.
// The command writes its normal output to stdout/stderr; goldweb does not use a
// special trace callback.

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/glycerine/ivy/goivy/xtracer"
	"github.com/gorilla/websocket"
)

const (
	writeWait    = 10 * time.Minute
	pongWait     = 60 * time.Second
	pingPeriod   = 30 * time.Second
	readLimit    = 64 << 20
	sendQueueLen = 8192
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type app struct {
	listen     string
	goivyRoot  string
	webvueDir  string
	staticDir  string
	workerDir  string
	goivyWasm  string
	ivyCheck   string
	httpServer *http.Server

	mu      sync.Mutex
	clients map[*wsClient]bool
	jobs    map[string]*job
}

type wsClient struct {
	app    *app
	conn   *websocket.Conn
	send   chan []byte
	done   chan struct{}
	closed sync.Once
}

type wsEnvelope struct {
	Type     string            `json:"type"`
	ID       string            `json:"id,omitempty"`
	Command  string            `json:"command,omitempty"`
	Filename string            `json:"filename,omitempty"`
	Spec     string            `json:"spec,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
	FD       int               `json:"fd,omitempty"`
	Data     string            `json:"data,omitempty"`
	Code     int               `json:"code,omitempty"`
	Error    string            `json:"error,omitempty"`
	Status   string            `json:"status,omitempty"`
	Message  string            `json:"message,omitempty"`
}

type goivyCheckRequest struct {
	Filename string            `json:"filename"`
	Spec     string            `json:"spec"`
	Params   map[string]string `json:"params"`
}

type job struct {
	id       string
	filename string
	spec     string
	params   map[string]string
	app      *app
	client   *wsClient

	pyLines      chan lineEvent
	browserLines chan lineEvent
	browserBuf   streamLineBuffer
	compareDone  chan struct{}

	mu           sync.Mutex
	status       string
	message      string
	xtraceCount  int
	browserCode  int
	pyErr        string
	tail         []string
	cancelPy     context.CancelFunc
	closeBrowser sync.Once
}

type lineEvent struct {
	FD   int
	Line string
}

type streamLineBuffer struct {
	partial string
	fd      int
}

type jobSnapshot struct {
	ID          string            `json:"id"`
	Filename    string            `json:"filename"`
	Params      map[string]string `json:"params,omitempty"`
	Status      string            `json:"status"`
	Message     string            `json:"message,omitempty"`
	XTraceCount int               `json:"xtrace_count"`
	BrowserCode int               `json:"browser_code,omitempty"`
	PythonError string            `json:"python_error,omitempty"`
	Tail        []string          `json:"tail,omitempty"`
}

func main() {
	rootDefault := defaultGoivyRoot()
	listen := flag.String("listen", "127.0.0.1:8998", "address for the goldweb HTTP/WebSocket server")
	root := flag.String("root", rootDefault, "goivy source root")
	goivyWasm := flag.String("goivy-wasm", filepath.Join(rootDefault, "webvue", "static", "goivy-check-wasip1.wasm"), "Go Ivy wasip1 wasm file to serve at /goivy-check.wasm")
	ivyCheck := flag.String("ivy-check", "ivy_check", "Python ivy_check executable")
	flag.Parse()

	a := newApp(*listen, *root, *goivyWasm, *ivyCheck)
	if err := a.listenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newApp(listen, root, goivyWasm, ivyCheck string) *app {
	root = filepath.Clean(root)
	webvueDir := filepath.Join(root, "webvue")
	return &app{
		listen:    listen,
		goivyRoot: root,
		webvueDir: webvueDir,
		staticDir: filepath.Join(webvueDir, "static"),
		workerDir: filepath.Join(webvueDir, "src", "workers"),
		goivyWasm: goivyWasm,
		ivyCheck:  ivyCheck,
		clients:   make(map[*wsClient]bool),
		jobs:      make(map[string]*job),
	}
}

func (a *app) listenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.serveIndex)
	mux.HandleFunc("/ws", a.serveWS)
	mux.HandleFunc("/goivy_check", a.handleGoivyCheck)
	mux.HandleFunc("/jobs/", a.handleJob)
	mux.HandleFunc("/goivy-check.wasm", serveFile(a.goivyWasm, "application/wasm"))
	mux.HandleFunc("/z3-471-api.js", serveFile(filepath.Join(a.staticDir, "z3-471-api.js"), "text/javascript; charset=utf-8"))
	mux.HandleFunc("/z3-471-api.wasm", serveFile(filepath.Join(a.staticDir, "z3-471-api.wasm"), "application/wasm"))
	mux.HandleFunc("/src/workers/smtZ3Imports.js", serveFile(filepath.Join(a.workerDir, "smtZ3Imports.js"), "text/javascript; charset=utf-8"))

	a.httpServer = &http.Server{
		Addr:              a.listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("goldweb listening on http://%s", a.listen)
	log.Printf("POST JSON to http://%s/goivy_check to start a browser-backed check", a.listen)
	return a.httpServer.ListenAndServe()
}

func (a *app) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = io.WriteString(w, indexHTML)
}

func serveFile(path, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if _, err := os.Stat(path); err != nil {
			http.Error(w, fmt.Sprintf("missing asset %s: %v", path, err), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, path)
	}
}

func (a *app) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	c := &wsClient{
		app:  a,
		conn: conn,
		send: make(chan []byte, sendQueueLen),
		done: make(chan struct{}),
	}
	a.register(c)
	go c.writePump()
	c.readPump()
}

func (a *app) register(c *wsClient) {
	a.mu.Lock()
	a.clients[c] = true
	n := len(a.clients)
	a.mu.Unlock()
	log.Printf("browser websocket connected: %s; clients=%d", c.conn.RemoteAddr(), n)
	_ = c.sendJSON(wsEnvelope{Type: "hello", Status: "ready", Message: "goldweb connected"})
}

func (a *app) unregister(c *wsClient) {
	c.close()
	a.mu.Lock()
	delete(a.clients, c)
	n := len(a.clients)
	a.mu.Unlock()
	log.Printf("browser websocket disconnected: %s; clients=%d", c.conn.RemoteAddr(), n)
}

func (a *app) firstClient() *wsClient {
	a.mu.Lock()
	defer a.mu.Unlock()
	for c := range a.clients {
		return c
	}
	return nil
}

func (a *app) addJob(j *job) {
	a.mu.Lock()
	a.jobs[j.id] = j
	a.mu.Unlock()
}

func (a *app) lookupJob(id string) *job {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.jobs[id]
}

func (a *app) handleWSMessage(c *wsClient, data []byte) {
	var msg wsEnvelope
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("bad websocket JSON from browser: %v", err)
		return
	}
	switch msg.Type {
	case "ready":
		return
	case "stream":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("stream for unknown job %q", msg.ID)
			return
		}
		j.addBrowserData(msg.FD, msg.Data)
	case "done":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("done for unknown job %q", msg.ID)
			return
		}
		j.setBrowserDone(msg.Code)
	case "error":
		j := a.lookupJob(msg.ID)
		if j == nil {
			log.Printf("error for unknown job %q: %s", msg.ID, msg.Error)
			return
		}
		j.fail("browser error: " + msg.Error)
		j.setBrowserDone(msg.Code)
	default:
		log.Printf("unknown websocket message type %q", msg.Type)
	}
}

func (a *app) handleGoivyCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req goivyCheckRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<20)).Decode(&req); err != nil {
		http.Error(w, "bad JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Spec) == "" {
		http.Error(w, "missing spec", http.StatusBadRequest)
		return
	}
	if req.Filename == "" {
		req.Filename = "browser_input.ivy"
	}
	if !strings.HasSuffix(req.Filename, ".ivy") {
		req.Filename += ".ivy"
	}
	if req.Params == nil {
		req.Params = make(map[string]string)
	}

	c := a.firstClient()
	if c == nil {
		http.Error(w, "no browser websocket client connected; open / first", http.StatusServiceUnavailable)
		return
	}

	j := newJob(a, c, req)
	a.addJob(j)
	go j.compareLoop()
	go j.runPythonIvyCheck()

	if err := c.sendJSON(wsEnvelope{
		Type:     "command",
		ID:       j.id,
		Command:  "goivy_check",
		Filename: j.filename,
		Spec:     j.spec,
		Params:   j.params,
	}); err != nil {
		j.fail("send browser command: " + err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	writeJSON(w, http.StatusAccepted, j.snapshot())
}

func (a *app) handleJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/jobs/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	j := a.lookupJob(id)
	if j == nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, j.snapshot())
}

func newJob(a *app, c *wsClient, req goivyCheckRequest) *job {
	return &job{
		id:           newID(),
		filename:     req.Filename,
		spec:         req.Spec,
		params:       cloneStringMap(req.Params),
		app:          a,
		client:       c,
		pyLines:      make(chan lineEvent, 8192),
		browserLines: make(chan lineEvent, 8192),
		compareDone:  make(chan struct{}),
		status:       "running",
	}
}

func (j *job) runPythonIvyCheck() {
	tmp, err := os.MkdirTemp("", "goldweb-ivy-*")
	if err != nil {
		j.fail("create temp dir: " + err.Error())
		close(j.pyLines)
		return
	}
	defer os.RemoveAll(tmp)

	filename := filepath.Base(j.filename)
	if filename == "." || filename == string(filepath.Separator) {
		filename = "browser_input.ivy"
	}
	path := filepath.Join(tmp, filename)
	if err := os.WriteFile(path, []byte(j.spec), 0o600); err != nil {
		j.fail("write temp spec: " + err.Error())
		close(j.pyLines)
		return
	}

	args := paramsAsArgs(j.params)
	args = append(args, path)
	ctx, cancel := context.WithCancel(context.Background())
	j.setPythonCancel(cancel)
	defer func() {
		j.setPythonCancel(nil)
		cancel()
	}()
	cmd := exec.CommandContext(ctx, j.app.ivyCheck, args...)
	cmd.Dir = j.app.goivyRoot

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		j.fail("start python ivy_check: " + err.Error())
		_ = pw.Close()
		close(j.pyLines)
		return
	}

	go func() {
		waitErr := cmd.Wait()
		if waitErr != nil {
			j.setPythonErr(waitErr.Error())
		}
		_ = pw.Close()
	}()

	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 0, 16<<20), 1<<30)
	for scanner.Scan() {
		j.pyLines <- lineEvent{FD: 1, Line: scanner.Text() + "\n"}
	}
	if err := scanner.Err(); err != nil {
		j.setPythonErr("scan python output: " + err.Error())
	}
	close(j.pyLines)
}

func (j *job) compareLoop() {
	defer close(j.compareDone)
	for {
		browser, browserOK := j.nextXTrace("browser", j.browserLines)
		python, pythonOK := j.nextXTrace("python", j.pyLines)

		switch {
		case !browserOK && !pythonOK:
			j.finishMatched()
			return
		case !browserOK:
			j.mismatch("browser ended before python", "", python)
			return
		case !pythonOK:
			j.mismatch("python ended before browser", browser, "")
			return
		}

		if browser != python {
			j.mismatch("xtrace divergence", browser, python)
			return
		}
		j.mu.Lock()
		j.xtraceCount++
		j.mu.Unlock()
	}
}

func (j *job) nextXTrace(side string, ch <-chan lineEvent) (string, bool) {
	for ev := range ch {
		line := xtracer.NormalizeLine(ev.Line)
		j.remember(side, line)
		if strings.HasPrefix(line, "XTRACE:") {
			return line, true
		}
	}
	return "", false
}

func (j *job) addBrowserData(fd int, data string) {
	lines := j.browserBuf.add(fd, data)
	for _, line := range lines {
		j.browserLines <- line
	}
}

func (j *job) setBrowserDone(code int) {
	j.mu.Lock()
	j.browserCode = code
	j.mu.Unlock()
	lines := j.browserBuf.flush()
	for _, line := range lines {
		j.browserLines <- line
	}
	j.closeBrowser.Do(func() {
		close(j.browserLines)
	})
}

func (j *job) setPythonErr(msg string) {
	j.mu.Lock()
	if j.pyErr == "" {
		j.pyErr = msg
	}
	j.mu.Unlock()
}

func (j *job) setPythonCancel(cancel context.CancelFunc) {
	j.mu.Lock()
	j.cancelPy = cancel
	j.mu.Unlock()
}

func (j *job) cancelPython() {
	j.mu.Lock()
	cancel := j.cancelPy
	j.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (j *job) fail(msg string) {
	j.mu.Lock()
	if j.status == "running" {
		j.status = "error"
		j.message = msg
	}
	j.mu.Unlock()
}

func (j *job) finishMatched() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.status != "running" {
		return
	}
	if j.pyErr != "" {
		j.status = "error"
		j.message = "python ivy_check failed: " + j.pyErr
		return
	}
	if j.browserCode != 0 {
		j.status = "error"
		j.message = fmt.Sprintf("browser goivy_check returned %d", j.browserCode)
		return
	}
	j.status = "matched"
	j.message = "browser Go Ivy and Python ivy_check XTRACE streams matched"
}

func (j *job) mismatch(reason, browser, python string) {
	j.mu.Lock()
	if j.status == "running" {
		j.status = "mismatch"
		j.message = fmt.Sprintf("%s after %d matching XTRACE lines\nbrowser: %.500s\npython : %.500s", reason, j.xtraceCount, strings.TrimSpace(browser), strings.TrimSpace(python))
	}
	j.mu.Unlock()
	j.cancelPython()
	_ = j.client.sendJSON(wsEnvelope{Type: "cancel", ID: j.id, Message: reason})
}

func (j *job) remember(side, line string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	entry := fmt.Sprintf("%s: %s", side, strings.TrimRight(line, "\n"))
	j.tail = append(j.tail, entry)
	if len(j.tail) > 200 {
		copy(j.tail, j.tail[len(j.tail)-200:])
		j.tail = j.tail[:200]
	}
}

func (j *job) snapshot() jobSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()
	return jobSnapshot{
		ID:          j.id,
		Filename:    j.filename,
		Params:      cloneStringMap(j.params),
		Status:      j.status,
		Message:     j.message,
		XTraceCount: j.xtraceCount,
		BrowserCode: j.browserCode,
		PythonError: j.pyErr,
		Tail:        append([]string(nil), j.tail...),
	}
}

func (b *streamLineBuffer) add(fd int, data string) []lineEvent {
	var lines []lineEvent
	if b.partial == "" {
		b.fd = fd
	}
	s := b.partial + data
	for {
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			b.partial = s
			if s != "" {
				b.fd = fd
			}
			return lines
		}
		lines = append(lines, lineEvent{FD: fd, Line: s[:idx+1]})
		s = s[idx+1:]
	}
}

func (b *streamLineBuffer) flush() []lineEvent {
	if b.partial == "" {
		return nil
	}
	line := b.partial
	fd := b.fd
	if fd == 0 {
		fd = 1
	}
	b.partial = ""
	b.fd = 0
	return []lineEvent{{FD: fd, Line: line}}
}

func (c *wsClient) readPump() {
	defer func() {
		c.app.unregister(c)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(readLimit)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			return
		}
		select {
		case <-c.done:
			return
		default:
			c.app.handleWSMessage(c, message)
		}
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message := <-c.send:
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *wsClient) write(mt int, payload []byte) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(mt, payload)
}

func (c *wsClient) sendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
		return nil
	case <-c.done:
		return errors.New("websocket client closed")
	case <-time.After(5 * time.Second):
		return errors.New("websocket send queue blocked")
	}
}

func (c *wsClient) close() {
	c.closed.Do(func() {
		close(c.done)
	})
}

func paramsAsArgs(params map[string]string) []string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	args := make([]string, 0, len(keys))
	for _, k := range keys {
		args = append(args, k+"="+params[k])
	}
	return args
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func newID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}

func defaultGoivyRoot() string {
	if wd, err := os.Getwd(); err == nil {
		if looksLikeGoivyRoot(wd) {
			return wd
		}
		if filepath.Base(wd) == "goldweb" && looksLikeGoivyRoot(filepath.Dir(wd)) {
			return filepath.Dir(wd)
		}
	}
	_, file, _, ok := runtime.Caller(0)
	if ok {
		dir := filepath.Dir(file)
		if filepath.Base(dir) == "goldweb" && looksLikeGoivyRoot(filepath.Dir(dir)) {
			return filepath.Dir(dir)
		}
	}
	return "."
}

func looksLikeGoivyRoot(dir string) bool {
	_, err1 := os.Stat(filepath.Join(dir, "go.mod"))
	_, err2 := os.Stat(filepath.Join(dir, "webvue", "static", "z3-471-api.js"))
	return err1 == nil && err2 == nil
}

const indexHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>goldweb</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 2rem; max-width: 70rem; }
    #status { font-weight: 700; }
    pre { border: 1px solid #ddd; padding: 1rem; min-height: 20rem; white-space: pre-wrap; overflow-wrap: anywhere; }
  </style>
</head>
<body>
  <h1>goldweb</h1>
  <p id="status">connecting</p>
  <pre id="log"></pre>

  <script type="text/plain" id="worker-source">
const textDecoder = new TextDecoder();
const textEncoder = new TextEncoder();

class WasiExit extends Error {
  constructor(code) {
    super('WASI exit ' + code);
    this.name = 'WasiExit';
    this.code = code >>> 0;
  }
}

function createWasiImports(getMemory, emit) {
  const errnoSuccess = 0;
  const errnoBadf = 8;

  function memoryBytes() {
    const memory = getMemory();
    if (!memory) {
      throw new Error('WASI import used before wasm memory was available');
    }
    return new Uint8Array(memory.buffer);
  }

  function memoryView() {
    const memory = getMemory();
    if (!memory) {
      throw new Error('WASI import used before wasm memory was available');
    }
    return new DataView(memory.buffer);
  }

  function writeU32(ptr, value) {
    memoryView().setUint32(ptr >>> 0, value >>> 0, true);
  }

  function writeU64(ptr, value) {
    memoryView().setBigUint64(ptr >>> 0, BigInt(value), true);
  }

  function zero(ptr, length) {
    memoryBytes().fill(0, ptr >>> 0, (ptr >>> 0) + (length >>> 0));
  }

  function copyRandom(ptr, length) {
    const heap = memoryBytes();
    let offset = ptr >>> 0;
    let remaining = length >>> 0;
    while (remaining > 0) {
      const chunkLength = Math.min(remaining, 65536);
      crypto.getRandomValues(heap.subarray(offset, offset + chunkLength));
      offset += chunkLength;
      remaining -= chunkLength;
    }
  }

  return {
    args_get() { return errnoSuccess; },
    args_sizes_get(argcPtr, argvBufSizePtr) {
      writeU32(argcPtr, 0);
      writeU32(argvBufSizePtr, 0);
      return errnoSuccess;
    },
    clock_time_get(clockId, precision, timePtr) {
      void clockId;
      void precision;
      writeU64(timePtr, BigInt(Date.now()) * 1000000n);
      return errnoSuccess;
    },
    environ_get() { return errnoSuccess; },
    environ_sizes_get(countPtr, bufSizePtr) {
      writeU32(countPtr, 0);
      writeU32(bufSizePtr, 0);
      return errnoSuccess;
    },
    fd_close() { return errnoSuccess; },
    fd_fdstat_get(fd, statPtr) {
      void fd;
      zero(statPtr, 24);
      return errnoSuccess;
    },
    fd_fdstat_set_flags() { return errnoSuccess; },
    fd_prestat_dir_name() { return errnoBadf; },
    fd_prestat_get() { return errnoBadf; },
    fd_read(fd, iovsPtr, iovsLen, nreadPtr) {
      void fd;
      void iovsPtr;
      void iovsLen;
      writeU32(nreadPtr, 0);
      return errnoSuccess;
    },
    fd_seek(fd, offset, whence, newOffsetPtr) {
      void fd;
      void offset;
      void whence;
      writeU64(newOffsetPtr, 0n);
      return errnoSuccess;
    },
    fd_write(fd, iovsPtr, iovsLen, nwrittenPtr) {
      const view = memoryView();
      const heap = memoryBytes();
      let written = 0;
      let output = '';
      for (let i = 0; i < iovsLen; i += 1) {
        const iov = (iovsPtr >>> 0) + i * 8;
        const ptr = view.getUint32(iov, true);
        const len = view.getUint32(iov + 4, true);
        written += len;
        if (fd === 1 || fd === 2) {
          output += textDecoder.decode(heap.subarray(ptr, ptr + len));
        }
      }
      if (output.length) {
        emit(fd, output);
      }
      writeU32(nwrittenPtr, written);
      return errnoSuccess;
    },
    path_open() { return errnoBadf; },
    poll_oneoff(inPtr, outPtr, nsubscriptions, neventsPtr) {
      void inPtr;
      void outPtr;
      void nsubscriptions;
      writeU32(neventsPtr, 0);
      return errnoSuccess;
    },
    proc_exit(code) { throw new WasiExit(code); },
    random_get(bufPtr, bufLen) {
      copyRandom(bufPtr, bufLen);
      return errnoSuccess;
    },
    sched_yield() { return errnoSuccess; }
  };
}

function requireExport(exports, name) {
  const fn = exports[name];
  if (typeof fn !== 'function') {
    throw new Error('Go Ivy wasm is missing export ' + name);
  }
  return fn;
}

function writeBytes(exports, memory, bytes) {
  const alloc = requireExport(exports, 'goivy_check_alloc');
  const ptr = alloc(bytes.length) >>> 0;
  new Uint8Array(memory.buffer, ptr, bytes.length).set(bytes);
  return ptr;
}

self.onmessage = async (event) => {
  const command = event.data;
  let z3;
  try {
    const assetBaseURL = self.location.origin;
    const z3Imports = await import(assetBaseURL + '/src/workers/smtZ3Imports.js');
    importScripts(assetBaseURL + '/z3-471-api.js');
    if (typeof initZ3 !== 'function') {
      throw new Error('z3-471-api.js did not expose initZ3');
    }
    z3 = await initZ3({
      locateFile(file) {
        if (file === 'z3-api.wasm') {
          return assetBaseURL + '/z3-471-api.wasm';
        }
        return assetBaseURL + '/' + file;
      }
    });

    let wasmMemory;
    const imports = {
      smt_z3: z3Imports.createSmtZ3Imports({ z3, getGoMemory: () => wasmMemory }),
      wasi_snapshot_preview1: createWasiImports(() => wasmMemory, (fd, data) => {
        self.postMessage({ type: 'stream', fd, data });
      })
    };

    const response = await fetch(assetBaseURL + '/goivy-check.wasm', { cache: 'no-store' });
    if (!response.ok) {
      throw new Error('could not fetch /goivy-check.wasm: HTTP ' + response.status);
    }
    const bytes = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, imports);
    const instance = result.instance;
    wasmMemory = instance.exports.memory;
    if (!wasmMemory) {
      throw new Error('Go Ivy wasm does not export memory');
    }

    try {
      if (typeof instance.exports._start === 'function') {
        instance.exports._start();
      }
    } catch (error) {
      if (!(error instanceof WasiExit) || error.code !== 0) {
        throw error;
      }
    }

    const run = requireExport(instance.exports, 'goivy_check_run');
    const free = instance.exports.goivy_check_free;
    const specBytes = textEncoder.encode(command.spec || '');
    const metaBytes = textEncoder.encode(JSON.stringify({
      filename: command.filename || 'browser_input.ivy',
      params: command.params || {}
    }));
    const specPtr = writeBytes(instance.exports, wasmMemory, specBytes);
    const metaPtr = writeBytes(instance.exports, wasmMemory, metaBytes);
    let code = 0;
    try {
      code = run(specPtr, specBytes.length, metaPtr, metaBytes.length) | 0;
    } finally {
      if (typeof free === 'function') {
        free(specPtr, specBytes.length);
        free(metaPtr, metaBytes.length);
      }
    }
    self.postMessage({ type: 'done', code });
  } catch (error) {
    self.postMessage({
      type: 'error',
      error: error && error.stack ? error.stack : String(error)
    });
  }
};
  </script>

  <script>
const statusEl = document.getElementById('status');
const logEl = document.getElementById('log');
const workerSource = document.getElementById('worker-source').textContent;
const workers = new Map();

function log(message) {
  logEl.textContent += message + '\n';
  logEl.scrollTop = logEl.scrollHeight;
}

function connect() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(proto + '//' + location.host + '/ws');

  ws.onopen = () => {
    statusEl.textContent = 'connected; waiting for goldweb commands';
    ws.send(JSON.stringify({ type: 'ready' }));
  };

  ws.onclose = () => {
    statusEl.textContent = 'disconnected; retrying';
    for (const worker of workers.values()) {
      worker.terminate();
    }
    workers.clear();
    setTimeout(connect, 1000);
  };

  ws.onerror = () => {
    statusEl.textContent = 'websocket error';
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.type === 'hello') {
      log(msg.message || 'hello');
      return;
    }
    if (msg.type === 'cancel') {
      const worker = workers.get(msg.id);
      if (worker) {
        worker.terminate();
        workers.delete(msg.id);
        log('cancelled ' + msg.id + ': ' + (msg.message || ''));
      }
      return;
    }
    if (msg.type === 'command' && msg.command === 'goivy_check') {
      runGoivyCheck(ws, msg);
      return;
    }
    log('unknown message: ' + event.data);
  };
}

function runGoivyCheck(ws, command) {
  log('starting goivy_check job ' + command.id + ' filename=' + command.filename);
  const url = URL.createObjectURL(new Blob([workerSource], { type: 'text/javascript' }));
  const worker = new Worker(url);
  workers.set(command.id, worker);

  worker.onmessage = (event) => {
    const msg = Object.assign({ id: command.id }, event.data);
    ws.send(JSON.stringify(msg));
    if (msg.type === 'stream') {
      log((msg.fd === 2 ? 'stderr: ' : 'stdout: ') + msg.data.replace(/\n$/, ''));
    }
    if (msg.type === 'done' || msg.type === 'error') {
      workers.delete(command.id);
      URL.revokeObjectURL(url);
      worker.terminate();
      log('finished ' + command.id + ' type=' + msg.type);
    }
  };

  worker.onerror = (error) => {
    ws.send(JSON.stringify({
      type: 'error',
      id: command.id,
      error: error.message || String(error)
    }));
    workers.delete(command.id);
    URL.revokeObjectURL(url);
    worker.terminate();
  };

  worker.postMessage(command);
}

connect();
  </script>
</body>
</html>
`
