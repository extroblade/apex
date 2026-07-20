import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { useTranslation } from '@/shared/i18n';
import { useResendEmailVerification } from '@/features/auth';
import { useViewer, viewerKeys } from '@/entities/viewer';
import { Button } from '@/shared/ui/button';
import { Card, CardContent } from '@/shared/ui/card';

/**
 * Banner shown on the profile page when the logged-in user's email isn't
 * verified yet. Lets them resend the verification email. After a successful
 * resend we don't flip `emailVerified` (only the link does) — we just show a
 * "sent" confirmation.
 */
export function VerifyEmailBanner() {
  const { t } = useTranslation();
  const { data: viewer } = useViewer();
  const resend = useResendEmailVerification();
  const qc = useQueryClient();
  const [resent, setResent] = useState(false);

  if (!viewer || viewer.emailVerified) return null;

  const onResend = () => {
    resend.mutate(undefined, {
      onSuccess: () => {
        setResent(true);
        // The banner stays until the user actually clicks the link, but we
        // refresh the viewer cache in case the backend updated anything.
        qc.invalidateQueries({ queryKey: viewerKeys.me });
      },
    });
  };

  return (
    <Card className="border-amber-500/40 bg-amber-500/5">
      <CardContent className="flex flex-col gap-3 py-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-0.5">
          <p className="text-sm font-medium text-foreground">{t('auth.verifyBanner')}</p>
          <p className="text-xs text-muted-foreground">{t('auth.verifyBannerHint')}</p>
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onResend}
          disabled={resend.isPending}
        >
          {resend.isPending
            ? t('auth.pleaseWait')
            : resent
              ? t('auth.verifyResent')
              : t('auth.verifyResend')}
        </Button>
      </CardContent>
    </Card>
  );
}
