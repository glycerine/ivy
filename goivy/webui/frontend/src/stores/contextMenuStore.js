import { defineStore } from 'pinia';

export const useContextMenuStore = defineStore('contextMenu', {
  state: () => ({
    visible: false,
    x: 0,
    y: 0,
    items: [],
  }),
  getters: {
    style: (state) => ({
      display: state.visible ? 'block' : 'none',
      left: `${state.x}px`,
      top: `${state.y}px`,
    }),
  },
  actions: {
    show(x, y, actions = []) {
      this.x = Number(x) || 0;
      this.y = Number(y) || 0;
      this.items = actions.map((action, index) => {
        if (action.separator) {
          return { kind: 'separator', key: `sep-${index}` };
        }
        if (action.header) {
          return { kind: 'header', key: `hdr-${index}`, header: action.header };
        }
        return {
          kind: 'item',
          key: `item-${index}-${action.id || action.name || ''}`,
          name: action.name || '',
          id: action.id || action.name || '',
          callback: typeof action.callback === 'function' ? action.callback : null,
        };
      });
      this.visible = true;
    },
    setPosition(x, y) {
      this.x = Number(x) || 0;
      this.y = Number(y) || 0;
    },
    hide() {
      this.visible = false;
      this.items = [];
    },
    runItem(item) {
      this.hide();
      if (item && typeof item.callback === 'function') {
        item.callback();
      }
    },
  },
});
