import { defineStore } from 'pinia';

export const useSessionStore = defineStore('session', {
  state: () => ({
    authenticated: false,
    user: null,
    accounts: [],
    teams: [],
    projects: [],
    roles: {},
    selectedProjectId: '',
  }),

  getters: {
    selectedProject(state) {
      return state.projects.find((project) => project.id === state.selectedProjectId) || null;
    },
  },

  actions: {
    applyAuthView(view) {
      this.authenticated = Boolean(view?.authenticated);
      this.user = view?.user || null;
      this.accounts = Array.isArray(view?.accounts) ? view.accounts : [];
      this.teams = Array.isArray(view?.teams) ? view.teams : [];
      this.projects = Array.isArray(view?.projects) ? view.projects : [];
      this.roles = view?.roles && typeof view.roles === 'object' ? view.roles : {};
      if (!this.authenticated || !this.projects.some((project) => project.id === this.selectedProjectId)) {
        this.selectedProjectId = '';
      }
    },

    selectProject(projectId) {
      const project = this.projects.find((candidate) => candidate.id === projectId);
      if (!project) {
        throw new Error(`project ${projectId} is not visible to the current user`);
      }
      this.selectedProjectId = project.id;
    },

    projectRole(projectId) {
      return this.roles[projectId] || '';
    },
  },
});
