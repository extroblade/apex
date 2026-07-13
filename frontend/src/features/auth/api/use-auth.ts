import { useMutation, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import { viewerKeys, type Viewer } from '@/entities/viewer';

export interface Credentials {
  email: string;
  password: string;
}

export function useLogin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (c: Credentials) =>
      apiFetch<Viewer>('/api/auth/login', { method: 'POST', body: JSON.stringify(c) }),
    // Prime the viewer cache so the UI updates immediately, no refetch needed.
    onSuccess: (user) => qc.setQueryData(viewerKeys.me, user),
  });
}

export function useRegister() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (c: Credentials) =>
      apiFetch<Viewer>('/api/auth/register', { method: 'POST', body: JSON.stringify(c) }),
    onSuccess: (user) => qc.setQueryData(viewerKeys.me, user),
  });
}

export function useLogout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch<void>('/api/auth/logout', { method: 'POST' }),
    onSuccess: () => qc.setQueryData(viewerKeys.me, null),
  });
}
