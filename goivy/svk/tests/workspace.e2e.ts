import { expect, test, type Locator, type Page } from '@playwright/test';

async function openWorkspace(page: Page) {
	await page.goto('/');
	await expect(page.getByRole('main')).toContainText('Editing: client_server_example.ivy');
	await expect(page.getByTestId('status-strip')).toContainText('ready');
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

	await expect(page.getByLabel('Editor')).toBeVisible();
	await expect(page.getByLabel('ARG graph')).toBeVisible();
	await expect(page.getByLabel('Concept graph')).toBeVisible();
	await expect(page.getByLabel('Details and checks')).toBeVisible();
});

test('editor accepts text and updates the dirty indicator', async ({ page }) => {
	await openWorkspace(page);

	const editor = page.getByTestId('model-editor');
	await editor.fill('type node\nrelation next(X:node,Y:node)');

	await expect(page.getByTestId('dirty-indicator')).toHaveText('Unsaved');
});

test('fake induction command creates a completed job row', async ({ page }) => {
	await openWorkspace(page);

	await page.getByTestId('run-induction').click();

	await expect(page.getByTestId('job-strip')).toContainText('check-induction');
	await expect(page.getByTestId('job-strip')).toContainText('succeeded');
	await expect(page.getByTestId('status-strip')).toContainText('Check PASS');
});

test('fake graph renders selectable nodes and updates details', async ({ page }) => {
	await openWorkspace(page);

	const firstNode = page.getByTestId('graph-node').first();
	await expect(firstNode).toBeVisible();
	await firstNode.click();

	await expect(page.getByTestId('details-pane')).toContainText('0');
	await expect(page.getByTestId('details-pane')).toContainText('fake.node');
});

test('graph action menu emits a command intent and remains usable after update', async ({ page }) => {
	await openWorkspace(page);

	await page.getByTestId('graph-node').first().click();
	await expect(page.getByTestId('graph-actions')).toContainText('Expand');
	await page.getByRole('button', { name: 'Expand' }).click();

	await expect(page.getByTestId('job-strip')).toContainText('arg-action');
	await expect(page.getByTestId('graph-node').first()).toBeVisible();
	await page.getByTestId('graph-node').first().click();
	await expect(page.getByTestId('details-pane')).toContainText('0');
});

test('major panes do not overlap on desktop or mobile', async ({ page }) => {
	await page.setViewportSize({ width: 1280, height: 820 });
	await openWorkspace(page);
	await expectNoOverlap(page.getByLabel('Editor'), page.getByLabel('ARG graph'));
	await expectNoOverlap(page.getByLabel('Editor'), page.getByLabel('Details and checks'));

	await page.setViewportSize({ width: 390, height: 820 });
	await openWorkspace(page);
	await expectNoOverlap(page.getByLabel('Editor'), page.getByLabel('ARG graph'));
	await expectNoOverlap(page.getByLabel('Concept graph'), page.getByLabel('Details and checks'));
});
