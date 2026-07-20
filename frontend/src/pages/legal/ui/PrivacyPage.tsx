import { useTranslation } from '@/shared/i18n';

import { LegalPage } from './LegalPage';

const PRIVACY_SECTIONS = [
  'collect',
  'why',
  'cookies',
  'thirdParties',
  'retention',
  'rights',
  'security',
  'changes',
] as const;

export function PrivacyPage() {
  const { t } = useTranslation();
  return (
    <LegalPage title={t('legal.privacy.title')} updated={t('legal.privacy.updated')}>
      <p>{t('legal.privacy.intro')}</p>
      {PRIVACY_SECTIONS.map((key) => (
        <section key={key}>
          <h2>{t(`legal.privacy.${key}.h`)}</h2>
          <p>{t(`legal.privacy.${key}.b`)}</p>
        </section>
      ))}
    </LegalPage>
  );
}
