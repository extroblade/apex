import { useState } from 'react';
import { Link, useLocation, useRoute } from 'wouter';
import { MoreHorizontal } from 'lucide-react';

import { useViewer } from '@/entities/viewer';
import { useFeatures } from '@/entities/features';
import { useNav, visibleNav, NavIcon, type NavItem } from '@/entities/nav';
import { useTranslation } from '@/shared/i18n';
import { cn } from '@/shared/lib/utils';

// The bar shows at most 5 buttons. Profile lives in the header, not here.
const MAX_SLOTS = 5;

/**
 * Instagram-style mobile bottom bar (hidden on md+). Which items appear — and
 * in what order — is configured on the backend (the nav service, "bottom"
 * placement); anything that doesn't fit the 5 slots folds behind a "More"
 * button.
 */
export function BottomNav() {
  const { data: items = [] } = useNav();
  const { data: viewer } = useViewer();
  const { data: flags = {} } = useFeatures();
  const { t } = useTranslation();
  const [moreOpen, setMoreOpen] = useState(false);
  const [location] = useLocation();

  const links = visibleNav(items, 'bottom', { isAuthed: Boolean(viewer), flags });

  // Fit everything if possible; otherwise keep a slot for "More".
  const needsMore = links.length > MAX_SLOTS;
  const primary = needsMore ? links.slice(0, MAX_SLOTS - 1) : links;
  const overflow = needsMore ? links.slice(MAX_SLOTS - 1) : [];
  const overflowActive = overflow.some((i) => i.href === location);

  return (
    <nav
      aria-label="Primary"
      className="fixed inset-x-0 bottom-0 z-40 border-t bg-background/95 backdrop-blur md:hidden"
    >
      {moreOpen && overflow.length > 0 && (
        <div className="border-b">
          <div className="mx-auto grid max-w-md grid-cols-4 gap-1 p-2">
            {overflow.map((item) => (
              <BarLink key={item.key} item={item} onNavigate={() => setMoreOpen(false)} />
            ))}
          </div>
        </div>
      )}

      <div className="mx-auto flex h-16 max-w-md items-stretch justify-around px-2 pb-[env(safe-area-inset-bottom)]">
        {primary.map((item) => (
          <BarLink key={item.key} item={item} onNavigate={() => setMoreOpen(false)} />
        ))}

        {overflow.length > 0 && (
          <button
            type="button"
            onClick={() => setMoreOpen((v) => !v)}
            aria-expanded={moreOpen}
            aria-label={t('nav.more')}
            className={cn(
              'flex min-w-14 cursor-pointer flex-col items-center justify-center gap-0.5 rounded-md text-[11px] font-medium',
              moreOpen || overflowActive ? 'text-foreground' : 'text-muted-foreground',
            )}
          >
            <MoreHorizontal className="size-5" aria-hidden />
            {t('nav.more')}
          </button>
        )}
      </div>
    </nav>
  );
}

function BarLink({ item, onNavigate }: { item: NavItem; onNavigate: () => void }) {
  const [isActive] = useRoute(item.href);
  const { t } = useTranslation();
  return (
    <Link
      href={item.href}
      onClick={onNavigate}
      aria-current={isActive ? 'page' : undefined}
      className={cn(
        'flex min-w-14 flex-col items-center justify-center gap-0.5 rounded-md py-1.5 text-[11px] font-medium',
        isActive ? 'text-foreground' : 'text-muted-foreground',
      )}
    >
      <NavIcon name={item.icon} className="size-5" />
      {t(item.labelKey)}
    </Link>
  );
}
