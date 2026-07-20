import { useTranslation } from '@/shared/i18n';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';

export function AboutPage() {
  const { t } = useTranslation();
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('about.title')}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3 text-sm text-muted-foreground">
        <p>{t('about.p1')}</p>
        <p>{t('about.p2')}</p>
      </CardContent>
    </Card>
  );
}
