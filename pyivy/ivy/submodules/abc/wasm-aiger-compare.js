"use strict";

const childProcess = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");

function argValue(name, defaultValue) {
  const index = process.argv.indexOf(name);
  if (index === -1) return defaultValue;
  if (index + 1 >= process.argv.length) {
    throw new Error(`missing value for ${name}`);
  }
  return process.argv[index + 1];
}

const modulePath = path.resolve(argValue("--module", "./build/wasm/abc.js"));
const nativeAbcPath = path.resolve(argValue("--abc", "./abc"));
const aigtoaigPath = path.resolve(argValue("--aigtoaig", "../../ivy/bin/aigtoaig"));
const caseCount = Number(argValue("--cases", "25"));
const seed = Number(argValue("--seed", "1"));

const initAbc = require(modulePath);
const abcRc = fs.readFileSync(path.resolve("abc.rc"), "utf8");

function makeRng(initialSeed) {
  let state = initialSeed >>> 0;
  if (state === 0) state = 0x6d2b79f5;
  return function rng() {
    state ^= state << 13;
    state ^= state >>> 17;
    state ^= state << 5;
    return (state >>> 0) / 0x100000000;
  };
}

function randInt(rng, limit) {
  return Math.floor(rng() * limit);
}

function chooseLiteral(rng, maxVar) {
  if (rng() < 0.08) return randInt(rng, 2);
  return 2 * (1 + randInt(rng, maxVar)) + randInt(rng, 2);
}

function generateAag(rng, index) {
  if (index === 0) return "aag 2 1 1 1 0\n2\n4 0\n4\n";
  if (index === 1) return "aag 2 1 1 1 0\n2\n4 2\n5\n";

  const inputs = 1 + randInt(rng, 4);
  const latches = 1 + randInt(rng, 3);
  const ands = randInt(rng, 8);
  const maxVar = inputs + latches + ands;
  const lines = [`aag ${maxVar} ${inputs} ${latches} 1 ${ands}`];

  for (let i = 1; i <= inputs; i++) {
    lines.push(String(2 * i));
  }

  for (let i = 0; i < latches; i++) {
    const latchVar = inputs + i + 1;
    lines.push(`${2 * latchVar} ${chooseLiteral(rng, maxVar)}`);
  }

  lines.push(String(chooseLiteral(rng, maxVar)));

  for (let i = 0; i < ands; i++) {
    const andVar = inputs + latches + i + 1;
    lines.push(`${2 * andVar} ${chooseLiteral(rng, andVar - 1)} ${chooseLiteral(rng, andVar - 1)}`);
  }

  return `${lines.join("\n")}\n`;
}

function convertAagToAig(aagText, dir, name) {
  const aagPath = path.join(dir, `${name}.aag`);
  const aigPath = path.join(dir, `${name}.aig`);
  fs.writeFileSync(aagPath, aagText, "utf8");
  childProcess.execFileSync(aigtoaigPath, [aagPath, aigPath], { stdio: "pipe" });
  return aigPath;
}

function classify(text, cexText) {
  if (/Property proved/.test(text)) return "proved";
  if (/was asserted in frame/.test(text) || cexText !== null) return "cex";
  return "error";
}

function normalizeCex(text) {
  return text == null ? null : text.replace(/\s+$/g, "");
}

function runNative(aigPath, outPath) {
  const command = `read_aiger ${aigPath}; pdr; write_aiger_cex ${outPath}`;
  try {
    const stdout = childProcess.execFileSync(nativeAbcPath, ["-c", command], {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
    });
    const cex = fs.existsSync(outPath) ? fs.readFileSync(outPath, "utf8") : null;
    return { status: classify(stdout, cex), text: stdout, cex: normalizeCex(cex) };
  } catch (error) {
    const stdout = error.stdout ? String(error.stdout) : "";
    const stderr = error.stderr ? String(error.stderr) : "";
    return { status: "error", text: `${stdout}\n${stderr}`, cex: null };
  }
}

async function runWasm(aigPath) {
  const out = [];
  const err = [];
  try {
    const Module = await initAbc({
      noInitialRun: true,
      print: (s) => out.push(s),
      printErr: (s) => err.push(s),
    });
    Module.FS.writeFile("/abc.rc", abcRc);
    Module.FS.writeFile("/case.aig", fs.readFileSync(aigPath));
    Module.callMain(["-c", "read_aiger /case.aig; pdr; write_aiger_cex /case.out"]);
    const cex = Module.FS.analyzePath("/case.out").exists
      ? Module.FS.readFile("/case.out", { encoding: "utf8" })
      : null;
    const text = `${out.join("\n")}\n${err.join("\n")}`;
    return { status: classify(text, cex), text, cex: normalizeCex(cex) };
  } catch (error) {
    const text = `${out.join("\n")}\n${err.join("\n")}\n${error && error.stack ? error.stack : error}`;
    return { status: "error", text, cex: null };
  }
}

(async function main() {
  if (!Number.isInteger(caseCount) || caseCount < 1) {
    throw new Error(`--cases must be a positive integer, got ${caseCount}`);
  }
  if (!Number.isInteger(seed)) {
    throw new Error(`--seed must be an integer, got ${seed}`);
  }

  const rng = makeRng(seed);
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "abc-aiger-compare-"));
  for (let i = 0; i < caseCount; i++) {
    const name = `case-${String(i).padStart(4, "0")}`;
    const aigPath = convertAagToAig(generateAag(rng, i), dir, name);
    const nativeResult = runNative(aigPath, path.join(dir, `${name}.native.out`));
    const wasmResult = await runWasm(aigPath);

    if (nativeResult.status !== wasmResult.status) {
      throw new Error([
        `${name}: status mismatch native=${nativeResult.status} wasm=${wasmResult.status}`,
        `case directory: ${dir}`,
        "--- native ---",
        nativeResult.text,
        "--- wasm ---",
        wasmResult.text,
      ].join("\n"));
    }
    if (nativeResult.status === "cex" && nativeResult.cex !== wasmResult.cex) {
      throw new Error([
        `${name}: CEX mismatch`,
        `case directory: ${dir}`,
        "--- native cex ---",
        nativeResult.cex,
        "--- wasm cex ---",
        wasmResult.cex,
      ].join("\n"));
    }
  }

  console.log(`abc native/wasm AIGER compare passed (${caseCount} cases, seed ${seed})`);
})().catch((error) => {
  console.error(error && error.stack ? error.stack : error);
  process.exit(1);
});
