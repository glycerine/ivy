const FLAG_CHUNK_START = 1 << 0;
const FLAG_CHUNK_END = 1 << 1;
const FLAG_PARENT = 1 << 2;
const FLAG_ROOT = 1 << 3;
const BLOCK_SIZE = 64;
const CHUNK_SIZE = 1024;

const IV = [
  0x6A09E667, 0xBB67AE85, 0x3C6EF372, 0xA54FF53A,
  0x510E527F, 0x9B05688C, 0x1F83D9AB, 0x5BE0CD19
] as const;

interface Node {
  cv: number[];
  block: number[];
  counter: bigint;
  blockLen: number;
  flags: number;
}

const encoder = new TextEncoder();

export function blake3HashString(text: string): string {
  return blake3HashBytes(encoder.encode(text));
}

export function blake3HashBytes(bytes: Uint8Array): string {
  return `blake3.33B-${base64Url(blake3RawBytes(bytes, 33))}`;
}

export function blake3RawBytes(bytes: Uint8Array, length = 64): Uint8Array {
  const out = new Uint8Array(length);
  let written = 0;
  let counter = 0n;
  const root = rootNode(bytes);
  while (written < length) {
    const block = wordsToBytes(compressNode({ ...root, counter }));
    const n = Math.min(block.length, length - written);
    out.set(block.subarray(0, n), written);
    written += n;
    counter++;
  }
  return out;
}

function rootNode(bytes: Uint8Array): Node {
  if (bytes.length <= BLOCK_SIZE) {
    const block = new Uint8Array(BLOCK_SIZE);
    block.set(bytes);
    return {
      cv: [...IV],
      block: bytesToWords(block),
      counter: 0n,
      blockLen: bytes.length,
      flags: FLAG_CHUNK_START | FLAG_CHUNK_END | FLAG_ROOT
    };
  }
  if (bytes.length <= CHUNK_SIZE) {
    const node = compressChunk(bytes, IV, 0n, 0);
    node.flags |= FLAG_ROOT;
    return node;
  }

  const cvs: number[][] = [];
  for (let offset = 0, chunkCounter = 0n; offset < bytes.length; offset += CHUNK_SIZE, chunkCounter++) {
    const chunk = bytes.subarray(offset, Math.min(offset + CHUNK_SIZE, bytes.length));
    cvs.push(chainingValue(compressChunk(chunk, IV, chunkCounter, 0)));
  }
  const node = rootParentNode(cvs, 0, cvs.length);
  node.flags |= FLAG_ROOT;
  return node;
}

function rootParentNode(cvs: number[][], start: number, end: number): Node {
  const count = end - start;
  if (count < 2) throw new Error("internal blake3 error: root parent needs at least two children");
  const split = start + largestPowerOfTwoLessThan(count);
  return parentNode(subtreeCV(cvs, start, split), subtreeCV(cvs, split, end), IV, 0);
}

function subtreeCV(cvs: number[][], start: number, end: number): number[] {
  const count = end - start;
  if (count === 1) return cvs[start]!.slice();
  return chainingValue(rootParentNode(cvs, start, end));
}

function largestPowerOfTwoLessThan(n: number): number {
  let power = 1;
  while (power * 2 < n) power *= 2;
  return power;
}

function compressChunk(chunk: Uint8Array, key: readonly number[], counter: bigint, flags: number): Node {
  let node: Node = {
    cv: [...key],
    block: new Array(16).fill(0),
    counter,
    blockLen: BLOCK_SIZE,
    flags: flags | FLAG_CHUNK_START
  };
  let offset = 0;
  while (chunk.length - offset > BLOCK_SIZE) {
    node.block = bytesToWords(chunk.subarray(offset, offset + BLOCK_SIZE));
    node.cv = chainingValue(node);
    node.flags &= ~FLAG_CHUNK_START;
    offset += BLOCK_SIZE;
  }
  const last = new Uint8Array(BLOCK_SIZE);
  const tail = chunk.subarray(offset);
  last.set(tail);
  node = {
    ...node,
    block: bytesToWords(last),
    blockLen: tail.length,
    flags: node.flags | FLAG_CHUNK_END
  };
  return node;
}

function parentNode(left: readonly number[], right: readonly number[], key: readonly number[], flags: number): Node {
  return {
    cv: [...key],
    block: [...left.slice(0, 8), ...right.slice(0, 8)],
    counter: 0n,
    blockLen: BLOCK_SIZE,
    flags: flags | FLAG_PARENT
  };
}

function chainingValue(node: Node): number[] {
  return compressNode(node).slice(0, 8);
}

function compressNode(node: Node): number[] {
  const block = node.block;
  const cv = node.cv;
  let s0: number, s1: number, s2: number, s3: number;
  let s4: number, s5: number, s6: number, s7: number;
  let s8: number, s9: number, s10: number, s11: number;
  let s12: number, s13: number, s14: number, s15: number;
  const counterLo = Number(node.counter & 0xffffffffn) >>> 0;
  const counterHi = Number((node.counter >> 32n) & 0xffffffffn) >>> 0;

  [s0, s4, s8, s12] = g(cv[0]!, cv[4]!, IV[0], counterLo, block[0]!, block[1]!);
  [s1, s5, s9, s13] = g(cv[1]!, cv[5]!, IV[1], counterHi, block[2]!, block[3]!);
  [s2, s6, s10, s14] = g(cv[2]!, cv[6]!, IV[2], node.blockLen, block[4]!, block[5]!);
  [s3, s7, s11, s15] = g(cv[3]!, cv[7]!, IV[3], node.flags, block[6]!, block[7]!);
  [s0, s5, s10, s15] = g(s0, s5, s10, s15, block[8]!, block[9]!);
  [s1, s6, s11, s12] = g(s1, s6, s11, s12, block[10]!, block[11]!);
  [s2, s7, s8, s13] = g(s2, s7, s8, s13, block[12]!, block[13]!);
  [s3, s4, s9, s14] = g(s3, s4, s9, s14, block[14]!, block[15]!);

  [s0, s4, s8, s12] = g(s0, s4, s8, s12, block[2]!, block[6]!);
  [s1, s5, s9, s13] = g(s1, s5, s9, s13, block[3]!, block[10]!);
  [s2, s6, s10, s14] = g(s2, s6, s10, s14, block[7]!, block[0]!);
  [s3, s7, s11, s15] = g(s3, s7, s11, s15, block[4]!, block[13]!);
  [s0, s5, s10, s15] = g(s0, s5, s10, s15, block[1]!, block[11]!);
  [s1, s6, s11, s12] = g(s1, s6, s11, s12, block[12]!, block[5]!);
  [s2, s7, s8, s13] = g(s2, s7, s8, s13, block[9]!, block[14]!);
  [s3, s4, s9, s14] = g(s3, s4, s9, s14, block[15]!, block[8]!);

  [s0, s4, s8, s12] = g(s0, s4, s8, s12, block[3]!, block[4]!);
  [s1, s5, s9, s13] = g(s1, s5, s9, s13, block[10]!, block[12]!);
  [s2, s6, s10, s14] = g(s2, s6, s10, s14, block[13]!, block[2]!);
  [s3, s7, s11, s15] = g(s3, s7, s11, s15, block[7]!, block[14]!);
  [s0, s5, s10, s15] = g(s0, s5, s10, s15, block[6]!, block[5]!);
  [s1, s6, s11, s12] = g(s1, s6, s11, s12, block[9]!, block[0]!);
  [s2, s7, s8, s13] = g(s2, s7, s8, s13, block[11]!, block[15]!);
  [s3, s4, s9, s14] = g(s3, s4, s9, s14, block[8]!, block[1]!);

  [s0, s4, s8, s12] = g(s0, s4, s8, s12, block[10]!, block[7]!);
  [s1, s5, s9, s13] = g(s1, s5, s9, s13, block[12]!, block[9]!);
  [s2, s6, s10, s14] = g(s2, s6, s10, s14, block[14]!, block[3]!);
  [s3, s7, s11, s15] = g(s3, s7, s11, s15, block[13]!, block[15]!);
  [s0, s5, s10, s15] = g(s0, s5, s10, s15, block[4]!, block[0]!);
  [s1, s6, s11, s12] = g(s1, s6, s11, s12, block[11]!, block[2]!);
  [s2, s7, s8, s13] = g(s2, s7, s8, s13, block[5]!, block[8]!);
  [s3, s4, s9, s14] = g(s3, s4, s9, s14, block[1]!, block[6]!);

  [s0, s4, s8, s12] = g(s0, s4, s8, s12, block[12]!, block[13]!);
  [s1, s5, s9, s13] = g(s1, s5, s9, s13, block[9]!, block[11]!);
  [s2, s6, s10, s14] = g(s2, s6, s10, s14, block[15]!, block[10]!);
  [s3, s7, s11, s15] = g(s3, s7, s11, s15, block[14]!, block[8]!);
  [s0, s5, s10, s15] = g(s0, s5, s10, s15, block[7]!, block[2]!);
  [s1, s6, s11, s12] = g(s1, s6, s11, s12, block[5]!, block[3]!);
  [s2, s7, s8, s13] = g(s2, s7, s8, s13, block[0]!, block[1]!);
  [s3, s4, s9, s14] = g(s3, s4, s9, s14, block[6]!, block[4]!);

  [s0, s4, s8, s12] = g(s0, s4, s8, s12, block[9]!, block[14]!);
  [s1, s5, s9, s13] = g(s1, s5, s9, s13, block[11]!, block[5]!);
  [s2, s6, s10, s14] = g(s2, s6, s10, s14, block[8]!, block[12]!);
  [s3, s7, s11, s15] = g(s3, s7, s11, s15, block[15]!, block[1]!);
  [s0, s5, s10, s15] = g(s0, s5, s10, s15, block[13]!, block[3]!);
  [s1, s6, s11, s12] = g(s1, s6, s11, s12, block[0]!, block[10]!);
  [s2, s7, s8, s13] = g(s2, s7, s8, s13, block[2]!, block[6]!);
  [s3, s4, s9, s14] = g(s3, s4, s9, s14, block[4]!, block[7]!);

  [s0, s4, s8, s12] = g(s0, s4, s8, s12, block[11]!, block[15]!);
  [s1, s5, s9, s13] = g(s1, s5, s9, s13, block[5]!, block[0]!);
  [s2, s6, s10, s14] = g(s2, s6, s10, s14, block[1]!, block[9]!);
  [s3, s7, s11, s15] = g(s3, s7, s11, s15, block[8]!, block[6]!);
  [s0, s5, s10, s15] = g(s0, s5, s10, s15, block[14]!, block[10]!);
  [s1, s6, s11, s12] = g(s1, s6, s11, s12, block[2]!, block[12]!);
  [s2, s7, s8, s13] = g(s2, s7, s8, s13, block[3]!, block[4]!);
  [s3, s4, s9, s14] = g(s3, s4, s9, s14, block[7]!, block[13]!);

  return [
    (s0 ^ s8) >>> 0, (s1 ^ s9) >>> 0, (s2 ^ s10) >>> 0, (s3 ^ s11) >>> 0,
    (s4 ^ s12) >>> 0, (s5 ^ s13) >>> 0, (s6 ^ s14) >>> 0, (s7 ^ s15) >>> 0,
    (s8 ^ cv[0]!) >>> 0, (s9 ^ cv[1]!) >>> 0, (s10 ^ cv[2]!) >>> 0, (s11 ^ cv[3]!) >>> 0,
    (s12 ^ cv[4]!) >>> 0, (s13 ^ cv[5]!) >>> 0, (s14 ^ cv[6]!) >>> 0, (s15 ^ cv[7]!) >>> 0
  ];
}

function g(a: number, b: number, c: number, d: number, mx: number, my: number): [number, number, number, number] {
  a = (a + b + mx) >>> 0;
  d = rotateRight(d ^ a, 16);
  c = (c + d) >>> 0;
  b = rotateRight(b ^ c, 12);
  a = (a + b + my) >>> 0;
  d = rotateRight(d ^ a, 8);
  c = (c + d) >>> 0;
  b = rotateRight(b ^ c, 7);
  return [a, b, c, d];
}

function rotateRight(value: number, amount: number): number {
  return ((value >>> amount) | (value << (32 - amount))) >>> 0;
}

function bytesToWords(bytes: Uint8Array): number[] {
  const words = new Array(16).fill(0);
  for (let index = 0; index < 16; index++) {
    const offset = index * 4;
    words[index] = (
      (bytes[offset] ?? 0) |
      ((bytes[offset + 1] ?? 0) << 8) |
      ((bytes[offset + 2] ?? 0) << 16) |
      ((bytes[offset + 3] ?? 0) << 24)
    ) >>> 0;
  }
  return words;
}

function wordsToBytes(words: readonly number[]): Uint8Array {
  const out = new Uint8Array(64);
  for (let index = 0; index < 16; index++) {
    const word = words[index] ?? 0;
    const offset = index * 4;
    out[offset] = word & 0xff;
    out[offset + 1] = (word >>> 8) & 0xff;
    out[offset + 2] = (word >>> 16) & 0xff;
    out[offset + 3] = (word >>> 24) & 0xff;
  }
  return out;
}

function base64Url(bytes: Uint8Array): string {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";
  let out = "";
  for (let index = 0; index < bytes.length; index += 3) {
    const a = bytes[index]!;
    const b = bytes[index + 1];
    const c = bytes[index + 2];
    out += alphabet[a >>> 2];
    out += alphabet[((a & 0x03) << 4) | ((b ?? 0) >>> 4)];
    out += b === undefined ? "=" : alphabet[((b & 0x0f) << 2) | ((c ?? 0) >>> 6)];
    out += c === undefined ? "=" : alphabet[c & 0x3f];
  }
  return out;
}
