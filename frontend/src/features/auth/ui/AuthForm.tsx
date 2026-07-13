import { useState } from 'react';
import { useLocation } from 'wouter';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

import { useTranslation } from '@/shared/i18n';
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
import { useLogin, useRegister } from '../api/use-auth';

type Mode = 'login' | 'register';

// Validation keys, not messages: the resolver stores the i18n key and the field
// renders it through t(), so errors switch language with the rest of the UI.
const credentialsSchema = z.object({
  email: z.string().email('auth.invalidEmail'),
  password: z.string().min(8, 'auth.passwordTooShort'),
});

type Credentials = z.infer<typeof credentialsSchema>;

export function AuthForm() {
  const { t } = useTranslation();
  const [mode, setMode] = useState<Mode>('login');
  const [, navigate] = useLocation();

  const login = useLogin();
  const register = useRegister();
  const active = mode === 'login' ? login : register;

  const {
    register: field,
    handleSubmit,
    formState: { errors },
  } = useForm<Credentials>({
    resolver: zodResolver(credentialsSchema),
    mode: 'onTouched',
  });

  const onSubmit = handleSubmit((values) => {
    active.mutate(values, { onSuccess: () => navigate('/') });
  });

  return (
    <Card className="mx-auto max-w-sm">
      <CardHeader>
        <CardTitle>
          {mode === 'login' ? t('auth.logInTitle') : t('auth.signUpTitle')}
        </CardTitle>
        <CardDescription>
          {mode === 'login' ? t('auth.welcomeBack') : t('auth.startPlanning')}
        </CardDescription>
      </CardHeader>
      <CardContent>
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
          <div className="space-y-1.5">
            <Label htmlFor="password">{t('auth.password')}</Label>
            <Input
              id="password"
              type="password"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              aria-invalid={!!errors.password}
              {...field('password')}
            />
            {errors.password && (
              <p className="text-sm text-destructive">
                {t(errors.password.message ?? '')}
              </p>
            )}
          </div>

          {active.error && (
            <p className="text-sm text-destructive">{(active.error as Error).message}</p>
          )}

          <Button type="submit" className="w-full" disabled={active.isPending}>
            {active.isPending
              ? t('auth.pleaseWait')
              : mode === 'login'
                ? t('auth.logIn')
                : t('auth.signUp')}
          </Button>
        </form>

        <p className="mt-4 text-center text-sm text-muted-foreground">
          {mode === 'login' ? t('auth.noAccount') : t('auth.haveAccount')}
          <button
            type="button"
            className="font-medium text-foreground underline underline-offset-4"
            onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
          >
            {mode === 'login' ? t('auth.signUp') : t('auth.logIn')}
          </button>
        </p>
      </CardContent>
    </Card>
  );
}
