import { useEffect, useState } from 'react';

import { fetchAvailableLocales, type LocaleInfo } from './config';

const FALLBACK: LocaleInfo[] = [
  { code: 'en', name: 'English' },
  { code: 'ru', name: 'Русский' },
];

/** The locales offered in the language menu, driven by the static manifest. */
export function useAvailableLocales(): LocaleInfo[] {
  const [locales, setLocales] = useState<LocaleInfo[]>(FALLBACK);

  useEffect(() => {
    let alive = true;
    void fetchAvailableLocales().then((list) => {
      if (alive) setLocales(list);
    });
    return () => {
      alive = false;
    };
  }, []);

  return locales;
}
