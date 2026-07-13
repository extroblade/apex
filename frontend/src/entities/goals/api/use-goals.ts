import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import type { Goal, GoalInput } from '../model/types';

export const goalsKeys = {
  all: ['goals'] as const,
  list: ['goals', 'list'] as const,
};

export function useGoals() {
  return useQuery({
    queryKey: goalsKeys.list,
    queryFn: () => apiFetch<Goal[]>('/api/goals/'),
  });
}

export function useCreateGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: GoalInput) =>
      apiFetch<Goal>('/api/goals/', { method: 'POST', body: JSON.stringify(input) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: goalsKeys.all }),
  });
}

export function useUpdateGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: GoalInput }) =>
      apiFetch<Goal>(`/api/goals/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: goalsKeys.all }),
  });
}

export function useDeleteGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => apiFetch<void>(`/api/goals/${id}`, { method: 'DELETE' }),
    onSuccess: () => qc.invalidateQueries({ queryKey: goalsKeys.all }),
  });
}
