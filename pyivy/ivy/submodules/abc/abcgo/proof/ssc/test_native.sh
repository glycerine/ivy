#!/bin/sh
# Real-source integration while unrelated GIA source files remain drafts.
set -eu
cd "$(dirname "$0")/../.."
ssc_source_root=$(pwd)
ssc_check_root=$(mktemp -d /tmp/abcgo-ssc-native.XXXXXX)
trap 'rm -rf "$ssc_check_root"' EXIT HUP INT TERM
python3 - "$ssc_source_root" "$ssc_check_root" <<'PY'
from pathlib import Path
import re,shutil,os,sys
src,dst=map(Path,sys.argv[1:])
shutil.copy(src/'go.mod',dst/'go.mod')
for n in ['misc','sat','bool','map','opt','base','bdd']:os.symlink(src/n,dst/n)
(dst/'aig').mkdir()
for n in ['aig','hop','miniaig','saig','ioa']:os.symlink(src/'aig'/n,dst/'aig'/n)
g=dst/'aig/gia';g.mkdir()
for n in ['mf','satcost']:os.symlink(src/'aig/gia'/n,g/n)
files=re.findall(r'aig/gia/[\w.]+\.go',(src/'aig/gia/test_ported_core.sh').read_text())
for f in files:
 if not f.endswith('_test.go'):shutil.copy(src/f,g/Path(f).name)
(dst/'proof').mkdir();os.symlink(src/'proof/cec',dst/'proof/cec');os.symlink(src/'proof/ssc',dst/'proof/ssc')
PY
SSC_SOURCE_ROOT="$ssc_source_root/../src" GOCACHE=${GOCACHE:-/tmp/abcgo-cache} go -C "$ssc_check_root" "${SSC_CHECK_MODE:-test}" "$@" ./proof/ssc
