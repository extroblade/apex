import { Link } from 'wouter';

import { useViewer } from '@/entities/viewer';
import { GarageManager } from '@/features/manage-content';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export function GaragePage() {
  const { data: viewer, isLoading } = useViewer();

  if (isLoading) return null;

  if (!viewer) {
    return (
      <Card className="mx-auto max-w-sm">
        <CardHeader>
          <CardTitle>Sign in required</CardTitle>
          <CardDescription>Log in to manage your garage.</CardDescription>
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
        <h1 className="text-2xl font-semibold">Garage</h1>
        <p className="text-sm text-muted-foreground">
          Mark the cars and tracks you own, and favorite the series you race — your set is
          saved and drives the planner.
        </p>
      </div>
      <GarageManager />
    </div>
  );
}
