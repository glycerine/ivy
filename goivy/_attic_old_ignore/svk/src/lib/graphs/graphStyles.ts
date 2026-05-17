import type cytoscape from 'cytoscape';
import type { GraphKind } from '$lib/types';

export function graphStyleFor(kind: GraphKind): cytoscape.StylesheetJson {
	const accent = kind === 'concept' ? '#6b7f1d' : kind === 'proof' ? '#7a4f11' : '#2f6f87';
	return [
		{
			selector: 'node',
			style: {
				label: 'data(label)',
				'background-color': '#e9f7f6',
				'border-color': accent,
				'border-width': 2,
				color: '#15363f',
				'font-size': 12,
				'text-valign': 'center',
				'text-halign': 'center',
				width: 'label',
				height: 'label',
				padding: '14px',
				shape: 'ellipse'
			}
		},
		{
			selector: 'node:selected',
			style: {
				'background-color': '#fff1d6',
				'border-color': '#7a4f11'
			}
		},
		{
			selector: 'edge',
			style: {
				label: 'data(label)',
				width: 2,
				'line-color': '#9aa8b7',
				'target-arrow-color': '#9aa8b7',
				'target-arrow-shape': 'triangle',
				'curve-style': 'bezier',
				'font-size': 11,
				color: '#425466'
			}
		}
	] as cytoscape.StylesheetJson;
}
