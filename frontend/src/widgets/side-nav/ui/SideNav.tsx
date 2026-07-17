import { Link, useRoute } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useFeatures } from '@/entities/features';
import { useNav, visibleNav, NavIcon, type NavItem } from '@/entities/nav';
import { useTranslation } from '@/shared/i18n';
import { cn } from '@/shared/lib/utils';

/**
 * Desktop side menu (hidden below md, where the bottom bar takes over). Its
 * items come from the nav service, so the menu is configured on the backend.
 */
export function SideNav() {
  const { data: items = [] } = useNav();
  const { data: viewer } = useViewer();
  const { data: flags = {} } = useFeatures();

  const links = visibleNav(items, 'side', { isAuthed: Boolean(viewer), flags });
  if (links.length === 0) return null;

  return (
    <nav
      aria-label="Main"
      className="hidden w-52 shrink-0 border-r md:block"
      data-testid="side-nav"
    >
      <ul className="sticky top-0 flex flex-col gap-1 p-3">
        {links.map((item) => (
          <li key={item.key}>
            <SideLink item={item} />
          </li>
        ))}
      </ul>
    </nav>
  );
}

function SideLink({ item }: { item: NavItem }) {
  const [isActive] = useRoute(item.href);
  const { t } = useTranslation();
  return (
    <Link
      href={item.href}
      aria-current={isActive ? 'page' : undefined}
      className={cn(
        'flex items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium transition-colors',
        isActive
          ? 'bg-accent text-accent-foreground'
          : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
      )}
    >
      <NavIcon name={item.icon} className="size-4" />
      {t(item.labelKey)}
    </Link>
  );
}
