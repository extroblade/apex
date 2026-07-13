import { useState } from 'react';
import { Link, useLocation, useRoute } from 'wouter';
import {
  Home,
  Fuel,
  CalendarRange,
  Warehouse,
  Wrench,
  Target,
  Users,
  Gauge,
  BarChart3,
  MoreHorizontal,
  type LucideIcon,
} from 'lucide-react';

import { useViewer } from '@/entities/viewer';
import { useFeature, IRACING_OAUTH } from '@/entities/features';
import { useTranslation } from '@/shared/i18n';
import { cn } from '@/shared/lib/utils';

interface NavItem {
  href: string;
  label: string;
  Icon: LucideIcon;
}

// The bar shows at most 5 buttons. Profile lives in the header, not here.
const MAX_SLOTS = 5;

/**
 * Instagram-style mobile bottom bar (hidden on md+). Everything that doesn't
 * fit into the 5 slots goes behind a right-side "More" button — unless only a
 * single item would overflow, in which case it's shown directly (no More).
 */
export function BottomNav() {
  const { data: viewer } = useViewer();
  const iracing = useFeature(IRACING_OAUTH);
  const { t } = useTranslation();
  const [moreOpen, setMoreOpen] = useState(false);
  const [location] = useLocation();

  const items: NavItem[] = [
    { href: '/', label: t('nav.home'), Icon: Home },
    { href: '/fuel', label: t('nav.fuel'), Icon: Fuel },
    ...(viewer
      ? [
          { href: '/planner', label: t('nav.planner'), Icon: CalendarRange },
          { href: '/garage', label: t('nav.garage'), Icon: Warehouse },
          { href: '/setups', label: t('nav.setups'), Icon: Wrench },
          { href: '/goals', label: t('nav.goals'), Icon: Target },
        ]
      : []),
    ...(viewer && iracing
      ? [
          { href: '/drivers', label: t('nav.drivers'), Icon: Users },
          { href: '/dashboard', label: t('nav.dashboard'), Icon: Gauge },
          { href: '/compare', label: t('nav.compare'), Icon: BarChart3 },
        ]
      : []),
  ];

  // Fit everything if possible; only fold into "More" when 2+ items overflow.
  const needsMore = items.length > MAX_SLOTS;
  const primary = needsMore ? items.slice(0, MAX_SLOTS - 1) : items;
  const overflow = needsMore ? items.slice(MAX_SLOTS - 1) : [];
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
              <BarLink key={item.href} item={item} onNavigate={() => setMoreOpen(false)} />
            ))}
          </div>
        </div>
      )}

      <div className="mx-auto flex h-16 max-w-md items-stretch justify-around px-2 pb-[env(safe-area-inset-bottom)]">
        {primary.map((item) => (
          <BarLink key={item.href} item={item} onNavigate={() => setMoreOpen(false)} />
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
  const { Icon } = item;
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
      <Icon className="size-5" aria-hidden />
      {item.label}
    </Link>
  );
}
