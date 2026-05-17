import { page } from 'vitest/browser';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import type { ComponentProps } from 'svelte';
import WorkbenchShell from './WorkbenchShell.svelte';
import type { CheckResult, ModelDocument } from '$lib/types';

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
		mode: 'pdr',
		activeModel: model,
		editorText: model.text,
		editorSaveState: 'idle',
		editorKeymap: 'sublime',
		tutorialVisible: false,
		tutorialUrl: '/static/tutorial/kenmcmil.github.io/ivy/language.html',
		tutorialInput: '/static/tutorial/kenmcmil.github.io/ivy/language.html',
		tutorialCanGoBack: false,
		tutorialCanGoForward: false,
		tutorialFrameKey: 0,
		tutorialButtonFlashing: false,
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
		onSetMode: vi.fn(),
		onToggleTutorial: vi.fn(),
		onSetTutorialInput: vi.fn(),
		onNavigateTutorial: vi.fn(),
		onTutorialBack: vi.fn(),
		onTutorialForward: vi.fn(),
		onTutorialReload: vi.fn(),
		onCloseTutorial: vi.fn(),
		onRunCommand: vi.fn(),
		onUpdateEditor: vi.fn(),
		onSetEditorKeymap: vi.fn(),
		onSelectNode: vi.fn(),
		onNodeAction: vi.fn(),
		onToggleRelation: vi.fn(),
		...overrides
	});
}

const failingCheck: CheckResult = {
	id: 'check-1',
	jobId: 'job-1',
	sessionId: 'session-1',
	mode: 'induction',
	z3Contacted: true,
	result: 'fail',
	message: 'The following conjecture is not relatively inductive:',
	createdAt
};

describe('WorkbenchShell', () => {
	it('renders the old workbench pane structure with empty data states', async () => {
		expect.hasAssertions();

		renderShell();

		await expect.element(page.getByRole('main')).toBeInTheDocument();
		await expect.element(page.getByLabelText('ARG graph')).toBeVisible();
		await expect.element(page.getByLabelText('Concept graph', { exact: true })).toBeVisible();
		await expect.element(page.getByLabelText('State relations', { exact: true })).toBeVisible();
		await expect.element(page.getByLabelText('Editor', { exact: true })).toBeVisible();
		await expect.element(page.getByLabelText('Details and checks', { exact: true })).toBeVisible();
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

	it('exposes webui mode choices and dispatches selected-mode checks', async () => {
		expect.hasAssertions();
		const onSetMode = vi.fn();
		const onRunCommand = vi.fn();

		renderShell({ mode: 'pdr', onSetMode, onRunCommand });

		const mode = page.getByLabelText('Mode');
		await expect.element(mode).toHaveValue('pdr');
		await expect.element(page.getByRole('option', { name: 'PDR' })).toBeInTheDocument();
		await mode.selectOptions('bounded');
		expect(onSetMode).toHaveBeenCalledWith('bounded');

		await page.getByTestId('run-check').click();
		expect(onRunCommand).toHaveBeenCalledWith('runCheck');
	});

	it('renders draggable splitters for the webui workbench panes', async () => {
		expect.hasAssertions();

		renderShell();

		await expect.element(page.getByTestId('arg-concept-splitter')).toBeInTheDocument();
		await expect.element(page.getByTestId('concept-state-splitter')).toBeInTheDocument();
		await expect.element(page.getByTestId('state-editor-splitter')).toBeInTheDocument();
		await expect.element(page.getByTestId('details-splitter')).toBeInTheDocument();
	});

	it('colors verification failures as status errors', async () => {
		expect.hasAssertions();

		renderShell({ latestCheck: failingCheck, statusLevel: 'success' });

		await expect.element(page.getByTestId('status-strip')).toHaveClass(/error/);
		await expect.element(page.getByTestId('status-strip')).toHaveTextContent('Check FAIL (induction)');
	});

	it('renders the webui save sheen while the editor is saving', async () => {
		expect.hasAssertions();

		renderShell({ editorSaveState: 'saving' });

		await expect.element(page.getByTestId('editor-label')).toHaveTextContent('client_server_example.ivy [saving...]');
		await expect.element(page.getByTestId('save-editor-sheen')).toBeInTheDocument();
	});

	it('shows and wires the tutorial pane from the menubar toggle', async () => {
		expect.hasAssertions();
		const onToggleTutorial = vi.fn();
		const onCloseTutorial = vi.fn();

		renderShell({ tutorialVisible: true, onToggleTutorial, onCloseTutorial });

		await expect.element(page.getByTestId('toggle-tutorial')).toHaveTextContent('Hide Tutorial');
		await expect.element(page.getByLabelText('Tutorial', { exact: true })).toBeVisible();
		await expect.element(page.getByLabelText('Tutorial URL')).toHaveValue('/static/tutorial/kenmcmil.github.io/ivy/language.html');
		await page.getByTestId('toggle-tutorial').click();
		expect(onToggleTutorial).toHaveBeenCalled();
		await page.getByTitle('Close tutorial').click();
		expect(onCloseTutorial).toHaveBeenCalled();
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
