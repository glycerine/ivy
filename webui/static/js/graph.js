/**
 * IvyGraph - Cytoscape.js graph wrapper for Ivy verification UI.
 *
 * Manages a Cytoscape instance with styles ported directly from
 * the Python cy_styles.py (concept_style, arg_style, proof_style).
 */

'use strict';

/* ============================================================
   Cytoscape.js Style Definitions
   Ported from ivy/cy_styles.py
   ============================================================ */

/**
 * Concept graph styles - used for displaying abstract concept graphs.
 * Node classes: non_existing, exactly_one, at_least_one, at_most_one, node_unknown
 * Edge classes: none_to_none, all_to_all, edge_unknown, total, functional, injective, surjective
 */
var CONCEPT_STYLE = [
    // Base node style
    {
        selector: 'node',
        style: {
            'content': 'data(label)',
            'text-wrap': 'wrap',
            'font-size': '14px',
            'text-outline-width': '3px',
            'text-outline-color': '#888',
            'text-valign': 'center',
            'color': '#fff',
            'width': 'data(width)',
            'height': 'data(height)',
            'border-color': 'data(border_color)',
            'shape': 'data(shape)',
            'background-color': '#888',
        },
    },

    // Node classes
    {
        selector: 'node.non_existing',
        style: {
            'display': 'none',
        },
    },
    {
        selector: 'node.exactly_one',
        style: {
            'border-width': '4px',
            'border-style': 'solid',
        },
    },
    {
        selector: 'node.at_least_one',
        style: {
            'border-width': '8px',
            'border-style': 'double',
        },
    },
    {
        selector: 'node.at_most_one',
        style: {
            'border-width': '3px',
            'border-style': 'dotted',
        },
    },
    {
        selector: 'node.node_unknown',
        style: {
            'border-width': '0px',
        },
    },

    // Base edge style
    {
        selector: 'edge',
        style: {
            'target-arrow-shape': 'triangle',
            'target-arrow-fill': 'filled',
            'source-arrow-fill': 'filled',
            'curve-style': 'bezier',
        },
    },

    // Edge classes
    {
        selector: 'edge.none_to_none',
        style: {
            'width': '4px',
            'line-style': 'dashed',
            'target-arrow-shape': 'triangle',
            'target-arrow-fill': 'filled',
            'source-arrow-fill': 'filled',
        },
    },
    {
        selector: 'edge.all_to_all',
        style: {
            'width': '4px',
            'line-style': 'solid',
        },
    },
    {
        selector: 'edge.edge_unknown',
        style: {
            'width': '4px',
            'line-style': 'dotted',
        },
    },
    {
        selector: 'edge.total',
        style: {
            'source-arrow-shape': 'circle',
            'source-arrow-fill': 'filled',
        },
    },
    {
        selector: 'edge.functional',
        style: {
            'source-arrow-shape': 'square',
        },
    },
    {
        selector: 'edge.injective',
        style: {
            'target-arrow-shape': 'triangle-backcurve',
        },
    },
    {
        selector: 'edge.surjective',
        style: {
            'target-arrow-fill': 'filled',
        },
    },

    // Selection
    {
        selector: 'node:selected',
        style: {
            'overlay-opacity': 0.2,
        },
    },
    {
        selector: 'edge:selected',
        style: {
            'overlay-opacity': 0.2,
        },
    },

    // Highlight class for active node
    {
        selector: 'node.highlighted',
        style: {
            'overlay-color': '#007acc',
            'overlay-opacity': 0.25,
            'overlay-padding': 6,
        },
    },
];


/**
 * ARG (Analysis Reachability Graph) styles.
 * Node classes: state, bottom_state
 * Edge classes: transition_action, transition_join, cover
 */
var ARG_STYLE = [
    // Base node style
    {
        selector: 'node',
        style: {
            'content': 'data(label)',
            'text-wrap': 'wrap',
            'text-outline-width': '3px',
            'text-outline-color': '#888',
            'text-valign': 'center',
            'color': '#fff',
            'width': '50',
            'height': '50',
            'border-color': '#000',
            'background-color': '#888',
        },
    },

    // Node classes
    {
        selector: 'node.state',
        style: {},
    },
    {
        selector: 'node.bottom_state',
        style: {
            'background-color': '#000',
            'text-outline-color': '#000',
        },
    },

    // Base edge style
    {
        selector: 'edge',
        style: {
            'content': 'data(label)',
            'width': '4px',
            'line-style': 'solid',
            'edge-text-rotation': 'none',
            'curve-style': 'bezier',
            'color': '#c8c8c8',
            'text-outline-width': '2px',
            'text-outline-color': '#1a1a2e',
            'font-size': '11px',
        },
    },

    // Edge classes
    {
        selector: 'edge.transition_join',
        style: {
            'target-arrow-shape': 'triangle-backcurve',
        },
    },
    {
        selector: 'edge.transition_action',
        style: {
            'target-arrow-shape': 'triangle',
        },
    },
    {
        selector: 'edge.cover',
        style: {
            'content': '',
            'target-arrow-shape': 'triangle',
            'line-style': 'dashed',
        },
    },

    // Selection
    {
        selector: 'node:selected',
        style: {
            'overlay-opacity': 0.2,
        },
    },
    {
        selector: 'edge:selected',
        style: {
            'overlay-opacity': 0.2,
        },
    },

    // Highlight class for selected ARG node
    {
        selector: 'node.highlighted',
        style: {
            'overlay-color': '#007acc',
            'overlay-opacity': 0.3,
            'overlay-padding': 8,
        },
    },
];


/**
 * Proof graph styles.
 * Node classes: proof_goal, refuted
 */
var PROOF_STYLE = [
    // Base node style
    {
        selector: 'node',
        style: {
            'content': 'data(label)',
            'text-wrap': 'wrap',
            'text-outline-width': '3px',
            'text-outline-color': '#888',
            'text-valign': 'center',
            'color': '#fff',
            'width': '50',
            'height': '50',
            'border-color': '#000',
            'background-color': '#888',
        },
    },

    // Base edge
    {
        selector: 'edge',
        style: {
            'target-arrow-shape': 'triangle',
            'curve-style': 'bezier',
        },
    },

    // Node classes
    {
        selector: 'node.refuted',
        style: {
            'background-color': '#000',
            'text-outline-color': '#000',
        },
    },

    // Selection
    {
        selector: 'node:selected',
        style: {
            'overlay-opacity': 0.2,
        },
    },
    {
        selector: 'edge:selected',
        style: {
            'overlay-opacity': 0.2,
        },
    },
];


/* ============================================================
   IvyGraph Class
   ============================================================ */

class IvyGraph {
    /**
     * @param {string} containerId - DOM element ID for the Cytoscape container
     * @param {Array} style - Cytoscape style array (CONCEPT_STYLE, ARG_STYLE, or PROOF_STYLE)
     */
    constructor(containerId, style) {
        this.containerId = containerId;
        this.cy = cytoscape({
            container: document.getElementById(containerId),
            style: style,
            layout: { name: 'preset' },
            userZoomingEnabled: true,
            userPanningEnabled: true,
            boxSelectionEnabled: false,
            minZoom: 0.2,
            maxZoom: 5,
        });

        // Register dagre layout if available
        if (typeof cytoscape !== 'undefined' && typeof cytoscapeDagre !== 'undefined') {
            cytoscape.use(cytoscapeDagre);
        }

        this._clickCallbacks = {};
        this._rightClickCallbacks = {};
    }

    /**
     * Update the graph with new Cytoscape elements JSON.
     * Elements are in the standard Cytoscape.js JSON format:
     * [{group: 'nodes', data: {...}, classes: '...'}, ...]
     *
     * @param {Array} elements - Cytoscape elements array
     * @param {object} [positions] - Optional map of node ID -> {x, y} positions
     */
    update(elements, positions) {
        this.cy.elements().remove();

        if (!elements || elements.length === 0) {
            return;
        }

        // Add elements, setting default data values for missing fields
        var toAdd = [];
        for (var i = 0; i < elements.length; i++) {
            var el = elements[i];
            // Default width/height for concept nodes if not provided
            if (el.group === 'nodes') {
                if (!el.data.width) {
                    el.data.width = el.data.label ? Math.max(50, el.data.label.length * 8 + 20) : 50;
                }
                if (!el.data.height) {
                    el.data.height = 50;
                }
                if (!el.data.shape) {
                    el.data.shape = 'ellipse';
                }
            }
            toAdd.push(el);
        }

        this.cy.add(toAdd);

        // Apply positions if provided (from server-side dot_layout)
        if (positions) {
            for (var nodeId in positions) {
                if (positions.hasOwnProperty(nodeId)) {
                    var node = this.cy.getElementById(nodeId);
                    if (node.length > 0) {
                        node.position(positions[nodeId]);
                    }
                }
            }
            this.cy.fit(undefined, 30);
        } else {
            // Auto-layout with dagre
            this.runLayout();
        }
    }

    /**
     * Run the dagre (hierarchical) layout algorithm.
     * @param {object} [options] - Additional dagre layout options
     */
    runLayout(options) {
        var defaults = {
            name: 'dagre',
            rankDir: 'TB',
            nodeSep: 50,
            rankSep: 80,
            edgeSep: 10,
            animate: false,
            fit: true,
            padding: 30,
        };

        if (options) {
            for (var key in options) {
                if (options.hasOwnProperty(key)) {
                    defaults[key] = options[key];
                }
            }
        }

        try {
            this.cy.layout(defaults).run();
        } catch (e) {
            // Fallback to grid if dagre fails (e.g., plugin not loaded)
            console.warn('Dagre layout failed, falling back to grid:', e);
            this.cy.layout({ name: 'grid', fit: true, padding: 30 }).run();
        }
    }

    /**
     * Register a callback for node tap (left-click).
     * @param {function} callback - function(nodeData) where nodeData is the node's data object
     */
    onNodeClick(callback) {
        this.cy.on('tap', 'node', function (evt) {
            callback(evt.target.data(), evt);
        });
    }

    /**
     * Register a callback for node right-click (context tap).
     * @param {function} callback - function(nodeData, renderedPosition)
     */
    onNodeRightClick(callback) {
        this.cy.on('cxttap', 'node', function (evt) {
            var pos = evt.renderedPosition || evt.target.renderedPosition();
            callback(evt.target.data(), pos, evt);
        });
    }

    /**
     * Register a callback for edge tap (left-click).
     * @param {function} callback - function(edgeData)
     */
    onEdgeClick(callback) {
        this.cy.on('tap', 'edge', function (evt) {
            callback(evt.target.data(), evt);
        });
    }

    /**
     * Register a callback for edge right-click (context tap).
     * @param {function} callback - function(edgeData, renderedPosition)
     */
    onEdgeRightClick(callback) {
        this.cy.on('cxttap', 'edge', function (evt) {
            var pos = evt.renderedPosition || evt.target.renderedMidpoint();
            callback(evt.target.data(), pos, evt);
        });
    }

    /**
     * Register a callback for background tap (deselect).
     * @param {function} callback
     */
    onBackgroundClick(callback) {
        this.cy.on('tap', function (evt) {
            if (evt.target === this.cy) {
                callback(evt);
            }
        }.bind(this));
    }

    /**
     * Get currently selected node data objects.
     * @returns {Array} array of data objects
     */
    getSelectedNodes() {
        var result = [];
        this.cy.$('node:selected').forEach(function (node) {
            result.push(node.data());
        });
        return result;
    }

    /**
     * Highlight a node by ID (add the 'highlighted' class).
     * @param {string} id - Node ID
     */
    highlightNode(id) {
        this.cy.nodes().removeClass('highlighted');
        var node = this.cy.getElementById(id);
        if (node.length > 0) {
            node.addClass('highlighted');
        }
    }

    /**
     * Clear all highlights.
     */
    clearHighlights() {
        this.cy.nodes().removeClass('highlighted');
    }

    /**
     * Center the view on a specific node.
     * @param {string} id - Node ID
     */
    centerOnNode(id) {
        var node = this.cy.getElementById(id);
        if (node.length > 0) {
            this.cy.animate({
                center: { eles: node },
                duration: 300,
            });
        }
    }

    /**
     * Fit the view to show all elements.
     */
    fit() {
        this.cy.fit(undefined, 30);
    }

    /**
     * Resize the graph (call after container size changes).
     */
    resize() {
        this.cy.resize();
        this.cy.fit(undefined, 30);
    }

    /**
     * Destroy the Cytoscape instance.
     */
    destroy() {
        if (this.cy) {
            this.cy.destroy();
            this.cy = null;
        }
    }
}
