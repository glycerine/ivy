export type LayoutState = {
	leftPaneWidth: number;
	rightPaneWidth: number;
	bottomPaneHeight: number;
	activeTabByRegion: Record<string, string>;
	tutorialVisible: boolean;
	compactMode: boolean;
};

export function createLayoutState(initial: Partial<LayoutState> = {}) {
	const state = $state<LayoutState>({
		leftPaneWidth: 360,
		rightPaneWidth: 420,
		bottomPaneHeight: 240,
		activeTabByRegion: {},
		tutorialVisible: false,
		compactMode: false,
		...initial
	});

	return {
		get current() {
			return state;
		},
		setPaneSizes(sizes: Partial<Pick<LayoutState, 'leftPaneWidth' | 'rightPaneWidth' | 'bottomPaneHeight'>>) {
			Object.assign(state, sizes);
		},
		setActiveTab(region: string, tabId: string) {
			state.activeTabByRegion[region] = tabId;
		},
		setTutorialVisible(visible: boolean) {
			state.tutorialVisible = visible;
		},
		setCompactMode(compact: boolean) {
			state.compactMode = compact;
		}
	};
}
