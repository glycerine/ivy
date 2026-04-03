"""
Canonical s-expression functions for fragment checker data structures.

Produces output compatible with Go goivy fragment/canon.go methods.
These serialize the module-level globals in ivy_fragment.py, which
correspond to the Go checker struct and its nested types.

Usage: call install() once at startup (after logic_sexp.install()).
"""

from .canon import node_canon, slice_canon, string_canon, bool_canon, canon_blake3


# --- UFNode ---

def uf_node_canon(n):
    """Return '(ufNode id:N)' or 'nil'. Matches Go ufNodeSexp()."""
    if n is None:
        return 'nil'
    return '(ufNode id:{})'.format(n.id)


def uf_node_set_canon(s):
    """Return sorted-id list '[id1 id2 ...]'. Matches Go ufNodeSetSexp()."""
    if s is None:
        return 'nil'
    ids = sorted(n.id for n in s)
    return '[' + ' '.join(str(i) for i in ids) + ']'


# --- Sorted map serialization ---

def sorted_map_canon(m, key_fn, val_fn):
    """Deterministic '(hash k:v ...)' with keys sorted lexicographically.
    Matches Go's sorted-map helpers in fragment/canon.go."""
    if m is None:
        return 'nil'
    pairs = sorted(((key_fn(k), val_fn(v)) for k, v in m.items()),
                   key=lambda p: p[0])
    return '(hash ' + ' '.join('{}:{}'.format(k, v) for k, v in pairs) + ')'


# --- strat_map key canonicalization ---
# Must match Go's "v:", "a:", "s:" prefixes from varKey/appKey/sortEqKey.

def strat_key_canon(key):
    """Convert a Python strat_map key to the same string Go uses.

    Python strat_map keys are:
      - Variable objects         -> Go varKey:  "v:" + var.Sexp()
      - (Symbol, int) tuples     -> Go appKey:  "a:" + sym.Sexp() + ":" + idx
      - Symbol('=', ...) objects -> Go sortEqKey: "s:" + sym.Sexp()
    """
    from . import ivy_logic as il
    if isinstance(key, tuple):
        sym, idx = key
        return 'a:{}:{}'.format(sym.sexp(), idx)
    if il.is_variable(key):
        return 'v:{}'.format(key.sexp())
    # Sort equality key: Symbol('=', ...)
    return 's:{}'.format(key.sexp())


# --- stratEntry ---

def strat_entry_canon(key):
    """Serialize strat_map key metadata matching Go stratEntry.Sexp().

    Go has an explicit stratEntry struct; Python infers it from key type."""
    from . import ivy_logic as il
    if isinstance(key, tuple):
        sym, idx = key
        return '(stratEntry sym:{} idx:{} v:nil isSort:false)'.format(
            sym.sexp(), idx)
    if il.is_variable(key):
        return '(stratEntry sym:nil idx:0 v:{} isSort:false)'.format(
            key.sexp())
    # Sort equality key
    return '(stratEntry sym:{} idx:0 v:nil isSort:true)'.format(key.sexp())


# --- arc ---

def arc_canon(arc_tuple):
    """Serialize an arc tuple. Matches Go arc.Sexp().

    Python arcs are 4-tuples (from, to, fmla, lineno)
    or 5-tuples (from, to, fmla, lineno, argIdx)."""
    if len(arc_tuple) == 5:
        v, anode, fmla, lineno, idx = arc_tuple
        return '(arc from:{} to:{} fmla:{} lineno:{} argIdx:{} hasIdx:true)'.format(
            uf_node_canon(v), uf_node_canon(anode),
            node_canon(fmla), lineno, idx)
    else:
        v, anode, fmla, lineno = arc_tuple
        return '(arc from:{} to:{} fmla:{} lineno:{} argIdx:-1 hasIdx:false)'.format(
            uf_node_canon(v), uf_node_canon(anode),
            node_canon(fmla), lineno)


# --- varID ---

def var_id_canon(v):
    """Serialize a variable as varID. Matches Go varID.Sexp().

    Go uses varID{name, sort} as map key; Python uses Variable directly."""
    return '(varID name:"{}" sort:"{}")'.format(v.name, str(v.sort))


# --- macroDef ---

def macro_def_canon(defn_lf_pair):
    """Serialize a macro_map value. Matches Go macroDef.Sexp().

    Python stores (definition, labeled_formula) tuples."""
    defn, lf = defn_lf_pair
    return '(macroDef def:{} lf:{})'.format(
        node_canon(defn) if defn is not None else 'nil',
        node_canon(lf) if lf is not None else 'nil')


# --- mapFmlaRes ---

def map_fmla_res_canon(res_pair):
    """Serialize a macro_value_map value. Matches Go mapFmlaRes.Sexp().

    Python stores (ufnode_or_none, set_of_ufnodes) tuples."""
    node, uvs = res_pair
    return '(mapFmlaRes node:{} uvs:{})'.format(
        uf_node_canon(node), uf_node_set_canon(uvs))


# --- skolemEntry ---

def skolem_entry_canon(entry_pair):
    """Serialize a skolem_map value. Matches Go skolemEntry.Sexp().

    Python stores (formula, ast_node) tuples."""
    fmla, ast_node = entry_pair
    return '(skolemEntry fmla:{} ast:{})'.format(
        node_canon(fmla), node_canon(ast_node))


# --- fmlaPair ---

def fmla_pair_canon(fmla, source, lineno):
    """Serialize a formula/source pair. Matches Go fmlaPair.Sexp()."""
    source_str = 'nil'
    if source is not None:
        if hasattr(source, 'canon'):
            source_str = source.canon()
        else:
            source_str = str(source)
    return '(fmlaPair fmla:{} source:{} lineno:{})'.format(
        node_canon(fmla), source_str, lineno)


# --- checker (top-level) ---

def checker_canon():
    """Serialize the entire fragment checker state from module globals.
    Matches Go checker.Sexp().

    Reads module-level globals from ivy_fragment."""
    from . import ivy_fragment as frag

    parts = []
    parts.append('(checker')

    # sig — not directly available as a module global in Python fragment;
    # Go's checker has it but Python reads from im.module.sig at call time.
    # Emit nil for now.
    parts.append(' sig:nil')

    # interp — same situation
    parts.append(' interp:nil')

    # universallyQuantifiedVars - can infinite loop, so skip.
    #uqv = getattr(frag, 'universally_quantified_variables', None)
    # parts.append(' universallyQuantifiedVars:')
    # if uqv is None:
    #     parts.append('nil')
    # else:
    #     parts.append(sorted_map_canon(uqv,
    #         lambda k: str(var_id_canon(k)),
    #         lambda v: node_canon(v)))

    # universalVarLineno — Python doesn't separate this; lineno comes from
    # the labeled_formula value in universally_quantified_variables.
    # Emit the lineno extracted from the lf source.
    # parts.append(' universalVarLineno:')
    # if uqv is None:
    #     parts.append('nil')
    # else:
    #     lineno_pairs = {}
    #     for v, lf in uqv.items():
    #         ln = getattr(lf, 'lineno', 0)
    #         if hasattr(ln, 'line'):
    #             ln = ln.line
    #         lineno_pairs[v] = ln
    #     parts.append(sorted_map_canon(lineno_pairs,
    #         lambda k: str(var_id_canon(k)),
    #         lambda v: str(v)))

    # stratMap
    sm = getattr(frag, 'strat_map', None)
    parts.append(' stratMap:')
    if sm is None:
        parts.append('nil')
    else:
        parts.append(sorted_map_canon(sm,
            lambda k: strat_key_canon(k),
            lambda v: uf_node_canon(v)))

    # stratInfo — Go has a separate map; Python infers metadata from the key type.
    # We iterate strat_map keys and derive the stratEntry from each key.
    parts.append(' stratInfo:')
    if sm is None:
        parts.append('nil')
    else:
        si_pairs = sorted(
            ((strat_key_canon(k), strat_entry_canon(k)) for k in sm),
            key=lambda p: p[0])
        parts.append('(hash ' + ' '.join('{}:{}'.format(k, v) for k, v in si_pairs) + ')')

    # arcs
    arc_list = getattr(frag, 'arcs', None)
    parts.append(' arcs:[')
    if arc_list:
        parts.append(' '.join(arc_canon(a) for a in arc_list))
    parts.append(']')

    # macroMap - can infinitely loop, skip.
    #mm = getattr(frag, 'macro_map', None)
    #parts.append(' macroMap:')
    #if mm is None:
    #    parts.append('nil')
    #else:
    #    parts.append(sorted_map_canon(mm,
    #        lambda k: '"{}"'.format(k),
    #        lambda v: macro_def_canon(v)))

    # macroValueMap
    mvm = getattr(frag, 'macro_value_map', None)
    parts.append(' macroValueMap:')
    if mvm is None:
        parts.append('nil')
    else:
        parts.append(sorted_map_canon(mvm,
            lambda k: '"{}"'.format(k),
            lambda v: map_fmla_res_canon(v)))

    # macroVarMap
    mvarm = getattr(frag, 'macro_var_map', None)
    parts.append(' macroVarMap:')
    if mvarm is None:
        parts.append('nil')
    else:
        parts.append(sorted_map_canon(mvarm,
            lambda k: str(var_id_canon(k)),
            lambda v: uf_node_canon(v)))

    # macroDepMap
    mdm = getattr(frag, 'macro_dep_map', None)
    parts.append(' macroDepMap:')
    if mdm is None:
        parts.append('nil')
    else:
        parts.append(sorted_map_canon(mdm,
            lambda k: str(var_id_canon(k)),
            lambda v: uf_node_set_canon(v)))

    # skolemMap - can inf loop, skip.
    # skm = getattr(frag, 'skolem_map', None)
    # parts.append(' skolemMap:')
    # if skm is None:
    #     parts.append('nil')
    # else:
    #     parts.append(sorted_map_canon(skm,
    #         lambda k: str(var_id_canon(k)),
    #         lambda v: skolem_entry_canon(v)))

    # varUniq
    vu = getattr(frag, 'var_uniq', None)
    parts.append(' varUniq:')
    if vu is not None:
        parts.append('(variableUniqifier)')
    else:
        parts.append('nil')

    parts.append(')')
    return ''.join(parts)


def install():
    """Monkey-patch canon() and blake3() onto UFNode. Call once at startup."""
    from .ivy_union_find2 import UFNode

    def _ufnode_canon(self):
        return '(ufNode id:{})'.format(self.id)

    def _ufnode_blake3(self):
        return canon_blake3(self.canon())

    UFNode.canon = _ufnode_canon
    UFNode.blake3 = _ufnode_blake3
