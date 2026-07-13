import { useQuery } from '@tanstack/react-query';

import { apiFetch, ApiError } from '@/shared/api/client';
import type { Viewer } from '../model/types';

export const viewerKeys = {
  me: ['viewer', 'me'] as const,
};

/** Fetches the current user; a 401 means "not logged in" and yields null. */
export async function fetchViewer(): Promise<Viewer | null> {
  try {
    return await apiFetch<Viewer>('/api/auth/me');
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) return null;
    throw e;
  }
}

/** Reads the current user from GET /api/auth/me. */
export function useViewer() {
  return useQuery({
    queryKey: viewerKeys.me,
    queryFn: fetchViewer,
    staleTime: 60_000,
  });
}
