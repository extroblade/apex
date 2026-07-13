import { useState } from 'react';
import { Link } from 'wouter';
import { Plus, X } from 'lucide-react';

import { useViewer } from '@/entities/viewer';
import { useTranslation } from '@/shared/i18n';
import { GoalTracker, GoalForm } from '@/features/goal-tracker';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export function GoalsPage() {
  const { t } = useTranslation();
  const { data: viewer, isLoading } = useViewer();
  const [creating, setCreating] = useState(false);

  if (isLoading) return null;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>{t('profile.signInRequired')}</CardTitle>
          <CardDescription>{t('goals.subtitle')}</CardDescription>
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
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t('goals.title')}</h1>
          <p className="text-sm text-muted-foreground">{t('goals.subtitle')}</p>
        </div>
        <Button size="sm" variant={creating ? 'outline' : 'default'} onClick={() => setCreating((v) => !v)}>
          {creating ? <X className="size-4" /> : <Plus className="size-4" />}
          {creating ? t('common.cancel') : t('goals.add')}
        </Button>
      </div>

      {creating && (
        <Card>
          <CardHeader>
            <CardTitle>{t('goals.add')}</CardTitle>
          </CardHeader>
          <CardContent>
            <GoalForm onCreated={() => setCreating(false)} />
          </CardContent>
        </Card>
      )}

      <GoalTracker />
    </div>
  );
}
