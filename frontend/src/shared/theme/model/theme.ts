import { create } from 'zustand';

export type Theme = 'light' | 'dark' | 'custom';

export const THEMES: readonly Theme[] = ['light', 'dark', 'custom'];

const STORAGE_KEY = 'theme';

function isTheme(v: unknown): v is Theme {
  return v === 'light' || v === 'dark' || v === 'custom';
}

/** Applies a theme by toggling classes on <html>. `light` uses the :root vars. */
export function applyTheme(theme: Theme): void {
  const el = document.documentElement;
  el.classList.remove('dark', 'custom');
  if (theme === 'dark') el.classList.add('dark');
  else if (theme === 'custom') el.classList.add('custom');
}

/** Reads the persisted theme (also set pre-hydration by an inline script). */
export function readStoredTheme(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (isTheme(stored)) return stored;
  } catch {
    // localStorage unavailable — fall through to default
  }
  return 'light';
}

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

export const useThemeStore = create<ThemeState>((set) => ({
  theme: readStoredTheme(),
  setTheme: (theme) => {
    try {
      localStorage.setItem(STORAGE_KEY, theme);
    } catch {
      // ignore persistence failures
    }
    applyTheme(theme);
    set({ theme });
  },
}));
