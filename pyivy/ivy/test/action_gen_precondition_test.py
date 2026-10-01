import os
import subprocess
import sys
import tempfile
import textwrap
import unittest


class ActionGenPreconditionTest(unittest.TestCase):
    def test_ivy_to_cpp_target_test_isolated_export_requires_drive_generator(self):
        spec = textwrap.dedent(
            """\
            #lang ivy1.7
            type epoch
            interpret epoch -> bv[1]

            type ts = struct {
                version : epoch,
                cid     : epoch
            }

            isolate test_timestamp = {
                action send_ts(a:ts)

                relation seen(X:ts)

                implementation {
                    implement send_ts(a:ts) {
                        require ~seen(a);
                        seen(a) := true;
                        call report_ts_a(a)
                    }
                }
            }

            import action report_ts_a(a:ts)
            export test_timestamp.send_ts
            """
        )
        with tempfile.TemporaryDirectory() as td:
            path = os.path.join(td, "action_gen_precondition.ivy")
            with open(path, "w") as f:
                f.write(spec)
            code = "\n".join(
                [
                    "import sys",
                    "from ivy import ivy_to_cpp",
                    "sys.argv = ['ivy_to_cpp', 'target=test', 'isolate=test_timestamp', 'outdir={}', 'action_gen_precondition.ivy']".format(
                        td.replace("\\", "\\\\")
                    ),
                    "ivy_to_cpp.main_int(False)",
                ]
            )
            env = os.environ.copy()
            env["PYTHONHASHSEED"] = "0"
            env["XTRACE_OFF"] = "1"
            env["PYTHONPATH"] = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
            proc = subprocess.run(
                [sys.executable, "-c", code],
                cwd=td,
                env=env,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
            )
            self.assertEqual(
                proc.returncode,
                0,
                "ivy_to_cpp target=test failed\nstdout:\n{}\nstderr:\n{}".format(
                    proc.stdout, proc.stderr
                ),
            )
            cpp_path = os.path.join(td, "action_gen_precondition.cpp")
            with open(cpp_path) as f:
                cpp = f.read()
            marker = "ext__test_timestamp__send_ts_gen::ext__test_timestamp__send_ts_gen"
            start = cpp.find(marker)
            self.assertNotEqual(start, -1, cpp)
            end = cpp.find("bool ext__test_timestamp__send_ts_gen::generate", start)
            self.assertNotEqual(end, -1, cpp[start:])
            body = cpp[start:end]
            self.assertIn("test_timestamp.seen", body)
            self.assertNotIn('add("(assert true)")', body)
            gen_start = end
            gen_end = cpp.find("void ext__test_timestamp__send_ts_gen::execute", gen_start)
            self.assertNotEqual(gen_end, -1, cpp[gen_start:])
            gen_body = cpp[gen_start:gen_end]
            self.assertIn("assumption_unsatisfied", gen_body)
            self.assertIn("action generator precondition cannot be satisfied", gen_body)


if __name__ == "__main__":
    unittest.main()
