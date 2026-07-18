import { useTranslation } from '@/shared/i18n';

/**
 * Site footer. Carries the non-affiliation disclaimer required because the app
 * uses the "iRacing" name nominatively to describe compatibility. Rendered at
 * the end of the main content column on every page.
 */
export function Footer() {
  const { t } = useTranslation();
  const year = new Date().getFullYear();

  return (
    <footer className="mt-12 border-t pt-6 text-xs text-muted-foreground">
      <p>{t('footer.disclaimer')}</p>
      <p className="mt-2">
        © {year} Apex. {t('footer.rights')}
      </p>
    </footer>
  );
}
