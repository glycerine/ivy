import { createAppServices, installAppServices } from './services/appServices.ts';
import { installIvyDiagnostics } from './services/diagnosticsService.ts';

async function bootstrap() {
  const services = installAppServices(createAppServices());
  installIvyDiagnostics({ services });
  await services.start();
}

function bootWhenReady() {
  bootstrap().catch((err) => {
    window.__ivyInitError = err && err.message ? err.message : String(err);
    console.error('Failed to boot Ivy app:', err);
  });
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bootWhenReady, { once: true });
} else {
  bootWhenReady();
}
