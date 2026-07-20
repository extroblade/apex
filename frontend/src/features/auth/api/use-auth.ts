import { useMutation, useQueryClient } from '@tanstack/react-query';

import { apiFetch } from '@/shared/api/client';
import { env } from '@/shared/config/env';
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
    mutationFn: () => apiFetch<void>('/api/auth/verify-email/resend', { method: 'POST' }),
  });
}

/** Delete the logged-in account (password-confirmed). On success the viewer
 * cache is cleared and the caller should redirect to a logged-out view. */
export function useDeleteAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (currentPassword: string) =>
      apiFetch<void>('/api/auth/account', {
        method: 'DELETE',
        body: JSON.stringify({ currentPassword }),
      }),
    onSuccess: () => qc.setQueryData(viewerKeys.me, null),
  });
}

/** Request an email change: verifies the current password, stages the new
 * email as pending, and emails a verification link to the new address. */
export function useRequestEmailChange() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { newEmail: string; currentPassword: string }) =>
      apiFetch<void>('/api/auth/account/email', {
        method: 'POST',
        body: JSON.stringify(input),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: viewerKeys.me }),
  });
}

/** Cancel a staged email change (clears pending_email + discards the token). */
export function useCancelEmailChange() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => apiFetch<void>('/api/auth/account/email', { method: 'DELETE' }),
    onSuccess: () => qc.invalidateQueries({ queryKey: viewerKeys.me }),
  });
}

/** Trigger a browser download of the user's full data export (GDPR). The
 * response is a JSON file with a Content-Disposition: attachment header. */
export async function downloadAccountExport(): Promise<void> {
  const res = await fetch(`${env.apiBaseUrl}/api/auth/account/export`, {
    method: 'GET',
    credentials: 'same-origin',
  });
  if (!res.ok) {
    let message = `Export failed: ${res.status}`;
    try {
      const body = await res.json();
      if (body && typeof body.error === 'string') message = body.error;
    } catch {
      // no JSON body
    }
    throw new Error(message);
  }
  // The backend sets Content-Disposition: attachment; filename="...". Pull the
  // filename from it (or fall back to a dated default) and trigger a download.
  const disposition = res.headers.get('Content-Disposition') ?? '';
  const match = /filename="([^"]+)"/.exec(disposition);
  const filename =
    match?.[1] ?? `apex-account-export-${new Date().toISOString().slice(0, 10)}.json`;
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
