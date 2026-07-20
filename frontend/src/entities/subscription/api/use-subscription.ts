import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';

export interface BillingPlan {
  key: string;
  name: string;
  price: string;
  interval: string;
  features: string[];
}

export interface SubscriptionSummary {
  tier: string;
  pro: boolean;
  status?: string;
  provider?: string;
  currentPeriodEnd?: string;
  cancelAtPeriodEnd: boolean;
}

export const subscriptionKeys = {
  all: ['subscription'] as const,
  plans: ['subscription', 'plans'] as const,
  mine: ['subscription', 'mine'] as const,
};

export function useBillingPlans() {
  return useQuery({
    queryKey: subscriptionKeys.plans,
    queryFn: () => apiFetch<{ plans: BillingPlan[] }>('/api/billing/plans').then((r) => r.plans),
    staleTime: 5 * 60_000,
  });
}

export function useSubscription(enabled = true) {
  return useQuery({
    queryKey: subscriptionKeys.mine,
    queryFn: () => apiFetch<SubscriptionSummary>('/api/billing/subscription'),
    enabled,
  });
}

export function useSetDevTier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (tier: 'free' | 'pro') =>
      apiFetch<void>('/api/billing/dev/tier', {
        method: 'PUT',
        body: JSON.stringify({ tier, reason: 'manual dev switch from upgrade page' }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: subscriptionKeys.all });
    },
  });
}
