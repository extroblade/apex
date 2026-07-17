import { env } from '@/shared/config/env';

import { createYandexProvider } from '../providers/yandex';
import type { MetricsProvider } from './types';

/**
 * The active metrics provider. Yandex Metrica when PUBLIC_YM_ID is set,
 * otherwise a dev no-op. Swap the default here to change analytics backends.
 */
export function createProvider(): MetricsProvider {
  return createYandexProvider(env.yandexMetricaId);
}
