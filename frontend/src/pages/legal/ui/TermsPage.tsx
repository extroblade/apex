import { useTranslation } from '@/shared/i18n';

import { LegalPage } from './LegalPage';

// Section keys live next to the component so the order is explicit and the
// translation file stays a flat map of `key → {heading, body}` pairs.
const TERMS_SECTIONS = [
  'service',
  'accounts',
  'content',
  'acceptable',
  'thirdParty',
  'subscriptions',
  'warranty',
  'changes',
] as const;

export function TermsPage() {
  const { t } = useTranslation();
  return (
    <LegalPage title={t('legal.terms.title')} updated={t('legal.terms.updated')}>
      <p>{t('legal.terms.intro')}</p>
      {TERMS_SECTIONS.map((key) => (
        <section key={key}>
          <h2>{t(`legal.terms.${key}.h`)}</h2>
          <p>{t(`legal.terms.${key}.b`)}</p>
        </section>
      ))}
    </LegalPage>
  );
}
