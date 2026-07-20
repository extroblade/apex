import { useState } from 'react';
import { Link } from 'wouter';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

import { useTranslation } from '@/shared/i18n';
import { useRequestPasswordReset } from '@/features/auth';
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
  email: z.string().email('auth.invalidEmail'),
});
type Form = z.infer<typeof schema>;

export function ResetPasswordPage() {
  const { t } = useTranslation();
  const req = useRequestPasswordReset();
  const [sent, setSent] = useState(false);

  const {
    register: field,
    handleSubmit,
    formState: { errors },
  } = useForm<Form>({ resolver: zodResolver(schema), mode: 'onTouched' });

  const onSubmit = handleSubmit((values) => {
    req.mutate(values.email, { onSuccess: () => setSent(true) });
  });

  return (
    <Card className="mx-auto max-w-sm">
      <CardHeader>
        <CardTitle>{t('auth.resetTitle')}</CardTitle>
        <CardDescription>{t('auth.resetSubtitle')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {sent ? (
          <>
            <p className="text-sm text-muted-foreground">{t('auth.resetSent')}</p>
            <Button asChild variant="outline" className="w-full">
              <Link href="/login">{t('auth.logIn')}</Link>
            </Button>
          </>
        ) : (
          <form onSubmit={onSubmit} className="space-y-4" noValidate>
            <div className="space-y-1.5">
              <Label htmlFor="email">{t('auth.email')}</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                aria-invalid={!!errors.email}
                {...field('email')}
              />
              {errors.email && (
                <p className="text-sm text-destructive">{t(errors.email.message ?? '')}</p>
              )}
            </div>
            {req.error && (
              <p className="text-sm text-destructive">{(req.error as Error).message}</p>
            )}
            <Button type="submit" className="w-full" disabled={req.isPending}>
              {req.isPending ? t('auth.pleaseWait') : t('auth.resetSend')}
            </Button>
          </form>
        )}
      </CardContent>
    </Card>
  );
}
