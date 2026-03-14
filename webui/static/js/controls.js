/**
 * IvyControls - UI control management for the Ivy verification UI.
 *
 * Manages edge/label visibility toggles, context menus, info panel,
 * and status bar.
 */

'use strict';

/**
 * Edge display class labels for toggle checkboxes.
 * Maps internal class names to user-visible symbols.
 */
var EDGE_DISPLAY_SYMBOLS = {
    'all_to_all': '+',
    'edge_unknown': '?',
    'none_to_none': '\u2013', // en-dash for minus
    'transitive': 'T',
};

/**
 * Node label display class labels.
 */
var LABEL_DISPLAY_SYMBOLS = {
    'node_necessarily': '+',
    'node_maybe': '?',
    'node_necessarily_not': '\u00AC', // negation sign
};


class IvyControls {
    /**
     * @param {IvyAPI} api - API instance for server communication
     */
    constructor(api) {
        this.api = api;
        this.edgeToggles = {};   // { edgeName: { className: checkbox } }
        this.labelToggles = {};  // { labelName: { className: checkbox } }
        this._contextMenuVisible = false;
    }

    /**
     * Build edge visibility toggle checkboxes from edge names.
     * Creates a checkbox group for each edge with +, ?, - and T (transitive) toggles.
     *
     * @param {Array<string>} edgeNames - List of edge concept names
     * @param {function} onChange - Callback when any toggle changes: onChange(edgeName, className, checked)
     */
    buildEdgeToggles(edgeNames, onChange) {
        var container = document.getElementById('edge-toggles');
        container.innerHTML = '';
        this.edgeToggles = {};

        for (var i = 0; i < edgeNames.length; i++) {
            var edgeName = edgeNames[i];
            var group = document.createElement('div');
            group.className = 'toggle-group';

            var groupLabel = document.createElement('span');
            groupLabel.className = 'toggle-group-label';
            groupLabel.textContent = edgeName;
            group.appendChild(groupLabel);

            this.edgeToggles[edgeName] = {};

            var classes = ['all_to_all', 'edge_unknown', 'none_to_none', 'transitive'];
            for (var j = 0; j < classes.length; j++) {
                var cls = classes[j];
                var item = this._createToggle(
                    edgeName + '_' + cls,
                    EDGE_DISPLAY_SYMBOLS[cls],
                    cls !== 'none_to_none', // default: show + and ?, hide -
                    edgeName,
                    cls,
                    onChange
                );
                group.appendChild(item.element);
                this.edgeToggles[edgeName][cls] = item.checkbox;
            }

            container.appendChild(group);
        }
    }

    /**
     * Build node label visibility toggle checkboxes.
     * Creates a checkbox group for each label with +, ?, and negation toggles.
     *
     * @param {Array<string>} labelNames - List of node label names
     * @param {function} onChange - Callback: onChange(labelName, className, checked)
     */
    buildLabelToggles(labelNames, onChange) {
        var container = document.getElementById('label-toggles');
        container.innerHTML = '';
        this.labelToggles = {};

        for (var i = 0; i < labelNames.length; i++) {
            var labelName = labelNames[i];
            var group = document.createElement('div');
            group.className = 'toggle-group';

            var groupLabel = document.createElement('span');
            groupLabel.className = 'toggle-group-label';
            groupLabel.textContent = labelName;
            group.appendChild(groupLabel);

            this.labelToggles[labelName] = {};

            var classes = ['node_necessarily', 'node_maybe', 'node_necessarily_not'];
            for (var j = 0; j < classes.length; j++) {
                var cls = classes[j];
                var item = this._createToggle(
                    'label_' + labelName + '_' + cls,
                    LABEL_DISPLAY_SYMBOLS[cls],
                    true, // default: show all labels
                    labelName,
                    cls,
                    onChange
                );
                group.appendChild(item.element);
                this.labelToggles[labelName][cls] = item.checkbox;
            }

            container.appendChild(group);
        }
    }

    /**
     * Create a single toggle checkbox item.
     * @private
     */
    _createToggle(id, symbol, defaultChecked, groupName, className, onChange) {
        var item = document.createElement('span');
        item.className = 'toggle-item';

        var checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.id = id;
        checkbox.checked = defaultChecked;
        checkbox.addEventListener('change', function () {
            if (onChange) {
                onChange(groupName, className, checkbox.checked);
            }
        });

        var label = document.createElement('label');
        label.htmlFor = id;
        label.textContent = symbol;
        label.title = className.replace(/_/g, ' ');

        item.appendChild(checkbox);
        item.appendChild(label);

        return { element: item, checkbox: checkbox };
    }

    /**
     * Get the current state of all edge toggles.
     * @returns {object} { edgeName: { className: boolean } }
     */
    getEdgeToggleState() {
        var state = {};
        for (var edge in this.edgeToggles) {
            if (this.edgeToggles.hasOwnProperty(edge)) {
                state[edge] = {};
                for (var cls in this.edgeToggles[edge]) {
                    if (this.edgeToggles[edge].hasOwnProperty(cls)) {
                        state[edge][cls] = this.edgeToggles[edge][cls].checked;
                    }
                }
            }
        }
        return state;
    }

    /**
     * Get the current state of all label toggles.
     * @returns {object} { labelName: { className: boolean } }
     */
    getLabelToggleState() {
        var state = {};
        for (var label in this.labelToggles) {
            if (this.labelToggles.hasOwnProperty(label)) {
                state[label] = {};
                for (var cls in this.labelToggles[label]) {
                    if (this.labelToggles[label].hasOwnProperty(cls)) {
                        state[label][cls] = this.labelToggles[label][cls].checked;
                    }
                }
            }
        }
        return state;
    }

    /**
     * Show a context menu at the given position.
     *
     * @param {number} x - Rendered X position (pixels)
     * @param {number} y - Rendered Y position (pixels)
     * @param {Array} actions - Array of action objects:
     *   { name: string, id: string, callback: function }
     *   or { separator: true } for a separator
     *   or { header: string } for a section header
     */
    showContextMenu(x, y, actions) {
        var menu = document.getElementById('context-menu');
        menu.innerHTML = '';
        menu.style.display = 'block';

        for (var i = 0; i < actions.length; i++) {
            var action = actions[i];

            if (action.separator) {
                var sep = document.createElement('div');
                sep.className = 'context-menu-separator';
                menu.appendChild(sep);
                continue;
            }

            if (action.header) {
                var hdr = document.createElement('div');
                hdr.className = 'context-menu-header';
                hdr.textContent = action.header;
                menu.appendChild(hdr);
                continue;
            }

            var item = document.createElement('div');
            item.className = 'context-menu-item';
            item.textContent = action.name;
            item.dataset.actionId = action.id || action.name;
            (function (act) {
                item.addEventListener('click', function (e) {
                    e.stopPropagation();
                    menu.style.display = 'none';
                    if (act.callback) {
                        act.callback();
                    }
                });
            })(action);
            menu.appendChild(item);
        }

        // Position the menu, ensuring it stays within the viewport
        var menuRect = menu.getBoundingClientRect();
        var viewW = window.innerWidth;
        var viewH = window.innerHeight;

        // Temporarily show to measure
        menu.style.left = x + 'px';
        menu.style.top = y + 'px';

        // After render, adjust if overflowing
        requestAnimationFrame(function () {
            var rect = menu.getBoundingClientRect();
            if (rect.right > viewW) {
                menu.style.left = (x - rect.width) + 'px';
            }
            if (rect.bottom > viewH) {
                menu.style.top = (y - rect.height) + 'px';
            }
        });

        this._contextMenuVisible = true;
    }

    /**
     * Hide the context menu.
     */
    hideContextMenu() {
        var menu = document.getElementById('context-menu');
        menu.style.display = 'none';
        menu.innerHTML = '';
        this._contextMenuVisible = false;
    }

    /**
     * Check if context menu is currently visible.
     * @returns {boolean}
     */
    isContextMenuVisible() {
        return this._contextMenuVisible;
    }

    /**
     * Display information in the bottom info panel.
     *
     * @param {string} shortInfo - Brief summary line
     * @param {string|Array<string>} longInfo - Detailed info (string or array of lines)
     */
    showInfo(shortInfo, longInfo) {
        var content = document.getElementById('info-content');
        var lines = [];

        if (shortInfo) {
            lines.push(shortInfo);
        }

        if (longInfo) {
            if (Array.isArray(longInfo)) {
                lines = lines.concat(longInfo);
            } else {
                lines.push(longInfo);
            }
        }

        content.textContent = lines.join('\n');
    }

    /**
     * Clear the info panel to its default state.
     */
    clearInfo() {
        document.getElementById('info-content').textContent = 'Select a node or edge to see details';
    }

    /**
     * Update the status bar message.
     * @param {string} msg - Status message
     * @param {string} [level] - 'normal', 'error', or 'success'
     */
    setStatus(msg, level) {
        var bar = document.getElementById('statusbar');
        bar.textContent = msg;
        bar.className = '';
        if (level === 'error') {
            bar.classList.add('error');
        } else if (level === 'success') {
            bar.classList.add('success');
        }
    }

    /**
     * Show the loading overlay.
     * @param {string} [message] - Loading message
     */
    showLoading(message) {
        var overlay = document.getElementById('loading-overlay');
        var msgEl = document.getElementById('loading-message');
        msgEl.textContent = message || 'Loading...';
        overlay.style.display = 'flex';
    }

    /**
     * Hide the loading overlay.
     */
    hideLoading() {
        document.getElementById('loading-overlay').style.display = 'none';
    }
}
