<script lang="ts">
	import { onMount } from 'svelte';
	import { BrowserWasmEngine, HostedGoEngine, HostedWebuiEngine } from '$lib/engines';
	import { createEngineService } from '$lib/services';
	import { nowIso } from '$lib/time';
	import WorkbenchShell from '$lib/workbench/components/WorkbenchShell.svelte';
	import { routeWorkbenchCommand } from '$lib/workbench/commands/commandRouting';
	import { createEditorUiState, createSessionUiState, type EditorKeymap, type SessionMode } from '$lib/workbench/state';
	import { downloadTextFile, modelDownloadFilename } from '$lib/workbench/services/fileLifecycle';
	import '$lib/workbench/workbench.css';
	import {
		createConceptsState,
		createChecksState,
		createEnginesState,
		createGraphsState,
		createJobsState,
		createStateRelationsState,
		createModelsState,
		createWorkspaceState
	} from '$lib/state';
	import type { EdgeDisplayClass, GraphNode, GraphSnapshot, ModelDocument, NodeAction, NodeLabelDisplayClass, Project } from '$lib/types';
	import type { IvyEngine } from '$lib/types';

	type EngineChoice = 'hosted-go' | 'hosted-webui' | 'browser-wasm';

	const createdAt = '2026-05-12T00:00:00.000Z';

	const project: Project = {
		id: 'project-local-1',
		accountId: 'account-local-1',
		ownerKind: 'user',
		ownerId: 'user-local-1',
		name: 'Local Ivy Workspace',
		slug: 'local-ivy-workspace',
		storageMode: 'local_indexeddb',
		createdAt,
		updatedAt: createdAt
	};

	const initialModel: ModelDocument = {
		id: 'model-local-1',
		projectId: project.id,
		filename: 'client_server_example.ivy',
		text: `#lang ivy1.7

type client
type server

relation link(X:client, Y:server)
relation semaphore(X:server)

after init {
    semaphore(W) := true;
    link(X,Y) := false
}

action connect(x:client,y:server) = {
    require semaphore(y);
    link(x,y) := true;
    semaphore(y) := false
}

action disconnect(x:client,y:server) = {
    require link(x,y);
    link(x,y) := false;
    semaphore(y) := true
}

invariant ~(X ~= Z & link(X,Y) & link(Z,Y))

export connect
export disconnect

# Add this to prove the invariant

#private {
#    invariant ~(link(X,Y) & semaphore(Y))
#}`,
		dirty: false,
		parseRevision: 1,
		engineRevision: 1,
		createdAt,
		updatedAt: createdAt
	};

	const stores = {
		workspace: createWorkspaceState({ hydrated: true, activeProjectId: project.id, activeModelId: initialModel.id }),
		engines: createEnginesState(),
		jobs: createJobsState(),
		graphs: createGraphsState(),
		concepts: createConceptsState(),
		checks: createChecksState(),
		stateRelations: createStateRelationsState()
	};
	const models = createModelsState();
	models.upsert(initialModel);
	const sessionUi = createSessionUiState();
	const editorUi = createEditorUiState({ path: initialModel.filename, content: initialModel.text });

	let editorText = $state(initialModel.text);
	let selectedNodeId = $state<string | null>(null);
	let selectedGraphId = $state<string | null>(null);
	let engineChoice = $state<EngineChoice>('hosted-go');
	let service = createEngineService({ engine: createEngine('hosted-go'), stores });
	let activationSerial = 0;
	let activationFailureMessage = '';

	const activeModel = $derived(models.table.byId[initialModel.id]);
	const jobs = $derived(stores.jobs.table.order.map((id) => stores.jobs.table.byId[id]));
	const latestGraph = $derived.by(() => {
		const graphId = selectedGraphId ?? stores.graphs.table.order.at(-1);
		return graphId ? stores.graphs.table.byId[graphId] : null;
	});
	const selectedNode = $derived(selectedNodeId && latestGraph ? latestGraph.nodes[selectedNodeId] : null);
	const latestConcept = $derived.by(() => {
		const conceptId = stores.concepts.table.order.at(-1);
		return conceptId ? stores.concepts.table.byId[conceptId] : null;
	});
	const latestCheck = $derived.by(() => {
		const sessionId = stores.workspace.current.activeSessionId;
		return sessionId ? stores.checks.latestForSession(sessionId) : null;
	});
	onMount(() => {
		void activateEngine(engineChoice);
		const keydown = (event: KeyboardEvent) => {
			const commandKey = event.metaKey || event.ctrlKey;
			if (!commandKey || event.shiftKey || event.altKey) {
				return;
			}
			if (event.key.toLowerCase() === 's') {
				event.preventDefault();
				void runCommand('file.save');
				return;
			}
			if (event.key.toLowerCase() === 'z') {
				event.preventDefault();
				void runCommand('doUndo');
			}
		};
		window.addEventListener('keydown', keydown);
		return () => window.removeEventListener('keydown', keydown);
	});

	function createEngine(choice: EngineChoice): IvyEngine {
		if (choice === 'hosted-go') {
			return new HostedGoEngine({
				baseUrl: import.meta.env.VITE_IVY_ENGINE_BASE_URL ?? ''
			});
		}
		if (choice === 'hosted-webui') {
			return new HostedWebuiEngine({
				baseUrl: import.meta.env.VITE_IVY_ENGINE_BASE_URL ?? '',
				now: nowIso
			});
		}
		if (choice === 'browser-wasm') {
			return new BrowserWasmEngine();
		}
		return new HostedGoEngine({
			baseUrl: import.meta.env.VITE_IVY_ENGINE_BASE_URL ?? ''
		});
	}

	async function activateEngine(choice: EngineChoice) {
		const serial = ++activationSerial;
		const oldService = service;
		activationFailureMessage = '';
		engineChoice = choice;
		sessionUi.setStatus(`Starting ${choice}`, 'info');
		sessionUi.showLoading(`Starting ${choice}`);
		await oldService.closeSession();
		if (serial !== activationSerial) {
			return;
		}
		const nextService = createEngineService({ engine: createEngine(choice), stores });
		service = nextService;
		try {
			const session = await nextService.startSession(project.id);
			if (serial !== activationSerial) {
				return;
			}
			sessionUi.setSessionId(session.id);
			await nextService.loadModel({ ...activeModel, text: editorText });
			if (serial !== activationSerial) {
				return;
			}
			sessionUi.setStatus(`${choice} ready`, 'success');
			selectedGraphId = stores.graphs.table.order.at(-1) ?? null;
		} catch (error) {
			if (serial !== activationSerial) {
				return;
			}
			activationFailureMessage = error instanceof Error ? error.message : `Failed to start ${choice}`;
			sessionUi.setStatus(activationFailureMessage, 'error');
		} finally {
			if (serial === activationSerial) {
				sessionUi.hideLoading();
			}
		}
	}

	function updateEditor(text: string) {
		editorText = text;
		editorUi.edit(text);
		models.updateText(initialModel.id, text);
	}

	async function runCommand(commandId: string, target?: { kind: GraphSnapshot['kind']; graphId?: string; nodeId?: string; obj?: string }) {
		if (runFileCommand(commandId)) {
			return;
		}
		const engineCommandId = commandId === 'runCheck' ? `check.${sessionUi.current.mode}` : routeWorkbenchCommand(commandId);
		try {
			if (!stores.workspace.current.activeSessionId) {
				sessionUi.setStatus('Starting engine session...', 'info');
				await activateEngine(engineChoice);
			}
			if (!stores.workspace.current.activeSessionId) {
				if (!activationFailureMessage) {
					sessionUi.setStatus('No active engine session', 'error');
				}
				return;
			}
			sessionUi.setStatus(`Running ${engineCommandId}`, 'info');
			if (engineCommandId.startsWith('check.')) {
				await service.loadModel({
					...activeModel,
					text: editorText,
					engineRevision: activeModel.engineRevision + 1,
					updatedAt: nowIso()
				});
			}
			const job = await service.runCommand({
				id: `intent-${crypto.randomUUID()}`,
				commandId: engineCommandId,
				target: target ?? { kind: engineCommandId === 'concept.action' ? 'concept' : 'arg' }
			});
			sessionUi.setStatus(`Finished ${job.kind}`, 'success');
			selectedGraphId = stores.graphs.table.order.at(-1) ?? null;
			selectedNodeId = null;
		} catch (error) {
			sessionUi.setStatus(error instanceof Error ? error.message : `Failed to run ${engineCommandId}`, 'error');
		}
	}

	function runFileCommand(commandId: string) {
		if (commandId === 'file.download') {
			downloadTextFile(modelDownloadFilename(activeModel.filename), editorText, 'text/plain');
			sessionUi.setStatus(`Downloaded: ${modelDownloadFilename(activeModel.filename)}`, 'success');
			return true;
		}
		if (commandId === 'file.saveAs' || commandId === 'file.save') {
			editorUi.markSaved(editorText);
			models.markSaved(initialModel.id, editorText);
			sessionUi.setStatus(`Saved: ${activeModel.filename}`, 'success');
			return true;
		}
		if (commandId === 'file.new') {
			editorText = '';
			editorUi.load('', '');
			models.updateText(initialModel.id, '');
			models.markSaved(initialModel.id, '');
			sessionUi.setStatus('New model ready', 'success');
			return true;
		}
		return false;
	}

	function setEditorKeymap(keymap: EditorKeymap) {
		editorUi.setKeymap(keymap);
		sessionUi.setStatus(`Editor keymap: ${keymap}`, 'success');
	}

	function setMode(mode: SessionMode) {
		sessionUi.setMode(mode);
		sessionUi.setStatus(`Mode: ${mode}`, 'success');
	}

	function toggleRelation(rowId: string, displayClass: string, checked: boolean) {
		const toggled = stores.stateRelations.toggle(rowId, displayClass as EdgeDisplayClass | NodeLabelDisplayClass, checked);
		sessionUi.setStatus(toggled ? `Updated relation visibility: ${rowId}` : `Relation toggle unavailable: ${rowId}`, toggled ? 'success' : 'warning');
	}

	function selectNode(nodeId: string) {
		selectedNodeId = nodeId;
	}

	async function runNodeAction(action: NodeAction, node: GraphNode) {
		await runCommand(action.action, {
			kind: latestGraph?.kind ?? 'arg',
			graphId: latestGraph?.id,
			nodeId: node.id,
			obj: node.obj
		});
	}
</script>

<svelte:head>
	<title>SVK Ivy Workspace</title>
</svelte:head>

<WorkbenchShell
	{engineChoice}
	mode={sessionUi.current.mode}
	{activeModel}
	{editorText}
	editorKeymap={editorUi.current.keymap}
	sessionLabel={sessionUi.current.sessionDisplay || stores.workspace.current.activeSessionId || 'none'}
	statusMessage={sessionUi.current.status}
	statusLevel={sessionUi.current.statusLevel}
	{latestGraph}
	{latestConcept}
	{latestCheck}
	{selectedNode}
	{selectedNodeId}
	{jobs}
	stateRelationRows={stores.stateRelations.current.rows}
	onActivateEngine={activateEngine}
	onSetMode={setMode}
	onRunCommand={runCommand}
	onUpdateEditor={updateEditor}
	onSetEditorKeymap={setEditorKeymap}
	onSelectNode={selectNode}
	onNodeAction={runNodeAction}
	onToggleRelation={toggleRelation}
/>
