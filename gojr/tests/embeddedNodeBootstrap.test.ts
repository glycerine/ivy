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
});
