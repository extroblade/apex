import { useQuery } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import type { DriverSearchResult, DriverProfile } from '../model/types';

export const driverKeys = {
  search: (term: string) => ['driver', 'search', term] as const,
  profile: (custId: number) => ['driver', 'profile', custId] as const,
};

/**
 * Driver search. Uses the logged-in user's own iRacing session, so it requires
 * a linked account — pass `enabled` accordingly from the page.
 */
export function useDriverSearch(term: string, enabled = true) {
  const trimmed = term.trim();
  return useQuery({
    queryKey: driverKeys.search(trimmed),
    queryFn: () =>
      apiFetch<DriverSearchResult[]>(
        `/api/drivers/search?q=${encodeURIComponent(trimmed)}`,
      ),
    enabled: enabled && trimmed.length >= 2,
  });
}

/** Driver profile by cust id (via the user's own session). */
export function useDriverProfile(custId: number, enabled = true) {
  return useQuery({
    queryKey: driverKeys.profile(custId),
    queryFn: () => apiFetch<DriverProfile>(`/api/drivers/${custId}`),
    enabled: enabled && custId > 0,
    retry: false,
  });
}
