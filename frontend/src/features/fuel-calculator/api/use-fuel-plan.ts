import { useMutation } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import type { FuelRequest, FuelPlan } from '../model/types';

/**
 * A React Query *mutation* (not a query) because computing a plan is an action
 * triggered on submit, not data we read/cache by key.
 */
export function useFuelPlan() {
  return useMutation({
    mutationFn: (req: FuelRequest) =>
      apiFetch<FuelPlan>('/api/fuel/plan', {
        method: 'POST',
        body: JSON.stringify(req),
      }),
  });
}
