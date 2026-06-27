import { stage1SharedSupportBindingNames, stage1SharedSupportSource } from "./emitter/package.js";
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
    const callerStage1Helpers = typeof runtime.stage1Helpers === "function"
        ? runtime.stage1Helpers
        : typeof runtime.__gojrStage1Helpers === "function"
            ? runtime.__gojrStage1Helpers
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
        },
        defaultTextSink(stream) {
            if (typeof process !== "undefined")
                return undefined;
            const target = stream === "stderr" ? globalThis.console?.error || globalThis.console?.log : globalThis.console?.log;
            if (typeof target !== "function")
                return undefined;
            return (text) => target.call(globalThis.console, text);
        },
        stage1Helpers(state) {
            const custom = callerStage1Helpers?.(state);
            return custom !== undefined ? custom : supportForOps(state);
        }
    };
}
function supportForOps(ops) {
    const key = typeof ops === "object" && ops !== null ? ops : Object(ops);
    const cached = supportCache.get(key);
    if (cached)
        return cached;
    const support = stage1SupportFactory()(key.gojrPackageArtifact, key.__gojrImportPathsByQualifier ?? {}, key.__gojrTypeDescriptors ?? {}, key.__gojrActiveImportsByPath ?? {}, key.__gojrActivePackage ?? {}, key.__gojrActiveRuntimeOptions ?? {});
    supportCache.set(key, support);
    return support;
}
function stage1SupportFactory() {
    if (cachedSupportFactory)
        return cachedSupportFactory;
    cachedSupportFactory = new Function("gojrPackageArtifact", "__gojrImportPathsByQualifier", "__gojrTypeDescriptors", "__gojrActiveImportsByPath", "__gojrActivePackage", "__gojrActiveRuntimeOptions", `
${stage1SharedSupportSource()}
return { ${stage1SharedSupportBindingNames().join(", ")} };
`);
    return cachedSupportFactory;
}
