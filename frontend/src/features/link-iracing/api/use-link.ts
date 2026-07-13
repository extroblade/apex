import { useMutation, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import { iracingKeys } from '@/entities/iracing';

/**
 * Linking is an OAuth redirect, not an API call: send the browser to the
 * backend's /authorize endpoint, which redirects on to iRacing's login page.
 */
export function startIRacingLink() {
  window.location.href = '/api/iracing/authorize';
}

export function useUnlinkIRacing() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch<void>('/api/iracing', { method: 'DELETE' }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: iracingKeys.status });
      qc.removeQueries({ queryKey: iracingKeys.dashboard });
    },
  });
}

export function useSyncIRacing() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      apiFetch<{ synced: number }>('/api/iracing/sync', { method: 'POST' }),
    onSuccess: () => {
      // Synced data changes the comparators.
      qc.invalidateQueries({ queryKey: ['iracing', 'compare'] });
    },
  });
}
