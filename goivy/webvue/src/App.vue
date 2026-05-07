<script setup>
import { computed } from 'vue';
import { onMounted } from 'vue';
import { ref } from 'vue';
import { useAuthClient } from './app/dependencies.js';
import { useSessionStore } from './stores/sessionStore.js';

const session = useSessionStore();
const authClient = useAuthClient();

const selectedProject = computed(() => session.selectedProject);
const isVerifiedLanding = ref(globalThis.location?.pathname === '/verified');
const email = ref('');
const sending = ref(false);
const sent = ref(false);
const failed = ref(false);
const authLoading = ref(false);
const authFailed = ref(false);
const isAdmin = ref(globalThis.location?.pathname?.startsWith('/admin') || false);
const adminEmails = ref([]);
const adminLoading = ref(false);
const adminFailed = ref(false);

async function loadCurrentSession() {
  authLoading.value = true;
  authFailed.value = false;
  try {
    const view = await authClient.currentSession();
    session.applyAuthView(view);
  } catch {
    authFailed.value = true;
  } finally {
    authLoading.value = false;
  }
}

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

async function loadAdminEmails() {
  adminLoading.value = true;
  adminFailed.value = false;
  try {
    const response = await authClient.listUnverifiedEmails();
    adminEmails.value = response.emails || [];
  } catch {
    adminFailed.value = true;
  } finally {
    adminLoading.value = false;
  }
}

function emailStatus(row) {
  if (row.usedAt) {
    return 'used';
  }
  if (row.expired) {
    return 'expired';
  }
  return 'pending';
}

onMounted(() => {
  if (isAdmin.value) {
    loadAdminEmails();
    return;
  }
  loadCurrentSession();
});
</script>

<template>
  <main class="ivy-webvue-shell">
    <section v-if="isAdmin" aria-label="Admin dashboard">
      <h1>Ivy admin</h1>
      <section aria-label="Unverified emails">
        <header>
          <h2>Unverified emails</h2>
          <button type="button" :disabled="adminLoading" @click="loadAdminEmails">
            {{ adminLoading ? 'Refreshing...' : 'Refresh' }}
          </button>
        </header>
        <p v-if="adminFailed" role="alert">Could not load unverified emails.</p>
        <p v-else-if="!adminLoading && adminEmails.length === 0">No unverified email links.</p>
        <table v-else>
          <thead>
            <tr>
              <th>Email</th>
              <th>Status</th>
              <th>Expires</th>
              <th>Link</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in adminEmails" :key="row.loginUrl">
              <td>{{ row.email }}</td>
              <td>{{ emailStatus(row) }}</td>
              <td>{{ new Date(row.expiresAt).toLocaleString() }}</td>
              <td><a :href="row.loginUrl">Open sign-in link</a></td>
            </tr>
          </tbody>
        </table>
      </section>
    </section>

    <section v-else-if="authLoading" aria-label="Loading session">
      <h1>Ivy</h1>
      <p>Loading your account...</p>
    </section>

    <section v-else-if="isVerifiedLanding && session.authenticated" aria-label="Email verified">
      <h1>Email verified</h1>
      <p>{{ session.user?.email }} is ready to use Ivy.</p>
      <a href="/">Continue to projects</a>
    </section>

    <section v-else-if="isVerifiedLanding && authFailed" aria-label="Verification status unavailable">
      <h1>Verification status unavailable</h1>
      <p>Your browser could not load the verified account session. Try opening the sign-in link again.</p>
    </section>

    <section v-else-if="!session.authenticated" aria-label="Sign in">
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
