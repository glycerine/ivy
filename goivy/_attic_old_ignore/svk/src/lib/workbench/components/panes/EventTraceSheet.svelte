<script lang="ts">
	import type { TraceEventSheet } from '$lib/types';
	import EventTraceNode from './EventTraceNode.svelte';

	type Props = {
		sheet: TraceEventSheet | null;
		onCommand?: (commandId: string, value?: string) => void | Promise<void>;
	};

	let { sheet, onCommand }: Props = $props();
	let pattern = $state('');
</script>

<section class="event-viewer" aria-label="Event trace">
	<div class="event-tree-panel panel">
		<div class="pane-title">
			<strong>{sheet?.label ?? 'Event trace'}</strong>
			<div class="event-pattern-buttons">
				<input aria-label="Event pattern" bind:value={pattern} />
				<button type="button" onclick={() => void onCommand?.('filterEventTrace', pattern)}>Filter</button>
				<button type="button" onclick={() => void onCommand?.('findEventTrace', pattern)}>Find</button>
				<button type="button" onclick={() => void onCommand?.('addEventPattern', pattern)}>+</button>
				<button type="button" onclick={() => void onCommand?.('removeSelectedEventPattern')}>-</button>
				<button type="button" onclick={() => void onCommand?.('saveEventPatterns')}>Save</button>
				<button type="button" onclick={() => void onCommand?.('loadEventPatterns')}>Load</button>
				<button type="button" onclick={() => void onCommand?.('clearEventPatterns')}>Clear</button>
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
