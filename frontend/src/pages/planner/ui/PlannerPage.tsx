import { Link } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useTranslation } from '@/shared/i18n';
import { SeasonPlanner } from '@/features/season-planner';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export function PlannerPage() {
  const { t } = useTranslation();
  const { data: viewer, isLoading } = useViewer();

  if (isLoading) return null;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>{t('profile.signInRequired')}</CardTitle>
          <CardDescription>{t('planner.subtitle')}</CardDescription>
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
        <h1 className="text-2xl font-semibold">{t('planner.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('planner.subtitle')}</p>
      </div>
      <SeasonPlanner />
    </div>
  );
}
