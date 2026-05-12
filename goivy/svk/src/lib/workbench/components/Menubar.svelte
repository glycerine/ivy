<script lang="ts">
	import { FILE_MENU } from '$lib/workbench/commands/menuDefinitions';
	import type { SessionMode } from '$lib/workbench/state';

	type EngineChoice = 'hosted-go' | 'hosted-webui' | 'browser-wasm';

	type Props = {
		engineChoice: EngineChoice;
		mode: SessionMode;
		onActivateEngine: (choice: EngineChoice) => void | Promise<void>;
		onSetMode: (mode: SessionMode) => void;
		onRunCommand: (commandId: string) => void | Promise<void>;
	};

	let {
		engineChoice,
		mode,
		onActivateEngine,
		onSetMode,
		onRunCommand
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
	<select aria-label="Mode" class="mode-select" value={mode} onchange={(event) => onSetMode(event.currentTarget.value as SessionMode)}>
		<option value="induction">Induction</option>
		<option value="pdr">PDR</option>
		<option value="concrete">Concrete</option>
		<option value="abstract">Abstract</option>
		<option value="bounded">Bounded</option>
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
		<button type="button" class="primary" data-testid="run-check" onclick={() => void onRunCommand('runCheck')}>
			Check
		</button>
		<button type="button" onclick={() => void onRunCommand('showReachableStates')}>Show Reachable</button>
		<button type="button" onclick={() => void onRunCommand('doUndo')}>Undo</button>
		<button type="button" onclick={() => void onRunCommand('resetDomain')}>Reset Domain</button>
		<button type="button" onclick={() => void onRunCommand('diagramDomain')}>Diagram Domain</button>
	</div>
	<button type="button" class="tutorial-button">Show Tutorial</button>
</header>
