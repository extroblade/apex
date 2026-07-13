import { useState } from 'react';
import { Wrench } from 'lucide-react';

import { useTranslation } from '@/shared/i18n';
import { isDev } from '@/shared/lib/dev';
import { useAllFeatures, useCockpitHealth, useToggleFeature } from '@/entities/features';
import { Button } from '@/shared/ui/button';
import { Switch } from '@/shared/ui/switch';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog';

/**
 * Cockpit dev-overlay. A floating wrench button (rendered only when the
 * `developer` cookie is set via ?dev=KEY) opens a modal listing every backend
 * feature flag with a live toggle, plus a backend health readout. All endpoints
 * 404 unless the cookie matches DEVELOPER_KEY, so this is inert in production.
 */
export function Cockpit() {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  // Gate on the cookie: no button at all for non-developers.
  if (!isDev()) return null;

  return (
    <>
      <Button
        variant="secondary"
        size="icon"
        aria-label={t('cockpit.open')}
        onClick={() => setOpen(true)}
        className="fixed bottom-24 right-4 z-40 rounded-full shadow-lg md:bottom-6"
      >
        <Wrench className="size-4" />
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('cockpit.title')}</DialogTitle>
            <DialogDescription>{t('cockpit.description')}</DialogDescription>
          </DialogHeader>
          {open && <CockpitBody />}
        </DialogContent>
      </Dialog>
    </>
  );
}

/** The overlay contents; mounted only while the dialog is open so the gated
 *  requests don't fire for everyone with the cookie on every page. */
function CockpitBody() {
  const { t } = useTranslation();
  const flags = useAllFeatures(true);
  const health = useCockpitHealth(true);
  const toggle = useToggleFeature();

  return (
    <div className="mt-2 flex flex-col gap-4">
      <section aria-labelledby="cockpit-flags">
        <h3 id="cockpit-flags" className="mb-2 text-sm font-medium">
          {t('cockpit.flags')}
        </h3>
        {flags.isLoading && (
          <p className="text-sm text-muted-foreground">{t('common.loading')}</p>
        )}
        {flags.isError && (
          <p className="text-sm text-destructive">{t('cockpit.error')}</p>
        )}
        {flags.data && (
          <ul className="flex flex-col gap-2">
            {Object.entries(flags.data)
              .sort(([a], [b]) => a.localeCompare(b))
              .map(([key, enabled]) => (
                <li
                  key={key}
                  className="flex items-center justify-between gap-4 rounded-md border px-3 py-2"
                >
                  <span className="font-mono text-sm" id={`flag-${key}`}>
                    {key}
                  </span>
                  <Switch
                    checked={enabled}
                    disabled={toggle.isPending}
                    aria-labelledby={`flag-${key}`}
                    onCheckedChange={(next) => toggle.mutate({ key, enabled: next })}
                  />
                </li>
              ))}
          </ul>
        )}
      </section>

      <section aria-labelledby="cockpit-health">
        <h3 id="cockpit-health" className="mb-2 text-sm font-medium">
          {t('cockpit.health')}
        </h3>
        <dl className="flex flex-col gap-1 text-sm">
          <HealthRow label={t('cockpit.db')} ok={health.data?.db} />
          <HealthRow
            label={t('cockpit.redis')}
            ok={health.data?.redis}
            disabled={health.data ? !health.data.redisEnabled : undefined}
            disabledLabel={t('cockpit.disabled')}
          />
        </dl>
      </section>
    </div>
  );
}

function HealthRow({
  label,
  ok,
  disabled,
  disabledLabel,
}: {
  label: string;
  ok?: boolean;
  disabled?: boolean;
  disabledLabel?: string;
}) {
  const { t } = useTranslation();
  let text = t('cockpit.unknown');
  let tone = 'text-muted-foreground';
  if (disabled) {
    text = disabledLabel ?? t('cockpit.disabled');
  } else if (ok === true) {
    text = t('cockpit.up');
    tone = 'text-primary';
  } else if (ok === false) {
    text = t('cockpit.down');
    tone = 'text-destructive';
  }
  return (
    <div className="flex items-center justify-between gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className={tone}>{text}</dd>
    </div>
  );
}
