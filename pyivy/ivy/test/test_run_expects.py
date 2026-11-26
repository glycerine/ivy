import unittest
from unittest.mock import Mock, patch

import pexpect

from run_expects import IvyToRtl


class IvyToRtlTests(unittest.TestCase):
    def test_reports_timeouts_and_closes_children(self):
        for timed_out in ('translator', 'validation'):
            with self.subTest(timed_out=timed_out):
                translator = Mock(exitstatus=0, before='translation output')
                validator = Mock(exitstatus=0, before='validation output')
                child = translator if timed_out == 'translator' else validator
                child.expect.side_effect = pexpect.TIMEOUT(timed_out)
                test = IvyToRtl('.', {'type': 'to_rtl', 'name': 'example', 'validate': 'true'})

                with patch('run_expects.spawn', side_effect=[translator, validator]):
                    self.assertFalse(test.run_translate())

                translator.close.assert_called_once()
                if timed_out == 'validation':
                    validator.close.assert_called_once()
