import os
import subprocess
import sys
import tempfile
import textwrap
import unittest


class VariantTargetTest(unittest.TestCase):
    def test_ivy_to_cpp_target_test_handles_three_variants(self):
        spec = textwrap.dedent(
            """\
            #lang ivy1.7
            type msg
            type color = {red, green}
            variant req of msg = struct {
                shade : color
            }
            variant ack of msg
            variant done of msg
            individual saved : msg
            action make_req returns(out:msg) = {
                var r : req;
                r.shade := green;
                saved := r;
                out := saved
            }
            export make_req
            """
        )
        with tempfile.TemporaryDirectory() as td:
            path = os.path.join(td, "variant_target_test.ivy")
            with open(path, "w") as f:
                f.write(spec)
            code = "\n".join(
                [
                    "import sys",
                    "from ivy import ivy_to_cpp",
                    "sys.argv = ['ivy_to_cpp', 'target=test', 'outdir={}', 'variant_target_test.ivy']".format(
                        td.replace("\\", "\\\\")
                    ),
                    "ivy_to_cpp.main_int(False)",
                ]
            )
            env = os.environ.copy()
            env["PYTHONHASHSEED"] = "0"
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
            cpp_path = os.path.join(td, "variant_target_test.cpp")
            with open(cpp_path) as f:
                cpp = f.read()
            self.assertIn("*>:msg:req", cpp)
            self.assertIn("*>:msg:ack", cpp)
            self.assertIn("*>:msg:done", cpp)


if __name__ == "__main__":
    unittest.main()
