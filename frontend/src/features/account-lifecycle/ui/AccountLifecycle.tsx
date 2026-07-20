import { useState } from 'react';
import { useLocation } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useTranslation } from '@/shared/i18n';
import {
  useCancelEmailChange,
  useDeleteAccount,
  useRequestEmailChange,
  downloadAccountExport,
} from '@/features/auth/api/use-auth';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog';

export function AccountLifecycle() {
  const { data: viewer } = useViewer();

  if (!viewer) return null;

  return (
    <div className="space-y-6">
      <EmailChangeCard pendingEmail={viewer.pendingEmail} currentEmail={viewer.email} />
      <DataExportCard />
      <DeleteAccountCard />
    </div>
  );
}

function EmailChangeCard({
  pendingEmail,
  currentEmail,
}: {
  pendingEmail?: string;
  currentEmail: string;
}) {
  const { t } = useTranslation();
  const request = useRequestEmailChange();
  const cancel = useCancelEmailChange();
  const [newEmail, setNewEmail] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    request.mutate(
      { newEmail, currentPassword },
      {
        onSuccess: () => {
          setNewEmail('');
          setCurrentPassword('');
        },
      },
    );
  };

  const onCancel = () => {
    cancel.mutate();
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('profile.changeEmail')}</CardTitle>
        <CardDescription>{t('profile.changeEmailHint')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {pendingEmail && pendingEmail !== currentEmail && (
          <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 p-3 text-sm">
            <span>{t('profile.emailChangePending', { email: pendingEmail })}</span>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onCancel}
              disabled={cancel.isPending}
            >
              {t('profile.cancelEmailChange')}
            </Button>
          </div>
        )}
        <form onSubmit={onSubmit} className="max-w-sm space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="new-email">{t('profile.newEmail')}</Label>
            <Input
              id="new-email"
              type="email"
              autoComplete="email"
              value={newEmail}
              onChange={(e) => setNewEmail(e.target.value)}
              required
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="ec-cur-pw">{t('profile.currentPassword')}</Label>
            <Input
              id="ec-cur-pw"
              type="password"
              autoComplete="current-password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
            />
          </div>
          {request.error && (
            <p className="text-sm text-destructive">{(request.error as Error).message}</p>
          )}
          {request.isSuccess && !request.isPending && (
            <p className="text-sm text-green-600">{t('profile.emailChangeRequested')}</p>
          )}
          {cancel.isSuccess && !cancel.isPending && (
            <p className="text-sm text-green-600">{t('profile.emailChangeCanceled')}</p>
          )}
          <Button type="submit" disabled={request.isPending}>
            {t('profile.requestEmailChange')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function DataExportCard() {
  const { t } = useTranslation();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const onExport = async () => {
    setError('');
    setLoading(true);
    try {
      await downloadAccountExport();
    } catch (e) {
      setError(e instanceof Error ? e.message : t('profile.exportFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('profile.dataExport')}</CardTitle>
        <CardDescription>{t('profile.dataExportHint')}</CardDescription>
      </CardHeader>
      <CardContent>
        <Button onClick={onExport} disabled={loading}>
          {loading ? t('common.loading') : t('profile.exportData')}
        </Button>
        {error && <p className="mt-2 text-sm text-destructive">{error}</p>}
      </CardContent>
    </Card>
  );
}

function DeleteAccountCard() {
  const { t } = useTranslation();
  const [, navigate] = useLocation();
  const del = useDeleteAccount();
  const [open, setOpen] = useState(false);
  const [password, setPassword] = useState('');

  const onConfirm = (e: React.FormEvent) => {
    e.preventDefault();
    del.mutate(password, {
      onSuccess: () => {
        setOpen(false);
        setPassword('');
        navigate('/');
      },
    });
  };

  return (
    <Card className="border-destructive/40">
      <CardHeader>
        <CardTitle className="text-destructive">{t('profile.dangerZone')}</CardTitle>
        <CardDescription>{t('profile.dangerZoneHint')}</CardDescription>
      </CardHeader>
      <CardContent>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button variant="destructive">{t('profile.deleteAccount')}</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t('profile.deleteAccountConfirmTitle')}</DialogTitle>
              <DialogDescription>{t('profile.deleteAccountConfirmBody')}</DialogDescription>
            </DialogHeader>
            <form onSubmit={onConfirm} className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="del-pw">{t('profile.deleteAccountPasswordLabel')}</Label>
                <Input
                  id="del-pw"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  autoFocus
                />
              </div>
              {del.error && (
                <p className="text-sm text-destructive">{(del.error as Error).message}</p>
              )}
              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setOpen(false)}
                  disabled={del.isPending}
                >
                  {t('common.cancel')}
                </Button>
                <Button type="submit" variant="destructive" disabled={del.isPending}>
                  {t('profile.deleteAccountConfirm')}
                </Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  );
}
