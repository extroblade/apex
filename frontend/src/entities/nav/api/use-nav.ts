import { useQuery } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';

/** One menu entry as served by the nav service (GET /api/nav). */
export interface NavItem {
  key: string;
  /** i18n key (e.g. "nav.planner") — the service ships keys, not translations. */
  labelKey: string;
  href: string;
  /** lucide icon name, resolved through a whitelist (see NavIcon). */
  icon: string;
  /** Where the item may appear: "side" and/or "bottom". */
  placements: string[];
  order: number;
  requiresAuth: boolean;
  featureFlag?: string;
}

export const navKeys = { all: ['nav'] as const };

export function fetchNav(): Promise<NavItem[]> {
  return apiFetch<{ items?: NavItem[] }>('/api/nav').then((r) => r.items ?? []);
}

/** Loads the backend-configured menu. It changes rarely, so cache it. */
export function useNav() {
  return useQuery({
    queryKey: navKeys.all,
    queryFn: fetchNav,
    staleTime: 5 * 60_000,
  });
}
