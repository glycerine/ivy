#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#

import sys

# pick up some paths if were are run as main
if __name__ == "__main__":
    mypath = sys.path[0]
    sys.path.insert(0,mypath+'/..')
    sys.path.insert(0,mypath+'/../tkui')
#    print "path: {}".format(sys.path)

import pickle
import json
import string
from .ivy_compiler import IvyError, ivy_new, ivy_load_file
from .ivy_utils import Parameter, set_parameters
from . import ivy_logic
from . import proof as pf
from . import ivy_utils as iu
from . import ivy_module
from . import xtracer
#import tactics_api as ta

# mode = Parameter("mode",None)

def usage():
    print("usage: \n  {} <file>.[a2g,ivy,dfy]\n {} <file>.a2g <file.[ivy,dfy]> ".format(sys.argv[0],sys.argv[0]))
    sys.exit(1)

def open_read(fn):
    try:
        f = open(fn,'rb' if fn.endswith('.a2g') else 'r')
        return f
    except:
        print("not found: %s" % fn)
        sys.exit(1)

def get_verb():
    if len(sys.argv) > 1 and '.' not in sys.argv[1] and '=' not in sys.argv[1]:
        return sys.argv.pop(1)
    return 'interact'

def read_params():
    ps = dict()
    args = sys.argv[1:]
    while args and '=' in args[0]:
        thing = str.split(args[0],'=')
        if len(thing) > 2:
            usage()
        ps[thing[0]] = thing[1]
        args = args[1:]
    try:
        set_parameters(ps)
    except IvyError as e:
        print(e)
        exit(1)
    sys.argv = sys.argv[0:1] + args

# def ivy_init2():
#     if len(sys.argv) < 2 or len(sys.argv) > 2:
#         usage()
#     fn = sys.argv[1]
#     try:
#         session = pf.AnalysisSession(fn)
#         ta.set_context(session)
# # tt.CheckCover.apply.im_func.func_code.co_argcount
#         return session.analysis_state.ivy_ag
#     except IvyError as e:
#         e.filename = fn
#         print repr(e)
#         sys.exit(1)
    
    
def source_file(fn,f,**kwargs):
    if __debug__: xtracer.trace("init.SourceFile ENTER file=%s" % xtracer.normalize_filename(fn))
    try:
        with iu.SourceFile(fn):
            ivy_load_file(f,**kwargs)
            ivy_module.module.name = fn[:fn.rindex('.')]
    except IvyError as e:
        if not hasattr(e,'filename'):
            e.filename = fn
        print(str(e))
        sys.exit(1)
    if __debug__: xtracer.trace("init.SourceFile EXIT file=%s" % xtracer.normalize_filename(fn))

def _is_portable_trace(stream):
    while chunk := stream.peek(1):
        remaining = chunk.lstrip(b' \t\r\n')
        stream.read(len(chunk) - len(remaining))
        if remaining:
            return remaining.startswith(b'{')
    return False


def ivy_init(**kwargs):
    verb = kwargs.pop('verb',None)
    read_params()

#    if mode.get() == "ivy2":
#        return ivy_init2()

    if len(sys.argv) < 2 or len(sys.argv) > 3:
        usage()
    if verb == 'replay' and (len(sys.argv) != 2 or not sys.argv[1].endswith('.a2g')):
        usage()

    from . import ivy_trace
    files = [(fn,open_read(fn)) for fn in sys.argv[1:]]

    if files[0][0].endswith('.a2g'):
        fn,f = files.pop(0)
        with f:
            if verb == 'replay' and _is_portable_trace(f):
                try:
                    archive = json.load(f)
                except (ValueError, UnicodeError) as error:
                    raise IvyError(None,'invalid trace archive {}: {}'.format(fn,error)) from error
                if not isinstance(archive,dict) or archive.get('format') != 'ivy-trace-1':
                    raise IvyError(None,'unsupported trace format in {}'.format(fn))
                return archive
            ag = pickle.load(f)
        ivy_module.module = ag.domain
        ivy_logic.sig = ag.domain.sig
        if hasattr(ag,'ivy_version'):
            iu.set_string_version(ag.ivy_version)
    else:
        ag = None

    if files:
        fn,f = files.pop(0)
        if not fn.endswith('.ivy') and not fn.endswith('.dfy'):
            usage()
        skwargs = dict(kwargs)
        skwargs.update(iu.get_default_ui_compile_kwargs())
        source_file(fn,f,**skwargs)

        if ag:
            ag.update_module()
        else: 
            ag = ivy_new()

    return ag
