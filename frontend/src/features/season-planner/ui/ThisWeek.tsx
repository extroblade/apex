import { useMemo } from 'react';
import { Link } from 'wouter';
import { CalendarCheck } from 'lucide-react';

import { useSeason, useSetRacePlanned, type SeasonSeries } from '@/entities/planner';
import { useTranslation } from '@/shared/i18n';
import { CatalogThumbnail } from '@/shared/ui/catalog-thumbnail';
import { Badge } from '@/shared/ui/badge';
import { SkeletonRows } from '@/shared/ui/skeleton';
import { cn } from '@/shared/lib/utils';
import { accessTextClass } from './access-colors';

/**
 * The current week's races for the user's favorite series, as tappable cards.
 * Lives on its own page so the season grid can use the full screen width.
 */
export function ThisWeek() {
  const { t } = useTranslation();
  const season = useSeason();
  const setPlanned = useSetRacePlanned();

  const currentWeek = season.data?.currentWeek ?? 1;
  const favorites = useMemo(
    () => (season.data?.series ?? []).filter((s) => s.favorite),
    [season.data],
  );

  if (season.isLoading) return <SkeletonRows rows={4} />;
  if (season.error) {
    return <p className="text-sm text-destructive">{(season.error as Error).message}</p>;
  }

  const thisWeek = favorites
    .map((s) => ({ series: s, week: s.weeks.find((w) => w.week === currentWeek) }))
    .filter(
      (x): x is { series: SeasonSeries; week: NonNullable<typeof x.week> } => !!x.week,
    );

  if (thisWeek.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t('planner.thisWeekEmpty')}{' '}
        <Link href="/garage" className="underline">
          {t('nav.garage')}
        </Link>
      </p>
    );
  }

  return (
    <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {thisWeek.map(({ series, week }) => (
        <li key={series.seriesId}>
          <button
            type="button"
            aria-pressed={week.planned}
            onClick={() =>
              setPlanned.mutate({
                seriesId: series.seriesId,
                week: week.week,
                planned: !week.planned,
              })
            }
            className={cn(
              'flex w-full cursor-pointer items-center gap-3 rounded-md border p-3 text-left transition-shadow',
              'outline-none focus-visible:ring-2 focus-visible:ring-ring',
              week.planned
                ? 'border-amber-500 bg-amber-400/15 ring-1 ring-amber-500'
                : 'hover:bg-muted/50',
            )}
          >
            <CatalogThumbnail
              category={series.category}
              name={series.seriesName}
              className="size-10"
            />
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-medium">{series.seriesName}</div>
              <div className={cn('truncate text-xs', accessTextClass(week.trackAccess))}>
                {week.trackName}
                {week.configName ? ` (${week.configName})` : ''}
              </div>
            </div>
            {week.planned && (
              <CalendarCheck className="size-4 shrink-0 text-amber-600" aria-hidden />
            )}
            {series.licenseNeeded && <Badge>Lic. {series.licenseNeeded}</Badge>}
          </button>
        </li>
      ))}
    </ul>
  );
}
