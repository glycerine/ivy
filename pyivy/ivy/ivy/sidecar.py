"""
Ivy Sidecar HTTP Server

A lightweight HTTP server that wraps the Ivy verification tool for
use as a backend by the Go ivyweb server. Exposes the same session-based
API that the Go handlers provide, producing canonical JSON responses
suitable for bit-for-bit conformance checking.

Usage:
    python3 -m ivy.sidecar --port 18080
"""

import argparse
import json
import sys
import threading
import traceback
from http.server import HTTPServer, BaseHTTPRequestHandler
from socketserver import ThreadingMixIn
from io import StringIO
from collections import OrderedDict
from queue import Queue, Empty

# ---------------------------------------------------------------------------
# Ivy imports — heavy; imported once at startup.
# ---------------------------------------------------------------------------

import ivy.ivy_logic as il
import ivy.ivy_module as im
import ivy.ivy_compiler as ic
import ivy.ivy_art as ia_art
import ivy.ivy_interp as itp
import ivy.ivy_actions as act
import ivy.ivy_trace as ivy_trace
import ivy.ivy_logic_utils as ilu
import ivy.ivy_solver as slv
import ivy.logic as lg
from ivy.cy_elements import CyElements


# ---------------------------------------------------------------------------
# Session state
# ---------------------------------------------------------------------------

class IvySession:
    """Holds per-session Ivy state."""

    def __init__(self, session_id):
        self.id = session_id
        self.ag = None           # AnalysisGraph
        self.concept_domain = None
        self.concept_session = None
        self.file_path = ""
        self.file_content = ""
        self.active_isolate = ""
        self.available_isolates = []
        self.toggles = {"edges": {}}
        self.events = Queue(maxsize=256)

    def emit(self, event):
        try:
            self.events.put_nowait(event)
        except Exception:
            pass  # drop if full


# Global session store — protected by lock since Python Ivy uses globals.
_sessions = {}
_session_counter = 0
_ivy_lock = threading.Lock()
NO_ISOLATES_FOUND_CHOICE = "no_isolates_found"


def _reset_ivy_globals():
    """Reset Ivy's module-level global state between compilations."""
    il.sig = il.Sig()
    im.module = im.Module()


# ---------------------------------------------------------------------------
# Session operations
# ---------------------------------------------------------------------------

def _new_session():
    global _session_counter
    _session_counter += 1
    sid = "s{}".format(_session_counter)
    sess = IvySession(sid)
    _sessions[sid] = sess
    return sid


def _webui_isolate_names():
    return sorted(
        name for name, iso in getattr(im.module, "isolates", {}).items()
        if not _webui_implicit_this_isolate(name, iso)
    )


def _webui_implicit_this_isolate(name, iso):
    args = getattr(iso, "args", ())
    if name != "this" or getattr(iso, "with_args", 0) != 0 or len(args) != 2:
        return False
    return all(getattr(arg, "relname", None) == "this" for arg in args)


def _load_content(sess, filename, content, isolate=""):
    """Compile Ivy source and populate session state."""
    with _ivy_lock:
        _reset_ivy_globals()
        sess.file_path = filename
        sess.file_content = content
        isolate = (isolate or "").strip()
        ic.isolate.set(isolate if isolate and isolate != NO_ISOLATES_FOUND_CHOICE else None)

        # ivy_load_file expects the full source including #lang header.
        sio = StringIO(content)
        # Note: ivy_compile has a variable shadowing bug where the loop
        # variable 'iso' (IsolateDef) overwrites the 'ivy_isolate as iso'
        # import. We work around this by calling ivy_load_file with
        # create_isolate=False, then calling ivy_isolate.create_isolate
        # directly.
        from ivy import ivy_isolate
        from ivy import ivy_utils as iu
        ic.ivy_load_file(sio, create_isolate=False)
        user_isolates = _webui_isolate_names()
        if not isolate and len(user_isolates) == 0:
            sess.active_isolate = NO_ISOLATES_FOUND_CHOICE
            sess.available_isolates = [NO_ISOLATES_FOUND_CHOICE]
        elif isolate == NO_ISOLATES_FOUND_CHOICE:
            if len(user_isolates) != 0:
                raise iu.IvyError(None, "{} is only valid when the source declares no isolates".format(NO_ISOLATES_FOUND_CHOICE))
            sess.active_isolate = NO_ISOLATES_FOUND_CHOICE
            sess.available_isolates = [NO_ISOLATES_FOUND_CHOICE]
        else:
            selected = isolate
            if not selected and len(user_isolates) > 0:
                selected = user_isolates[0]
            ivy_isolate.create_isolate(selected if selected else None, im.module)
            sess.available_isolates = _webui_isolate_names()
            sess.active_isolate = selected
        im.module.labeled_axioms.extend(im.module.labeled_props)
        im.module.theory_context().__enter__()
        sess.ag = ic.ivy_new()

        # Build concept domain from compiled signature.
        # Note: get_initial_concept_domain has a bug in ConceptDict.add
        # (references self.category instead of self[category]).
        # We build the concept data directly from the signature instead.
        sess.concept_domain = None  # built lazily in _get_concept


def _load_path(sess, path):
    """Load Ivy source from a file path."""
    with open(path, "r") as f:
        content = f.read()
    _load_content(sess, path, content)


def _get_concept(sess):
    """Return concept graph data matching Go's apiConcept JSON format."""
    relations = []
    edges = []
    node_labels = []
    nodes = []
    label_sorts = {}

    # Build concept data from the module's signature (im.module.sig),
    # not il.sig (they can be different objects).
    sig = im.module.sig if im.module is not None else il.sig
    sort_names = set()
    if sig is not None:
        for s in sig.sorts.values():
            if s.name != "bool":
                sort_names.add(s.name)
                nodes.append(s.name)

        for sym_name, sym in sig.symbols.items():
            if not hasattr(sym, 'sort') or sym.sort is None:
                continue
            sort = sym.sort
            if hasattr(sort, 'arity'):
                if sort.arity == 1 and hasattr(sort, 'range') and sort.range == lg.Boolean:
                    # Unary relation -> node label
                    node_labels.append(sym_name)
                    if hasattr(sort, 'domain') and len(sort.domain) > 0:
                        label_sorts[sym_name] = str(sort.domain[0])
                elif sort.arity == 2 and hasattr(sort, 'range') and sort.range == lg.Boolean:
                    # Binary relation -> edge
                    edges.append(sym_name)
                relations.append(sym_name)

    nodes.sort()
    edges.sort()
    node_labels.sort()
    relations.sort()

    # Render concept graph elements via CyElements.
    cy = CyElements()
    for name in nodes:
        cy.add_node(
            obj=name,
            label=name,
            classes=["node_unknown"],
            short_info=name,
            long_info=name,
            shape="ellipse",
        )

    # Serialize elements, stripping non-JSON-safe fields.
    elements = _clean_elements(cy.elements)

    return {
        "abstract_value": {},
        "edges": edges,
        "elements": elements,
        "label_sorts": label_sorts,
        "node_labels": node_labels,
        "nodes": nodes,
        "relations": relations,
    }


def _clean_elements(elements):
    """Strip callback/function fields that aren't JSON serializable."""
    cleaned = []
    for el in elements:
        cel = {"group": el["group"], "data": {}, "classes": el.get("classes", "")}
        if "locked" in el:
            cel["locked"] = el["locked"]
        for k, v in el["data"].items():
            if k in ("events", "actions"):
                continue  # callbacks, not serializable
            if callable(v):
                continue
            cel["data"][k] = v
        cleaned.append(cel)
    return cleaned


def _get_arg(sess):
    """Return ARG data matching Go's apiARG JSON format."""
    cy = CyElements()
    if sess.ag is not None:
        for s in sess.ag.states:
            classes = ["bottom_state"] if s.is_bottom() else ["state"]
            cy.add_node(
                obj=id(s),
                label=str(s.id),
                classes=classes,
                short_info=str(s.id),
                long_info=str(s.id),
                shape="ellipse",
            )
        for source, op, label, target in sess.ag.transitions:
            cls = "transition_join" if label == "join" else "transition_action"
            cy.add_edge(
                obj=id(op),
                source_obj=id(source),
                target_obj=id(target),
                label=str(label),
                classes=[cls],
                short_info=str(op),
                long_info=str(op),
            )
    elements = _clean_elements(cy.elements)
    return {"elements": elements}


def _run_check(sess, mode):
    """Run verification check matching Go's apiCheck JSON format."""
    result = {
        "failed_conjecture": "",
        "failed_label": "",
        "message": "",
        "mode": mode,
        "result": "error",
        "status": "ok",
        "used_relations": None,
    }

    if sess.ag is None or im.module is None:
        result["message"] = "No module loaded \u2014 load an .ivy file first"
        return result

    with _ivy_lock:
        if mode == "induction":
            conjs = im.module.conjs
            if len(conjs) == 0:
                result["result"] = "pass"
                result["message"] = "No conjectures to check"
                return result

            try:
                ag, succeed, fail = ivy_trace.make_check_art(precond=list(conjs))
            except Exception as e:
                result["message"] = "Error building check art: {}".format(e)
                return result

            # Check each conjecture.
            to_test = [None] + list(conjs)  # None = safety check
            used_names = frozenset(x.name for x in list(il.sig.symbols.values()))

            def witness(v):
                c = lg.Const("@" + v.name, v.sort)
                return c

            for conj in to_test:
                try:
                    if conj is None:
                        clauses = ilu.true_clauses()
                        post = fail
                    else:
                        clauses = ilu.dual_clauses(conj, witness)
                        post = succeed

                    clauses.annot = act.EmptyAnnotation()
                    res = ivy_trace.check_final_cond(ag, post, clauses, [], True)

                    if res is not None:
                        if conj is None:
                            result["result"] = "fail"
                            result["message"] = "An assertion failed."
                        else:
                            formula = str(il.drop_universals(conj.to_formula()))
                            result["result"] = "fail"
                            result["message"] = "The following conjecture is not relatively inductive:"
                            result["failed_conjecture"] = formula
                            # Collect used relations.
                            used = []
                            for sym_name, sym in il.sig.symbols.items():
                                if (hasattr(sym.sort, 'range') and
                                        sym.sort.range == lg.Boolean):
                                    used.append(sym_name)
                            used.sort()
                            result["used_relations"] = used
                        return result
                except Exception as e:
                    traceback.print_exc()
                    formula = ""
                    if conj is not None:
                        formula = str(il.drop_universals(conj.to_formula()))
                    result["result"] = "fail"
                    result["message"] = "Could not check conjecture (solver error): {}".format(e)
                    result["failed_conjecture"] = formula
                    return result

            # All conjectures passed.
            lines = [str(c) for c in conjs]
            result["result"] = "pass"
            result["message"] = "Inductive invariant found:\n" + "\n".join(lines)
            return result

        else:
            result["result"] = "error"
            result["message"] = "Unknown mode: " + mode
            return result


def _execute_action(sess, action_name, args):
    """Execute a named action."""
    if not action_name:
        raise ValueError("empty action name")
    return {"status": "ok"}


def _concept_op(sess, op, params):
    """Execute a concept domain operation."""
    # Stub — concept ops will be wired incrementally.
    return {"status": "ok"}


def _get_toggles(sess):
    return sess.toggles


def _set_toggle(sess, edge, display_class, value):
    if edge not in sess.toggles["edges"]:
        sess.toggles["edges"][edge] = {}
    sess.toggles["edges"][edge][display_class] = value


def _save_state(sess):
    return {
        "file_content": sess.file_content,
        "file_path": sess.file_path,
        "session_id": sess.id,
        "toggles": sess.toggles,
    }


# ---------------------------------------------------------------------------
# Canonical JSON serialization
# ---------------------------------------------------------------------------

def canonical_json(obj):
    """Produce compact JSON with sorted keys — must match Go's json.Marshal."""
    return json.dumps(obj, sort_keys=True, separators=(",", ":"))


# ---------------------------------------------------------------------------
# HTTP Handler
# ---------------------------------------------------------------------------

class SidecarHandler(BaseHTTPRequestHandler):
    """Handles HTTP requests for the Ivy sidecar."""

    def log_message(self, format, *args):
        # Suppress default logging to stderr.
        pass

    def _send_json(self, data, status=200):
        body = canonical_json(data).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _send_error(self, status, msg):
        self._send_json({"error": msg}, status)

    def _read_json(self):
        length = int(self.headers.get("Content-Length", 0))
        if length == 0:
            return {}
        body = self.rfile.read(length)
        return json.loads(body)

    def _get_session(self, sid):
        sess = _sessions.get(sid)
        if sess is None:
            self._send_error(404, "session not found")
        return sess

    def do_GET(self):
        path = self.path.rstrip("/")

        if path == "/health":
            self._send_json({"status": "ok"})
            return

        parts = path.lstrip("/").split("/")
        if len(parts) < 3 or parts[0] != "session":
            self._send_error(404, "unknown endpoint")
            return

        sid = parts[1]
        rest = "/".join(parts[2:])
        sess = self._get_session(sid)
        if sess is None:
            return

        try:
            if rest == "arg":
                self._send_json(_get_arg(sess))
            elif rest == "concept":
                self._send_json(_get_concept(sess))
            elif rest == "toggles":
                self._send_json(_get_toggles(sess))
            elif rest == "proof":
                self._send_json({"elements": []})
            elif rest == "save":
                self._send_json(_save_state(sess))
            elif rest == "events":
                self._send_sse(sess)
            else:
                self._send_error(404, "unknown endpoint")
        except Exception as e:
            traceback.print_exc()
            self._send_error(400, str(e))

    def do_POST(self):
        path = self.path.rstrip("/")
        parts = path.lstrip("/").split("/")

        if len(parts) == 2 and parts[0] == "session" and parts[1] == "new":
            sid = _new_session()
            self._send_json({"session_id": sid})
            return

        if len(parts) < 3 or parts[0] != "session":
            self._send_error(404, "unknown endpoint")
            return

        sid = parts[1]
        rest = "/".join(parts[2:])
        sess = self._get_session(sid)
        if sess is None:
            return

        try:
            body = self._read_json()

            if rest == "load":
                if "content" in body:
                    _load_content(sess, body.get("filename", ""), body["content"], body.get("isolate", ""))
                    self._send_json({
                        "filename": body.get("filename", ""),
                        "isolate": sess.active_isolate,
                        "isolates": list(sess.available_isolates),
                        "status": "ok",
                    })
                elif "path" in body:
                    p = body["path"]
                    if not p:
                        self._send_error(400, "empty file path")
                        return
                    _load_path(sess, p)
                    self._send_json({"status": "ok"})
                else:
                    self._send_error(400, "missing content or path")

            elif rest == "action":
                result = _execute_action(
                    sess, body.get("action", ""), body.get("args", {}))
                self._send_json(result)

            elif rest == "check":
                mode = body.get("mode", "induction")
                result = _run_check(sess, mode)
                self._send_json(result)

            elif rest.startswith("concept/"):
                op = rest[len("concept/"):]
                result = _concept_op(sess, op, body)
                self._send_json(result)

            elif rest == "toggles":
                _set_toggle(
                    sess,
                    body.get("edge", ""),
                    body.get("display_class", ""),
                    body.get("value", False),
                )
                self._send_json({"status": "ok"})

            elif rest == "arg/action":
                result = {
                    "action": body.get("action", ""),
                    "node": body.get("node", ""),
                    "status": "ok",
                }
                self._send_json(result)

            elif rest == "proof/action":
                result = {
                    "action": body.get("action", ""),
                    "goal": body.get("goal", ""),
                    "status": "ok",
                }
                self._send_json(result)

            else:
                self._send_error(404, "unknown endpoint")

        except Exception as e:
            traceback.print_exc()
            self._send_error(400, str(e))

    def _send_sse(self, sess):
        """Stream Server-Sent Events."""
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.send_header("Connection", "keep-alive")
        self.end_headers()
        self.wfile.flush()

        while True:
            try:
                evt = sess.events.get(timeout=30)
                data = canonical_json(evt)
                self.wfile.write("data: {}\n\n".format(data).encode("utf-8"))
                self.wfile.flush()
            except Empty:
                # Send keepalive comment.
                try:
                    self.wfile.write(b": keepalive\n\n")
                    self.wfile.flush()
                except BrokenPipeError:
                    return
            except BrokenPipeError:
                return


# ---------------------------------------------------------------------------
# Threaded server
# ---------------------------------------------------------------------------

class ThreadedHTTPServer(ThreadingMixIn, HTTPServer):
    daemon_threads = True
    allow_reuse_address = True


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description="Ivy Sidecar HTTP Server")
    parser.add_argument("--port", type=int, default=18080, help="port to listen on")
    args = parser.parse_args()

    server = ThreadedHTTPServer(("127.0.0.1", args.port), SidecarHandler)
    print("ivy sidecar: listening on 127.0.0.1:{}".format(args.port), file=sys.stderr)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    server.server_close()


if __name__ == "__main__":
    main()
