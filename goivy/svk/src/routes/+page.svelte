<script lang="ts">
	import { onMount } from 'svelte';
	import { BrowserWasmEngine, FakeEngine, HostedWebuiEngine } from '$lib/engines';
	import GraphSnapshotView from '$lib/graphs/GraphSnapshotView.svelte';
	import { createEngineService } from '$lib/services';
	import { nowIso } from '$lib/time';
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
	import type { GraphNode, GraphSnapshot, ModelDocument, NodeAction, Project } from '$lib/types';
	import type { IvyEngine } from '$lib/types';

	type EngineChoice = 'fake' | 'hosted-webui' | 'browser-wasm';

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

	let editorText = $state(initialModel.text);
	let selectedNodeId = $state<string | null>(null);
	let selectedGraphId = $state<string | null>(null);
	let statusMessage = $state('Starting local fake engine');
	let engineChoice = $state<EngineChoice>('fake');
	let service = createEngineService({ engine: createEngine('fake'), stores });

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
	const lineNumbers = $derived(editorText.split('\n').map((_, index) => index + 1));

	onMount(() => {
		void activateEngine(engineChoice);
	});

	function createEngine(choice: EngineChoice): IvyEngine {
		if (choice === 'hosted-webui') {
			return new HostedWebuiEngine({
				baseUrl: import.meta.env.VITE_IVY_ENGINE_BASE_URL ?? '',
				now: nowIso
			});
		}
		if (choice === 'browser-wasm') {
			return new BrowserWasmEngine();
		}
		return new FakeEngine({ now: nowIso });
	}

	async function activateEngine(choice: EngineChoice) {
		engineChoice = choice;
		statusMessage = `Starting ${choice}`;
		await service.closeSession();
		service = createEngineService({ engine: createEngine(choice), stores });
		try {
			const session = await service.startSession(project.id);
			await service.loadModel({ ...activeModel, text: editorText });
			statusMessage = `${choice} ready: ${session.id}`;
			selectedGraphId = stores.graphs.table.order.at(-1) ?? null;
		} catch (error) {
			statusMessage = error instanceof Error ? error.message : `Failed to start ${choice}`;
		}
	}

	function updateEditor(text: string) {
		editorText = text;
		models.updateText(initialModel.id, text);
	}

	async function runCommand(commandId: string, target?: { kind: GraphSnapshot['kind']; graphId?: string; nodeId?: string; obj?: string }) {
		statusMessage = `Running ${commandId}`;
		if (commandId.startsWith('check.')) {
			await service.loadModel({
				...activeModel,
				text: editorText,
				engineRevision: activeModel.engineRevision + 1,
				updatedAt: nowIso()
			});
		}
		const job = await service.runCommand({
			id: `intent-${crypto.randomUUID()}`,
			commandId,
			target: target ?? { kind: commandId === 'concept.action' ? 'concept' : 'arg' }
		});
		statusMessage = `Finished ${job.kind}`;
		selectedGraphId = stores.graphs.table.order.at(-1) ?? null;
		selectedNodeId = null;
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

<main class="workspace-shell">
	<header class="topbar">
		<button type="button" class="menu-button">File</button>
		<span class="mode-label">MODE</span>
		<select aria-label="Mode" class="mode-select">
			<option>Induction</option>
			<option>Bounded</option>
			<option>Concrete</option>
		</select>
		<select
			aria-label="Engine"
			class="engine-select"
			value={engineChoice}
			onchange={(event) => void activateEngine(event.currentTarget.value as EngineChoice)}
		>
			<option value="fake">Fake engine</option>
			<option value="hosted-webui">Hosted webui</option>
			<option value="browser-wasm">Browser wasm</option>
		</select>
		<div class="toolbar" aria-label="Workspace commands">
			<button type="button" class="primary" data-testid="run-induction" onclick={() => void runCommand('check.induction')}>
				Check
			</button>
			<button type="button" onclick={() => void runCommand('concept.action')}>Show Reachable</button>
			<button type="button" onclick={() => void service.loadModel(activeModel)} disabled={!stores.workspace.current.activeSessionId}>Undo</button>
			<button type="button" onclick={() => models.markSaved(initialModel.id, editorText)} disabled={!activeModel.dirty}>Reset Domain</button>
			<button type="button" onclick={() => void runCommand('check.bounded')}>Bounded</button>
			<button type="button">Diagram Domain</button>
		</div>
		<button type="button" class="tutorial-button">Show Tutorial</button>
	</header>

	<section class="workspace-grid" aria-label="Ivy workspace">
		<nav class="sheet-tabs" aria-label="Sheets">
			<button type="button" class="active">Sheet 1</button>
			<button type="button">Step: call ext</button>
		</nav>

		<section class="pane arg-pane graph-pane" aria-label="ARG graph">
			<div class="pane-title stacked">
				<strong>ARG (Abstract Reachability Graph)</strong>
				<span>Invariant</span>
			</div>
			<div class="subtoolbar">
				<span>File</span>
				<span>Mode</span>
				<span>Action</span>
			</div>
			<GraphSnapshotView
				snapshot={latestGraph}
				{selectedNodeId}
				testId="arg-graph"
				onSelect={(node) => selectNode(node.id)}
				onAction={runNodeAction}
			/>
		</section>

		<section class="pane concept-pane" aria-label="Concept graph">
			<div class="pane-title">
				<strong>Concept graph</strong>
			</div>
			<div class="subtoolbar concept-toolbar">
				<span>Conjecture</span>
				<span>View</span>
				<span>Action</span>
				<span>View</span>
			</div>
			<div class="concept-canvas" data-testid="concept-graph">
				<div class="concept-node blue concept-client-zero">client<br />=0</div>
				<div class="concept-node blue concept-client-one">client<br />=1</div>
				<div class="concept-node red concept-server-zero">server<br />=0</div>
				<div class="concept-edge"></div>
				{#if latestConcept}
					<span class="concept-hidden">{Object.keys(latestConcept.concepts).join(', ')}</span>
				{/if}
			</div>
		</section>

		<section class="pane state-pane" aria-label="State relations">
			<div class="pane-title">
				<strong>State/relations</strong>
			</div>
			<div class="state-panel">
				<strong>State: 0</strong>
				<div class="relation-grid" aria-label="Relation toggles">
					<span>+</span><span>?</span><span>-</span><span>T</span><span></span>
					<label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><span>=@X</span>
					<label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><span>=@Y</span>
					<label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><span>=@Z</span>
					<label><input type="checkbox" checked /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><span>link(X,Y)</span>
					<label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><label><input type="checkbox" /></label><span>semaphore</span>
				</div>
			</div>
		</section>

		<section class="pane editor-pane" aria-label="Editor">
			<div class="pane-title">
				<strong>Editing: {activeModel.filename} [{activeModel.dirty ? 'unsaved' : 'saved'}]</strong>
				<span data-testid="dirty-indicator">{activeModel.dirty ? 'Unsaved' : 'Saved'}</span>
			</div>
			<div class="editor-controls">
				<button type="button">x</button>
				<label><input type="radio" checked /> Sublime</label>
				<label><input type="radio" /> Emacs-ish</label>
				<label><input type="radio" /> Vim</label>
				<a href="https://microsoft.github.io/monaco-editor/">keymap docs</a>
			</div>
			<div class="editor-wrap">
				<div class="line-gutter" aria-hidden="true">
					{#each lineNumbers as number (number)}
						<span>{number}</span>
					{/each}
				</div>
				<textarea
					data-testid="model-editor"
					spellcheck="false"
					value={editorText}
					oninput={(event) => updateEditor(event.currentTarget.value)}
				></textarea>
			</div>
		</section>

		<section class="pane details-pane" aria-label="Details and checks">
			<div class="pane-title compact-title">
				<strong>DETAILS</strong>
			</div>
			<div class="details-body" data-testid="details-pane">
				<p>Verification Result</p>
				{#if latestCheck}
					<p>{latestCheck.result.toUpperCase()} [Z3: {latestCheck.z3Contacted ? 'yes' : 'no'}]: {latestCheck.message}</p>
					{#if latestCheck.failedConjecture}
						<p>{latestCheck.failedConjecture}</p>
					{/if}
					{#if latestCheck.counterexampleTrace}
						<p>{latestCheck.counterexampleTrace}</p>
					{/if}
				{:else}
					<p>FAILED [Z3: yes]: The following conjecture is not relatively inductive:</p>
					<p>~(X:client ~= Z &amp; link(X,Y) &amp; link(Z,Y))</p>
				{/if}
				{#if selectedNode}
					<h2>{selectedNode.label}</h2>
					<p>{selectedNode.obj}</p>
					<p>{selectedNode.classes.join(', ')}</p>
				{/if}
			</div>
			<div class="job-strip" data-testid="job-strip">
				{#each jobs as job (job.id)}
					<div class="job-row">
						<span>{job.kind}</span>
						<strong>{job.status}</strong>
					</div>
				{/each}
			</div>
		</section>

		<footer class="statusbar" data-testid="status-strip">
			<span>
				{#if latestCheck}
					Check {latestCheck.result.toUpperCase()} ({latestCheck.mode}) [Z3: {latestCheck.z3Contacted ? 'yes' : 'no'}] - {latestCheck.message}
				{:else}
					Check FAILED (induction) [Z3: yes] - counterexample found
				{/if}
			</span>
			<span>Session: {stores.workspace.current.activeSessionId ?? statusMessage}</span>
		</footer>
	</section>
</main>

<style>
	:global(body) {
		margin: 0;
		font-family:
			Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
		background: #0e0f15;
		color: #d9d9df;
	}

	button,
	select,
	textarea {
		font: inherit;
	}

	.workspace-shell {
		height: 100vh;
		display: grid;
		grid-template-rows: auto 1fr;
		overflow: hidden;
		background: #101116;
	}

	.topbar {
		display: flex;
		gap: 7px;
		align-items: center;
		min-height: 36px;
		padding: 4px 6px;
		border-bottom: 1px solid #303238;
		background: #202123;
		box-sizing: border-box;
	}

	.toolbar {
		display: flex;
		gap: 4px;
		align-items: center;
		justify-content: flex-start;
		min-width: 0;
		flex: 1;
	}

	.toolbar button,
	.topbar select,
	.menu-button,
	.tutorial-button {
		border: 1px solid #3c3f45;
		background: #303236;
		color: #d7d7dc;
		border-radius: 3px;
		min-height: 26px;
		padding: 0 12px;
		cursor: pointer;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
	}

	.menu-button {
		border-color: transparent;
		background: transparent;
		padding: 0 8px;
	}

	.toolbar button:disabled {
		color: #777a82;
		cursor: default;
	}

	.toolbar .primary {
		background: #1265a8;
		border-color: #1d70b4;
		color: #ffffff;
		font-weight: 600;
	}

	.mode-label {
		color: #92939a;
		margin-left: 8px;
	}

	.mode-select {
		width: 118px;
	}

	.engine-select {
		width: 136px;
	}

	.tutorial-button {
		margin-left: auto;
	}

	.workspace-grid {
		display: grid;
		grid-template-columns: minmax(170px, 14%) minmax(360px, 32%) minmax(220px, 16%) minmax(430px, 38%);
		grid-template-rows: 36px minmax(0, 1fr) 170px 24px;
		gap: 0;
		min-height: 0;
		border-top: 1px solid #111;
		background: #111217;
	}

	.pane {
		min-width: 0;
		min-height: 0;
		display: grid;
		grid-template-rows: auto 1fr;
		border-right: 2px solid #2a2c32;
		border-bottom: 1px solid #303238;
		background: #151718;
		border-radius: 0;
		overflow: hidden;
	}

	.sheet-tabs {
		grid-column: 1 / 4;
		grid-row: 1;
		display: flex;
		align-items: stretch;
		background: #111315;
		border-bottom: 1px solid #292c33;
	}

	.sheet-tabs button {
		border: 0;
		border-right: 1px solid #25282e;
		background: transparent;
		color: #8f929a;
		padding: 0 18px;
		cursor: pointer;
	}

	.sheet-tabs .active {
		color: #f0f0f4;
		border-bottom: 3px solid #0c6db5;
	}

	.arg-pane {
		grid-column: 1;
		grid-row: 2;
		grid-template-rows: auto auto 1fr;
	}

	.concept-pane {
		grid-column: 2;
		grid-row: 2;
		grid-template-rows: auto auto 1fr;
	}

	.state-pane {
		grid-column: 3;
		grid-row: 2;
	}

	.editor-pane {
		grid-column: 4;
		grid-row: 1 / 4;
		grid-template-rows: auto auto 1fr;
		background: #10101d;
		border-right: 0;
	}

	.pane-title {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		min-height: 34px;
		padding: 6px 10px;
		border-bottom: 1px solid #2c2e34;
		background: #181a1c;
		font-size: 0.9rem;
		color: #e2e2e8;
		box-sizing: border-box;
	}

	.pane-title.stacked {
		display: grid;
		align-content: center;
		min-height: 64px;
	}

	.pane-title.stacked span {
		margin-top: 6px;
		color: #b5b6bd;
	}

	.pane-title span {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.subtoolbar {
		display: flex;
		gap: 26px;
		align-items: center;
		min-height: 32px;
		padding: 0 24px;
		color: #b9bbc2;
		background: #202123;
		border-bottom: 1px solid #303238;
	}

	.concept-toolbar {
		justify-content: space-around;
		gap: 10px;
	}

	.editor-controls {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 30px;
		padding: 0 10px;
		background: #17121f;
		border-bottom: 1px solid #262236;
		color: #cdccd4;
		font-size: 0.86rem;
	}

	.editor-controls button {
		width: 24px;
		height: 24px;
		border: 1px solid #353544;
		background: #20202a;
		color: #bfc0c8;
		border-radius: 4px;
	}

	.editor-controls a {
		color: #d7a543;
	}

	.editor-wrap {
		position: relative;
		min-height: 0;
		background: #11101d;
		overflow: hidden;
	}

	.line-gutter {
		position: absolute;
		inset: 0 auto 0 0;
		width: 44px;
		padding-top: 8px;
		background: #121120;
		border-right: 1px solid #242236;
		color: #55596b;
		font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace;
		font-size: 0.92rem;
		line-height: 1.45;
		text-align: right;
		box-sizing: border-box;
		pointer-events: none;
	}

	.line-gutter span {
		display: block;
		padding-right: 8px;
		height: 1.45em;
	}

	textarea {
		width: 100%;
		height: 100%;
		box-sizing: border-box;
		resize: none;
		border: 0;
		outline: 0;
		padding: 8px 14px 8px 58px;
		font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace;
		font-size: 0.92rem;
		line-height: 1.45;
		background: transparent;
		color: #eeeef4;
		caret-color: #ffffff;
	}

	.concept-canvas {
		position: relative;
		min-height: 0;
		background: #11101d;
		overflow: hidden;
	}

	.concept-node {
		position: absolute;
		display: grid;
		place-items: center;
		width: 112px;
		height: 112px;
		background: #f8f8f6;
		color: #111;
		--concept-border: #001eff;
		font-size: 1.7rem;
		line-height: 1.05;
		text-align: center;
		clip-path: polygon(28% 0, 72% 0, 100% 28%, 100% 72%, 72% 100%, 28% 100%, 0 72%, 0 28%);
	}

	.concept-node::after {
		content: "";
		position: absolute;
		inset: 0;
		clip-path: inherit;
		border: 8px solid var(--concept-border);
		pointer-events: none;
	}

	.concept-node.blue {
		--concept-border: #001eff;
	}

	.concept-node.red {
		--concept-border: #f10012;
	}

	.concept-client-zero {
		left: 10%;
		top: 8%;
	}

	.concept-client-one {
		right: 12%;
		top: 8%;
	}

	.concept-server-zero {
		left: 10%;
		bottom: 8%;
	}

	.concept-edge {
		position: absolute;
		left: calc(10% + 56px);
		top: calc(8% + 112px);
		width: 8px;
		height: 34%;
		background: #8d8d8d;
	}

	.concept-edge::after {
		content: "";
		position: absolute;
		left: -15px;
		bottom: -18px;
		border-left: 19px solid transparent;
		border-right: 19px solid transparent;
		border-top: 34px solid #8d8d8d;
	}

	.concept-hidden {
		position: absolute;
		left: -9999px;
	}

	.state-panel {
		padding: 10px 14px;
		color: #d8d8dd;
	}

	.relation-grid {
		display: grid;
		grid-template-columns: repeat(4, 28px) minmax(80px, 1fr);
		gap: 9px 8px;
		align-items: center;
		margin-top: 18px;
		color: #3c99d3;
		font-size: 0.86rem;
	}

	.relation-grid span:nth-child(-n + 4) {
		color: #b8bac1;
		text-align: center;
	}

	.relation-grid input {
		width: 18px;
		height: 18px;
		accent-color: #1377bd;
	}

	.details-pane {
		grid-column: 1 / 4;
		grid-row: 3;
		grid-template-rows: auto 1fr auto;
	}

	.compact-title {
		min-height: 28px;
		padding: 4px 10px;
		color: #75777f;
	}

	.details-body {
		padding: 8px 10px;
		overflow: auto;
		background: #181a1c;
		color: #dedee5;
		font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace;
		font-size: 0.9rem;
	}

	.details-body h2 {
		margin: 10px 0 4px;
		font-size: 0.95rem;
		font-family: inherit;
	}

	.details-body p {
		margin: 4px 0 16px;
		color: #dedee5;
	}

	.job-strip {
		border-top: 1px solid #303238;
		max-height: 56px;
		overflow: auto;
		background: #151718;
	}

	.job-row {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		padding: 4px 10px;
		font-size: 0.86rem;
		border-top: 1px solid #272a30;
		color: #dfe0e7;
	}

	.job-row:first-child {
		border-top: 0;
	}

	.statusbar {
		grid-column: 1 / 5;
		grid-row: 4;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 14px;
		padding: 0 10px;
		background: #c91524;
		color: #ffffff;
		font-weight: 700;
		font-size: 0.9rem;
	}

	@media (max-width: 900px) {
		.topbar {
			overflow-x: auto;
		}

		.toolbar {
			justify-content: flex-start;
		}

		.workspace-grid {
			grid-template-columns: 1fr;
			grid-template-rows: 36px 280px 300px 220px 420px 180px 24px;
		}

		.sheet-tabs,
		.arg-pane,
		.concept-pane,
		.state-pane,
		.editor-pane,
		.details-pane,
		.statusbar {
			grid-column: 1;
		}

		.sheet-tabs {
			grid-row: 1;
		}

		.arg-pane {
			grid-row: 2;
		}

		.concept-pane {
			grid-row: 3;
		}

		.state-pane {
			grid-row: 4;
		}

		.editor-pane {
			grid-row: 5;
		}

		.details-pane {
			grid-row: 6;
		}

		.statusbar {
			grid-row: 7;
		}

		.pane {
			border-right: 0;
		}

		.concept-node {
			width: 88px;
			height: 88px;
			font-size: 1.25rem;
		}
	}

	@media (max-width: 1280px) and (min-width: 901px) {
		.workspace-grid {
			grid-template-columns: minmax(150px, 16%) minmax(280px, 30%) minmax(190px, 16%) minmax(340px, 38%);
		}

		.concept-node {
			width: 92px;
			height: 92px;
			font-size: 1.25rem;
		}

		.subtoolbar {
			gap: 14px;
			padding: 0 14px;
		}

		.toolbar button:nth-last-child(-n + 2) {
			display: none;
		}
	}

	@media (max-height: 720px) {
		.workspace-grid {
			grid-template-rows: 34px minmax(0, 1fr) 130px 24px;
		}

		.pane-title.stacked {
			min-height: 52px;
		}

		.subtoolbar {
			min-height: 28px;
		}
	}

	@media (max-width: 520px) {
		.statusbar {
			font-size: 0.75rem;
		}

		.tutorial-button,
		.mode-label {
			display: none;
		}

		.mode-select {
			width: 104px;
		}

		.toolbar button {
			padding: 0 8px;
		}

		.editor-controls {
			flex-wrap: wrap;
			min-height: 58px;
			align-content: center;
		}

		.editor-pane {
			grid-row: auto;
		}
	}
</style>
