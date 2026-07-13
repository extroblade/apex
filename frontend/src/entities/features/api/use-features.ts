import { useQuery } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';

export type Features = Record<string, boolean>;

export const featureKeys = { all: ['features'] as const };

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

export const IRACING_OAUTH = 'iracing_oauth';
