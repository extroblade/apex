import { Link } from 'wouter';

import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

/** Shown on iRacing-only pages when the iracing_oauth feature is disabled. */
export function IRacingUnavailable() {
  return (
    <Card className="mx-auto max-w-md">
      <CardHeader>
        <CardTitle>iRacing features are unavailable</CardTitle>
        <CardDescription>
          iRacing account linking is currently turned off. You can still plan races
          using the Planner and Garage.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Button asChild>
          <Link href="/planner">Go to planner</Link>
        </Button>
      </CardContent>
    </Card>
  );
}
