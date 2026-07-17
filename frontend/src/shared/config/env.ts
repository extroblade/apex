/**
 * Public runtime configuration. `apiBaseUrl` is empty by default so requests
 * hit the same origin and are proxied to the backend (dev) or reverse-proxied
 * by nginx (prod). Override via the PUBLIC_API_BASE_URL build-time env var.
 *
 * Uses `import.meta.env` (always defined by rsbuild) rather than `process.env`,
 * which has no `process` global in the browser and would throw at runtime.
 */
export const env = {
  apiBaseUrl: import.meta.env.PUBLIC_API_BASE_URL ?? '',
  // Yandex Metrica counter id. Empty (the default) disables analytics — the
  // metrics layer no-ops and just logs events in dev. Set PUBLIC_YM_ID to enable.
  yandexMetricaId: import.meta.env.PUBLIC_YM_ID ?? '',
} as const;
