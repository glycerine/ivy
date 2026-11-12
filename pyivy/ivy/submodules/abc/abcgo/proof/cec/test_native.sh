#!/bin/sh
# Focused integration only. Copies native CEC/GIA sources into a temporary
# module because both parent packages still contain untranslated C drafts.
set -eu
cd "$(dirname "$0")/../.."
source_root=$(pwd)
check_root=$(mktemp -d /tmp/abcgo-cec-native.XXXXXX)
trap 'rm -rf "$check_root"' EXIT HUP INT TERM
python3 - "$source_root" "$check_root" <<'PY'
from pathlib import Path
import re,shutil,os,sys
src,dst=map(Path,sys.argv[1:])
shutil.copy(src/'go.mod',dst/'go.mod')
for n in ['misc','sat','bool','map','opt','base','bdd']:
 if (src/n).exists():os.symlink(src/n,dst/n)
(dst/'aig').mkdir()
for n in ['aig','hop','miniaig','saig','ioa']:os.symlink(src/'aig'/n,dst/'aig'/n)
g=dst/'aig/gia';g.mkdir()
for n in ['mf','satcost']:os.symlink(src/'aig/gia'/n,g/n)
files=re.findall(r'aig/gia/[\w.]+\.go',(src/'aig/gia/test_ported_core.sh').read_text())
files += ['aig/gia/aig_conversion.go','aig/gia/aig_representatives.go']
for f in files:
 if not f.endswith('_test.go'):shutil.copy(src/f,g/Path(f).name)
(dst/'proof').mkdir(exist_ok=True)
os.symlink(src/'proof/dch',dst/'proof/dch')
p=dst/'proof/cec';p.mkdir(parents=True)
os.symlink(src/'proof/cec/core',p/'core')
for n in ['cecInt.go','cec.h.go','params.go','aig_aliases.go','verify_aig.go','cecMan.go','sat_model.go','sat_queries.go','sat_miter.go','choice.go','choice_test.go','sweep.go','fraig.go','sweep_test.go','sweep_reference_test.go','verify.go','native_test.go']:
 shutil.copy(src/'proof/cec'/n,p/n)
PY
ABC_SOURCE_BINARY="$source_root/../abc" GOCACHE=${GOCACHE:-/tmp/abcgo-cache} go -C "$check_root" test "$@" ./proof/cec
