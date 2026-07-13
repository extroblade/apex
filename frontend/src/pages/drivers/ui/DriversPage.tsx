import { useState } from 'react';
import { Link } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { useFeature, IRACING_OAUTH, IRacingUnavailable } from '@/entities/features';
import { useIRacingStatus } from '@/entities/iracing';
import { useDriverSearch } from '@/entities/driver';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export function DriversPage() {
  const { data: viewer, isLoading: viewerLoading } = useViewer();
  const iracing = useFeature(IRACING_OAUTH);
  const status = useIRacingStatus();
  const linked = Boolean(status.data?.linked);

  const [input, setInput] = useState('');
  const [term, setTerm] = useState('');
  const search = useDriverSearch(term, linked && iracing);

  if (viewerLoading) return null;

  if (viewer && !iracing) return <IRacingUnavailable />;

  if (!viewer) {
    return (
      <Gate
        title="Sign in to search drivers"
        description="Driver lookups use your own iRacing session, so you need an account."
        href="/login"
        cta="Log in"
      />
    );
  }

  if (status.data && !linked) {
    return (
      <Gate
        title="Link your iRacing account"
        description="Searches run through your own iRacing login. Link it on the dashboard first."
        href="/dashboard"
        cta="Go to dashboard"
      />
    );
  }

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setTerm(input.trim());
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Driver search</h1>
        <p className="text-sm text-muted-foreground">
          Look up any iRacing driver — fetched through your own linked account.
        </p>
      </div>

      <form onSubmit={onSubmit} className="flex gap-2">
        <Input
          placeholder="Driver name…"
          value={input}
          onChange={(e) => setInput(e.target.value)}
        />
        <Button type="submit" disabled={input.trim().length < 2}>
          Search
        </Button>
      </form>

      <Card>
        <CardContent className="pt-6">
          {!term && (
            <p className="text-sm text-muted-foreground">
              Enter a name and search to begin.
            </p>
          )}
          {search.isLoading && (
            <p className="text-sm text-muted-foreground">Searching…</p>
          )}
          {search.error && (
            <p className="text-sm text-destructive">{(search.error as Error).message}</p>
          )}
          {search.data && search.data.length === 0 && (
            <p className="text-sm text-muted-foreground">
              No drivers found for “{term}”.
            </p>
          )}
          {search.data && search.data.length > 0 && (
            <ul className="text-sm">
              {search.data.map((d) => (
                <li key={d.cust_id} className="border-b last:border-0">
                  <Link
                    href={`/drivers/${d.cust_id}`}
                    className="flex items-center justify-between py-2 hover:text-foreground"
                  >
                    <span className="font-medium">{d.display_name}</span>
                    <span className="text-muted-foreground">#{d.cust_id}</span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function Gate({
  title,
  description,
  href,
  cta,
}: {
  title: string;
  description: string;
  href: string;
  cta: string;
}) {
  return (
    <Card className="mx-auto max-w-sm">
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <Button asChild>
          <Link href={href}>{cta}</Link>
        </Button>
      </CardContent>
    </Card>
  );
}
