import { useEffect, useState } from 'react';
import { Link, useSearch } from 'wouter';

import { useTranslation } from '@/shared/i18n';
import { useConfirmEmailVerification } from '@/features/auth';
import { Button } from '@/shared/ui/button';
import { Card, CardContent } from '@/shared/ui/card';

type State = 'verifying' | 'done' | 'failed';

export function VerifyEmailPage() {
  const { t } = useTranslation();
  const confirm = useConfirmEmailVerification();

  const searchStr = useSearch();
  const token = new URLSearchParams(searchStr).get('token') ?? '';

  const [state, setState] = useState<State>(token ? 'verifying' : 'failed');

  useEffect(() => {
    if (!token) {
      setState('failed');
      return;
    }
    confirm.mutate(token, {
      onSuccess: () => setState('done'),
      onError: () => setState('failed'),
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  return (
    <Card className="mx-auto max-w-sm">
      <CardContent className="space-y-4 py-6">
        {state === 'verifying' && (
          <p className="text-sm text-muted-foreground">{t('auth.verifySubtitle')}</p>
        )}
        {state === 'done' && (
          <>
            <p className="text-sm text-foreground">{t('auth.verifyDone')}</p>
            <Button asChild className="w-full">
              <Link href="/">{t('nav.home')}</Link>
            </Button>
          </>
        )}
        {state === 'failed' && (
          <>
            <p className="text-sm text-destructive">{t('auth.verifyFailed')}</p>
            <Button asChild variant="outline" className="w-full">
              <Link href="/login">{t('auth.logIn')}</Link>
            </Button>
          </>
        )}
      </CardContent>
    </Card>
  );
}
