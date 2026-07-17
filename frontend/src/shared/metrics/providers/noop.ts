import { devlog } from '@/shared/lib/dev';

import type { MetricsProvider } from '../model/types';

/**
 * Used when analytics are disabled (e.g. no Yandex id). It still surfaces events
 * through devlog, so with the developer cookie set you can watch what fires
 * without sending anything anywhere.
 */
export const noopProvider: MetricsProvider = {
  track: (event, params) => devlog('[metric]', event, params ?? {}),
};
