import { describe, expect, test } from "./testHarness.js";
import "../src/embeddedNodeBootstrap.js";

describe("embedded Node bootstrap", () => {
  test("loads bundled modules through the shared bootstrap transformer", () => {
    const root = globalThis as Record<string, any>;
    const createLoader = root.__gojrCreateEmbeddedModuleLoader as (json: string) => (specifier: string, parent?: string) => Record<string, any>;
    expect(createLoader).toBeDefined();

    const requireEmbedded = createLoader(JSON.stringify({
      modules: [
        {
          path: "/dep.js",
          source: "export const value = 41;"
        },
        {
          path: "/main.js",
          source: "import { value as base } from './dep.js';\nexport const answer = base + 1;\nexport function show() { return answer; }"
        }
      ]
    }));

    const module = requireEmbedded("/main.js");
    expect(module.answer).toBe(42);
    expect(module.show()).toBe(42);
  });

  test("maps terminal Ctrl-D reads to EOF without changing piped EOT bytes", () => {
    const root = globalThis as Record<string, any>;
    const install = root.__gojrInstallEmbeddedRuntime as (json: string) => void;
    expect(install).toBeDefined();

    const previousRead = root.__gojrReadSync;
    const previousWrite = root.__gojrWriteSync;
    const previousRequire = root.require;
    const installWithTTY = (stdinIsTTY: boolean) => {
      delete root.__gojrReadSync;
      delete root.__gojrWriteSync;
      root.require = (specifier: string) => {
        if (specifier === "node:fs" || specifier === "fs") {
          return {
            readSync(_fd: number, buffer: Uint8Array, offset: number) {
              buffer[offset] = 4;
              return 1;
            },
            writeSync(_fd: number, _buffer: Uint8Array, _offset: number, length: number) {
              return length;
            }
          };
        }
        if (specifier === "node:tty") {
          return {
            isatty(fd: number) {
              return fd === 0 && stdinIsTTY;
            }
          };
        }
        throw new Error(`unexpected host require ${specifier}`);
      };
      install(JSON.stringify({
        modules: [{
          path: "/src/index.js",
          source: "export function runtimeOptionsFromEnvironment() { return {}; }\nexport class GoJuniorSession { constructor(_options) {} }\n"
        }]
      }));
    };

    try {
      installWithTTY(true);
      const terminalBuffer = new Uint8Array(1);
      expect(root.__gojrReadSync(0, terminalBuffer, 0, 1, null)).toBe(0);

      installWithTTY(false);
      const pipeBuffer = new Uint8Array(1);
      expect(root.__gojrReadSync(0, pipeBuffer, 0, 1, null)).toBe(1);
      expect(pipeBuffer[0]).toBe(4);
    } finally {
      if (previousRead) root.__gojrReadSync = previousRead;
      else delete root.__gojrReadSync;
      if (previousWrite) root.__gojrWriteSync = previousWrite;
      else delete root.__gojrWriteSync;
      if (previousRequire) root.require = previousRequire;
      else delete root.require;
    }
  });
});
