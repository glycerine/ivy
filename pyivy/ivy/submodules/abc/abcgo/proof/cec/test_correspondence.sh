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
import ast,re,shutil,os,sys
src,dst=map(Path,sys.argv[1:])
shutil.copy(src/'go.mod',dst/'go.mod')
for n in ['misc','sat','bool','map','opt','base','bdd']:
 if (src/n).exists():os.symlink(src/n,dst/n)
(dst/'aig').mkdir()
for n in ['aig','hop','miniaig','saig','ioa']:os.symlink(src/'aig'/n,dst/'aig'/n)
g=dst/'aig/gia';g.mkdir()
for n in ['mf','satcost']:os.symlink(src/'aig/gia'/n,g/n)
native=(src/'proof/cec/test_native.sh').read_text()
files=re.findall(r'aig/gia/[\w.]+\.go',(src/'aig/gia/test_ported_core.sh').read_text())
files += ['aig/gia/identity.go']
files+=re.findall(r"'(aig/gia/[\w.]+\.go)'",native)
for f in files:
 if not f.endswith('_test.go'):shutil.copy(src/f,g/Path(f).name)
p=dst/'proof/cec';p.mkdir(parents=True)
os.symlink(src/'proof/cec/core',p/'core')
os.symlink(src/'proof/dch',dst/'proof/dch')
# Reuse the parent owner's native list so new SAT/verify APIs and tests do
# not silently diverge between this subgroup gate and its integration gate.
names=ast.literal_eval(re.search(r"for n in (\[[^\]]+\]):\n shutil.copy\(src/'proof/cec'",native).group(1))
for n in dict.fromkeys(names+['cecPat.go','cecClass.go','cecSim.go','cecSimBack.go','simulation_words.go','simulation_words_test.go','simulation_patterns_test.go','simulation_classes_test.go','simulation_c_oracle_test.go','simulation_review_test.go','cecSeq.go','cecCorr.go','cecChoice.go','cecChoice_spec_test.go','cecCore.go','cecCore_simulation_test.go','seq_divide.go','seq_divide_test.go','original_ids.go','original_ids_test.go','spec_refine.go','spec_refine_test.go','sat_sequential.go','correspondence_test.go']):
 shutil.copy(src/'proof/cec'/n,p/n)
PY
ABC_SOURCE_ROOT="$source_root/../src" ABC_SOURCE_BINARY="$source_root/../abc" GOCACHE=${GOCACHE:-/tmp/abcgo-cache} go -C "$check_root" test "$@" ./proof/cec
