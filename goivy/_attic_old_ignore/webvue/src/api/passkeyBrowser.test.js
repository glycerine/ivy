import { describe, expect, it, vi } from 'vitest';
import {
  base64URLToBuffer,
  bufferToBase64URL,
  createPasskeyBrowser,
  decodePublicKeyCreationOptions,
  decodePublicKeyRequestOptions,
} from './passkeyBrowser.js';

describe('passkeyBrowser', () => {
  it('round-trips base64url encoded ArrayBuffers', () => {
    const encoded = bufferToBase64URL(new Uint8Array([1, 2, 3, 254, 255]).buffer);

    expect(encoded).toBe('AQID_v8');
    expect(Array.from(new Uint8Array(base64URLToBuffer(encoded)))).toEqual([1, 2, 3, 254, 255]);
  });

  it('decodes passkey registration options for navigator.credentials.create', () => {
    const decoded = decodePublicKeyCreationOptions({
      challenge: 'AQID',
      user: { id: 'BAUG', name: 'alice@example.test', displayName: 'Alice' },
      excludeCredentials: [{ type: 'public-key', id: 'BwgJ' }],
    });

    expect(Array.from(new Uint8Array(decoded.challenge))).toEqual([1, 2, 3]);
    expect(Array.from(new Uint8Array(decoded.user.id))).toEqual([4, 5, 6]);
    expect(Array.from(new Uint8Array(decoded.excludeCredentials[0].id))).toEqual([7, 8, 9]);
  });

  it('decodes passkey login options for navigator.credentials.get', () => {
    const decoded = decodePublicKeyRequestOptions({
      challenge: 'AQID',
      allowCredentials: [{ type: 'public-key', id: 'BAUG' }],
    });

    expect(Array.from(new Uint8Array(decoded.challenge))).toEqual([1, 2, 3]);
    expect(Array.from(new Uint8Array(decoded.allowCredentials[0].id))).toEqual([4, 5, 6]);
  });

  it('serializes a browser passkey assertion after automatic startup authentication', async () => {
    globalThis.PublicKeyCredential = function PublicKeyCredential() {};
    const get = vi.fn(async () => ({
      id: 'credential-1',
      rawId: new Uint8Array([1]).buffer,
      type: 'public-key',
      response: {
        clientDataJSON: new Uint8Array([2]).buffer,
        authenticatorData: new Uint8Array([3]).buffer,
        signature: new Uint8Array([4]).buffer,
        userHandle: new Uint8Array([5]).buffer,
      },
    }));
    const browser = createPasskeyBrowser({ credentials: { get } });

    const credential = await browser.authenticate({ publicKey: { challenge: 'Bg' } });

    expect(get).toHaveBeenCalledWith({
      publicKey: {
        challenge: expect.any(ArrayBuffer),
        allowCredentials: [],
      },
    });
    expect(credential).toEqual({
      id: 'credential-1',
      rawId: 'AQ',
      type: 'public-key',
      response: {
        clientDataJSON: 'Ag',
        authenticatorData: 'Aw',
        signature: 'BA',
        userHandle: 'BQ',
      },
    });
    delete globalThis.PublicKeyCredential;
  });
});
