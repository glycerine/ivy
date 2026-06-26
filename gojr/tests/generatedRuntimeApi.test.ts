import { describe, expect, test } from "./testHarness.js";
import {
  RuntimeInterfaceValue,
  RuntimeMap,
  RuntimePointer,
  gojrGeneratedRuntimeApi
} from "../src/index.js";

describe("GoJr generated-code runtime API", () => {
  test("creates package contexts and declares typed package variables", () => {
    const ctx = gojrGeneratedRuntimeApi.createPackageContext({
      importPath: "example.com/generated",
      packageName: "generated"
    });

    gojrGeneratedRuntimeApi.declarePackageVar(ctx, "Count", 3n, "int64");
    gojrGeneratedRuntimeApi.declarePackageVar(ctx, "Name", "ivy", "string");
    const finished = gojrGeneratedRuntimeApi.finishPackage(ctx);

    expect(ctx.package.Count).toBe(3n);
    expect((ctx.package.Name as { text(): string }).text()).toBe("ivy");
    expect(finished.diagnostics).toEqual([]);
    expect(finished.package).toBe(ctx.package);
    expect(finished.runtime).toBe(ctx.runtime);
  });

  test("wraps map helpers around the shared RuntimeMap implementation", () => {
    const ctx = gojrGeneratedRuntimeApi.createPackageContext({});
    const map = gojrGeneratedRuntimeApi.makeMap("string", "int64", [["a", 1n]], ctx);

    expect(map).toBeInstanceOf(RuntimeMap);
    expect(gojrGeneratedRuntimeApi.mapGet(map, "a")).toBe(1n);
    expect(gojrGeneratedRuntimeApi.mapGetOk(map, "missing")).toEqual([0n, false]);
    gojrGeneratedRuntimeApi.mapSet(map, "b", 2n);
    expect(gojrGeneratedRuntimeApi.mapGetOk(map, "b")).toEqual([2n, true]);
    gojrGeneratedRuntimeApi.mapDelete(map, "a");
    expect(gojrGeneratedRuntimeApi.mapGetOk(map, "a")).toEqual([0n, false]);
  });

  test("creates slices, shared slice views, append, and copy helpers", () => {
    const ctx = gojrGeneratedRuntimeApi.createPackageContext({});
    const slice = gojrGeneratedRuntimeApi.makeSlice("int64", 3, 6, ctx);

    expect(slice).toEqual([0n, 0n, 0n]);
    gojrGeneratedRuntimeApi.sliceSet(slice, 1, 11n);
    expect(gojrGeneratedRuntimeApi.sliceGet(slice, 1)).toBe(11n);

    const view = gojrGeneratedRuntimeApi.sliceRange(slice, 1, 3, 5);
    gojrGeneratedRuntimeApi.sliceSet(view, 0, 22n);
    expect(gojrGeneratedRuntimeApi.sliceGet(slice, 1)).toBe(22n);

    const appended = gojrGeneratedRuntimeApi.append(slice, 33n);
    expect(appended).toEqual([0n, 22n, 0n, 33n]);

    const dst = gojrGeneratedRuntimeApi.makeSlice("int64", 2, 2, ctx);
    expect(gojrGeneratedRuntimeApi.copy(dst, appended)).toBe(2);
    expect(dst).toEqual([0n, 22n]);
  });

  test("creates pointer helpers without losing identity", () => {
    let value = 7n;
    const pointer = gojrGeneratedRuntimeApi.newPointer("int64", () => value, (next) => {
      value = next as bigint;
    }, "test-pointer");

    expect(pointer).toBeInstanceOf(RuntimePointer);
    expect(pointer.identityKey()).toBe("test-pointer");
    pointer.set(9n);
    expect(value).toBe(9n);
    expect(pointer.get()).toBe(9n);
  });

  test("wraps channel send and receive helpers around the shared scheduler", async () => {
    const ctx = gojrGeneratedRuntimeApi.createPackageContext({});
    const channel = gojrGeneratedRuntimeApi.makeChan("int64", 1, ctx);

    await gojrGeneratedRuntimeApi.chanSend(ctx, channel, 42n);
    expect(await gojrGeneratedRuntimeApi.chanRecv(ctx, channel)).toEqual([42n, true]);
  });

  test("runs generated defers in LIFO order and supports recover", async () => {
    const ctx = gojrGeneratedRuntimeApi.createPackageContext({});
    const order: string[] = [];

    await gojrGeneratedRuntimeApi.deferScope(ctx, async () => {
      gojrGeneratedRuntimeApi.defer(ctx, () => {
        order.push("first");
        return null;
      });
      gojrGeneratedRuntimeApi.defer(ctx, () => {
        order.push("second");
        return null;
      });
      return null;
    });
    expect(order).toEqual(["second", "first"]);

    let recovered: unknown;
    await gojrGeneratedRuntimeApi.deferScope(ctx, async () => {
      gojrGeneratedRuntimeApi.defer(ctx, () => {
        recovered = gojrGeneratedRuntimeApi.recover(ctx);
        return null;
      });
      gojrGeneratedRuntimeApi.panic("boom");
    });
    expect(recovered).toBe("boom");
  });

  test("boxes values into interface values through the shared interface path", () => {
    const ctx = gojrGeneratedRuntimeApi.createPackageContext({});
    const value = gojrGeneratedRuntimeApi.toInterface(12n, "any", ctx);

    expect(value).toBeInstanceOf(RuntimeInterfaceValue);
    expect(value.interfaceType).toBe("any");
    expect(value.value).toBe(12n);
  });
});
