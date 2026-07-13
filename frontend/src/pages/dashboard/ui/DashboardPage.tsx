import { Link } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useFeature, IRACING_OAUTH, IRacingUnavailable } from '@/entities/features';
import { useIRacingStatus, useDashboard } from '@/entities/iracing';
import { LinkIRacingCard } from '@/features/link-iracing';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

const CATEGORY_LABELS: Record<number, string> = {
  1: 'Oval',
  2: 'Road',
  3: 'Dirt Oval',
  4: 'Dirt Road',
  5: 'Sports Car',
  6: 'Formula Car',
};

function categoryLabel(id: number, fallback: string) {
  return CATEGORY_LABELS[id] ?? fallback ?? `Category ${id}`;
}

export function DashboardPage() {
  const { data: viewer, isLoading: viewerLoading } = useViewer();
  const iracing = useFeature(IRACING_OAUTH);
  const status = useIRacingStatus();
  const linked = Boolean(status.data?.linked);
  const dashboard = useDashboard(linked && iracing);

  if (viewerLoading) return null;

  if (viewer && !iracing) return <IRacingUnavailable />;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>Sign in required</CardTitle>
          <CardDescription>Log in to view your dashboard.</CardDescription>
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
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Your current iRacing licenses, career, and recent races.
        </p>
      </div>

      <LinkIRacingCard />

      {linked && (
        <>
          {dashboard.isLoading && (
            <p className="text-sm text-muted-foreground">
              Loading your stats from iRacing…
            </p>
          )}
          {dashboard.error && (
            <p className="text-sm text-destructive">
              {(dashboard.error as Error).message}
            </p>
          )}
          {dashboard.data && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Licenses &amp; iRating</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {dashboard.data.licenses.map((l) => (
                      <div key={l.category_id} className="rounded-md border p-3">
                        <div className="text-sm font-medium">
                          {categoryLabel(l.category_id, l.category)}
                        </div>
                        <div className="mt-1 text-sm text-muted-foreground">
                          iR {l.irating} · SR {l.safety_rating.toFixed(2)}
                        </div>
                      </div>
                    ))}
                    {dashboard.data.licenses.length === 0 && (
                      <p className="text-sm text-muted-foreground">No licenses found.</p>
                    )}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Recent races</CardTitle>
                </CardHeader>
                <CardContent>
                  {dashboard.data.recent.length === 0 ? (
                    <p className="text-sm text-muted-foreground">No recent races.</p>
                  ) : (
                    <ul className="text-sm">
                      {dashboard.data.recent.slice(0, 10).map((r) => {
                        const delta = r.newi_rating - r.oldi_rating;
                        return (
                          <li
                            key={r.subsession_id}
                            className="flex items-center justify-between gap-3 border-b py-1.5 last:border-0"
                          >
                            <span className="min-w-0 flex-1 truncate">
                              {r.series_name}
                            </span>
                            <span className="text-muted-foreground">
                              P{r.finish_position} · {r.incidents}x
                            </span>
                            <span
                              className={
                                delta >= 0 ? 'text-green-600' : 'text-destructive'
                              }
                            >
                              {delta >= 0 ? '+' : ''}
                              {delta}
                            </span>
                          </li>
                        );
                      })}
                    </ul>
                  )}
                </CardContent>
              </Card>
            </div>
          )}
        </>
      )}
    </div>
  );
}
