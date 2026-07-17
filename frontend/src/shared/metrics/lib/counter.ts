import type { CounterHelper, MetricsProvider } from '../model/types';

/**
 * Builds the curried counter over a provider:
 *   const counter = makeCounterHelper(provider);
 *   counter('setup_saved')({ carId });
 * Separated from the provider/hook so it's trivially unit-testable with a fake.
 */
export function makeCounterHelper(provider: MetricsProvider): CounterHelper {
  return (event) => (params) => provider.track(event, params);
}
