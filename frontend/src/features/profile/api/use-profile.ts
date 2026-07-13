import { useMutation, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import { viewerKeys, type Viewer } from '@/entities/viewer';

export function useUpdateProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { nickname: string; email: string }) =>
      apiFetch<Viewer>('/api/auth/profile', {
        method: 'PATCH',
        body: JSON.stringify(input),
      }),
    onSuccess: (user) => qc.setQueryData(viewerKeys.me, user),
  });
}

export function useUpdateAvatar() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (avatar: string) =>
      apiFetch<Viewer>('/api/auth/avatar', {
        method: 'PUT',
        body: JSON.stringify({ avatar }),
      }),
    onSuccess: (user) => qc.setQueryData(viewerKeys.me, user),
  });
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (input: { currentPassword: string; newPassword: string }) =>
      apiFetch<void>('/api/auth/password', {
        method: 'POST',
        body: JSON.stringify(input),
      }),
  });
}
