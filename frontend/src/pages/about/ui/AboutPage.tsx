import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';

export function AboutPage() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>About</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-sm text-muted-foreground">
        <p>
          Full-stack starter: Go + chi + MySQL backend, React + TypeScript frontend built
          with rsbuild.
        </p>
        <p>
          Frontend stack: wouter, zustand, TanStack Query, shadcn/ui, Tailwind CSS v4,
          clsx — organized with Feature-Sliced Design.
        </p>
      </CardContent>
    </Card>
  );
}
