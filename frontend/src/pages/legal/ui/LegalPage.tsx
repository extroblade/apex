import type { ReactNode } from 'react';

import { useTranslation } from '@/shared/i18n';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';

/** Shared layout for the Terms / Privacy pages: a titled card with prose. */
export function LegalPage({
  title,
  updated,
  children,
}: {
  title: string;
  updated: string;
  children: ReactNode;
}) {
  const { t } = useTranslation();
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <p className="text-xs text-muted-foreground">{t('legal.updated', { date: updated })}</p>
      </CardHeader>
      <CardContent className="space-y-4 text-sm leading-relaxed text-muted-foreground [&_h2]:mt-2 [&_h2]:font-medium [&_h2]:text-foreground">
        {children}
      </CardContent>
    </Card>
  );
}
