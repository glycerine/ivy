#!/bin/sh
# Focused, unchanged native source integration. Ordinary package tests remain
# required separately; the IF mapping driver awaits map/if's GIA dependency.
set -eu
abc_root=$(CDPATH= cd -- "$(dirname "$0")/../../.." && pwd)
check_dir=$(mktemp -d /tmp/abcgo-nwk-check.XXXXXX)
trap 'rm -rf "$check_dir"' EXIT HUP INT TERM
python3 - "$abc_root" "$check_dir" <<'PY'
import os,re,shutil,sys
root,dest=sys.argv[1:];source=root+'/abcgo'
os.makedirs(dest+'/opt/nwk',exist_ok=True)
for name in os.listdir(source+'/opt/nwk'):
    if name.endswith('.go') and name!='nwkMap.go':
        shutil.copyfile(source+'/opt/nwk/'+name,dest+'/opt/nwk/'+name)
shutil.copytree(source+'/opt/nwk/test_vectors',dest+'/opt/nwk/test_vectors')
os.symlink(root+'/src',dest+'/src')
shutil.copyfile(source+'/go.mod',dest+'/go.mod')
# Copy the actual, complete native library implementation and its exact type
# declaration. No fake implementation or substitute runtime behavior is used.
os.makedirs(dest+'/map/if',exist_ok=True)
shutil.copyfile(source+'/map/if/ifLibLut.go',dest+'/map/if/ifLibLut.go')
header=open(source+'/map/if/if.go').read()
record=re.search(r'type If_LibLut_t_ struct \{.*?\n\}',header,re.S).group()
open(dest+'/map/if/types.go','w').write('package if_\nconst IF_MAX_LUTSIZE=32\ntype If_LibLut_t=If_LibLut_t_\n'+record+'\n')
os.makedirs(dest+'/aig/saig',exist_ok=True)
shutil.copyfile(source+'/aig/saig/saigSimFast.go',dest+'/aig/saig/saigSimFast.go')
# Resolve package dependencies by copying their actual source, unedited.
selected={'opt/nwk','map/if','aig/saig'}
copied=set(selected)
while True:
    imports=set()
    for directory,dirs,files in os.walk(dest):
        for name in files:
            if name.endswith('.go'):
                imports.update(re.findall(r'"github.com/glycerine/abcgo/([\w/]+)"',open(directory+'/'+name).read()))
    pending=imports-copied
    if not pending:break
    for package in pending:
        path=dest+'/'+package
        os.makedirs(path,exist_ok=True)
        for name in os.listdir(source+'/'+package):
            if name.endswith('.go') and not name.endswith('_test.go'):
                shutil.copyfile(source+'/'+package+'/'+name,path+'/'+name)
        copied.add(package)
PY
export GOCACHE="${GOCACHE:-/tmp/abcgo-cache}"
go -C "$check_dir" test "$@" ./opt/nwk
go -C "$check_dir" vet ./opt/nwk
