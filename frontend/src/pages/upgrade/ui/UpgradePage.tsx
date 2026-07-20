import { Link } from 'wouter';

import { useViewer } from '@/entities/viewer';
import {
  useBillingPlans,
  useSetDevTier,
  useStartCheckout,
  useStartPortal,
  useSubscription,
} from '@/entities/subscription';
import { isDev } from '@/shared/lib/dev';
import { useTranslation } from '@/shared/i18n';
import { Button } from '@/shared/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card';
import { Badge } from '@/shared/ui/badge';

export function UpgradePage() {
  const { t } = useTranslation();
  const { data: viewer, isLoading: viewerLoading } = useViewer();
  const plans = useBillingPlans();
  const subscription = useSubscription(Boolean(viewer));
  const setTier = useSetDevTier();
  const checkout = useStartCheckout();
  const portal = useStartPortal();
  const dev = Boolean(isDev());

  if (viewerLoading) return null;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t('billing.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('billing.subtitle')}</p>
      </div>

      {viewer && (
        <Card>
          <CardHeader>
            <CardTitle>{t('billing.currentPlan')}</CardTitle>
            <CardDescription>{t('billing.currentPlanHint')}</CardDescription>
          </CardHeader>
          <CardContent>
            <Badge className={subscription.data?.pro ? 'border-primary/40 text-primary' : ''}>
              {subscription.data?.pro ? t('billing.proName') : t('billing.freeName')}
            </Badge>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-2">
        {(plans.data ?? []).map((plan) => {
          const isPro = plan.key === 'pro';
          const isCurrent = isPro ? Boolean(subscription.data?.pro) : !subscription.data?.pro;
          return (
            <Card key={plan.key}>
              <CardHeader>
                <CardTitle>{isPro ? t('billing.proName') : t('billing.freeName')}</CardTitle>
                <CardDescription>
                  {plan.price} / {plan.interval}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <ul className="space-y-1 text-sm text-muted-foreground">
                  {plan.features.map((f) => (
                    <li key={f}>• {f}</li>
                  ))}
                </ul>
                {!viewer ? (
                  <Button asChild>
                    <Link href="/login">{t('billing.loginToUpgrade')}</Link>
                  </Button>
                ) : isCurrent ? (
                  isPro ? (
                    <Button
                      type="button"
                      variant="outline"
                      disabled={portal.isPending}
                      onClick={() =>
                        portal.mutate(undefined, {
                          onSuccess: ({ url }) => window.location.assign(url),
                        })
                      }
                    >
                      {portal.isPending ? t('common.loading') : t('billing.manageBilling')}
                    </Button>
                  ) : (
                    <Button disabled>{t('billing.currentPlanButton')}</Button>
                  )
                ) : isPro ? (
                  <div className="space-y-2">
                    <Button
                      type="button"
                      disabled={checkout.isPending}
                      onClick={() =>
                        checkout.mutate('pro', {
                          onSuccess: ({ url }) => window.location.assign(url),
                        })
                      }
                    >
                      {checkout.isPending ? t('common.loading') : t('billing.startCheckout')}
                    </Button>
                    {dev && (
                      <Button
                        type="button"
                        variant="outline"
                        disabled={setTier.isPending}
                        onClick={() => setTier.mutate('pro')}
                      >
                        {setTier.isPending ? t('common.loading') : t('billing.devActivatePro')}
                      </Button>
                    )}
                  </div>
                ) : (
                  dev && (
                    <Button
                      type="button"
                      variant="outline"
                      disabled={setTier.isPending}
                      onClick={() => setTier.mutate('free')}
                    >
                      {setTier.isPending ? t('common.loading') : t('billing.devSwitchFree')}
                    </Button>
                  )
                )}
              </CardContent>
            </Card>
          );
        })}
      </div>

      {plans.error && <p className="text-sm text-destructive">{(plans.error as Error).message}</p>}
      {subscription.error && (
        <p className="text-sm text-destructive">{(subscription.error as Error).message}</p>
      )}
      {checkout.error && (
        <p className="text-sm text-destructive">{(checkout.error as Error).message}</p>
      )}
      {portal.error && <p className="text-sm text-destructive">{(portal.error as Error).message}</p>}
    </div>
  );
}
