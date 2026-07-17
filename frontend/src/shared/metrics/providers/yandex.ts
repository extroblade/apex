import type { MetricsProvider } from '../model/types';
import { noopProvider } from './noop';

// The Yandex Metrica global (`ym`). It queues calls until the tag.js loads.
interface YmFn {
  (id: number, action: string, ...args: unknown[]): void;
  a?: unknown[][];
  l?: number;
}

declare global {
  interface Window {
    ym?: YmFn;
  }
}

// loadYandex injects the Metrica tag once and installs the queueing stub, so
// events fired before the script finishes loading aren't lost.
function loadYandex(id: number): void {
  if (window.ym) return;
  const ym: YmFn = (...args: unknown[]) => {
    (ym.a ??= []).push(args);
  };
  ym.l = Date.now();
  window.ym = ym;

  const s = document.createElement('script');
  s.async = true;
  s.src = 'https://mc.yandex.ru/metrika/tag.js';
  document.head.appendChild(s);

  ym(id, 'init', { clickmap: true, trackLinks: true, accurateTrackBounce: true });
}

/**
 * Yandex Metrica provider. A blank/invalid id disables it (returns the no-op),
 * which is the current default — no id is set, so nothing loads or sends.
 */
export function createYandexProvider(counterId: string): MetricsProvider {
  const id = Number(counterId);
  if (!counterId || Number.isNaN(id)) return noopProvider;
  loadYandex(id);
  return {
    track: (event, params) => window.ym?.(id, 'reachGoal', event, params),
  };
}
