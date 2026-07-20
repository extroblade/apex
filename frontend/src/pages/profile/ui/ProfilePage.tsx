import { Link } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useTranslation } from '@/shared/i18n';
import { ProfileSettings } from '@/features/profile';
import { ThemeConfigurator } from '@/features/customize-theme';
import { VerifyEmailBanner } from '@/features/auth';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export function ProfilePage() {
  const { t } = useTranslation();
  const { data: viewer, isLoading } = useViewer();

  if (isLoading) return null;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>{t('profile.signInRequired')}</CardTitle>
          <CardDescription>{t('profile.signInToManage')}</CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild>
            <Link href="/login">{t('common.logIn')}</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t('profile.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('profile.subtitle')}</p>
      </div>
      <VerifyEmailBanner />
      <ProfileSettings />
      <ThemeConfigurator />
    </div>
  );
}
