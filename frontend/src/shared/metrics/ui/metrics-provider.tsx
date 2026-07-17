import { useMemo } from 'react';

import { makeCounterHelper } from '../lib/counter';
import { createProvider } from '../model/provider-factory';
import type { MetricsProvider as Provider } from '../model/types';
import { MetricsContext } from './context';

/**
 * Provides the counter to the tree. Builds it once from the configured provider
 * (Yandex by default). Pass `provider` to inject a fake in tests/stories.
 */
export function MetricsProvider({
  children,
  provider,
}: {
  children: React.ReactNode;
  provider?: Provider;
}) {
  const counter = useMemo(
    () => makeCounterHelper(provider ?? createProvider()),
    [provider],
  );
  return <MetricsContext.Provider value={counter}>{children}</MetricsContext.Provider>;
}
