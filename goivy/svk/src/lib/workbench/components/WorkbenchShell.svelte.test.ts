import { page } from 'vitest/browser';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import type { ComponentProps } from 'svelte';
import WorkbenchShell from './WorkbenchShell.svelte';
import type { ModelDocument } from '$lib/types';

const createdAt = '2026-05-12T00:00:00.000Z';

const model: ModelDocument = {
	id: 'model-1',
	projectId: 'project-1',
	filename: 'client_server_example.ivy',
	text: 'type node',
	dirty: false,
	parseRevision: 1,
	engineRevision: 1,
	createdAt,
	updatedAt: createdAt
};

type ShellProps = ComponentProps<typeof WorkbenchShell>;

function renderShell(overrides: Partial<ShellProps> = {}) {
	return render(WorkbenchShell, {
		engineChoice: 'hosted-go',
		activeModel: model,
		editorText: model.text,
		editorKeymap: 'sublime',
		hasActiveSession: true,
		sessionLabel: 'session-1',
		statusMessage: 'Ready',
		statusLevel: '',
		latestGraph: null,
		latestConcept: null,
		latestCheck: null,
		selectedNode: null,
		selectedNodeId: null,
		jobs: [],
		stateRelationRows: [],
		onActivateEngine: vi.fn(),
		onRunCommand: vi.fn(),
		onReloadModel: vi.fn(),
		onMarkSaved: vi.fn(),
		onUpdateEditor: vi.fn(),
		onSetEditorKeymap: vi.fn(),
		onSelectNode: vi.fn(),
		onNodeAction: vi.fn(),
		onToggleRelation: vi.fn(),
		...overrides
	});
}

describe('WorkbenchShell', () => {
	it('renders the old workbench pane structure with empty data states', async () => {
		expect.hasAssertions();

		renderShell();

		await expect.element(page.getByRole('main')).toBeInTheDocument();
		await expect.element(page.getByLabelText('ARG graph')).toBeVisible();
		await expect.element(page.getByLabelText('Concept graph')).toBeVisible();
		await expect.element(page.getByLabelText('State relations')).toBeVisible();
		await expect.element(page.getByLabelText('Editor')).toBeVisible();
		await expect.element(page.getByLabelText('Details and checks')).toBeVisible();
		await expect.element(page.getByText('No graph loaded').first()).toBeInTheDocument();
		await expect.element(page.getByText('No relations loaded')).toBeInTheDocument();
		await expect.element(page.getByText('No verification result yet.')).toBeInTheDocument();
		await expect.element(page.getByTestId('status-strip')).toHaveTextContent('Ready');
	});

	it('renders the restored static workbench menu labels', async () => {
		expect.hasAssertions();

		renderShell();

		await expect.element(page.getByText('Open Event Trace')).toBeInTheDocument();
		await expect.element(page.getByText('Save Analysis State')).toBeInTheDocument();
		await expect.element(page.getByText('Check induction')).toBeInTheDocument();
		await expect.element(page.getByText('Bounded check')).toBeInTheDocument();
		await expect.element(page.getByText('PDR step')).toBeInTheDocument();
		await expect.element(page.getByText('Add relation')).toBeInTheDocument();
	});

	it('routes editor edits through the shell callback', async () => {
		expect.hasAssertions();
		const onUpdateEditor = vi.fn();

		renderShell({ onUpdateEditor });

		const editor = page.getByTestId('model-editor');
		await editor.fill('type other');

		expect(onUpdateEditor).toHaveBeenCalledWith('type other');
	});

	it('routes editor keymap changes through the shell callback', async () => {
		expect.hasAssertions();
		const onSetEditorKeymap = vi.fn();

		renderShell({ onSetEditorKeymap });

		await page.getByLabelText('Vim').click();

		expect(onSetEditorKeymap).toHaveBeenCalledWith('vim');
	});
});
