import { expect, test } from '@playwright/test';

test('workbench splitters resize the webui parity panes', async ({ page }) => {
	await page.goto('/');
	await expect(page.getByLabel('Editor', { exact: true })).toBeVisible();

	const argPane = page.getByLabel('ARG graph', { exact: true });
	const conceptPane = page.getByLabel('Concept graph', { exact: true });
	const statePane = page.getByLabel('State relations', { exact: true });
	const detailsPane = page.getByLabel('Details and checks');

	const argBefore = await boundingBox(argPane);
	await dragSplitter(page, 'arg-concept-splitter', 90, 0);
	const argAfter = await boundingBox(argPane);
	expect(argAfter.width).toBeGreaterThan(argBefore.width + 20);

	const conceptBefore = await boundingBox(conceptPane);
	await dragSplitter(page, 'concept-state-splitter', 90, 0);
	const conceptAfter = await boundingBox(conceptPane);
	expect(conceptAfter.width).toBeGreaterThan(conceptBefore.width + 20);

	const stateBefore = await boundingBox(statePane);
	await dragSplitter(page, 'state-editor-splitter', 90, 0);
	const stateAfter = await boundingBox(statePane);
	expect(stateAfter.width).toBeGreaterThan(stateBefore.width + 20);

	const detailsBefore = await boundingBox(detailsPane);
	await dragSplitter(page, 'details-splitter', 0, -70);
	const detailsAfter = await boundingBox(detailsPane);
	expect(detailsAfter.height).toBeGreaterThan(detailsBefore.height + 20);
});

test('mode Check uses PDR and Ctrl-S saves the editor', async ({ page }) => {
	await page.goto('/');
	await expect(page.getByLabel('Editor', { exact: true })).toBeVisible();

	await page.getByLabel('Mode').selectOption('pdr');
	await expect(page.getByLabel('Mode')).toHaveValue('pdr');

	const editor = page.getByTestId('model-editor');
	await editor.fill('type node\nrelation link(X:node,Y:node)\n');
	await expect(page.getByTestId('dirty-indicator')).toHaveText('Unsaved');
	await page.keyboard.press(process.platform === 'darwin' ? 'Meta+S' : 'Control+S');
	await expect(page.getByTestId('dirty-indicator')).toHaveText('Saved');
	await expect(page.getByTestId('status-strip')).toContainText('Saved: client_server_example.ivy');

	await page.getByLabel('Engine').selectOption('browser-wasm');
	await expect(page.getByTestId('status-strip')).toContainText('browser-wasm ready', { timeout: 30_000 });
	await page.getByTestId('run-check').click();
	await expect(page.getByTestId('job-strip')).toContainText('check-pdr');
});

test('tutorial button opens the webui tutorial pane', async ({ page }) => {
	await page.goto('/');
	await page.getByTestId('toggle-tutorial').click();

	await expect(page.getByLabel('Tutorial', { exact: true })).toBeVisible();
	await expect(page.getByLabel('Tutorial URL')).toHaveValue('/static/tutorial/kenmcmil.github.io/ivy/language.html');
	await expect(page.getByTestId('toggle-tutorial')).toHaveText('Hide Tutorial');

	await page.getByTitle('Close tutorial').click();
	await expect(page.getByLabel('Tutorial', { exact: true })).toHaveCount(0);
	await expect(page.getByTestId('toggle-tutorial')).toHaveText('Show Tutorial');
});

async function dragSplitter(page: import('@playwright/test').Page, testId: string, dx: number, dy: number) {
	const box = await boundingBox(page.getByTestId(testId));
	await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
	await page.mouse.down();
	await page.mouse.move(box.x + box.width / 2 + dx, box.y + box.height / 2 + dy, { steps: 6 });
	await page.mouse.up();
}

async function boundingBox(locator: import('@playwright/test').Locator) {
	const box = await locator.boundingBox();
	expect(box).not.toBeNull();
	return box!;
}
