<script lang="ts">
	import type { TraceEventSheet } from '$lib/types';
	import EventTraceNode from './EventTraceNode.svelte';

	type Props = {
		sheet: TraceEventSheet | null;
	};

	let { sheet }: Props = $props();
</script>

<section class="event-viewer" aria-label="Event trace">
	<div class="event-tree-panel panel">
		<div class="pane-title">
			<strong>{sheet?.label ?? 'Event trace'}</strong>
			<div class="event-pattern-buttons">
				<button type="button">Filter</button>
				<button type="button">Find</button>
			</div>
		</div>
		<div class="event-tree">
			{#if sheet}
				<ul class="event-tree-list">
					{#each sheet.events as event (event.address)}
						<EventTraceNode node={event} />
					{/each}
				</ul>
			{:else}
				<div class="empty-graph-state">No event trace loaded</div>
			{/if}
		</div>
	</div>
</section>
