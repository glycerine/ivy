import { describe, expect, test } from "./testHarness.js";
import { Environment } from "../src/transactionalEnvironment.js";

describe("Transactional Environment", () => {
  test("keeps transaction writes in a copy-on-write overlay until commit", () => {
    const root = Environment.fromEntries<string, number>([
      ["a", 1],
      ["b", 2]
    ]);

    const tx = root.beginTransaction();
    tx.environment.set("a", 10);
    tx.environment.set("c", 3);

    expect(tx.environment.get("a")).toBe(10);
    expect(tx.environment.get("b")).toBe(2);
    expect(tx.environment.get("c")).toBe(3);
    expect(root.get("a")).toBe(1);
    expect(root.has("c")).toBe(false);

    tx.commit();

    expect(root.get("a")).toBe(10);
    expect(root.get("b")).toBe(2);
    expect(root.get("c")).toBe(3);
  });

  test("rolls back overlay writes and deletes without changing the parent", () => {
    const root = Environment.fromEntries<string, number>([
      ["a", 1],
      ["b", 2]
    ]);

    const tx = root.beginTransaction();
    tx.environment.set("a", 10);
    tx.environment.delete("b");
    tx.environment.set("c", 3);

    expect(tx.environment.get("a")).toBe(10);
    expect(tx.environment.has("b")).toBe(false);
    expect(tx.environment.get("c")).toBe(3);

    tx.rollback();

    expect(root.get("a")).toBe(1);
    expect(root.get("b")).toBe(2);
    expect(root.has("c")).toBe(false);
  });

  test("overlay bindings shadow parent bindings while other names fall through", () => {
    const root = Environment.fromEntries<string, string>([
      ["name", "global"],
      ["shared", "visible from parent"]
    ]);

    const tx = root.beginTransaction();
    tx.environment.set("name", "transaction-local");

    expect(tx.environment.has("name")).toBe(true);
    expect(tx.environment.get("name")).toBe("transaction-local");
    expect(tx.environment.get("shared")).toBe("visible from parent");
    expect(root.get("name")).toBe("global");

    tx.rollback();

    expect(root.get("name")).toBe("global");
    expect(root.get("shared")).toBe("visible from parent");
  });

  test("makes completed transactions unusable", () => {
    const root = Environment.fromEntries<string, number>([["a", 1]]);
    const tx = root.beginTransaction();

    tx.environment.set("a", 2);
    tx.commit();

    expect(() => tx.environment.set("a", 3)).toThrow(/closed transaction/);
    expect(() => tx.rollback()).toThrow(/already closed/);
  });
});
