import { useMutation, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import {
  plannerKeys,
  type CatalogCounts,
  type CarItem,
  type TrackItem,
  type SeriesItem,
} from '@/entities/planner';

export function useSyncCatalog() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      apiFetch<CatalogCounts>('/api/planner/catalog/sync', { method: 'POST' }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['planner'] }),
  });
}

// Toggles use optimistic updates so the checkbox flips instantly (no wait for
// the round-trip), then reconcile with the server on settle — reverting on error.
export function useSetCarOwned() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ carId, owned }: { carId: number; owned: boolean }) =>
      apiFetch<void>(`/api/planner/cars/${carId}`, {
        method: 'PUT',
        body: JSON.stringify({ owned }),
      }),
    onMutate: async ({ carId, owned }) => {
      await qc.cancelQueries({ queryKey: plannerKeys.cars });
      const prev = qc.getQueryData<CarItem[]>(plannerKeys.cars);
      qc.setQueryData<CarItem[]>(plannerKeys.cars, (old) =>
        old?.map((c) => (c.carId === carId ? { ...c, owned } : c)),
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(plannerKeys.cars, ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: plannerKeys.cars }),
  });
}

export function useSetTrackOwned() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ trackId, owned }: { trackId: number; owned: boolean }) =>
      apiFetch<void>(`/api/planner/tracks/${trackId}`, {
        method: 'PUT',
        body: JSON.stringify({ owned }),
      }),
    onMutate: async ({ trackId, owned }) => {
      await qc.cancelQueries({ queryKey: plannerKeys.tracks });
      const prev = qc.getQueryData<TrackItem[]>(plannerKeys.tracks);
      qc.setQueryData<TrackItem[]>(plannerKeys.tracks, (old) =>
        old?.map((t) => (t.trackId === trackId ? { ...t, owned } : t)),
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(plannerKeys.tracks, ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: plannerKeys.tracks }),
  });
}

export function useSetSeriesFavorite() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ seriesId, favorite }: { seriesId: number; favorite: boolean }) =>
      apiFetch<void>(`/api/planner/series/${seriesId}`, {
        method: 'PUT',
        body: JSON.stringify({ favorite }),
      }),
    onMutate: async ({ seriesId, favorite }) => {
      await qc.cancelQueries({ queryKey: plannerKeys.series });
      const prev = qc.getQueryData<SeriesItem[]>(plannerKeys.series);
      qc.setQueryData<SeriesItem[]>(plannerKeys.series, (old) =>
        old?.map((s) => (s.seriesId === seriesId ? { ...s, favorite } : s)),
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(plannerKeys.series, ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: plannerKeys.series }),
  });
}
