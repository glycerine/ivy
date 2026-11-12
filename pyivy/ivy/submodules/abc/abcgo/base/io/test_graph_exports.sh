#!/bin/sh
# Focused real-source integration while unrelated ABC and I/O drafts remain.
set -eu
cd "$(dirname "$0")/../.."
graph_source_root=$(pwd)
graph_check_root=$(mktemp -d /tmp/abcgo-graph-exports.XXXXXX)
trap 'rm -rf "$graph_check_root"' EXIT HUP INT TERM
python3 - "$graph_source_root" "$graph_check_root" <<'PY'
from pathlib import Path
import os,shutil,sys
src,dst=map(Path,sys.argv[1:]);shutil.copy(src/'go.mod',dst/'go.mod')
for path in src.iterdir():
 if path.is_dir() and path.name!='base':os.symlink(path,dst/path.name)
(dst/'base').mkdir()
for path in (src/'base').iterdir():
 if path.is_dir() and path.name not in ['abc','io']:os.symlink(path,dst/'base'/path.name)
a=dst/'base/abc';a.mkdir()
files='network.go network_aig.go hierarchy.go abc.go network_api.go abcFanio.go abcObj.go abcRefs.go abcAig.go abcBarBuf.go abcBlifMv.go abcCheck.go mv_table.go mv_evaluate.go abcDfs.go dfs_iso.go abcFanOrder.go abcFunc.go abcHie.go abcHieCec.go scl_adapter.go fxu_adapter.go abcSop.go abcNtk.go ntk_core.go'.split()
for f in files:shutil.copy(src/'base/abc'/f,a/f)
i=dst/'base/io';i.mkdir()
for f in ['ioRegistry.go','ioAbc.go','ioWriteDot.go','ioWriteGml.go','ioWriteList.go','ioWriteBook.go','graph_export_test.go','graph_export_reference_test.go']:
 shutil.copy(src/'base/io'/f,i/f)
PY
GRAPH_SOURCE_ROOT="$graph_source_root/../src" GOCACHE=${GOCACHE:-/tmp/abcgo-cache} go -C "$graph_check_root" "${GRAPH_CHECK_MODE:-test}" "$@" ./base/io
