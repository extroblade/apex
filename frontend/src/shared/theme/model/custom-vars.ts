import type { Theme } from './theme';

/** User-tweakable variables for the `custom` theme. */
export interface CustomThemeVars {
  background: string;
  foreground: string;
  primary: string;
  accent: string;
  font: FontKey;
}

export type FontKey = 'system' | 'serif' | 'mono' | 'rounded';

export const FONT_STACKS: Record<FontKey, string> = {
  system: 'ui-sans-serif, system-ui, sans-serif',
  serif: 'Georgia, "Times New Roman", serif',
  mono: 'ui-monospace, "SF Mono", Menlo, monospace',
  rounded: '"Trebuchet MS", Verdana, ui-rounded, sans-serif',
};

export const DEFAULT_CUSTOM_VARS: CustomThemeVars = {
  background: '#17181c',
  foreground: '#f4f4f5',
  primary: '#f97316',
  accent: '#f97316',
  font: 'system',
};

const STORAGE_KEY = 'custom-theme-vars';

export function readCustomVars(): CustomThemeVars {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return { ...DEFAULT_CUSTOM_VARS, ...JSON.parse(raw) };
  } catch {
    // corrupt/unavailable storage — fall through to defaults
  }
  return DEFAULT_CUSTOM_VARS;
}

export function saveCustomVars(vars: CustomThemeVars): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(vars));
  } catch {
    // ignore persistence failures
  }
}

// CSS custom properties driven by the configurator. Surface colors reuse the
// background/foreground picks so cards and popovers stay coherent.
const VAR_MAP: Array<[cssVar: string, pick: (v: CustomThemeVars) => string]> = [
  ['--background', (v) => v.background],
  ['--card', (v) => v.background],
  ['--popover', (v) => v.background],
  ['--foreground', (v) => v.foreground],
  ['--card-foreground', (v) => v.foreground],
  ['--popover-foreground', (v) => v.foreground],
  ['--primary', (v) => v.primary],
  ['--ring', (v) => v.primary],
  ['--accent', (v) => v.accent],
  ['--app-font', (v) => FONT_STACKS[v.font] ?? FONT_STACKS.system],
];

/**
 * Applies (or clears) the stored overrides on <html>. Only the `custom` theme
 * uses them; light/dark always render their stock palettes.
 */
export function applyCustomVars(theme: Theme): void {
  const style = document.documentElement.style;
  if (theme !== 'custom') {
    for (const [cssVar] of VAR_MAP) style.removeProperty(cssVar);
    return;
  }
  const vars = readCustomVars();
  for (const [cssVar, pick] of VAR_MAP) style.setProperty(cssVar, pick(vars));
}
