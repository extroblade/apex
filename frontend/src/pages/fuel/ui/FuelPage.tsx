import { useTranslation } from '@/shared/i18n';
import { FuelCalculator } from '@/features/fuel-calculator';

export function FuelPage() {
  const { t } = useTranslation();
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t('fuel.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('fuel.subtitle')}</p>
      </div>
      <FuelCalculator />
    </div>
  );
}
