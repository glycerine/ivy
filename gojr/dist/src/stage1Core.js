import { stage1IntrinsicSupportSource } from "./emitter/package.js";
const supportCache = new WeakMap();
let cachedSupportFactory;
export function createStage1RuntimeCore(runtime = {}) {
    const callerBuiltinImport = typeof runtime.builtinImport === "function"
        ? runtime.builtinImport
        : typeof runtime.__gojrBuiltinImport === "function"
            ? runtime.__gojrBuiltinImport
            : undefined;
    const callerIntrinsicZero = typeof runtime.intrinsicZero === "function"
        ? runtime.intrinsicZero
        : typeof runtime.__gojrIntrinsicZero === "function"
            ? runtime.__gojrIntrinsicZero
            : undefined;
    const callerIntrinsicDescriptor = typeof runtime.intrinsicDescriptor === "function"
        ? runtime.intrinsicDescriptor
        : typeof runtime.__gojrIntrinsicDescriptor === "function"
            ? runtime.__gojrIntrinsicDescriptor
            : undefined;
    const callerIntrinsicStruct = typeof runtime.intrinsicStruct === "function"
        ? runtime.intrinsicStruct
        : typeof runtime.__gojrIntrinsicStruct === "function"
            ? runtime.__gojrIntrinsicStruct
            : undefined;
    return {
        ...runtime,
        builtinImport(path, importsByPath, ops) {
            const custom = callerBuiltinImport?.(path, importsByPath, ops);
            return custom !== undefined ? custom : supportForOps(ops).__gojrBuiltinImport(path, importsByPath);
        },
        intrinsicZero(typeText, pkgPath, ops) {
            const custom = callerIntrinsicZero?.(typeText, pkgPath, ops);
            if (custom !== undefined)
                return custom;
            const support = supportForOps(ops);
            return support.__gojrSyncZero?.(typeText, pkgPath) ??
                support.__gojrAtomicZero?.(typeText) ??
                support.__gojrWeakZero?.(typeText);
        },
        intrinsicDescriptor(typeText, pkgPath, ops) {
            const custom = callerIntrinsicDescriptor?.(typeText, pkgPath, ops);
            if (custom !== undefined)
                return custom;
            const support = supportForOps(ops);
            return support.__gojrSyncDescriptorForTypeName?.(typeText, pkgPath) ??
                support.__gojrAtomicDescriptorForTypeName?.(typeText) ??
                support.__gojrWeakDescriptorForTypeName?.(typeText);
        },
        intrinsicStruct(typeName, value, pkgPath, ops) {
            const custom = callerIntrinsicStruct?.(typeName, value, pkgPath, ops);
            return custom !== undefined ? custom : supportForOps(ops).__gojrWeakStruct?.(typeName, value, pkgPath);
        }
    };
}
function supportForOps(ops) {
    const key = typeof ops === "object" && ops !== null ? ops : Object(ops);
    const cached = supportCache.get(key);
    if (cached)
        return cached;
    const support = stage1SupportFactory()(key);
    supportCache.set(key, support);
    return support;
}
function stage1SupportFactory() {
    if (cachedSupportFactory)
        return cachedSupportFactory;
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
`);
    return cachedSupportFactory;
}
