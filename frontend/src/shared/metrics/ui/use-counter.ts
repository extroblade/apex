import { useContext } from 'react';

import type { CounterHelper } from '../model/types';
import { MetricsContext } from './context';

/**
 * The universal counter hook:
 *   const counter = useCounter();
 *   counter('setup_saved')({ carId });
 */
export function useCounter(): CounterHelper {
  return useContext(MetricsContext);
}
