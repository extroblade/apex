import { useState } from 'react';
import { Link, useSearch } from 'wouter';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

import { useTranslation } from '@/shared/i18n';
import { useConfirmPasswordReset } from '@/features/auth';
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

const schema = z.object({
  newPassword: z.string().min(8, 'auth.passwordTooShort'),
});
type Form = z.infer<typeof schema>;

export function ResetPasswordConfirmPage() {
  const { t } = useTranslation();
  const confirm = useConfirmPasswordReset();
  const [done, setDone] = useState(false);

  // The token arrives in the email link's query string: /reset-password/confirm?token=...
  const searchStr = useSearch();
  const token = new URLSearchParams(searchStr).get('token') ?? '';

  const {
    register: field,
    handleSubmit,
    formState: { errors },
  } = useForm<Form>({ resolver: zodResolver(schema), mode: 'onTouched' });

  const onSubmit = handleSubmit((values) => {
    confirm.mutate(
      { token, newPassword: values.newPassword },
      { onSuccess: () => setDone(true) },
    );
  });

  // No token in the URL — the link was malformed or expired.
  if (!token && !done) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>{t('auth.resetConfirmTitle')}</CardTitle>
          <CardDescription>{t('auth.resetTokenMissing')}</CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild variant="outline" className="w-full">
            <Link href="/reset-password">{t('auth.resetSend')}</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="mx-auto max-w-sm">
      <CardHeader>
        <CardTitle>{t('auth.resetConfirmTitle')}</CardTitle>
        <CardDescription>{t('auth.resetConfirmSubtitle')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {done ? (
          <>
            <p className="text-sm text-muted-foreground">{t('auth.resetConfirmDone')}</p>
            <Button asChild className="w-full">
              <Link href="/login">{t('auth.logIn')}</Link>
            </Button>
          </>
        ) : (
          <form onSubmit={onSubmit} className="space-y-4" noValidate>
            <div className="space-y-1.5">
              <Label htmlFor="newPassword">{t('auth.newPassword')}</Label>
              <Input
                id="newPassword"
                type="password"
                autoComplete="new-password"
                aria-invalid={!!errors.newPassword}
                {...field('newPassword')}
              />
              {errors.newPassword && (
                <p className="text-sm text-destructive">
                  {t(errors.newPassword.message ?? '')}
                </p>
              )}
            </div>
            {confirm.error && (
              <p className="text-sm text-destructive">{(confirm.error as Error).message}</p>
            )}
            <Button type="submit" className="w-full" disabled={confirm.isPending}>
              {confirm.isPending ? t('auth.pleaseWait') : t('auth.resetConfirmButton')}
            </Button>
          </form>
        )}
      </CardContent>
    </Card>
  );
}
