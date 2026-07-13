import { ThemeProvider } from '@/shared/theme';
import { QueryProvider } from './query-provider';

/** Composes all app-wide providers. */
export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider>
      <QueryProvider>{children}</QueryProvider>
    </ThemeProvider>
  );
}
