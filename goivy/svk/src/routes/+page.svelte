<script lang="ts">
	import { onMount } from 'svelte';
	import { FakeEngine } from '$lib/engines';
	import GraphSnapshotView from '$lib/graphs/GraphSnapshotView.svelte';
	import { createEngineService } from '$lib/services';
	import { nowIso } from '$lib/time';
	import {
		createConceptsState,
		createEnginesState,
		createGraphsState,
		createJobsState,
		createModelsState,
		createWorkspaceState
	} from '$lib/state';
	import type { GraphNode, GraphSnapshot, ModelDocument, NodeAction, Project } from '$lib/types';

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
		filename: 'demo.ivy',
		text: `#lang ivy1.8
type node
relation link(X:node,Y:node)
individual root:node
conjecture root_reaches_self = link(root,root)`,
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
		concepts: createConceptsState()
	};
	const models = createModelsState();
	models.upsert(initialModel);

	const engine = new FakeEngine({ now: nowIso });
	const service = createEngineService({ engine, stores });

	let editorText = $state(initialModel.text);
	let selectedNodeId = $state<string | null>(null);
	let selectedGraphId = $state<string | null>(null);
	let statusMessage = $state('Starting local fake engine');

	const activeModel = $derived(models.table.byId[initialModel.id]);
	const jobs = $derived(stores.jobs.table.order.map((id) => stores.jobs.table.byId[id]));
	const latestGraph = $derived.by(() => {
		const graphId = selectedGraphId ?? stores.graphs.table.order.at(-1);
		return graphId ? stores.graphs.table.byId[graphId] : null;
	});
	const graphNodes = $derived<GraphNode[]>(latestGraph ? latestGraph.nodeOrder.map((id) => latestGraph.nodes[id]) : []);
	const selectedNode = $derived(selectedNodeId && latestGraph ? latestGraph.nodes[selectedNodeId] : null);
	const latestConcept = $derived.by(() => {
		const conceptId = stores.concepts.table.order.at(-1);
		return conceptId ? stores.concepts.table.byId[conceptId] : null;
	});

	onMount(() => {
		void (async () => {
			const session = await service.startSession(project.id);
			await service.loadModel(initialModel);
			statusMessage = `Local engine ready: ${session.id}`;
			selectedGraphId = stores.graphs.table.order.at(-1) ?? null;
			selectedNodeId = graphNodes[0]?.id ?? null;
		})();
	});

	function updateEditor(text: string) {
		editorText = text;
		models.updateText(initialModel.id, text);
	}

	async function runCommand(commandId: string, target?: { kind: GraphSnapshot['kind']; graphId?: string; nodeId?: string; obj?: string }) {
		statusMessage = `Running ${commandId}`;
		const job = await service.runCommand({
			id: `intent-${crypto.randomUUID()}`,
			commandId,
			target: target ?? { kind: commandId === 'concept.action' ? 'concept' : 'arg' }
		});
		statusMessage = `Finished ${job.kind}`;
		selectedGraphId = stores.graphs.table.order.at(-1) ?? null;
		selectedNodeId = graphNodes[0]?.id ?? null;
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
		<div class="project-mark">
			<strong>SVK</strong>
			<span>{project.name}</span>
		</div>
		<div class="toolbar" aria-label="Workspace commands">
			<button type="button" onclick={() => void service.loadModel(activeModel)} disabled={!stores.workspace.current.activeSessionId}>
				Load
			</button>
			<button type="button" onclick={() => models.markSaved(initialModel.id, editorText)} disabled={!activeModel.dirty}>
				Save local
			</button>
			<button type="button" data-testid="run-induction" onclick={() => void runCommand('check.induction')}>
				Induction
			</button>
			<button type="button" onclick={() => void runCommand('check.bounded')}>Bounded</button>
			<button type="button" onclick={() => void runCommand('concept.action')}>Concept</button>
			<select aria-label="Engine" value={engine.kind}>
				<option value="fake">Fake engine</option>
			</select>
		</div>
		<div class="session-status" data-testid="status-strip">{statusMessage}</div>
	</header>

	<section class="workspace-grid" aria-label="Ivy workspace">
		<section class="pane editor-pane" aria-label="Editor">
			<div class="pane-title">
				<span>{activeModel.filename}</span>
				<span data-testid="dirty-indicator">{activeModel.dirty ? 'Unsaved' : 'Saved'}</span>
			</div>
			<textarea
				data-testid="model-editor"
				spellcheck="false"
				value={editorText}
				oninput={(event) => updateEditor(event.currentTarget.value)}
			></textarea>
		</section>

		<section class="pane graph-pane" aria-label="ARG graph">
			<div class="pane-title">
				<span>ARG graph</span>
				<span>{latestGraph?.kind ?? 'empty'}</span>
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
				<span>Concept graph</span>
				<span>{latestConcept ? `${latestConcept.revision}` : 'empty'}</span>
			</div>
			<div class="concept-list" data-testid="concept-graph">
				{#if latestConcept}
					{#each Object.values(latestConcept.concepts) as concept (concept.name)}
						<button type="button">{concept.name}</button>
					{/each}
				{/if}
			</div>
		</section>

		<section class="pane details-pane" aria-label="Details and checks">
			<div class="pane-title">
				<span>Details</span>
				<span>{jobs.length} jobs</span>
			</div>
			<div class="details-body" data-testid="details-pane">
				{#if selectedNode}
					<h2>{selectedNode.label}</h2>
					<p>{selectedNode.obj}</p>
					<p>{selectedNode.classes.join(', ')}</p>
				{:else}
					<h2>No node selected</h2>
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
	</section>
</main>

<style>
	:global(body) {
		margin: 0;
		font-family:
			Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
		background: #f6f7f9;
		color: #17202a;
	}

	button,
	select,
	textarea {
		font: inherit;
	}

	.workspace-shell {
		min-height: 100vh;
		display: grid;
		grid-template-rows: auto 1fr;
	}

	.topbar {
		display: grid;
		grid-template-columns: minmax(180px, 1fr) minmax(360px, 2fr) minmax(220px, 1fr);
		gap: 12px;
		align-items: center;
		padding: 10px 14px;
		border-bottom: 1px solid #d7dde5;
		background: #ffffff;
	}

	.project-mark {
		display: flex;
		align-items: baseline;
		gap: 10px;
		min-width: 0;
	}

	.project-mark strong {
		font-size: 0.95rem;
		letter-spacing: 0;
	}

	.project-mark span,
	.session-status {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.toolbar {
		display: flex;
		gap: 8px;
		align-items: center;
		justify-content: center;
		min-width: 0;
	}

	.toolbar button,
	.toolbar select,
	.concept-list button {
		border: 1px solid #bbc6d4;
		background: #ffffff;
		color: #17202a;
		border-radius: 6px;
		min-height: 32px;
		padding: 0 10px;
		cursor: pointer;
	}

	.toolbar button:disabled {
		color: #8a97a6;
		cursor: default;
	}

	.session-status {
		text-align: right;
		color: #425466;
		font-size: 0.9rem;
	}

	.workspace-grid {
		display: grid;
		grid-template-columns: minmax(280px, 1.1fr) minmax(320px, 1fr);
		grid-template-rows: minmax(280px, 1.1fr) minmax(220px, 0.9fr);
		gap: 10px;
		padding: 10px;
		min-height: 0;
	}

	.pane {
		min-width: 0;
		min-height: 0;
		display: grid;
		grid-template-rows: auto 1fr;
		border: 1px solid #d7dde5;
		background: #ffffff;
		border-radius: 8px;
		overflow: hidden;
	}

	.editor-pane {
		grid-row: 1 / span 2;
	}

	.pane-title {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		padding: 8px 10px;
		border-bottom: 1px solid #e1e6ed;
		background: #fbfcfd;
		font-size: 0.86rem;
		color: #425466;
	}

	.pane-title span {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	textarea {
		width: 100%;
		height: 100%;
		box-sizing: border-box;
		resize: none;
		border: 0;
		outline: 0;
		padding: 12px;
		font-family: "SFMono-Regular", Consolas, "Liberation Mono", monospace;
		font-size: 0.92rem;
		line-height: 1.45;
		background: #fcfdff;
		color: #111820;
	}

	.concept-list {
		display: flex;
		align-content: flex-start;
		gap: 8px;
		flex-wrap: wrap;
		padding: 12px;
	}

	.details-pane {
		grid-template-rows: auto 1fr auto;
	}

	.details-body {
		padding: 12px;
		overflow: auto;
	}

	.details-body h2 {
		margin: 0 0 8px;
		font-size: 1rem;
	}

	.details-body p {
		margin: 6px 0;
		color: #425466;
	}

	.job-strip {
		border-top: 1px solid #e1e6ed;
		max-height: 118px;
		overflow: auto;
	}

	.job-row {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		padding: 7px 10px;
		font-size: 0.86rem;
		border-top: 1px solid #eef2f5;
	}

	.job-row:first-child {
		border-top: 0;
	}

	@media (max-width: 820px) {
		.topbar {
			grid-template-columns: 1fr;
			align-items: stretch;
		}

		.toolbar {
			justify-content: flex-start;
			overflow-x: auto;
			padding-bottom: 2px;
		}

		.session-status {
			text-align: left;
		}

		.workspace-grid {
			grid-template-columns: 1fr;
			grid-template-rows: 360px 260px 220px 280px;
		}

		.editor-pane {
			grid-row: auto;
		}
	}
</style>
