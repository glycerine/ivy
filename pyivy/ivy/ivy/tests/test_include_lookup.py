import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from ivy import ivy_compiler, ivy_utils


class IncludeLookupTest(unittest.TestCase):
    def test_nested_include_search_order(self):
        with tempfile.TemporaryDirectory() as tmp:
            work, library, target = (Path(tmp) / name for name in ('work', 'lib', 'target'))
            for directory in (work, library, target):
                directory.mkdir()

            main = target / 'main.ivy'
            main.write_text('#lang ivy1.6\ninclude bridge\n')
            (work / 'bridge.ivy').write_text('#lang ivy1.6\ninclude leaf\n')
            (target / 'leaf.ivy').write_text('#lang ivy1.6\ntype from_leaf\n')
            (library / 'bridge.ivy').write_text('#lang ivy1.6\ntype from_library\n')
            (target / 'bridge.ivy').write_text('#lang ivy1.6\ntype from_target\n')

            def type_names():
                with main.open() as source:
                    return [decl.args[0].name for decl in ivy_compiler.read_module(source).decls]

            old_cwd = os.getcwd()
            try:
                os.chdir(work)
                with patch.object(ivy_utils, 'get_std_include_dir', return_value=str(library)):
                    self.assertEqual(type_names(), ['from_leaf'])
                    (work / 'bridge.ivy').unlink()
                    self.assertEqual(type_names(), ['from_library'])
                    (library / 'bridge.ivy').unlink()
                    self.assertEqual(type_names(), ['from_target'])
                with patch.object(ivy_utils, 'get_std_include_dir', side_effect=FileNotFoundError):
                    self.assertEqual(type_names(), ['from_target'])
            finally:
                os.chdir(old_cwd)
