import { describe, expect, it, vi } from 'vitest';
import {
	analysisStateFilename,
	downloadTextFile,
	mergeDiskVersionIntoEditBuffer,
	modelDownloadFilename
} from './fileLifecycle';

describe('file lifecycle helpers', () => {
	it('merges external disk edits using conflict markers when both sides changed', () => {
		expect.hasAssertions();
		const merged = mergeDiskVersionIntoEditBuffer('type t\n', 'type local\n', 'type disk\n');

		expect(merged).toContain('<<<<<<< EDIT BUFFER');
		expect(merged).toContain('type local');
		expect(merged).toContain('||||||| LAST SAVED');
		expect(merged).toContain('type t');
		expect(merged).toContain('>>>>>>> ON DISK');
	});

	it('takes the changed side directly when only one side changed', () => {
		expect.hasAssertions();
		expect(mergeDiskVersionIntoEditBuffer('base', 'base', 'disk')).toBe('disk');
		expect(mergeDiskVersionIntoEditBuffer('base', 'editor', 'base')).toBe('editor');
	});

	it('derives download filenames for models and analysis state', () => {
		expect.hasAssertions();
		expect(modelDownloadFilename('')).toBe('model.ivy');
		expect(modelDownloadFilename('client.ivy')).toBe('client.ivy');
		expect(analysisStateFilename('client.ivy')).toBe('client.ivyweb.json');
	});

	it('downloads text through a temporary anchor', () => {
		expect.hasAssertions();
		const clicks: string[] = [];
		const anchor = {
			href: '',
			download: '',
			click: vi.fn(() => clicks.push('clicked'))
		};
		const doc = {
			body: {
				appendChild: vi.fn(),
				removeChild: vi.fn()
			},
			createElement: vi.fn(() => anchor)
		} as unknown as Document;
		const createObjectURL = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:test');
		const revokeObjectURL = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined);

		downloadTextFile('model.ivy', 'type t', 'text/plain', doc);

		expect(anchor.download).toBe('model.ivy');
		expect(anchor.href).toBe('blob:test');
		expect(clicks).toEqual(['clicked']);
		expect(doc.body.appendChild).toHaveBeenCalledWith(anchor);
		expect(doc.body.removeChild).toHaveBeenCalledWith(anchor);
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:test');

		createObjectURL.mockRestore();
		revokeObjectURL.mockRestore();
	});
});
