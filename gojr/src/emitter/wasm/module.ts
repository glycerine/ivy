export type WasmValueType =
  | "i32"
  | "i64"
  | "f32"
  | "f64"
  | "v128"
  | "funcref"
  | "externref"
  | `unknown(0x${string})`;

export interface WasmFunctionType {
  params: WasmValueType[];
  results: WasmValueType[];
}

export type WasmImportKind = "func" | "table" | "memory" | "global";
export type WasmExportKind = WasmImportKind;

export interface WasmLimits {
  min: number;
  max?: number;
}

export interface WasmImport {
  module: string;
  name: string;
  kind: WasmImportKind;
  index: number;
  typeIndex?: number;
  limits?: WasmLimits;
}

export interface WasmExport {
  name: string;
  kind: WasmExportKind;
  index: number;
}

export interface WasmLocalDecl {
  count: number;
  type: WasmValueType;
}

export interface WasmCodeBody {
  functionIndex: number;
  codeIndex: number;
  locals: WasmLocalDecl[];
  bodyBytes: Uint8Array;
  instructionBytes: Uint8Array;
}

export interface WasmSection {
  id: number;
  name: string;
  payloadStart: number;
  payloadEnd: number;
}

export interface WasmCustomSection extends WasmSection {
  customName: string;
}

export interface WasmModuleInfo {
  bytes: Uint8Array;
  sections: WasmSection[];
  customSections: WasmCustomSection[];
  types: WasmFunctionType[];
  imports: WasmImport[];
  functionTypeIndices: number[];
  exports: WasmExport[];
  code: WasmCodeBody[];
  importedFunctionCount: number;
}

export interface ExtractWasmFunctionOptions {
  allowImportedMemory?: boolean;
  memoryModule?: string;
  memoryName?: string;
}

export interface ExtractedWasmFunction {
  module: WasmModuleInfo;
  exportName: string;
  functionIndex: number;
  typeIndex: number;
  signature: WasmFunctionType;
  locals: WasmLocalDecl[];
  bodyBytes: Uint8Array;
  instructionBytes: Uint8Array;
  importedMemory?: WasmImport;
}

export function parseWasmModule(input: Uint8Array | ArrayBuffer): WasmModuleInfo {
  const bytes = input instanceof Uint8Array ? input : new Uint8Array(input);
  if (bytes.length < 8 ||
    bytes[0] !== 0x00 ||
    bytes[1] !== 0x61 ||
    bytes[2] !== 0x73 ||
    bytes[3] !== 0x6d ||
    bytes[4] !== 0x01 ||
    bytes[5] !== 0x00 ||
    bytes[6] !== 0x00 ||
    bytes[7] !== 0x00) {
    throw new Error("invalid WebAssembly module header");
  }

  const sections: WasmSection[] = [];
  const customSections: WasmCustomSection[] = [];
  let types: WasmFunctionType[] = [];
  let imports: WasmImport[] = [];
  let functionTypeIndices: number[] = [];
  let exports: WasmExport[] = [];
  let rawCodeBodies: Uint8Array[] = [];
  let offset = 8;

  while (offset < bytes.length) {
    const id = requireByte(bytes, offset);
    const size = readU32(bytes, offset + 1);
    const payloadStart = size.next;
    const payloadEnd = payloadStart + size.value;
    if (payloadEnd > bytes.length) throw new Error(`truncated WebAssembly section ${id}`);
    const section: WasmSection = {
      id,
      name: sectionName(id),
      payloadStart,
      payloadEnd
    };
    sections.push(section);
    const payload = bytes.subarray(payloadStart, payloadEnd);

    switch (id) {
      case 0: {
        const named = readName(bytes, payloadStart, payloadEnd);
        customSections.push({
          ...section,
          customName: named.value
        });
        break;
      }
      case 1:
        types = parseTypeSection(payload, payloadStart);
        break;
      case 2:
        imports = parseImportSection(payload, payloadStart);
        break;
      case 3:
        functionTypeIndices = parseFunctionSection(payload, payloadStart);
        break;
      case 7:
        exports = parseExportSection(payload, payloadStart);
        break;
      case 10:
        rawCodeBodies = parseRawCodeSection(payload, payloadStart);
        break;
      default:
        break;
    }

    offset = payloadEnd;
  }

  if (offset !== bytes.length) throw new Error("trailing bytes after WebAssembly module");
  const importedFunctionCount = imports.filter((item) => item.kind === "func").length;
  const code = rawCodeBodies.map((bodyBytes, codeIndex) => parseCodeBody(bodyBytes, importedFunctionCount + codeIndex, codeIndex));
  if (rawCodeBodies.length !== functionTypeIndices.length) {
    throw new Error(`function section count ${functionTypeIndices.length} does not match code section count ${rawCodeBodies.length}`);
  }

  return {
    bytes: bytes.slice(),
    sections,
    customSections,
    types,
    imports,
    functionTypeIndices,
    exports,
    code,
    importedFunctionCount
  };
}

export function extractWasmFunction(
  input: Uint8Array | ArrayBuffer,
  exportName: string,
  options: ExtractWasmFunctionOptions = {}
): ExtractedWasmFunction {
  const module = parseWasmModule(input);
  const exported = module.exports.find((item) => item.name === exportName && item.kind === "func");
  if (!exported) throw new Error(`WebAssembly function export ${exportName} not found`);
  if (exported.index < module.importedFunctionCount) {
    throw new Error(`WebAssembly function export ${exportName} is an imported function`);
  }
  const codeIndex = exported.index - module.importedFunctionCount;
  const body = module.code[codeIndex];
  if (!body) throw new Error(`WebAssembly function export ${exportName} has no code body`);
  const typeIndex = module.functionTypeIndices[codeIndex];
  if (typeIndex === undefined) throw new Error(`WebAssembly function export ${exportName} has no type index`);
  const signature = module.types[typeIndex];
  if (!signature) throw new Error(`WebAssembly function export ${exportName} has invalid type index ${typeIndex}`);

  const memoryModule = options.memoryModule ?? "env";
  const memoryName = options.memoryName ?? "memory";
  let importedMemory: WasmImport | undefined;
  for (const imported of module.imports) {
    if (imported.kind === "memory") {
      if (options.allowImportedMemory === false) {
        throw new Error(`WebAssembly function export ${exportName} unexpectedly imports memory`);
      }
      if (imported.module !== memoryModule || imported.name !== memoryName) {
        throw new Error(`WebAssembly function export ${exportName} imports unexpected memory ${imported.module}.${imported.name}`);
      }
      importedMemory = imported;
      continue;
    }
    throw new Error(`WebAssembly function export ${exportName} has unexpected ${imported.kind} import ${imported.module}.${imported.name}`);
  }

  const out: ExtractedWasmFunction = {
    module,
    exportName,
    functionIndex: exported.index,
    typeIndex,
    signature,
    locals: body.locals,
    bodyBytes: body.bodyBytes,
    instructionBytes: body.instructionBytes
  };
  if (importedMemory) out.importedMemory = importedMemory;
  return out;
}

function parseTypeSection(payload: Uint8Array, absoluteStart: number): WasmFunctionType[] {
  const cursor = new Cursor(payload, absoluteStart);
  const count = cursor.u32();
  const out: WasmFunctionType[] = [];
  for (let index = 0; index < count; index += 1) {
    const marker = cursor.byte();
    if (marker !== 0x60) throw new Error(`unsupported WebAssembly type form 0x${marker.toString(16)}`);
    out.push({
      params: parseValueTypeVector(cursor),
      results: parseValueTypeVector(cursor)
    });
  }
  cursor.done();
  return out;
}

function parseImportSection(payload: Uint8Array, absoluteStart: number): WasmImport[] {
  const cursor = new Cursor(payload, absoluteStart);
  const count = cursor.u32();
  const out: WasmImport[] = [];
  const nextIndex = new Map<WasmImportKind, number>([
    ["func", 0],
    ["table", 0],
    ["memory", 0],
    ["global", 0]
  ]);
  for (let index = 0; index < count; index += 1) {
    const module = cursor.name();
    const name = cursor.name();
    const kindByte = cursor.byte();
    const kind = importExportKind(kindByte);
    const itemIndex = nextIndex.get(kind) ?? 0;
    nextIndex.set(kind, itemIndex + 1);
    const base = { module, name, kind, index: itemIndex };
    switch (kind) {
      case "func":
        out.push({ ...base, typeIndex: cursor.u32() });
        break;
      case "table":
        cursor.byte();
        skipLimits(cursor);
        out.push(base);
        break;
      case "memory":
        out.push({ ...base, limits: parseLimits(cursor) });
        break;
      case "global":
        cursor.byte();
        cursor.byte();
        out.push(base);
        break;
    }
  }
  cursor.done();
  return out;
}

function parseFunctionSection(payload: Uint8Array, absoluteStart: number): number[] {
  const cursor = new Cursor(payload, absoluteStart);
  const count = cursor.u32();
  const out: number[] = [];
  for (let index = 0; index < count; index += 1) out.push(cursor.u32());
  cursor.done();
  return out;
}

function parseExportSection(payload: Uint8Array, absoluteStart: number): WasmExport[] {
  const cursor = new Cursor(payload, absoluteStart);
  const count = cursor.u32();
  const out: WasmExport[] = [];
  for (let index = 0; index < count; index += 1) {
    out.push({
      name: cursor.name(),
      kind: importExportKind(cursor.byte()),
      index: cursor.u32()
    });
  }
  cursor.done();
  return out;
}

function parseRawCodeSection(payload: Uint8Array, absoluteStart: number): Uint8Array[] {
  const cursor = new Cursor(payload, absoluteStart);
  const count = cursor.u32();
  const out: Uint8Array[] = [];
  for (let index = 0; index < count; index += 1) {
    const size = cursor.u32();
    out.push(cursor.readBytes(size));
  }
  cursor.done();
  return out;
}

function parseCodeBody(bodyBytes: Uint8Array, functionIndex: number, codeIndex: number): WasmCodeBody {
  const cursor = new Cursor(bodyBytes, 0);
  const localGroupCount = cursor.u32();
  const locals: WasmLocalDecl[] = [];
  for (let index = 0; index < localGroupCount; index += 1) {
    locals.push({
      count: cursor.u32(),
      type: valueType(cursor.byte())
    });
  }
  const instructionBytes = bodyBytes.slice(cursor.offset);
  return {
    functionIndex,
    codeIndex,
    locals,
    bodyBytes: bodyBytes.slice(),
    instructionBytes
  };
}

function parseValueTypeVector(cursor: Cursor): WasmValueType[] {
  const count = cursor.u32();
  const out: WasmValueType[] = [];
  for (let index = 0; index < count; index += 1) out.push(valueType(cursor.byte()));
  return out;
}

function parseLimits(cursor: Cursor): WasmLimits {
  const flag = cursor.byte();
  const min = cursor.u32();
  if ((flag & 0x01) === 0) return { min };
  return {
    min,
    max: cursor.u32()
  };
}

function skipLimits(cursor: Cursor): void {
  parseLimits(cursor);
}

function importExportKind(kind: number): WasmImportKind {
  switch (kind) {
    case 0: return "func";
    case 1: return "table";
    case 2: return "memory";
    case 3: return "global";
    default: throw new Error(`unsupported WebAssembly import/export kind ${kind}`);
  }
}

function valueType(byte: number): WasmValueType {
  switch (byte) {
    case 0x7f: return "i32";
    case 0x7e: return "i64";
    case 0x7d: return "f32";
    case 0x7c: return "f64";
    case 0x7b: return "v128";
    case 0x70: return "funcref";
    case 0x6f: return "externref";
    default: return `unknown(0x${byte.toString(16)})`;
  }
}

function sectionName(id: number): string {
  switch (id) {
    case 0: return "custom";
    case 1: return "type";
    case 2: return "import";
    case 3: return "function";
    case 4: return "table";
    case 5: return "memory";
    case 6: return "global";
    case 7: return "export";
    case 8: return "start";
    case 9: return "element";
    case 10: return "code";
    case 11: return "data";
    case 12: return "data-count";
    default: return `section-${id}`;
  }
}

function readU32(bytes: Uint8Array, offset: number): { value: number; next: number } {
  let result = 0;
  let shift = 0;
  for (let index = 0; index < 5; index += 1) {
    const byte = requireByte(bytes, offset + index);
    result |= (byte & 0x7f) << shift;
    if ((byte & 0x80) === 0) return { value: result >>> 0, next: offset + index + 1 };
    shift += 7;
  }
  throw new Error(`unsigned LEB128 value at ${offset} is too large`);
}

function readName(bytes: Uint8Array, offset: number, limit: number): { value: string; next: number } {
  const size = readU32(bytes, offset);
  const start = size.next;
  const end = start + size.value;
  if (end > limit) throw new Error(`truncated WebAssembly name at ${offset}`);
  return {
    value: new TextDecoder("utf-8").decode(bytes.subarray(start, end)),
    next: end
  };
}

function requireByte(bytes: Uint8Array, offset: number): number {
  const byte = bytes[offset];
  if (byte === undefined) throw new Error(`unexpected end of WebAssembly bytes at ${offset}`);
  return byte;
}

class Cursor {
  public offset = 0;

  public constructor(
    private readonly bytes: Uint8Array,
    private readonly absoluteStart: number
  ) {}

  public byte(): number {
    const out = requireByte(this.bytes, this.offset);
    this.offset += 1;
    return out;
  }

  public u32(): number {
    const out = readU32(this.bytes, this.offset);
    this.offset = out.next;
    return out.value;
  }

  public name(): string {
    const out = readName(this.bytes, this.offset, this.bytes.length);
    this.offset = out.next;
    return out.value;
  }

  public readBytes(count: number): Uint8Array {
    const end = this.offset + count;
    if (end > this.bytes.length) {
      throw new Error(`truncated WebAssembly bytes at ${this.absoluteStart + this.offset}`);
    }
    const out = this.bytes.slice(this.offset, end);
    this.offset = end;
    return out;
  }

  public done(): void {
    if (this.offset !== this.bytes.length) {
      throw new Error(`unconsumed WebAssembly section bytes at ${this.absoluteStart + this.offset}`);
    }
  }
}
