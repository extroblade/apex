import { useState } from 'react';
import { Link } from 'wouter';
import { Plus, X } from 'lucide-react';

import { useViewer } from '@/entities/viewer';
import { useTranslation } from '@/shared/i18n';
import { SetupsShowroom, SetupForm } from '@/features/setups-manager';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export function SetupsPage() {
  const { t } = useTranslation();
  const { data: viewer, isLoading } = useViewer();
  const [creating, setCreating] = useState(false);

  if (isLoading) return null;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>{t('profile.signInRequired')}</CardTitle>
          <CardDescription>{t('setups.subtitle')}</CardDescription>
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
          <h1 className="text-2xl font-semibold">{t('setups.title')}</h1>
          <p className="text-sm text-muted-foreground">{t('setups.subtitle')}</p>
        </div>
        <Button size="sm" variant={creating ? 'outline' : 'default'} onClick={() => setCreating((v) => !v)}>
          {creating ? <X className="size-4" /> : <Plus className="size-4" />}
          {creating ? t('common.cancel') : t('setups.add')}
        </Button>
      </div>

      {creating && (
        <Card>
          <CardHeader>
            <CardTitle>{t('setups.add')}</CardTitle>
            <CardDescription>{t('setups.addHint')}</CardDescription>
          </CardHeader>
          <CardContent>
            <SetupForm onCreated={() => setCreating(false)} />
          </CardContent>
        </Card>
      )}

      <SetupsShowroom />
    </div>
  );
}
