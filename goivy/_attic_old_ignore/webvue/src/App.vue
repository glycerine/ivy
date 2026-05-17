<script setup>
import { computed } from 'vue';
import { onMounted } from 'vue';
import { ref } from 'vue';
import { useAuthClient, usePasskeyBrowser } from './app/dependencies.js';
import { useSessionStore } from './stores/sessionStore.js';

const session = useSessionStore();
const authClient = useAuthClient();
const passkeyBrowser = usePasskeyBrowser();
const currentPath = globalThis.location?.pathname || '/';

const selectedProject = computed(() => session.selectedProject);
const isVerifiedLanding = ref(currentPath === '/verified');
const isEmailContinue = ref(currentPath === '/auth/email/continue');
const emailContinueStatus = ref('');
const linkExpired = computed(() => emailContinueStatus.value === 'expired');
const showVerifiedLanding = computed(() => (isVerifiedLanding.value || emailContinueStatus.value === 'verified') && session.authenticated);
const email = ref('');
const sending = ref(false);
const sent = ref(false);
const failed = ref(false);
const isAdmin = ref(globalThis.location?.pathname?.startsWith('/admin') || false);
const authLoading = ref(!isAdmin.value);
const authLoadingMessage = ref('Loading your account...');
const authFailed = ref(false);
const adminEmails = ref([]);
const adminLoading = ref(false);
const adminFailed = ref(false);
const passkeyEnrollmentDismissed = ref(false);
const passkeyWorking = ref(false);
const passkeyFailed = ref(false);
const showPasskeyEnrollment = computed(() => {
  return session.authenticated && passkeyBrowser.isSupported() && !session.passkey.registered && !passkeyEnrollmentDismissed.value;
});

async function loadCurrentSession() {
  authLoading.value = true;
  authLoadingMessage.value = 'Loading your account...';
  authFailed.value = false;
  try {
    const view = await authClient.currentSession();
    session.applyAuthView(view);
    if (!session.authenticated) {
      await attemptPasskeyStartup();
    }
  } catch {
    authFailed.value = true;
  } finally {
    authLoading.value = false;
  }
}

async function attemptPasskeyStartup() {
  if (!passkeyBrowser.isSupported()) {
    return false;
  }
  authLoadingMessage.value = 'Checking for your passkey...';
  try {
    const options = await authClient.beginPasskeyLogin();
    const credential = await passkeyBrowser.authenticate(options);
    const view = await authClient.finishPasskeyLogin(credential);
    session.applyAuthView(view);
    return session.authenticated;
  } catch {
    return false;
  }
}

async function retryPasskeySignIn() {
  authLoading.value = true;
  authFailed.value = false;
  await attemptPasskeyStartup();
  authLoading.value = false;
}

async function createPasskey() {
  passkeyWorking.value = true;
  passkeyFailed.value = false;
  try {
    const options = await authClient.beginPasskeyRegistration();
    const credential = await passkeyBrowser.create(options);
    const view = await authClient.finishPasskeyRegistration(credential);
    session.applyAuthView(view);
    passkeyEnrollmentDismissed.value = true;
  } catch {
    passkeyFailed.value = true;
  } finally {
    passkeyWorking.value = false;
  }
}

async function requestSignupLink() {
  sending.value = true;
  failed.value = false;
  emailContinueStatus.value = '';
  try {
    await authClient.requestEmailLogin(email.value);
    sent.value = true;
  } catch {
    failed.value = true;
  } finally {
    sending.value = false;
  }
}

function readMagicLinkToken() {
  const params = new URLSearchParams(globalThis.location?.hash?.slice(1) || '');
  return params.get('token') || '';
}

function replaceBrowserPath(path) {
  if (typeof globalThis.history?.replaceState !== 'function') {
    return;
  }
  globalThis.history.replaceState({}, '', path);
}

async function continueEmailLogin() {
  authLoading.value = true;
  authLoadingMessage.value = 'Checking your sign-in link...';
  authFailed.value = false;
  emailContinueStatus.value = 'checking';
  const token = readMagicLinkToken();
  replaceBrowserPath('/auth/email/continue');
  if (!token) {
    session.applyAuthView({ authenticated: false });
    emailContinueStatus.value = 'expired';
    replaceBrowserPath('/');
    authLoading.value = false;
    return;
  }
  try {
    await authClient.consumeEmailLogin(token);
    const view = await authClient.currentSession();
    session.applyAuthView(view);
    emailContinueStatus.value = 'verified';
    replaceBrowserPath('/verified');
  } catch {
    session.applyAuthView({ authenticated: false });
    emailContinueStatus.value = 'expired';
    replaceBrowserPath('/');
  } finally {
    authLoading.value = false;
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
  if (isEmailContinue.value) {
    continueEmailLogin();
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
      <p>{{ authLoadingMessage }}</p>
    </section>

    <section v-else-if="showVerifiedLanding" aria-label="Email verified">
      <h1>Email verified</h1>
      <p>{{ session.user?.email }} is ready to use Ivy.</p>
      <section v-if="showPasskeyEnrollment" aria-label="Passkey setup">
        <h2>Use a passkey next time?</h2>
        <p>When this browser has your passkey, Ivy can open directly without stopping at email sign-in.</p>
        <button type="button" :disabled="passkeyWorking" @click="createPasskey">
          {{ passkeyWorking ? 'Creating passkey...' : 'Create passkey' }}
        </button>
        <button type="button" :disabled="passkeyWorking" @click="passkeyEnrollmentDismissed = true">
          Not now
        </button>
        <p v-if="passkeyFailed" role="alert">The passkey was not created. You can keep using email sign-in.</p>
      </section>
      <a href="/">Continue to projects</a>
    </section>

    <section v-else-if="isVerifiedLanding && authFailed" aria-label="Verification status unavailable">
      <h1>Verification status unavailable</h1>
      <p>Your browser could not load the verified account session. Try opening the sign-in link again.</p>
    </section>

    <section v-else-if="!session.authenticated" aria-label="Sign in">
      <h1>Ivy</h1>
      <p>Create your account or sign in with your verified email.</p>
      <p v-if="linkExpired && !sent" role="alert">
        That sign-in link is invalid or expired. Enter your email and we will send a new one.
      </p>
      <form v-if="!sent" @submit.prevent="requestSignupLink">
        <label>
          Email
          <input v-model="email" type="email" autocomplete="email" required />
        </label>
        <button type="submit" :disabled="sending">
          {{ sending ? 'Sending...' : 'Send sign-in link' }}
        </button>
      </form>
      <button v-if="passkeyBrowser.isSupported() && !sent" type="button" @click="retryPasskeySignIn">
        Use passkey
      </button>
      <p v-if="sent" role="status">Check your email for a sign-in link.</p>
      <p v-if="failed" role="alert">We could not send a sign-in link. Try again in a moment.</p>
    </section>

    <section v-else-if="!selectedProject" aria-label="Project picker">
      <h1>Projects</h1>
      <section v-if="showPasskeyEnrollment" aria-label="Passkey setup">
        <h2>Use a passkey next time?</h2>
        <p>When this browser has your passkey, Ivy can open directly without stopping at email sign-in.</p>
        <button type="button" :disabled="passkeyWorking" @click="createPasskey">
          {{ passkeyWorking ? 'Creating passkey...' : 'Create passkey' }}
        </button>
        <button type="button" :disabled="passkeyWorking" @click="passkeyEnrollmentDismissed = true">
          Not now
        </button>
        <p v-if="passkeyFailed" role="alert">The passkey was not created. You can keep using email sign-in.</p>
      </section>
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
