import { createContext } from 'react';

import type { CounterHelper } from '../model/types';

/** Default counter: a no-op, so calling useCounter() outside a provider is safe. */
export const noopCounter: CounterHelper = () => () => {};

export const MetricsContext = createContext<CounterHelper>(noopCounter);
