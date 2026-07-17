import { ThemeProvider } from '@/shared/theme';
import { MetricsProvider } from '@/shared/metrics';
import { QueryProvider } from './query-provider';

/** Composes all app-wide providers. */
export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider>
      <MetricsProvider>
        <QueryProvider>{children}</QueryProvider>
      </MetricsProvider>
    </ThemeProvider>
  );
}
