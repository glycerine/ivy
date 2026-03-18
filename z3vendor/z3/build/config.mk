PREFIX=/usr/local
CC=gcc
CXX=g++
CXXFLAGS=-I/usr/local/opt/openjdk@17/include -I/usr/local/opt/zlib/include -D_MP_INTERNAL -DNDEBUG -D_EXTERNAL_RELEASE -D_AMD64_ -Wno-deprecated-declarations -std=c++11 -fvisibility=hidden -c -mfpmath=sse -msse -msse2 -D_NO_OMP_ -O3 -Wno-unknown-pragmas -Wno-overloaded-virtual -Wno-unused-value -fPIC
CFLAGS=-I/usr/local/opt/openjdk@17/include -I/usr/local/opt/zlib/include -D_MP_INTERNAL -DNDEBUG -D_EXTERNAL_RELEASE -D_AMD64_ -Wno-deprecated-declarations  -fvisibility=hidden -c -mfpmath=sse -msse -msse2 -D_NO_OMP_ -O3 -Wno-unknown-pragmas -Wno-overloaded-virtual -Wno-unused-value -fPIC
EXAMP_DEBUG_FLAG=
CXX_OUT_FLAG=-o 
C_OUT_FLAG=-o 
OBJ_EXT=.o
LIB_EXT=.a
AR=ar
AR_FLAGS=rcs
AR_OUTFLAG=
EXE_EXT=
LINK=g++
LINK_FLAGS=
LINK_OUT_FLAG=-o 
LINK_EXTRA_FLAGS=-lpthread -L/usr/local/lib -L/usr/local/opt/zlib/lib
SO_EXT=.dylib
SLINK=g++
SLINK_FLAGS=-dynamiclib
SLINK_EXTRA_FLAGS=
SLINK_OUT_FLAG=-o 
OS_DEFINES=
