import { Link, useRoute } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useFeature, IRACING_OAUTH } from '@/entities/features';
import { useTranslation } from '@/shared/i18n';
import { UserMenu } from '@/widgets/user-menu';
import { AppLogo } from '@/shared/ui/logo';
import { cn } from '@/shared/lib/utils';

function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  const [isActive] = useRoute(href);
  return (
    <Link
      href={href}
      aria-current={isActive ? 'page' : undefined}
      className={cn(
        'text-sm font-medium transition-colors hover:text-foreground',
        isActive ? 'text-foreground' : 'text-muted-foreground',
      )}
    >
      {children}
    </Link>
  );
}

/**
 * Top header. On desktop it carries the full nav + user menu; on mobile it is
 * just the brand — navigation lives in the bottom bar (widgets/bottom-nav).
 */
export function Header() {
  const { data: viewer } = useViewer();
  const iracing = useFeature(IRACING_OAUTH);
  const { t } = useTranslation();

  const links = [
    { href: '/', label: t('nav.home') },
    { href: '/fuel', label: t('nav.fuel') },
    ...(viewer
      ? [
          { href: '/planner', label: t('nav.planner') },
          { href: '/garage', label: t('nav.garage') },
          { href: '/setups', label: t('nav.setups') },
          { href: '/goals', label: t('nav.goals') },
        ]
      : []),
    ...(viewer && iracing
      ? [
          { href: '/drivers', label: t('nav.drivers') },
          { href: '/dashboard', label: t('nav.dashboard') },
          { href: '/compare', label: t('nav.compare') },
        ]
      : []),
  ];

  return (
    <header className="border-b">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between gap-4 px-4">
        <Link
          href="/"
          className="flex shrink-0 items-center gap-2 font-semibold"
          aria-label={t('brand')}
        >
          <AppLogo className="size-6 text-primary" />
          <span>{t('brand')}</span>
        </Link>

        <nav className="hidden items-center gap-5 md:flex" aria-label="Main">
          {links.map((l) => (
            <NavLink key={l.href} href={l.href}>
              {l.label}
            </NavLink>
          ))}
        </nav>

        {/* Profile/preferences stays in the header on every viewport. */}
        <div className="flex items-center">
          <UserMenu />
        </div>
      </div>
    </header>
  );
}
