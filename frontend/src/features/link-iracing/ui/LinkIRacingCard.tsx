import { useEffect } from 'react';
import { useSearch } from 'wouter';

import { useIRacingStatus } from '@/entities/iracing';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';
import { startIRacingLink, useUnlinkIRacing, useSyncIRacing } from '../api/use-link';

export function LinkIRacingCard() {
  const status = useIRacingStatus();
  const unlink = useUnlinkIRacing();
  const sync = useSyncIRacing();

  // The OAuth callback redirects back with ?iracing=linked or ?iracing_error=...
  const searchStr = useSearch();
  const params = new URLSearchParams(searchStr);
  const oauthError = params.get('iracing_error');
  const justLinked = params.get('iracing') === 'linked';

  useEffect(() => {
    if (justLinked) status.refetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [justLinked]);

  if (status.isLoading) {
    return (
      <Card>
        <CardContent className="py-6 text-sm text-muted-foreground">Loading…</CardContent>
      </Card>
    );
  }

  // Linked: show account details + sync/unlink.
  if (status.data?.linked && status.data.account) {
    const acc = status.data.account;
    return (
      <Card>
        <CardHeader>
          <CardTitle>iRacing account linked</CardTitle>
          <CardDescription>
            {acc.displayName} · cust #{acc.custId}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap items-center gap-3">
          <Button onClick={() => sync.mutate()} disabled={sync.isPending}>
            {sync.isPending ? 'Syncing…' : 'Sync recent races'}
          </Button>
          <Button
            variant="outline"
            onClick={() => unlink.mutate()}
            disabled={unlink.isPending}
          >
            Unlink
          </Button>
          {sync.data && (
            <span className="text-sm text-muted-foreground">
              Synced {sync.data.synced} races.
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

  // Not linked: OAuth connect.
  return (
    <Card>
      <CardHeader>
        <CardTitle>Connect your iRacing account</CardTitle>
        <CardDescription>
          You'll sign in on iRacing's own page — we never see your password. Your own
          session then fetches your stats.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {oauthError && (
          <p className="text-sm text-destructive">Linking failed: {oauthError}</p>
        )}
        <Button onClick={startIRacingLink}>Connect iRacing</Button>
      </CardContent>
    </Card>
  );
}
