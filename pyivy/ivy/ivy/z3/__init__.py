import os
from pathlib import Path
import sys

_use_external_z3 = os.environ.get('IVY_EXTERNAL_Z3') == '1'
if sys.platform == 'darwin':
    _vendored_library = Path(__file__).with_name('libz3.dylib')
    _use_external_z3 = _use_external_z3 or not _vendored_library.is_file()

if _use_external_z3:
    from z3 import *
    from z3 import z3, z3core, z3consts, z3num, z3poly, z3printer, z3rcf, z3types, z3util
else:
    from .z3 import *
    from . import z3core, z3consts, z3num, z3poly, z3printer, z3rcf, z3types, z3util
