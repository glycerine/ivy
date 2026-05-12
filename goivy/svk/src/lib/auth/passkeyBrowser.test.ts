import { describe, expect, it } from 'vitest';
import { base64UrlToBytes, bytesToBase64Url } from './passkeyBrowser';

describe('passkeyBrowser encoding helpers', () => {
	it('round-trips WebAuthn challenge bytes with base64url encoding', () => {
		expect.hasAssertions();

		const bytes = new Uint8Array([0, 1, 2, 251, 252, 253, 254, 255]);
		const encoded = bytesToBase64Url(bytes);

		expect(encoded).not.toContain('+');
		expect(encoded).not.toContain('/');
		expect(encoded).not.toContain('=');
		expect(base64UrlToBytes(encoded)).toEqual(bytes);
	});
});
