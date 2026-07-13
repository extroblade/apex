import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import type { Setup, NewSetup } from '../model/types';

export const setupsKeys = {
  all: ['setups'] as const,
  list: (mine: boolean) => ['setups', 'list', mine] as const,
  one: (id: number) => ['setups', 'one', id] as const,
};

/** The showroom (public + own) or, with mine=true, only the caller's setups. */
export function useSetups(mine = false) {
  return useQuery({
    queryKey: setupsKeys.list(mine),
    queryFn: () => apiFetch<Setup[]>(`/api/setups/${mine ? '?mine=1' : ''}`),
  });
}

/** One setup with its data; pass download to bump the counter server-side. */
export function useSetup(id: number | null, download = false) {
  return useQuery({
    queryKey: setupsKeys.one(id ?? 0),
    queryFn: () => apiFetch<Setup>(`/api/setups/${id}${download ? '?download=1' : ''}`),
    enabled: id != null,
  });
}

export interface GeneratedSetup {
  name: string;
  notes: string;
  data: string;
}

/** Ask the backend for a generated baseline for a car+track combo. */
export function useGenerateSetup() {
  return useMutation({
    mutationFn: (input: { carId: number; trackId: number }) =>
      apiFetch<GeneratedSetup>('/api/setups/generate', {
        method: 'POST',
        body: JSON.stringify(input),
      }),
  });
}

export function useCreateSetup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: NewSetup) =>
      apiFetch<Setup>('/api/setups/', {
        method: 'POST',
        body: JSON.stringify(input),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: setupsKeys.all }),
  });
}

export function useSetSetupPublic() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { id: number; public: boolean }) =>
      apiFetch<void>(`/api/setups/${input.id}/public`, {
        method: 'PUT',
        body: JSON.stringify({ public: input.public }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: setupsKeys.all }),
  });
}

export function useDeleteSetup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => apiFetch<void>(`/api/setups/${id}`, { method: 'DELETE' }),
    onSuccess: () => qc.invalidateQueries({ queryKey: setupsKeys.all }),
  });
}
