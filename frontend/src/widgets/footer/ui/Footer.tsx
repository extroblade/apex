import { Link } from 'wouter';

import { useTranslation } from '@/shared/i18n';

/**
 * Site footer. Carries the non-affiliation disclaimer (the app uses the
 * "iRacing" name nominatively) and the legal/about links. Rendered at the end of
 * the main content column on every page.
 */
export function Footer() {
  const { t } = useTranslation();
  const year = new Date().getFullYear();

  const links = [
    { href: '/about', label: t('footer.about') },
    { href: '/terms', label: t('footer.terms') },
    { href: '/privacy', label: t('footer.privacy') },
  ];

  return (
    <footer className="mt-12 border-t pt-6 text-xs text-muted-foreground">
      <p>{t('footer.disclaimer')}</p>
      <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1">
        <span>
          © {year} ContentPilot. {t('footer.rights')}
        </span>
        {links.map((l) => (
          <Link
            key={l.href}
            href={l.href}
            className="underline-offset-4 hover:text-foreground hover:underline"
          >
            {l.label}
          </Link>
        ))}
      </div>
    </footer>
  );
}
