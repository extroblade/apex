import { useMemo, useState } from 'react';
import { Link } from 'wouter';
import { Star, ArrowLeftRight, CalendarCheck, CalendarClock } from 'lucide-react';

import {
  useSeason,
  useSetRacePlanned,
  type SeasonSeries,
  type SeasonWeek,
} from '@/entities/planner';
import { useTranslation } from '@/shared/i18n';
import { CatalogThumbnail } from '@/shared/ui/catalog-thumbnail';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { SkeletonRows } from '@/shared/ui/skeleton';
import { cn } from '@/shared/lib/utils';
import { accessCellClasses } from './access-colors';

/**
 * The season grid: series × weeks (or transposed). Cells are colored by track
 * access — green (free/included), aquamarine (owned), red (missing) — and a
 * planned race glows amber. Every cell is a button: click to plan/unplan.
 *
 * The grid breaks out of the page's centered container to use ~the full screen
 * width; the current week's races live on their own page (ThisWeek).
 */
export function SeasonPlanner() {
  const { t } = useTranslation();
  const season = useSeason();
  const [favoritesOnly, setFavoritesOnly] = useState(true);
  const [transposed, setTransposed] = useState(false);

  const allSeries = useMemo(() => season.data?.series ?? [], [season.data]);
  const favorites = useMemo(() => allSeries.filter((s) => s.favorite), [allSeries]);
  const shown = favoritesOnly ? favorites : allSeries;
  const currentWeek = season.data?.currentWeek ?? 1;
  const totalWeeks = season.data?.totalWeeks ?? 13;
  const plannedTotal = useMemo(
    () => allSeries.reduce((sum, s) => sum + s.weeks.filter((w) => w.planned).length, 0),
    [allSeries],
  );

  if (season.isLoading) return <SkeletonRows rows={8} />;
  if (season.error) {
    return <p className="text-sm text-destructive">{(season.error as Error).message}</p>;
  }
  if (allSeries.length === 0) {
    return <p className="text-sm text-muted-foreground">{t('planner.empty')}</p>;
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div
          className="flex flex-wrap gap-2"
          role="group"
          aria-label={t('planner.title')}
        >
          <Button
            size="sm"
            variant={favoritesOnly ? 'default' : 'outline'}
            aria-pressed={favoritesOnly}
            onClick={() => setFavoritesOnly(true)}
          >
            <Star className="size-3.5" />
            {t('planner.favoritesOnly')}
          </Button>
          <Button
            size="sm"
            variant={!favoritesOnly ? 'default' : 'outline'}
            aria-pressed={!favoritesOnly}
            onClick={() => setFavoritesOnly(false)}
          >
            {t('planner.allSeries')}
          </Button>
          <Button
            size="sm"
            variant="outline"
            aria-pressed={transposed}
            onClick={() => setTransposed((v) => !v)}
          >
            <ArrowLeftRight className="size-3.5" />
            {t('planner.transpose')}
          </Button>
          <Button size="sm" variant="outline" asChild>
            <Link href="/this-week">
              <CalendarClock className="size-3.5" />
              {t('planner.openThisWeek')}
            </Link>
          </Button>
        </div>
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-1">
            <CalendarCheck className="size-3.5" />
            {t('planner.plannedCount', { n: plannedTotal })}
          </span>
          <LegendSwatch
            className="bg-green-500/25 ring-1 ring-green-600"
            label={t('planner.legendFree')}
          />
          <LegendSwatch
            className="bg-teal-500/25 ring-1 ring-teal-600"
            label={t('planner.legendOwned')}
          />
          <LegendSwatch
            className="bg-red-500/20 ring-1 ring-red-600/70"
            label={t('planner.legendMissing')}
          />
          <LegendSwatch
            className="bg-amber-400/40 ring-2 ring-amber-500"
            label={t('planner.legendPlanned')}
          />
        </div>
      </div>

      {favoritesOnly && favorites.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          {t('planner.noFavorites')}{' '}
          <Link href="/garage" className="underline">
            {t('nav.garage')}
          </Link>
        </p>
      ) : (
        // Break out of the centered <main> container so the grid uses ~full width.
        <div className="mx-[calc(50%-50vw)] w-screen px-4 sm:px-6 lg:px-8">
          {transposed ? (
            <WeeksBySeries
              series={shown}
              currentWeek={currentWeek}
              totalWeeks={totalWeeks}
            />
          ) : (
            <SeriesByWeeks
              series={shown}
              currentWeek={currentWeek}
              totalWeeks={totalWeeks}
            />
          )}
        </div>
      )}
    </div>
  );
}

function LegendSwatch({ className, label }: { className: string; label: string }) {
  return (
    <span className="inline-flex items-center gap-1">
      <span className={cn('inline-block size-3 rounded-sm', className)} />
      {label}
    </span>
  );
}

/** A clickable schedule cell colored by access; amber when planned. */
function WeekCell({ series, week }: { series: SeasonSeries; week: SeasonWeek }) {
  const { t } = useTranslation();
  const setPlanned = useSetRacePlanned();
  const label = week.trackName + (week.configName ? ` (${week.configName})` : '');
  const dated = week.raceDate ? `${label} · ${week.raceDate}` : label;

  return (
    <button
      type="button"
      aria-pressed={week.planned}
      aria-label={`${series.seriesName} — ${t('planner.week', { n: week.week })} — ${label}`}
      title={`${dated} — ${t('planner.cellHint')}`}
      onClick={() =>
        setPlanned.mutate({
          seriesId: series.seriesId,
          week: week.week,
          planned: !week.planned,
        })
      }
      className={cn(
        'inline-block w-full cursor-pointer rounded-sm px-1.5 py-1 text-left leading-tight outline-none transition-shadow',
        'focus-visible:ring-2 focus-visible:ring-ring',
        accessCellClasses(week.trackAccess, week.planned),
      )}
    >
      {week.planned && <CalendarCheck className="mr-1 inline size-3" aria-hidden />}
      {shortTrack(week.trackName)}
    </button>
  );
}

function SeriesHeaderCell({ s, weeks }: { s: SeasonSeries; weeks: SeasonWeek[] }) {
  const { t } = useTranslation();
  const ownedWeeks = weeks.filter((w) => w.trackOwned).length;
  return (
    <div className="flex items-center gap-2">
      <CatalogThumbnail
        category={s.category}
        name={s.seriesName}
        className="size-7 shrink-0"
      />
      <div className="min-w-0">
        <div className="truncate text-sm font-medium">
          {s.favorite && (
            <Star className="mr-1 inline size-3 fill-current text-primary" aria-hidden />
          )}
          {s.seriesName}
        </div>
        <div className="mt-0.5 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
          {s.licenseNeeded && <Badge>Lic. {s.licenseNeeded}</Badge>}
          <span title={t('planner.canRunHint')}>
            {t('planner.canRun', { n: ownedWeeks, total: weeks.length })}
          </span>
          {!s.carOwned && <span className="text-red-500">{t('planner.noCar')}</span>}
        </div>
      </div>
    </div>
  );
}

/** Default orientation: rows = series, columns = weeks. */
function SeriesByWeeks({
  series,
  currentWeek,
  totalWeeks,
}: {
  series: SeasonSeries[];
  currentWeek: number;
  totalWeeks: number;
}) {
  const { t } = useTranslation();
  const weekNumbers = Array.from({ length: totalWeeks }, (_, i) => i + 1);

  return (
    <div className="overflow-x-auto rounded-md border">
      <table className="w-full table-fixed border-collapse text-xs">
        <thead>
          <tr className="border-b bg-muted/40 text-muted-foreground">
            <th className="sticky left-0 z-10 w-52 bg-background px-3 py-2 text-left font-medium">
              {t('planner.allSeries')}
            </th>
            {weekNumbers.map((n) => (
              <th
                key={n}
                scope="col"
                className={cn(
                  'px-2 py-2 text-left font-medium',
                  n === currentWeek && 'bg-primary/15 text-foreground',
                )}
              >
                {t('planner.week', { n })}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {series.map((s) => (
            <tr key={s.seriesId} className="border-b align-top last:border-0">
              <th
                scope="row"
                className="sticky left-0 z-10 w-52 bg-background px-3 py-2 text-left"
              >
                <SeriesHeaderCell s={s} weeks={s.weeks} />
              </th>
              {weekNumbers.map((n) => {
                // Real schedules have gaps — a series may not race every week.
                const w = s.weeks.find((x) => x.week === n);
                return (
                  <td
                    key={n}
                    className={cn('px-1.5 py-1.5', n === currentWeek && 'bg-primary/10')}
                  >
                    {w && <WeekCell series={s} week={w} />}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Transposed orientation: rows = weeks, columns = series. */
function WeeksBySeries({
  series,
  currentWeek,
  totalWeeks,
}: {
  series: SeasonSeries[];
  currentWeek: number;
  totalWeeks: number;
}) {
  const { t } = useTranslation();
  const weekNumbers = Array.from({ length: totalWeeks }, (_, i) => i + 1);

  return (
    <div className="overflow-x-auto rounded-md border">
      <table className="w-full table-fixed border-collapse text-xs">
        <thead>
          <tr className="border-b bg-muted/40 text-muted-foreground">
            <th className="sticky left-0 z-10 w-24 bg-background px-3 py-2 text-left font-medium" />
            {series.map((s) => (
              <th key={s.seriesId} scope="col" className="px-2 py-2 text-left">
                <SeriesHeaderCell s={s} weeks={s.weeks} />
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {weekNumbers.map((n) => (
            <tr
              key={n}
              className={cn(
                'border-b align-top last:border-0',
                n === currentWeek && 'bg-primary/10',
              )}
            >
              <th
                scope="row"
                className="sticky left-0 z-10 w-24 whitespace-nowrap bg-background px-3 py-2 text-left font-medium"
              >
                {t('planner.week', { n })}
              </th>
              {series.map((s) => {
                const w = s.weeks.find((x) => x.week === n);
                return (
                  <td key={s.seriesId} className="px-1.5 py-1.5">
                    {w && <WeekCell series={s} week={w} />}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Compacts long track names for grid cells. */
function shortTrack(name: string): string {
  const stripped = name
    .replace(
      /^(Circuit de |Circuit |Autodromo Nazionale |Autodromo Internazionale |Autódromo |Mobility Resort |WeatherTech Raceway )/i,
      '',
    )
    .trim();
  return stripped.length > 22 ? `${stripped.slice(0, 21)}…` : stripped;
}
