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

/** Request a password-reset link be emailed. Always 204 — the API never
 * reveals whether the email belongs to an account. */
export function useRequestPasswordReset() {
  return useMutation({
    mutationFn: (email: string) =>
      apiFetch<void>('/api/auth/password-reset/request', {
        method: 'POST',
        body: JSON.stringify({ email }),
      }),
  });
}

/** Confirm a password reset with the token from the email + a new password. */
export function useConfirmPasswordReset() {
  return useMutation({
    mutationFn: (input: { token: string; newPassword: string }) =>
      apiFetch<void>('/api/auth/password-reset/confirm', {
        method: 'POST',
        body: JSON.stringify(input),
      }),
  });
}

/** Request a verification link be emailed to `email` (pre-login resend form). */
export function useRequestEmailVerification() {
  return useMutation({
    mutationFn: (email: string) =>
      apiFetch<void>('/api/auth/verify-email/request', {
        method: 'POST',
        body: JSON.stringify({ email }),
      }),
  });
}

/** Confirm an email verification with the token from the email link. */
export function useConfirmEmailVerification() {
  return useMutation({
    mutationFn: (token: string) =>
      apiFetch<void>(`/api/auth/verify-email?token=${encodeURIComponent(token)}`, {
        method: 'GET',
      }),
  });
}

/** Resend the verification email for the logged-in user (profile banner). */
export function useResendEmailVerification() {
  return useMutation({
    mutationFn: () =>
      apiFetch<void>('/api/auth/verify-email/resend', { method: 'POST' }),
  });
}
