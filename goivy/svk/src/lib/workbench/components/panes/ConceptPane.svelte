<script lang="ts">
	import type { ConceptState } from '$lib/types';
	import { CONCEPT_CONJECTURE_MENU, CONCEPT_VIEW_MENU } from '$lib/workbench/commands/menuDefinitions';
	import GraphSnapshotView from '$lib/graphs/GraphSnapshotView.svelte';
	import { conceptStateToGraphSnapshot } from '$lib/graphs/conceptGraph';

	type Props = {
		latestConcept: ConceptState | null;
		onRunCommand?: (commandId: string) => void | Promise<void>;
	};

	let { latestConcept, onRunCommand }: Props = $props();
	const conceptGraph = $derived(latestConcept ? conceptStateToGraphSnapshot(latestConcept) : null);
</script>

<section class="pane concept-pane" aria-label="Concept graph">
	<div class="pane-title">
		<strong>Concept graph</strong>
	</div>
	<div class="subtoolbar concept-toolbar">
		<span class="dropdown">
			<button type="button" class="panel-menu">Conjecture</button>
			<span class="dropdown-content">
				{#each CONCEPT_CONJECTURE_MENU as item (item.id)}
					{#if item.separatorBefore}
						<span class="dropdown-sep"></span>
					{/if}
					<button type="button" onclick={() => void onRunCommand?.(item.commandId)}>{item.label}</button>
				{/each}
			</span>
		</span>
		<span class="dropdown">
			<button type="button" class="panel-menu">View</button>
			<span class="dropdown-content">
				{#each CONCEPT_VIEW_MENU as item (item.id)}
					<button type="button" onclick={() => void onRunCommand?.(item.commandId)}>{item.label}</button>
				{/each}
			</span>
		</span>
		<span>Action</span>
		<span>View</span>
	</div>
	<GraphSnapshotView snapshot={conceptGraph} testId="concept-graph" />
</section>
