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
const seconds = Number(argValue("--seconds", "0"));
const progressEvery = Number(argValue("--progress-every", "50"));
const logPathArg = argValue("--log", "");
const logPath = logPathArg ? path.resolve(logPathArg) : "";

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

function logWrite(text) {
  if (logPath) fs.appendFileSync(logPath, text, "utf8");
}

function section(title, body) {
  return `=== ${title} ===\n${body == null || body === "" ? "(empty)\n" : `${body.replace(/\s+$/g, "")}\n`}`;
}

function writeCaseLog(caseInfo) {
  logWrite([
    `=== case ${caseInfo.name} seed ${caseInfo.seed} index ${caseInfo.index} ===`,
    `case directory: ${caseInfo.dir}`,
    `native command: ${caseInfo.nativeCommand}`,
    "wasm command: read_aiger /case.aig; pdr; write_aiger_cex /case.out",
    section("input", caseInfo.aagText),
    section("input aig path", caseInfo.aigPath),
    section("input aig base64", caseInfo.aigBase64),
    section("native status", caseInfo.nativeResult.status),
    section("native stdout", caseInfo.nativeResult.stdout),
    section("native stderr", caseInfo.nativeResult.stderr),
    section("native cex", caseInfo.nativeResult.cex),
    section("wasm status", caseInfo.wasmResult.status),
    section("wasm stdout", caseInfo.wasmResult.stdout),
    section("wasm stderr", caseInfo.wasmResult.stderr),
    section("wasm cex", caseInfo.wasmResult.cex),
    "",
  ].join("\n"));
}

function runNative(aigPath, outPath) {
  const command = `read_aiger ${aigPath}; pdr; write_aiger_cex ${outPath}`;
  const result = childProcess.spawnSync(nativeAbcPath, ["-c", command], {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });
  const stdout = result.stdout ? String(result.stdout) : "";
  const stderr = result.stderr ? String(result.stderr) : "";
  const cex = fs.existsSync(outPath) ? fs.readFileSync(outPath, "utf8") : null;
  const status = result.status === 0 && !result.error ? classify(`${stdout}\n${stderr}`, cex) : "error";
  return {
    command,
    status,
    stdout,
    stderr: result.error ? `${stderr}\n${result.error.stack || result.error}` : stderr,
    text: `${stdout}\n${stderr}`,
    cex: normalizeCex(cex),
  };
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
    return {
      status: classify(text, cex),
      stdout: out.join("\n"),
      stderr: err.join("\n"),
      text,
      cex: normalizeCex(cex),
    };
  } catch (error) {
    const stderr = `${err.join("\n")}\n${error && error.stack ? error.stack : error}`;
    const text = `${out.join("\n")}\n${stderr}`;
    return { status: "error", stdout: out.join("\n"), stderr, text, cex: null };
  }
}

(async function main() {
  if (!Number.isInteger(caseCount) || caseCount < 1) {
    throw new Error(`--cases must be a positive integer, got ${caseCount}`);
  }
  if (!Number.isInteger(seed)) {
    throw new Error(`--seed must be an integer, got ${seed}`);
  }
  if (!Number.isFinite(seconds) || seconds < 0) {
    throw new Error(`--seconds must be a non-negative number, got ${seconds}`);
  }
  if (!Number.isInteger(progressEvery) || progressEvery < 1) {
    throw new Error(`--progress-every must be a positive integer, got ${progressEvery}`);
  }

  const rng = makeRng(seed);
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "abc-aiger-compare-"));
  const startTime = Date.now();
  const deadline = seconds > 0 ? Date.now() + seconds * 1000 : 0;
  let i = 0;
  logWrite([
    `=== run seed ${seed} ===`,
    `started: ${new Date(startTime).toISOString()}`,
    `mode: ${seconds > 0 ? `${seconds}s` : `${caseCount} cases`}`,
    `case directory: ${dir}`,
    "",
  ].join("\n"));
  console.log(seconds > 0
    ? `abc native/wasm AIGER compare started (seed ${seed}, ${seconds}s, dir ${dir}${logPath ? `, log ${logPath}` : ""})`
    : `abc native/wasm AIGER compare started (${caseCount} cases, seed ${seed}, dir ${dir}${logPath ? `, log ${logPath}` : ""})`);

  while (seconds > 0 ? Date.now() < deadline : i < caseCount) {
    const name = `case-${String(i).padStart(4, "0")}`;
    const aagText = generateAag(rng, i);
    const aigPath = convertAagToAig(aagText, dir, name);
    const aigBase64 = fs.readFileSync(aigPath).toString("base64");
    const nativeResult = runNative(aigPath, path.join(dir, `${name}.native.out`));
    const wasmResult = await runWasm(aigPath);
    writeCaseLog({
      name,
      seed,
      index: i,
      dir,
      aagText,
      aigPath,
      aigBase64,
      nativeCommand: nativeResult.command,
      nativeResult,
      wasmResult,
    });

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
    i++;
    if (i % progressEvery === 0) {
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      console.log(`abc native/wasm AIGER compare progress: seed ${seed}, ${i} cases, ${elapsed}s`);
    }
  }

  console.log(seconds > 0
    ? `abc native/wasm AIGER compare passed (${i} cases, seed ${seed}, ${seconds}s budget)`
    : `abc native/wasm AIGER compare passed (${i} cases, seed ${seed})`);
})().catch((error) => {
  console.error(error && error.stack ? error.stack : error);
  process.exit(1);
});
