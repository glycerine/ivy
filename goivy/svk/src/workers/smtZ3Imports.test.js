/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck
import { describe, expect, it } from 'vitest';
import { createSmtZ3Imports } from './smtZ3Imports.js';

describe('smtZ3Imports string bridge', () => {
	it('routes context interrupts to the Z3 wasm export', () => {
		expect.hasAssertions();

		const calls = [];
		const z3 = {
			_Z3_interrupt(ctx) {
				calls.push(ctx);
			}
		};

		const imports = createSmtZ3Imports({ z3, getGoMemory: () => null });
		imports.Z3_interrupt(0xabc);

		expect(calls).toEqual([0xabc]);
	});

	it('copies Z3 C string bytes without UTF-8 decoding', () => {
		expect.hasAssertions();

		const heap = new Uint8Array(64);
		heap.set([0x41, 0xa2, 0xa3, 0x00], 8);
		let utf8ToStringCalled = false;
		const z3 = {
			HEAPU8: heap,
			UTF8ToString() {
				utf8ToStringCalled = true;
				throw new Error('UTF8ToString should not be used for Z3 C string results');
			},
			_Z3_get_error_msg() {
				return 8;
			}
		};

		const imports = createSmtZ3Imports({ z3, getGoMemory: () => null });
		const handle = imports.Z3_get_error_msg(1, 0);

		expect(imports.Z3_string_len(handle)).toBe(3);
		expect(imports.Z3_string_word(handle, 0)).toBe(0x00a3a241);
		expect(utf8ToStringCalled).toBe(false);
	});
});
