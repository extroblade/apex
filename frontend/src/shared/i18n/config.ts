import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import { en } from './locales/en';

/**
 * `en` is the only bundled locale: it is the instant, offline default AND the
 * source of the `Translation` type. Every other language — including `ru` — is
 * DATA served by the backend (GET /api/locales for the list, GET
 * /api/locales/{code} for a bundle), so adding a language needs no app rebuild.
 * `ru.ts` still lives in the tree as the authored, type-checked source that
 * `npm run gen:locales` exports to the backend; it is just not bundled here.
 */
export const BUNDLED = 'en';
export type Language = string;

export interface LocaleInfo {
  code: string;
  name: string;
}

const STORAGE_KEY = 'lang';
// Shown only if the backend list can't be reached — the real list is /api/locales.
const FALLBACK_LOCALES: LocaleInfo[] = [
  { code: 'en', name: 'English' },
  { code: 'ru', name: 'Русский' },
];

function initialLanguage(): string {
  try {
    return localStorage.getItem(STORAGE_KEY) ?? BUNDLED;
  } catch {
    return BUNDLED;
  }
}

void i18n.use(initReactI18next).init({
  resources: { en: { translation: en } },
  lng: initialLanguage(),
  fallbackLng: BUNDLED,
  interpolation: { escapeValue: false },
});

// If the persisted language isn't the bundled one, fetch its bundle now.
if (i18n.language !== BUNDLED && !i18n.hasResourceBundle(i18n.language, 'translation')) {
  void setLanguage(i18n.language);
}

/**
 * Change and persist the active language, fetching its bundle from the backend
 * when it isn't already loaded. Missing keys fall back to English (i18next
 * fallbackLng), so a partial or unreachable bundle degrades gracefully.
 */
export async function setLanguage(code: string): Promise<void> {
  try {
    localStorage.setItem(STORAGE_KEY, code);
  } catch {
    // ignore persistence failures
  }
  if (!i18n.hasResourceBundle(code, 'translation')) {
    try {
      const res = await fetch(`/api/locales/${encodeURIComponent(code)}`);
      if (res.ok) {
        i18n.addResourceBundle(code, 'translation', await res.json());
      }
    } catch {
      // backend unreachable — i18next falls back to English keys
    }
  }
  await i18n.changeLanguage(code);
}

let listPromise: Promise<LocaleInfo[]> | null = null;

/**
 * The languages offered in the menu, from the backend (GET /api/locales) so new
 * languages appear without an app deploy. Cached for the session.
 */
export function fetchAvailableLocales(): Promise<LocaleInfo[]> {
  listPromise ??= fetch('/api/locales')
    .then((r) => (r.ok ? r.json() : Promise.reject(new Error(String(r.status)))))
    .then((data: { locales?: LocaleInfo[] }) =>
      Array.isArray(data.locales) && data.locales.length > 0
        ? data.locales
        : FALLBACK_LOCALES,
    )
    .catch(() => FALLBACK_LOCALES);
  return listPromise;
}

export default i18n;
