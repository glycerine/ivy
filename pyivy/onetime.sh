## these command line utilities are called from 
## python liveness-to-safety checks, and must be 
## pre-installed to the bin directories to be found.
## they are vendored in the ivy/submodules/ dir.

mkdir -p ivy/bin
mkdir -p ivy/ivy/bin
cd ivy/submodules/aiger
gcc -O2 -o aigtoaig aigtoaig.c aiger.c

## install to both bin directories, since
## ivy_mc.py:1721 constructs the path:
##   aigtoaig_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'bin', 'aigtoaig')
## --> ~/ivy/pyivy/ivy/ivy/bin/aigtoaig
cp -p aigtoaig ../../bin/
cp -p aigtoaig ../../ivy/bin/
cp -p aigtoaig ${GOPATH}/bin/
cd ../../..

cd ivy/submodules/abc
make -j4 ## it takes a few minutes to compile binary 'abc'.
cp -p abc ../../bin/
cp -p abc ../../ivy/bin/
cp -p abc ${GOPATH}/bin/
cd ../../..
