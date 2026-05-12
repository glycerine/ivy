import { page } from 'vitest/browser';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DialogHost from './DialogHost.svelte';
import ContextMenuHost from './ContextMenuHost.svelte';
import ToastHost from './ToastHost.svelte';

describe('overlay hosts', () => {
	it('renders confirm dialogs and resolves button choices', async () => {
		expect.hasAssertions();
		const onResolve = vi.fn();
		render(DialogHost, {
			dialog: { id: 'd1', kind: 'confirm', title: 'Close file', message: 'Discard changes?' },
			onResolve
		});

		await expect.element(page.getByRole('dialog')).toBeVisible();
		await page.getByRole('button', { name: 'Cancel' }).click();
		expect(onResolve).toHaveBeenCalledWith(false);
	});

	it('renders context menu items and toasts', async () => {
		expect.hasAssertions();
		const onSelect = vi.fn();
		render(ContextMenuHost, { visible: true, x: 10, y: 20, items: [{ id: 'expand', label: 'Expand' }], onSelect });
		await page.getByRole('button', { name: 'Expand' }).click();
		expect(onSelect).toHaveBeenCalledWith({ id: 'expand', label: 'Expand' });

		const onDismiss = vi.fn();
		render(ToastHost, { toasts: [{ id: 't1', level: 'success', message: 'Saved', createdAt: 0 }], onDismiss });
		await page.getByRole('button', { name: 'Dismiss Saved' }).click();
		expect(onDismiss).toHaveBeenCalledWith('t1');
	});
});
