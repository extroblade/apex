import { useQuery } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import type {
  CompareDimension,
  Dashboard,
  GroupStat,
  IRacingStatus,
} from '../model/types';

export const iracingKeys = {
  status: ['iracing', 'status'] as const,
  dashboard: ['iracing', 'dashboard'] as const,
  compare: (dim: CompareDimension) => ['iracing', 'compare', dim] as const,
};

/** Whether the current user has linked an iRacing account. */
export function useIRacingStatus() {
  return useQuery({
    queryKey: iracingKeys.status,
    queryFn: () => apiFetch<IRacingStatus>('/api/iracing'),
  });
}

/** Live dashboard stats. Only enabled once an account is linked. */
export function useDashboard(enabled: boolean) {
  return useQuery({
    queryKey: iracingKeys.dashboard,
    queryFn: () => apiFetch<Dashboard>('/api/iracing/stats'),
    enabled,
    retry: false,
  });
}

/** Comparator aggregates over synced races. */
export function useComparator(dimension: CompareDimension, enabled: boolean) {
  return useQuery({
    queryKey: iracingKeys.compare(dimension),
    queryFn: () => apiFetch<GroupStat[]>(`/api/compare/${dimension}`),
    enabled,
  });
}
