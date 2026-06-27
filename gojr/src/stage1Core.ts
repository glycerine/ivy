import { stage1IntrinsicSupportSource } from "./emitter/package.js";

type Stage1IntrinsicOps = Record<PropertyKey, unknown>;

interface Stage1BuiltinSupport {
  __gojrBuiltinImport(path: string, importsByPath: Record<string, unknown>): unknown;
  __gojrSyncZero?: (typeText: string, pkgPath?: string) => unknown;
  __gojrAtomicZero?: (typeText: string) => unknown;
  __gojrWeakZero?: (typeText: string) => unknown;
  __gojrSyncDescriptorForTypeName?: (typeText: string, pkgPath?: string) => unknown;
  __gojrAtomicDescriptorForTypeName?: (typeText: string) => unknown;
  __gojrWeakDescriptorForTypeName?: (typeText: string) => unknown;
  __gojrWeakStruct?: (typeName: string, value: unknown, pkgPath?: string) => unknown;
}

export type Stage1RuntimeCore = Record<string, unknown> & {
  builtinImport(path: string, importsByPath: Record<string, unknown>, ops: Stage1IntrinsicOps): unknown;
  intrinsicZero(typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps): unknown;
  intrinsicDescriptor(typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps): unknown;
  intrinsicStruct(typeName: string, value: unknown, pkgPath: string | undefined, ops: Stage1IntrinsicOps): unknown;
};

type Stage1SupportFactory = (ops: Stage1IntrinsicOps) => Stage1BuiltinSupport;

const supportCache = new WeakMap<object, Stage1BuiltinSupport>();
let cachedSupportFactory: Stage1SupportFactory | undefined;

export function createStage1RuntimeCore(runtime: Record<string, unknown> = {}): Stage1RuntimeCore {
  const callerBuiltinImport = typeof runtime.builtinImport === "function"
    ? runtime.builtinImport as (path: string, importsByPath: Record<string, unknown>, ops: Stage1IntrinsicOps) => unknown
    : typeof runtime.__gojrBuiltinImport === "function"
      ? runtime.__gojrBuiltinImport as (path: string, importsByPath: Record<string, unknown>, ops: Stage1IntrinsicOps) => unknown
      : undefined;
  const callerIntrinsicZero = typeof runtime.intrinsicZero === "function"
    ? runtime.intrinsicZero as (typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps) => unknown
    : typeof runtime.__gojrIntrinsicZero === "function"
      ? runtime.__gojrIntrinsicZero as (typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps) => unknown
      : undefined;
  const callerIntrinsicDescriptor = typeof runtime.intrinsicDescriptor === "function"
    ? runtime.intrinsicDescriptor as (typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps) => unknown
    : typeof runtime.__gojrIntrinsicDescriptor === "function"
      ? runtime.__gojrIntrinsicDescriptor as (typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps) => unknown
      : undefined;
  const callerIntrinsicStruct = typeof runtime.intrinsicStruct === "function"
    ? runtime.intrinsicStruct as (typeName: string, value: unknown, pkgPath: string | undefined, ops: Stage1IntrinsicOps) => unknown
    : typeof runtime.__gojrIntrinsicStruct === "function"
      ? runtime.__gojrIntrinsicStruct as (typeName: string, value: unknown, pkgPath: string | undefined, ops: Stage1IntrinsicOps) => unknown
      : undefined;

  return {
    ...runtime,
    builtinImport(path: string, importsByPath: Record<string, unknown>, ops: Stage1IntrinsicOps) {
      const custom = callerBuiltinImport?.(path, importsByPath, ops);
      return custom !== undefined ? custom : supportForOps(ops).__gojrBuiltinImport(path, importsByPath);
    },
    intrinsicZero(typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps) {
      const custom = callerIntrinsicZero?.(typeText, pkgPath, ops);
      if (custom !== undefined) return custom;
      const support = supportForOps(ops);
      return support.__gojrSyncZero?.(typeText, pkgPath) ??
        support.__gojrAtomicZero?.(typeText) ??
        support.__gojrWeakZero?.(typeText);
    },
    intrinsicDescriptor(typeText: string, pkgPath: string | undefined, ops: Stage1IntrinsicOps) {
      const custom = callerIntrinsicDescriptor?.(typeText, pkgPath, ops);
      if (custom !== undefined) return custom;
      const support = supportForOps(ops);
      return support.__gojrSyncDescriptorForTypeName?.(typeText, pkgPath) ??
        support.__gojrAtomicDescriptorForTypeName?.(typeText) ??
        support.__gojrWeakDescriptorForTypeName?.(typeText);
    },
    intrinsicStruct(typeName: string, value: unknown, pkgPath: string | undefined, ops: Stage1IntrinsicOps) {
      const custom = callerIntrinsicStruct?.(typeName, value, pkgPath, ops);
      return custom !== undefined ? custom : supportForOps(ops).__gojrWeakStruct?.(typeName, value, pkgPath);
    }
  };
}

function supportForOps(ops: Stage1IntrinsicOps): Stage1BuiltinSupport {
  const key = typeof ops === "object" && ops !== null ? ops : Object(ops);
  const cached = supportCache.get(key);
  if (cached) return cached;
  const support = stage1SupportFactory()(key as Stage1IntrinsicOps);
  supportCache.set(key, support);
  return support;
}

function stage1SupportFactory(): Stage1SupportFactory {
  if (cachedSupportFactory) return cachedSupportFactory;
  cachedSupportFactory = new Function("ops", `
const __gojrScope = new Proxy(ops || {}, {
  has(target, key) {
    return key in target || key in globalThis;
  },
  get(target, key) {
    return key in target ? target[key] : globalThis[key];
  }
});
with (__gojrScope) {
${stage1IntrinsicSupportSource()}
return {
  __gojrBuiltinImport,
  __gojrSyncZero,
  __gojrAtomicZero,
  __gojrWeakZero,
  __gojrSyncDescriptorForTypeName,
  __gojrAtomicDescriptorForTypeName,
  __gojrWeakDescriptorForTypeName,
  __gojrWeakStruct
};
}
`) as Stage1SupportFactory;
  return cachedSupportFactory;
}
