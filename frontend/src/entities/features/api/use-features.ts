import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';

export type Features = Record<string, boolean>;

/** Cockpit health summary from GET /api/health/cockpit. */
export interface CockpitHealth {
  db: boolean;
  redisEnabled: boolean;
  redis: boolean;
}

export const featureKeys = {
  all: ['features'] as const,
  cockpit: ['features', 'all'] as const,
  health: ['features', 'health'] as const,
};

/** Fetches the backend feature flags. */
export function fetchFeatures(): Promise<Features> {
  return apiFetch<Features>('/api/features');
}

/** Loads the backend feature flags (cached; they rarely change). */
export function useFeatures() {
  return useQuery({
    queryKey: featureKeys.all,
    queryFn: fetchFeatures,
    staleTime: 5 * 60_000,
  });
}

/** Convenience: whether a single flag is on. */
export function useFeature(key: string): boolean {
  const { data } = useFeatures();
  return Boolean(data?.[key]);
}

/** Cockpit: fetches the full feature list (requires developer cookie). */
export function fetchAllFeatures(): Promise<Features> {
  return apiFetch<Features>('/api/features/all');
}

/** Cockpit: loads the full feature list. Enable only when the overlay is open. */
export function useAllFeatures(enabled: boolean) {
  return useQuery({
    queryKey: featureKeys.cockpit,
    queryFn: fetchAllFeatures,
    enabled,
    staleTime: 0,
  });
}

/** Cockpit: loads the backend health summary (requires developer cookie). */
export function useCockpitHealth(enabled: boolean) {
  return useQuery({
    queryKey: featureKeys.health,
    queryFn: () => apiFetch<CockpitHealth>('/api/health/cockpit'),
    enabled,
    staleTime: 0,
  });
}

/** Cockpit: toggles a single feature flag (requires developer cookie). */
export function toggleFeature(key: string, enabled: boolean): Promise<void> {
  return apiFetch<void>('/api/features/' + key, {
    method: 'PUT',
    body: JSON.stringify({ enabled }),
  });
}

/** Cockpit: mutation hook for toggling a flag. Invalidates the features cache. */
export function useToggleFeature() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ key, enabled }: { key: string; enabled: boolean }) =>
      toggleFeature(key, enabled),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: featureKeys.all });
    },
  });
}

export const IRACING_OAUTH = 'iracing_oauth';
export const COCKPIT = 'cockpit';
