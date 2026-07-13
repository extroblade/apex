import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import type { CarItem, TrackItem, SeriesItem, Season } from '../model/types';

export const plannerKeys = {
  cars: ['planner', 'cars'] as const,
  tracks: ['planner', 'tracks'] as const,
  series: ['planner', 'series'] as const,
  season: ['planner', 'season'] as const,
};

export function useCars(enabled = true) {
  return useQuery({
    queryKey: plannerKeys.cars,
    queryFn: () => apiFetch<CarItem[]>('/api/planner/cars'),
    enabled,
  });
}

export function useTracks(enabled = true) {
  return useQuery({
    queryKey: plannerKeys.tracks,
    queryFn: () => apiFetch<TrackItem[]>('/api/planner/tracks'),
    enabled,
  });
}

export function useSeriesList(enabled = true) {
  return useQuery({
    queryKey: plannerKeys.series,
    queryFn: () => apiFetch<SeriesItem[]>('/api/planner/series'),
    enabled,
  });
}

export function useSeason(enabled = true) {
  return useQuery({
    queryKey: plannerKeys.season,
    queryFn: () => apiFetch<Season>('/api/planner/season'),
    enabled,
  });
}

/** Toggle a (series, week) race in the plan; optimistic so cells flip instantly. */
export function useSetRacePlanned() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { seriesId: number; week: number; planned: boolean }) =>
      apiFetch<void>('/api/planner/season/plan', {
        method: 'PUT',
        body: JSON.stringify(input),
      }),
    onMutate: async ({ seriesId, week, planned }) => {
      await qc.cancelQueries({ queryKey: plannerKeys.season });
      const prev = qc.getQueryData<Season>(plannerKeys.season);
      qc.setQueryData<Season>(plannerKeys.season, (old) =>
        old
          ? {
              ...old,
              series: old.series.map((s) =>
                s.seriesId === seriesId
                  ? {
                      ...s,
                      weeks: s.weeks.map((w) =>
                        w.week === week ? { ...w, planned } : w,
                      ),
                    }
                  : s,
              ),
            }
          : old,
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(plannerKeys.season, ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: plannerKeys.season }),
  });
}
