<script lang="ts">
	import { FILE_MENU } from '$lib/workbench/commands/menuDefinitions';

	type EngineChoice = 'hosted-go' | 'hosted-webui' | 'browser-wasm';

	type Props = {
		engineChoice: EngineChoice;
		hasActiveSession: boolean;
		dirty: boolean;
		onActivateEngine: (choice: EngineChoice) => void | Promise<void>;
		onRunCommand: (commandId: string) => void | Promise<void>;
		onReloadModel: () => void | Promise<void>;
		onMarkSaved: () => void;
	};

	let {
		engineChoice,
		hasActiveSession,
		dirty,
		onActivateEngine,
		onRunCommand,
		onReloadModel,
		onMarkSaved
	}: Props = $props();
</script>

<header class="topbar">
	<div class="dropdown">
		<button type="button" class="menu-button">File</button>
		<div class="dropdown-content">
			{#each FILE_MENU as item (item.id)}
				{#if item.separatorBefore}
					<div class="dropdown-sep"></div>
				{/if}
				<button type="button" onclick={() => void onRunCommand(item.commandId)}>{item.label}</button>
			{/each}
		</div>
	</div>
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
		onchange={(event) => void onActivateEngine(event.currentTarget.value as EngineChoice)}
	>
		<option value="hosted-go">Native Go engine</option>
		<option value="browser-wasm">Browser wasm</option>
		<option value="hosted-webui">Legacy webui server</option>
	</select>
	<div class="toolbar" aria-label="Workspace commands">
		<button type="button" class="primary" data-testid="run-induction" onclick={() => void onRunCommand('check.induction')}>
			Check
		</button>
		<button type="button" onclick={() => void onRunCommand('concept.action')}>Show Reachable</button>
		<button type="button" onclick={() => void onReloadModel()} disabled={!hasActiveSession}>Undo</button>
		<button type="button" onclick={onMarkSaved} disabled={!dirty}>Reset Domain</button>
		<button type="button" onclick={() => void onRunCommand('check.bounded')}>Bounded</button>
		<button type="button">Diagram Domain</button>
	</div>
	<button type="button" class="tutorial-button">Show Tutorial</button>
</header>
