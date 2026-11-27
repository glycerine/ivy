import os
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest.mock import patch

from ivy import ivy_rtlil_eqv

yosys_command = shutil.which(os.environ.get('IVY_YOSYS', 'yosys'))
package_bin = Path(ivy_rtlil_eqv.__file__).parent / 'bin'
abc_command = shutil.which(os.environ.get('IVY_ABC', str(package_bin / 'abc'))) or shutil.which('abc')
aigtoaig_command = shutil.which(os.environ.get('IVY_AIGTOAIG', str(package_bin / 'aigtoaig'))) or shutil.which('aigtoaig')


class RTLILEquivalenceTests(unittest.TestCase):
    def test_finds_tools_bundled_in_an_installed_package(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            package_bin = root / 'ivy' / 'bin'
            package_bin.mkdir(parents=True)
            yosys = root / 'yosys'
            abc = package_bin / 'abc'
            aigtoaig = package_bin / 'aigtoaig'
            for tool in (yosys, abc, aigtoaig):
                tool.touch(mode=0o700)
            with (
                patch.dict('os.environ', {'IVY_YOSYS': str(yosys)}, clear=True),
                patch.object(ivy_rtlil_eqv, '_repo_root', return_value=str(root)),
                patch.object(ivy_rtlil_eqv.shutil, 'which', return_value=None),
            ):
                self.assertEqual(ivy_rtlil_eqv.tools(), (str(yosys), str(abc), str(aigtoaig)))

    @unittest.skipUnless(yosys_command, 'requires Yosys')
    def test_normalizes_designs_from_paths_with_spaces(self):
        with tempfile.TemporaryDirectory(suffix=' with spaces') as directory:
            root = Path(directory)
            design = root / 'source design.il'
            design.write_text(r'''module \top
  wire input 1 \clk
  wire input 2 \rst
  wire input 3 \data
  wire output 4 \q
  connect \q \data
end
''')
            work = root / 'working directory'
            work.mkdir()
            config = ivy_rtlil_eqv.DesignCfg('gold')
            config.file, config.top, config.clk = str(design), 'top', 'clk'
            aig, mapping, edge = ivy_rtlil_eqv.run_yosys(yosys_command, config, str(work))
            self.assertTrue(Path(aig).is_file())
            self.assertTrue(Path(mapping).is_file())
            self.assertIsNone(edge)

    @unittest.skipUnless(abc_command and aigtoaig_command, 'requires ABC and aigtoaig')
    def test_abc_writes_counterexample_in_directory_with_spaces(self):
        with tempfile.TemporaryDirectory(suffix=' with spaces') as directory:
            aig = ivy_rtlil_eqv.AIG()
            bit = aig.newvar()
            inputs = [(('in', 'data', 0), bit)]
            cex = ivy_rtlil_eqv.run_abc(abc_command, aigtoaig_command, aig, inputs, bit, directory)
            self.assertEqual(cex, [1])
