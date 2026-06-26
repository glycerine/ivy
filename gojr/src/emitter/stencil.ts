import type { WasmImport, WasmLocalDecl, WasmValueType } from "./wasm/module.js";

export type SlotKind =
  | "identifier"
  | "property"
  | "expression"
  | "statementList"
  | "typeDescriptor"
  | "json"
  | "stringLiteral"
  | "label"
  | "rawTrustedJs";

export interface JsStencil {
  name: string;
  template: string;
  slots: Record<string, SlotKind>;
}

export interface WasmFunctionSignature {
  params: WasmValueType[];
  results: WasmValueType[];
}

export type WasmPatchHoleKind =
  | "typeIndex"
  | "functionIndex"
  | "localIndex"
  | "globalIndex"
  | "branchDepth"
  | "lebI32"
  | "lebI64"
  | "memoryOffset"
  | "dataOffset"
  | "callTarget";

export interface WasmPatchHole {
  kind: WasmPatchHoleKind;
  offset: number;
  width: number;
  signed: boolean;
  semanticName: string;
}

export interface WasmMemoryRequirement {
  importMemory: boolean;
  module?: string;
  name?: string;
  minPages?: number;
}

export interface WasmStencilMetadata {
  sourceLanguage?: string;
  sourceName?: string;
  description?: string;
}

export interface WasmStencil {
  name: string;
  signature: WasmFunctionSignature;
  locals: WasmLocalDecl[];
  bodyBytes: Uint8Array;
  holes: WasmPatchHole[];
  imports: WasmImport[];
  memory: WasmMemoryRequirement;
  metadata: WasmStencilMetadata;
}
