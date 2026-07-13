import { Link } from 'wouter';
import { Fuel, CalendarRange, Wrench, Target, type LucideIcon } from 'lucide-react';

import { useViewer } from '@/entities/viewer';
import { useFeature, IRACING_OAUTH } from '@/entities/features';
import { useTranslation } from '@/shared/i18n';
import { Button } from '@/shared/ui/button';
import { Aurora } from '@/shared/ui/fx/aurora';
import { ShinyText } from '@/shared/ui/fx/shiny-text';
import { SpotlightCard } from '@/shared/ui/fx/spotlight-card';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export function HomePage() {
  const { t } = useTranslation();
  const { data: viewer } = useViewer();
  const iracing = useFeature(IRACING_OAUTH);

  const features: { href: string; title: string; desc: string; Icon: LucideIcon }[] = [
    { href: '/fuel', title: t('home.fuelTitle'), desc: t('home.fuelDesc'), Icon: Fuel },
    {
      href: '/planner',
      title: t('nav.planner'),
      desc: t('home.plannerDesc'),
      Icon: CalendarRange,
    },
    { href: '/setups', title: t('nav.setups'), desc: t('home.setupsDesc'), Icon: Wrench },
    { href: '/goals', title: t('nav.goals'), desc: t('home.goalsDesc'), Icon: Target },
  ];

  return (
    <div className="space-y-10">
      {/* Hero with an animated aurora backdrop. */}
      <section className="relative overflow-hidden rounded-2xl border px-6 py-14 sm:px-10">
        <Aurora />
        <div className="relative max-w-xl">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">
            <ShinyText>{t('brand')}</ShinyText>
          </h1>
          <p className="mt-3 text-base text-muted-foreground sm:text-lg">
            {viewer ? t('home.signedInAs', { email: viewer.email }) : t('home.tagline')}
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Button asChild size="lg">
              <Link href={viewer ? '/this-week' : '/login'}>
                {viewer ? t('planner.thisWeekTitle') : t('common.logIn')}
              </Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href="/fuel">{t('home.openPlanner')}</Link>
            </Button>
          </div>
        </div>
      </section>

      <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {features.map(({ href, title, desc, Icon }) => (
          <Link key={href} href={href} className="group outline-none">
            <SpotlightCard className="h-full transition-colors group-focus-visible:ring-2 group-focus-visible:ring-ring">
              <CardHeader>
                <Icon className="mb-1 size-6 text-primary" aria-hidden />
                <CardTitle>{title}</CardTitle>
                <CardDescription>{desc}</CardDescription>
              </CardHeader>
            </SpotlightCard>
          </Link>
        ))}
      </section>

      {/* iRacing linking is feature-gated; hide the card entirely when off. */}
      {iracing && (
        <Card>
          <CardHeader>
            <CardTitle>{t('home.iracingTitle')}</CardTitle>
            <CardDescription>
              {viewer ? t('home.iracingSignedIn') : t('home.iracingSignedOut')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild>
              <Link href={viewer ? '/dashboard' : '/login'}>
                {viewer ? t('home.goToDashboard') : t('common.logIn')}
              </Link>
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
