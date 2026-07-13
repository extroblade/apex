import { useMemo, useState } from 'react';

import { useCars, useTracks, useSeriesList, type TrackItem } from '@/entities/planner';
import { useFeature, IRACING_OAUTH } from '@/entities/features';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';
import {
  useSyncCatalog,
  useSetCarOwned,
  useSetTrackOwned,
  useSetSeriesFavorite,
} from '../api/use-manage';
import { CatalogList, type CatalogRow } from './CatalogList';

type Tab = 'cars' | 'tracks' | 'series';

export function GarageManager() {
  const [tab, setTab] = useState<Tab>('cars');
  const iracing = useFeature(IRACING_OAUTH);

  return (
    <div className="space-y-6">
      {iracing && <SyncCard />}

      <div className="flex gap-2">
        {(['cars', 'tracks', 'series'] as Tab[]).map((t) => (
          <Button
            key={t}
            size="sm"
            variant={tab === t ? 'default' : 'outline'}
            onClick={() => setTab(t)}
          >
            {t[0].toUpperCase() + t.slice(1)}
          </Button>
        ))}
      </div>

      {tab === 'cars' && <CarsPanel />}
      {tab === 'tracks' && <TracksPanel />}
      {tab === 'series' && <SeriesPanel />}
    </div>
  );
}

function CarsPanel() {
  const cars = useCars();
  const setOwned = useSetCarOwned();
  const rows: CatalogRow[] = (cars.data ?? []).map((c) => ({
    id: c.carId,
    name: c.carName,
    category: c.category,
    description: c.description,
    imagePath: c.imagePath || undefined,
    checked: c.owned,
  }));
  return (
    <CatalogList
      items={rows}
      loading={cars.isLoading}
      onToggle={(carId, owned) => setOwned.mutate({ carId, owned })}
    />
  );
}

function TracksPanel() {
  const tracks = useTracks();
  const setOwned = useSetTrackOwned();

  // iRacing sells a track as one package (all its configs at once), so we show a
  // single row per base track — no per-layout duplicates — and list the layouts
  // in the info dialog. Owning is per-config under the hood, so toggling the row
  // toggles every config of that track together.
  const { rows, configsByGroup } = useMemo(() => {
    const groups = new Map<
      string,
      {
        rep: TrackItem;
        configs: number[];
        layouts: { name: string; free?: boolean }[];
        owned: boolean;
      }
    >();
    for (const t of tracks.data ?? []) {
      const g = groups.get(t.trackName);
      if (g) {
        g.configs.push(t.trackId);
        if (t.configName) g.layouts.push({ name: t.configName, free: t.free });
        g.owned = g.owned || t.owned;
      } else {
        groups.set(t.trackName, {
          rep: t,
          configs: [t.trackId],
          layouts: t.configName ? [{ name: t.configName, free: t.free }] : [],
          owned: t.owned,
        });
      }
    }
    const rows: CatalogRow[] = [];
    const configsByGroup = new Map<number, number[]>();
    for (const g of groups.values()) {
      configsByGroup.set(g.rep.trackId, g.configs);
      rows.push({
        id: g.rep.trackId,
        name: g.rep.trackName,
        category: g.rep.category,
        description: g.rep.description,
        layouts: g.layouts,
        imagePath: g.rep.imagePath || undefined,
        checked: g.owned,
      });
    }
    return { rows, configsByGroup };
  }, [tracks.data]);

  return (
    <CatalogList
      items={rows}
      loading={tracks.isLoading}
      onToggle={(groupId, owned) => {
        for (const trackId of configsByGroup.get(groupId) ?? [groupId]) {
          setOwned.mutate({ trackId, owned });
        }
      }}
    />
  );
}

function SeriesPanel() {
  const series = useSeriesList();
  const setFav = useSetSeriesFavorite();
  const rows: CatalogRow[] = (series.data ?? []).map((s) => ({
    id: s.seriesId,
    name: s.seriesName,
    category: s.category,
    description: s.description,
    badge: s.licenseNeeded ? `Lic. ${s.licenseNeeded}` : undefined,
    imagePath: s.imagePath || undefined,
    checked: s.favorite,
  }));
  return (
    <CatalogList
      items={rows}
      loading={series.isLoading}
      onToggle={(seriesId, favorite) => setFav.mutate({ seriesId, favorite })}
    />
  );
}

function SyncCard() {
  const sync = useSyncCatalog();
  return (
    <Card>
      <CardHeader>
        <CardTitle>Catalog</CardTitle>
        <CardDescription>Refresh cars, tracks, and series from iRacing.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-wrap items-center gap-3">
        <Button onClick={() => sync.mutate()} disabled={sync.isPending}>
          {sync.isPending ? 'Syncing…' : 'Sync catalog from iRacing'}
        </Button>
        {sync.data && (
          <span className="text-sm text-muted-foreground">
            {sync.data.cars} cars · {sync.data.tracks} tracks · {sync.data.series} series
          </span>
        )}
        {sync.error && (
          <span className="text-sm text-destructive">
            {(sync.error as Error).message}
          </span>
        )}
      </CardContent>
    </Card>
  );
}
