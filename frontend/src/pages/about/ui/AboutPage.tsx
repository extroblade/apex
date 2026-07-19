import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';

export function AboutPage() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>About Apex</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3 text-sm text-muted-foreground">
        <p>
          Apex is a companion app for sim racers. It helps you plan and get faster: a fuel
          &amp; stint calculator, a setup generator, a full season planner, your garage,
          and season goals — in one place.
        </p>
        <p>
          The core tools work without linking any external account. Apex is an independent
          project and is not affiliated with, endorsed by, or sponsored by iRacing.com
          Motorsport Simulations, LLC.
        </p>
      </CardContent>
    </Card>
  );
}
