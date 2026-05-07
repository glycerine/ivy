<script setup>
import { computed } from 'vue';
import { useSessionStore } from './stores/sessionStore.js';

const session = useSessionStore();

const selectedProject = computed(() => session.selectedProject);
</script>

<template>
  <main class="ivy-webvue-shell">
    <section v-if="!session.authenticated" aria-label="Sign in">
      <h1>Ivy</h1>
      <p>Sign in with your email to open your projects.</p>
      <form>
        <label>
          Email
          <input type="email" autocomplete="email" />
        </label>
        <button type="submit">Send sign-in link</button>
      </form>
    </section>

    <section v-else-if="!selectedProject" aria-label="Project picker">
      <h1>Projects</h1>
      <ul>
        <li v-for="project in session.projects" :key="project.id">
          <button type="button" @click="session.selectProject(project.id)">
            {{ project.displayName }}
          </button>
        </li>
      </ul>
    </section>

    <section v-else aria-label="Ivy workspace">
      <header>
        <strong>{{ selectedProject.displayName }}</strong>
        <span>{{ session.projectRole(selectedProject.id) }}</span>
      </header>
      <div class="workspace-grid">
        <section aria-label="ARG (Abstract Reachability Graph)">ARG</section>
        <section aria-label="Concept graph">Concept graph</section>
        <section aria-label="State/relations">State/relations</section>
        <section aria-label="Editing">Editing</section>
      </div>
    </section>
  </main>
</template>
