import { describe, expect, it } from 'vitest';
import { bytesToBase64, base64ToBytes, createCloudflareOpaqueClient, createOpaqueHttpClient } from './opaqueClient';

describe('opaqueClient', () => {
	it('round-trips binary messages through base64 transport encoding', () => {
		expect.hasAssertions();

		const bytes = new Uint8Array([0, 1, 2, 253, 254, 255]);
		expect(base64ToBytes(bytesToBase64(bytes))).toEqual(bytes);
	});

	it('posts OPAQUE protocol bytes without raw password fields', async () => {
		expect.hasAssertions();

		let body = '';
		const client = createOpaqueHttpClient(async (_input, init) => {
			body = String(init?.body ?? '');
			return new Response(JSON.stringify({ flowId: 'flow-1', message: 'AQID' }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			});
		});

		await expect(client.registrationStart('user-1', new Uint8Array([1, 2, 3]))).resolves.toMatchObject({
			flowId: 'flow-1'
		});
		expect(JSON.parse(body)).toEqual({ userId: 'user-1', message: 'AQID' });
		expect(body).not.toContain('password');
	});

	it('constructs the Cloudflare OPAQUE client used by the browser flow', () => {
		expect.hasAssertions();

		const client = createCloudflareOpaqueClient();

		expect(client).toHaveProperty('registerInit');
		expect(client).toHaveProperty('authInit');
	});
});
