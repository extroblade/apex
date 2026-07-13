import { useState } from 'react';
import { Link } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useFeature, IRACING_OAUTH, IRacingUnavailable } from '@/entities/features';
import {
  useIRacingStatus,
  useComparator,
  type CompareDimension,
  type GroupStat,
} from '@/entities/iracing';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

const DIMENSIONS: { key: CompareDimension; label: string }[] = [
  { key: 'categories', label: 'Categories' },
  { key: 'cars', label: 'Cars' },
  { key: 'tracks', label: 'Tracks' },
];

export function ComparePage() {
  const { data: viewer, isLoading: viewerLoading } = useViewer();
  const iracing = useFeature(IRACING_OAUTH);
  const status = useIRacingStatus();
  const linked = Boolean(status.data?.linked);
  const [dimension, setDimension] = useState<CompareDimension>('categories');
  const compare = useComparator(dimension, linked && iracing);

  if (viewerLoading) return null;

  if (viewer && !iracing) return <IRacingUnavailable />;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>Sign in required</CardTitle>
          <CardDescription>Log in to compare your results.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild>
            <Link href="/login">Log in</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Comparators</h1>
        <p className="text-sm text-muted-foreground">
          Aggregated from your synced races. Ranked by average finish (best first). Sync
          races on the{' '}
          <Link href="/dashboard" className="underline">
            dashboard
          </Link>{' '}
          first.
        </p>
      </div>

      {!linked ? (
        <Card>
          <CardContent className="py-6 text-sm text-muted-foreground">
            Link your iRacing account on the dashboard to unlock comparators.
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="flex gap-2">
            {DIMENSIONS.map((d) => (
              <Button
                key={d.key}
                size="sm"
                variant={dimension === d.key ? 'default' : 'outline'}
                onClick={() => setDimension(d.key)}
              >
                {d.label}
              </Button>
            ))}
          </div>

          <Card>
            <CardContent className="pt-6">
              {compare.isLoading && (
                <p className="text-sm text-muted-foreground">Loading…</p>
              )}
              {compare.error && (
                <p className="text-sm text-destructive">
                  {(compare.error as Error).message}
                </p>
              )}
              {compare.data && <CompareTable rows={compare.data} />}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}

function CompareTable({ rows }: { rows: GroupStat[] }) {
  if (rows.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No synced races yet. Run a sync on the dashboard.
      </p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b text-left text-muted-foreground">
            <th className="py-2 pr-4 font-medium">Name</th>
            <th className="py-2 pr-4 font-medium">Races</th>
            <th className="py-2 pr-4 font-medium">Avg finish</th>
            <th className="py-2 pr-4 font-medium">Avg incidents</th>
            <th className="py-2 font-medium">Avg iR Δ</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.key} className="border-b last:border-0">
              <td className="py-2 pr-4">{r.label || r.key}</td>
              <td className="py-2 pr-4">{r.races}</td>
              <td className="py-2 pr-4">{r.avgFinish.toFixed(1)}</td>
              <td className="py-2 pr-4">{r.avgIncidents.toFixed(1)}</td>
              <td
                className={
                  r.avgIRatingGain >= 0 ? 'py-2 text-green-600' : 'py-2 text-destructive'
                }
              >
                {r.avgIRatingGain >= 0 ? '+' : ''}
                {r.avgIRatingGain.toFixed(0)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
