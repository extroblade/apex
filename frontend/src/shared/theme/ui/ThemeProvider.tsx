import { useEffect } from 'react';

import { applyTheme, useThemeStore } from '../model/theme';
import { applyCustomVars } from '../model/custom-vars';

/** Keeps the <html> theme classes + custom-theme overrides in sync. */
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const theme = useThemeStore((s) => s.theme);

  useEffect(() => {
    applyTheme(theme);
    applyCustomVars(theme);
  }, [theme]);

  return <>{children}</>;
}
