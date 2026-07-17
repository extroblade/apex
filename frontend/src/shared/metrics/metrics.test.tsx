import type { ReactNode } from 'react';
import { describe, it, expect, vi } from 'vitest';
import { renderHook } from '@testing-library/react';

import { makeCounterHelper } from './lib/counter';
import { noopProvider } from './providers/noop';
import { createYandexProvider } from './providers/yandex';
import { MetricsProvider } from './ui/metrics-provider';
import { useCounter } from './ui/use-counter';
import type { MetricsProvider as Provider } from './model/types';

describe('makeCounterHelper', () => {
  it('is curried and forwards event + params to the provider', () => {
    const track = vi.fn();
    const counter = makeCounterHelper({ track });
    counter('setup_saved')({ carId: 3 });
    expect(track).toHaveBeenCalledWith('setup_saved', { carId: 3 });
  });

  it('works without params', () => {
    const track = vi.fn();
    makeCounterHelper({ track })('language_changed')();
    expect(track).toHaveBeenCalledWith('language_changed', undefined);
  });

  it('returns a reusable event fn (call the same event many times)', () => {
    const track = vi.fn();
    const saved = makeCounterHelper({ track })('setup_saved');
    saved({ id: 1 });
    saved({ id: 2 });
    expect(track).toHaveBeenNthCalledWith(1, 'setup_saved', { id: 1 });
    expect(track).toHaveBeenNthCalledWith(2, 'setup_saved', { id: 2 });
  });
});

describe('providers', () => {
  it('noop provider never throws', () => {
    expect(() => noopProvider.track('x', { a: 1 })).not.toThrow();
  });

  it('yandex provider with a blank id is disabled (no tag loaded, no throw)', () => {
    const p = createYandexProvider('');
    expect(() => p.track('x')).not.toThrow();
    expect(window.ym).toBeUndefined();
  });
});

describe('useCounter', () => {
  it('returns a counter wired to the injected provider', () => {
    const track = vi.fn();
    const provider: Provider = { track };
    const wrapper = ({ children }: { children: ReactNode }) => (
      <MetricsProvider provider={provider}>{children}</MetricsProvider>
    );
    const { result } = renderHook(() => useCounter(), { wrapper });
    result.current('goal_created')({ id: 1 });
    expect(track).toHaveBeenCalledWith('goal_created', { id: 1 });
  });
});
