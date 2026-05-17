import { expect, test, type Locator, type Page } from '@playwright/test';

async function openWorkspace(page: Page) {
	await page.goto('/');
	await expect(page.getByRole('main')).toContainText('Editing: client_server_example.ivy');
	await page.getByLabel('Engine').selectOption('browser-wasm');
	await expect(page.getByTestId('status-strip')).toContainText('browser-wasm ready', { timeout: 30_000 });
}

async function expectNoOverlap(a: Locator, b: Locator) {
	const first = await a.boundingBox();
	const second = await b.boundingBox();
	expect(first).not.toBeNull();
	expect(second).not.toBeNull();
	if (!first || !second) {
		return;
	}
	const separated =
		first.x + first.width <= second.x ||
		second.x + second.width <= first.x ||
		first.y + first.height <= second.y ||
		second.y + second.height <= first.y;
	expect(separated).toBe(true);
}

test('first screen is the authenticated workspace shell', async ({ page }) => {
	await openWorkspace(page);

	await expect(page.getByLabel('Editor', { exact: true })).toBeVisible();
	await expect(page.getByLabel('ARG graph', { exact: true })).toBeVisible();
	await expect(page.getByLabel('Concept graph', { exact: true })).toBeVisible();
	await expect(page.getByLabel('Details and checks', { exact: true })).toBeVisible();
});

test('editor accepts text and updates the dirty indicator', async ({ page }) => {
	await openWorkspace(page);

	const editor = page.getByTestId('model-editor');
	await editor.fill('type node\nrelation next(X:node,Y:node)');

	await expect(page.getByTestId('dirty-indicator')).toHaveText('Unsaved');
});

test('browser wasm induction command reports a real solver result', async ({ page }) => {
	await openWorkspace(page);

	await page.getByLabel('Mode').selectOption('induction');
	await page.getByTestId('run-check').click();

	await expect(page.getByTestId('job-strip')).toContainText('check-induction');
	await expect(page.getByTestId('job-strip')).toContainText('succeeded');
	await expect(page.getByTestId('status-strip')).toContainText(/Check (PASS|FAIL|ERROR)/);
});

test('real engine result populates details with solver output', async ({ page }) => {
	await openWorkspace(page);
	await page.getByLabel('Mode').selectOption('induction');
	await page.getByTestId('run-check').click();

	await expect(page.getByTestId('details-pane')).toContainText('Verification Result');
});

test('engine selection remains usable after a real command', async ({ page }) => {
	await openWorkspace(page);
	await page.getByLabel('Mode').selectOption('induction');
	await page.getByTestId('run-check').click();
	await expect(page.getByTestId('job-strip')).toContainText('check-induction');

	await page.getByLabel('Engine').selectOption('browser-wasm');
	await expect(page.getByTestId('status-strip')).toContainText('browser-wasm ready', { timeout: 30_000 });
});

test('major panes do not overlap on desktop or mobile', async ({ page }) => {
	await page.setViewportSize({ width: 1280, height: 820 });
	await openWorkspace(page);
	await expectNoOverlap(page.getByLabel('Editor', { exact: true }), page.getByLabel('ARG graph', { exact: true }));
	await expectNoOverlap(page.getByLabel('Editor', { exact: true }), page.getByLabel('Details and checks', { exact: true }));

	await page.setViewportSize({ width: 390, height: 820 });
	await openWorkspace(page);
	await expectNoOverlap(page.getByLabel('Editor', { exact: true }), page.getByLabel('ARG graph', { exact: true }));
	await expectNoOverlap(page.getByLabel('Concept graph', { exact: true }), page.getByLabel('Details and checks', { exact: true }));
});
