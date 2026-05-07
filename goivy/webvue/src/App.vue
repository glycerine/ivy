<script setup>
import { computed } from 'vue';
import { ref } from 'vue';
import { useAuthClient } from './app/dependencies.js';
import { useSessionStore } from './stores/sessionStore.js';

const session = useSessionStore();
const authClient = useAuthClient();

const selectedProject = computed(() => session.selectedProject);
const email = ref('');
const sending = ref(false);
const sent = ref(false);
const failed = ref(false);

async function requestSignupLink() {
  sending.value = true;
  failed.value = false;
  try {
    await authClient.requestEmailLogin(email.value);
    sent.value = true;
  } catch {
    failed.value = true;
  } finally {
    sending.value = false;
  }
}
</script>

<template>
  <main class="ivy-webvue-shell">
    <section v-if="!session.authenticated" aria-label="Sign in">
      <h1>Ivy</h1>
      <p>Create your account or sign in with your verified email.</p>
      <form v-if="!sent" @submit.prevent="requestSignupLink">
        <label>
          Email
          <input v-model="email" type="email" autocomplete="email" required />
        </label>
        <button type="submit" :disabled="sending">
          {{ sending ? 'Sending...' : 'Send sign-in link' }}
        </button>
      </form>
      <p v-if="sent" role="status">Check your email for a sign-in link.</p>
      <p v-if="failed" role="alert">We could not send a sign-in link. Try again in a moment.</p>
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
