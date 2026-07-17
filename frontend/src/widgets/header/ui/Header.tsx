import { Link } from 'wouter';

import { useTranslation } from '@/shared/i18n';
import { UserMenu } from '@/widgets/user-menu';
import { AppLogo } from '@/shared/ui/logo';

/**
 * Top header, deliberately minimal: brand + user menu. Navigation lives in the
 * side menu on desktop (widgets/side-nav) and the bottom bar on mobile
 * (widgets/bottom-nav), both configured by the nav service.
 */
export function Header() {
  const { t } = useTranslation();

  return (
    <header className="sticky top-0 z-30 border-b bg-background/95 backdrop-blur">
      <div className="flex h-14 items-center justify-between gap-4 px-4">
        <Link
          href="/"
          className="flex shrink-0 items-center gap-2 font-semibold"
          aria-label={t('brand')}
        >
          <AppLogo className="size-6 text-primary" />
          <span>{t('brand')}</span>
        </Link>

        {/* Profile/preferences stays in the header on every viewport. */}
        <div className="flex items-center">
          <UserMenu />
        </div>
      </div>
    </header>
  );
}
