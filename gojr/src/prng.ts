import { blake3RawBytes } from "./blake3.js";

const seedEncoder = new TextEncoder();

export class DeterministicPrng {
  private state: bigint;

  public constructor(seed: number | string | bigint | undefined) {
    this.state = normalizeRandomSeed(seed);
  }

  public nextIndex(length: number): number {
    if (length <= 0) throw new RangeError("cannot sample from empty random set");
    this.state = (this.state * 6364136223846793005n + 1442695040888963407n) & 0xffffffffffffffffn;
    return Number(this.state % BigInt(length));
  }
}

export function normalizeRandomSeed(seed: number | string | bigint | undefined): bigint {
  if (seed === undefined) return 0x6a09e667f3bcc909n;
  if (typeof seed === "bigint") return seed & 0xffffffffffffffffn;
  if (typeof seed === "number") return BigInt(Math.trunc(seed)) & 0xffffffffffffffffn;
  const digest = blake3RawBytes(seedEncoder.encode(seed), 8);
  let value = 0n;
  for (let index = 0; index < digest.length; index += 1) {
    value |= BigInt(digest[index]!) << BigInt(index * 8);
  }
  return value;
}
