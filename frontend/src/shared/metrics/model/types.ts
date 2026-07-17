import type { MetricEvent } from './events';

export type MetricParams = Record<string, unknown>;

/**
 * A metrics backend. Implement `track` to add a provider (Yandex Metrica by
 * default; GA, PostHog, a custom endpoint… are one file each).
 */
export interface MetricsProvider {
  track(event: string, params?: MetricParams): void;
}

/**
 * The universal counter. Curried so a call site reads:
 *   const counter = useCounter();
 *   counter('setup_saved')({ carId });
 */
export type CounterHelper = (event: MetricEvent) => (params?: MetricParams) => void;
