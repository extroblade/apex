import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import { en } from './locales/en';
import { ru } from './locales/ru';

/** Locales bundled into the app (instant + offline). More can live on the
 *  static service (`/media/locales/*.json`) and are fetched on demand. */
export const LANGUAGES = ['en', 'ru'] as const;
export type Language = string;

export interface LocaleInfo {
  code: string;
  name: string;
}

const STORAGE_KEY = 'lang';
const FALLBACK_LOCALES: LocaleInfo[] = [
  { code: 'en', name: 'English' },
  { code: 'ru', name: 'Русский' },
];

function initialLanguage(): string {
  try {
    return localStorage.getItem(STORAGE_KEY) ?? 'en';
  } catch {
    return 'en';
  }
}

void i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    ru: { translation: ru },
  },
  lng: initialLanguage(),
  fallbackLng: 'en',
  interpolation: { escapeValue: false },
});

// If the persisted language isn't bundled (added later via static), load it.
if (!i18n.hasResourceBundle(i18n.language, 'translation')) {
  void setLanguage(i18n.language);
}

/**
 * Change and persist the active language, fetching the locale bundle from the
 * static service when it isn't shipped with the app.
 */
export async function setLanguage(code: string): Promise<void> {
  try {
    localStorage.setItem(STORAGE_KEY, code);
  } catch {
    // ignore persistence failures
  }
  if (!i18n.hasResourceBundle(code, 'translation')) {
    try {
      const res = await fetch(`/media/locales/${encodeURIComponent(code)}.json`);
      if (res.ok) {
        i18n.addResourceBundle(code, 'translation', await res.json());
      }
    } catch {
      // unreachable static service — i18next falls back to English keys
    }
  }
  await i18n.changeLanguage(code);
}

let manifestPromise: Promise<LocaleInfo[]> | null = null;

/**
 * The list of available locales, from the static manifest
 * (`/media/locales/index.json`) so new languages appear without an app deploy.
 */
export function fetchAvailableLocales(): Promise<LocaleInfo[]> {
  manifestPromise ??= fetch('/media/locales/index.json')
    .then((r) => (r.ok ? r.json() : Promise.reject(new Error(String(r.status)))))
    .then((data: { locales?: LocaleInfo[] }) =>
      Array.isArray(data.locales) && data.locales.length > 0
        ? data.locales
        : FALLBACK_LOCALES,
    )
    .catch(() => FALLBACK_LOCALES);
  return manifestPromise;
}

export default i18n;
