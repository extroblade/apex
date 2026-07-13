import { Sun, Moon, Palette } from 'lucide-react';

import { cn } from '@/shared/lib/utils';
import { THEMES, useThemeStore, type Theme } from '../model/theme';

const ICONS: Record<Theme, typeof Sun> = {
  light: Sun,
  dark: Moon,
  custom: Palette,
};

const LABELS: Record<Theme, string> = {
  light: 'Light theme',
  dark: 'Dark theme',
  custom: 'Custom theme',
};

/** A compact 3-way theme switcher (light / dark / custom). */
export function ThemeToggle({ className }: { className?: string }) {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);

  return (
    <div
      role="group"
      aria-label="Theme"
      className={cn('inline-flex items-center rounded-md border p-0.5', className)}
    >
      {THEMES.map((t) => {
        const Icon = ICONS[t];
        const active = theme === t;
        return (
          <button
            key={t}
            type="button"
            title={LABELS[t]}
            aria-label={LABELS[t]}
            aria-pressed={active}
            onClick={() => setTheme(t)}
            className={cn(
              'inline-flex size-7 items-center justify-center rounded-sm transition-colors',
              active
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            <Icon className="size-4" />
          </button>
        );
      })}
    </div>
  );
}
