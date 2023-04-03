import { keccak256 } from "js-sha3";

import { JsonLdType } from "src/domain";

const HEX_TABLE = "0123456789abcdef";

// Helpers

function decode(src: string): Uint8Array {
  let i = 0;
  let j = 1;
  const dst: number[] = [];

  for (; j < src.length; j += 2) {
    const a = fromHexChar(src[j - 1].charCodeAt(0));
    const b = fromHexChar(src[j].charCodeAt(0));

    dst[i] = (a << 4) | b;
    i++;
  }

  if (src.length % 2 == 1) {
    throw new Error("Invalid hex string");
  }

  return Uint8Array.from(dst);
}

function encode(src: Uint8Array): Uint8Array {
  const dst = new Uint8Array(encodeLength(src.length));
  let j = 0;

  for (let i = 0; i < src.length; i++) {
    dst[j] = HEX_TABLE[src[i] >> 4].charCodeAt(0);
    dst[j + 1] = HEX_TABLE[src[i] & 0x0f].charCodeAt(0);
    j += 2;
  }

  return dst;
}

function encodeLength(n: number): number {
  return n * 2;
}

function encodeString(b: Uint8Array): string {
  return new TextDecoder().decode(encode(b));
}

function fromHexChar(c: number): number {
  if ("0".charCodeAt(0) <= c && c <= "9".charCodeAt(0)) {
    return c - "0".charCodeAt(0);
  } else if ("a".charCodeAt(0) <= c && c <= "f".charCodeAt(0)) {
    return c - "a".charCodeAt(0) + 10;
  }

  if ("A".charCodeAt(0) <= c && c <= "F".charCodeAt(0)) {
    return c - "A".charCodeAt(0) + 10;
  }

  throw new Error(`Invalid byte char ${c}`);
}

function fromLittleEndian(bytes: Uint8Array): bigint {
  const n256 = BigInt(256);
  let result = BigInt(0);
  let base = BigInt(1);

  bytes.forEach((byte) => {
    result += base * BigInt(byte);
    base = base * n256;
  });

  return result;
}

function schemaIdToBytes(schemaId: Uint8Array): Uint8Array {
  const sHash = decode(keccak256(schemaId));

  return sHash.slice(sHash.length - 16, sHash.length);
}

// Exports

export function getBigint(
  jsonLdType: JsonLdType
): { data: string; success: true } | { success: false } {
  try {
    const data = fromLittleEndian(
      schemaIdToBytes(new TextEncoder().encode(jsonLdType.id))
    ).toString();
    return {
      data,
      success: true,
    };
  } catch (e) {
    return {
      success: false,
    };
  }
}

export function getSchemaHash(
  jsonLdType: JsonLdType
): { data: string; success: true } | { success: false } {
  try {
    const data = encodeString(schemaIdToBytes(new TextEncoder().encode(jsonLdType.id)));
    return {
      data,
      success: true,
    };
  } catch (e) {
    return {
      success: false,
    };
  }
}
