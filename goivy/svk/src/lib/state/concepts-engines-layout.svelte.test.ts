import { describe, expect, it } from 'vitest';
import { createConceptsState } from './concepts.svelte';
import { createEnginesState } from './engines.svelte';
import { createLayoutState } from './layout.svelte';
import type { ConceptState, EngineSession } from '$lib/types';

const concept: ConceptState = {
	id: 'concept-1',
	sessionId: 'session-1',
	sheetId: 'sheet-1',
	concepts: {},
	sortNodes: [],
	relationEdges: [],
	nodeLabels: [],
	abstractValue: {},
	toggles: { edges: {}, labels: {} },
	revision: 0
};

const engine: EngineSession = {
	id: 'engine-1',
	kind: 'fake',
	status: 'ready',
	projectId: 'project-1',
	capabilities: {
		offline: true,
		persistentJobs: false,
		cancelJob: true,
		eventStream: true,
		parallelJobs: false
	},
	createdAt: '2026-05-12T00:00:00.000Z',
	updatedAt: '2026-05-12T00:00:00.000Z'
};

describe('concept, engine, and layout state', () => {
	it('updates selected concept with revision bumps', () => {
		expect.hasAssertions();

		const concepts = createConceptsState();
		concepts.upsert(concept);
		concepts.selectConcept('concept-1', 'ready');

		expect(concepts.table.byId['concept-1'].selectedConcept).toBe('ready');
		expect(concepts.table.byId['concept-1'].revision).toBe(1);
		expect(concepts.table.revision).toBe(2);
	});

	it('updates engine status', () => {
		expect.hasAssertions();

		const engines = createEnginesState();
		engines.upsert(engine);
		engines.setStatus('engine-1', 'failed', 'boom', '2026-05-12T00:00:01.000Z');

		expect(engines.table.byId['engine-1']).toMatchObject({
			status: 'failed',
			error: 'boom',
			updatedAt: '2026-05-12T00:00:01.000Z'
		});
	});

	it('updates layout preferences without effects', () => {
		expect.hasAssertions();

		const layout = createLayoutState();
		layout.setPaneSizes({ leftPaneWidth: 300, bottomPaneHeight: 180 });
		layout.setActiveTab('main', 'arg');
		layout.setTutorialVisible(true);

		expect(layout.current).toMatchObject({
			leftPaneWidth: 300,
			bottomPaneHeight: 180,
			tutorialVisible: true
		});
		expect(layout.current.activeTabByRegion.main).toBe('arg');
	});
});
