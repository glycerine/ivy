/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck
import { describe, expect, it } from 'vitest';
import { installGoIvyNodeFS } from './goivyNodeFS.js';

function callFs(method, ...args) {
	return new Promise((resolve, reject) => {
		globalThis.fs[method](...args, (error, value) => {
			if (error) {
				reject(error);
				return;
			}
			resolve(value);
		});
	});
}

describe('goivyNodeFS', () => {
	it('serves a read-only include tree to Go wasm hooks', async () => {
		expect.hasAssertions();

		const host = installGoIvyNodeFS({
			includeRoot: '/include',
			includeTree: {
				files: [{ path: 'stdlib/core.ivy', data: 'type node' }]
			}
		});

		const entries = await callFs('readdir', '/include/stdlib');
		const fd = await callFs('open', '/include/stdlib/core.ivy', 0, 0);
		const stat = await callFs('fstat', fd);
		const bytes = new Uint8Array(64);
		const read = await callFs('read', fd, bytes, 0, 64, 0);
		await callFs('close', fd);

		expect(host.includeRoot).toBe('/include');
		expect(entries).toEqual(['core.ivy']);
		expect(stat.isFile()).toBe(true);
		expect(host.decode(bytes.subarray(0, read))).toBe('type node');
	});

	it('rejects writes so the worker filesystem stays local-first and immutable', async () => {
		expect.hasAssertions();

		installGoIvyNodeFS();

		await expect(callFs('mkdir', '/include/new', 0o755)).rejects.toMatchObject({ code: 'EROFS' });
	});
});
