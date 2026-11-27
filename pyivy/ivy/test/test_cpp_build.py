import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest


@unittest.skipUnless(shutil.which('g++') and sys.platform != 'win32', 'requires POSIX g++')
class CppBuildTests(unittest.TestCase):
    def test_support_headers_directory_may_contain_spaces(self):
        source = '''#lang ivy1.8
relation ready
after init { ready := false }
export action step = { ready := true }
'''
        compile_source = '''
from pathlib import Path
import shutil
from unittest.mock import patch
from ivy import ivy_to_cpp

headers = Path("support headers").resolve()
shutil.copytree(ivy_to_cpp.get_support_include_dir(), headers)
with patch.object(ivy_to_cpp, "get_support_include_dir", return_value=str(headers)):
    ivy_to_cpp.ivyc()
'''
        with tempfile.TemporaryDirectory() as directory:
            Path(directory, 'demo.ivy').write_text(source)
            env = os.environ.copy()
            env['XTRACE_OFF'] = '1'
            repo = Path(__file__).resolve().parent.parent
            env['PYTHONPATH'] = os.pathsep.join(filter(None, (str(repo), env.get('PYTHONPATH'))))
            result = subprocess.run(
                [sys.executable, '-O', '-c', compile_source, 'target=test', 'demo.ivy'],
                cwd=directory, env=env, capture_output=True, text=True, timeout=20,
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertTrue(Path(directory, 'demo').is_file())
