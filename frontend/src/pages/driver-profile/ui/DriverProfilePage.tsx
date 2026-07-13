import { Link, useParams } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useDriverProfile } from '@/entities/driver';
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

export function DriverProfilePage() {
  const params = useParams();
  const custId = Number(params.custId);
  const { data: viewer, isLoading: viewerLoading } = useViewer();
  const profile = useDriverProfile(custId, Boolean(viewer));

  if (viewerLoading) return null;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>Sign in required</CardTitle>
          <CardDescription>
            Driver profiles are fetched through your own iRacing session.
          </CardDescription>
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">
            {profile.data?.displayName ?? `Driver #${custId}`}
          </h1>
          {profile.data && (
            <p className="text-sm text-muted-foreground">
              cust #{profile.data.custId}
              {profile.data.cachedAt &&
                ` · data as of ${new Date(profile.data.cachedAt).toLocaleString()}`}
            </p>
          )}
        </div>
        <Button asChild variant="outline" size="sm">
          <Link href="/drivers">← Search</Link>
        </Button>
      </div>

      {profile.isLoading && (
        <p className="text-sm text-muted-foreground">Loading driver…</p>
      )}
      {profile.error && (
        <p className="text-sm text-destructive">{(profile.error as Error).message}</p>
      )}

      {profile.data && (
        <>
          <Card>
            <CardHeader>
              <CardTitle>Licenses &amp; iRating</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {profile.data.licenses.map((l) => (
                  <div key={l.category_id} className="rounded-md border p-3">
                    <div className="text-sm font-medium">
                      {categoryLabel(l.category_id, l.category)}
                    </div>
                    <div className="mt-1 text-sm text-muted-foreground">
                      iR {l.irating} · SR {l.safety_rating.toFixed(2)}
                    </div>
                  </div>
                ))}
                {profile.data.licenses.length === 0 && (
                  <p className="text-sm text-muted-foreground">No licenses.</p>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Career</CardTitle>
            </CardHeader>
            <CardContent>
              {profile.data.career.length === 0 ? (
                <p className="text-sm text-muted-foreground">No career stats.</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-left text-muted-foreground">
                        <th className="py-2 pr-4 font-medium">Category</th>
                        <th className="py-2 pr-4 font-medium">Starts</th>
                        <th className="py-2 pr-4 font-medium">Wins</th>
                        <th className="py-2 pr-4 font-medium">Top 5</th>
                        <th className="py-2 pr-4 font-medium">Avg finish</th>
                        <th className="py-2 font-medium">Avg inc</th>
                      </tr>
                    </thead>
                    <tbody>
                      {profile.data.career.map((c) => (
                        <tr key={c.category_id} className="border-b last:border-0">
                          <td className="py-2 pr-4">
                            {categoryLabel(c.category_id, c.category)}
                          </td>
                          <td className="py-2 pr-4">{c.starts}</td>
                          <td className="py-2 pr-4">{c.wins}</td>
                          <td className="py-2 pr-4">{c.top5}</td>
                          <td className="py-2 pr-4">
                            {c.avg_finish_position.toFixed(1)}
                          </td>
                          <td className="py-2">{c.avg_incidents.toFixed(1)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Recent races</CardTitle>
            </CardHeader>
            <CardContent>
              {profile.data.recent.length === 0 ? (
                <p className="text-sm text-muted-foreground">No recent races.</p>
              ) : (
                <ul className="text-sm">
                  {profile.data.recent.slice(0, 15).map((r) => {
                    const delta = r.newi_rating - r.oldi_rating;
                    return (
                      <li
                        key={r.subsession_id}
                        className="flex items-center justify-between gap-3 border-b py-1.5 last:border-0"
                      >
                        <span className="min-w-0 flex-1 truncate">{r.series_name}</span>
                        <span className="text-muted-foreground">
                          P{r.finish_position} · {r.incidents}x
                        </span>
                        <span
                          className={delta >= 0 ? 'text-green-600' : 'text-destructive'}
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
        </>
      )}
    </div>
  );
}
