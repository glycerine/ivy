import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


class ModelExportTests(unittest.TestCase):
    export_model = '''
import sys
from ivy import ivy_init, ivy_isolate, ivy_module, {exporter}
with ivy_module.Module():
    with open(sys.argv[1]) as source:
        ivy_init.source_file(sys.argv[1], source, create_isolate=False)
    ivy_isolate.create_isolate(None)
    {exporter}.check_isolate()
'''

    def run_ivy(self, source, script, filename='model.ivy'):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / filename).write_text(source)
            env = os.environ.copy()
            env['XTRACE_OFF'] = '1'
            repo = Path(__file__).resolve().parent.parent
            env['PYTHONPATH'] = os.pathsep.join(filter(None, (str(repo), env.get('PYTHONPATH'))))
            result = subprocess.run(
                [sys.executable, '-O', '-c', script, filename],
                cwd=root, env=env, capture_output=True, text=True, timeout=20,
            )
            models = {path.suffix: path.read_text() for path in root.glob('model.*')
                      if path.suffix != '.ivy'}
            duoai = root / 'model_duoai.ivy'
            if duoai.exists():
                models['duoai'] = duoai.read_text()
            return result, models

    def test_export_does_not_skip_checking_later_isolates(self):
        for method, extension in [('convert_to_mypyvy', '.pyv'), ('vmt', '.vmt')]:
            with self.subTest(method=method):
                source = '''#lang ivy1.8
isolate a_model = {{
    attribute method = {}
}}
isolate z_proof = {{
    invariant [must_fail] false
}}
'''.format(method)
                result, models = self.run_ivy(source, 'from ivy import ivy_check; ivy_check.main()')
                self.assertIn(extension, models, result.stdout + result.stderr)
                self.assertIn('failed checks:', result.stdout + result.stderr)
                self.assertIn('must_fail', result.stdout + result.stderr)
                self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_mypyvy_uses_declared_init_before_running_initializer(self):
        source = '''#lang ivy1.6
relation p
relation q
init p
after init { q := p; p := false }
'''
        result, models = self.run_ivy(source, self.export_model.format(exporter='ivy_mypyvy'))
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        init, = [line for line in models['.pyv'].splitlines() if line.startswith('init ')]
        formula = init.removeprefix('init ').split('. ')[-1]
        self.assertIn('!p', formula.split(' & '))
        self.assertIn('q', formula.split(' & '))

    def test_mypyvy_enum_axioms_close_the_domain(self):
        source = '''#lang ivy1.8
type color = {red, green}
individual chosen : color
after init { chosen := red }
export action step = { chosen := green }
invariant [known_color] chosen = red | chosen = green
'''
        result, models = self.run_ivy(source, self.export_model.format(exporter='ivy_mypyvy'))
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        model = models['.pyv']
        self.assertIn('immutable constant red: color', model)
        self.assertIn('immutable constant green: color', model)
        self.assertRegex(model, r'forall (\w+):color\. \1 = red \| \1 = green')

    def test_duoai_keeps_definitions_used_by_the_exported_model(self):
        source = '''#lang ivy1.8
relation driving
relation ready
definition ready = driving
after init { driving := true }
export action step = { driving := ~driving }
invariant [same] ready <-> driving
'''
        result, models = self.run_ivy(source, self.export_model.format(exporter='ivy_duoai'))
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertRegex(models['duoai'], r'(?m)^(?:definition|axiom).*\bready\b.*\bdriving\b')
